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

	frames, next, gone := r.since(2, replMaxBatchBytes, never)
	if gone {
		t.Fatal("cursor 2 reported as fallen off a ring holding 1-5")
	}
	if next != 5 || len(frames) != 3 {
		t.Fatalf("since(2) = %d frames, next %d; want 3 frames, next 5", len(frames), next)
	}

	// A cursor one behind the oldest frame is still serviceable: the next record
	// it needs is exactly firstSeq, and that is still here.
	if _, _, gone := r.since(0, replMaxBatchBytes, never); gone {
		t.Error("cursor 0 reported as fallen off a ring whose firstSeq is 1")
	}
}

// The frames since() hands back are applied in the order they are returned, and
// the follower rejects a sequence gap — so an out-of-order or incomplete run
// ends the session. Pinned here because it is an invariant of the return value,
// not of the loop that happens to produce it today.
func TestReplRingSinceReturnsFramesInSeqOrder(t *testing.T) {
	t.Parallel()

	r := newReplRing(1 << 20)
	for i := 1; i <= 8; i++ {
		r.append(uint64(i), []byte{byte(i)})
	}

	frames, next, gone := r.since(3, replMaxBatchBytes, func() bool { return false })
	if gone {
		t.Fatal("cursor 3 reported as fallen off a ring holding 1-8")
	}
	if len(frames) != 5 {
		t.Fatalf("since(3) returned %d frames, want 5 (seqs 4-8)", len(frames))
	}
	for i, f := range frames {
		if wantSeq := byte(4 + i); f[0] != wantSeq {
			t.Errorf("frame %d carries seq %d, want %d — frames must come back in seq order", i, f[0], wantSeq)
		}
	}
	// next is the seq of the last frame handed back: it becomes the caller's new
	// cursor, so it must never run ahead of what was actually sent.
	if next != uint64(3+len(frames)) {
		t.Errorf("next = %d after returning %d frames from cursor 3, want %d", next, len(frames), 3+len(frames))
	}
}

// since() must not park a second time once it has been woken. wake() is also how
// a queued reply and the heartbeat ask to be sent, and only the caller knows it
// has either — waiting again here would strand both.
func TestReplRingSinceReturnsAfterWakeWithNothingNew(t *testing.T) {
	t.Parallel()

	r := newReplRing(1 << 20)
	for i := 1; i <= 3; i++ {
		r.append(uint64(i), []byte{byte(i)})
	}

	type result struct {
		frames [][]byte
		next   uint64
		gone   bool
	}
	done := make(chan result, 1)
	go func() {
		frames, next, gone := r.since(3, replMaxBatchBytes, func() bool { return false })
		done <- result{frames, next, gone}
	}()

	// Give the caller time to park, then wake it with no new frames.
	time.Sleep(50 * time.Millisecond)
	r.wake()

	select {
	case got := <-done:
		if len(got.frames) != 0 || got.gone {
			t.Errorf("since() returned %d frames, gone=%v; want 0 frames and gone=false", len(got.frames), got.gone)
		}
		if got.next != 3 {
			t.Errorf("next = %d, want the cursor back unchanged (3)", got.next)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("since() parked again after wake() — a queued reply or heartbeat would never be sent")
	}
}

// A disconnecting follower must not sit in the ring waiting for a frame that
// will never come, even when frames are available to return.
func TestReplRingSinceHonorsDone(t *testing.T) {
	t.Parallel()

	r := newReplRing(1 << 20)
	for i := 1; i <= 5; i++ {
		r.append(uint64(i), []byte{byte(i)})
	}

	frames, next, gone := r.since(0, replMaxBatchBytes, func() bool { return true })
	if len(frames) != 0 || gone {
		t.Errorf("since() with done()==true returned %d frames, gone=%v; want 0 frames and gone=false",
			len(frames), gone)
	}
	if next != 0 {
		t.Errorf("next = %d, want the cursor back unchanged (0)", next)
	}
}

// serve() concatenates everything since() hands back into one websocket message,
// so an unbounded batch is an unbounded message. A follower resuming after a
// blip is handed every frame it missed — built entirely out of small values —
// and before the batch was bounded that message went over the payload cap, the
// follower reconnected at the same cursor, and the leader sent the same
// oversized batch again, forever.
func TestReplRingSinceBoundsBatchSize(t *testing.T) {
	t.Parallel()

	// 8MB of ordinary small records: 1000 frames of 8KB. Nothing here is a large
	// value; this is a write burst a follower fell behind on.
	r := newReplRing(64 * 1024 * 1024)
	for i := 1; i <= 1000; i++ {
		r.append(uint64(i), make([]byte, 8*1024))
	}

	frames, next, gone := r.since(0, replMaxBatchBytes, func() bool { return false })
	if gone {
		t.Fatal("cursor 0 fell off a 64MB ring holding 8MB")
	}

	total := 0
	for _, f := range frames {
		total += len(f)
	}
	if total > replMaxBatchBytes {
		t.Errorf("since() returned %d bytes in one batch (%d frames), over the %d-byte budget; "+
			"serve() builds one websocket message this size", total, len(frames), replMaxBatchBytes)
	}

	// The budget must not stall the stream: a batch has to carry something.
	if len(frames) == 0 {
		t.Fatal("since() returned no frames from a ring holding 1000 of them")
	}

	// next must name the last frame actually included, so the caller's cursor
	// never runs ahead of what it sent.
	if next != uint64(len(frames)) {
		t.Errorf("next = %d after returning %d frames from cursor 0, want %d — "+
			"a partial batch must report the seq it actually reached",
			next, len(frames), len(frames))
	}

	// The backlog drains across successive calls rather than being dropped.
	drained := len(frames)
	for drained < 1000 {
		var more [][]byte
		more, next, gone = r.since(next, replMaxBatchBytes, func() bool { return false })
		if gone || len(more) == 0 {
			t.Fatalf("draining stalled at seq %d after %d frames (gone=%v)", next, drained, gone)
		}
		drained += len(more)
	}
	if drained != 1000 {
		t.Errorf("drained %d frames in total, want 1000", drained)
	}
}

func TestReplRingReportsFallenOffFollower(t *testing.T) {
	t.Parallel()

	r := newReplRing(50)
	for i := 1; i <= 20; i++ {
		r.append(uint64(i), make([]byte, 25))
	}

	// The budget holds ~2 frames, so a follower still at seq 1 is long gone.
	if _, _, gone := r.since(1, replMaxBatchBytes, func() bool { return false }); !gone {
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

// A follower forwards writes rather than rejecting them, so writeGuard must let
// them through — but save() and load() act on local files and are never
// forwarded, and a disconnected follower has to say plainly that nothing was
// written.
func TestFollowerWriteGuards(t *testing.T) {
	t.Parallel()

	ds := newReplTestStore(t, replRoleFollower)
	ds.repl.leaderURL = "ws://leader:7777"

	if err := ds.writeGuard(); err != nil {
		t.Errorf("a follower rejected a forwardable write: %v", err)
	}
	for _, op := range []string{"save", "load"} {
		if err := ds.fileGuard(op); err == nil {
			t.Errorf("a follower accepted %s()", op)
		}
	}

	// No connection: the caller must be told the write did not happen, and where
	// it was supposed to go.
	err := ds.Set("k", "v")
	if err == nil {
		t.Fatal("a disconnected follower accepted a write")
	}
	if !strings.Contains(err.Error(), "ws://leader:7777") {
		t.Errorf("error does not name the leader: %v", err)
	}
	if !strings.Contains(err.Error(), "not applied") {
		t.Errorf("error does not say the write did not land: %v", err)
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
		{"readonly secret on a follower",
			map[string]any{"replicate_from": "ws://x", "replicate_secret": "s",
				"replicate_readonly_secret": "r"},
			"leader option"},
		{"readonly secret same as the write secret",
			map[string]any{"replicate_listen": ":1", "replicate_secret": "s",
				"replicate_readonly_secret": "s"},
			"must differ"},
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
		{"cert reload interval with no certificate",
			map[string]any{"replicate_listen": ":1", "replicate_secret": "s",
				"replicate_cert_reload_interval": 60.0},
			"replicate_cert_file"},
		{"cert reload interval with no role",
			map[string]any{"replicate_cert_reload_interval": 60.0},
			"replicate_listen"},
		{"negative cert reload interval",
			map[string]any{"replicate_listen": ":1", "replicate_secret": "s",
				"replicate_cert_file": "c.pem", "replicate_key_file": "k.pem",
				"replicate_cert_reload_interval": -1.0},
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

	// Forwarded writes: issued on the follower, applied on the leader, and
	// visible locally the moment the call returns.
	n, err := follower.Increment("counter", 8)
	if err != nil {
		t.Fatalf("forwarded increment: %v", err)
	}
	if n != 58.0 {
		t.Errorf("forwarded increment returned %v, want 58 — the leader computes the result", n)
	}
	if local, _ := follower.Get("counter"); local != 58.0 {
		t.Errorf("counter reads %v locally right after the forwarded write returned, want 58", local)
	}
	if onLeader, _ := leader.Get("counter"); onLeader != 58.0 {
		t.Errorf("leader has counter = %v after a forwarded write, want 58", onLeader)
	}

	// A no-op forward still has to come back: set_once on an existing key
	// produces no WAL record, so its reply cannot ride behind a frame.
	stored, err := follower.SetOnce("mutable", "ignored")
	if err != nil {
		t.Fatalf("forwarded set_once: %v", err)
	}
	if stored {
		t.Error("set_once reported storing over an existing key")
	}

	// save() and load() are local file operations and stay blocked.
	if err := follower.fileGuard("load"); err == nil {
		t.Error("a streaming follower accepted load()")
	}
}

// newReplTestPair spins a leader and a follower over a real localhost socket and
// waits for the follower's first sync. Both sides name the same logical store,
// which is why each pair needs a namespace of its own.
// maxValueSize is explicit because it sets the connection's payload limit, and
// the relationship between that limit and replicate_buffer is what these tests
// are about.
func newReplTestPair(t *testing.T, namespace string, maxValueSize int64) (leader, follower *DatastoreValue) {
	t.Helper()

	dir := t.TempDir()
	newStore := func(name string) *DatastoreValue {
		return &DatastoreValue{
			namespace:       namespace,
			data:            make(map[string]any),
			conditions:      make(map[string]*sync.Cond),
			expiryTimes:     make(map[string]time.Time),
			expiryHeap:      make(ExpiryHeap, 0),
			persistPath:     filepath.Join(dir, name+".dusnap"),
			walPath:         filepath.Join(dir, name+".duwal"),
			walSyncInterval: 10 * time.Millisecond,
			maxValueSize:    maxValueSize,
		}
	}

	leader = newStore("leader")
	if err := replConfigure(leader, map[string]any{
		"replicate_listen": "127.0.0.1:0",
		"replicate_secret": "shared",
	}); err != nil {
		t.Fatalf("configure leader: %v", err)
	}
	if err := leader.openWALForWrites(); err != nil {
		t.Fatalf("open leader WAL: %v", err)
	}
	if err := replStart(leader); err != nil {
		t.Fatalf("start leader: %v", err)
	}
	t.Cleanup(leader.replShutdown)

	follower = newStore("follower")
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
	t.Cleanup(follower.replShutdown)

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) && follower.repl.conn.Load() == nil {
		time.Sleep(5 * time.Millisecond)
	}
	if follower.repl.conn.Load() == nil {
		t.Fatal("follower never connected to the leader")
	}
	return leader, follower
}

// replWaitForCursor reports whether the follower caught up to want before the
// deadline. It returns rather than failing so callers can attribute the stall.
func replWaitForCursor(follower *DatastoreValue, want uint64, within time.Duration) bool {
	deadline := time.Now().Add(within)
	for time.Now().Before(deadline) {
		if follower.repl.cursor.Load() >= want {
			return true
		}
		time.Sleep(5 * time.Millisecond)
	}
	return false
}

// RED until the 1.2 fix lands.
//
// The user-visible half of the same bug TestReplRingSinceBoundsBatchSize pins at
// the ring: a value under maxValueSize but over the websocket payload cap can
// never reach a follower, and the follower does not degrade — it reconnects at
// the same cursor every replReconnectMin forever, because replFollowLoop resets
// the backoff whenever a session streamed.
func TestReplicationDeliversValueOverPayloadCap(t *testing.T) {
	leader, follower := newReplTestPair(t, "payload_cap_e2e", defaultMaxValueSize)

	// Baseline: ordinary streaming works, so a failure below is about size.
	if err := leader.Set("small", "ok"); err != nil {
		t.Fatalf("leader set: %v", err)
	}
	if !replWaitForCursor(follower, leader.walSeq.Load(), 10*time.Second) {
		t.Fatal("baseline streaming is broken; the rest of this test proves nothing")
	}

	// Comfortably under the 64MB maxValueSize, comfortably over the 4MB cap.
	big := strings.Repeat("x", 6*1024*1024)
	if err := leader.Set("big", big); err != nil {
		t.Fatalf("leader refused a 6MB value: %v", err)
	}

	if !replWaitForCursor(follower, leader.walSeq.Load(), 3*time.Second) {
		last := "<none>"
		if p := follower.repl.lastError.Load(); p != nil {
			last = *p
		}
		t.Errorf("follower stalled at seq %d, leader at seq %d, last error: %s "+
			"(expected until the 1.2 fix — see docs/ideas/replication-hardening-plan.md)",
			follower.repl.cursor.Load(), leader.walSeq.Load(), last)
		return
	}

	follower.dataMutex.RLock()
	got, ok := follower.data["big"].(string)
	follower.dataMutex.RUnlock()
	if !ok || len(got) != len(big) {
		t.Errorf("follower has big = %d bytes, want %d", len(got), len(big))
	}
}

// The socket-level case for the bounded batch, and the one that reaches ordinary
// workloads: no value here is remotely large. A follower drops its connection,
// the leader writes several megabytes of small records while it is away, and the
// follower resumes from inside the ring — so the leader streams the backlog
// rather than sending a snapshot. Unbounded, that backlog was one websocket
// message over the payload cap, and the follower could never get past it.
func TestReplicationResumesFromLargeBacklogOfSmallWrites(t *testing.T) {
	// A store holding small values gets a small payload limit, while the ring
	// still defaults to 64MB. That gap is the point: the batch bound, not the
	// payload limit, is what keeps the backlog deliverable here.
	leader, follower := newReplTestPair(t, "backlog_e2e", 1024*1024)

	if err := leader.Set("first", "ok"); err != nil {
		t.Fatalf("leader set: %v", err)
	}
	if !replWaitForCursor(follower, leader.walSeq.Load(), 10*time.Second) {
		t.Fatal("baseline streaming is broken; the rest of this test proves nothing")
	}
	resumeFrom := follower.repl.cursor.Load()

	// Drop the connection out from under the follower. It reconnects on its own
	// after replReconnectMin, which is the window the backlog is written in.
	conn := follower.repl.conn.Load()
	if conn == nil {
		t.Fatal("follower has no connection to drop")
	}
	_ = conn.ws.Close()

	// 8MB across 2000 ordinary records. Well inside the 64MB ring, so the
	// follower resumes mid-stream instead of taking a snapshot.
	value := strings.Repeat("v", 4*1024)
	for i := 0; i < 2000; i++ {
		if err := leader.Set(fmt.Sprintf("backlog:%d", i), value); err != nil {
			t.Fatalf("backlog write %d: %v", i, err)
		}
	}

	if !replWaitForCursor(follower, leader.walSeq.Load(), 30*time.Second) {
		last := "<none>"
		if p := follower.repl.lastError.Load(); p != nil {
			last = *p
		}
		t.Fatalf("follower stalled at seq %d, leader at seq %d, last error: %s",
			follower.repl.cursor.Load(), leader.walSeq.Load(), last)
	}

	// It resumed rather than resyncing — otherwise this exercised the snapshot
	// path and says nothing about batching.
	if follower.repl.cursor.Load() <= resumeFrom {
		t.Fatalf("follower cursor went backwards: %d, was %d", follower.repl.cursor.Load(), resumeFrom)
	}

	follower.dataMutex.RLock()
	got := len(follower.data)
	sample, hasSample := follower.data["backlog:1999"].(string)
	follower.dataMutex.RUnlock()

	leader.dataMutex.RLock()
	want := len(leader.data)
	leader.dataMutex.RUnlock()

	if got != want {
		t.Errorf("follower has %d keys, leader has %d", got, want)
	}
	if !hasSample || len(sample) != len(value) {
		t.Errorf("follower is missing the last record of the backlog")
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

// The forwarding wire format carries duso values, including the ones JSON would
// mangle, so it round-trips through the same codec the WAL uses.
func TestForwardWireRoundTrip(t *testing.T) {
	t.Parallel()

	// Binary values are not covered here: forwarding calls the same
	// script.EncodeValue the WAL does, so a case for them would test the codec
	// rather than this file. test/ covers it through a real script.
	reqs := []replRequest{
		{ID: 1, Op: opSet, Key: "k", Value: "v", HasValue: true},
		{ID: 2, Op: opIncr, Key: "n", Num: -2.5},
		{ID: 3, Op: opRename, Key: "a", Key2: "b"},
		{ID: 4, Op: opClear},
		{ID: 5, Op: opReqShiftWait, Key: "q", Timeout: 30},
		{ID: 6, Op: opSet, Key: "obj", HasValue: true,
			Value: map[string]any{"nested": []any{1.0, "two", true, nil}}},
	}

	for _, want := range reqs {
		body, err := encodeReplRequest(&want)
		if err != nil {
			t.Fatalf("encode %d: %v", want.ID, err)
		}
		got, err := decodeReplRequest(body)
		if err != nil {
			t.Fatalf("decode %d: %v", want.ID, err)
		}
		if got.ID != want.ID || got.Op != want.Op || got.Key != want.Key ||
			got.Key2 != want.Key2 || got.Num != want.Num || got.Timeout != want.Timeout {
			t.Errorf("request %d round-tripped as %+v, want %+v", want.ID, got, want)
		}
		if want.HasValue && !reflect.DeepEqual(got.Value, want.Value) {
			t.Errorf("request %d value = %#v, want %#v", want.ID, got.Value, want.Value)
		}
	}

	replies := []replReply{
		{ID: 1, Seq: 42, Value: 58.0, HasValue: true},
		{ID: 2, Seq: 0, Value: false, HasValue: true}, // set_once that stored nothing
		{ID: 3, Err: "rename() target already exists"},
		{ID: 4, Seq: 7, Value: []any{"a", 2.0}, HasValue: true},
	}
	for _, want := range replies {
		body, err := encodeReplReply(&want)
		if err != nil {
			t.Fatalf("encode reply %d: %v", want.ID, err)
		}
		got, err := decodeReplReply(body)
		if err != nil {
			t.Fatalf("decode reply %d: %v", want.ID, err)
		}
		if got.ID != want.ID || got.Seq != want.Seq || got.Err != want.Err {
			t.Errorf("reply %d round-tripped as %+v, want %+v", want.ID, got, want)
		}
		if want.HasValue && !reflect.DeepEqual(got.Value, want.Value) {
			t.Errorf("reply %d value = %#v, want %#v", want.ID, got.Value, want.Value)
		}
	}
}

// A reply may only be sent once the frame it refers to has gone out, or the
// follower would be handed a result for a write it cannot yet see. Rejections
// carry no frame and must not be held back.
func TestReplRepliesWaitForTheirFrame(t *testing.T) {
	t.Parallel()

	state := &replConnState{}
	state.queue(&replReply{ID: 1, Seq: 10})
	state.queue(&replReply{ID: 2, Seq: 0, Err: "rejected"})
	state.queue(&replReply{ID: 3, Seq: 25})

	ready := state.take(10)
	if len(ready) != 2 {
		t.Fatalf("take(10) returned %d replies, want 2 (seq 10 and the rejection)", len(ready))
	}
	for _, r := range ready {
		if r.ID == 3 {
			t.Error("a reply was sent before the frame it refers to")
		}
	}

	if ready := state.take(24); len(ready) != 0 {
		t.Errorf("take(24) released the seq-25 reply early")
	}
	if ready := state.take(25); len(ready) != 1 || ready[0].ID != 3 {
		t.Errorf("take(25) did not release the seq-25 reply: %+v", ready)
	}
}

// A leader writes locally and so starts writable; a follower only learns its
// grant from the leader's welcome, so it must not assume one before connecting.
func TestReplGrantDefaults(t *testing.T) {
	t.Parallel()

	leader := newReplTestStore(t, replRoleNone)
	leader.repl = nil
	if err := replConfigure(leader, map[string]any{
		"replicate_listen":          "127.0.0.1:0",
		"replicate_secret":          "rw",
		"replicate_readonly_secret": "ro",
	}); err != nil {
		t.Fatalf("configure leader: %v", err)
	}
	if !leader.repl.canWrite.Load() {
		t.Error("a leader started read-only")
	}
	if leader.repl.readSecret != "ro" {
		t.Errorf("readSecret = %q, want \"ro\"", leader.repl.readSecret)
	}

	follower := newReplTestStore(t, replRoleNone)
	follower.repl = nil
	if err := replConfigure(follower, map[string]any{
		"replicate_from":   "ws://leader:7777",
		"replicate_secret": "ro",
	}); err != nil {
		t.Fatalf("configure follower: %v", err)
	}
	if follower.repl.canWrite.Load() {
		t.Error("a follower assumed write access before the leader granted it")
	}

	// Never connected: the failure names the connection, not the grant, because
	// that is the actionable half.
	err := follower.Set("k", "v")
	if err == nil {
		t.Fatal("a follower with no connection accepted a write")
	}
	if !strings.Contains(err.Error(), "not connected") {
		t.Errorf("error should name the missing connection: %v", err)
	}
	if !strings.Contains(err.Error(), "not applied") {
		t.Errorf("error does not say the write did not land: %v", err)
	}
}

// A replication leader is the process followers reconnect to, so its
// certificate outliving it is the case that matters: a renewal must be picked
// up without a restart that would drop every follower. This drives a real
// TLS leader with a real follower attached, rotates the certificate on disk,
// and checks both that new handshakes get the new certificate and that the
// follower already streaming never notices.
func TestReplicationReloadsRenewedCertificate(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")
	writeKeyPair(t, certPath, keyPath, 4001)

	newStore := func(name string) *DatastoreValue {
		return &DatastoreValue{
			namespace:       "cert_rotate_e2e",
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
		"replicate_listen":               "127.0.0.1:0",
		"replicate_secret":               "shared",
		"replicate_cert_file":            certPath,
		"replicate_key_file":             keyPath,
		"replicate_cert_reload_interval": 0.01,
	}); err != nil {
		t.Fatalf("configure leader: %v", err)
	}
	if err := leader.openWALForWrites(); err != nil {
		t.Fatalf("open leader WAL: %v", err)
	}
	if err := replStart(leader); err != nil {
		t.Fatalf("start leader: %v", err)
	}
	defer leader.replShutdown()

	addr := replBoundAddr("127.0.0.1:0")
	if got := servedSerial(t, addr); got != 4001 {
		t.Fatalf("before renewal: leader served serial %d, want 4001", got)
	}

	// The self-signed leaf is its own CA as far as the follower is concerned.
	follower := newStore("follower")
	if err := replConfigure(follower, map[string]any{
		"replicate_from":    "wss://" + addr,
		"replicate_secret":  "shared",
		"replicate_ca_file": certPath,
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

	if err := leader.Set("before", "renewal"); err != nil {
		t.Fatalf("set before renewal: %v", err)
	}
	waitForCursor(leader.walSeq.Load())

	// Renew underneath the running leader.
	time.Sleep(20 * time.Millisecond)
	writeKeyPair(t, certPath, keyPath, 5002)

	deadline := time.Now().Add(5 * time.Second)
	for {
		if got := servedSerial(t, addr); got == 5002 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("after renewal: leader still serving the old certificate, want serial 5002")
		}
		time.Sleep(20 * time.Millisecond)
	}

	// The follower attached before the rotation must still be streaming: a
	// reload swaps the certificate, it does not restart the listener.
	if !follower.repl.connected.Load() {
		t.Error("follower dropped its connection across the certificate rotation")
	}
	if err := leader.Set("after", "renewal"); err != nil {
		t.Fatalf("set after renewal: %v", err)
	}
	waitForCursor(leader.walSeq.Load())

	follower.dataMutex.RLock()
	got := follower.data["after"]
	follower.dataMutex.RUnlock()
	if got != "renewal" {
		t.Errorf("after the rotation the follower has after=%v, want \"renewal\"", got)
	}
}
