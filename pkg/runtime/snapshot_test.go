package runtime

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/duso-org/duso/pkg/script"
)

// newTestStore builds a bare DatastoreValue for format tests, skipping
// GetDatastore so no background tickers are started.
func newTestStore(key []byte) *DatastoreValue {
	return &DatastoreValue{
		namespace:   "test",
		persistPath: "test.dsnap",
		data:        make(map[string]any),
		expiryTimes: make(map[string]time.Time),
		expiryHeap:  make(ExpiryHeap, 0),
		encryptKey:  key,
	}
}

func TestSnapshotRoundTrip(t *testing.T) {
	t.Parallel()

	ds := newTestStore(nil)
	ds.data = map[string]any{
		"number":  42.0,
		"string":  "héllo 世界",
		"bool":    true,
		"nil":     nil,
		"array":   []any{1.0, "two", false, nil},
		"object":  map[string]any{"nested": map[string]any{"deep": 3.0}},
		"empties": map[string]any{},
	}
	ds.walSeq.Store(12345)

	raw, err := ds.encodeSnapshot()
	if err != nil {
		t.Fatalf("encodeSnapshot failed: %v", err)
	}
	if !bytes.HasPrefix(raw, []byte(snapshotMagic)) {
		t.Fatalf("snapshot does not start with magic: %q", raw[:8])
	}

	out := newTestStore(nil)
	data, expiry, seq, err := out.decodeSnapshot(raw)
	if err != nil {
		t.Fatalf("decodeSnapshot failed: %v", err)
	}
	if !reflect.DeepEqual(data, ds.data) {
		t.Errorf("data mismatch:\n got %#v\nwant %#v", data, ds.data)
	}
	if seq != 12345 {
		t.Errorf("seq watermark: got %d, want 12345", seq)
	}
	if len(expiry) != 0 {
		t.Errorf("expected no deadlines, got %d", len(expiry))
	}
}

func TestSnapshotExpiryRoundTrip(t *testing.T) {
	t.Parallel()

	// Truncate to milliseconds: that is the format's resolution.
	future := time.Now().Add(time.Hour).Truncate(time.Millisecond)
	past := time.Now().Add(-time.Hour).Truncate(time.Millisecond)

	ds := newTestStore(nil)
	ds.data = map[string]any{"live": 1.0, "stale": 2.0, "noexpiry": 3.0}
	ds.expiryTimes = map[string]time.Time{"live": future, "stale": past}

	raw, err := ds.encodeSnapshot()
	if err != nil {
		t.Fatalf("encodeSnapshot failed: %v", err)
	}

	out := newTestStore(nil)
	data, expiry, _, err := out.decodeSnapshot(raw)
	if err != nil {
		t.Fatalf("decodeSnapshot failed: %v", err)
	}
	if len(expiry) != 2 {
		t.Fatalf("expected 2 deadlines, got %d", len(expiry))
	}
	if !expiry["live"].Equal(future) {
		t.Errorf("live deadline: got %v, want %v", expiry["live"], future)
	}
	// A deadline already in the past must survive decode rather than being
	// dropped — the sweeper is what removes the key, and that is what produces
	// the authoritative deletion record.
	if !expiry["stale"].Equal(past) {
		t.Errorf("stale deadline: got %v, want %v", expiry["stale"], past)
	}

	out.data = data
	out.applySnapshotExpiry(expiry)
	if len(out.expiryTimes) != 2 {
		t.Errorf("applySnapshotExpiry: got %d deadlines, want 2", len(out.expiryTimes))
	}
	if len(out.expiryHeap) != 2 {
		t.Errorf("applySnapshotExpiry: heap has %d entries, want 2", len(out.expiryHeap))
	}
}

// A deadline for a key the snapshot no longer holds must not resurrect anything.
func TestSnapshotExpiryForMissingKeyIgnored(t *testing.T) {
	t.Parallel()

	ds := newTestStore(nil)
	ds.data = map[string]any{"present": 1.0}
	ds.applySnapshotExpiry(map[string]time.Time{
		"present": time.Now().Add(time.Hour),
		"absent":  time.Now().Add(time.Hour),
	})

	if _, ok := ds.expiryTimes["absent"]; ok {
		t.Error("deadline for a missing key was applied")
	}
	if len(ds.expiryTimes) != 1 {
		t.Errorf("got %d deadlines, want 1", len(ds.expiryTimes))
	}
}

func TestSnapshotDeterministic(t *testing.T) {
	t.Parallel()

	ds := newTestStore(nil)
	ds.data = map[string]any{
		"zebra": 1.0, "alpha": 2.0, "mike": 3.0, "": 4.0, "🐈": 5.0,
	}
	ds.expiryTimes = map[string]time.Time{
		"zebra": time.UnixMilli(1000), "alpha": time.UnixMilli(2000),
	}

	first, err := ds.encodeSnapshot()
	if err != nil {
		t.Fatalf("encodeSnapshot failed: %v", err)
	}
	for i := 0; i < 50; i++ {
		again, err := ds.encodeSnapshot()
		if err != nil {
			t.Fatalf("encodeSnapshot failed on pass %d: %v", i, err)
		}
		if !bytes.Equal(first, again) {
			t.Fatalf("snapshot encoding is not deterministic: pass %d differs", i)
		}
	}
}

func TestSnapshotEncrypted(t *testing.T) {
	t.Parallel()

	key := []byte("0123456789abcdef0123456789abcdef")
	ds := newTestStore(key)
	ds.data = map[string]any{"secret": "classified payload"}
	ds.walSeq.Store(7)

	raw, err := ds.encodeSnapshot()
	if err != nil {
		t.Fatalf("encodeSnapshot failed: %v", err)
	}

	// Header must stay readable without the key.
	if !bytes.HasPrefix(raw, []byte(snapshotMagic)) {
		t.Error("encrypted snapshot lost its plaintext magic")
	}
	if flags := binary.LittleEndian.Uint16(raw[10:12]); flags&snapshotFlagEncrypted == 0 {
		t.Error("encrypted flag not set in header")
	}
	if seq := binary.LittleEndian.Uint64(raw[16:24]); seq != 7 {
		t.Errorf("seq should be readable without the key: got %d, want 7", seq)
	}
	if bytes.Contains(raw, []byte("classified")) {
		t.Error("plaintext leaked into an encrypted snapshot")
	}

	out := newTestStore(key)
	data, _, _, err := out.decodeSnapshot(raw)
	if err != nil {
		t.Fatalf("decodeSnapshot failed: %v", err)
	}
	if data["secret"] != "classified payload" {
		t.Errorf("secret: got %#v", data["secret"])
	}
}

// A key mismatch in either direction must be a clear error, never a silently
// empty store.
func TestSnapshotKeyMismatch(t *testing.T) {
	t.Parallel()

	key := []byte("0123456789abcdef0123456789abcdef")

	encrypted := newTestStore(key)
	encrypted.data = map[string]any{"a": 1.0}
	encRaw, err := encrypted.encodeSnapshot()
	if err != nil {
		t.Fatalf("encodeSnapshot failed: %v", err)
	}

	plain := newTestStore(nil)
	plain.data = map[string]any{"a": 1.0}
	plainRaw, err := plain.encodeSnapshot()
	if err != nil {
		t.Fatalf("encodeSnapshot failed: %v", err)
	}

	t.Run("encrypted file, no key configured", func(t *testing.T) {
		t.Parallel()
		if _, _, _, err := newTestStore(nil).decodeSnapshot(encRaw); err == nil {
			t.Fatal("expected an error, got nil")
		}
	})

	t.Run("encrypted file, wrong key", func(t *testing.T) {
		t.Parallel()
		wrong := []byte("WRONGWRONGWRONGWRONGWRONGWRONGWR")
		if _, _, _, err := newTestStore(wrong).decodeSnapshot(encRaw); err == nil {
			t.Fatal("expected an error, got nil")
		}
	})

	t.Run("plaintext file, key configured", func(t *testing.T) {
		t.Parallel()
		if _, _, _, err := newTestStore(key).decodeSnapshot(plainRaw); err == nil {
			t.Fatal("expected an error, got nil")
		}
	})
}

// A file written by a newer duso must be refused by name, not misparsed.
func TestSnapshotRejectsFutureVersion(t *testing.T) {
	t.Parallel()

	ds := newTestStore(nil)
	ds.data = map[string]any{"a": 1.0}
	raw, err := ds.encodeSnapshot()
	if err != nil {
		t.Fatalf("encodeSnapshot failed: %v", err)
	}
	binary.LittleEndian.PutUint16(raw[8:10], snapshotVersion+1)

	if _, _, _, err := newTestStore(nil).decodeSnapshot(raw); err == nil {
		t.Fatal("expected a version mismatch error, got nil")
	}
}

func TestSnapshotRejectsTruncated(t *testing.T) {
	t.Parallel()

	ds := newTestStore(nil)
	ds.data = map[string]any{
		"alpha": "a reasonably long string value",
		"bravo": []any{1.0, 2.0, 3.0},
	}
	ds.expiryTimes = map[string]time.Time{"alpha": time.UnixMilli(5000)}

	raw, err := ds.encodeSnapshot()
	if err != nil {
		t.Fatalf("encodeSnapshot failed: %v", err)
	}

	// Cut anywhere inside the header or body. Truncation must error, never panic.
	// (An empty prefix is a legacy-format read attempt, handled separately.)
	for cut := 1; cut < len(raw); cut++ {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("panic decoding snapshot truncated to %d bytes: %v", cut, r)
				}
			}()
			if _, _, _, err := newTestStore(nil).decodeSnapshot(raw[:cut]); err == nil {
				t.Errorf("expected an error decoding snapshot truncated to %d bytes", cut)
			}
		}()
	}
}

// Pre-v1 files are a bare gob map with no header and must still load.
func TestSnapshotReadsLegacyGob(t *testing.T) {
	t.Parallel()

	want := map[string]any{
		"greeting": "hello from the old format",
		"count":    99.0,
		"nested":   map[string]any{"ok": true, "list": []any{1.0, "two"}},
	}

	var b bytes.Buffer
	if err := gob.NewEncoder(&b).Encode(want); err != nil {
		t.Fatalf("test setup: %v", err)
	}

	data, expiry, seq, err := newTestStore(nil).decodeSnapshot(b.Bytes())
	if err != nil {
		t.Fatalf("decodeSnapshot on legacy file failed: %v", err)
	}
	if !reflect.DeepEqual(data, want) {
		t.Errorf("legacy data mismatch:\n got %#v\nwant %#v", data, want)
	}
	// Pre-v1 files carry neither deadlines nor a watermark.
	if len(expiry) != 0 {
		t.Errorf("expected no deadlines from a legacy file, got %d", len(expiry))
	}
	if seq != 0 {
		t.Errorf("expected seq 0 from a legacy file, got %d", seq)
	}
}

// Reading a legacy file and re-encoding must produce a v1 file — the
// read-and-convert migration.
func TestSnapshotLegacyConvertsToV1(t *testing.T) {
	t.Parallel()

	var b bytes.Buffer
	if err := gob.NewEncoder(&b).Encode(map[string]any{"k": "v"}); err != nil {
		t.Fatalf("test setup: %v", err)
	}

	ds := newTestStore(nil)
	data, _, _, err := ds.decodeSnapshot(b.Bytes())
	if err != nil {
		t.Fatalf("decodeSnapshot failed: %v", err)
	}
	ds.data = data

	raw, err := ds.encodeSnapshot()
	if err != nil {
		t.Fatalf("encodeSnapshot failed: %v", err)
	}
	if !bytes.HasPrefix(raw, []byte(snapshotMagic)) {
		t.Fatal("re-encoded legacy snapshot is not in v1 format")
	}

	back, _, _, err := newTestStore(nil).decodeSnapshot(raw)
	if err != nil {
		t.Fatalf("decoding the converted snapshot failed: %v", err)
	}
	if back["k"] != "v" {
		t.Errorf("value lost in conversion: got %#v", back["k"])
	}
}

func TestSnapshotEmptyStore(t *testing.T) {
	t.Parallel()

	ds := newTestStore(nil)
	raw, err := ds.encodeSnapshot()
	if err != nil {
		t.Fatalf("encodeSnapshot failed: %v", err)
	}
	data, expiry, _, err := newTestStore(nil).decodeSnapshot(raw)
	if err != nil {
		t.Fatalf("decodeSnapshot failed: %v", err)
	}
	if len(data) != 0 || len(expiry) != 0 {
		t.Errorf("expected an empty store, got %d keys and %d deadlines", len(data), len(expiry))
	}
}

// Binary values could not be persisted at all before the codec landed: the gob
// encoder rejected them and the whole write failed.
func TestSnapshotHoldsBinaryValues(t *testing.T) {
	t.Parallel()

	payload := []byte{0x00, 0xFF, 'h', 'i'}
	ds := newTestStore(nil)
	ds.data = map[string]any{"file": &ValueRef{Val: script.NewBinary(payload)}}

	raw, err := ds.encodeSnapshot()
	if err != nil {
		t.Fatalf("encodeSnapshot failed: %v", err)
	}
	data, _, _, err := newTestStore(nil).decodeSnapshot(raw)
	if err != nil {
		t.Fatalf("decodeSnapshot failed: %v", err)
	}

	ref, ok := data["file"].(*ValueRef)
	if !ok {
		t.Fatalf("got %T, want *ValueRef", data["file"])
	}
	if got := *ref.Val.AsBinary().Data; !bytes.Equal(got, payload) {
		t.Errorf("binary payload: got %v, want %v", got, payload)
	}
}

// A store that cannot load its snapshot or validate its key must fail the
// datastore() call outright, not start up degraded.
func TestApplyConfigFailsLoudly(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	good := base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef"))

	encPath := filepath.Join(dir, "enc.dsnap")
	enc := newTestStore([]byte("0123456789abcdef0123456789abcdef"))
	enc.data = map[string]any{"a": 1.0}
	raw, err := enc.encodeSnapshot()
	if err != nil {
		t.Fatalf("test setup: %v", err)
	}
	if err := os.WriteFile(encPath, raw, 0644); err != nil {
		t.Fatalf("test setup: %v", err)
	}

	garbagePath := filepath.Join(dir, "garbage.dsnap")
	if err := os.WriteFile(garbagePath, []byte("DUSOSNAP\x01\x00"), 0644); err != nil {
		t.Fatalf("test setup: %v", err)
	}

	tests := []struct {
		name   string
		config map[string]any
	}{
		{"encrypt_key not base64", map[string]any{"encrypt_key": "not!valid!base64!"}},
		{"encrypt_key wrong length", map[string]any{"encrypt_key": base64.StdEncoding.EncodeToString([]byte("tooshort"))}},
		{"encrypt_key not a string", map[string]any{"encrypt_key": 42.0}},
		{"encrypted file, no key", map[string]any{"persist": encPath}},
		{"encrypted file, wrong key", map[string]any{
			"persist":     encPath,
			"encrypt_key": base64.StdEncoding.EncodeToString([]byte("WRONGWRONGWRONGWRONGWRONGWRONGWR")),
		}},
		{"truncated snapshot", map[string]any{"persist": garbagePath}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if err := applyDatastoreConfig(newTestStore(nil), tc.config); err == nil {
				t.Fatal("expected applyDatastoreConfig to fail, got nil")
			}
		})
	}

	t.Run("valid config succeeds", func(t *testing.T) {
		t.Parallel()
		cfg := map[string]any{"persist": encPath, "encrypt_key": good}
		store := newTestStore(nil)
		if err := applyDatastoreConfig(store, cfg); err != nil {
			t.Fatalf("expected success, got %v", err)
		}
		if store.data["a"] != 1.0 {
			t.Errorf("snapshot not loaded: got %#v", store.data["a"])
		}
	})
}

// TestSaveReplacesFileAtomically pins the property that makes a crash during a
// save survivable: the snapshot file is never observed in a partial state, and
// a failed save leaves the previous snapshot untouched. An in-place O_TRUNC
// write satisfies neither.
func TestSaveReplacesFileAtomically(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "store.dsnap")

	original := []byte("the snapshot that must survive a failed save")
	if err := writeFileAtomic(path, original, 0644); err != nil {
		t.Fatalf("writeFileAtomic: %v", err)
	}

	// A save that cannot complete must not consume the old file. Renaming onto
	// a directory fails at the rename, which is the last and least recoverable
	// step -- if the old contents survive that, they survive the earlier ones.
	dirTarget := filepath.Join(dir, "subdir")
	if err := os.Mkdir(dirTarget, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := writeFileAtomic(dirTarget, []byte("nope"), 0644); err == nil {
		t.Fatal("expected writeFileAtomic to fail when the target is a directory")
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading original after failed save: %v", err)
	}
	if !bytes.Equal(got, original) {
		t.Errorf("previous snapshot was damaged by a failed save: got %q", got)
	}

	// A failed save must not litter the directory with temp files either --
	// they would accumulate one per failure next to a live store.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	for _, e := range entries {
		if e.Name() != "store.dsnap" && e.Name() != "subdir" {
			t.Errorf("leftover temp file after failed save: %q", e.Name())
		}
	}

	// A successful save replaces the contents and keeps the requested mode.
	replacement := []byte("the new snapshot")
	if err := writeFileAtomic(path, replacement, 0644); err != nil {
		t.Fatalf("writeFileAtomic (replace): %v", err)
	}
	if got, err = os.ReadFile(path); err != nil {
		t.Fatalf("reading replaced file: %v", err)
	}
	if !bytes.Equal(got, replacement) {
		t.Errorf("file not replaced: got %q, want %q", got, replacement)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode().Perm() != 0644 {
		t.Errorf("mode not preserved through rename: got %v, want 0644", info.Mode().Perm())
	}
}

// TestSaveToDiskSurvivesConcurrentReaders is the property a torn write breaks in
// practice: a reader opening the snapshot at an arbitrary moment during a run of
// saves always sees a complete, decodable file.
func TestSaveToDiskSurvivesConcurrentReaders(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	ds := newTestStore(nil)
	ds.persistPath = filepath.Join(dir, "store.dsnap")
	for i := 0; i < 2000; i++ {
		ds.data[fmt.Sprintf("key%d", i)] = strings.Repeat("x", 256)
	}

	done := make(chan struct{})
	var saveErr error
	go func() {
		defer close(done)
		for i := 0; i < 40; i++ {
			if err := ds.saveToDisk(); err != nil {
				saveErr = err
				return
			}
		}
	}()

	reads, partial := 0, 0
	for {
		select {
		case <-done:
			if saveErr != nil {
				t.Fatalf("saveToDisk: %v", saveErr)
			}
			if reads == 0 {
				t.Skip("no reads landed during the save loop")
			}
			if partial > 0 {
				t.Errorf("%d of %d reads saw a torn snapshot", partial, reads)
			}
			return
		default:
		}
		raw, err := os.ReadFile(ds.persistPath)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		reads++
		if _, _, _, err := ds.decodeSnapshot(raw); err != nil {
			partial++
		}
	}
}

// TestSaveFollowsSymlinks pins behaviour an in-place write got for free and a
// rename does not: a persist path that is a symlink must be written *through*.
// Replacing the link with a regular file would strand the real store wherever
// the link pointed -- typically another volume, which is why people symlink it.
func TestSaveFollowsSymlinks(t *testing.T) {
	t.Parallel()

	t.Run("existing target", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		target := filepath.Join(dir, "real.dsnap")
		link := filepath.Join(dir, "link.dsnap")

		if err := os.WriteFile(target, []byte("old"), 0644); err != nil {
			t.Fatalf("seed target: %v", err)
		}
		if err := os.Symlink(target, link); err != nil {
			t.Fatalf("symlink: %v", err)
		}

		if err := writeFileAtomic(link, []byte("new contents"), 0644); err != nil {
			t.Fatalf("writeFileAtomic: %v", err)
		}

		info, err := os.Lstat(link)
		if err != nil {
			t.Fatalf("lstat link: %v", err)
		}
		if info.Mode()&os.ModeSymlink == 0 {
			t.Error("the save replaced the symlink with a regular file")
		}
		got, err := os.ReadFile(target)
		if err != nil {
			t.Fatalf("read target: %v", err)
		}
		if string(got) != "new contents" {
			t.Errorf("target contents: got %q, want %q", got, "new contents")
		}
	})

	// A link whose target does not exist yet is an ordinary first run: the
	// O_TRUNC|O_CREATE write created the target, and so must this one.
	t.Run("dangling target", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		target := filepath.Join(dir, "not-yet.dsnap")
		link := filepath.Join(dir, "link.dsnap")

		if err := os.Symlink(target, link); err != nil {
			t.Fatalf("symlink: %v", err)
		}
		if err := writeFileAtomic(link, []byte("created"), 0644); err != nil {
			t.Fatalf("writeFileAtomic: %v", err)
		}

		got, err := os.ReadFile(target)
		if err != nil {
			t.Fatalf("read target: %v", err)
		}
		if string(got) != "created" {
			t.Errorf("target contents: got %q", got)
		}
	})

	t.Run("chain of links", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		target := filepath.Join(dir, "real.dsnap")
		mid := filepath.Join(dir, "mid.dsnap")
		link := filepath.Join(dir, "link.dsnap")

		if err := os.WriteFile(target, []byte("old"), 0644); err != nil {
			t.Fatalf("seed: %v", err)
		}
		// Relative link, to exercise resolution against the link's own directory.
		if err := os.Symlink("real.dsnap", mid); err != nil {
			t.Fatalf("symlink mid: %v", err)
		}
		if err := os.Symlink(mid, link); err != nil {
			t.Fatalf("symlink link: %v", err)
		}

		if err := writeFileAtomic(link, []byte("through both"), 0644); err != nil {
			t.Fatalf("writeFileAtomic: %v", err)
		}

		got, err := os.ReadFile(target)
		if err != nil {
			t.Fatalf("read target: %v", err)
		}
		if string(got) != "through both" {
			t.Errorf("target contents: got %q", got)
		}
	})
}

// TestSaveToDiskLeavesNoTempFiles: the temp file is an implementation detail. If
// one survives a save it survives every save, silting up the data directory next
// to a live store.
func TestSaveToDiskLeavesNoTempFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	ds := newTestStore(nil)
	ds.persistPath = filepath.Join(dir, "store.dsnap")
	ds.data["k"] = "v"

	for i := 0; i < 20; i++ {
		if err := ds.saveToDisk(); err != nil {
			t.Fatalf("saveToDisk: %v", err)
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	for _, e := range entries {
		if e.Name() != "store.dsnap" {
			t.Errorf("save left a file behind: %q", e.Name())
		}
	}
}

// TestSaveToDiskBareFilename covers a persist path with no directory component,
// which is what a script that just says persist = "store.db" produces. The temp
// file has to land in the working directory, not in "".
func TestSaveToDiskBareFilename(t *testing.T) {
	dir := t.TempDir()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer os.Chdir(prev)

	ds := newTestStore(nil)
	ds.persistPath = "store.dsnap"
	ds.data["k"] = "v"

	if err := ds.saveToDisk(); err != nil {
		t.Fatalf("saveToDisk: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(dir, "store.dsnap"))
	if err != nil {
		t.Fatalf("read snapshot: %v", err)
	}
	data, _, _, err := ds.decodeSnapshot(raw)
	if err != nil {
		t.Fatalf("decodeSnapshot: %v", err)
	}
	if data["k"] != "v" {
		t.Errorf("value did not round-trip: got %#v", data["k"])
	}
}
