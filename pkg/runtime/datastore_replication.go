package runtime

import (
	"bytes"
	"crypto/subtle"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/websocket"

	"github.com/duso-org/duso/pkg/core"
)

// Datastore replication: one leader streams its WAL to followers over a
// websocket. The unit shipped is the frame writeWALOp already built, forked
// before it reaches the file, so snapshotting and truncation stay invisible to
// the stream. Followers apply through applyWALRecord, the same function crash
// recovery uses.
//
// No consensus, no automatic promotion. A follower forwards writes to the leader
// rather than applying them; promotion is a config change plus a restart.
// See docs/reference/datastore_replication.md.

const (
	// replProtoVersion is bumped when the handshake or stream framing changes in
	// a way an older peer cannot read.
	replProtoVersion = 2

	// Recent frames kept so a blipping follower resumes instead of resyncing.
	// Falling past this costs a snapshot encode, which stalls leader writes.
	replDefaultBuffer = 64 * 1024 * 1024

	// replSnapshotChunk bounds one websocket message during a resync so a large
	// store does not become one enormous allocation on either side.
	replSnapshotChunk = 4 * 1024 * 1024

	// replHandshakeLimit is the payload cap in force before a connection has been
	// authenticated. Deliberately small: until lookup() succeeds the peer is
	// anonymous, and the limit is the only thing bounding what it can make this
	// process allocate. Raised to replPayloadLimit once the secret checks out.
	replHandshakeLimit = replSnapshotChunk + 1024

	// replMaxBatchBytes bounds one coalesced frame message. Without a bound,
	// since() hands back everything a follower missed in a single message, so a
	// follower resuming after a blip could be given the whole ring — megabytes
	// built entirely out of small records, over any sane payload cap, with no
	// large value involved anywhere. Small enough to stay deliverable, large
	// enough that a busy leader still amortizes syscalls across many records.
	replMaxBatchBytes = 1024 * 1024

	// replFrameOverhead is slack above maxValueSize for everything a frame wraps
	// around the value: frame header, key, opcode, and the nonce and tag added
	// when the log is encrypted.
	replFrameOverhead = 1024 * 1024

	// replUncappedPayloadLimit applies when max_value_size is 0, which disables
	// the value cap entirely. Something still has to bound one allocation, so
	// this is the ceiling for a store that asked for no ceiling.
	replUncappedPayloadLimit = 256 * 1024 * 1024

	// replHeartbeat is how often an idle leader sends an empty message. Without
	// it a leader with no writes cannot tell a live follower from a dead socket,
	// and neither can the follower.
	replHeartbeat = 10 * time.Second

	// replReadTimeout is how long a follower waits for anything at all before
	// deciding the connection is gone. Three missed heartbeats.
	replReadTimeout = 35 * time.Second

	// replReconnectMin/Max bound the follower's reconnect backoff.
	replReconnectMin = 500 * time.Millisecond
	replReconnectMax = 15 * time.Second
)

// replRole is how a store participates in replication.
type replRole int

const (
	replRoleNone     replRole = iota // not replicated
	replRoleLeader                   // serves the stream, accepts writes
	replRoleFollower                 // consumes the stream, forwards writes
)

func (r replRole) String() string {
	switch r {
	case replRoleLeader:
		return "leader"
	case replRoleFollower:
		return "follower"
	default:
		return "standalone"
	}
}

// replState holds everything replication adds to a datastore. It hangs off
// DatastoreValue as a single nil-able pointer so a store that does not replicate
// carries one word and no behavior.
type replState struct {
	role replRole

	// secret grants forwarding; readSecret grants the stream only. A follower
	// presents one of them and the leader tags the connection accordingly.
	secret     string
	readSecret string

	// canWrite is what the leader granted this follower, learned at handshake so
	// a read-only replica fails writes locally instead of round-tripping.
	canWrite atomic.Bool

	// epoch identifies a leadership term. It rides in the reserved field of both
	// the WAL and snapshot headers, and is bumped exactly once — when a store
	// that was last written as a follower starts as a leader. See replBumpEpoch.
	epoch uint32

	// leader side
	listenAddr string
	certFile   string
	keyFile    string
	ring       *replRing

	// leader side, observable
	followers atomic.Int64 // currently connected followers

	// follower side
	leaderURL string
	rootCAs   *x509.CertPool // trusted issuers for wss://; nil uses the system pool
	cursor    atomic.Uint64  // highest seq applied
	conn      atomic.Pointer[replConn]
	connected atomic.Bool
	lastError atomic.Pointer[string]

	stop     chan struct{}
	stopOnce sync.Once
	wg       sync.WaitGroup
}

// ---------------------------------------------------------------------------
// Configuration
// ---------------------------------------------------------------------------

// replConfigKeys is every option this file consumes, so applyDatastoreConfig can
// reject a store that sets one without setting a role.
var replConfigKeys = []string{
	"replicate_listen", "replicate_from", "replicate_secret",
	"replicate_readonly_secret", "replicate_buffer",
	"replicate_cert_file", "replicate_key_file", "replicate_ca_file",
}

// replPayloadLimit is the largest websocket message a connection to this store
// will accept: big enough for the largest frame the store can produce, and never
// smaller than a snapshot chunk. A frame carrying a legal value must always be
// deliverable — a value the store accepts but cannot replicate would strand the
// follower at the record before it, with no way for the caller to know.
func replPayloadLimit(maxValueSize int64) int {
	limit := int64(replHandshakeLimit)
	if maxValueSize <= 0 {
		limit = max(limit, replUncappedPayloadLimit)
	} else {
		limit = max(limit, maxValueSize+replFrameOverhead)
	}
	return int(limit)
}

// replIsLoopbackHost reports whether traffic stays on this machine. An empty
// host (":7777") is every interface. Names other than localhost count as remote
// rather than making startup depend on DNS.
func replIsLoopbackHost(host string) bool {
	if host == "" {
		return false
	}
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// replPeerSpeaksTLS reports whether the address answers with a TLS handshake.
// Diagnostic only, called after a plaintext dial already failed. Answers yes
// only on positive evidence: treating anything that is not RecordHeaderError as
// TLS would tell someone whose leader is merely down to go configure TLS.
func replPeerSpeaksTLS(hostport string) bool {
	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 3 * time.Second}, "tcp", hostport, nil)
	if err == nil {
		_ = conn.Close()
		return true // handshake completed against a trusted certificate
	}

	// A certificate this process does not trust, or a name that does not match,
	// still proves the peer served one. So does a TLS alert.
	var certErr *tls.CertificateVerificationError
	var alertErr tls.AlertError
	return errors.As(err, &certErr) || errors.As(err, &alertErr)
}

// replConfigure reads the replicate_* options and decides this store's role. It
// does not open sockets — replStart does that, after recovery has finished, so a
// follower never connects to a store still replaying its log.
func replConfigure(ds *DatastoreValue, config map[string]any) error {
	str := func(key string) (string, bool, error) {
		raw, ok := config[key]
		if !ok {
			return "", false, nil
		}
		s, ok := raw.(string)
		if !ok {
			return "", false, fmt.Errorf("datastore(%q): %s must be a string", ds.namespace, key)
		}
		return s, s != "", nil
	}

	listen, isLeader, err := str("replicate_listen")
	if err != nil {
		return err
	}
	from, isFollower, err := str("replicate_from")
	if err != nil {
		return err
	}
	secret, hasSecret, err := str("replicate_secret")
	if err != nil {
		return err
	}
	certFile, _, err := str("replicate_cert_file")
	if err != nil {
		return err
	}
	keyFile, _, err := str("replicate_key_file")
	if err != nil {
		return err
	}
	caFile, _, err := str("replicate_ca_file")
	if err != nil {
		return err
	}
	readSecret, hasReadSecret, err := str("replicate_readonly_secret")
	if err != nil {
		return err
	}

	if isLeader && isFollower {
		return fmt.Errorf("datastore(%q): set replicate_listen to lead or replicate_from to follow, not both", ds.namespace)
	}
	if !isLeader && !isFollower {
		// Catch the half-configured case rather than starting standalone and
		// letting someone discover months later that nothing ever replicated.
		for _, k := range replConfigKeys {
			if _, set := config[k]; set {
				return fmt.Errorf("datastore(%q): %s needs either replicate_listen (to lead) or replicate_from (to follow)",
					ds.namespace, k)
			}
		}
		return nil
	}

	// The stream carries the whole store. An unauthenticated listener would hand
	// it to anyone who can reach the port.
	if !hasSecret {
		return fmt.Errorf("datastore(%q): replication requires replicate_secret, shared by the leader and every follower", ds.namespace)
	}
	if (certFile == "") != (keyFile == "") {
		return fmt.Errorf("datastore(%q): replicate_cert_file and replicate_key_file must be set together", ds.namespace)
	}
	if isFollower && (certFile != "" || keyFile != "") {
		return fmt.Errorf("datastore(%q): replicate_cert_file and replicate_key_file configure a leader's listener; a follower sets replicate_from to a wss:// URL instead", ds.namespace)
	}
	if isFollower && hasReadSecret {
		return fmt.Errorf("datastore(%q): replicate_readonly_secret is a leader option; a follower presents whichever secret it was given as replicate_secret", ds.namespace)
	}
	if hasReadSecret && readSecret == secret {
		return fmt.Errorf("datastore(%q): replicate_readonly_secret must differ from replicate_secret, or the grant it implies means nothing", ds.namespace)
	}
	if isLeader && caFile != "" {
		return fmt.Errorf("datastore(%q): replicate_ca_file tells a follower which CA to trust; a leader sets replicate_cert_file and replicate_key_file instead", ds.namespace)
	}
	if ds.readonly {
		return fmt.Errorf("datastore(%q) is read-only and cannot participate in replication", ds.namespace)
	}

	buffer := replDefaultBuffer
	if raw, ok := config["replicate_buffer"]; ok {
		size, ok := raw.(float64)
		if !ok {
			return fmt.Errorf("datastore(%q): replicate_buffer must be a number of bytes", ds.namespace)
		}
		if size <= 0 {
			return fmt.Errorf("datastore(%q): replicate_buffer must be positive", ds.namespace)
		}
		buffer = int(size)
	}

	repl := &replState{secret: secret, readSecret: readSecret, stop: make(chan struct{})}
	// A leader writes locally. A follower is granted its access by the leader's
	// welcome and assumes nothing until then.
	repl.canWrite.Store(isLeader)

	// What the last snapshot says about this store decides whether starting as a
	// leader is a restart or a promotion.
	var snapEpoch uint32
	var wasFollower bool
	if ds.persistPath != "" {
		snapEpoch, wasFollower = readSnapshotProvenance(ds.persistPath)
	}

	if isFollower {
		repl.role = replRoleFollower
		repl.leaderURL = from

		// Validate the URL here rather than letting the first dial fail: a typo in
		// a scheme should stop the process, not become a reconnect loop.
		u, err := url.Parse(from)
		if err != nil {
			return fmt.Errorf("datastore(%q): replicate_from %q is not a valid URL: %v", ds.namespace, from, err)
		}
		switch u.Scheme {
		case "ws":
			// The handshake carries replicate_secret, and over ws:// it carries it
			// in the clear. Anyone who can read the link gets a live copy of the
			// whole store, so this is worth saying out loud rather than leaving in
			// the documentation.
			if !replIsLoopbackHost(u.Hostname()) {
				fmt.Fprintf(os.Stderr, "duso: warning: datastore %q replicates from %s over an unencrypted connection — replicate_secret and every value cross the network in the clear. Use wss:// unless this link is already private (VPN, WireGuard, internal VPC).\n",
					ds.namespace, from)
			}
		case "wss":
		default:
			return fmt.Errorf("datastore(%q): replicate_from must be a ws:// or wss:// URL, got %q", ds.namespace, from)
		}

		// An internal leader almost never has a publicly-trusted certificate, so
		// without this wss:// would only work against a public CA. Loading the
		// pool here makes a bad path fail the datastore() call rather than
		// surfacing as a connection that never succeeds.
		if caFile != "" {
			pem, err := os.ReadFile(caFile)
			if err != nil {
				return fmt.Errorf("datastore(%q): cannot read replicate_ca_file %q: %v", ds.namespace, caFile, err)
			}
			pool := x509.NewCertPool()
			if !pool.AppendCertsFromPEM(pem) {
				return fmt.Errorf("datastore(%q): replicate_ca_file %q contains no PEM certificates", ds.namespace, caFile)
			}
			repl.rootCAs = pool
		}
		repl.epoch = snapEpoch // the leader's welcome replaces this
	} else {
		repl.role = replRoleLeader
		repl.listenAddr = listen
		repl.certFile = certFile
		repl.keyFile = keyFile
		repl.ring = newReplRing(buffer)

		// The same warning from the other end. A leader listening on a public
		// interface without a certificate is handing out the store to anyone who
		// can sniff the link and read the secret out of a handshake.
		if certFile == "" {
			if host, _, err := net.SplitHostPort(listen); err != nil || !replIsLoopbackHost(host) {
				fmt.Fprintf(os.Stderr, "duso: warning: datastore %q serves replication on %s without TLS — replicate_secret and every value cross the network in the clear. Set replicate_cert_file and replicate_key_file unless this link is already private (VPN, WireGuard, internal VPC).\n",
					ds.namespace, listen)
			}
		}

		switch {
		case wasFollower:
			// Promotion. The new term is what stops a returning old leader, which
			// stayed at snapEpoch, from being mistaken for the current one.
			repl.epoch = snapEpoch + 1
			fmt.Fprintf(os.Stderr, "duso: datastore %q promoted to replication leader, epoch %d\n", ds.namespace, repl.epoch)
		case snapEpoch > 0:
			repl.epoch = snapEpoch // ordinary leader restart, same term
		default:
			repl.epoch = 1 // first time this store has replicated
		}
	}

	ds.repl = repl
	return nil
}

// replStart opens sockets. Called once recovery is complete.
func replStart(ds *DatastoreValue) error {
	repl := ds.repl
	if repl == nil {
		return nil
	}

	switch repl.role {
	case replRoleLeader:
		if err := replStartLeader(ds); err != nil {
			return err
		}
		repl.wg.Add(1)
		go repl.replHeartbeatLoop()
	case replRoleFollower:
		repl.cursor.Store(ds.walSeq.Load())
		repl.wg.Add(1)
		go ds.replFollowLoop()
	}
	return nil
}

// ---------------------------------------------------------------------------
// Ring buffer
// ---------------------------------------------------------------------------

type replEntry struct {
	seq   uint64
	frame []byte
}

// replRing is the leader's window of recent frames. Every follower reads this
// one buffer, so memory is bounded regardless of follower count and a slow
// follower falls off the front instead of inflating the leader.
type replRing struct {
	mu       sync.Mutex
	cond     *sync.Cond
	entries  []replEntry
	bytes    int
	maxBytes int
	firstSeq uint64 // seq of entries[0]; 0 when empty
	lastSeq  uint64 // seq of the newest entry
}

func newReplRing(maxBytes int) *replRing {
	if maxBytes <= 0 {
		maxBytes = replDefaultBuffer
	}
	r := &replRing{maxBytes: maxBytes}
	r.cond = sync.NewCond(&r.mu)
	return r
}

// append records a frame and wakes every waiting sender. The frame is retained
// by reference, so callers must not reuse the slice — writeWALOp builds a fresh
// one per record, which is what makes that safe.
func (r *replRing) append(seq uint64, frame []byte) {
	r.mu.Lock()
	r.entries = append(r.entries, replEntry{seq: seq, frame: frame})
	r.bytes += len(frame)
	r.lastSeq = seq
	if len(r.entries) == 1 {
		r.firstSeq = seq
	}
	// Always keep the newest entry, even if it alone exceeds the budget: a
	// single oversized value should cost a resync, not an empty ring.
	for r.bytes > r.maxBytes && len(r.entries) > 1 {
		r.bytes -= len(r.entries[0].frame)
		r.entries = r.entries[1:]
		r.firstSeq = r.entries[0].seq
	}
	r.mu.Unlock()
	r.cond.Broadcast()
}

// wake releases senders blocked in since() so they can re-check their own stop
// conditions. Called on shutdown and by the heartbeat.
func (r *replRing) wake() {
	r.cond.Broadcast()
}

// since blocks until frames after cursor exist. gone=true means the cursor fell
// off the front and the caller must resync from a snapshot. done() is polled on
// each wakeup so a disconnecting follower does not park here forever.
//
// maxBytes bounds one batch. next reports the seq of the last frame actually
// returned, which is not necessarily lastSeq: a caller behind by more than
// maxBytes gets the backlog over several calls, and its cursor must never run
// ahead of what it was handed.
func (r *replRing) since(cursor uint64, maxBytes int, done func() bool) (frames [][]byte, next uint64, gone bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	waited := false
	for {
		if done() {
			return nil, cursor, false
		}
		if len(r.entries) > 0 {
			// The cursor is behind the oldest frame we still hold. Note the
			// equality: firstSeq itself is fine, since a cursor of firstSeq-1
			// still finds firstSeq in the ring.
			if cursor+1 < r.firstSeq {
				return nil, cursor, true
			}
			if cursor < r.lastSeq {
				out := make([][]byte, 0, len(r.entries))
				last, total := cursor, 0
				for _, e := range r.entries {
					if e.seq <= cursor {
						continue
					}
					// The len(out) > 0 guard keeps a single frame larger than the
					// whole budget deliverable: it goes out alone rather than
					// being skipped, which would leave a hole the follower can
					// never fill. The payload limit is what makes it fit.
					if len(out) > 0 && total+len(e.frame) > maxBytes {
						break
					}
					out = append(out, e.frame)
					total += len(e.frame)
					last = e.seq
				}
				return out, last, false
			}
		}
		// Woken with nothing new. Return anyway: wake() is also how a queued
		// reply and the heartbeat ask to be sent, and only the caller knows it
		// has either. Waiting again here would strand both.
		if waited {
			return nil, cursor, false
		}
		r.cond.Wait()
		waited = true
	}
}

// ---------------------------------------------------------------------------
// Handshake
// ---------------------------------------------------------------------------

type replHello struct {
	Proto     int    `json:"proto"`
	Namespace string `json:"namespace"`
	Secret    string `json:"secret"`
	Cursor    uint64 `json:"cursor"`
	Epoch     uint32 `json:"epoch"`
	Encrypted bool   `json:"encrypted"`
}

type replWelcome struct {
	OK        bool   `json:"ok"`
	Error     string `json:"error,omitempty"`
	Mode      string `json:"mode,omitempty"` // "stream" or "snapshot"
	Epoch     uint32 `json:"epoch,omitempty"`
	Seq       uint64 `json:"seq,omitempty"`    // snapshot mode: watermark the snapshot covers
	Chunks    int    `json:"chunks,omitempty"` // snapshot mode: binary messages to follow
	Encrypted bool   `json:"encrypted"`
	CanWrite  bool   `json:"can_write"` // whether this follower may forward writes

	// PayloadLimit is the largest message this leader may send, derived from its
	// own max_value_size. The follower raises its own cap to match, so a leader
	// configured to hold larger values than the follower cannot produce a frame
	// the follower is unable to receive.
	PayloadLimit int `json:"payload_limit,omitempty"`
}

// ---------------------------------------------------------------------------
// Leader: publishing
// ---------------------------------------------------------------------------

// replPublish hands a frame to followers. Called from writeWALOp under
// dataMutex, so ring order is log order. One append and a broadcast; sockets are
// written elsewhere, so a wedged replica cannot wedge the leader.
func (ds *DatastoreValue) replPublish(seq uint64, frame []byte) {
	repl := ds.repl
	if repl == nil || repl.role != replRoleLeader || repl.ring == nil {
		return
	}
	repl.ring.append(seq, frame)
}

// ---------------------------------------------------------------------------
// Leader: listener
// ---------------------------------------------------------------------------

// replListener is one address serving one or more namespaces. Sharing a listener
// means a server replicating three datastores opens one port, not three.
type replListener struct {
	addr   string
	mu     sync.Mutex
	stores map[string]*DatastoreValue
	server *http.Server
	ln     net.Listener
}

var (
	replListenersMu sync.Mutex
	replListeners   = make(map[string]*replListener)
)

// replStartLeader registers the store with a listener for its address, starting
// the listener if this is the first store to ask for it.
func replStartLeader(ds *DatastoreValue) error {
	repl := ds.repl

	replListenersMu.Lock()
	defer replListenersMu.Unlock()

	l, exists := replListeners[repl.listenAddr]
	if exists {
		l.mu.Lock()
		defer l.mu.Unlock()
		if other, dup := l.stores[ds.namespace]; dup && other != ds {
			return fmt.Errorf("datastore(%q): another datastore is already replicating namespace %q on %s",
				ds.namespace, ds.namespace, repl.listenAddr)
		}
		l.stores[ds.namespace] = ds
		return nil
	}

	l = &replListener{addr: repl.listenAddr, stores: map[string]*DatastoreValue{ds.namespace: ds}}

	// Bind before returning so a busy port fails the datastore() call rather
	// than surfacing later as followers that mysteriously cannot connect.
	ln, err := net.Listen("tcp", repl.listenAddr)
	if err != nil {
		return fmt.Errorf("datastore(%q): cannot listen for replicas on %s: %v", ds.namespace, repl.listenAddr, err)
	}
	l.ln = ln

	mux := http.NewServeMux()
	mux.Handle("/replicate", websocket.Handler(l.serve))
	l.server = &http.Server{Handler: mux}

	replListeners[repl.listenAddr] = l

	go func() {
		defer core.RecoverPanic(fmt.Sprintf("datastore_replication_listener (addr=%s)", l.addr))
		var err error
		if repl.certFile != "" && repl.keyFile != "" {
			err = l.server.ServeTLS(ln, repl.certFile, repl.keyFile)
		} else {
			err = l.server.Serve(ln)
		}
		if err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "duso: datastore replication listener on %s stopped: %v\n", l.addr, err)
		}
	}()

	return nil
}

// lookup resolves a namespace to a store, and checks the secret in constant time.
func (l *replListener) lookup(hello *replHello) (ds *DatastoreValue, canWrite bool, err error) {
	l.mu.Lock()
	ds = l.stores[hello.Namespace]
	l.mu.Unlock()

	if ds == nil {
		return nil, false, fmt.Errorf("no datastore named %q is replicated on this server", hello.Namespace)
	}
	repl := ds.repl
	if repl == nil || repl.role != replRoleLeader {
		return nil, false, fmt.Errorf("datastore %q is not a replication leader", hello.Namespace)
	}
	// Both compares always run: returning early on the first match would leak,
	// by timing, which secret a guess was closer to.
	rw := subtle.ConstantTimeCompare([]byte(hello.Secret), []byte(repl.secret)) == 1
	ro := repl.readSecret != "" &&
		subtle.ConstantTimeCompare([]byte(hello.Secret), []byte(repl.readSecret)) == 1
	if !rw && !ro {
		return nil, false, fmt.Errorf("replicate_secret does not match")
	}
	return ds, rw, nil
}

// serve runs one follower connection start to finish: handshake, optional
// snapshot, then frames until someone hangs up.
func (l *replListener) serve(ws *websocket.Conn) {
	defer core.RecoverPanic(fmt.Sprintf("datastore_replication_serve (addr=%s)", l.addr))
	defer ws.Close()

	ws.MaxPayloadBytes = replHandshakeLimit

	var hello replHello
	_ = ws.SetReadDeadline(time.Now().Add(15 * time.Second))
	if err := websocket.JSON.Receive(ws, &hello); err != nil {
		return
	}
	_ = ws.SetReadDeadline(time.Time{})

	ds, canWrite, err := l.lookup(&hello)
	if err != nil {
		_ = websocket.JSON.Send(ws, replWelcome{OK: false, Error: err.Error()})
		return
	}
	repl := ds.repl

	// Authenticated: this peer is now entitled to send a forwarded write as large
	// as any value this store accepts, so the handshake cap comes off.
	payloadLimit := replPayloadLimit(ds.maxValueSize)
	ws.MaxPayloadBytes = payloadLimit

	if hello.Proto != replProtoVersion {
		_ = websocket.JSON.Send(ws, replWelcome{OK: false,
			Error: fmt.Sprintf("replication protocol version %d, this leader speaks %d", hello.Proto, replProtoVersion)})
		return
	}

	// Frames are shipped exactly as they are written, so an encrypted leader can
	// only be followed by a store holding the same key. Catching it here beats
	// letting the follower fail to decrypt every record it receives.
	leaderEncrypted := len(ds.encryptKey) > 0
	if hello.Encrypted != leaderEncrypted {
		msg := "this leader encrypts its log; the follower must set the same encrypt_key"
		if !leaderEncrypted {
			msg = "this leader does not encrypt its log; the follower must not set encrypt_key"
		}
		_ = websocket.JSON.Send(ws, replWelcome{OK: false, Error: msg})
		return
	}

	// A follower that has seen a later term than ours means this process has been
	// superseded — almost always an old leader that came back after a promotion.
	// Refusing here stops it feeding stale data to a replica that already moved on.
	if hello.Epoch > repl.epoch {
		msg := fmt.Sprintf("follower has seen epoch %d but this leader is epoch %d — this leader has been superseded and must be reconfigured as a follower",
			hello.Epoch, repl.epoch)
		_ = websocket.JSON.Send(ws, replWelcome{OK: false, Error: msg, Epoch: repl.epoch})
		fmt.Fprintf(os.Stderr, "duso: datastore %q: %s\n", ds.namespace, msg)
		return
	}

	done := make(chan struct{})
	defer close(done)
	isDone := func() bool {
		select {
		case <-done:
			return true
		case <-repl.stop:
			return true
		default:
			return false
		}
	}

	// Wake this sender when the store shuts down, so it does not sit in the ring
	// waiting for a frame that will never come.
	go func() {
		defer core.RecoverPanic("datastore_replication_waker")
		select {
		case <-done:
		case <-repl.stop:
			repl.ring.wake()
		}
	}()

	// Counted from here rather than from the top of serve: a connection that
	// failed authentication or the epoch check was never a follower of this
	// leader, and counting it would make a brute-force attempt look like healthy
	// replication.
	repl.followers.Add(1)
	defer repl.followers.Add(-1)

	// Forwarded writes arrive on their own goroutine; serve() stays the only
	// writer to this socket, so replies are queued here and shipped by the loop
	// below in the right order relative to the frames they refer to.
	state := &replConnState{canWrite: canWrite}
	go func() {
		defer core.RecoverPanic("datastore_replication_requests")
		l.replServeRequests(ws, ds, state, done)
	}()

	cursor := hello.Cursor

	// Decide between resuming mid-stream and shipping a snapshot. A follower
	// ahead of the leader is the post-promotion case: its extra records belong to
	// a timeline that no longer exists, and an operation log has no partial
	// rollback, so it starts over.
	needSnapshot := hello.Epoch != repl.epoch || cursor > ds.walSeq.Load()
	if !needSnapshot {
		repl.ring.mu.Lock()
		empty := len(repl.ring.entries) == 0
		firstSeq := repl.ring.firstSeq
		repl.ring.mu.Unlock()
		// An empty ring is only safe to resume from when the follower is already
		// current; otherwise the frames it is missing were written before this
		// leader started and exist nowhere in memory.
		if empty {
			needSnapshot = cursor != ds.walSeq.Load()
		} else {
			needSnapshot = cursor+1 < firstSeq
		}
	}

	if needSnapshot {
		// Worth a log line: a resync holds dataMutex for the length of the encode,
		// so an operator seeing write latency spikes needs to be able to connect
		// them to a replica reconnecting.
		fmt.Fprintf(os.Stderr, "duso: datastore %q: follower at seq %d needs a full resync (leader at seq %d, epoch %d, %s)\n",
			ds.namespace, hello.Cursor, ds.walSeq.Load(), repl.epoch, replGrantName(canWrite))

		snap, seq, err := ds.replSnapshot()
		if err != nil {
			_ = websocket.JSON.Send(ws, replWelcome{OK: false, Error: fmt.Sprintf("snapshot failed: %v", err)})
			return
		}
		chunks := (len(snap) + replSnapshotChunk - 1) / replSnapshotChunk
		if err := websocket.JSON.Send(ws, replWelcome{
			OK: true, Mode: "snapshot", Epoch: repl.epoch, Seq: seq, Chunks: chunks,
			Encrypted: leaderEncrypted, CanWrite: canWrite, PayloadLimit: payloadLimit,
		}); err != nil {
			return
		}
		for off := 0; off < len(snap); off += replSnapshotChunk {
			end := min(off+replSnapshotChunk, len(snap))
			if err := websocket.Message.Send(ws, snap[off:end]); err != nil {
				return
			}
		}
		cursor = seq
	} else {
		if err := websocket.JSON.Send(ws, replWelcome{
			OK: true, Mode: "stream", Epoch: repl.epoch,
			Encrypted: leaderEncrypted, CanWrite: canWrite, PayloadLimit: payloadLimit,
		}); err != nil {
			return
		}
		fmt.Fprintf(os.Stderr, "duso: datastore %q: follower resumed at seq %d (leader at seq %d, epoch %d, %s)\n",
			ds.namespace, cursor, ds.walSeq.Load(), repl.epoch, replGrantName(canWrite))
	}

	// Steady state. since() blocks until there is something to send, so an idle
	// leader costs one parked goroutine per follower and no syscalls; the
	// heartbeat below is what keeps a dead socket from going unnoticed.
	for {
		frames, next, gone := repl.ring.since(cursor, replMaxBatchBytes, isDone)
		if isDone() {
			return
		}
		if gone {
			// Falling off the ring is recoverable, just not from here: the
			// follower reconnects and takes a snapshot.
			_ = websocket.JSON.Send(ws, replWelcome{OK: false,
				Error: fmt.Sprintf("follower at seq %d has fallen behind the leader's replicate_buffer — reconnect for a snapshot", cursor)})
			return
		}
		if len(frames) == 0 {
			// Woken with nothing new. Rejected writes produce no frame, so they can
			// only leave from here; otherwise this is a heartbeat, which is what
			// surfaces a socket that has gone away.
			if err := replSendReplies(ws, state, cursor); err != nil {
				return
			}
			if err := websocket.Message.Send(ws, []byte{}); err != nil {
				return
			}
			continue
		}

		// Coalesce into one message. At high write rates this is the difference
		// between one syscall per record and one per wakeup.
		total := 0
		for _, f := range frames {
			total += len(f)
		}
		batch := make([]byte, 1, 1+total)
		batch[0] = replMsgFrames
		for _, f := range frames {
			batch = append(batch, f...)
		}
		if err := websocket.Message.Send(ws, batch); err != nil {
			return
		}
		cursor = next

		// Replies go out only after the frames they refer to, so a follower has
		// already applied its own write by the time it is handed the result.
		if err := replSendReplies(ws, state, cursor); err != nil {
			return
		}
	}
}

// replSnapshot encodes the store for a resyncing follower and reports the
// watermark it covers. Holds dataMutex for the encode, blocking writes — the
// cost saveToDisk already pays, now reachable by a reconnect.
func (ds *DatastoreValue) replSnapshot() ([]byte, uint64, error) {
	ds.dataMutex.RLock()
	defer ds.dataMutex.RUnlock()

	seq := ds.walSeq.Load()
	body, err := ds.encodeSnapshot()
	if err != nil {
		return nil, 0, err
	}
	return body, seq, nil
}

// replHeartbeatLoop nudges idle senders so they emit a keepalive. Without it a
// leader with no writes never touches the socket and cannot notice a dead peer.
func (repl *replState) replHeartbeatLoop() {
	defer core.RecoverPanic("datastore_replication_heartbeat")
	defer repl.wg.Done()

	t := time.NewTicker(replHeartbeat)
	defer t.Stop()
	for {
		select {
		case <-repl.stop:
			repl.ring.wake()
			return
		case <-t.C:
			repl.ring.wake()
		}
	}
}

// ---------------------------------------------------------------------------
// Follower: streaming
// ---------------------------------------------------------------------------

// replFollowLoop keeps the store connected to its leader, reconnecting with
// backoff. Failures are logged, never fatal: a stale replica still serves reads.
func (ds *DatastoreValue) replFollowLoop() {
	defer core.RecoverPanic(fmt.Sprintf("datastore_replication_follow (namespace=%s)", ds.namespace))
	defer ds.repl.wg.Done()

	repl := ds.repl
	backoff := replReconnectMin

	for {
		select {
		case <-repl.stop:
			return
		default:
		}

		streamed, err := ds.replSession()
		repl.connected.Store(false)

		select {
		case <-repl.stop:
			return
		default:
		}

		if err != nil {
			msg := err.Error()
			repl.lastError.Store(&msg)
			fmt.Fprintf(os.Stderr, "duso: datastore %q replication from %s: %v (retrying in %s)\n",
				ds.namespace, repl.leaderURL, err, backoff.Round(time.Millisecond))
		}

		// Report where this replica stopped, but only when a live stream ended —
		// which is exactly the moment a leader died, and exactly when someone
		// choosing which follower to promote needs this number. Logging it on
		// every failed dial would bury it; logging it periodically would be noise
		// for a healthy system that never fails over.
		if streamed {
			fmt.Fprintf(os.Stderr, "duso: datastore %q: replication stopped at seq %d (epoch %d) — this is the position to compare when choosing a replica to promote\n",
				ds.namespace, repl.cursor.Load(), repl.epoch)
		}

		// Back off only against a leader we cannot reach. A session that streamed
		// and then ended is an ordinary leader restart, and making the follower
		// wait out a backoff it grew during an unrelated outage days ago would
		// stretch a two-second gap into fifteen.
		if streamed {
			backoff = replReconnectMin
		}

		select {
		case <-repl.stop:
			return
		case <-time.After(backoff):
		}
		if backoff *= 2; backoff > replReconnectMax {
			backoff = replReconnectMax
		}
	}
}

// replSession runs one connection to the leader: dial, handshake, then apply
// frames until the connection ends. The bool reports whether the session ever
// reached the streaming state, which is what distinguishes a leader that
// restarted from one that cannot be reached or will not have us.
func (ds *DatastoreValue) replSession() (streamed bool, err error) {
	repl := ds.repl

	cfg, err := websocket.NewConfig(repl.leaderURL+"/replicate", repl.leaderURL)
	if err != nil {
		return false, fmt.Errorf("bad replicate_from URL: %v", err)
	}
	if repl.rootCAs != nil {
		cfg.TlsConfig = &tls.Config{RootCAs: repl.rootCAs}
	}
	ws, err := websocket.DialConfig(cfg)
	if err != nil {
		return false, ds.replExplainDialError(cfg, err)
	}
	defer ws.Close()
	ws.MaxPayloadBytes = replHandshakeLimit

	// Reads here block for up to replReadTimeout. Closing the socket on shutdown
	// is what turns that into an immediate return, so a stopping process does not
	// sit for half a minute waiting on a heartbeat that is not coming.
	sessionDone := make(chan struct{})
	defer close(sessionDone)
	go func() {
		defer core.RecoverPanic("datastore_replication_session_closer")
		select {
		case <-sessionDone:
		case <-repl.stop:
			_ = ws.Close()
		}
	}()

	if err := websocket.JSON.Send(ws, replHello{
		Proto:     replProtoVersion,
		Namespace: ds.namespace,
		Secret:    repl.secret,
		Cursor:    repl.cursor.Load(),
		Epoch:     repl.epoch,
		Encrypted: len(ds.encryptKey) > 0,
	}); err != nil {
		return false, fmt.Errorf("handshake failed: %v", err)
	}

	var welcome replWelcome
	_ = ws.SetReadDeadline(time.Now().Add(replReadTimeout))
	if err := websocket.JSON.Receive(ws, &welcome); err != nil {
		return false, fmt.Errorf("no handshake reply: %v", err)
	}
	if !welcome.OK {
		return false, fmt.Errorf("leader refused: %s", welcome.Error)
	}

	// Take the larger of the two limits. The leader's frames are sized by the
	// leader's max_value_size, so a follower configured to hold smaller values
	// than its leader must still be able to receive them — refusing here would
	// strand this replica at the record before the first oversized one, which is
	// the failure this limit exists to prevent.
	ws.MaxPayloadBytes = max(replPayloadLimit(ds.maxValueSize), welcome.PayloadLimit)

	if welcome.Mode == "snapshot" {
		if err := ds.replReceiveSnapshot(ws, &welcome); err != nil {
			return false, err
		}
	}

	repl.epoch = welcome.Epoch
	repl.canWrite.Store(welcome.CanWrite)
	repl.connected.Store(true)
	repl.lastError.Store(nil)

	// Publish the connection so script goroutines can forward writes over it, and
	// wake anyone still waiting on it when this session ends.
	conn := newReplConn(ws)
	repl.conn.Store(conn)
	defer func() {
		repl.conn.CompareAndSwap(conn, nil)
		conn.shutdown()
	}()

	for {
		select {
		case <-repl.stop:
			return true, nil
		default:
		}

		var msg []byte
		_ = ws.SetReadDeadline(time.Now().Add(replReadTimeout))
		if err := websocket.Message.Receive(ws, &msg); err != nil {
			return true, fmt.Errorf("stream ended: %v", err)
		}
		if len(msg) == 0 {
			continue // heartbeat
		}

		switch msg[0] {
		case replMsgFrames:
			if err := ds.replApplyBatch(msg[1:]); err != nil {
				return true, err
			}

		case replMsgReply:
			reply, err := decodeReplReply(msg[1:])
			if err != nil {
				return true, fmt.Errorf("malformed reply from leader: %v", err)
			}
			// Delivered from this goroutine, which has already applied every
			// frame that preceded this message — including the one this reply
			// refers to. That is what makes a forwarded write readable locally
			// the moment the call returns.
			conn.deliver(reply)

		default:
			return true, fmt.Errorf("unknown message type %#x from leader", msg[0])
		}
	}
}

// replExplainDialError names the fix for a ws:/wss: mismatch, since "bad status"
// says nothing about the cause. Runs only on the failure path.
func (ds *DatastoreValue) replExplainDialError(cfg *websocket.Config, err error) error {
	hostport := cfg.Location.Host
	if cfg.Location.Port() == "" {
		return err // no port to probe; nothing useful to add
	}

	switch cfg.Location.Scheme {
	case "ws":
		if replPeerSpeaksTLS(hostport) {
			return fmt.Errorf("%v — the leader is serving TLS; change replicate_from to wss://%s", err, hostport)
		}
	case "wss":
		// websocket.DialError predates error wrapping and has no Unwrap, so
		// errors.As cannot reach the TLS error inside it on its own.
		cause := err
		var dialErr *websocket.DialError
		if errors.As(err, &dialErr) {
			cause = dialErr.Err
		}
		var recordErr tls.RecordHeaderError
		if errors.As(cause, &recordErr) {
			return fmt.Errorf("%v — the leader is not serving TLS; either set replicate_cert_file and replicate_key_file on the leader, or change replicate_from to ws://%s", err, hostport)
		}
	}
	return err
}

// replReceiveSnapshot replaces local state wholesale. Replace, not merge: the
// frames that follow are an operation log applied on top of this exact state, so
// keys the leader does not have are drift, not extra data.
func (ds *DatastoreValue) replReceiveSnapshot(ws *websocket.Conn, welcome *replWelcome) error {
	buf := make([]byte, 0, replSnapshotChunk)
	for i := 0; i < welcome.Chunks; i++ {
		var chunk []byte
		_ = ws.SetReadDeadline(time.Now().Add(replReadTimeout))
		if err := websocket.Message.Receive(ws, &chunk); err != nil {
			return fmt.Errorf("snapshot transfer failed at chunk %d/%d: %v", i+1, welcome.Chunks, err)
		}
		buf = append(buf, chunk...)
	}

	// decodeSnapshot falls back to a gob decoder for pre-v1 files. That fallback
	// exists to read this machine's own old snapshots and must not be reachable
	// from a socket: encodeSnapshot only ever writes v1, so no real leader can
	// produce these bytes, and gob decoding of anything a peer influenced is a
	// cheap way to be made to allocate.
	if !bytes.HasPrefix(buf, []byte(snapshotMagic)) {
		return fmt.Errorf("leader sent a snapshot that is not in the v1 format")
	}

	data, expiry, seq, err := ds.decodeSnapshot(buf)
	if err != nil {
		return fmt.Errorf("leader snapshot: %v", err)
	}

	ds.dataMutex.Lock()
	ds.data = data
	ds.expiryTimes = make(map[string]time.Time, len(expiry))
	ds.expiryHeap = ds.expiryHeap[:0]
	ds.applySnapshotExpiry(expiry)
	ds.walSeq.Store(seq)
	// Anything blocked in wait() is looking at state that just changed entirely.
	for _, cond := range ds.conditions {
		cond.Broadcast()
	}
	ds.dataMutex.Unlock()

	ds.repl.cursor.Store(seq)
	ds.repl.epoch = welcome.Epoch

	// Persist immediately so a restart resumes from here instead of pulling the
	// whole snapshot again, and truncate the local log: every frame in it belongs
	// to a timeline this snapshot has just replaced.
	if ds.persistPath != "" {
		if err := ds.saveToDisk(); err != nil {
			return fmt.Errorf("failed to persist leader snapshot: %v", err)
		}
	} else if ds.walPath != "" {
		if err := ds.truncateWAL(); err != nil {
			return fmt.Errorf("failed to truncate WAL after resync: %v", err)
		}
	}

	return nil
}

// replApplyBatch splits a coalesced message into frames and applies each. Same
// framing as the log, but no torn tail exists here: a websocket message arrives
// whole or not at all, so a short or corrupt frame ends the session.
func (ds *DatastoreValue) replApplyBatch(batch []byte) error {
	repl := ds.repl

	for len(batch) > 0 {
		if len(batch) < walFrameLen {
			return fmt.Errorf("replication stream: %d trailing bytes, too short for a frame header", len(batch))
		}
		length := int(binary.LittleEndian.Uint32(batch[0:4]))
		wantCRC := binary.LittleEndian.Uint32(batch[4:8])
		if length > len(batch)-walFrameLen {
			return fmt.Errorf("replication stream: frame declares %d bytes but only %d remain", length, len(batch)-walFrameLen)
		}
		stored := batch[walFrameLen : walFrameLen+length]
		frame := batch[:walFrameLen+length]
		batch = batch[walFrameLen+length:]

		if got := crc32.Checksum(stored, crc32cTable); got != wantCRC {
			return fmt.Errorf("replication stream: frame failed its checksum (want %08x, got %08x)", wantCRC, got)
		}

		body := stored
		if len(ds.encryptKey) > 0 {
			plain, err := ds.decryptBytes(stored)
			if err != nil {
				return fmt.Errorf("replication stream: failed to decrypt frame: %v", err)
			}
			body = plain
		}

		rec, err := decodeWALBody(body)
		if err != nil {
			if errors.Is(err, errSkippableRecord) {
				// Control record from a newer leader. Skipping is safe by the
				// opcode reservation, but the seq still has to advance or the
				// cursor stalls and every reconnect resyncs.
				repl.cursor.Store(rec.Seq)
				ds.walSeq.Store(rec.Seq)
				continue
			}
			return fmt.Errorf("replication stream: %v", err)
		}

		cursor := repl.cursor.Load()
		if rec.Seq <= cursor {
			continue // already applied; a resumed stream can overlap by design
		}
		if rec.Seq != cursor+1 {
			// The log is an operation log. A hole means state has diverged and no
			// amount of further streaming closes it.
			return fmt.Errorf("replication stream: gap at seq %d (expected %d) — resyncing", rec.Seq, cursor+1)
		}

		if err := ds.replApplyRecord(rec, frame); err != nil {
			return err
		}
		repl.cursor.Store(rec.Seq)
	}
	return nil
}

// replApplyRecord applies one record and appends its frame verbatim to the
// follower's own log, so a restart resumes from its cursor and the ordinary
// recovery path reads it back without knowing replication produced it.
func (ds *DatastoreValue) replApplyRecord(rec *walRecord, frame []byte) error {
	ds.dataMutex.Lock()
	defer ds.dataMutex.Unlock()

	if ds.walFile != nil {
		ds.walMutex.Lock()
		_, err := ds.walFile.Write(frame)
		ds.walMutex.Unlock()
		if err != nil {
			return fmt.Errorf("failed to append replicated frame to WAL: %v", err)
		}
	}

	if err := ds.applyWALRecord(rec); err != nil {
		return fmt.Errorf("failed to apply replicated record at seq %d: %v", rec.Seq, err)
	}
	ds.walSeq.Store(rec.Seq)

	// Waiters on a follower are watching the leader's writes arrive.
	if cond, exists := ds.conditions[rec.Key]; exists {
		cond.Broadcast()
	}
	if rec.Key2 != "" {
		if cond, exists := ds.conditions[rec.Key2]; exists {
			cond.Broadcast()
		}
	}
	if rec.Op == opClear {
		for _, cond := range ds.conditions {
			cond.Broadcast()
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Lifecycle
// ---------------------------------------------------------------------------

// replShutdown stops replication and waits for its goroutines. Safe on a store
// that never replicated, and safe to call twice. The wait matters: Shutdown
// snapshots right after, and a frame applied in between would understate the
// watermark.
func (ds *DatastoreValue) replShutdown() {
	repl := ds.repl
	if repl == nil {
		return
	}

	repl.stopOnce.Do(func() {
		close(repl.stop)
		if repl.ring != nil {
			repl.ring.wake()
		}
		if repl.role == replRoleLeader {
			replStopLeader(ds)
		}
	})
	repl.wg.Wait()
}

// replStopLeader unregisters a store from its listener, closing the listener
// once the last store on that address has gone.
func replStopLeader(ds *DatastoreValue) {
	replListenersMu.Lock()
	defer replListenersMu.Unlock()

	l, ok := replListeners[ds.repl.listenAddr]
	if !ok {
		return
	}

	l.mu.Lock()
	delete(l.stores, ds.namespace)
	empty := len(l.stores) == 0
	l.mu.Unlock()

	if empty {
		delete(replListeners, ds.repl.listenAddr)
		_ = l.server.Close()
	}
}

// replStatus backs replication_status(). A standalone store answers rather than
// throwing, so one script can run on a leader, a follower, and a laptop.
func (ds *DatastoreValue) replStatus() map[string]any {
	repl := ds.repl
	if repl == nil {
		return map[string]any{"role": replRoleNone.String()}
	}

	out := map[string]any{
		"role":  repl.role.String(),
		"epoch": float64(repl.epoch),
	}

	switch repl.role {
	case replRoleLeader:
		out["listen"] = repl.listenAddr
		out["seq"] = float64(ds.walSeq.Load())
		out["followers"] = float64(repl.followers.Load())
		repl.ring.mu.Lock()
		out["buffered_bytes"] = float64(repl.ring.bytes)
		out["buffered_frames"] = float64(len(repl.ring.entries))
		repl.ring.mu.Unlock()

	case replRoleFollower:
		out["leader"] = repl.leaderURL
		out["connected"] = repl.connected.Load()
		out["can_write"] = repl.canWrite.Load()
		// The position to compare when choosing which replica to promote. There
		// is deliberately no "lag" field: a follower is never told the leader's
		// current sequence, so any lag it reported would be a guess.
		out["cursor"] = float64(repl.cursor.Load())
		if msg := repl.lastError.Load(); msg != nil {
			out["last_error"] = *msg
		}
	}

	return out
}

// replGrantName labels a connection's access in log lines.
func replGrantName(canWrite bool) string {
	if canWrite {
		return "read-write"
	}
	return "read-only"
}
