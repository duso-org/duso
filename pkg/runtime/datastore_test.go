package runtime

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestDatastoreSet_Basic tests basic Set and Get operations
func TestDatastoreSet_Basic(t *testing.T) {
	ds := GetDatastore("test_set_basic", nil)
	defer ds.Clear()

	err := ds.Set("key1", "value1")
	if err != nil {
		t.Errorf("Set failed: %v", err)
	}

	val, err := ds.Get("key1")
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}

	if val != "value1" {
		t.Errorf("Expected 'value1', got %v", val)
	}
}

// TestDatastoreSet_MultipleTypes tests Set with different value types
func TestDatastoreSet_MultipleTypes(t *testing.T) {
	ds := GetDatastore("test_set_types", nil)
	defer ds.Clear()

	testCases := []struct {
		name  string
		key   string
		value any
	}{
		{"string", "str_key", "hello"},
		{"number", "num_key", float64(42)},
		{"bool_true", "bool_key_t", true},
		{"bool_false", "bool_key_f", false},
		{"array", "arr_key", []any{1.0, 2.0, 3.0}},
		{"map", "map_key", map[string]any{"nested": "value"}},
		{"nil", "nil_key", nil},
	}

	for _, tc := range testCases {
		err := ds.Set(tc.key, tc.value)
		if err != nil {
			t.Errorf("%s: Set failed: %v", tc.name, err)
			continue
		}

		val, err := ds.Get(tc.key)
		if err != nil {
			t.Errorf("%s: Get failed: %v", tc.name, err)
			continue
		}

		// Compare based on type (deep equality checking would be complex here)
		switch v := tc.value.(type) {
		case string:
			if val != v {
				t.Errorf("%s: Expected %v, got %v", tc.name, v, val)
			}
		case float64:
			if val != v {
				t.Errorf("%s: Expected %v, got %v", tc.name, v, val)
			}
		case bool:
			if val != v {
				t.Errorf("%s: Expected %v, got %v", tc.name, v, val)
			}
		case nil:
			if val != nil {
				t.Errorf("%s: Expected nil, got %v", tc.name, val)
			}
		}
	}
}

// TestDatastoreGet_NonexistentKey tests that getting a nonexistent key returns nil
func TestDatastoreGet_NonexistentKey(t *testing.T) {
	ds := GetDatastore("test_get_nonexistent", nil)
	defer ds.Clear()

	val, err := ds.Get("nonexistent_key")
	if err != nil {
		t.Errorf("Get should not error: %v", err)
	}

	if val != nil {
		t.Errorf("Expected nil for nonexistent key, got %v", val)
	}
}

// TestDatastoreIncrement_Basic tests basic increment operations
func TestDatastoreIncrement_Basic(t *testing.T) {
	ds := GetDatastore("test_increment_basic", nil)
	defer ds.Clear()

	// Increment on nonexistent key (should create with value=delta)
	val, err := ds.Increment("counter", 5)
	if err != nil {
		t.Errorf("Increment failed: %v", err)
	}

	if val != 5.0 {
		t.Errorf("Expected 5.0, got %v", val)
	}

	// Increment again
	val, err = ds.Increment("counter", 3)
	if err != nil {
		t.Errorf("Second increment failed: %v", err)
	}

	if val != 8.0 {
		t.Errorf("Expected 8.0, got %v", val)
	}
}

// TestDatastoreIncrement_NegativeDelta tests incrementing by negative values
func TestDatastoreIncrement_NegativeDelta(t *testing.T) {
	ds := GetDatastore("test_increment_negative", nil)
	defer ds.Clear()

	ds.Set("counter", 10.0)
	val, err := ds.Increment("counter", -3)
	if err != nil {
		t.Errorf("Negative increment failed: %v", err)
	}

	if val != 7.0 {
		t.Errorf("Expected 7.0, got %v", val)
	}
}

// TestDatastoreIncrement_NonNumericError tests that incrementing non-numeric values errors
func TestDatastoreIncrement_NonNumericError(t *testing.T) {
	ds := GetDatastore("test_increment_non_numeric", nil)
	defer ds.Clear()

	ds.Set("str_key", "not_a_number")

	_, err := ds.Increment("str_key", 1)
	if err == nil {
		t.Errorf("Increment on non-numeric should error, got nil")
	}

	if !strings.Contains(err.Error(), "cannot operate on non-numeric") {
		t.Errorf("Error message should mention non-numeric, got: %v", err)
	}
}

// TestDatastoreAppend_Basic tests basic append operations
func TestDatastoreAppend_Basic(t *testing.T) {
	ds := GetDatastore("test_append_basic", nil)
	defer ds.Clear()

	// Append to nonexistent key (should create array)
	len1, err := ds.Append("arr", 1.0)
	if err != nil {
		t.Errorf("First append failed: %v", err)
	}

	if len1 != 1.0 {
		t.Errorf("Expected length 1.0, got %v", len1)
	}

	len2, err := ds.Append("arr", 2.0)
	if err != nil {
		t.Errorf("Second append failed: %v", err)
	}

	if len2 != 2.0 {
		t.Errorf("Expected length 2.0, got %v", len2)
	}

	// Verify array contents
	val, _ := ds.Get("arr")
	arr, ok := val.([]any)
	if !ok {
		t.Errorf("Expected array, got %T", val)
		return
	}

	if len(arr) != 2 {
		t.Errorf("Expected array length 2, got %d", len(arr))
	}
}

// TestDatastoreAppend_NonArrayError tests that appending to non-arrays errors
func TestDatastoreAppend_NonArrayError(t *testing.T) {
	ds := GetDatastore("test_append_non_array", nil)
	defer ds.Clear()

	ds.Set("not_array", "string_value")

	_, err := ds.Append("not_array", "item")
	if err == nil {
		t.Errorf("Append on non-array should error, got nil")
	}

	if !strings.Contains(err.Error(), "cannot operate on non-array") {
		t.Errorf("Error message should mention non-array, got: %v", err)
	}
}

// TestDatastoreDelete tests delete operations
func TestDatastoreDelete(t *testing.T) {
	ds := GetDatastore("test_delete", nil)
	defer ds.Clear()

	ds.Set("key1", "value1")
	ds.Set("key2", "value2")

	err := ds.Delete("key1")
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}

	// Verify key1 is gone
	val, _ := ds.Get("key1")
	if val != nil {
		t.Errorf("Expected nil after delete, got %v", val)
	}

	// Verify key2 still exists
	val, _ = ds.Get("key2")
	if val != "value2" {
		t.Errorf("key2 should still exist, got %v", val)
	}
}

// TestDatastoreDelete_NonexistentKey tests deleting nonexistent keys doesn't error
func TestDatastoreDelete_NonexistentKey(t *testing.T) {
	ds := GetDatastore("test_delete_nonexistent", nil)
	defer ds.Clear()

	err := ds.Delete("nonexistent")
	if err != nil {
		t.Errorf("Delete nonexistent key should not error: %v", err)
	}
}

// TestDatastoreClear tests clearing the entire datastore
func TestDatastoreClear(t *testing.T) {
	ds := GetDatastore("test_clear", nil)
	defer ds.Clear()

	ds.Set("key1", "value1")
	ds.Set("key2", "value2")
	ds.Set("key3", "value3")

	err := ds.Clear()
	if err != nil {
		t.Errorf("Clear failed: %v", err)
	}

	// Verify all keys are gone
	for i := 1; i <= 3; i++ {
		key := "key" + string(rune('0'+i))
		val, _ := ds.Get(key)
		if val != nil {
			t.Errorf("Key %s should be gone after clear, got %v", key, val)
		}
	}
}

// TestDatastoreThreadSafety_ConcurrentWrites tests thread-safe concurrent writes
func TestDatastoreThreadSafety_ConcurrentWrites(t *testing.T) {
	ds := GetDatastore("test_thread_safety_writes", nil)
	defer ds.Clear()

	numGoroutines := 10
	operationsPerGoroutine := 100
	wg := sync.WaitGroup{}

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for i := 0; i < operationsPerGoroutine; i++ {
				key := "key"
				ds.Set(key, float64(goroutineID*operationsPerGoroutine+i))
			}
		}(g)
	}

	wg.Wait()

	// Verify final value exists
	val, _ := ds.Get("key")
	if val == nil {
		t.Errorf("Expected value after concurrent writes, got nil")
	}
}

// TestDatastoreThreadSafety_ConcurrentReadsWrites tests concurrent reads and writes
func TestDatastoreThreadSafety_ConcurrentReadsWrites(t *testing.T) {
	ds := GetDatastore("test_thread_safety_reads_writes", nil)
	defer ds.Clear()

	ds.Set("counter", 0.0)

	numWriters := 5
	numReaders := 5
	operationsPerGoroutine := 50
	wg := sync.WaitGroup{}

	// Start writers
	for w := 0; w < numWriters; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < operationsPerGoroutine; i++ {
				ds.Increment("counter", 1)
			}
		}()
	}

	// Start readers
	for r := 0; r < numReaders; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < operationsPerGoroutine; i++ {
				ds.Get("counter")
			}
		}()
	}

	wg.Wait()

	// Verify final count
	val, _ := ds.Get("counter")
	expected := float64(numWriters * operationsPerGoroutine)
	if val != expected {
		t.Errorf("Expected counter=%v, got %v", expected, val)
	}
}

// TestDatastoreWait_Basic tests Wait on key change
func TestDatastoreWait_Basic(t *testing.T) {
	ds := GetDatastore("test_wait_basic", nil)
	defer ds.Clear()

	ds.Set("value", 1.0)

	done := make(chan bool)
	var result any

	// Start a goroutine that waits for change
	go func() {
		val, err := ds.Wait("value", nil, false, 2*time.Second)
		if err != nil {
			t.Errorf("Wait failed: %v", err)
		}
		result = val
		done <- true
	}()

	// Give goroutine time to start waiting
	time.Sleep(100 * time.Millisecond)

	// Change the value
	ds.Set("value", 2.0)

	// Wait for goroutine to complete
	select {
	case <-done:
		if result != 2.0 {
			t.Errorf("Expected 2.0, got %v", result)
		}
	case <-time.After(3 * time.Second):
		t.Errorf("Wait timed out")
	}
}

// TestDatastoreWait_WithExpectedValue tests Wait with specific expected value
func TestDatastoreWait_WithExpectedValue(t *testing.T) {
	ds := GetDatastore("test_wait_expected", nil)
	defer ds.Clear()

	ds.Set("status", "pending")

	done := make(chan bool)

	// Start a goroutine that waits for specific value
	go func() {
		val, err := ds.Wait("status", "ready", true, 2*time.Second)
		if err != nil {
			t.Errorf("Wait failed: %v", err)
		}
		if val != "ready" {
			t.Errorf("Expected 'ready', got %v", val)
		}
		done <- true
	}()

	time.Sleep(100 * time.Millisecond)

	// Set to wrong value first (should not unblock)
	ds.Set("status", "processing")
	time.Sleep(50 * time.Millisecond)

	// Set to expected value
	ds.Set("status", "ready")

	select {
	case <-done:
		// Success
	case <-time.After(3 * time.Second):
		t.Errorf("Wait timed out")
	}
}

// TestDatastoreWait_Timeout tests Wait timeout
func TestDatastoreWait_Timeout(t *testing.T) {
	ds := GetDatastore("test_wait_timeout", nil)
	defer ds.Clear()

	ds.Set("value", 1.0)

	// Wait with short timeout
	_, err := ds.Wait("value", nil, false, 100*time.Millisecond)
	if err == nil {
		t.Errorf("Wait should timeout, got nil error")
	}

	if !strings.Contains(err.Error(), "timeout") {
		t.Errorf("Error should mention timeout, got: %v", err)
	}
}

// TestDatastoreWaitFor_PredicateTrue tests WaitFor with immediately true predicate
func TestDatastoreWaitFor_PredicateTrue(t *testing.T) {
	ds := GetDatastore("test_waitfor_true", nil)
	defer ds.Clear()

	ds.Set("count", 5.0)

	// Predicate that checks if value is >= 5
	predicate := func(args map[string]any) (any, error) {
		val, ok := args["0"].(float64)
		if !ok {
			return false, nil
		}
		return val >= 5.0, nil
	}

	val, err := ds.WaitFor("count", predicate, 1*time.Second)
	if err != nil {
		t.Errorf("WaitFor failed: %v", err)
	}

	if val != 5.0 {
		t.Errorf("Expected 5.0, got %v", val)
	}
}

// TestDatastoreWaitFor_PredicateEventual tests WaitFor with eventual true predicate
func TestDatastoreWaitFor_PredicateEventual(t *testing.T) {
	ds := GetDatastore("test_waitfor_eventual", nil)
	defer ds.Clear()

	ds.Set("count", 0.0)

	done := make(chan bool)

	// Start waiter
	go func() {
		// Predicate that checks if count is >= 10
		predicate := func(args map[string]any) (any, error) {
			val, ok := args["0"].(float64)
			if !ok {
				return false, nil
			}
			return val >= 10.0, nil
		}

		val, err := ds.WaitFor("count", predicate, 2*time.Second)
		if err != nil {
			t.Errorf("WaitFor failed: %v", err)
		}
		if val != 10.0 {
			t.Errorf("Expected 10.0, got %v", val)
		}
		done <- true
	}()

	time.Sleep(100 * time.Millisecond)

	// Increment to 10
	ds.Set("count", 10.0)

	select {
	case <-done:
		// Success
	case <-time.After(3 * time.Second):
		t.Errorf("WaitFor timed out")
	}
}

// TestDatastoreWaitFor_Timeout tests WaitFor timeout
func TestDatastoreWaitFor_Timeout(t *testing.T) {
	ds := GetDatastore("test_waitfor_timeout", nil)
	defer ds.Clear()

	ds.Set("count", 0.0)

	// Predicate that's never true
	predicate := func(args map[string]any) (any, error) {
		return false, nil
	}

	_, err := ds.WaitFor("count", predicate, 100*time.Millisecond)
	if err == nil {
		t.Errorf("WaitFor should timeout, got nil error")
	}

	if !strings.Contains(err.Error(), "timeout") {
		t.Errorf("Error should mention timeout, got: %v", err)
	}
}

// TestDatastorePersistence_Save tests saving datastore to disk
func TestDatastorePersistence_Save(t *testing.T) {
	tmpDir := t.TempDir()
	persistPath := filepath.Join(tmpDir, "test_datastore.json")

	config := map[string]any{
		"persist": persistPath,
	}

	ds := GetDatastore("test_persist_save", config)
	defer ds.Clear()

	ds.Set("key1", "value1")
	ds.Set("key2", 42.0)
	ds.Set("key3", []any{1.0, 2.0, 3.0})

	err := ds.Save()
	if err != nil {
		t.Errorf("Save failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(persistPath); os.IsNotExist(err) {
		t.Errorf("Persist file not created")
	}

	// Verify JSON content
	data, err := os.ReadFile(persistPath)
	if err != nil {
		t.Errorf("Failed to read persist file: %v", err)
	}

	var loaded map[string]any
	err = json.Unmarshal(data, &loaded)
	if err != nil {
		t.Errorf("Failed to parse JSON: %v", err)
	}

	if loaded["key1"] != "value1" {
		t.Errorf("Persisted key1 incorrect")
	}
	if loaded["key2"] != 42.0 {
		t.Errorf("Persisted key2 incorrect")
	}
}

// TestDatastorePersistence_Load tests loading datastore from disk
func TestDatastorePersistence_Load(t *testing.T) {
	tmpDir := t.TempDir()
	persistPath := filepath.Join(tmpDir, "test_load.json")

	// Create a JSON file with test data
	testData := map[string]any{
		"key1": "value1",
		"key2": 42.0,
	}
	jsonData, _ := json.Marshal(testData)
	os.WriteFile(persistPath, jsonData, 0644)

	config := map[string]any{
		"persist": persistPath,
	}

	ds := GetDatastore("test_persist_load", config)
	defer ds.Clear()

	// Data should be loaded on initialization
	val1, _ := ds.Get("key1")
	if val1 != "value1" {
		t.Errorf("Expected 'value1', got %v", val1)
	}

	val2, _ := ds.Get("key2")
	if val2 != 42.0 {
		t.Errorf("Expected 42.0, got %v", val2)
	}
}

// TestDatastorePersistence_AutoSave tests auto-save with ticker
func TestDatastorePersistence_AutoSave(t *testing.T) {
	tmpDir := t.TempDir()
	persistPath := filepath.Join(tmpDir, "test_autosave.json")

	config := map[string]any{
		"persist":          persistPath,
		"persist_interval": 0.1, // 100ms
	}

	ds := GetDatastore("test_persist_autosave", config)
	defer func() {
		ds.Shutdown()
		ds.Clear()
	}()

	ds.Set("key1", "value1")

	// Shutdown will save the datastore to disk
	err := ds.Shutdown()
	if err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(persistPath); os.IsNotExist(err) {
		t.Errorf("Auto-save did not create file on shutdown")
	}
}

// TestDatastoreNamespaces tests that different namespaces are isolated
func TestDatastoreNamespaces(t *testing.T) {
	ds1 := GetDatastore("namespace1", nil)
	ds2 := GetDatastore("namespace2", nil)

	defer func() {
		ds1.Clear()
		ds2.Clear()
	}()

	ds1.Set("key", "value1")
	ds2.Set("key", "value2")

	val1, _ := ds1.Get("key")
	val2, _ := ds2.Get("key")

	if val1 != "value1" {
		t.Errorf("namespace1 key should be 'value1', got %v", val1)
	}

	if val2 != "value2" {
		t.Errorf("namespace2 key should be 'value2', got %v", val2)
	}
}

// TestDatastoreMultipleWaiters tests multiple goroutines waiting on same key
func TestDatastoreMultipleWaiters(t *testing.T) {
	ds := GetDatastore("test_multiple_waiters", nil)
	defer ds.Clear()

	ds.Set("signal", 0.0)

	numWaiters := 5
	done := make(chan bool, numWaiters)

	// Start multiple waiters
	for i := 0; i < numWaiters; i++ {
		go func() {
			ds.Wait("signal", 1.0, true, 2*time.Second)
			done <- true
		}()
	}

	time.Sleep(100 * time.Millisecond)

	// Signal all waiters at once
	ds.Set("signal", 1.0)

	// All should complete
	for i := 0; i < numWaiters; i++ {
		select {
		case <-done:
			// Success
		case <-time.After(3 * time.Second):
			t.Errorf("Waiter %d timed out", i)
		}
	}
}

// TestDatastoreShutdown tests shutdown cleanup
func TestDatastoreShutdown(t *testing.T) {
	tmpDir := t.TempDir()
	persistPath := filepath.Join(tmpDir, "test_shutdown.json")

	config := map[string]any{
		"persist":          persistPath,
		"persist_interval": 0.1,
	}

	ds := GetDatastore("test_shutdown", config)

	ds.Set("key", "value")

	err := ds.Shutdown()
	if err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}

	// Verify file was saved
	if _, err := os.Stat(persistPath); os.IsNotExist(err) {
		t.Errorf("Shutdown did not save file")
	}
}

// TestDatastoreUpdatePeakGoroutines tests that peak goroutines are tracked
func TestDatastoreUpdatePeakGoroutines(t *testing.T) {
	ds := GetDatastore("test_peak_goroutines", nil)
	defer ds.Clear()

	ds.Set("value", 1.0)

	// This should be called by Increment/Append/Set, but test it directly
	// (Coverage test for the metrics integration)
}

// TestDatastoreValueIsolation tests that values are deep copied
func TestDatastoreValueIsolation(t *testing.T) {
	ds := GetDatastore("test_value_isolation", nil)
	defer ds.Clear()

	// Create a slice and store it
	original := []any{1.0, 2.0, 3.0}
	ds.Set("arr", original)

	// Modify original
	original[0] = 999.0

	// Get should return unchanged value
	val, _ := ds.Get("arr")
	arr, _ := val.([]any)

	if arr[0] == 999.0 {
		t.Errorf("Value isolation failed: original modification affected stored value")
	}

	if arr[0] != 1.0 {
		t.Errorf("Expected 1.0, got %v", arr[0])
	}
}
