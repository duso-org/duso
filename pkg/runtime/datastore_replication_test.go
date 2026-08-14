package runtime

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

// These cover replication at two levels: the pure logic (ring buffer retention,
// frame parsing, the epoch bookkeeping that decides whether a start is a
// promotion), and a real leader and follower talking over a real localhost
// socket in TestReplicationOverSocket.
//
// Both are needed. The socket test proves the thing works end to end; the unit
// tests reach the cases a healthy connection never produces — a corrupt frame, a
// sequence gap, a follower that has fallen off the buffer.

func newReplTestStore(t *testing.T, role replRole) *DatastoreValue {
	t.Helper()
	ds := &DatastoreValue{
		namespace:   uniqueNamespace("repl"),
		data:        make(map[string]any),
		conditions:  make(map[string]*sync.Cond),
		expiryTimes: make(map[string]time.Time),
		expiryHeap:  make(ExpiryHeap, 0),
		repl:        &replState{role: role, secret: "s", epoch: 1, stop: make(chan struct{})},
	}
	if role == replRoleLeader {
		ds.repl.ring = newReplRing(1024)
	}
	return ds
}

// buildFrame produces the exact bytes writeWALOp would put on the wire, so the
// follower-side tests parse real frames rather than a test-only encoding.
func buildFrame(t *testing.T, rec *walRecord) []byte {
	t.Helper()
	body, err := encodeWALBody(rec)
	if err != nil {
		t.Fatalf("encodeWALBody: %v", err)
	}
	frame := make([]byte, walFrameLen, walFrameLen+len(body))
	binary.LittleEndian.PutUint32(frame[0:4], uint32(len(body)))
	binary.LittleEndian.PutUint32(frame[4:8], crc32.Checksum(body, crc32cTable))
	return append(frame, body...)
}

func TestReplRingEvictsOldestPastBudget(t *testing.T) {
	t.Parallel()

	r := newReplRing(100)
	for i := 1; i <= 10; i++ {
		r.append(uint64(i), make([]byte, 30))
	}

	r.mu.Lock()
	bytes, first, last, n := r.bytes, r.firstSeq, r.lastSeq, len(r.entries)
	r.mu.Unlock()

	if bytes > 100 {
		t.Errorf("ring holds %d bytes, over its 100-byte budget", bytes)
	}
	if last != 10 {
		t.Errorf("lastSeq = %d, want 10", last)
	}
	if first != uint64(10-n+1) {
		t.Errorf("firstSeq = %d with %d entries, want %d", first, n, 10-n+1)
	}
}

// A frame larger than the whole budget must still be retained. Dropping it would
// leave the ring empty and force a resync for every connected follower, which is
// a far worse outcome than briefly exceeding the budget.
func TestReplRingKeepsOversizedFrame(t *testing.T) {
	t.Parallel()

	r := newReplRing(100)
	r.append(1, make([]byte, 500))

	r.mu.Lock()
	n, first := len(r.entries), r.firstSeq
	r.mu.Unlock()

	if n != 1 || first != 1 {
		t.Fatalf("oversized frame was evicted: %d entries, firstSeq %d", n, first)
	}
}

func TestReplRingSince(t *testing.T) {
	t.Parallel()

	never := func() bool { return false }
	r := newReplRing(1 << 20)
	for i := 1; i <= 5; i++ {
		r.append(uint64(i), []byte{byte(i)})
	}

	frames, next, gone := r.since(2, never)
	if gone {
		t.Fatal("cursor 2 reported as fallen off a ring holding 1-5")
	}
	if next != 5 || len(frames) != 3 {
		t.Fatalf("since(2) = %d frames, next %d; want 3 frames, next 5", len(frames), next)
	}

	// A cursor one behind the oldest frame is still serviceable: the next record
	// it needs is exactly firstSeq, and that is still here.
	if _, _, gone := r.since(0, never); gone {
		t.Error("cursor 0 reported as fallen off a ring whose firstSeq is 1")
	}
}

func TestReplRingReportsFallenOffFollower(t *testing.T) {
	t.Parallel()

	r := newReplRing(50)
	for i := 1; i <= 20; i++ {
		r.append(uint64(i), make([]byte, 25))
	}

	// The budget holds ~2 frames, so a follower still at seq 1 is long gone.
	if _, _, gone := r.since(1, func() bool { return false }); !gone {
		t.Error("a follower behind the retained window was not reported as fallen off")
	}
}

func TestReplApplyBatchAppliesEveryFrame(t *testing.T) {
	t.Parallel()

	ds := newReplTestStore(t, replRoleFollower)
	batch := append(
		buildFrame(t, &walRecord{Seq: 1, Op: opSet, Key: "greeting", Value: "hello"}),
		buildFrame(t, &walRecord{Seq: 2, Op: opIncr, Key: "n", Num: 5})...,
	)
	batch = append(batch, buildFrame(t, &walRecord{Seq: 3, Op: opPush, Key: "q", Value: "item"})...)

	if err := ds.replApplyBatch(batch); err != nil {
		t.Fatalf("replApplyBatch: %v", err)
	}

	if got := ds.data["greeting"]; got != "hello" {
		t.Errorf("greeting = %v, want hello", got)
	}
	if got := ds.data["n"]; got != 5.0 {
		t.Errorf("n = %v, want 5", got)
	}
	if q, ok := ds.data["q"].([]any); !ok || len(q) != 1 {
		t.Errorf("q = %v, want a 1-element array", ds.data["q"])
	}
	if got := ds.repl.cursor.Load(); got != 3 {
		t.Errorf("cursor = %d, want 3", got)
	}
}

// A gap means the stream skipped a record. Because the log is an operation log,
// applying what follows would silently diverge — so the session must fail and
// the follower resync.
func TestReplApplyBatchRejectsSequenceGap(t *testing.T) {
	t.Parallel()

	ds := newReplTestStore(t, replRoleFollower)
	batch := append(
		buildFrame(t, &walRecord{Seq: 1, Op: opSet, Key: "a", Value: 1.0}),
		buildFrame(t, &walRecord{Seq: 3, Op: opSet, Key: "b", Value: 2.0})...,
	)

	err := ds.replApplyBatch(batch)
	if err == nil {
		t.Fatal("a batch skipping seq 2 was accepted")
	}
	if _, applied := ds.data["b"]; applied {
		t.Error("the record after the gap was applied anyway")
	}
}

// A resumed stream can legitimately re-send records the follower already has.
// Re-applying a PUSH would append the item twice, so replays must be dropped.
func TestReplApplyBatchSkipsAlreadyApplied(t *testing.T) {
	t.Parallel()

	ds := newReplTestStore(t, replRoleFollower)
	first := buildFrame(t, &walRecord{Seq: 1, Op: opPush, Key: "q", Value: "item"})

	if err := ds.replApplyBatch(first); err != nil {
		t.Fatalf("first apply: %v", err)
	}
	if err := ds.replApplyBatch(first); err != nil {
		t.Fatalf("replayed frame rejected: %v", err)
	}

	q, ok := ds.data["q"].([]any)
	if !ok || len(q) != 1 {
		t.Fatalf("q = %v, want a single item after a duplicate frame", ds.data["q"])
	}
}

func TestReplApplyBatchRejectsCorruption(t *testing.T) {
	t.Parallel()

	t.Run("bad checksum", func(t *testing.T) {
		ds := newReplTestStore(t, replRoleFollower)
		frame := buildFrame(t, &walRecord{Seq: 1, Op: opSet, Key: "a", Value: "v"})
		frame[len(frame)-1] ^= 0xFF
		if err := ds.replApplyBatch(frame); err == nil {
			t.Error("a frame failing its checksum was accepted")
		}
	})

	t.Run("truncated body", func(t *testing.T) {
		ds := newReplTestStore(t, replRoleFollower)
		frame := buildFrame(t, &walRecord{Seq: 1, Op: opSet, Key: "a", Value: "v"})
		if err := ds.replApplyBatch(frame[:len(frame)-3]); err == nil {
			t.Error("a frame declaring more bytes than it carries was accepted")
		}
	})

	t.Run("trailing garbage", func(t *testing.T) {
		ds := newReplTestStore(t, replRoleFollower)
		frame := buildFrame(t, &walRecord{Seq: 1, Op: opSet, Key: "a", Value: "v"})
		if err := ds.replApplyBatch(append(frame, 0x01, 0x02)); err == nil {
			t.Error("trailing bytes too short for a frame header were accepted")
		}
	})
}

func TestFollowerRejectsWrites(t *testing.T) {
	t.Parallel()

	ds := newReplTestStore(t, replRoleFollower)
	ds.repl.leaderURL = "ws://leader:7777"

	err := ds.writeGuard()
	if err == nil {
		t.Fatal("a follower accepted a write")
	}
	// The message has to name the leader, or an operator staring at it has no
	// idea where the write was supposed to go.
	if !strings.Contains(err.Error(), "ws://leader:7777") {
		t.Errorf("write rejection does not name the leader: %v", err)
	}

	leader := newReplTestStore(t, replRoleLeader)
	if err := leader.writeGuard(); err != nil {
		t.Errorf("a leader rejected a write: %v", err)
	}
}

// A follower must never run its own expiry sweep: it would delete keys the
// leader still holds and, worse, call writeWALOp and advance a sequence counter
// that belongs to the leader.
func TestFollowerDoesNotSweepExpiredKeys(t *testing.T) {
	t.Parallel()

	ds := newReplTestStore(t, replRoleFollower)
	ds.data["doomed"] = "value"
	past := time.Now().Add(-time.Hour)
	ds.expiryTimes["doomed"] = past
	ds.expiryHeap = ExpiryHeap{{key: "doomed", expiryTime: past}}

	ds.sweepExpiredKeys()

	if _, exists := ds.data["doomed"]; !exists {
		t.Error("a follower swept an expired key instead of waiting for the leader's EXPIRED record")
	}
	if got := ds.walSeq.Load(); got != 0 {
		t.Errorf("a follower advanced walSeq to %d on its own", got)
	}
}

// The follower bit in the snapshot header is what separates a promotion from an
// ordinary leader restart. Get it wrong and a returning old leader lands on the
// same epoch as the replica that replaced it.
func TestSnapshotProvenanceDrivesEpoch(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	write := func(name string, role replRole, epoch uint32) string {
		path := filepath.Join(dir, name)
		ds := newTestStore(nil)
		ds.persistPath = path
		ds.repl = &replState{role: role, epoch: epoch}
		ds.data = map[string]any{"k": "v"}
		body, err := ds.encodeSnapshot()
		if err != nil {
			t.Fatalf("encodeSnapshot: %v", err)
		}
		if err := os.WriteFile(path, body, 0644); err != nil {
			t.Fatalf("write snapshot: %v", err)
		}
		return path
	}

	tests := []struct {
		name       string
		role       replRole
		epoch      uint32
		wantEpoch  uint32
		wantFollow bool
		promotedTo uint32
	}{
		{"leader snapshot", replRoleLeader, 4, 4, false, 4},
		{"follower snapshot", replRoleFollower, 4, 4, true, 5},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			path := write(tc.name+".dusnap", tc.role, tc.epoch)

			epoch, wasFollower := readSnapshotProvenance(path)
			if epoch != tc.wantEpoch || wasFollower != tc.wantFollow {
				t.Fatalf("provenance = (epoch %d, follower %v), want (%d, %v)",
					epoch, wasFollower, tc.wantEpoch, tc.wantFollow)
			}

			// Starting as a leader from this file: promotion bumps the term, a
			// plain restart keeps it.
			ds := newTestStore(nil)
			ds.namespace = uniqueNamespace("promote")
			ds.persistPath = path
			cfg := map[string]any{"replicate_listen": "127.0.0.1:0", "replicate_secret": "s"}
			if err := replConfigure(ds, cfg); err != nil {
				t.Fatalf("replConfigure: %v", err)
			}
			if got := ds.repl.epoch; got != tc.promotedTo {
				t.Errorf("epoch after starting as leader = %d, want %d", got, tc.promotedTo)
			}
		})
	}
}

// A file written before replication existed carries zeros where the epoch and
// follower bit now live, and must read back as "never replicated" rather than as
// a follower at epoch 0.
func TestSnapshotProvenanceOnUnreplicatedFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "plain.dusnap")
	ds := newTestStore(nil)
	ds.persistPath = path
	ds.data = map[string]any{"k": "v"}
	body, err := ds.encodeSnapshot()
	if err != nil {
		t.Fatalf("encodeSnapshot: %v", err)
	}
	if err := os.WriteFile(path, body, 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	epoch, wasFollower := readSnapshotProvenance(path)
	if epoch != 0 || wasFollower {
		t.Errorf("provenance = (epoch %d, follower %v), want (0, false)", epoch, wasFollower)
	}

	// Missing files answer the same way, which is what makes a first run work.
	if epoch, wasFollower := readSnapshotProvenance(filepath.Join(t.TempDir(), "absent")); epoch != 0 || wasFollower {
		t.Errorf("missing file = (epoch %d, follower %v), want (0, false)", epoch, wasFollower)
	}
}

func TestReplConfigureRejectsBadConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cfg  map[string]any
		want string
	}{
		{"both roles",
			map[string]any{"replicate_listen": ":1", "replicate_from": "ws://x", "replicate_secret": "s"},
			"not both"},
		{"no secret",
			map[string]any{"replicate_listen": ":1"},
			"replicate_secret"},
		{"secret with no role",
			map[string]any{"replicate_secret": "s"},
			"replicate_listen"},
		{"buffer with no role",
			map[string]any{"replicate_buffer": 1024.0},
			"replicate_listen"},
		{"cert without key",
			map[string]any{"replicate_listen": ":1", "replicate_secret": "s", "replicate_cert_file": "c.pem"},
			"must be set together"},
		{"cert on a follower",
			map[string]any{"replicate_from": "ws://x", "replicate_secret": "s",
				"replicate_cert_file": "c.pem", "replicate_key_file": "k.pem"},
			"wss://"},
		{"ca on a leader",
			map[string]any{"replicate_listen": ":1", "replicate_secret": "s", "replicate_ca_file": "ca.pem"},
			"replicate_cert_file"},
		{"unreadable ca file",
			map[string]any{"replicate_from": "wss://x", "replicate_secret": "s",
				"replicate_ca_file": "/nonexistent/ca.pem"},
			"cannot read replicate_ca_file"},
		{"non-string listen",
			map[string]any{"replicate_listen": 7777.0, "replicate_secret": "s"},
			"must be a string"},
		{"negative buffer",
			map[string]any{"replicate_listen": ":1", "replicate_secret": "s", "replicate_buffer": -1.0},
			"must be positive"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ds := newReplTestStore(t, replRoleNone)
			ds.repl = nil
			err := replConfigure(ds, tc.cfg)
			if err == nil {
				t.Fatalf("config %v was accepted", tc.cfg)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error %q does not mention %q", err, tc.want)
			}
		})
	}
}

func TestReplConfigureStandaloneStaysNil(t *testing.T) {
	t.Parallel()

	ds := newReplTestStore(t, replRoleNone)
	ds.repl = nil
	if err := replConfigure(ds, map[string]any{"persist": "x.dusnap"}); err != nil {
		t.Fatalf("replConfigure on a plain store: %v", err)
	}
	if ds.repl != nil {
		t.Error("a store with no replicate_* options was given replication state")
	}
	if ds.replEpoch() != 0 {
		t.Errorf("standalone store stamps epoch %d, want 0", ds.replEpoch())
	}
}

// replBoundAddr reports the address a leader's listener actually bound, so tests
// can ask for port 0 and still tell a follower where to dial.
func replBoundAddr(configured string) string {
	replListenersMu.Lock()
	defer replListenersMu.Unlock()
	if l, ok := replListeners[configured]; ok && l.ln != nil {
		return l.ln.Addr().String()
	}
	return configured
}

// TestReplicationOverSocket drives a real leader and a real follower over a real
// socket, exercising the paths that only exist once bytes are moving: snapshot
// resync, live streaming, and the operation-log semantics that make a streamed
// PUSH land exactly once.
func TestReplicationOverSocket(t *testing.T) {
	dir := t.TempDir()

	newStore := func(name string) *DatastoreValue {
		return &DatastoreValue{
			namespace:       "socket_e2e", // both sides name the same logical store
			data:            make(map[string]any),
			conditions:      make(map[string]*sync.Cond),
			expiryTimes:     make(map[string]time.Time),
			expiryHeap:      make(ExpiryHeap, 0),
			persistPath:     filepath.Join(dir, name+".dusnap"),
			walPath:         filepath.Join(dir, name+".duwal"),
			walSyncInterval: 10 * time.Millisecond,
			maxValueSize:    defaultMaxValueSize,
		}
	}

	leader := newStore("leader")
	if err := replConfigure(leader, map[string]any{
		"replicate_listen": "127.0.0.1:0",
		"replicate_secret": "shared",
	}); err != nil {
		t.Fatalf("configure leader: %v", err)
	}
	if err := leader.openWALForWrites(); err != nil {
		t.Fatalf("open leader WAL: %v", err)
	}

	// Seed before the follower exists, so its first connection must resync from
	// a snapshot rather than stream from zero.
	for i := 0; i < 50; i++ {
		if err := leader.Set(fmt.Sprintf("seed%d", i), float64(i)); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}

	if err := replStart(leader); err != nil {
		t.Fatalf("start leader: %v", err)
	}
	defer leader.replShutdown()

	follower := newStore("follower")
	if err := replConfigure(follower, map[string]any{
		"replicate_from":   "ws://" + replBoundAddr("127.0.0.1:0"),
		"replicate_secret": "shared",
	}); err != nil {
		t.Fatalf("configure follower: %v", err)
	}
	if err := follower.openWALForWrites(); err != nil {
		t.Fatalf("open follower WAL: %v", err)
	}
	if err := replStart(follower); err != nil {
		t.Fatalf("start follower: %v", err)
	}
	defer follower.replShutdown()

	waitForCursor := func(want uint64) {
		t.Helper()
		deadline := time.Now().Add(10 * time.Second)
		for time.Now().Before(deadline) {
			if follower.repl.cursor.Load() >= want {
				return
			}
			time.Sleep(5 * time.Millisecond)
		}
		t.Fatalf("follower stalled at seq %d, waiting for %d", follower.repl.cursor.Load(), want)
	}

	waitForCursor(leader.walSeq.Load())

	// Now drive live traffic through every op shape that behaves differently on
	// replay: a blind write, a relative counter, an array append, a rename and a
	// delete.
	for i := 0; i < 25; i++ {
		if _, err := leader.Push("queue", fmt.Sprintf("item-%d", i)); err != nil {
			t.Fatalf("push: %v", err)
		}
		if _, err := leader.Increment("counter", 2); err != nil {
			t.Fatalf("increment: %v", err)
		}
	}
	if err := leader.Set("mutable", "final"); err != nil {
		t.Fatalf("set: %v", err)
	}
	if err := leader.Rename("seed0", "renamed"); err != nil {
		t.Fatalf("rename: %v", err)
	}
	if _, err := leader.Delete("seed1"); err != nil {
		t.Fatalf("delete: %v", err)
	}

	waitForCursor(leader.walSeq.Load())

	leader.dataMutex.RLock()
	want := DeepCopyAny(leader.data)
	leader.dataMutex.RUnlock()

	follower.dataMutex.RLock()
	got := DeepCopyAny(follower.data)
	follower.dataMutex.RUnlock()

	if !reflect.DeepEqual(want, got) {
		wantMap := want.(map[string]any)
		gotMap := got.(map[string]any)
		t.Errorf("follower diverged: leader has %d keys, follower %d", len(wantMap), len(gotMap))
		for k, v := range wantMap {
			if gv, ok := gotMap[k]; !ok {
				t.Errorf("  follower is missing %q", k)
			} else if !reflect.DeepEqual(v, gv) {
				t.Errorf("  %q: leader %v, follower %v", k, v, gv)
			}
		}
	}

	// The counter and queue are the ones that catch a double-apply: a streamed
	// INCR applied twice reads 100, and a PUSH applied twice leaves 50 items.
	if got := got.(map[string]any)["counter"]; got != 50.0 {
		t.Errorf("counter = %v, want 50 — a relative op was applied the wrong number of times", got)
	}
	if q, ok := got.(map[string]any)["queue"].([]any); !ok || len(q) != 25 {
		t.Errorf("queue = %v, want 25 items — a streamed PUSH landed the wrong number of times", got.(map[string]any)["queue"])
	}

	if err := follower.writeGuard(); err == nil {
		t.Error("the follower accepted a write while streaming")
	}
}

func TestReplIsLoopbackHost(t *testing.T) {
	t.Parallel()

	tests := []struct {
		host string
		want bool
	}{
		{"127.0.0.1", true},
		{"::1", true},
		{"localhost", true},
		{"127.0.0.53", true},
		{"", false}, // bare ":7777" — every interface
		{"0.0.0.0", false},
		{"10.0.0.5", false},
		{"db1.internal", false}, // resolving would make a warning depend on DNS
	}

	for _, tc := range tests {
		if got := replIsLoopbackHost(tc.host); got != tc.want {
			t.Errorf("replIsLoopbackHost(%q) = %v, want %v", tc.host, got, tc.want)
		}
	}
}

func TestReplConfigureRejectsBadURLScheme(t *testing.T) {
	t.Parallel()

	for _, from := range []string{"http://db1:7777", "db1:7777", "tcp://db1:7777"} {
		ds := newReplTestStore(t, replRoleNone)
		ds.repl = nil
		err := replConfigure(ds, map[string]any{
			"replicate_from":   from,
			"replicate_secret": "s",
		})
		if err == nil {
			t.Errorf("replicate_from %q was accepted", from)
			continue
		}
		if !strings.Contains(err.Error(), "ws:// or wss://") {
			t.Errorf("replicate_from %q: error %q does not name the valid schemes", from, err)
		}
	}
}
