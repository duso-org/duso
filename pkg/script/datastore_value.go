package script

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// Global registry of namespaced datastores
var (
	datastoreRegistry = make(map[string]*DatastoreValue)
	registryMutex     sync.RWMutex
)

// DatastoreValue represents an in-memory thread-safe key/value store
// scoped to a specific namespace. Multiple scripts can access the same
// store by using the same namespace. Optionally persists to JSON.
type DatastoreValue struct {
	namespace         string
	data              map[string]any
	dataMutex         sync.RWMutex
	conditions        map[string]*sync.Cond // Per-key condition variables for wait operations
	persistPath       string                // Optional: path to JSON file
	persistInterval   time.Duration         // Optional: auto-save interval
	ticker            *time.Ticker          // Auto-save ticker
	stopTicker        chan bool              // Signal to stop ticker
	fileWriteMutex    sync.Mutex             // Serialize file writes
	statsFn           func(key string) any  // Function to compute stats dynamically (for sys datastore)
}

// GetDatastore returns or creates a namespaced datastore with optional persistence config
func GetDatastore(namespace string, config map[string]any) *DatastoreValue {
	registryMutex.Lock()
	defer registryMutex.Unlock()

	if store, exists := datastoreRegistry[namespace]; exists {
		return store
	}

	store := &DatastoreValue{
		namespace:      namespace,
		data:           make(map[string]any),
		conditions:     make(map[string]*sync.Cond),
		stopTicker:     make(chan bool, 1),
	}

	// For sys datastore, set up dynamic metric computation
	if namespace == "sys" {
		store.statsFn = GetMetric
	}

	// Parse persistence config
	if config != nil {
		if persistPath, ok := config["persist"]; ok {
			store.persistPath = fmt.Sprintf("%v", persistPath)
		}
		if persistInterval, ok := config["persist_interval"]; ok {
			if intervalSecs, ok := persistInterval.(float64); ok {
				store.persistInterval = time.Duration(intervalSecs) * time.Second
			}
		}
	}

	// Auto-load from disk if file exists
	if store.persistPath != "" {
		_ = store.loadFromDisk() // Ignore error if file doesn't exist yet
	}

	// Start auto-save ticker if configured
	if store.persistInterval > 0 {
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

	datastoreRegistry[namespace] = store
	return store
}

// Set stores a value by key (thread-safe)
func (ds *DatastoreValue) Set(key string, value any) error {
	ds.dataMutex.Lock()
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
	ds.dataMutex.Lock()

	// Check if key already exists
	if _, exists := ds.data[key]; exists {
		ds.dataMutex.Unlock()
		return false // Key already exists, don't overwrite
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
	ds.dataMutex.RLock()
	defer ds.dataMutex.RUnlock()

	// Check for dynamic stats computation (e.g., memory stats)
	if ds.statsFn != nil {
		if val := ds.statsFn(key); val != nil {
			return val, nil
		}
	}

	value, exists := ds.data[key]
	if !exists {
		return nil, nil // Return nil if key doesn't exist
	}

	// Deep copy to isolate returned values from datastore's scope
	// Prevents concurrent requests from accidentally sharing mutable data
	return DeepCopyAny(value), nil
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

	if val, exists := ds.data[key]; exists {
		// Key exists - must be an array
		if arr, ok := val.([]any); ok {
			// Deep copy the item before appending
			arr = append(arr, DeepCopyAny(item))
			ds.data[key] = arr
			// Notify any waiters on this key (value changed)
			if cond, exists := ds.conditions[key]; exists {
				cond.Broadcast()
			}
			return float64(len(arr)), nil
		}
		// Handle *[]Value (when passing Duso array back in)
		if _, ok := val.(*[]Value); ok {
			// This shouldn't happen since Set should convert *[]Value to []any
			return 0, fmt.Errorf("push() found unexpected *[]Value at key %q (should be []any)", key)
		}
		return 0, fmt.Errorf("push() cannot operate on non-array value at key %q", key)
	}

	// Key doesn't exist - create new array with the item
	arr := []any{DeepCopyAny(item)}
	ds.data[key] = arr

	// Notify any waiters on this key (value changed)
	if cond, exists := ds.conditions[key]; exists {
		cond.Broadcast()
	}

	return 1, nil
}

// Wait blocks until the key changes (if no expectedValue) or equals expectedValue (if provided)
// If expectedValue is nil (omitted), waits for ANY change to the key
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

// WaitFor blocks until predicate(value) returns true
// For array values, predicate receives the array length as a number
// Predicate is a Duso function that takes one argument and returns a boolean
// Timeout is optional (pass 0 for no timeout)
// Returns the current value of the key after the predicate is true, or error on timeout
func (ds *DatastoreValue) WaitFor(evaluator *Evaluator, key string, predicateFn GoFunction, timeout time.Duration) (any, error) {
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
			// For arrays, pass the length to predicate. Otherwise pass the value itself.
			predicateArg := current
			if isArray(current) {
				predicateArg = float64(getLength(current))
			}

			// Call the predicate function directly (it's a GoFunction func type)
			result, err := predicateFn(evaluator, map[string]any{"0": predicateArg})
			if err != nil {
				ds.dataMutex.Unlock()
				return nil, fmt.Errorf("waitFor() predicate error: %v", err)
			}
			if resultBool, ok := result.(bool); ok && resultBool {
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
				return nil, fmt.Errorf("waitFor() timeout exceeded for key %q", key)
			}
			// Otherwise, loop will re-check the condition
		} else {
			// No timeout - just wait
			cond.Wait()
		}
	}
}

// Delete removes a key from the store
func (ds *DatastoreValue) Delete(key string) error {
	ds.dataMutex.Lock()
	defer ds.dataMutex.Unlock()

	delete(ds.data, key)
	delete(ds.conditions, key)

	return nil
}

// Clear removes all keys from the store
func (ds *DatastoreValue) Clear() error {
	ds.dataMutex.Lock()
	defer ds.dataMutex.Unlock()

	ds.data = make(map[string]any)
	ds.conditions = make(map[string]*sync.Cond)

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

// Shutdown stops the auto-save ticker and saves final state
func (ds *DatastoreValue) Shutdown() error {
	if ds.ticker != nil {
		ds.ticker.Stop()
		select {
		case ds.stopTicker <- true:
		default:
		}
	}

	// Final save if configured
	if ds.persistPath != "" {
		return ds.saveToDisk()
	}
	return nil
}

// saveToDisk serializes the datastore to JSON file and flushes to disk
func (ds *DatastoreValue) saveToDisk() error {
	if ds.persistPath == "" {
		return nil // No persistence configured
	}

	ds.fileWriteMutex.Lock()
	defer ds.fileWriteMutex.Unlock()

	ds.dataMutex.RLock()
	defer ds.dataMutex.RUnlock()

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(ds.data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize datastore %q: %v", ds.namespace, err)
	}

	// Open file for writing
	file, err := os.OpenFile(ds.persistPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open datastore %q at %q: %v", ds.namespace, ds.persistPath, err)
	}
	defer file.Close()

	// Write JSON data
	if _, err := file.Write(jsonData); err != nil {
		return fmt.Errorf("failed to write datastore %q to %q: %v", ds.namespace, ds.persistPath, err)
	}

	// Flush to disk to ensure data hits storage
	if err := file.Sync(); err != nil {
		return fmt.Errorf("failed to sync datastore %q to disk: %v", ds.namespace, err)
	}

	return nil
}

// loadFromDisk deserializes the datastore from JSON file
func (ds *DatastoreValue) loadFromDisk() error {
	if ds.persistPath == "" {
		return nil // No persistence configured
	}

	ds.fileWriteMutex.Lock()
	defer ds.fileWriteMutex.Unlock()

	// Read file (fail silently if not exists)
	jsonData, err := os.ReadFile(ds.persistPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist yet - OK
		}
		return fmt.Errorf("failed to read datastore %q from %q: %v", ds.namespace, ds.persistPath, err)
	}

	ds.dataMutex.Lock()
	defer ds.dataMutex.Unlock()

	// Unmarshal from JSON
	if err := json.Unmarshal(jsonData, &ds.data); err != nil {
		return fmt.Errorf("failed to deserialize datastore %q: %v", ds.namespace, err)
	}

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

	// Fall back to interface equality (for nil, maps, etc.)
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
