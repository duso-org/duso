package runtime

import (
	"fmt"
	"path/filepath"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
)

// The invariant these tests exist for: **the log must describe the store you
// have.** Every mutation writes its WAL record while holding dataMutex, so log
// order matches the order writes reach memory. If that ever stops being true,
// recovery reconstructs a store that never existed — and because an operation
// log is not idempotent, the damage is silent rather than loud.
//
// Each test drives concurrent traffic through a WAL-backed store, then replays
// the log into a second store and requires the two to be identical. Whatever the
// racing writers happened to do, replay has to reproduce it.
//
// These exist alongside test/test_wal_concurrency.du, which covers the same
// invariant end to end through spawn(). The two are not redundant: the script
// test drives the real path a Duso program takes, while these run under
// `go test -race`, which is the only way to catch an actual data race in the
// datastore. Keep both.
//
// Writes here are batched (wal_sync_interval=1) rather than fsynced per write.
// The comparison reads the same file on the same machine, so the page cache
// makes the records visible either way, and fsyncing every write made this suite
// take 88 seconds instead of a few.

// nsCounter keeps namespaces unique; datastores are registered globally.
var nsCounter atomic.Uint64

func uniqueNamespace(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, nsCounter.Add(1))
}

// newDurableStore creates a WAL-backed store. Records are written to the file
// immediately; only the fsync is batched, which is all these tests need.
func newDurableStore(t *testing.T, dir, name string) (*DatastoreValue, string, string) {
	t.Helper()

	snapPath := filepath.Join(dir, name+".dusnap")
	walPath := filepath.Join(dir, name+".duwal")

	store := GetDatastore(uniqueNamespace(name), nil)
	err := applyDatastoreConfig(store, map[string]any{
		"persist":           snapPath,
		"wal":               walPath,
		"wal_sync_interval": 1.0,
	})
	if err != nil {
		t.Fatalf("configuring the writer store failed: %v", err)
	}
	return store, snapPath, walPath
}

// recoverFrom loads a fresh store from the same files. readonly keeps it from
// truncating the log the writer still holds open, so the comparison is against
// the log exactly as it was written.
func recoverFrom(t *testing.T, snapPath, walPath string) *DatastoreValue {
	t.Helper()

	store := GetDatastore(uniqueNamespace("recovered"), nil)
	err := applyDatastoreConfig(store, map[string]any{
		"persist":  snapPath,
		"wal":      walPath,
		"readonly": true,
	})
	if err != nil {
		t.Fatalf("recovery failed: %v", err)
	}
	return store
}

// snapshotData copies a store's contents so it can be compared after the fact.
func snapshotData(ds *DatastoreValue) map[string]any {
	ds.dataMutex.RLock()
	defer ds.dataMutex.RUnlock()

	out := make(map[string]any, len(ds.data))
	for k, v := range ds.data {
		out[k] = DeepCopyAny(v)
	}
	return out
}

// requireSameState reports the first divergence rather than dumping two large
// maps, since a single wrong key is the realistic failure.
func requireSameState(t *testing.T, want, got map[string]any) {
	t.Helper()

	if len(want) != len(got) {
		t.Errorf("key count differs: memory has %d, recovered has %d", len(want), len(got))
	}
	for k, wantVal := range want {
		gotVal, ok := got[k]
		if !ok {
			t.Errorf("key %q is in memory but missing after recovery (memory value: %#v)", k, wantVal)
			continue
		}
		if !reflect.DeepEqual(wantVal, gotVal) {
			t.Errorf("key %q diverged:\n  memory:    %#v\n  recovered: %#v", k, wantVal, gotVal)
		}
	}
	for k := range got {
		if _, ok := want[k]; !ok {
			t.Errorf("key %q appeared during recovery but is not in memory: %#v", k, got[k])
		}
	}
}

// Mixed concurrent traffic — the test that covers the lock-ordering change.
// Logging outside dataMutex would let two writers land in one order on disk and
// the opposite order in memory; that shows up here as a diverged key.
func TestConcurrentWritesRecoverExactly(t *testing.T) {
	const (
		workers      = 8
		opsPerWorker = 150
	)

	store, snapPath, walPath := newDurableStore(t, t.TempDir(), "mixed")

	// Seed the keys that workers will contend over, so operations that require
	// an existing value have one.
	if err := store.Set("shared_counter", 0.0); err != nil {
		t.Fatalf("seed failed: %v", err)
	}
	if err := store.Set("shared_object", map[string]any{"base": 1.0}); err != nil {
		t.Fatalf("seed failed: %v", err)
	}
	if _, err := store.Push("shared_queue", "seed"); err != nil {
		t.Fatalf("seed failed: %v", err)
	}

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()

			for i := 0; i < opsPerWorker; i++ {
				own := fmt.Sprintf("worker_%d_key_%d", worker, i%7)

				switch i % 7 {
				case 0:
					_ = store.Set(own, map[string]any{"worker": float64(worker), "i": float64(i)})
				case 1:
					_, _ = store.Increment("shared_counter", 1)
				case 2:
					_, _ = store.Push("shared_queue", map[string]any{"w": float64(worker), "i": float64(i)})
				case 3:
					// May find the queue empty; that path logs nothing, which is
					// itself part of what recovery has to agree with.
					_, _ = store.Shift("shared_queue")
				case 4:
					_, _ = store.Update("shared_object", map[string]any{
						fmt.Sprintf("w%d", worker): float64(i),
					})
				case 5:
					_, _ = store.Swap(own, fmt.Sprintf("swapped_%d_%d", worker, i))
				case 6:
					_, _ = store.Delete(own)
				}
			}
		}(w)
	}
	wg.Wait()

	want := snapshotData(store)
	if len(want) == 0 {
		t.Fatal("test did no work: store is empty")
	}

	got := snapshotData(recoverFrom(t, snapPath, walPath))
	requireSameState(t, want, got)
}

// Counters are the case where a lost or doubled record is silent: the value is
// still a plausible number. The total is deterministic even though the
// interleaving is not, so recovery has an exact target to hit.
func TestConcurrentIncrementsRecoverExactly(t *testing.T) {
	const (
		workers      = 8
		opsPerWorker = 250
	)

	store, snapPath, walPath := newDurableStore(t, t.TempDir(), "counters")

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < opsPerWorker; i++ {
				if _, err := store.Increment("hits", 1); err != nil {
					t.Errorf("increment failed: %v", err)
					return
				}
			}
		}()
	}
	wg.Wait()

	expected := float64(workers * opsPerWorker)
	if got := snapshotData(store)["hits"]; got != expected {
		t.Fatalf("in-memory counter is already wrong: got %v, want %v", got, expected)
	}

	recovered := snapshotData(recoverFrom(t, snapPath, walPath))
	if got := recovered["hits"]; got != expected {
		t.Errorf("counter after recovery: got %v, want %v — a delta was lost or applied twice", got, expected)
	}
}

// Producers and consumers through the blocking path. shift_wait() takes an item
// while holding the lock and logs from inside that critical section; nothing
// else exercises that branch.
func TestConcurrentBlockingQueueRecoversExactly(t *testing.T) {
	const (
		producers   = 4
		consumers   = 4
		perProducer = 100
	)

	store, snapPath, walPath := newDurableStore(t, t.TempDir(), "blocking")

	var consumed atomic.Int64
	var wg sync.WaitGroup

	for c := 0; c < consumers; c++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				if consumed.Load() >= int64(producers*perProducer) {
					return
				}
				// Short timeout so the test cannot hang if a producer dies.
				item, err := store.ShiftWait(nil, "jobs", 500_000_000) // 500ms
				if err != nil {
					return
				}
				if item == nil {
					continue // timed out with an empty queue
				}
				consumed.Add(1)
			}
		}()
	}

	for p := 0; p < producers; p++ {
		wg.Add(1)
		go func(producer int) {
			defer wg.Done()
			for i := 0; i < perProducer; i++ {
				if _, err := store.Push("jobs", map[string]any{
					"producer": float64(producer),
					"i":        float64(i),
				}); err != nil {
					t.Errorf("push failed: %v", err)
					return
				}
			}
		}(p)
	}

	wg.Wait()

	if got := consumed.Load(); got != int64(producers*perProducer) {
		t.Errorf("consumed %d items, want %d", got, producers*perProducer)
	}

	want := snapshotData(store)
	got := snapshotData(recoverFrom(t, snapPath, walPath))
	requireSameState(t, want, got)

	// The point of logging shift_wait at all: consumed work must not come back.
	if arr, ok := got["jobs"].([]any); ok && len(arr) != 0 {
		t.Errorf("recovery resurrected %d consumed jobs", len(arr))
	}
}

// Two callers racing to configure the same namespace: exactly one may win, and
// the loser must be told rather than silently re-running recovery.
func TestConcurrentConfigureAllowsExactlyOne(t *testing.T) {
	const racers = 8

	dir := t.TempDir()
	namespace := uniqueNamespace("race_config")
	store := GetDatastore(namespace, nil)

	config := map[string]any{
		"persist": filepath.Join(dir, "race.dusnap"),
		"wal":     filepath.Join(dir, "race.duwal"),
	}

	var succeeded atomic.Int64
	var failed atomic.Int64

	var start sync.WaitGroup
	start.Add(1)

	var wg sync.WaitGroup
	for i := 0; i < racers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			start.Wait() // release them together
			if err := applyDatastoreConfig(store, config); err != nil {
				failed.Add(1)
			} else {
				succeeded.Add(1)
			}
		}()
	}
	start.Done()
	wg.Wait()

	if succeeded.Load() != 1 {
		t.Errorf("%d callers configured the store; exactly 1 may", succeeded.Load())
	}
	if failed.Load() != racers-1 {
		t.Errorf("%d callers were rejected, want %d", failed.Load(), racers-1)
	}
}

// A snapshot taken while writers are running truncates the log and records the
// watermark it covers. If that watermark is wrong, replay either skips records
// the snapshot never had or re-applies ones it did.
func TestSnapshotDuringConcurrentWritesRecoversExactly(t *testing.T) {
	const (
		workers      = 6
		opsPerWorker = 150
	)

	store, snapPath, walPath := newDurableStore(t, t.TempDir(), "snapmid")

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for i := 0; i < opsPerWorker; i++ {
				_ = store.Set(fmt.Sprintf("k_%d_%d", worker, i), map[string]any{"i": float64(i)})
				if i%3 == 0 {
					_, _ = store.Increment("total", 1)
				}
			}
		}(w)
	}

	// Snapshot repeatedly while the writers are mid-flight.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 5; i++ {
			if err := store.saveToDisk(); err != nil {
				t.Errorf("snapshot during writes failed: %v", err)
				return
			}
		}
	}()

	wg.Wait()

	want := snapshotData(store)
	got := snapshotData(recoverFrom(t, snapPath, walPath))
	requireSameState(t, want, got)
}
