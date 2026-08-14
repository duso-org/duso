package runtime

import (
	"bytes"
	"container/heap"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/duso-org/duso/pkg/core"
	"github.com/duso-org/duso/pkg/script"
)

// Global registry of namespaced datastores
var (
	datastoreRegistry = make(map[string]*DatastoreValue)
	registryMutex     sync.RWMutex
)

// Register gob types once at init
func init() {
	gob.Register(WALEntry{})
	gob.Register([]any{})
	gob.Register(map[string]any{})
}

// WALEntry represents a key-value write in the Write-Ahead Log
type WALEntry struct {
	Key   string
	Value any
}

// ExpiryEntry represents a key and its expiration time in the min-heap
type ExpiryEntry struct {
	key        string
	expiryTime time.Time
}

// ExpiryHeap implements container/heap.Interface for a min-heap sorted by expiryTime
type ExpiryHeap []ExpiryEntry

func (h ExpiryHeap) Len() int           { return len(h) }
func (h ExpiryHeap) Less(i, j int) bool { return h[i].expiryTime.Before(h[j].expiryTime) }
func (h ExpiryHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *ExpiryHeap) Push(x any)        { *h = append(*h, x.(ExpiryEntry)) }
func (h *ExpiryHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// DatastoreValue represents an in-memory thread-safe key/value store
// scoped to a specific namespace. Multiple scripts can access the same
// store by using the same namespace. Optionally persists to JSON and/or WAL.
type DatastoreValue struct {
	namespace          string
	data               map[string]any
	dataMutex          sync.RWMutex
	conditions         map[string]*sync.Cond // Per-key condition variables for wait operations
	persistPath        string                // Optional: path to JSON file
	persistInterval    time.Duration         // Optional: auto-save interval
	ticker             *time.Ticker          // Auto-save ticker
	stopTicker         chan bool             // Signal to stop ticker
	expiryTicker       *time.Ticker          // Expiry sweep ticker
	fileWriteMutex     sync.Mutex            // Serialize file writes
	statsFn            func(key string) any  // Function to compute stats dynamically (for sys datastore)
	expiryTimes        map[string]time.Time  // Quick lookup: when does each key expire?
	expiryHeap         ExpiryHeap            // Min-heap sorted by expiration time
	expiryStopTicker   chan bool             // Signal to stop expiry sweep ticker
	defaultExpiryTTL   time.Duration         // Default TTL for expired keys (60 minutes)
	readonly           bool                  // If true, builtin write operations are forbidden
	returnDeletedValue bool                  // If true, delete() returns the deleted value
	walPath            string                // Optional: path to WAL file
	walFile            *os.File              // Open WAL file handle
	walEncoder         *gob.Encoder          // WAL encoder for writing entries
	walMutex           sync.Mutex            // Protect concurrent WAL writes
	walSyncInterval    time.Duration         // 0=sync every write, >0=batch writes
	walSyncTicker      *time.Ticker          // Periodic WAL sync (if batching)
	walStopSync        chan bool             // Signal to stop WAL sync ticker
	encryptKey         []byte                // Optional: 32-byte AES-256 key for encrypting snapshot/WAL files
	walSeq             atomic.Uint64         // Monotonic op-log sequence; snapshots record the watermark they cover
	configMutex        sync.Mutex            // Serializes configuration so two racing callers can't both recover
	configured         bool                  // Configuration and recovery have already run for this store
	maxValueSize       int64                 // Largest value a single write may store; 0 disables the check
	repl               *replState            // Replication role and state; nil when the store is standalone
}

// replEpoch is the leadership term to stamp into file headers. Zero for a store
// that has never replicated, which is what every pre-replication file already
// carries in that position.
func (ds *DatastoreValue) replEpoch() uint32 {
	if ds.repl == nil {
		return 0
	}
	return ds.repl.epoch
}

// writeGuard reports why this store rejects mutations, or nil if it accepts
// them. A follower is not rejected here: its mutations are forwarded to the
// leader by the methods themselves.
func (ds *DatastoreValue) writeGuard() error {
	if ds.readonly {
		return fmt.Errorf("datastore(%q) is read-only", ds.namespace)
	}
	return nil
}

// fileGuard covers save() and load(), which act on this machine's files rather
// than on the data, and so are never forwarded. load() in particular would
// overwrite replicated state while the cursor kept pointing at the leader's
// sequence, leaving memory and cursor disagreeing with nothing to notice it.
func (ds *DatastoreValue) fileGuard(op string) error {
	if err := ds.writeGuard(); err != nil {
		return err
	}
	if ds.isFollower() {
		return fmt.Errorf("datastore(%q): %s() is not available on a replication follower — its snapshot is written from the stream",
			ds.namespace, op)
	}
	return nil
}

// defaultWALSyncInterval is how often the log is fsynced when the caller does
// not say otherwise.
//
// Records are written to the file as each operation completes; only the fsync is
// batched. That means a process crash — a panic, an OOM kill, a bad deploy, by
// far the common case — loses nothing, because the pages are already handed to
// the kernel. Only losing the machine itself (power, kernel panic, hypervisor)
// can cost the last interval's writes.
//
// Fsyncing every write instead costs roughly three orders of magnitude of write
// throughput (~300/sec vs ~400,000/sec on an M1), because throughput is then
// bound by fsync latency rather than by anything Duso does. 100ms was measured
// as indistinguishable from Redis's 1-second default in throughput while
// exposing a tenth as much, so there is no reason to wait longer.
//
// Set wal_sync_interval = 0 to fsync every write.
const defaultWALSyncInterval = 100 * time.Millisecond

// defaultMaxValueSize caps what one write may put under a key.
//
// Binary values persist properly now, which removed the accidental wall that
// used to stop them reaching disk at all. Without a cap, a stray upload becomes
// an fsync of that size on every write and a rewrite of that size in every
// snapshot. 64MB is generous for uploads, images and thumbnails while still
// failing a runaway immediately on the small VMs Duso targets.
const defaultMaxValueSize int64 = 64 * 1024 * 1024

// applyDatastoreConfig applies configuration to a datastore and triggers recovery.
// IMPORTANT: Paths in config must be pre-resolved (caller is responsible).
//
// Every failure here is returned rather than logged. A store that cannot load
// its snapshot, replay its WAL, or open its WAL for writing is not a degraded
// store — it is one that will overwrite good data at the next save, or drop
// writes that the caller believes are durable. Failing the datastore() call
// makes that a startup failure instead of a line in the log.
func applyDatastoreConfig(store *DatastoreValue, config map[string]any) error {
	if config == nil {
		return nil
	}

	// A datastore is configured in exactly one place. Serialized so two callers
	// racing to configure the same namespace cannot both run recovery.
	store.configMutex.Lock()
	defer store.configMutex.Unlock()

	if store.configured {
		return fmt.Errorf("datastore(%q) is already configured — configure a datastore in one place and open it elsewhere with datastore(%q)",
			store.namespace, store.namespace)
	}

	// Apply config options - paths must already be resolved by caller.
	// Every option is type-checked rather than silently skipped: a mistyped
	// persist or wal path used to leave the store in memory only, with the
	// caller believing it was durable.
	if persistPath, ok := config["persist"]; ok {
		p, ok := persistPath.(string)
		if !ok {
			return fmt.Errorf("datastore(%q): persist must be a file path string", store.namespace)
		}
		store.persistPath = p
	}
	// A bad value here used to be ignored, leaving auto-save silently switched
	// off while the caller believed they had configured it.
	if persistInterval, ok := config["persist_interval"]; ok {
		intervalSecs, ok := persistInterval.(float64)
		if !ok {
			return fmt.Errorf("datastore(%q): persist_interval must be a number of seconds", store.namespace)
		}
		if intervalSecs < 0 {
			return fmt.Errorf("datastore(%q): persist_interval cannot be negative", store.namespace)
		}
		// Convert seconds (as float64) to nanoseconds as int64
		store.persistInterval = time.Duration(int64(intervalSecs*1e9)) * time.Nanosecond
	}
	if walPath, ok := config["wal"]; ok {
		w, ok := walPath.(string)
		if !ok {
			return fmt.Errorf("datastore(%q): wal must be a file path string", store.namespace)
		}
		store.walPath = w
	}
	// Only an explicit setting overrides the default; 0 here means the caller
	// asked for an fsync on every write.
	if walSyncInterval, ok := config["wal_sync_interval"]; ok {
		intervalSecs, ok := walSyncInterval.(float64)
		if !ok {
			return fmt.Errorf("datastore(%q): wal_sync_interval must be a number of seconds", store.namespace)
		}
		if intervalSecs < 0 {
			return fmt.Errorf("datastore(%q): wal_sync_interval cannot be negative", store.namespace)
		}
		store.walSyncInterval = time.Duration(int64(intervalSecs*1e9)) * time.Nanosecond
	}
	if readonly, ok := config["readonly"]; ok {
		r, ok := readonly.(bool)
		if !ok {
			return fmt.Errorf("datastore(%q): readonly must be true or false", store.namespace)
		}
		store.readonly = r
	}
	if returnDeletedValue, ok := config["return_deleted_value"]; ok {
		r, ok := returnDeletedValue.(bool)
		if !ok {
			return fmt.Errorf("datastore(%q): return_deleted_value must be true or false", store.namespace)
		}
		store.returnDeletedValue = r
	}
	if maxValueSize, ok := config["max_value_size"]; ok {
		size, ok := maxValueSize.(float64)
		if !ok {
			return fmt.Errorf("datastore(%q): max_value_size must be a number of bytes", store.namespace)
		}
		if size < 0 {
			return fmt.Errorf("datastore(%q): max_value_size cannot be negative", store.namespace)
		}
		store.maxValueSize = int64(size) // 0 disables the check
	}

	// Handle encryption key (base64-encoded string).
	// A bad key is fatal, not a warning: ignoring it would silently write the
	// store to disk in plaintext when the caller asked for encryption.
	if encryptKey, ok := config["encrypt_key"]; ok {
		keyStr, ok := encryptKey.(string)
		if !ok {
			return fmt.Errorf("datastore(%q): encrypt_key must be a base64-encoded string", store.namespace)
		}
		keyBytes, err := base64.StdEncoding.DecodeString(keyStr)
		if err != nil {
			return fmt.Errorf("datastore(%q): encrypt_key is not valid base64: %v", store.namespace, err)
		}
		if len(keyBytes) != 32 {
			return fmt.Errorf("datastore(%q): encrypt_key must decode to 32 bytes for AES-256, got %d", store.namespace, len(keyBytes))
		}
		store.encryptKey = keyBytes
	}

	// Decide the replication role before any file is read: the role determines
	// the epoch stamped into everything written from here on, and whether this
	// start is a promotion. No sockets open until recovery has finished.
	if err := replConfigure(store, config); err != nil {
		return err
	}

	// Step 1: Load persist if it exists.
	// A missing file is normal (first run) and returns nil. Anything else means
	// the file is there but unreadable — a wrong encrypt_key, a newer format, or
	// corruption — and continuing would start on an empty store and overwrite
	// the real data at the next save.
	if store.persistPath != "" {
		if err := store.loadFromDisk(true); err != nil {
			return fmt.Errorf("datastore(%q): %v", store.namespace, err)
		}
	}

	// Step 2: Replay WAL if it exists. An unreplayable WAL means committed
	// writes are unrecoverable; starting anyway would quietly discard them.
	//
	// A readonly store replays but never writes back: no re-save, no truncate,
	// no opening the log for appends. That makes it safe to point an inspector
	// script at the files of a store another process is actively writing —
	// without it, recovery would truncate that process's WAL out from under it.
	if store.walPath != "" {
		if err := store.replayWALForRecovery(); err != nil {
			return fmt.Errorf("datastore(%q): failed to recover from WAL %q: %v", store.namespace, store.walPath, err)
		}
	}

	// Step 3: Open WAL for new writes if configured. Without this the store runs
	// with durability silently switched off.
	if store.walPath != "" && store.walFile == nil && !store.readonly {
		if err := store.openWALForWrites(); err != nil {
			return fmt.Errorf("datastore(%q): failed to open WAL %q for writes: %v", store.namespace, store.walPath, err)
		}
	}

	store.configured = true

	// Start auto-save ticker if configured (never for a readonly store)
	if store.persistInterval > 0 && store.ticker == nil && !store.readonly {
		store.ticker = time.NewTicker(store.persistInterval)
		go func() {
			defer core.RecoverPanic(fmt.Sprintf("datastore_autosave (namespace=%s)", store.namespace))
			for {
				select {
				case <-store.ticker.C:
					_ = store.saveToDisk()
				case <-store.stopTicker:
					return
				}
			}
		}()
	}

	// Last: the store is now recovered and consistent, so it is safe to serve it
	// to followers or to start consuming a leader's stream.
	if err := replStart(store); err != nil {
		return err
	}

	return nil
}

// GetDatastore returns or creates a namespaced datastore with optional persistence config
func GetDatastore(namespace string, config map[string]any) *DatastoreValue {
	registryMutex.Lock()
	defer registryMutex.Unlock()

	if store, exists := datastoreRegistry[namespace]; exists {
		return store
	}

	store := &DatastoreValue{
		namespace:          namespace,
		data:               make(map[string]any),
		conditions:         make(map[string]*sync.Cond),
		stopTicker:         make(chan bool, 1),
		expiryTimes:        make(map[string]time.Time),
		expiryHeap:         make(ExpiryHeap, 0),
		expiryStopTicker:   make(chan bool, 1),
		defaultExpiryTTL:   60 * time.Minute, // Default 60-minute TTL
		returnDeletedValue: true,             // Default: return deleted values
		maxValueSize:       defaultMaxValueSize,
		walSyncInterval:    defaultWALSyncInterval,
	}

	// Apply namespace defaults
	if namespace == "duso_sys" {
		store.readonly = true
	}
	if namespace == "duso_vfs" {
		store.returnDeletedValue = false // Don't copy large files on delete
	}
	if namespace == "duso_schedule" {
		store.readonly = true // schedule()/unschedule() are the only writers; select()/watch()/wait() still work
	}

	// For duso_sys datastore, set up dynamic metric computation
	// TODO: Implement metrics system properly (currently disabled)
	// if namespace == "duso_sys" {
	//	store.statsFn = GetMetric
	// }

	// Start expiry sweep ticker (1-second sweep)
	store.expiryTicker = time.NewTicker(1 * time.Second)
	go func() {
		defer core.RecoverPanic(fmt.Sprintf("datastore_expiry_sweep (namespace=%s)", store.namespace))
		for {
			select {
			case <-store.expiryTicker.C:
				store.sweepExpiredKeys()
			case <-store.expiryStopTicker:
				store.expiryTicker.Stop()
				return
			}
		}
	}()

	datastoreRegistry[namespace] = store
	return store
}

// GetDatastoreCount returns the number of registered datastores
// Used by system metrics to report datastore count
func GetDatastoreCount() int {
	registryMutex.RLock()
	defer registryMutex.RUnlock()
	return len(datastoreRegistry)
}

// checkValueSize rejects a write whose value exceeds the store's cap.
//
// Checked before the value is copied or encoded: discovering a 500MB value by
// encoding it has already paid most of the cost. Replay deliberately bypasses
// this — data already committed must load even if the cap was lowered since.
func (ds *DatastoreValue) checkValueSize(op, key string, value any) error {
	if ds.maxValueSize <= 0 {
		return nil // check disabled
	}
	if size := script.ValueSize(value); size > ds.maxValueSize {
		return fmt.Errorf("%s(%q): value is %d bytes, over the %d-byte max_value_size for datastore %q",
			op, key, size, ds.maxValueSize, ds.namespace)
	}
	return nil
}

// Set stores a value by key (thread-safe)
func (ds *DatastoreValue) Set(key string, value any) error {
	if ds.isFollower() {
		_, err := ds.forwardSimple(opSet, key, "", 0, value, true)
		return err
	}
	if err := ds.checkValueSize("set", key, value); err != nil {
		return err
	}

	// Deep copy the value to prevent external mutations
	// Handle *[]Value (mutable arrays from script)
	var storedValue any
	if arrPtr, ok := value.(*[]Value); ok {
		// Convert *[]Value to []any for storage
		anyArr := make([]any, len(*arrPtr))
		for i, v := range *arrPtr {
			anyArr[i] = DeepCopyAny(ValueToInterface(v))
		}
		storedValue = anyArr
	} else {
		storedValue = DeepCopyAny(value)
	}

	ds.dataMutex.Lock()

	// Log inside the lock, so log order matches the order writes reach memory.
	// Logging first and locking after lets two concurrent Set calls land in one
	// order on disk and the opposite order in memory, and recovery then produces
	// a store that never existed.
	if err := ds.writeWALOp(opSet, key, "", 0, storedValue); err != nil {
		ds.dataMutex.Unlock()
		return err
	}

	ds.data[key] = storedValue

	// Notify any waiters on this key
	if cond, exists := ds.conditions[key]; exists {
		ds.dataMutex.Unlock()
		cond.Broadcast()
	} else {
		ds.dataMutex.Unlock()
	}

	return nil
}

// SetOnce stores a value by key only if the key doesn't already exist (thread-safe)
// Returns true if the value was set, false if the key already existed
// Useful for caching patterns where multiple concurrent requests might try to set the same key
//
// Returns an error separately from the false result: "the key already existed"
// and "this write was rejected" are different outcomes, and collapsing them into
// one bool hides a failed write behind an ordinary-looking cache miss.
func (ds *DatastoreValue) SetOnce(key string, value any) (bool, error) {
	if ds.isFollower() {
		v, err := ds.forwardSimple(opSetOnce, key, "", 0, value, true)
		stored, _ := v.(bool)
		return stored, err
	}
	if err := ds.checkValueSize("set_once", key, value); err != nil {
		return false, err
	}

	// Deep copy the value to prevent external mutations
	// Handle *[]Value (mutable arrays from script)
	var storedValue any
	if arrPtr, ok := value.(*[]Value); ok {
		// Convert *[]Value to []any for storage
		anyArr := make([]any, len(*arrPtr))
		for i, v := range *arrPtr {
			anyArr[i] = DeepCopyAny(ValueToInterface(v))
		}
		storedValue = anyArr
	} else {
		storedValue = DeepCopyAny(value)
	}

	ds.dataMutex.Lock()

	// Check if key already exists
	if _, exists := ds.data[key]; exists {
		ds.dataMutex.Unlock()
		return false, nil // Key already exists, don't overwrite
	}

	// Logged only on success, so replay stays a blind write
	if err := ds.writeWALOp(opSetOnce, key, "", 0, storedValue); err != nil {
		ds.dataMutex.Unlock()
		return false, err // WAL write failed, don't apply
	}

	ds.data[key] = storedValue

	// Notify any waiters on this key
	if cond, exists := ds.conditions[key]; exists {
		ds.dataMutex.Unlock()
		cond.Broadcast()
	} else {
		ds.dataMutex.Unlock()
	}

	return true, nil // Value was successfully set
}

// Get retrieves a value by key (thread-safe)
func (ds *DatastoreValue) Get(key string) (any, error) {
	ds.dataMutex.Lock()
	defer ds.dataMutex.Unlock()

	// Check for dynamic stats computation (e.g., memory stats)
	if ds.statsFn != nil {
		if val := ds.statsFn(key); val != nil {
			return val, nil
		}
	}

	// Lazy expiry check
	if ds.checkExpired(key) {
		return nil, nil // Key expired
	}

	value, exists := ds.data[key]
	if !exists {
		return nil, nil // Return nil if key doesn't exist
	}

	// Deep copy to isolate returned values from datastore's scope
	// Prevents concurrent requests from accidentally sharing mutable data
	return DeepCopyAny(value), nil
}

// Swap atomically exchanges a key's value for a new value (thread-safe)
// Returns the old value that was at the key
// Useful for consuming inboxes or implementing atomic exchange patterns
func (ds *DatastoreValue) Swap(key string, newValue any) (any, error) {
	if ds.isFollower() {
		return ds.forwardSimple(opSwap, key, "", 0, newValue, true)
	}
	if err := ds.checkValueSize("swap", key, newValue); err != nil {
		return nil, err
	}

	// Deep copy the new value to prevent external mutations
	// Handle *[]Value (mutable arrays from script)
	var storedValue any
	if arrPtr, ok := newValue.(*[]Value); ok {
		// Convert *[]Value to []any for storage
		anyArr := make([]any, len(*arrPtr))
		for i, v := range *arrPtr {
			anyArr[i] = DeepCopyAny(ValueToInterface(v))
		}
		storedValue = anyArr
	} else {
		storedValue = DeepCopyAny(newValue)
	}

	ds.dataMutex.Lock()
	defer ds.dataMutex.Unlock()

	// Logged inside the lock so log order matches apply order (see Set)
	if err := ds.writeWALOp(opSwap, key, "", 0, storedValue); err != nil {
		return nil, err
	}

	// Lazy expiry check
	ds.checkExpired(key)

	// Get the old value
	oldValue, exists := ds.data[key]
	if !exists {
		oldValue = nil
	}

	ds.data[key] = storedValue

	// Notify any waiters on this key
	if cond, exists := ds.conditions[key]; exists {
		cond.Broadcast()
	}

	// Return the old value (deep copied to isolate from datastore's scope)
	return DeepCopyAny(oldValue), nil
}

// Increment atomically increments a numeric value by delta
// Creates the key with value delta if it doesn't exist
func (ds *DatastoreValue) Increment(key string, delta float64) (any, error) {
	if ds.isFollower() {
		return ds.forwardSimple(opIncr, key, "", delta, nil, false)
	}
	ds.dataMutex.Lock()
	defer ds.dataMutex.Unlock()

	current := 0.0
	if val, exists := ds.data[key]; exists {
		// Try to convert existing value to number
		if f, ok := val.(float64); ok {
			current = f
		} else {
			return nil, fmt.Errorf("increment() cannot operate on non-numeric value at key %q", key)
		}
	}

	newValue := current + delta

	// Log the delta, not the result: the op is what replays.
	if err := ds.writeWALOp(opIncr, key, "", delta, nil); err != nil {
		return nil, err
	}

	ds.data[key] = newValue

	// Notify any waiters on this key
	if cond, exists := ds.conditions[key]; exists {
		cond.Broadcast()
	}

	return newValue, nil
}

// Push atomically pushes an item to an array
// Creates the array if key doesn't exist. Returns new array length.
// Returns error if key exists but is not an array.
func (ds *DatastoreValue) Push(key string, item any) (float64, error) {
	if ds.isFollower() {
		v, err := ds.forwardSimple(opPush, key, "", 0, item, true)
		n, _ := toFloat64(v)
		return n, err
	}
	// Bounds the item being appended, not the resulting array — sizing the whole
	// array on every push would make filling a queue O(n²), which is exactly the
	// cost the op-log removed.
	if err := ds.checkValueSize("push", key, item); err != nil {
		return 0, err
	}

	ds.dataMutex.Lock()
	defer ds.dataMutex.Unlock()

	var newArr []any
	if val, exists := ds.data[key]; exists {
		// Key exists - must be an array
		if arr, ok := val.([]any); ok {
			// Deep copy the item before appending
			newArr = append(arr, DeepCopyAny(item))
		} else if _, ok := val.(*[]Value); ok {
			// This shouldn't happen since Set should convert *[]Value to []any
			return 0, fmt.Errorf("push() found unexpected *[]Value at key %q (should be []any)", key)
		} else {
			return 0, fmt.Errorf("push() cannot operate on non-array value at key %q", key)
		}
	} else {
		// Key doesn't exist - create new array with the item
		newArr = []any{DeepCopyAny(item)}
	}

	// Log the appended item only — logging newArr would make every push cost
	// the size of the whole queue.
	if err := ds.writeWALOp(opPush, key, "", 0, item); err != nil {
		return 0, err
	}

	ds.data[key] = newArr

	// Notify any waiters on this key (value changed)
	if cond, exists := ds.conditions[key]; exists {
		cond.Broadcast()
	}

	return float64(len(newArr)), nil
}

// Shift atomically removes and returns the first element from an array
// Returns error if key doesn't exist or is not an array.
// Returns nil if array is empty.
func (ds *DatastoreValue) Shift(key string) (any, error) {
	if ds.isFollower() {
		return ds.forwardSimple(opShift, key, "", 0, nil, false)
	}
	ds.dataMutex.Lock()
	defer ds.dataMutex.Unlock()

	// Lazy expiry check
	if ds.checkExpired(key) {
		return nil, fmt.Errorf("shift() key %q does not exist", key)
	}

	val, exists := ds.data[key]
	if !exists {
		return nil, fmt.Errorf("shift() key %q does not exist", key)
	}

	// Must be an array
	if arr, ok := val.([]any); ok {
		if len(arr) == 0 {
			return nil, nil // Empty array
		}
		item := arr[0]
		newArr := arr[1:]

		// The op alone is enough to replay; the remaining array is not logged.
		if err := ds.writeWALOp(opShift, key, "", 0, nil); err != nil {
			return nil, err
		}

		ds.data[key] = newArr
		// Notify any waiters on this key (value changed)
		if cond, exists := ds.conditions[key]; exists {
			cond.Broadcast()
		}
		return DeepCopyAny(item), nil
	}

	return nil, fmt.Errorf("shift() cannot operate on non-array value at key %q", key)
}

// Pop atomically removes and returns the last element from an array
// Returns error if key doesn't exist or is not an array.
// Returns nil if array is empty.
func (ds *DatastoreValue) Pop(key string) (any, error) {
	if ds.isFollower() {
		return ds.forwardSimple(opPop, key, "", 0, nil, false)
	}
	ds.dataMutex.Lock()
	defer ds.dataMutex.Unlock()

	// Lazy expiry check
	if ds.checkExpired(key) {
		return nil, fmt.Errorf("pop() key %q does not exist", key)
	}

	val, exists := ds.data[key]
	if !exists {
		return nil, fmt.Errorf("pop() key %q does not exist", key)
	}

	// Must be an array
	if arr, ok := val.([]any); ok {
		if len(arr) == 0 {
			return nil, nil // Empty array
		}
		item := arr[len(arr)-1]
		newArr := arr[:len(arr)-1]

		// The op alone is enough to replay; the remaining array is not logged.
		if err := ds.writeWALOp(opPop, key, "", 0, nil); err != nil {
			return nil, err
		}

		ds.data[key] = newArr
		// Notify any waiters on this key (value changed)
		if cond, exists := ds.conditions[key]; exists {
			cond.Broadcast()
		}
		return DeepCopyAny(item), nil
	}

	return nil, fmt.Errorf("pop() cannot operate on non-array value at key %q", key)
}

// ShiftWait atomically removes and returns the first element from an array
// Blocks until array has items or timeout expires
// Returns nil if timeout exceeded and array is still empty
// Returns error if key exists but is not an array
func (ds *DatastoreValue) ShiftWait(procCtx context.Context, key string, timeout time.Duration) (any, error) {
	if ds.isFollower() {
		return ds.replForward(procCtx, &replRequest{
			Op: opReqShiftWait, Key: key, Timeout: timeout.Seconds(),
		}, timeout)
	}
	if procCtx == nil {
		procCtx = context.Background()
	}
	ds.dataMutex.Lock()

	// Get or create condition variable for this key
	cond, exists := ds.conditions[key]
	if !exists {
		cond = sync.NewCond(&ds.dataMutex)
		ds.conditions[key] = cond
	}

	// Child context so kill(pid) (cancelling procCtx) can wake this specific wait
	// early via Broadcast - cheap (one alloc, no channel until Done() is read),
	// and Go's context tree already fans this out correctly to concurrent waiters.
	waitCtx, cancelWait := context.WithCancel(procCtx)
	defer cancelWait()
	var killWatcherArmed bool

	// Loop until we have an item, timeout, or kill()
	for {
		// Check if key exists and is an array with items
		val, keyExists := ds.data[key]
		if keyExists {
			if arr, ok := val.([]any); ok {
				if len(arr) > 0 {
					// We have an item - atomically shift and return it
					item := arr[0]
					if err := ds.writeWALOp(opShift, key, "", 0, nil); err != nil {
						ds.dataMutex.Unlock()
						return nil, err
					}
					ds.data[key] = arr[1:]
					cond.Broadcast()
					ds.dataMutex.Unlock()
					return DeepCopyAny(item), nil
				}
				// Array is empty, keep waiting
			} else {
				// Key exists but is not an array
				ds.dataMutex.Unlock()
				return nil, fmt.Errorf("shift_wait() cannot operate on non-array value at key %q", key)
			}
		}
		// Key doesn't exist or array is empty - wait for change

		if timeout > 0 {
			// Start a goroutine that will broadcast on timeout or kill()
			go func() {
				defer core.RecoverPanic(fmt.Sprintf("datastore_wait_timeout (namespace=%s)", ds.namespace))
				select {
				case <-time.After(timeout):
				case <-waitCtx.Done():
				}
				ds.dataMutex.Lock()
				cond.Broadcast()
				ds.dataMutex.Unlock()
			}()

			// Record start time for checking actual timeout
			startTime := time.Now()
			cond.Wait() // Called with lock held - safe

			if waitCtx.Err() != nil {
				ds.dataMutex.Unlock()
				return nil, fmt.Errorf("killed")
			}
			// Check if we actually timed out
			if time.Since(startTime) >= timeout {
				ds.dataMutex.Unlock()
				return nil, nil // Timeout with no item
			}
			// Otherwise, loop will re-check the condition
		} else {
			// No timeout - arm a one-time watcher for kill(), then just wait
			if !killWatcherArmed {
				killWatcherArmed = true
				go func() {
					defer core.RecoverPanic(fmt.Sprintf("datastore_wait_kill (namespace=%s)", ds.namespace))
					<-waitCtx.Done()
					ds.dataMutex.Lock()
					cond.Broadcast()
					ds.dataMutex.Unlock()
				}()
			}
			cond.Wait()
			if waitCtx.Err() != nil {
				ds.dataMutex.Unlock()
				return nil, fmt.Errorf("killed")
			}
		}
	}
}

// PopWait atomically removes and returns the last element from an array
// Blocks until array has items or timeout expires
// Returns nil if timeout exceeded and array is still empty
// Returns error if key exists but is not an array
func (ds *DatastoreValue) PopWait(procCtx context.Context, key string, timeout time.Duration) (any, error) {
	if ds.isFollower() {
		return ds.replForward(procCtx, &replRequest{
			Op: opReqPopWait, Key: key, Timeout: timeout.Seconds(),
		}, timeout)
	}
	if procCtx == nil {
		procCtx = context.Background()
	}
	ds.dataMutex.Lock()

	// Get or create condition variable for this key
	cond, exists := ds.conditions[key]
	if !exists {
		cond = sync.NewCond(&ds.dataMutex)
		ds.conditions[key] = cond
	}

	// Child context so kill(pid) (cancelling procCtx) can wake this specific wait
	// early via Broadcast - cheap (one alloc, no channel until Done() is read),
	// and Go's context tree already fans this out correctly to concurrent waiters.
	waitCtx, cancelWait := context.WithCancel(procCtx)
	defer cancelWait()
	var killWatcherArmed bool

	// Loop until we have an item, timeout, or kill()
	for {
		// Check if key exists and is an array with items
		val, keyExists := ds.data[key]
		if keyExists {
			if arr, ok := val.([]any); ok {
				if len(arr) > 0 {
					// We have an item - atomically pop and return it
					item := arr[len(arr)-1]
					if err := ds.writeWALOp(opPop, key, "", 0, nil); err != nil {
						ds.dataMutex.Unlock()
						return nil, err
					}
					ds.data[key] = arr[:len(arr)-1]
					cond.Broadcast()
					ds.dataMutex.Unlock()
					return DeepCopyAny(item), nil
				}
				// Array is empty, keep waiting
			} else {
				// Key exists but is not an array
				ds.dataMutex.Unlock()
				return nil, fmt.Errorf("pop_wait() cannot operate on non-array value at key %q", key)
			}
		}
		// Key doesn't exist or array is empty - wait for change

		if timeout > 0 {
			// Start a goroutine that will broadcast on timeout or kill()
			go func() {
				defer core.RecoverPanic(fmt.Sprintf("datastore_wait_timeout (namespace=%s)", ds.namespace))
				select {
				case <-time.After(timeout):
				case <-waitCtx.Done():
				}
				ds.dataMutex.Lock()
				cond.Broadcast()
				ds.dataMutex.Unlock()
			}()

			// Record start time for checking actual timeout
			startTime := time.Now()
			cond.Wait() // Called with lock held - safe

			if waitCtx.Err() != nil {
				ds.dataMutex.Unlock()
				return nil, fmt.Errorf("killed")
			}
			// Check if we actually timed out
			if time.Since(startTime) >= timeout {
				ds.dataMutex.Unlock()
				return nil, nil // Timeout with no item
			}
			// Otherwise, loop will re-check the condition
		} else {
			// No timeout - arm a one-time watcher for kill(), then just wait
			if !killWatcherArmed {
				killWatcherArmed = true
				go func() {
					defer core.RecoverPanic(fmt.Sprintf("datastore_wait_kill (namespace=%s)", ds.namespace))
					<-waitCtx.Done()
					ds.dataMutex.Lock()
					cond.Broadcast()
					ds.dataMutex.Unlock()
				}()
			}
			cond.Wait()
			if waitCtx.Err() != nil {
				ds.dataMutex.Unlock()
				return nil, fmt.Errorf("killed")
			}
		}
	}
}

// Unshift atomically prepends an item to an array
// Creates the array if key doesn't exist. Returns new array length.
// Returns error if key exists but is not an array.
func (ds *DatastoreValue) Unshift(key string, item any) (float64, error) {
	if ds.isFollower() {
		v, err := ds.forwardSimple(opUnshift, key, "", 0, item, true)
		n, _ := toFloat64(v)
		return n, err
	}
	if err := ds.checkValueSize("unshift", key, item); err != nil {
		return 0, err
	}

	ds.dataMutex.Lock()
	defer ds.dataMutex.Unlock()

	// Lazy expiry check - if expired, treat as non-existent
	ds.checkExpired(key)

	var newArr []any
	if val, exists := ds.data[key]; exists {
		// Key exists - must be an array
		if arr, ok := val.([]any); ok {
			// Deep copy the item before prepending
			newArr = []any{DeepCopyAny(item)}
			newArr = append(newArr, arr...)
		} else {
			return 0, fmt.Errorf("unshift() cannot operate on non-array value at key %q", key)
		}
	} else {
		// Key doesn't exist - create new array with the item
		newArr = []any{DeepCopyAny(item)}
	}

	// Log the prepended item only (see Push)
	if err := ds.writeWALOp(opUnshift, key, "", 0, item); err != nil {
		return 0, err
	}

	ds.data[key] = newArr

	// Notify any waiters on this key (value changed)
	if cond, exists := ds.conditions[key]; exists {
		cond.Broadcast()
	}

	return float64(len(newArr)), nil
}

// Exists checks if a key exists in the datastore (thread-safe)
func (ds *DatastoreValue) Exists(key string) bool {
	ds.dataMutex.RLock()
	defer ds.dataMutex.RUnlock()
	_, exists := ds.data[key]
	return exists
}

// Rename atomically renames a key (moves value to new key, deletes old key)
// Returns error if oldKey doesn't exist or if newKey already exists
func (ds *DatastoreValue) Rename(oldKey, newKey string) error {
	if ds.isFollower() {
		_, err := ds.forwardSimple(opRename, oldKey, newKey, 0, nil, false)
		return err
	}
	ds.dataMutex.Lock()
	defer ds.dataMutex.Unlock()

	// Old key must exist
	oldValue, exists := ds.data[oldKey]
	if !exists {
		return fmt.Errorf("rename() old key %q does not exist", oldKey)
	}

	// New key must not exist
	if _, exists := ds.data[newKey]; exists {
		return fmt.Errorf("rename() new key %q already exists", newKey)
	}

	if err := ds.writeWALOp(opRename, oldKey, newKey, 0, nil); err != nil {
		return err
	}

	// Move the value
	ds.data[newKey] = oldValue
	delete(ds.data, oldKey)

	// Move condition variable if it exists
	if cond, exists := ds.conditions[oldKey]; exists {
		ds.conditions[newKey] = cond
		delete(ds.conditions, oldKey)
	}

	// Broadcast to both keys
	if cond, exists := ds.conditions[oldKey]; exists {
		cond.Broadcast()
	}
	if cond, exists := ds.conditions[newKey]; exists {
		cond.Broadcast()
	}

	return nil
}

// Expire sets a time-to-live (TTL) for a key in seconds
// The key will be automatically deleted when the TTL expires
// Calling expire() on an existing key resets the TTL
// Returns error if the key doesn't exist
func (ds *DatastoreValue) Expire(key string, ttlSeconds float64) error {
	if ds.isFollower() {
		_, err := ds.forwardSimple(opExpire, key, "", ttlSeconds, nil, false)
		return err
	}
	ds.dataMutex.Lock()
	defer ds.dataMutex.Unlock()

	// Key must exist
	if _, exists := ds.data[key]; !exists {
		return fmt.Errorf("expire() key %q does not exist", key)
	}

	// Calculate expiry time
	ttl := time.Duration(ttlSeconds) * time.Second
	expiryTime := time.Now().Add(ttl)

	// Log the absolute deadline, never the TTL — replaying "60 seconds from now"
	// three hours after a crash would resurrect the key.
	if err := ds.writeWALOp(opExpire, key, "", float64(expiryTime.UnixMilli()), nil); err != nil {
		return err
	}

	// Update expiryTimes map (quick lookup)
	ds.expiryTimes[key] = expiryTime

	// Push to min-heap
	heap.Push(&ds.expiryHeap, ExpiryEntry{key: key, expiryTime: expiryTime})

	return nil
}

// sweepExpiredKeys removes keys that have expired from the heap
// This is called by the background ticker every 1 second
// Uses lazy deletion: checks expiryTimes[key] before deleting
func (ds *DatastoreValue) sweepExpiredKeys() {
	// Expiry is the leader's decision, and it reaches followers as an EXPIRED
	// record like any other write. A follower that swept on its own schedule
	// would delete keys the leader still holds, and — worse — would call
	// writeWALOp, advancing a sequence counter that belongs to the leader.
	if ds.repl != nil && ds.repl.role == replRoleFollower {
		return
	}

	ds.dataMutex.Lock()
	defer ds.dataMutex.Unlock()

	now := time.Now()

	// Pop expired entries from the heap
	for len(ds.expiryHeap) > 0 && (ds.expiryHeap[0].expiryTime.Before(now) || ds.expiryHeap[0].expiryTime.Equal(now)) {
		entry := heap.Pop(&ds.expiryHeap).(ExpiryEntry)

		// Lazy deletion check: only delete if the key still has this expiry time
		if expiryTime, exists := ds.expiryTimes[entry.key]; exists && expiryTime.Equal(entry.expiryTime) {
			// Log before applying. If the log write fails, leave the key in place
			// and re-queue it rather than dropping it from memory only — memory
			// diverging from the log is worse than a late expiry.
			if err := ds.writeWALOp(opExpired, entry.key, "", 0, nil); err != nil {
				heap.Push(&ds.expiryHeap, entry)
				return
			}
			// Key is still expired, delete it
			delete(ds.data, entry.key)
			delete(ds.expiryTimes, entry.key)
			delete(ds.conditions, entry.key)
		}
		// If expiryTimes[key] doesn't match, it was re-expired, so skip this old heap entry
	}
}

// checkExpired is called before returning values to catch lazily-deleted keys
// Returns true if the key is expired and was deleted, false otherwise
func (ds *DatastoreValue) checkExpired(key string) bool {
	now := time.Now()
	if expiryTime, exists := ds.expiryTimes[key]; exists && (expiryTime.Before(now) || expiryTime.Equal(now)) {
		// Key is expired, delete it
		delete(ds.data, key)
		delete(ds.expiryTimes, key)
		delete(ds.conditions, key)
		return true
	}
	return false
}

// Wait blocks until the key changes (if no expectedValue) or equals expectedValue (if provided)
// If expectedValue is nil (omitted), waits for ANY change to the key
// WaitWithPredicate waits until a predicate function returns true for the key's value
// The predicate is called with the current value and should return true when condition is met
// Timeout is optional (pass 0 for no timeout)
// Returns the current value of the key after the predicate returns true, or error on timeout
func (ds *DatastoreValue) WaitWithPredicate(procCtx context.Context, evaluator *Evaluator, key string, predicateFn Value, timeout time.Duration) (any, error) {
	if procCtx == nil {
		procCtx = context.Background()
	}
	ds.dataMutex.Lock()

	// Get or create condition variable for this key
	cond, exists := ds.conditions[key]
	if !exists {
		cond = sync.NewCond(&ds.dataMutex)
		ds.conditions[key] = cond
	}

	// Child context so kill(pid) (cancelling procCtx) can wake this specific wait
	// early via Broadcast - cheap (one alloc, no channel until Done() is read),
	// and Go's context tree already fans this out correctly to concurrent waiters.
	waitCtx, cancelWait := context.WithCancel(procCtx)
	defer cancelWait()
	var killWatcherArmed bool

	// Loop until predicate returns true
	for {
		current, keyExists := ds.data[key]
		if keyExists {
			// Call the predicate function with the current value
			fnArgs := map[string]Value{"0": InterfaceToValue(current)}
			result, err := evaluator.CallFunction(predicateFn, fnArgs)
			if err != nil {
				ds.dataMutex.Unlock()
				return nil, fmt.Errorf("wait() predicate error: %v", err)
			}
			if result.IsTruthy() {
				ds.dataMutex.Unlock()
				return current, nil
			}
		}

		// Wait for notification
		if timeout > 0 {
			// Start a goroutine that will broadcast on timeout or kill()
			go func() {
				defer core.RecoverPanic(fmt.Sprintf("datastore_wait_timeout (namespace=%s)", ds.namespace))
				select {
				case <-time.After(timeout):
				case <-waitCtx.Done():
				}
				ds.dataMutex.Lock()
				cond.Broadcast()
				ds.dataMutex.Unlock()
			}()

			// Record start time for checking actual timeout
			startTime := time.Now()
			cond.Wait() // Called with lock held - safe

			if waitCtx.Err() != nil {
				ds.dataMutex.Unlock()
				return nil, fmt.Errorf("killed")
			}
			// Check if we actually timed out
			if time.Since(startTime) >= timeout {
				ds.dataMutex.Unlock()
				return nil, fmt.Errorf("wait() timeout exceeded for key %q", key)
			}
			// Otherwise, loop will re-check the condition
		} else {
			// No timeout - arm a one-time watcher for kill(), then just wait
			if !killWatcherArmed {
				killWatcherArmed = true
				go func() {
					defer core.RecoverPanic(fmt.Sprintf("datastore_wait_kill (namespace=%s)", ds.namespace))
					<-waitCtx.Done()
					ds.dataMutex.Lock()
					cond.Broadcast()
					ds.dataMutex.Unlock()
				}()
			}
			cond.Wait()
			if waitCtx.Err() != nil {
				ds.dataMutex.Unlock()
				return nil, fmt.Errorf("killed")
			}
		}
	}
}

// For array values, this means waiting for length to change (new append)
// If expectedValue is provided, waits until key equals that value
// Timeout is optional (pass 0 for no timeout)
// Returns the current value of the key after the condition is met, or error on timeout
func (ds *DatastoreValue) Wait(procCtx context.Context, key string, expectedValue any, hasExpectedValue bool, timeout time.Duration) (any, error) {
	if procCtx == nil {
		procCtx = context.Background()
	}
	ds.dataMutex.Lock()

	// Get initial value and its length (for arrays)
	initialValue, _ := ds.data[key]
	initialLen := getLength(initialValue)

	// Get or create condition variable for this key
	cond, exists := ds.conditions[key]
	if !exists {
		cond = sync.NewCond(&ds.dataMutex)
		ds.conditions[key] = cond
	}

	// Child context so kill(pid) (cancelling procCtx) can wake this specific wait
	// early via Broadcast - cheap (one alloc, no channel until Done() is read),
	// and Go's context tree already fans this out correctly to concurrent waiters.
	waitCtx, cancelWait := context.WithCancel(procCtx)
	defer cancelWait()
	var killWatcherArmed bool

	// Loop until condition is met
	for {
		current, keyExists := ds.data[key]

		if hasExpectedValue {
			// Wait until key equals specific value
			if keyExists && valuesEqual(current, expectedValue) {
				ds.dataMutex.Unlock()
				return current, nil
			}
		} else {
			// Wait until key changes from initial value
			// For arrays, check if length changed
			currentLen := getLength(current)
			if keyExists && (currentLen != initialLen || !valuesEqual(current, initialValue)) {
				ds.dataMutex.Unlock()
				return current, nil
			}
		}

		// Wait for notification
		if timeout > 0 {
			// Start a goroutine that will broadcast on timeout or kill()
			go func() {
				defer core.RecoverPanic(fmt.Sprintf("datastore_wait_timeout (namespace=%s)", ds.namespace))
				select {
				case <-time.After(timeout):
				case <-waitCtx.Done():
				}
				ds.dataMutex.Lock()
				cond.Broadcast()
				ds.dataMutex.Unlock()
			}()

			// Record start time for checking actual timeout
			startTime := time.Now()
			cond.Wait() // Called with lock held - safe

			if waitCtx.Err() != nil {
				ds.dataMutex.Unlock()
				return nil, fmt.Errorf("killed")
			}
			// Check if we actually timed out
			if time.Since(startTime) >= timeout {
				ds.dataMutex.Unlock()
				return nil, fmt.Errorf("wait() timeout exceeded for key %q", key)
			}
			// Otherwise, loop will re-check the condition
		} else {
			// No timeout - arm a one-time watcher for kill(), then just wait
			if !killWatcherArmed {
				killWatcherArmed = true
				go func() {
					defer core.RecoverPanic(fmt.Sprintf("datastore_wait_kill (namespace=%s)", ds.namespace))
					<-waitCtx.Done()
					ds.dataMutex.Lock()
					cond.Broadcast()
					ds.dataMutex.Unlock()
				}()
			}
			cond.Wait()
			if waitCtx.Err() != nil {
				ds.dataMutex.Unlock()
				return nil, fmt.Errorf("killed")
			}
		}
	}
}

// Update atomically reads, deep merges updates into an object, and returns the updated object
// Creates an empty object if key doesn't exist
// Returns error if key exists but is not an object
// Supports nil values to delete keys from the object (shallow deletion only)
func (ds *DatastoreValue) Update(key string, updates any) (any, error) {
	if ds.isFollower() {
		return ds.forwardSimple(opUpdate, key, "", 0, updates, true)
	}
	if err := ds.checkValueSize("update", key, updates); err != nil {
		return nil, err
	}

	ds.dataMutex.Lock()
	defer ds.dataMutex.Unlock()

	// Lazy expiry check
	ds.checkExpired(key)

	// Get current value or create empty object
	current, exists := ds.data[key]
	var obj map[string]any

	if exists {
		// Key exists - must be an object
		if o, ok := current.(map[string]any); ok {
			// Deep copy the current object to avoid mutations
			obj = DeepCopyAny(o).(map[string]any)
		} else {
			return nil, fmt.Errorf("update() cannot operate on non-object value at key %q", key)
		}
	} else {
		// Key doesn't exist - create empty object
		obj = make(map[string]any)
	}

	// Updates must be an object
	updateMap, ok := updates.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("update() updates argument must be an object")
	}

	// Deep merge updates into object
	deepMerge(obj, updateMap)

	storedObj := DeepCopyAny(obj)

	// Log the patch, not the merged object: replay re-runs the same deep merge,
	// and the patch is usually a fraction of the record's size.
	if err := ds.writeWALOp(opUpdate, key, "", 0, updateMap); err != nil {
		return nil, err
	}

	// Store the updated object
	ds.data[key] = storedObj

	// Notify any waiters on this key
	if cond, exists := ds.conditions[key]; exists {
		cond.Broadcast()
	}

	// Return the updated object (deep copied to isolate from datastore's scope)
	return DeepCopyAny(obj), nil
}

// deepMerge recursively merges src into dst
// Handles nil values as deletion markers
func deepMerge(dst, src map[string]any) {
	for k, v := range src {
		if v == nil {
			// Nil values delete the key
			delete(dst, k)
		} else if srcMap, ok := v.(map[string]any); ok {
			// Recursive merge for nested objects
			if dstVal, exists := dst[k]; exists {
				if dstMap, ok := dstVal.(map[string]any); ok {
					deepMerge(dstMap, srcMap)
					continue
				}
			}
			// If dst doesn't have this key or it's not an object, copy the nested object
			dst[k] = DeepCopyAny(srcMap)
		} else {
			// For all other types, just copy the value
			dst[k] = DeepCopyAny(v)
		}
	}
}

// WaitFor blocks until predicate(value) returns true
// For array values, predicate receives the array length as a number
// Predicate is a Duso function that takes one argument and returns a boolean
// Timeout is optional (pass 0 for no timeout)
// Returns the current value of the key after the predicate is true, or error on timeout
// Delete removes a key from the store and returns the deleted value (or nil if key didn't exist)
func (ds *DatastoreValue) Delete(key string) (any, error) {
	if ds.isFollower() {
		return ds.forwardSimple(opDelete, key, "", 0, nil, false)
	}
	ds.dataMutex.Lock()
	defer ds.dataMutex.Unlock()

	// A real DELETE opcode: the pre-v1 log wrote a nil value here, which was
	// indistinguishable from set(key, nil) and recovered as present-with-nil.
	if err := ds.writeWALOp(opDelete, key, "", 0, nil); err != nil {
		return nil, err
	}

	value := ds.data[key]
	delete(ds.data, key)
	delete(ds.conditions, key)
	delete(ds.expiryTimes, key)
	// Note: We don't remove from expiryHeap - it will be cleaned up lazily during sweep

	if !ds.returnDeletedValue {
		return nil, nil
	}

	// Deep copy to isolate returned values from datastore's scope
	if value != nil {
		return DeepCopyAny(value), nil
	}
	return nil, nil
}

// Clear removes all keys from the store
func (ds *DatastoreValue) Clear() error {
	if ds.isFollower() {
		_, err := ds.forwardSimple(opClear, "", "", 0, nil, false)
		return err
	}
	ds.dataMutex.Lock()
	defer ds.dataMutex.Unlock()

	if err := ds.writeWALOp(opClear, "", "", 0, nil); err != nil {
		return err
	}

	ds.data = make(map[string]any)
	ds.conditions = make(map[string]*sync.Cond)
	ds.expiryTimes = make(map[string]time.Time)
	ds.expiryHeap = make(ExpiryHeap, 0)

	return nil
}

// Save explicitly saves the datastore to disk (JSON)
func (ds *DatastoreValue) Save() error {
	if ds.persistPath == "" {
		return fmt.Errorf("datastore %q has no persist path configured", ds.namespace)
	}
	return ds.saveToDisk()
}

// Load explicitly loads the datastore from disk (JSON)
func (ds *DatastoreValue) Load() error {
	if ds.persistPath == "" {
		return fmt.Errorf("datastore %q has no persist path configured", ds.namespace)
	}
	if err := ds.loadFromDisk(false); err != nil {
		return fmt.Errorf("datastore(%q): %v", ds.namespace, err)
	}
	return nil
}

// Keys returns a slice of all keys in the datastore
func (ds *DatastoreValue) Keys() []string {
	ds.dataMutex.RLock()
	defer ds.dataMutex.RUnlock()

	keys := make([]string, 0, len(ds.data))
	for k := range ds.data {
		keys = append(keys, k)
	}
	return keys
}

// Select queries the datastore by running a predicate function on each key-value pair.
// The predicate receives (key, value) and returns:
// - nil to exclude this entry
// - any non-nil value to include it in the results
// Results are deep-copied to isolate from datastore mutations.
// Snapshot keys at start, then lock per-key during iteration for minimal blocking.
// Returns error if the predicate throws.
// Select runs predicate on each key/value, collecting non-nil returns.
// If max > 0, iteration stops as soon as max results are collected.
// Map iteration order is non-deterministic, so with max > 0 you get *any*
// matching entries, not a deterministic "first N".
func (ds *DatastoreValue) Select(evaluator *Evaluator, predicateFn Value, max int) ([]any, error) {
	// Snapshot keys (lightweight)
	ds.dataMutex.RLock()
	keys := make([]string, 0, len(ds.data))
	for k := range ds.data {
		keys = append(keys, k)
	}
	ds.dataMutex.RUnlock()

	// Iterate keys with per-key locking
	results := make([]any, 0)
	for _, key := range keys {
		// Lock only to read and copy this value
		ds.dataMutex.Lock()
		val, exists := ds.data[key]
		if !exists {
			ds.dataMutex.Unlock()
			continue // key was deleted
		}
		valCopy := DeepCopyAny(val)
		ds.dataMutex.Unlock()

		// Call predicate (unlocked)
		fnArgs := map[string]Value{
			"0": NewString(key),
			"1": InterfaceToValue(valCopy),
		}
		result, err := evaluator.CallFunction(predicateFn, fnArgs)
		if err != nil {
			return nil, fmt.Errorf("select() predicate error on key %q: %v", key, err)
		}

		// Test the type tag, not Data: numbers live in Value.Num and leave Data
		// nil, so checking Data dropped every numeric result.
		if !result.IsNil() {
			results = append(results, DeepCopyAny(ValueToInterface(result)))
			if max > 0 && len(results) >= max {
				break
			}
		}
	}

	return results, nil
}

// Count returns the number of entries for which the predicate returns a truthy value.
// Like Select but counts instead of collecting — avoids building/copying a result array.
// Predicate receives (key, value); truthy returns are counted, falsy (nil/false/0/"") are not.
func (ds *DatastoreValue) Count(evaluator *Evaluator, predicateFn Value) (float64, error) {
	ds.dataMutex.RLock()
	keys := make([]string, 0, len(ds.data))
	for k := range ds.data {
		keys = append(keys, k)
	}
	ds.dataMutex.RUnlock()

	var count float64
	for _, key := range keys {
		ds.dataMutex.Lock()
		val, exists := ds.data[key]
		if !exists {
			ds.dataMutex.Unlock()
			continue
		}
		valCopy := DeepCopyAny(val)
		ds.dataMutex.Unlock()

		fnArgs := map[string]Value{
			"0": NewString(key),
			"1": InterfaceToValue(valCopy),
		}
		result, err := evaluator.CallFunction(predicateFn, fnArgs)
		if err != nil {
			return 0, fmt.Errorf("count() predicate error on key %q: %v", key, err)
		}

		if result.IsTruthy() {
			count++
		}
	}

	return count, nil
}

// Shutdown stops the auto-save ticker and expiry ticker, and saves final state
func (ds *DatastoreValue) Shutdown() error {
	// Stop replication first: a follower must not apply another record once the
	// final snapshot below has been taken, or the file's watermark understates
	// what the store actually holds.
	ds.replShutdown()

	if ds.ticker != nil {
		ds.ticker.Stop()
		select {
		case ds.stopTicker <- true:
		default:
		}
	}

	// Stop expiry sweep ticker
	select {
	case ds.expiryStopTicker <- true:
	default:
	}

	// Stop WAL sync ticker if running
	if ds.walSyncTicker != nil {
		ds.walSyncTicker.Stop()
		select {
		case ds.walStopSync <- true:
		default:
		}
	}

	// Sync WAL before shutdown. Closing under walMutex for the same reason
	// syncWAL locks before reading the handle: a save racing shutdown can be
	// replacing it.
	_ = ds.syncWAL()
	ds.walMutex.Lock()
	if ds.walFile != nil {
		_ = ds.walFile.Close()
		ds.walFile = nil
	}
	ds.walMutex.Unlock()

	// Final save if configured
	if ds.persistPath != "" {
		return ds.saveToDisk()
	}
	return nil
}

// encryptBytes encrypts data using AES-256-GCM if encryption key is configured
func (ds *DatastoreValue) encryptBytes(data []byte) ([]byte, error) {
	if len(ds.encryptKey) == 0 {
		return data, nil // No encryption configured
	}

	block, err := aes.NewCipher(ds.encryptKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %v", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %v", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %v", err)
	}

	ciphertext := gcm.Seal(nil, nonce, data, nil)
	return append(nonce, ciphertext...), nil
}

// decryptBytes decrypts data using AES-256-GCM if encryption key is configured
func (ds *DatastoreValue) decryptBytes(data []byte) ([]byte, error) {
	if len(ds.encryptKey) == 0 {
		return data, nil // No encryption configured
	}

	block, err := aes.NewCipher(ds.encryptKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %v", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %v", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce := data[:nonceSize]
	ciphertext := data[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: authentication tag mismatch or corrupted data")
	}

	return plaintext, nil
}

// saveToDisk serializes the datastore to a gob file and flushes to disk
// After successful save, truncates the WAL (if configured)
func (ds *DatastoreValue) saveToDisk() error {
	if ds.persistPath == "" {
		return nil // No persistence configured
	}

	ds.fileWriteMutex.Lock()
	defer ds.fileWriteMutex.Unlock()

	ds.dataMutex.RLock()
	defer ds.dataMutex.RUnlock()

	// Create parent directory if needed
	persistDir := core.Dir(ds.persistPath)
	if persistDir != "" && persistDir != "." {
		if err := os.MkdirAll(persistDir, 0755); err != nil {
			return fmt.Errorf("failed to create datastore directory %q: %v", persistDir, err)
		}
	}

	// Encode as a v1 snapshot: plaintext header, then a codec-encoded body
	// (encrypted as a unit if a key is configured). Reading still accepts pre-v1
	// files — see decodeSnapshot.
	dataToWrite, err := ds.encodeSnapshot()
	if err != nil {
		return fmt.Errorf("failed to serialize datastore %q: %v", ds.namespace, err)
	}

	// Open file for writing
	file, err := os.OpenFile(ds.persistPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open datastore %q at %q: %v", ds.namespace, ds.persistPath, err)
	}
	defer file.Close()

	// Write data to file
	if _, err := file.Write(dataToWrite); err != nil {
		return fmt.Errorf("failed to write datastore %q: %v", ds.namespace, err)
	}

	// Flush to disk to ensure data hits storage
	if err := file.Sync(); err != nil {
		return fmt.Errorf("failed to sync datastore %q to disk: %v", ds.namespace, err)
	}

	// Truncate WAL after successful snapshot (it's captured in the snapshot now)
	if ds.walPath != "" {
		if err := ds.truncateWAL(); err != nil {
			// Log but don't fail - snapshot succeeded even if WAL truncate failed
			fmt.Fprintf(os.Stderr, "warning: failed to truncate WAL for %q: %v\n", ds.namespace, err)
		}
	}

	return nil
}

// loadFromDisk deserializes the datastore from its snapshot file.
// Accepts both v1 snapshots and pre-v1 bare-gob files; the next save() rewrites
// whatever was read as v1.
//
// replace distinguishes the two callers:
//
//   - Recovery (replace=true) must land on exactly the snapshot's state, because
//     the WAL replayed on top of it is an operation log. Replaying PUSH over a
//     store that already holds the item appends it twice. Starting from the
//     snapshot exactly is what makes the sequence watermark mean anything.
//   - An explicit load() (replace=false) merges, preserving the long-standing
//     behavior where keys the file doesn't mention survive the call.
func (ds *DatastoreValue) loadFromDisk(replace bool) error {
	if ds.persistPath == "" {
		return nil // No persistence configured
	}

	ds.fileWriteMutex.Lock()
	defer ds.fileWriteMutex.Unlock()

	// Callers name the namespace; these messages name the file and the cause.
	fileBytes, err := os.ReadFile(ds.persistPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist yet - OK
		}
		return fmt.Errorf("cannot read %q: %v", ds.persistPath, err)
	}

	data, expiry, seq, err := ds.decodeSnapshot(fileBytes)
	if err != nil {
		return err
	}

	ds.dataMutex.Lock()
	defer ds.dataMutex.Unlock()

	if replace {
		ds.data = data
		ds.expiryTimes = make(map[string]time.Time, len(expiry))
		ds.expiryHeap = make(ExpiryHeap, 0, len(expiry))
		ds.applySnapshotExpiry(expiry)
		// The watermark is what this snapshot covers, so replay applies exactly
		// the records written after it and no others.
		ds.walSeq.Store(seq)
		return nil
	}

	for k, v := range data {
		ds.data[k] = v
	}
	ds.applySnapshotExpiry(expiry)
	if seq > ds.walSeq.Load() {
		ds.walSeq.Store(seq)
	}

	return nil
}

// replayWALForRecovery replays the log and, unless the store is readonly,
// folds the result back into a fresh snapshot and truncates the log.
//
// The readonly path deliberately stops after the replay. Writing back would
// truncate the WAL of whatever process is actively writing these files, which
// is the difference between inspecting a live store and destroying it.
func (ds *DatastoreValue) replayWALForRecovery() error {
	if ds.walPath == "" {
		return nil
	}
	if ds.readonly {
		ds.walMutex.Lock()
		defer ds.walMutex.Unlock()
		return ds.replayWAL()
	}
	return ds.recoverFromWAL()
}

// recoverFromWAL replays WAL entries, saves merged state, and truncates WAL
func (ds *DatastoreValue) recoverFromWAL() error {
	if ds.walPath == "" {
		return nil
	}

	ds.walMutex.Lock()
	defer ds.walMutex.Unlock()

	// Replay WAL entries on top of loaded snapshot.
	// Callers name the namespace and the WAL path.
	if err := ds.replayWAL(); err != nil {
		return err
	}

	// Save merged state (snapshot + replayed WAL)
	// Release walMutex since saveToDisk needs other locks
	ds.walMutex.Unlock()
	if ds.persistPath != "" {
		if err := ds.saveToDisk(); err != nil {
			ds.walMutex.Lock()
			return err // saveToDisk already calls truncateWAL on success
		}
	} else {
		// No persist file, just truncate WAL
		if err := ds.truncateWAL(); err != nil {
			ds.walMutex.Lock()
			return fmt.Errorf("failed to truncate WAL for %q: %v", ds.namespace, err)
		}
	}
	ds.walMutex.Lock()

	return nil
}

// openWALForWrites opens the WAL file for appending new entries
func (ds *DatastoreValue) openWALForWrites() error {
	if ds.walPath == "" {
		return nil
	}

	ds.walMutex.Lock()
	defer ds.walMutex.Unlock()

	// Create parent directory if needed
	walDir := core.Dir(ds.walPath)
	if walDir != "" && walDir != "." {
		if err := os.MkdirAll(walDir, 0755); err != nil {
			return fmt.Errorf("failed to create WAL directory %q: %v", walDir, err)
		}
	}

	file, err := os.OpenFile(ds.walPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open WAL file %q: %v", ds.walPath, err)
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return fmt.Errorf("failed to stat WAL file %q: %v", ds.walPath, err)
	}

	if info.Size() == 0 {
		if err := ds.appendWALHeader(file); err != nil {
			file.Close()
			return err
		}
	} else {
		// Recovery truncates before reaching here, so an existing file must
		// already be v1. Validating stops us appending v1 records onto a file
		// this build cannot read back.
		hdr := make([]byte, walHeaderLen)
		if _, err := file.ReadAt(hdr, 0); err != nil {
			file.Close()
			return fmt.Errorf("failed to read WAL header from %q: %v", ds.walPath, err)
		}
		if _, err := ds.readWALHeader(hdr); err != nil {
			file.Close()
			return err
		}
	}

	ds.walFile = file
	ds.walEncoder = nil

	// Start WAL sync ticker if batching is configured
	if ds.walSyncInterval > 0 {
		ds.walSyncTicker = time.NewTicker(ds.walSyncInterval)
		ds.walStopSync = make(chan bool, 1)
		go func() {
			defer core.RecoverPanic(fmt.Sprintf("datastore_wal_sync (namespace=%s)", ds.namespace))
			for {
				select {
				case <-ds.walSyncTicker.C:
					_ = ds.syncWAL()
				case <-ds.walStopSync:
					ds.walSyncTicker.Stop()
					return
				}
			}
		}()
	}

	return nil
}

// replayWAL applies the log on top of the loaded snapshot. Both v1 op-logs and
// pre-v1 gob state-logs are accepted; the next truncate rewrites the file as v1.
// Callers hold dataMutex via recoverFromWAL.
func (ds *DatastoreValue) replayWAL() error {
	if ds.walPath == "" {
		return nil
	}

	file, err := os.Open(ds.walPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No WAL file yet - OK
		}
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat WAL %q: %v", ds.walPath, err)
	}
	if info.Size() == 0 {
		return nil
	}

	hdr := make([]byte, walHeaderLen)
	n, err := file.ReadAt(hdr, 0)
	if err != nil && n < walHeaderLen {
		// Too short to hold a header: can only be a pre-v1 file.
		return ds.replayLegacyWAL()
	}
	if string(hdr[:8]) != walMagic {
		return ds.replayLegacyWAL()
	}

	encrypted, err := ds.readWALHeader(hdr)
	if err != nil {
		return err
	}
	if _, err := file.Seek(walHeaderLen, io.SeekStart); err != nil {
		return err
	}

	res, err := ds.replayWALv1(file, encrypted, info.Size())
	if err != nil {
		return err
	}

	// Drop an incomplete trailing record so the log never keeps bytes that can
	// no longer be reached. The writer paths truncate the whole log after
	// recovery anyway, but this keeps replayWAL correct on its own rather than
	// relying on what the caller happens to do next. A readonly store must not
	// touch the file at all — another process may be writing it.
	if res.tornAt >= 0 && !ds.readonly {
		file.Close()
		if err := os.Truncate(ds.walPath, res.lastGood); err != nil {
			return fmt.Errorf("failed to drop torn tail from WAL %q: %v", ds.walPath, err)
		}
	}
	return nil
}

// replayLegacyWAL reads a pre-v1 log: gob-encoded {Key, Value} entries applied
// as blind writes, optionally length-prefixed and encrypted.
//
// This log shape cannot distinguish delete(key) from set(key, nil) — both were
// written as a nil value — so a deleted key recovers as present-with-nil, which
// is what pre-v1 duso did. Deprecated; remove once the migration window closes.
func (ds *DatastoreValue) replayLegacyWAL() error {
	file, err := os.Open(ds.walPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	if len(ds.encryptKey) > 0 {
		for {
			lenBuf := make([]byte, 4)
			n, err := io.ReadFull(file, lenBuf)
			if err != nil || n != 4 {
				break // clean or ragged end of a legacy log
			}
			encLength := binary.BigEndian.Uint32(lenBuf)
			encryptedEntry := make([]byte, encLength)
			if _, err := io.ReadFull(file, encryptedEntry); err != nil {
				break
			}
			decrypted, err := ds.decryptBytes(encryptedEntry)
			if err != nil {
				return fmt.Errorf("failed to decrypt legacy WAL entry: %v", err)
			}
			var entry WALEntry
			if err := gob.NewDecoder(bytes.NewReader(decrypted)).Decode(&entry); err != nil {
				return fmt.Errorf("failed to decode legacy WAL entry: %v", err)
			}
			ds.data[entry.Key] = entry.Value
		}
		return nil
	}

	decoder := gob.NewDecoder(file)
	for {
		var entry WALEntry
		if err := decoder.Decode(&entry); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to decode legacy WAL entry: %v", err)
		}
		ds.data[entry.Key] = entry.Value
	}
	return nil
}

// syncWAL flushes buffered WAL writes to disk.
//
// The nil check is inside the lock, not before it: truncateWAL closes the handle
// and installs a new one while holding walMutex, so reading walFile unlocked
// races it, and a sync ticker that read the old pointer would fsync a closed
// file.
func (ds *DatastoreValue) syncWAL() error {
	ds.walMutex.Lock()
	defer ds.walMutex.Unlock()

	if ds.walFile == nil {
		return nil
	}

	if err := ds.walFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync WAL: %v", err)
	}
	return nil
}

// truncateWAL clears the WAL after a successful snapshot and re-establishes the
// v1 header. The snapshot records the sequence watermark it covers, so replay
// after this point resumes from the right place.
func (ds *DatastoreValue) truncateWAL() error {
	if ds.walPath == "" {
		return nil
	}

	ds.walMutex.Lock()
	defer ds.walMutex.Unlock()

	if ds.walFile != nil {
		ds.walFile.Close()
		ds.walFile = nil
		ds.walEncoder = nil
	}

	if err := os.Truncate(ds.walPath, 0); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to truncate WAL %q: %v", ds.walPath, err)
		}
		// File doesn't exist - that's OK, nothing to truncate
	}

	file, err := os.OpenFile(ds.walPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to reopen WAL file %q after truncate: %v", ds.walPath, err)
	}
	if err := ds.appendWALHeader(file); err != nil {
		file.Close()
		return err
	}

	ds.walFile = file
	return nil
}

// Handles numeric comparisons (int/float) appropriately
func valuesEqual(a, b any) bool {
	// Handle numeric comparisons
	aFloat, aIsFloat := toFloat64(a)
	bFloat, bIsFloat := toFloat64(b)

	if aIsFloat && bIsFloat {
		return aFloat == bFloat
	}

	// String comparison
	if aStr, ok := a.(string); ok {
		if bStr, ok := b.(string); ok {
			return aStr == bStr
		}
	}

	// Boolean comparison
	if aBool, ok := a.(bool); ok {
		if bBool, ok := b.(bool); ok {
			return aBool == bBool
		}
	}

	// Array comparison
	if aArr, ok := a.([]any); ok {
		if bArr, ok := b.([]any); ok {
			if len(aArr) != len(bArr) {
				return false
			}
			for i := range aArr {
				if !valuesEqual(aArr[i], bArr[i]) {
					return false
				}
			}
			return true
		}
	}

	// Object/map comparison - treat as equal (don't walk the tree)
	// Maps can't be compared with ==, so we assume they're equal unless type mismatch
	if _, ok := a.(map[string]any); ok {
		_, ok := b.(map[string]any)
		return ok // True if both are maps (equal), false if types differ
	}

	// Fall back to interface equality (for nil, etc.)
	return a == b
}

// toFloat64 attempts to convert a value to float64
func toFloat64(val any) (float64, bool) {
	switch v := val.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	default:
		return 0, false
	}
}

// isArray checks if a value is an array/slice
func isArray(val any) bool {
	switch val.(type) {
	case []any:
		return true
	default:
		return false
	}
}

// getLength gets the length of a value (works for arrays, strings, etc.)
func getLength(val any) int {
	switch v := val.(type) {
	case []any:
		return len(v)
	case string:
		return len(v)
	case map[string]any:
		return len(v)
	default:
		return 0
	}
}
