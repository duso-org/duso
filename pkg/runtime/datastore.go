package runtime

import (
	"container/heap"
	"encoding/gob"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/duso-org/duso/pkg/core"
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
	stopTicker         chan bool              // Signal to stop ticker
	expiryTicker       *time.Ticker          // Expiry sweep ticker
	fileWriteMutex     sync.Mutex             // Serialize file writes
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
	walMutex           sync.Mutex             // Protect concurrent WAL writes
	walSyncInterval    time.Duration         // 0=sync every write, >0=batch writes
	walSyncTicker      *time.Ticker          // Periodic WAL sync (if batching)
	walStopSync        chan bool              // Signal to stop WAL sync ticker
}

// applyDatastoreConfig applies configuration to a datastore and triggers recovery.
// IMPORTANT: Paths in config must be pre-resolved (caller is responsible).
func applyDatastoreConfig(store *DatastoreValue, config map[string]any) {
	if config == nil {
		return
	}

	// Apply config options - paths must already be resolved by caller
	if persistPath, ok := config["persist"].(string); ok {
		store.persistPath = persistPath
	}
	if persistInterval, ok := config["persist_interval"]; ok {
		if intervalSecs, ok := persistInterval.(float64); ok {
			store.persistInterval = time.Duration(intervalSecs) * time.Second
		}
	}
	if walPath, ok := config["wal"].(string); ok {
		store.walPath = walPath
	}
	if walSyncInterval, ok := config["wal_sync_interval"]; ok {
		if intervalSecs, ok := walSyncInterval.(float64); ok {
			store.walSyncInterval = time.Duration(intervalSecs) * time.Second
		}
	}
	if readonly, ok := config["readonly"]; ok {
		if r, ok := readonly.(bool); ok {
			store.readonly = r
		}
	}
	if returnDeletedValue, ok := config["return_deleted_value"]; ok {
		if r, ok := returnDeletedValue.(bool); ok {
			store.returnDeletedValue = r
		}
	}

	// Step 1: Load persist if it exists
	if store.persistPath != "" {
		_ = store.loadFromDisk()
	}

	// Step 2: Replay WAL if it exists
	if store.walPath != "" {
		if err := store.recoverFromWAL(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to recover from WAL for %q: %v\n", store.namespace, err)
		}
	}

	// Step 3: Open WAL for new writes if configured
	if store.walPath != "" && store.walFile == nil {
		if err := store.openWALForWrites(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to open WAL for writes for %q: %v\n", store.namespace, err)
		}
	}

	// Start auto-save ticker if configured
	if store.persistInterval > 0 && store.ticker == nil {
		store.ticker = time.NewTicker(store.persistInterval)
		go func() {
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
		returnDeletedValue: true,              // Default: return deleted values
	}

	// Apply namespace defaults
	if namespace == "sys" {
		store.readonly = true
	}
	if namespace == "vfs" {
		store.returnDeletedValue = false // Don't copy large files on delete
	}

	// For sys datastore, set up dynamic metric computation
	// TODO: Implement metrics system properly (currently disabled)
	// if namespace == "sys" {
	//	store.statsFn = GetMetric
	// }

	// Start expiry sweep ticker (1-second sweep)
	store.expiryTicker = time.NewTicker(1 * time.Second)
	go func() {
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

// Set stores a value by key (thread-safe)
func (ds *DatastoreValue) Set(key string, value any) error {
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

	// Write to WAL before applying to memory (durability guarantee)
	if err := ds.writeWAL(key, storedValue); err != nil {
		return err
	}

	ds.dataMutex.Lock()
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
func (ds *DatastoreValue) SetOnce(key string, value any) bool {
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
		return false // Key already exists, don't overwrite
	}

	// Write to WAL before applying to memory
	if err := ds.writeWAL(key, storedValue); err != nil {
		ds.dataMutex.Unlock()
		return false // WAL write failed, don't apply
	}

	ds.data[key] = storedValue

	// Notify any waiters on this key
	if cond, exists := ds.conditions[key]; exists {
		ds.dataMutex.Unlock()
		cond.Broadcast()
	} else {
		ds.dataMutex.Unlock()
	}

	return true // Value was successfully set
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

	// Write to WAL before applying to memory
	if err := ds.writeWAL(key, storedValue); err != nil {
		return nil, err
	}

	ds.dataMutex.Lock()
	defer ds.dataMutex.Unlock()

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

	// Write to WAL before applying to memory
	if err := ds.writeWAL(key, newValue); err != nil {
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

	// Write to WAL before applying to memory
	if err := ds.writeWAL(key, newArr); err != nil {
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

		// Write to WAL before applying to memory
		if err := ds.writeWAL(key, newArr); err != nil {
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

		// Write to WAL before applying to memory
		if err := ds.writeWAL(key, newArr); err != nil {
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
func (ds *DatastoreValue) ShiftWait(key string, timeout time.Duration) (any, error) {
	ds.dataMutex.Lock()

	// Get or create condition variable for this key
	cond, exists := ds.conditions[key]
	if !exists {
		cond = sync.NewCond(&ds.dataMutex)
		ds.conditions[key] = cond
	}

	// Loop until we have an item or timeout
	for {
		// Check if key exists and is an array with items
		val, keyExists := ds.data[key]
		if keyExists {
			if arr, ok := val.([]any); ok {
				if len(arr) > 0 {
					// We have an item - atomically shift and return it
					item := arr[0]
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
			// Start a goroutine that will broadcast on timeout
			timerDone := make(chan struct{})
			go func() {
				<-time.After(timeout)
				ds.dataMutex.Lock()
				cond.Broadcast()
				ds.dataMutex.Unlock()
				close(timerDone)
			}()

			// Record start time for checking actual timeout
			startTime := time.Now()
			cond.Wait() // Called with lock held - safe

			// Check if we actually timed out
			if time.Since(startTime) >= timeout {
				ds.dataMutex.Unlock()
				return nil, nil // Timeout with no item
			}
			// Otherwise, loop will re-check the condition
		} else {
			// No timeout - just wait
			cond.Wait()
		}
	}
}

// PopWait atomically removes and returns the last element from an array
// Blocks until array has items or timeout expires
// Returns nil if timeout exceeded and array is still empty
// Returns error if key exists but is not an array
func (ds *DatastoreValue) PopWait(key string, timeout time.Duration) (any, error) {
	ds.dataMutex.Lock()

	// Get or create condition variable for this key
	cond, exists := ds.conditions[key]
	if !exists {
		cond = sync.NewCond(&ds.dataMutex)
		ds.conditions[key] = cond
	}

	// Loop until we have an item or timeout
	for {
		// Check if key exists and is an array with items
		val, keyExists := ds.data[key]
		if keyExists {
			if arr, ok := val.([]any); ok {
				if len(arr) > 0 {
					// We have an item - atomically pop and return it
					item := arr[len(arr)-1]
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
			// Start a goroutine that will broadcast on timeout
			timerDone := make(chan struct{})
			go func() {
				<-time.After(timeout)
				ds.dataMutex.Lock()
				cond.Broadcast()
				ds.dataMutex.Unlock()
				close(timerDone)
			}()

			// Record start time for checking actual timeout
			startTime := time.Now()
			cond.Wait() // Called with lock held - safe

			// Check if we actually timed out
			if time.Since(startTime) >= timeout {
				ds.dataMutex.Unlock()
				return nil, nil // Timeout with no item
			}
			// Otherwise, loop will re-check the condition
		} else {
			// No timeout - just wait
			cond.Wait()
		}
	}
}

// Unshift atomically prepends an item to an array
// Creates the array if key doesn't exist. Returns new array length.
// Returns error if key exists but is not an array.
func (ds *DatastoreValue) Unshift(key string, item any) (float64, error) {
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

	// Write to WAL before applying to memory
	if err := ds.writeWAL(key, newArr); err != nil {
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
	ds.dataMutex.Lock()
	defer ds.dataMutex.Unlock()

	// Key must exist
	if _, exists := ds.data[key]; !exists {
		return fmt.Errorf("expire() key %q does not exist", key)
	}

	// Calculate expiry time
	ttl := time.Duration(ttlSeconds) * time.Second
	expiryTime := time.Now().Add(ttl)

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
	ds.dataMutex.Lock()
	defer ds.dataMutex.Unlock()

	now := time.Now()

	// Pop expired entries from the heap
	for len(ds.expiryHeap) > 0 && (ds.expiryHeap[0].expiryTime.Before(now) || ds.expiryHeap[0].expiryTime.Equal(now)) {
		entry := heap.Pop(&ds.expiryHeap).(ExpiryEntry)

		// Lazy deletion check: only delete if the key still has this expiry time
		if expiryTime, exists := ds.expiryTimes[entry.key]; exists && expiryTime.Equal(entry.expiryTime) {
			// Key is still expired, delete it
			delete(ds.data, entry.key)
			delete(ds.expiryTimes, entry.key)
			delete(ds.conditions, entry.key)

			// Notify any waiters that the key was deleted
			// (they're already deleted from conditions, but this is for consistency)
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
func (ds *DatastoreValue) WaitWithPredicate(evaluator *Evaluator, key string, predicateFn Value, timeout time.Duration) (any, error) {
	ds.dataMutex.Lock()

	// Get or create condition variable for this key
	cond, exists := ds.conditions[key]
	if !exists {
		cond = sync.NewCond(&ds.dataMutex)
		ds.conditions[key] = cond
	}

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
			// Start a goroutine that will broadcast on timeout
			timerDone := make(chan struct{})
			go func() {
				<-time.After(timeout)
				ds.dataMutex.Lock()
				cond.Broadcast()
				ds.dataMutex.Unlock()
				close(timerDone)
			}()

			// Record start time for checking actual timeout
			startTime := time.Now()
			cond.Wait() // Called with lock held - safe

			// Check if we actually timed out
			if time.Since(startTime) >= timeout {
				ds.dataMutex.Unlock()
				return nil, fmt.Errorf("wait() timeout exceeded for key %q", key)
			}
			// Otherwise, loop will re-check the condition
		} else {
			// No timeout - just wait
			cond.Wait()
		}
	}
}

// For array values, this means waiting for length to change (new append)
// If expectedValue is provided, waits until key equals that value
// Timeout is optional (pass 0 for no timeout)
// Returns the current value of the key after the condition is met, or error on timeout
func (ds *DatastoreValue) Wait(key string, expectedValue any, hasExpectedValue bool, timeout time.Duration) (any, error) {
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
			// Start a goroutine that will broadcast on timeout
			timerDone := make(chan struct{})
			go func() {
				<-time.After(timeout)
				ds.dataMutex.Lock()
				cond.Broadcast()
				ds.dataMutex.Unlock()
				close(timerDone)
			}()

			// Record start time for checking actual timeout
			startTime := time.Now()
			cond.Wait() // Called with lock held - safe

			// Check if we actually timed out
			if time.Since(startTime) >= timeout {
				ds.dataMutex.Unlock()
				return nil, fmt.Errorf("wait() timeout exceeded for key %q", key)
			}
			// Otherwise, loop will re-check the condition
		} else {
			// No timeout - just wait
			cond.Wait()
		}
	}
}

// Update atomically reads, deep merges updates into an object, and returns the updated object
// Creates an empty object if key doesn't exist
// Returns error if key exists but is not an object
// Supports nil values to delete keys from the object (shallow deletion only)
func (ds *DatastoreValue) Update(key string, updates any) (any, error) {
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

	// Write to WAL before applying to memory
	if err := ds.writeWAL(key, storedObj); err != nil {
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
	// Write nil to WAL to represent deletion
	if err := ds.writeWAL(key, nil); err != nil {
		return nil, err
	}

	ds.dataMutex.Lock()
	defer ds.dataMutex.Unlock()

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
	ds.dataMutex.Lock()
	defer ds.dataMutex.Unlock()

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
	return ds.loadFromDisk()
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

		// If result is not nil, include it
		if result.Data != nil {
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

	// Sync WAL before shutdown
	if ds.walFile != nil {
		_ = ds.syncWAL()
		_ = ds.walFile.Close()
	}

	// Final save if configured
	if ds.persistPath != "" {
		return ds.saveToDisk()
	}
	return nil
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

	// Open file for writing
	file, err := os.OpenFile(ds.persistPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open datastore %q at %q: %v", ds.namespace, ds.persistPath, err)
	}
	defer file.Close()

	// Encode to gob
	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(ds.data); err != nil {
		return fmt.Errorf("failed to serialize datastore %q: %v", ds.namespace, err)
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

// loadFromDisk deserializes the datastore from gob file
func (ds *DatastoreValue) loadFromDisk() error {
	if ds.persistPath == "" {
		return nil // No persistence configured
	}

	ds.fileWriteMutex.Lock()
	defer ds.fileWriteMutex.Unlock()

	// Open file (fail silently if not exists)
	file, err := os.Open(ds.persistPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist yet - OK
		}
		return fmt.Errorf("failed to read datastore %q from %q: %v", ds.namespace, ds.persistPath, err)
	}
	defer file.Close()

	ds.dataMutex.Lock()
	defer ds.dataMutex.Unlock()

	// Decode from gob
	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(&ds.data); err != nil {
		return fmt.Errorf("failed to deserialize datastore %q: %v", ds.namespace, err)
	}

	return nil
}

// recoverFromWAL replays WAL entries, saves merged state, and truncates WAL
func (ds *DatastoreValue) recoverFromWAL() error {
	if ds.walPath == "" {
		return nil
	}

	ds.walMutex.Lock()
	defer ds.walMutex.Unlock()

	// Replay WAL entries on top of loaded snapshot
	if err := ds.replayWAL(); err != nil {
		return fmt.Errorf("failed to replay WAL for %q: %v", ds.namespace, err)
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

	// Open WAL file for appending
	file, err := os.OpenFile(ds.walPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open WAL file %q: %v", ds.walPath, err)
	}

	ds.walFile = file
	ds.walEncoder = gob.NewEncoder(file)

	// Start WAL sync ticker if batching is configured
	if ds.walSyncInterval > 0 {
		ds.walSyncTicker = time.NewTicker(ds.walSyncInterval)
		ds.walStopSync = make(chan bool, 1)
		go func() {
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

// replayWAL reads all entries from the WAL file and applies them to data
func (ds *DatastoreValue) replayWAL() error {
	if ds.walPath == "" {
		return nil
	}

	// Check if WAL file exists
	file, err := os.Open(ds.walPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No WAL file yet - OK
		}
		return err
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	entryCount := 0
	for {
		var entry WALEntry
		if err := decoder.Decode(&entry); err != nil {
			if err.Error() == "EOF" {
				break // End of file
			}
			return fmt.Errorf("failed to decode WAL entry: %v", err)
		}

		// Apply entry to data (replays the exact key-value state)
		ds.data[entry.Key] = entry.Value
		entryCount++
	}

	return nil
}

// writeWAL appends a key-value entry to the WAL file
// Caller must hold dataMutex if this is part of an atomic operation
func (ds *DatastoreValue) writeWAL(key string, value any) error {
	if ds.walPath == "" || ds.walFile == nil {
		return nil // WAL not configured
	}

	ds.walMutex.Lock()
	defer ds.walMutex.Unlock()

	entry := WALEntry{Key: key, Value: value}
	if err := ds.walEncoder.Encode(entry); err != nil {
		return fmt.Errorf("failed to write WAL entry for key %q: %v", key, err)
	}

	// Sync immediately if configured (0 = sync every write)
	if ds.walSyncInterval == 0 {
		if err := ds.walFile.Sync(); err != nil {
			return fmt.Errorf("failed to sync WAL: %v", err)
		}
	}

	return nil
}

// syncWAL flushes buffered WAL writes to disk
func (ds *DatastoreValue) syncWAL() error {
	if ds.walFile == nil {
		return nil
	}

	ds.walMutex.Lock()
	defer ds.walMutex.Unlock()

	if err := ds.walFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync WAL: %v", err)
	}
	return nil
}

// truncateWAL clears the WAL after a successful snapshot save
func (ds *DatastoreValue) truncateWAL() error {
	if ds.walPath == "" {
		return nil
	}

	ds.walMutex.Lock()
	defer ds.walMutex.Unlock()

	// Close current WAL file if it's open
	if ds.walFile != nil {
		ds.walFile.Close()
		ds.walFile = nil
		ds.walEncoder = nil
	}

	// Truncate the WAL file (even if it wasn't previously open)
	if err := os.Truncate(ds.walPath, 0); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to truncate WAL %q: %v", ds.walPath, err)
		}
		// File doesn't exist - that's OK, nothing to truncate
	}

	// Reopen for appending
	file, err := os.OpenFile(ds.walPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to reopen WAL file %q after truncate: %v", ds.walPath, err)
	}

	ds.walFile = file
	ds.walEncoder = gob.NewEncoder(file)

	return nil
}

// valuesEqual compares two values for equality
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
