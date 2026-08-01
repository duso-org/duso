package runtime

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/duso-org/duso/pkg/script"
)

func newLimitTestStore(max int64) *DatastoreValue {
	return &DatastoreValue{
		namespace:          "test",
		data:               make(map[string]any),
		conditions:         make(map[string]*sync.Cond),
		expiryTimes:        make(map[string]time.Time),
		expiryHeap:         make(ExpiryHeap, 0),
		returnDeletedValue: true,
		maxValueSize:       max,
	}
}

func TestMaxValueSizeRejectsOversizedWrites(t *testing.T) {
	t.Parallel()

	big := strings.Repeat("x", 4096)

	t.Run("set", func(t *testing.T) {
		t.Parallel()
		ds := newLimitTestStore(1024)
		if err := ds.Set("k", big); err == nil {
			t.Fatal("expected set to be rejected")
		}
		if _, exists := ds.data["k"]; exists {
			t.Error("rejected value was stored anyway")
		}
	})

	t.Run("swap", func(t *testing.T) {
		t.Parallel()
		ds := newLimitTestStore(1024)
		if _, err := ds.Swap("k", big); err == nil {
			t.Fatal("expected swap to be rejected")
		}
	})

	t.Run("push", func(t *testing.T) {
		t.Parallel()
		ds := newLimitTestStore(1024)
		if _, err := ds.Push("q", big); err == nil {
			t.Fatal("expected push to be rejected")
		}
	})

	t.Run("unshift", func(t *testing.T) {
		t.Parallel()
		ds := newLimitTestStore(1024)
		if _, err := ds.Unshift("q", big); err == nil {
			t.Fatal("expected unshift to be rejected")
		}
	})

	t.Run("update", func(t *testing.T) {
		t.Parallel()
		ds := newLimitTestStore(1024)
		if _, err := ds.Update("k", map[string]any{"field": big}); err == nil {
			t.Fatal("expected update to be rejected")
		}
	})

	// A rejected set_once must not look like an ordinary "key already existed".
	t.Run("set_once reports rejection distinctly", func(t *testing.T) {
		t.Parallel()
		ds := newLimitTestStore(1024)

		ok, err := ds.SetOnce("k", big)
		if err == nil {
			t.Fatal("expected an error for an oversized set_once")
		}
		if ok {
			t.Error("rejected set_once reported success")
		}

		ok, err = ds.SetOnce("k", "small")
		if err != nil || !ok {
			t.Fatalf("genuine set_once failed: ok=%v err=%v", ok, err)
		}
		ok, err = ds.SetOnce("k", "small again")
		if err != nil {
			t.Fatalf("set_once on an existing key should not error: %v", err)
		}
		if ok {
			t.Error("set_once overwrote an existing key")
		}
	})
}

func TestMaxValueSizeAllowsWritesUnderLimit(t *testing.T) {
	t.Parallel()

	ds := newLimitTestStore(1024)
	if err := ds.Set("k", strings.Repeat("x", 100)); err != nil {
		t.Fatalf("write under the limit was rejected: %v", err)
	}
	if _, err := ds.Push("q", map[string]any{"id": 1.0}); err != nil {
		t.Fatalf("push under the limit was rejected: %v", err)
	}
}

func TestMaxValueSizeZeroDisablesCheck(t *testing.T) {
	t.Parallel()

	ds := newLimitTestStore(0)
	if err := ds.Set("k", strings.Repeat("x", 1<<20)); err != nil {
		t.Fatalf("max_value_size=0 should disable the check, got: %v", err)
	}
}

// The cap bounds the item appended, not the whole array: sizing the array on
// every push would reintroduce O(n²) queue filling.
func TestMaxValueSizeBoundsPushItemNotArray(t *testing.T) {
	t.Parallel()

	ds := newLimitTestStore(1024)
	item := strings.Repeat("x", 200)
	for i := 0; i < 50; i++ {
		if _, err := ds.Push("q", item); err != nil {
			t.Fatalf("push %d rejected even though each item is under the limit: %v", i, err)
		}
	}
	if arr, ok := ds.data["q"].([]any); !ok || len(arr) != 50 {
		t.Errorf("expected 50 queued items, got %#v", ds.data["q"])
	}
}

// Replay must not enforce the cap: data already committed has to load even if
// the limit was lowered afterwards.
func TestMaxValueSizeNotEnforcedOnReplay(t *testing.T) {
	t.Parallel()

	ds := newLimitTestStore(10)
	recs := []walRecord{
		{Op: opSet, Key: "big", Value: strings.Repeat("x", 4096)},
		{Op: opPush, Key: "q", Value: strings.Repeat("y", 4096)},
	}
	for i := range recs {
		if err := ds.applyWALRecord(&recs[i]); err != nil {
			t.Fatalf("replay rejected already-committed data: %v", err)
		}
	}
	if s, _ := ds.data["big"].(string); len(s) != 4096 {
		t.Errorf("replayed value was altered: %d bytes", len(s))
	}
}

// A mistyped option used to be skipped in silence, which is the worst possible
// outcome: a store the caller believes is durable, rate-limited or read-only,
// and isn't.
func TestConfigRejectsWrongTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config map[string]any
	}{
		{"persist not a string", map[string]any{"persist": 42.0}},
		{"wal not a string", map[string]any{"wal": true}},
		{"persist_interval not a number", map[string]any{"persist_interval": "60s"}},
		{"persist_interval negative", map[string]any{"persist_interval": -1.0}},
		{"wal_sync_interval not a number", map[string]any{"wal_sync_interval": "1s"}},
		{"wal_sync_interval negative", map[string]any{"wal_sync_interval": -1.0}},
		{"readonly not a bool", map[string]any{"readonly": "yes"}},
		{"return_deleted_value not a bool", map[string]any{"return_deleted_value": 1.0}},
		{"max_value_size not a number", map[string]any{"max_value_size": "64MB"}},
		{"encrypt_key not a string", map[string]any{"encrypt_key": 12345.0}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if err := applyDatastoreConfig(newLimitTestStore(defaultMaxValueSize), tc.config); err == nil {
				t.Fatalf("expected %v to be rejected, got nil", tc.config)
			}
		})
	}
}

// The default is batched, not per-write. Fsyncing every write costs ~1000x
// throughput and only protects against losing the machine, not the process.
func TestWALSyncIntervalDefaultsToBatched(t *testing.T) {
	t.Parallel()

	store := GetDatastore(uniqueNamespace("sync_default"), nil)
	if store.walSyncInterval != defaultWALSyncInterval {
		t.Errorf("default wal_sync_interval: got %v, want %v", store.walSyncInterval, defaultWALSyncInterval)
	}
	if store.walSyncInterval <= 0 {
		t.Error("default must be batched; 0 means fsync on every write")
	}

	// An explicit 0 still selects fsync-per-write.
	if err := applyDatastoreConfig(store, map[string]any{"wal_sync_interval": 0.0}); err != nil {
		t.Fatalf("explicit 0 was rejected: %v", err)
	}
	if store.walSyncInterval != 0 {
		t.Errorf("explicit wal_sync_interval=0: got %v, want 0", store.walSyncInterval)
	}
}

func TestMaxValueSizeConfigValidation(t *testing.T) {
	t.Parallel()

	t.Run("accepts a number", func(t *testing.T) {
		t.Parallel()
		ds := newLimitTestStore(defaultMaxValueSize)
		if err := applyDatastoreConfig(ds, map[string]any{"max_value_size": 2048.0}); err != nil {
			t.Fatalf("valid config rejected: %v", err)
		}
		if ds.maxValueSize != 2048 {
			t.Errorf("maxValueSize: got %d, want 2048", ds.maxValueSize)
		}
	})

	t.Run("rejects a non-number", func(t *testing.T) {
		t.Parallel()
		ds := newLimitTestStore(defaultMaxValueSize)
		if err := applyDatastoreConfig(ds, map[string]any{"max_value_size": "64MB"}); err == nil {
			t.Fatal("expected a non-numeric max_value_size to be rejected")
		}
	})

	t.Run("rejects a negative", func(t *testing.T) {
		t.Parallel()
		ds := newLimitTestStore(defaultMaxValueSize)
		if err := applyDatastoreConfig(ds, map[string]any{"max_value_size": -1.0}); err == nil {
			t.Fatal("expected a negative max_value_size to be rejected")
		}
	})
}

// A large binary is the case the cap exists for, and the one the earlier
// limits.md sketch would have measured as zero bytes.
func TestValueSizeMeasuresBinary(t *testing.T) {
	t.Parallel()

	payload := make([]byte, 5000)
	ref := &ValueRef{Val: script.NewBinary(payload)}

	if got := script.ValueSize(ref); got < 5000 {
		t.Errorf("binary sized as %d bytes, want at least 5000", got)
	}

	ds := newLimitTestStore(1024)
	if err := ds.Set("file", ref); err == nil {
		t.Fatal("expected an oversized binary to be rejected")
	}
}

// The estimate has to track the real encoding closely enough to enforce against.
func TestValueSizeTracksEncodedSize(t *testing.T) {
	t.Parallel()

	cases := []any{
		"a string value",
		strings.Repeat("x", 10000),
		map[string]any{"a": 1.0, "b": "two", "c": []any{1.0, 2.0, 3.0}},
		[]any{1.0, "two", true, nil},
		&ValueRef{Val: script.NewBinary(make([]byte, 2048))},
	}

	for _, v := range cases {
		encoded, err := script.EncodeValue(nil, v)
		if err != nil {
			t.Fatalf("EncodeValue failed: %v", err)
		}
		estimate := script.ValueSize(v)
		actual := int64(len(encoded))

		// The estimate must never undercount, or the cap could be bypassed.
		if estimate < actual {
			t.Errorf("estimate %d undercounts actual encoded size %d for %T", estimate, actual, v)
		}
		// And it must stay in the same ballpark to be a meaningful limit.
		if estimate > actual*4+64 {
			t.Errorf("estimate %d wildly overcounts actual %d for %T", estimate, actual, v)
		}
	}
}
