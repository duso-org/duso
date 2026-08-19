package runtime

import (
	"crypto/tls"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/duso-org/duso/pkg/core"
)

// defaultCertReloadInterval is how often a running server re-checks its TLS
// files. Certificates renew on the order of weeks and renewers swap them well
// before expiry, so a daily check is plenty; cert_reload_interval takes it
// lower for anyone who wants a tighter bound.
const defaultCertReloadInterval = 24 * time.Hour

// certReloadRetryInterval is the shorter delay used after a failed reload. The
// usual cause is catching the renewer between writing the cert and writing the
// key, where the two momentarily disagree; a minute later they match. A server
// checking more often than this retries on its own cadence instead -- a retry
// must never be slower than the regular check.
const certReloadRetryInterval = time.Minute

// fileStamp is what "the file changed" means here: mtime plus size. Comparable
// with ==, unlike time.Time.
type fileStamp struct {
	modNano int64
	size    int64
}

// certReloader owns the keypair a TLS server hands out during handshakes, and
// swaps in a new one when the files on disk change. Renewals (certbot and
// friends) rewrite the cert in place every 60-90 days; without this the process
// would keep serving the old pair until someone restarted it, which is how
// certificates expire on servers that were renewing correctly the whole time.
type certReloader struct {
	certFile string
	keyFile  string
	interval time.Duration

	// cert is read on every handshake and written only by the reload goroutine,
	// so it is an atomic pointer rather than a mutex: handshakes pay a load.
	cert atomic.Pointer[tls.Certificate]

	// The stamps are touched only by load(), which runs once before the reload
	// goroutine starts and thereafter only inside it -- no concurrent access.
	certStamp fileStamp
	keyStamp  fileStamp
}

// newCertReloader loads the initial pair. A bad pair at startup is fatal, same
// as it was when the files went straight to ListenAndServeTLS.
func newCertReloader(certFile, keyFile string, interval time.Duration) (*certReloader, error) {
	if interval <= 0 {
		interval = defaultCertReloadInterval
	}
	r := &certReloader{certFile: certFile, keyFile: keyFile, interval: interval}
	if err := r.load(); err != nil {
		return nil, err
	}
	return r, nil
}

// GetCertificate is what crypto/tls calls on every handshake.
func (r *certReloader) GetCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	return r.cert.Load(), nil
}

// stat reads both files' stamps. Both or neither: a stamp for one file is not
// enough to decide anything.
func (r *certReloader) stat() (certStat, keyStat fileStamp, err error) {
	certInfo, err := os.Stat(r.certFile)
	if err != nil {
		return fileStamp{}, fileStamp{}, err
	}
	keyInfo, err := os.Stat(r.keyFile)
	if err != nil {
		return fileStamp{}, fileStamp{}, err
	}
	return fileStamp{certInfo.ModTime().UnixNano(), certInfo.Size()},
		fileStamp{keyInfo.ModTime().UnixNano(), keyInfo.Size()}, nil
}

// load reads the pair from disk and, only if it parses and the key matches the
// cert, swaps it in. A failure leaves the previously loaded pair serving.
func (r *certReloader) load() error {
	// Stamp before reading, so a write that lands during the read shows up as a
	// change on the next check instead of being stamped as already-loaded.
	certStat, keyStat, statErr := r.stat()

	cert, err := tls.LoadX509KeyPair(r.certFile, r.keyFile)
	if err != nil {
		return err
	}
	r.cert.Store(&cert)

	if statErr != nil {
		// Unstampable but readable. Zero stamps mean the next check sees a
		// difference and reloads, rather than assuming this pair is current.
		r.certStamp, r.keyStamp = fileStamp{}, fileStamp{}
		return nil
	}
	r.certStamp, r.keyStamp = certStat, keyStat
	return nil
}

// changed reports whether either file moved since the last successful load.
func (r *certReloader) changed() bool {
	certStat, keyStat, err := r.stat()
	if err != nil {
		// A file is missing or unreadable -- likely the renewer mid-swap. There
		// is nothing to reload from, so hold what we have and look again later.
		return false
	}
	return certStat != r.certStamp || keyStat != r.keyStamp
}

// watch re-checks the files until stop is closed. It runs for the life of the
// server, so a certificate renewed underneath a long-lived process is picked up
// without a restart.
func (r *certReloader) watch(stop <-chan struct{}) {
	defer core.RecoverPanic("http_server TLS certificate reload")

	timer := time.NewTimer(r.interval)
	defer timer.Stop()

	for {
		select {
		case <-stop:
			return
		case <-timer.C:
		}

		next := r.interval
		if r.changed() {
			if err := r.load(); err != nil {
				// Keep serving the pair already in hand: a certificate with days
				// left on it beats no certificate at all.
				retry := certReloadRetryInterval
				if r.interval < retry {
					retry = r.interval
				}
				fmt.Fprintf(os.Stderr, "duso: TLS certificate %s changed but could not be loaded: %v (still serving the previous certificate, retrying in %s)\n",
					r.certFile, err, retry)
				next = retry
			} else {
				fmt.Fprintf(os.Stderr, "duso: reloaded TLS certificate %s\n", r.certFile)
			}
		}
		timer.Reset(next)
	}
}
