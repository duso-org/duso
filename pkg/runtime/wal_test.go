package runtime

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"
)

func newWALTestStore() *DatastoreValue {
	return &DatastoreValue{
		namespace:   "test",
		walPath:     "test.wal",
		data:        make(map[string]any),
		conditions:  make(map[string]*sync.Cond),
		expiryTimes: make(map[string]time.Time),
		expiryHeap:  make(ExpiryHeap, 0),
	}
}

func TestWALRecordRoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		rec  walRecord
	}{
		{"SET", walRecord{Seq: 1, Op: opSet, Key: "k", Value: "v"}},
		{"SET nested value", walRecord{Seq: 2, Op: opSet, Key: "k", Value: map[string]any{"a": []any{1.0, "b"}}}},
		{"SET_ONCE", walRecord{Seq: 3, Op: opSetOnce, Key: "k", Value: 1.0}},
		{"DELETE", walRecord{Seq: 4, Op: opDelete, Key: "k"}},
		{"SWAP", walRecord{Seq: 5, Op: opSwap, Key: "k", Value: true}},
		{"UPDATE", walRecord{Seq: 6, Op: opUpdate, Key: "k", Value: map[string]any{"patch": 1.0}}},
		{"INCR", walRecord{Seq: 7, Op: opIncr, Key: "k", Num: -2.5}},
		{"RENAME", walRecord{Seq: 8, Op: opRename, Key: "old", Key2: "new"}},
		{"EXPIRE", walRecord{Seq: 9, Op: opExpire, Key: "k", Num: 1700000000000}},
		{"EXPIRED", walRecord{Seq: 10, Op: opExpired, Key: "k"}},
		{"PUSH", walRecord{Seq: 11, Op: opPush, Key: "q", Value: map[string]any{"id": 1.0}}},
		{"UNSHIFT", walRecord{Seq: 12, Op: opUnshift, Key: "q", Value: "x"}},
		{"SHIFT", walRecord{Seq: 13, Op: opShift, Key: "q"}},
		{"POP", walRecord{Seq: 14, Op: opPop, Key: "q"}},
		{"CLEAR", walRecord{Seq: 15, Op: opClear}},
		{"empty key", walRecord{Seq: 16, Op: opDelete, Key: ""}},
		{"unicode key", walRecord{Seq: 17, Op: opSet, Key: "héllo 世界", Value: "v"}},
		{"large seq", walRecord{Seq: 1 << 40, Op: opShift, Key: "q"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			body, err := encodeWALBody(&tc.rec)
			if err != nil {
				t.Fatalf("encodeWALBody failed: %v", err)
			}
			got, err := decodeWALBody(body)
			if err != nil {
				t.Fatalf("decodeWALBody failed: %v", err)
			}
			if !reflect.DeepEqual(*got, tc.rec) {
				t.Errorf("round trip:\n got %#v\nwant %#v", *got, tc.rec)
			}
		})
	}
}

// The opcode determines the shape, so records carry only the fields their op
// needs. A drift between the shape table and the encoder shows up here.
func TestWALRecordIsCompact(t *testing.T) {
	t.Parallel()

	shift, err := encodeWALBody(&walRecord{Seq: 7, Op: opShift, Key: "jobs"})
	if err != nil {
		t.Fatalf("encodeWALBody failed: %v", err)
	}
	// seq(1) + op(1) + keylen(1) + "jobs"(4)
	if len(shift) != 7 {
		t.Errorf("SHIFT body: got %d bytes, want 7 (%v)", len(shift), shift)
	}

	clear, err := encodeWALBody(&walRecord{Seq: 7, Op: opClear})
	if err != nil {
		t.Fatalf("encodeWALBody failed: %v", err)
	}
	// seq(1) + op(1), no key at all
	if len(clear) != 2 {
		t.Errorf("CLEAR body: got %d bytes, want 2 (%v)", len(clear), clear)
	}
}

func TestWALRejectsUnknownDataOpcode(t *testing.T) {
	t.Parallel()

	// 0x20 is inside the data range and unallocated: fatal, since skipping a
	// data record would silently diverge state.
	body := append(binary.AppendUvarint(nil, 1), 0x20)
	if _, err := decodeWALBody(body); err == nil {
		t.Fatal("expected an unknown data opcode to be rejected")
	} else if errors.Is(err, errSkippableRecord) {
		t.Fatal("unknown data opcode must not be skippable")
	}
}

func TestWALSkipsUnknownControlOpcode(t *testing.T) {
	t.Parallel()

	// At or above 0x80 is control/metadata: an older reader can ignore it
	// without its key-value state being wrong.
	for _, op := range []byte{0x80, 0x90, 0xBF, 0xC0} {
		body := append(binary.AppendUvarint(nil, 42), op)
		rec, err := decodeWALBody(body)
		if !errors.Is(err, errSkippableRecord) {
			t.Errorf("opcode 0x%02X: got err %v, want errSkippableRecord", op, err)
		}
		if rec == nil || rec.Seq != 42 {
			t.Errorf("opcode 0x%02X: sequence number should still be readable", op)
		}
	}
}

func TestWALRejectsTruncatedRecord(t *testing.T) {
	t.Parallel()

	full, err := encodeWALBody(&walRecord{
		Seq: 9, Op: opSet, Key: "some_key", Value: map[string]any{"a": "a longer string value"},
	})
	if err != nil {
		t.Fatalf("encodeWALBody failed: %v", err)
	}

	for cut := 0; cut < len(full); cut++ {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("panic decoding record truncated to %d bytes: %v", cut, r)
				}
			}()
			if _, err := decodeWALBody(full[:cut]); err == nil {
				t.Errorf("expected an error decoding record truncated to %d bytes", cut)
			}
		}()
	}
}

// Trailing bytes mean the shape table and the writer disagree, which must not
// be silently tolerated.
func TestWALRejectsTrailingBytes(t *testing.T) {
	t.Parallel()

	body, err := encodeWALBody(&walRecord{Seq: 1, Op: opShift, Key: "q"})
	if err != nil {
		t.Fatalf("encodeWALBody failed: %v", err)
	}
	if _, err := decodeWALBody(append(body, 0xFF)); err == nil {
		t.Fatal("expected trailing bytes to be rejected")
	}
}

func TestWALRefusesToWriteUnknownOpcode(t *testing.T) {
	t.Parallel()

	if _, err := encodeWALBody(&walRecord{Seq: 1, Op: walOp(0x77), Key: "k"}); err == nil {
		t.Fatal("expected encoding an unknown opcode to fail")
	}
}

// Replay semantics: each opcode must reconstruct exactly what the live
// operation did.
func TestWALApplyRecord(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		before map[string]any
		recs   []walRecord
		want   map[string]any
	}{
		{
			name:   "SET overwrites",
			before: map[string]any{"k": "old"},
			recs:   []walRecord{{Op: opSet, Key: "k", Value: "new"}},
			want:   map[string]any{"k": "new"},
		},
		{
			name:   "DELETE removes the key rather than nilling it",
			before: map[string]any{"k": "v"},
			recs:   []walRecord{{Op: opDelete, Key: "k"}},
			want:   map[string]any{},
		},
		{
			name:   "SET nil keeps the key present",
			before: map[string]any{},
			recs:   []walRecord{{Op: opSet, Key: "k", Value: nil}},
			want:   map[string]any{"k": nil},
		},
		{
			name:   "INCR applies the delta",
			before: map[string]any{"n": 10.0},
			recs:   []walRecord{{Op: opIncr, Key: "n", Num: 5}, {Op: opIncr, Key: "n", Num: -2}},
			want:   map[string]any{"n": 13.0},
		},
		{
			name:   "INCR creates from zero",
			before: map[string]any{},
			recs:   []walRecord{{Op: opIncr, Key: "n", Num: 3}},
			want:   map[string]any{"n": 3.0},
		},
		{
			name:   "PUSH appends in order",
			before: map[string]any{},
			recs: []walRecord{
				{Op: opPush, Key: "q", Value: "a"},
				{Op: opPush, Key: "q", Value: "b"},
			},
			want: map[string]any{"q": []any{"a", "b"}},
		},
		{
			name:   "UNSHIFT prepends",
			before: map[string]any{"q": []any{"b"}},
			recs:   []walRecord{{Op: opUnshift, Key: "q", Value: "a"}},
			want:   map[string]any{"q": []any{"a", "b"}},
		},
		{
			name:   "SHIFT removes the first element",
			before: map[string]any{"q": []any{"a", "b", "c"}},
			recs:   []walRecord{{Op: opShift, Key: "q"}},
			want:   map[string]any{"q": []any{"b", "c"}},
		},
		{
			name:   "POP removes the last element",
			before: map[string]any{"q": []any{"a", "b", "c"}},
			recs:   []walRecord{{Op: opPop, Key: "q"}},
			want:   map[string]any{"q": []any{"a", "b"}},
		},
		{
			name:   "SHIFT on an empty array is a no-op",
			before: map[string]any{"q": []any{}},
			recs:   []walRecord{{Op: opShift, Key: "q"}},
			want:   map[string]any{"q": []any{}},
		},
		{
			name:   "push then shift leaves the remainder",
			before: map[string]any{},
			recs: []walRecord{
				{Op: opPush, Key: "q", Value: "a"},
				{Op: opPush, Key: "q", Value: "b"},
				{Op: opShift, Key: "q"},
			},
			want: map[string]any{"q": []any{"b"}},
		},
		{
			name:   "UPDATE deep merges the patch",
			before: map[string]any{"c": map[string]any{"a": 1.0, "nested": map[string]any{"x": 1.0}}},
			recs:   []walRecord{{Op: opUpdate, Key: "c", Value: map[string]any{"b": 2.0, "nested": map[string]any{"y": 2.0}}}},
			want: map[string]any{"c": map[string]any{
				"a": 1.0, "b": 2.0,
				"nested": map[string]any{"x": 1.0, "y": 2.0},
			}},
		},
		{
			name:   "UPDATE nil deletes a field",
			before: map[string]any{"c": map[string]any{"a": 1.0, "b": 2.0}},
			recs:   []walRecord{{Op: opUpdate, Key: "c", Value: map[string]any{"b": nil}}},
			want:   map[string]any{"c": map[string]any{"a": 1.0}},
		},
		{
			name:   "RENAME moves the value",
			before: map[string]any{"old": "v"},
			recs:   []walRecord{{Op: opRename, Key: "old", Key2: "new"}},
			want:   map[string]any{"new": "v"},
		},
		{
			name:   "RENAME of a missing key is a no-op",
			before: map[string]any{"other": "v"},
			recs:   []walRecord{{Op: opRename, Key: "old", Key2: "new"}},
			want:   map[string]any{"other": "v"},
		},
		{
			name:   "CLEAR empties the store",
			before: map[string]any{"a": 1.0, "b": 2.0},
			recs:   []walRecord{{Op: opClear}},
			want:   map[string]any{},
		},
		{
			name:   "CLEAR then write",
			before: map[string]any{"a": 1.0},
			recs:   []walRecord{{Op: opClear}, {Op: opSet, Key: "b", Value: 2.0}},
			want:   map[string]any{"b": 2.0},
		},
		{
			name:   "EXPIRED removes the key",
			before: map[string]any{"s": "v"},
			recs:   []walRecord{{Op: opExpired, Key: "s"}},
			want:   map[string]any{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ds := newWALTestStore()
			ds.data = tc.before
			for i := range tc.recs {
				if err := ds.applyWALRecord(&tc.recs[i]); err != nil {
					t.Fatalf("applyWALRecord(%s) failed: %v", walShapes[tc.recs[i].Op].name, err)
				}
			}
			if !reflect.DeepEqual(ds.data, tc.want) {
				t.Errorf("after replay:\n got %#v\nwant %#v", ds.data, tc.want)
			}
		})
	}
}

// EXPIRE carries an absolute deadline. Replaying it long after the fact must not
// extend the key's life the way a relative TTL would.
func TestWALExpireIsAbsolute(t *testing.T) {
	t.Parallel()

	past := time.Now().Add(-time.Hour).Truncate(time.Millisecond)

	ds := newWALTestStore()
	ds.data = map[string]any{"s": "v"}
	rec := walRecord{Op: opExpire, Key: "s", Num: float64(past.UnixMilli())}
	if err := ds.applyWALRecord(&rec); err != nil {
		t.Fatalf("applyWALRecord failed: %v", err)
	}

	got, ok := ds.expiryTimes["s"]
	if !ok {
		t.Fatal("deadline was not restored")
	}
	if !got.Equal(past) {
		t.Errorf("deadline: got %v, want %v", got, past)
	}
	if got.After(time.Now()) {
		t.Error("a deadline that was already past replayed as being in the future")
	}
	if len(ds.expiryHeap) != 1 {
		t.Errorf("expiry heap has %d entries, want 1", len(ds.expiryHeap))
	}
}

func TestWALExpireForMissingKeyIgnored(t *testing.T) {
	t.Parallel()

	ds := newWALTestStore()
	rec := walRecord{Op: opExpire, Key: "absent", Num: float64(time.Now().UnixMilli())}
	if err := ds.applyWALRecord(&rec); err != nil {
		t.Fatalf("applyWALRecord failed: %v", err)
	}
	if len(ds.expiryTimes) != 0 {
		t.Error("a deadline was recorded for a key that does not exist")
	}
}

func TestWALIncrOnNonNumericFails(t *testing.T) {
	t.Parallel()

	ds := newWALTestStore()
	ds.data = map[string]any{"k": "not a number"}
	rec := walRecord{Op: opIncr, Key: "k", Num: 1}
	if err := ds.applyWALRecord(&rec); err == nil {
		t.Fatal("expected INCR against a non-numeric value to fail")
	}
}

func TestWALHeaderRoundTrip(t *testing.T) {
	t.Parallel()

	t.Run("plaintext", func(t *testing.T) {
		t.Parallel()
		ds := newWALTestStore()
		hdr := buildWALHeader(nil)
		encrypted, err := ds.readWALHeader(hdr)
		if err != nil {
			t.Fatalf("readWALHeader failed: %v", err)
		}
		if encrypted {
			t.Error("plaintext header reported as encrypted")
		}
	})

	t.Run("encrypted", func(t *testing.T) {
		t.Parallel()
		key := []byte("0123456789abcdef0123456789abcdef")
		ds := newWALTestStore()
		ds.encryptKey = key
		encrypted, err := ds.readWALHeader(buildWALHeader(key))
		if err != nil {
			t.Fatalf("readWALHeader failed: %v", err)
		}
		if !encrypted {
			t.Error("encrypted header not reported as encrypted")
		}
	})

	t.Run("encrypted file, no key configured", func(t *testing.T) {
		t.Parallel()
		ds := newWALTestStore()
		if _, err := ds.readWALHeader(buildWALHeader([]byte("0123456789abcdef0123456789abcdef"))); err == nil {
			t.Fatal("expected an error, got nil")
		}
	})

	t.Run("plaintext file, key configured", func(t *testing.T) {
		t.Parallel()
		ds := newWALTestStore()
		ds.encryptKey = []byte("0123456789abcdef0123456789abcdef")
		if _, err := ds.readWALHeader(buildWALHeader(nil)); err == nil {
			t.Fatal("expected an error, got nil")
		}
	})

	t.Run("future version rejected", func(t *testing.T) {
		t.Parallel()
		hdr := buildWALHeader(nil)
		binary.LittleEndian.PutUint16(hdr[8:10], walVersion+1)
		if _, err := newWALTestStore().readWALHeader(hdr); err == nil {
			t.Fatal("expected a version mismatch error, got nil")
		}
	})

	t.Run("wrong magic rejected", func(t *testing.T) {
		t.Parallel()
		hdr := buildWALHeader(nil)
		copy(hdr[:8], "NOTADUSO")
		if _, err := newWALTestStore().readWALHeader(hdr); err == nil {
			t.Fatal("expected a magic mismatch error, got nil")
		}
	})
}

// buildWALHeader mirrors appendWALHeader without needing a file.
func buildWALHeader(key []byte) []byte {
	var flags uint16
	if len(key) > 0 {
		flags |= walFlagEncrypted
	}
	hdr := make([]byte, 0, walHeaderLen)
	hdr = append(hdr, walMagic...)
	hdr = binary.LittleEndian.AppendUint16(hdr, walVersion)
	hdr = binary.LittleEndian.AppendUint16(hdr, flags)
	return binary.LittleEndian.AppendUint32(hdr, 0)
}

// The magic must not collide with what a pre-v1 gob WAL starts with, or
// migration would misread every legacy file.
func TestWALMagicDistinguishesLegacy(t *testing.T) {
	t.Parallel()

	if len(walMagic) != 8 {
		t.Fatalf("magic is %d bytes, expected 8", len(walMagic))
	}
	if bytes.HasPrefix([]byte{0x0d, 0x7f, 0x04, 0x01, 0x02, 0xff, 0x80, 0x00}, []byte(walMagic)) {
		t.Error("magic collides with a gob stream prefix")
	}
}

// writeTestWAL builds a real v1 WAL file containing the given records and
// returns its path.
func writeTestWAL(t *testing.T, dir string, key []byte, recs ...walRecord) string {
	t.Helper()

	path := filepath.Join(dir, "test.duwal")
	ds := newWALTestStore()
	ds.walPath = path
	ds.encryptKey = key
	ds.walSyncInterval = 0

	if err := ds.openWALForWrites(); err != nil {
		t.Fatalf("openWALForWrites failed: %v", err)
	}
	for i := range recs {
		r := recs[i]
		if err := ds.writeWALOp(r.Op, r.Key, r.Key2, r.Num, r.Value); err != nil {
			t.Fatalf("writeWALOp failed: %v", err)
		}
	}
	if err := ds.walFile.Close(); err != nil {
		t.Fatalf("closing WAL failed: %v", err)
	}
	return path
}

// replayFile runs a fresh store's replay over path and returns it.
func replayFile(t *testing.T, path string, key []byte, readonly bool) (*DatastoreValue, error) {
	t.Helper()

	ds := newWALTestStore()
	ds.walPath = path
	ds.encryptKey = key
	ds.readonly = readonly
	return ds, ds.replayWAL()
}

// A crash almost always leaves a partial final record. That is the normal end of
// a write-ahead log, so replay must stop cleanly and keep everything before it.
func TestWALTornTailIsRecoverable(t *testing.T) {
	t.Parallel()

	recs := []walRecord{
		{Op: opSet, Key: "a", Value: "first"},
		{Op: opSet, Key: "b", Value: "second"},
		{Op: opSet, Key: "c", Value: "third"},
	}

	full, err := os.ReadFile(writeTestWAL(t, t.TempDir(), nil, recs...))
	if err != nil {
		t.Fatalf("test setup: %v", err)
	}

	// Cut at every byte past the header. Every prefix must replay without error
	// and yield a prefix of the records — never a partial or invented one.
	for cut := walHeaderLen; cut <= len(full); cut++ {
		dir := t.TempDir()
		path := filepath.Join(dir, "torn.duwal")
		if err := os.WriteFile(path, full[:cut], 0644); err != nil {
			t.Fatalf("test setup: %v", err)
		}

		ds, err := replayFile(t, path, nil, false)
		if err != nil {
			t.Fatalf("cut at %d: replay failed on a torn tail: %v", cut, err)
		}

		// Whatever survived must be a prefix: no "c" without "b".
		_, hasA := ds.data["a"]
		_, hasB := ds.data["b"]
		_, hasC := ds.data["c"]
		if hasC && !hasB || hasB && !hasA {
			t.Errorf("cut at %d: recovered a gap, got a=%v b=%v c=%v", cut, hasA, hasB, hasC)
		}
		if cut == len(full) && !hasC {
			t.Errorf("an untruncated log lost its last record")
		}
	}
}

// The torn tail must be physically removed, or later appends would sit behind
// bytes nothing can reach.
func TestWALTornTailIsTruncated(t *testing.T) {
	t.Parallel()

	full, err := os.ReadFile(writeTestWAL(t, t.TempDir(), nil,
		walRecord{Op: opSet, Key: "a", Value: "first"},
		walRecord{Op: opSet, Key: "b", Value: "second"},
	))
	if err != nil {
		t.Fatalf("test setup: %v", err)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "torn.duwal")
	if err := os.WriteFile(path, full[:len(full)-3], 0644); err != nil {
		t.Fatalf("test setup: %v", err)
	}

	if _, err := replayFile(t, path, nil, false); err != nil {
		t.Fatalf("replay failed: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	if info.Size() >= int64(len(full)-3) {
		t.Errorf("torn tail was not truncated: file is still %d bytes", info.Size())
	}
	if info.Size() < walHeaderLen {
		t.Errorf("truncation ate the header: file is %d bytes", info.Size())
	}
}

// A readonly store may be looking at files another process is writing. It must
// never modify them, even to tidy a torn tail.
func TestWALReadonlyDoesNotTruncate(t *testing.T) {
	t.Parallel()

	full, err := os.ReadFile(writeTestWAL(t, t.TempDir(), nil,
		walRecord{Op: opSet, Key: "a", Value: "first"},
		walRecord{Op: opSet, Key: "b", Value: "second"},
	))
	if err != nil {
		t.Fatalf("test setup: %v", err)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "torn.duwal")
	truncated := full[:len(full)-3]
	if err := os.WriteFile(path, truncated, 0644); err != nil {
		t.Fatalf("test setup: %v", err)
	}

	ds, err := replayFile(t, path, nil, true)
	if err != nil {
		t.Fatalf("readonly replay failed: %v", err)
	}
	if ds.data["a"] != "first" {
		t.Errorf("readonly replay lost intact records: %#v", ds.data)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	if info.Size() != int64(len(truncated)) {
		t.Errorf("readonly replay modified the file: %d bytes, want %d", info.Size(), len(truncated))
	}
}

// Damage in the middle is not tornness. Everything after it is unreachable
// because the framing can no longer be trusted, so it must be loud.
func TestWALMidLogCorruptionIsFatal(t *testing.T) {
	t.Parallel()

	path := writeTestWAL(t, t.TempDir(), nil,
		walRecord{Op: opSet, Key: "a", Value: "first"},
		walRecord{Op: opSet, Key: "b", Value: "second"},
		walRecord{Op: opSet, Key: "c", Value: "third"},
	)
	full, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("test setup: %v", err)
	}

	// Flip a byte inside the first record's body, well before the end.
	corrupt := make([]byte, len(full))
	copy(corrupt, full)
	corrupt[walHeaderLen+walFrameLen+2] ^= 0xFF
	if err := os.WriteFile(path, corrupt, 0644); err != nil {
		t.Fatalf("test setup: %v", err)
	}

	if _, err := replayFile(t, path, nil, false); err == nil {
		t.Fatal("expected mid-log corruption to be fatal, got nil")
	}
}

// The same damage on the final record is indistinguishable from a half-flushed
// write, so it is treated as a torn tail.
func TestWALCorruptFinalRecordIsTornTail(t *testing.T) {
	t.Parallel()

	path := writeTestWAL(t, t.TempDir(), nil,
		walRecord{Op: opSet, Key: "a", Value: "first"},
		walRecord{Op: opSet, Key: "b", Value: "second"},
	)
	full, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("test setup: %v", err)
	}

	corrupt := make([]byte, len(full))
	copy(corrupt, full)
	corrupt[len(corrupt)-1] ^= 0xFF
	if err := os.WriteFile(path, corrupt, 0644); err != nil {
		t.Fatalf("test setup: %v", err)
	}

	ds, err := replayFile(t, path, nil, false)
	if err != nil {
		t.Fatalf("a corrupt final record should be treated as a torn tail: %v", err)
	}
	if ds.data["a"] != "first" {
		t.Errorf("intact earlier record was lost: %#v", ds.data)
	}
	if _, ok := ds.data["b"]; ok {
		t.Error("the corrupt final record was applied anyway")
	}
}

// An encrypted log tears the same way; the checksum covers the sealed bytes so
// integrity is checkable without the key.
func TestWALTornTailEncrypted(t *testing.T) {
	t.Parallel()

	key := []byte("0123456789abcdef0123456789abcdef")
	full, err := os.ReadFile(writeTestWAL(t, t.TempDir(), key,
		walRecord{Op: opSet, Key: "a", Value: "first"},
		walRecord{Op: opSet, Key: "b", Value: "second"},
	))
	if err != nil {
		t.Fatalf("test setup: %v", err)
	}

	for cut := walHeaderLen; cut <= len(full); cut++ {
		dir := t.TempDir()
		path := filepath.Join(dir, "torn.duwal")
		if err := os.WriteFile(path, full[:cut], 0644); err != nil {
			t.Fatalf("test setup: %v", err)
		}
		if _, err := replayFile(t, path, key, false); err != nil {
			t.Fatalf("cut at %d: encrypted torn tail failed to replay: %v", cut, err)
		}
	}
}

// A log holding only a header is the state right after a truncate.
func TestWALHeaderOnlyReplaysClean(t *testing.T) {
	t.Parallel()

	path := writeTestWAL(t, t.TempDir(), nil)
	ds, err := replayFile(t, path, nil, false)
	if err != nil {
		t.Fatalf("header-only replay failed: %v", err)
	}
	if len(ds.data) != 0 {
		t.Errorf("header-only log produced %d keys", len(ds.data))
	}
}
