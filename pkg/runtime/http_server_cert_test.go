package runtime

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// writeKeyPair writes a self-signed cert/key pair carrying the given serial, so
// a test can tell which generation of the certificate a handshake served.
func writeKeyPair(t *testing.T, certPath, keyPath string, serial int64) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(serial),
		Subject:      pkix.Name{CommonName: "duso-test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatalf("marshal key: %v", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
	if err := os.WriteFile(certPath, certPEM, 0o600); err != nil {
		t.Fatalf("write cert: %v", err)
	}
	if err := os.WriteFile(keyPath, keyPEM, 0o600); err != nil {
		t.Fatalf("write key: %v", err)
	}
}

// servedSerial completes a real handshake against addr and reports the serial
// of the certificate the server presented.
func servedSerial(t *testing.T, addr string) int64 {
	t.Helper()
	conn, err := tls.Dial("tcp", addr, &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()
	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		t.Fatal("server presented no certificate")
	}
	return certs[0].SerialNumber.Int64()
}

// A certificate rewritten under a running server is picked up on the next
// check, without a restart -- the whole point of the reloader.
func TestCertReloaderPicksUpRenewal(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")
	writeKeyPair(t, certPath, keyPath, 1001)

	reloader, err := newCertReloader(certPath, keyPath, 10*time.Millisecond)
	if err != nil {
		t.Fatalf("newCertReloader: %v", err)
	}
	stop := make(chan struct{})
	defer close(stop)
	go reloader.watch(stop)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := &http.Server{
		Handler:   http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
		TLSConfig: &tls.Config{GetCertificate: reloader.GetCertificate},
	}
	go srv.ServeTLS(ln, "", "")
	defer srv.Close()

	addr := ln.Addr().String()
	if got := servedSerial(t, addr); got != 1001 {
		t.Fatalf("before renewal: served serial %d, want 1001", got)
	}

	// Renew: same paths, new pair. mtime granularity on some filesystems is
	// coarse, so make sure the write lands in a later nanosecond than the load.
	time.Sleep(20 * time.Millisecond)
	writeKeyPair(t, certPath, keyPath, 2002)

	deadline := time.Now().Add(5 * time.Second)
	for {
		if got := servedSerial(t, addr); got == 2002 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("after renewal: still serving old certificate, want serial 2002")
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// A pair that does not load -- here the renewer's cert has landed but its key
// has not, so the two disagree -- must not take the server's certificate away.
func TestCertReloaderKeepsOldCertOnBadPair(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")
	writeKeyPair(t, certPath, keyPath, 1001)

	reloader, err := newCertReloader(certPath, keyPath, time.Hour)
	if err != nil {
		t.Fatalf("newCertReloader: %v", err)
	}

	// Write only the cert half of a new pair: it no longer matches the key.
	mismatchedDir := t.TempDir()
	otherCert := filepath.Join(mismatchedDir, "cert.pem")
	otherKey := filepath.Join(mismatchedDir, "key.pem")
	writeKeyPair(t, otherCert, otherKey, 2002)
	newCertPEM, err := os.ReadFile(otherCert)
	if err != nil {
		t.Fatalf("read replacement cert: %v", err)
	}
	if err := os.WriteFile(certPath, newCertPEM, 0o600); err != nil {
		t.Fatalf("write replacement cert: %v", err)
	}

	if !reloader.changed() {
		t.Fatal("changed() did not notice the rewritten certificate")
	}
	if err := reloader.load(); err == nil {
		t.Fatal("load() accepted a cert that does not match the key")
	}
	cert, err := reloader.GetCertificate(nil)
	if err != nil {
		t.Fatalf("GetCertificate: %v", err)
	}
	leaf, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		t.Fatalf("parse served certificate: %v", err)
	}
	if leaf.SerialNumber.Int64() != 1001 {
		t.Fatalf("served serial %d after a failed reload, want the previous 1001", leaf.SerialNumber.Int64())
	}

	// The failed load must not stamp the bad pair as current, or the fixed pair
	// would never be noticed.
	if !reloader.changed() {
		t.Fatal("changed() went quiet after a failed load; a repaired pair would never reload")
	}

	// Now the key catches up and the pair matches again.
	newKeyPEM, err := os.ReadFile(otherKey)
	if err != nil {
		t.Fatalf("read replacement key: %v", err)
	}
	if err := os.WriteFile(keyPath, newKeyPEM, 0o600); err != nil {
		t.Fatalf("write replacement key: %v", err)
	}
	if err := reloader.load(); err != nil {
		t.Fatalf("load() rejected a valid renewed pair: %v", err)
	}
	if reloader.changed() {
		t.Fatal("changed() still reports a change right after a successful load")
	}
}

// An unreadable pair at startup is a startup error, as it was when the paths
// went straight to ListenAndServeTLS.
func TestCertReloaderStartupFailure(t *testing.T) {
	dir := t.TempDir()
	if _, err := newCertReloader(filepath.Join(dir, "missing.pem"), filepath.Join(dir, "missing.key"), time.Hour); err == nil {
		t.Fatal("newCertReloader accepted missing certificate files")
	}
}
