package script

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// Helper to execute Duso code and capture stdout
func testDatastore(t *testing.T, code string, expected string) {
	// Capture stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	oldStdout := os.Stdout
	os.Stdout = w

	interp := NewInterpreter(false)
	_, execErr := interp.Execute(code)

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var output strings.Builder
	_, err = io.Copy(&output, r)
	r.Close()
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	if execErr != nil {
		t.Fatalf("execution error: %v", execErr)
	}

	if output.String() != expected {
		t.Errorf("expected %q, got %q", expected, output.String())
	}
}

// Helper to verify code produces an error
func testDatastoreError(t *testing.T, code string) {
	interp := NewInterpreter(false)
	_, err := interp.Execute(code)
	if err == nil {
		t.Fatal("expected error but execution succeeded")
	}
}

// TestDatastore_Basic tests basic set/get operations
func TestDatastore_Basic(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{
			"set and get string",
			`store = datastore("test_basic")
store.set("name", "Alice")
print(store.get("name"))`,
			"Alice\n",
		},
		{
			"set and get number",
			`store = datastore("test_basic_num")
store.set("count", 42)
print(store.get("count"))`,
			"42\n",
		},
		{
			"set and get boolean",
			`store = datastore("test_basic_bool")
store.set("flag", true)
print(store.get("flag"))`,
			"true\n",
		},
		{
			"get nonexistent key returns nil",
			`store = datastore("test_basic_nil")
print(store.get("missing"))`,
			"nil\n",
		},
		{
			"set and get array",
			`store = datastore("test_basic_arr")
store.set("items", [1, 2, 3])
print(store.get("items"))`,
			"[1, 2, 3]\n",
		},
		{
			"set and get object",
			`store = datastore("test_basic_obj")
store.set("config", {host = "localhost", port = 5432})
result = store.get("config")
print(result.host)`,
			"localhost\n",
		},
		{
			"overwrite existing key",
			`store = datastore("test_basic_overwrite")
store.set("x", 10)
store.set("x", 20)
print(store.get("x"))`,
			"20\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDatastore(t, tt.code, tt.expected)
		})
	}
}

// TestDatastore_Increment tests atomic increment operations
func TestDatastore_Increment(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{
			"increment nonexistent key starts at 0",
			`store = datastore("test_incr_new")
result = store.increment("counter", 5)
print(result)`,
			"5\n",
		},
		{
			"increment existing value",
			`store = datastore("test_incr_existing")
store.set("counter", 10)
result = store.increment("counter", 5)
print(result)`,
			"15\n",
		},
		{
			"multiple increments",
			`store = datastore("test_incr_multi")
store.increment("counter", 1)
store.increment("counter", 2)
store.increment("counter", 3)
print(store.get("counter"))`,
			"6\n",
		},
		{
			"increment negative delta",
			`store = datastore("test_incr_negative")
store.set("counter", 10)
result = store.increment("counter", -3)
print(result)`,
			"7\n",
		},
		{
			"increment float values",
			`store = datastore("test_incr_float")
store.set("balance", 100.5)
result = store.increment("balance", 50.25)
print(result)`,
			"150.75\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDatastore(t, tt.code, tt.expected)
		})
	}
}

// TestDatastore_Push tests atomic array append operations
func TestDatastore_Push(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{
			"append to nonexistent key creates array",
			`store = datastore("test_append_new")
len = store.push("items", "first")
print(len)
print(store.get("items"))`,
			"1\n[first]\n",
		},
		{
			"append multiple items",
			`store = datastore("test_append_multi")
store.push("items", "a")
store.push("items", "b")
store.push("items", "c")
print(store.get("items"))`,
			"[a, b, c]\n",
		},
		{
			"append returns new length",
			`store = datastore("test_append_len")
len1 = store.push("items", 1)
len2 = store.push("items", 2)
len3 = store.push("items", 3)
print(len1)
print(len2)
print(len3)`,
			"1\n2\n3\n",
		},
		{
			"append various types",
			`store = datastore("test_append_types")
store.push("mixed", 42)
store.push("mixed", "text")
store.push("mixed", true)
store.push("mixed", {key = "value"})
print(format_json(store.get("mixed")))`,
			"[42,\"text\",true,{\"key\":\"value\"}]\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDatastore(t, tt.code, tt.expected)
		})
	}
}

// TestDatastore_Delete tests key deletion
func TestDatastore_Delete(t *testing.T) {
	code := `store = datastore("test_delete")
store.set("x", 10)
store.set("y", 20)
store.delete("x")
print(store.get("x"))
print(store.get("y"))
`
	testDatastore(t, code, "nil\n20\n")
}

// TestDatastore_Clear tests clearing all data
func TestDatastore_Clear(t *testing.T) {
	code := `store = datastore("test_clear")
store.set("a", 1)
store.set("b", 2)
store.set("c", 3)
store.clear()
print(store.get("a"))
print(store.get("b"))
print(store.get("c"))
`
	testDatastore(t, code, "nil\nnil\nnil\n")
}

// TestDatastore_Namespacing tests namespace isolation
func TestDatastore_Namespacing(t *testing.T) {
	code := `store1 = datastore("namespace_1")
store2 = datastore("namespace_2")

store1.set("key", "value1")
store2.set("key", "value2")

print(store1.get("key"))
print(store2.get("key"))
`
	testDatastore(t, code, "value1\nvalue2\n")
}

// TestDatastore_SameNamespace tests that same namespace returns same store
func TestDatastore_SameNamespace(t *testing.T) {
	code := `store1 = datastore("shared_namespace")
store1.set("counter", 10)

store2 = datastore("shared_namespace")
result = store2.get("counter")
print(result)

store2.increment("counter", 5)
print(store1.get("counter"))
`
	testDatastore(t, code, "10\n15\n")
}

// TestDatastore_ErrorOnIncrementNonNumeric tests error when incrementing non-numeric
func TestDatastore_ErrorOnIncrementNonNumeric(t *testing.T) {
	code := `store = datastore("test_incr_error")
store.set("text", "hello")
store.increment("text", 1)
`
	testDatastoreError(t, code)
}

// TestDatastore_ErrorOnPushNonArray tests error when appending to non-array
func TestDatastore_ErrorOnPushNonArray(t *testing.T) {
	code := `store = datastore("test_append_error")
store.set("scalar", 42)
store.push("scalar", "item")
`
	testDatastoreError(t, code)
}

// TestDatastore_WaitWithConcurrentUpdate tests wait with concurrent value update
func TestDatastore_WaitWithConcurrentUpdate(t *testing.T) {
	namespace := "test_wait_concurrent"
	store := GetDatastore(namespace, nil)

	// Set initial value
	store.Set("status", "pending")

	// Channel to signal when wait completes
	done := make(chan any, 1)
	waitErr := make(chan error, 1)

	// Goroutine to wait for change
	go func() {
		result, err := store.Wait("status", nil, false, 5*time.Second)
		if err != nil {
			waitErr <- err
		} else {
			done <- result
		}
	}()

	// Give goroutine time to start waiting
	time.Sleep(100 * time.Millisecond)

	// Update value from another goroutine
	store.Set("status", "done")

	// Wait for result with timeout
	select {
	case result := <-done:
		if result != "done" {
			t.Errorf("expected 'done', got %v", result)
		}
	case err := <-waitErr:
		t.Errorf("wait failed: %v", err)
	case <-time.After(2 * time.Second):
		t.Errorf("wait timed out")
	}
}

// TestDatastore_WaitForExpectedValue tests wait returns when value matches expected
func TestDatastore_WaitForExpectedValue(t *testing.T) {
	namespace := "test_wait_expected_value"
	store := GetDatastore(namespace, nil)

	store.Set("counter", float64(5))

	done := make(chan any, 1)
	waitErr := make(chan error, 1)

	go func() {
		result, err := store.Wait("counter", float64(10), true, 5*time.Second)
		if err != nil {
			waitErr <- err
		} else {
			done <- result
		}
	}()

	time.Sleep(100 * time.Millisecond)

	// Update to expected value
	store.Set("counter", float64(10))

	select {
	case result := <-done:
		if result != float64(10) {
			t.Errorf("expected 10, got %v", result)
		}
	case err := <-waitErr:
		t.Errorf("wait failed: %v", err)
	case <-time.After(2 * time.Second):
		t.Errorf("wait timed out")
	}
}

// TestDatastore_WaitTimeout tests wait times out when value never changes
func TestDatastore_WaitTimeout(t *testing.T) {
	namespace := "test_wait_timeout_val"
	store := GetDatastore(namespace, nil)

	store.Set("status", "initial")

	done := make(chan any, 1)
	waitErr := make(chan error, 1)

	go func() {
		// Wait for a value that will never be set, with short timeout
		result, err := store.Wait("status", "never_happens", true, 100*time.Millisecond)
		if err != nil {
			waitErr <- err
		} else {
			done <- result
		}
	}()

	// Expect timeout error
	select {
	case <-done:
		t.Errorf("expected timeout error, but wait succeeded")
	case err := <-waitErr:
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		// Expected - timeout should occur
	case <-time.After(2 * time.Second):
		t.Errorf("test timed out waiting for wait timeout")
	}
}

// TestDatastore_WaitArrayChange tests wait detects array changes
func TestDatastore_WaitArrayChange(t *testing.T) {
	namespace := "test_wait_array_change"
	store := GetDatastore(namespace, nil)

	// Set initial empty array
	store.Set("items", []any{})

	done := make(chan any, 1)
	waitErr := make(chan error, 1)

	go func() {
		result, err := store.Wait("items", nil, false, 5*time.Second)
		if err != nil {
			waitErr <- err
		} else {
			done <- result
		}
	}()

	time.Sleep(100 * time.Millisecond)

	// Push to array
	store.Push("items", "first")

	select {
	case result := <-done:
		if arr, ok := result.([]any); !ok || len(arr) != 1 {
			t.Errorf("expected array with 1 item, got %v", result)
		}
	case err := <-waitErr:
		t.Errorf("wait failed: %v", err)
	case <-time.After(2 * time.Second):
		t.Errorf("wait timed out")
	}
}

// TestDatastore_Persistence tests save and load functionality
func TestDatastore_Persistence(t *testing.T) {
	// Create temporary file
	tmpdir := t.TempDir()
	datafile := filepath.Join(tmpdir, "test_data.json")

	code := fmt.Sprintf(`store = datastore("persist_test", {persist = %q})
store.set("name", "Alice")
store.set("count", 42)
store.save()
print("saved")
`, datafile)

	testDatastore(t, code, "saved\n")

	// Verify file was created
	if _, err := os.Stat(datafile); os.IsNotExist(err) {
		t.Fatalf("persistence file not created at %s", datafile)
	}

	// Now load in a new datastore instance with different namespace to avoid cache
	code2 := fmt.Sprintf(`store2 = datastore("persist_test_new", {persist = %q})
store2.load()
print(store2.get("name"))
print(store2.get("count"))
`, datafile)

	testDatastore(t, code2, "Alice\n42\n")
}

// TestDatastore_PersistenceAutoLoad tests auto-load on creation
func TestDatastore_PersistenceAutoLoad(t *testing.T) {
	tmpdir := t.TempDir()
	datafile := filepath.Join(tmpdir, "autoload_data.json")

	// First: Create and save data
	code1 := fmt.Sprintf(`store = datastore("autoload_test", {persist = %q})
store.set("value", 123)
store.save()
`, datafile)
	testDatastore(t, code1, "")

	// Second: Create new interpreter and let it auto-load
	code2 := fmt.Sprintf(`store = datastore("autoload_test2", {persist = %q})
print(store.get("value"))
`, datafile)
	testDatastore(t, code2, "123\n")
}

// TestDatastore_ComplexData tests storing complex nested structures
func TestDatastore_ComplexData(t *testing.T) {
	code := `store = datastore("test_complex")
data = {
  users = [
    {name = "Alice", age = 30},
    {name = "Bob", age = 25}
  ],
  config = {
    timeout = 30,
    retries = 3
  }
}
store.set("app_state", data)
result = store.get("app_state")
print(result.users[0].name)
print(result.config.timeout)
`
	testDatastore(t, code, "Alice\n30\n")
}

// TestDatastore_IncrementReturnsValue tests increment returns the new value
func TestDatastore_IncrementReturnsValue(t *testing.T) {
	code := `store = datastore("test_incr_return")
v1 = store.increment("counter", 10)
v2 = store.increment("counter", 5)
v3 = store.increment("counter", -3)
print(v1)
print(v2)
print(v3)
`
	testDatastore(t, code, "10\n15\n12\n")
}

// TestDatastore_PushReturnsLength tests append returns array length
func TestDatastore_PushReturnsLength(t *testing.T) {
	code := `store = datastore("test_append_return")
len1 = store.push("arr", "x")
len2 = store.push("arr", "y")
len3 = store.push("arr", "z")
print(len1)
print(len2)
print(len3)
`
	testDatastore(t, code, "1\n2\n3\n")
}

// TestDatastore_TypeCoercion tests that types are preserved
func TestDatastore_TypeCoercion(t *testing.T) {
	code := `store = datastore("test_types")
store.set("num", 42)
store.set("str", "hello")
store.set("bool", true)
store.set("arr", [1, 2])
store.set("obj", {x = 1})

print(type(store.get("num")))
print(type(store.get("str")))
print(type(store.get("bool")))
print(type(store.get("arr")))
print(type(store.get("obj")))
`
	testDatastore(t, code, "number\nstring\nboolean\narray\nobject\n")
}

// TestDatastoreSwap_Basic tests basic swap functionality
func TestDatastoreSwap_Basic(t *testing.T) {
	ds := GetDatastore("test_swap_basic", nil)
	defer ds.Clear()

	// Set initial array
	ds.Set("inbox", []any{1.0, 2.0, 3.0})

	// Swap with empty array
	oldVal, err := ds.Swap("inbox", []any{})
	if err != nil {
		t.Errorf("Swap failed: %v", err)
	}

	// Check old value
	oldArr, ok := oldVal.([]any)
	if !ok || len(oldArr) != 3 || oldArr[0] != 1.0 {
		t.Errorf("Expected old value [1, 2, 3], got %v", oldVal)
	}

	// Check new value
	newVal, _ := ds.Get("inbox")
	newArr, ok := newVal.([]any)
	if !ok || len(newArr) != 0 {
		t.Errorf("Expected new value [], got %v", newVal)
	}
}

// TestDatastoreSwap_NonexistentKey tests swap returns nil on missing key
func TestDatastoreSwap_NonexistentKey(t *testing.T) {
	ds := GetDatastore("test_swap_nonexist", nil)
	defer ds.Clear()

	oldVal, err := ds.Swap("missing", []any{})
	if err != nil {
		t.Errorf("Swap should not error: %v", err)
	}
	if oldVal != nil {
		t.Errorf("Swap on missing key should return nil, got %v", oldVal)
	}
}

// TestDatastoreSwap_Concurrency tests that swap is atomic with concurrent operations
func TestDatastoreSwap_Concurrency(t *testing.T) {
	ds := GetDatastore("test_swap_concurrency", nil)
	defer ds.Clear()

	ds.Set("data", []any{})

	numGoroutines := 10
	operationsPerGoroutine := 50
	wg := sync.WaitGroup{}

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < operationsPerGoroutine; i++ {
				// Swap should atomically exchange value
				_, err := ds.Swap("data", []any{float64(id*operationsPerGoroutine + i)})
				if err != nil {
					t.Errorf("Swap failed: %v", err)
				}
			}
		}(g)
	}

	wg.Wait()

	// Verify final value exists
	val, _ := ds.Get("data")
	if val == nil {
		t.Errorf("Expected value after concurrent swaps, got nil")
	}
}

// TestDatastoreShift_Basic tests basic shift (FIFO dequeue) functionality
func TestDatastoreShift_Basic(t *testing.T) {
	ds := GetDatastore("test_shift_basic", nil)
	defer ds.Clear()

	ds.Set("queue", []any{1.0, 2.0, 3.0})

	first, err := ds.Shift("queue")
	if err != nil {
		t.Errorf("Shift failed: %v", err)
	}

	if first != 1.0 {
		t.Errorf("Expected first element 1.0, got %v", first)
	}

	// Check remaining array
	remaining, _ := ds.Get("queue")
	arr, ok := remaining.([]any)
	if !ok || len(arr) != 2 || arr[0] != 2.0 {
		t.Errorf("Expected [2.0, 3.0], got %v", remaining)
	}
}

// TestDatastoreShift_EmptyArray tests shift returns nil on empty array
func TestDatastoreShift_EmptyArray(t *testing.T) {
	ds := GetDatastore("test_shift_empty", nil)
	defer ds.Clear()

	ds.Set("queue", []any{})

	result, err := ds.Shift("queue")
	if err != nil {
		t.Errorf("Shift on empty array should not error: %v", err)
	}
	if result != nil {
		t.Errorf("Shift on empty array should return nil, got %v", result)
	}
}

// TestDatastoreShift_NonexistentKey tests shift error on missing key
func TestDatastoreShift_NonexistentKey(t *testing.T) {
	ds := GetDatastore("test_shift_nonexist", nil)
	defer ds.Clear()

	_, err := ds.Shift("missing")
	if err == nil {
		t.Errorf("Shift should error on nonexistent key")
	}
}

// TestDatastoreShift_NonArrayValue tests shift error on non-array value
func TestDatastoreShift_NonArrayValue(t *testing.T) {
	ds := GetDatastore("test_shift_nonarray", nil)
	defer ds.Clear()

	ds.Set("value", "string")

	_, err := ds.Shift("value")
	if err == nil {
		t.Errorf("Shift should error on non-array value")
	}
}

// TestDatastorePop_Basic tests basic pop (LIFO) functionality
func TestDatastorePop_Basic(t *testing.T) {
	ds := GetDatastore("test_pop_basic", nil)
	defer ds.Clear()

	ds.Set("stack", []any{"a", "b", "c"})

	last, err := ds.Pop("stack")
	if err != nil {
		t.Errorf("Pop failed: %v", err)
	}

	if last != "c" {
		t.Errorf("Expected last element 'c', got %v", last)
	}

	// Check remaining array
	remaining, _ := ds.Get("stack")
	arr, ok := remaining.([]any)
	if !ok || len(arr) != 2 || arr[0] != "a" {
		t.Errorf("Expected ['a', 'b'], got %v", remaining)
	}
}

// TestDatastorePop_EmptyArray tests pop returns nil on empty array
func TestDatastorePop_EmptyArray(t *testing.T) {
	ds := GetDatastore("test_pop_empty", nil)
	defer ds.Clear()

	ds.Set("stack", []any{})

	result, err := ds.Pop("stack")
	if err != nil {
		t.Errorf("Pop on empty array should not error: %v", err)
	}
	if result != nil {
		t.Errorf("Pop on empty array should return nil, got %v", result)
	}
}

// TestDatastorePop_NonexistentKey tests pop error on missing key
func TestDatastorePop_NonexistentKey(t *testing.T) {
	ds := GetDatastore("test_pop_nonexist", nil)
	defer ds.Clear()

	_, err := ds.Pop("missing")
	if err == nil {
		t.Errorf("Pop should error on nonexistent key")
	}
}

// TestDatastorePop_NonArrayValue tests pop error on non-array value
func TestDatastorePop_NonArrayValue(t *testing.T) {
	ds := GetDatastore("test_pop_nonarray", nil)
	defer ds.Clear()

	ds.Set("value", 42.0)

	_, err := ds.Pop("value")
	if err == nil {
		t.Errorf("Pop should error on non-array value")
	}
}

// TestDatastoreUnshift_Basic tests basic unshift (LIFO push) functionality
func TestDatastoreUnshift_Basic(t *testing.T) {
	ds := GetDatastore("test_unshift_basic", nil)
	defer ds.Clear()

	ds.Set("stack", []any{2.0, 3.0})

	_, err := ds.Unshift("stack", 1.0)
	if err != nil {
		t.Errorf("Unshift failed: %v", err)
	}

	val, _ := ds.Get("stack")
	arr, ok := val.([]any)
	if !ok || len(arr) != 3 || arr[0] != 1.0 {
		t.Errorf("Expected [1.0, 2.0, 3.0], got %v", val)
	}
}

// TestDatastoreUnshift_CreatesArray tests unshift creates new array
func TestDatastoreUnshift_CreatesArray(t *testing.T) {
	ds := GetDatastore("test_unshift_new", nil)
	defer ds.Clear()

	_, err := ds.Unshift("stack", "first")
	if err != nil {
		t.Errorf("Unshift failed: %v", err)
	}

	val, _ := ds.Get("stack")
	arr, ok := val.([]any)
	if !ok || len(arr) != 1 || arr[0] != "first" {
		t.Errorf("Expected ['first'], got %v", val)
	}
}

// TestDatastoreUnshift_NonArrayValue tests unshift error on non-array value
func TestDatastoreUnshift_NonArrayValue(t *testing.T) {
	ds := GetDatastore("test_unshift_nonarray", nil)
	defer ds.Clear()

	ds.Set("value", false)

	_, err := ds.Unshift("value", "item")
	if err == nil {
		t.Errorf("Unshift should error on non-array value")
	}
}

// TestDatastoreExpire_Basic tests basic TTL expiration
func TestDatastoreExpire_Basic(t *testing.T) {
	ds := GetDatastore("test_expire_basic", nil)
	defer ds.Clear()

	ds.Set("key1", "value1")

	// Set TTL to 1 second
	err := ds.Expire("key1", 1.0)
	if err != nil {
		t.Errorf("Expire failed: %v", err)
	}

	// Key should exist immediately
	val, _ := ds.Get("key1")
	if val != "value1" {
		t.Errorf("Expected 'value1', got %v", val)
	}

	// Sleep 1.2 seconds (120% of TTL to account for timing variations)
	time.Sleep(1200 * time.Millisecond)

	// Key should be expired now
	val, _ = ds.Get("key1")
	if val != nil {
		t.Errorf("Expected nil after expiry, got %v", val)
	}
}

// TestDatastoreExpire_NonexistentKey tests expire error on missing key
func TestDatastoreExpire_NonexistentKey(t *testing.T) {
	ds := GetDatastore("test_expire_nonexist", nil)
	defer ds.Clear()

	err := ds.Expire("missing", 10.0)
	if err == nil {
		t.Errorf("Expire should error on nonexistent key")
	}
}

// TestDatastoreExpire_ReExpire tests resetting TTL timer
func TestDatastoreExpire_ReExpire(t *testing.T) {
	ds := GetDatastore("test_expire_reexpire", nil)
	defer ds.Clear()

	ds.Set("key2", "value2")

	// Set initial 3-second TTL
	ds.Expire("key2", 3.0)

	// Wait 1 second
	time.Sleep(1000 * time.Millisecond)

	// Key should still exist
	val, _ := ds.Get("key2")
	if val == nil {
		t.Errorf("Key should still exist after 1 second")
	}

	// Reset TTL to 3 seconds again
	err := ds.Expire("key2", 3.0)
	if err != nil {
		t.Errorf("Re-expire failed: %v", err)
	}

	// Wait 1 more second (total 2, but timer reset)
	time.Sleep(1000 * time.Millisecond)

	// Key should still exist (would be expired with original timer)
	val, _ = ds.Get("key2")
	if val == nil {
		t.Errorf("Key should still exist after re-expire, but got nil")
	}

	// Wait 2.5 more seconds (4.5 total, past the re-expire time)
	time.Sleep(2500 * time.Millisecond)

	// Now key should be expired
	val, _ = ds.Get("key2")
	if val != nil {
		t.Errorf("Expected nil after re-expire time, got %v", val)
	}
}

// TestDatastoreExpire_LazyDeletion tests lazy deletion on access
func TestDatastoreExpire_LazyDeletion(t *testing.T) {
	ds := GetDatastore("test_expire_lazy", nil)
	defer ds.Clear()

	ds.Set("key3", "value3")
	ds.Expire("key3", 0.5)

	// Sleep to expire the key
	time.Sleep(600 * time.Millisecond)

	// Accessing the expired key should trigger lazy deletion
	val, _ := ds.Get("key3")
	if val != nil {
		t.Errorf("Lazy deletion failed: expected nil, got %v", val)
	}

	// Verify it's actually deleted from the store
	val, _ = ds.Get("key3")
	if val != nil {
		t.Errorf("Key should remain deleted on subsequent access, got %v", val)
	}
}

// TestDatastoreExpire_ArrayOperationsAfterExpiry tests expired keys with array operations
func TestDatastoreExpire_ArrayOperationsAfterExpiry(t *testing.T) {
	ds := GetDatastore("test_expire_array", nil)
	defer ds.Clear()

	ds.Set("queue", []any{1.0, 2.0, 3.0})
	ds.Expire("queue", 0.5)

	// Sleep to expire the key
	time.Sleep(600 * time.Millisecond)

	// Array operations on expired key should behave like it doesn't exist
	_, err := ds.Shift("queue")
	if err == nil {
		t.Errorf("Shift on expired key should error")
	}
}

// TestDatastoreExpire_SweepMultipleKeys tests background sweep with multiple keys
func TestDatastoreExpire_SweepMultipleKeys(t *testing.T) {
	ds := GetDatastore("test_expire_sweep", nil)
	defer ds.Clear()

	// Set multiple keys with different TTLs
	ds.Set("key_short", "short")
	ds.Set("key_long", "long")
	ds.Set("key_medium", "medium")

	ds.Expire("key_short", 0.5)
	ds.Expire("key_long", 5.0)
	ds.Expire("key_medium", 2.0)

	// Wait for short to expire
	time.Sleep(700 * time.Millisecond)

	val, _ := ds.Get("key_short")
	if val != nil {
		t.Errorf("key_short should be expired, got %v", val)
	}

	// Medium and long should still exist
	val, _ = ds.Get("key_medium")
	if val == nil {
		t.Errorf("key_medium should still exist, got nil")
	}

	val, _ = ds.Get("key_long")
	if val == nil {
		t.Errorf("key_long should still exist, got nil")
	}

	// Wait for medium to expire
	time.Sleep(1500 * time.Millisecond)

	val, _ = ds.Get("key_medium")
	if val != nil {
		t.Errorf("key_medium should be expired, got %v", val)
	}

	// Long should still exist
	val, _ = ds.Get("key_long")
	if val == nil {
		t.Errorf("key_long should still exist, got nil")
	}
}

// TestDatastoreSetOnce_Basic tests basic set_once functionality
func TestDatastoreSetOnce_Basic(t *testing.T) {
	ds := GetDatastore("test_setonce_basic", nil)
	defer ds.Clear()

	// SetOnce on new key should succeed
	set1 := ds.SetOnce("key1", "value1")
	if !set1 {
		t.Errorf("SetOnce on new key should return true")
	}

	// Verify the value was set
	val, _ := ds.Get("key1")
	if val != "value1" {
		t.Errorf("Expected 'value1', got %v", val)
	}

	// SetOnce on existing key should fail
	set2 := ds.SetOnce("key1", "value2")
	if set2 {
		t.Errorf("SetOnce on existing key should return false")
	}

	// Verify the value wasn't changed
	val, _ = ds.Get("key1")
	if val != "value1" {
		t.Errorf("Value should remain 'value1', got %v", val)
	}
}

// TestDatastoreSetOnce_Concurrency tests SetOnce is atomic
func TestDatastoreSetOnce_Concurrency(t *testing.T) {
	ds := GetDatastore("test_setonce_concurrency", nil)
	defer ds.Clear()

	numGoroutines := 10
	successCount := 0
	var successMutex sync.Mutex

	wg := sync.WaitGroup{}
	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			if ds.SetOnce("shared_key", float64(id)) {
				successMutex.Lock()
				successCount++
				successMutex.Unlock()
			}
		}(g)
	}

	wg.Wait()

	// Only one goroutine should have succeeded
	if successCount != 1 {
		t.Errorf("Expected exactly 1 success, got %d", successCount)
	}

	// Verify exactly one value was set
	val, _ := ds.Get("shared_key")
	if val == nil {
		t.Errorf("Expected value to be set")
	}
}

// TestDatastoreExists_True tests exists returns true for existing key
func TestDatastoreExists_True(t *testing.T) {
	ds := GetDatastore("test_exists_true", nil)
	defer ds.Clear()

	ds.Set("key1", "value1")

	exists := ds.Exists("key1")
	if !exists {
		t.Errorf("Exists should return true for existing key")
	}
}

// TestDatastoreExists_False tests exists returns false for missing key
func TestDatastoreExists_False(t *testing.T) {
	ds := GetDatastore("test_exists_false", nil)
	defer ds.Clear()

	exists := ds.Exists("missing")
	if exists {
		t.Errorf("Exists should return false for missing key")
	}
}

// TestDatastoreExists_AfterDelete tests exists after key is deleted
func TestDatastoreExists_AfterDelete(t *testing.T) {
	ds := GetDatastore("test_exists_after_delete", nil)
	defer ds.Clear()

	ds.Set("key1", "value1")
	if !ds.Exists("key1") {
		t.Errorf("Key should exist after Set")
	}

	ds.Delete("key1")
	if ds.Exists("key1") {
		t.Errorf("Key should not exist after Delete")
	}
}

// TestDatastoreKeys_Empty tests Keys returns empty array for empty store
func TestDatastoreKeys_Empty(t *testing.T) {
	ds := GetDatastore("test_keys_empty", nil)
	defer ds.Clear()

	keys := ds.Keys()
	if len(keys) != 0 {
		t.Errorf("Expected empty keys, got %v", keys)
	}
}

// TestDatastoreKeys_Multiple tests Keys returns all keys
func TestDatastoreKeys_Multiple(t *testing.T) {
	ds := GetDatastore("test_keys_multiple", nil)
	defer ds.Clear()

	ds.Set("key1", "value1")
	ds.Set("key2", "value2")
	ds.Set("key3", "value3")

	keys := ds.Keys()
	if len(keys) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(keys))
	}

	// Check all keys are present (order doesn't matter)
	keySet := make(map[string]bool)
	for _, k := range keys {
		keySet[k] = true
	}

	if !keySet["key1"] || !keySet["key2"] || !keySet["key3"] {
		t.Errorf("Missing expected keys in %v", keys)
	}
}

// TestDatastoreRename_Basic tests basic rename functionality
func TestDatastoreRename_Basic(t *testing.T) {
	ds := GetDatastore("test_rename_basic", nil)
	defer ds.Clear()

	ds.Set("oldKey", "value")

	err := ds.Rename("oldKey", "newKey")
	if err != nil {
		t.Errorf("Rename failed: %v", err)
	}

	// Old key should not exist
	val, _ := ds.Get("oldKey")
	if val != nil {
		t.Errorf("Old key should not exist after rename, got %v", val)
	}

	// New key should have the value
	val, _ = ds.Get("newKey")
	if val != "value" {
		t.Errorf("New key should have the value, got %v", val)
	}
}

// TestDatastoreRename_NonexistentOldKey tests rename error on missing old key
func TestDatastoreRename_NonexistentOldKey(t *testing.T) {
	ds := GetDatastore("test_rename_nonexist_old", nil)
	defer ds.Clear()

	err := ds.Rename("missing", "newKey")
	if err == nil {
		t.Errorf("Rename should error on nonexistent old key")
	}
}

// TestDatastoreRename_ExistingNewKey tests rename error when new key exists
func TestDatastoreRename_ExistingNewKey(t *testing.T) {
	ds := GetDatastore("test_rename_exist_new", nil)
	defer ds.Clear()

	ds.Set("oldKey", "value1")
	ds.Set("newKey", "value2")

	err := ds.Rename("oldKey", "newKey")
	if err == nil {
		t.Errorf("Rename should error when new key already exists")
	}

	// Both keys should still exist with original values
	val, _ := ds.Get("oldKey")
	if val != "value1" {
		t.Errorf("Old key should still exist with original value, got %v", val)
	}

	val, _ = ds.Get("newKey")
	if val != "value2" {
		t.Errorf("New key should still exist with original value, got %v", val)
	}
}
