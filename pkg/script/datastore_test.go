package script

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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

// TestDatastore_Append tests atomic array append operations
func TestDatastore_Append(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{
			"append to nonexistent key creates array",
			`store = datastore("test_append_new")
len = store.append("items", "first")
print(len)
print(store.get("items"))`,
			"1\n[first]\n",
		},
		{
			"append multiple items",
			`store = datastore("test_append_multi")
store.append("items", "a")
store.append("items", "b")
store.append("items", "c")
print(store.get("items"))`,
			"[a, b, c]\n",
		},
		{
			"append returns new length",
			`store = datastore("test_append_len")
len1 = store.append("items", 1)
len2 = store.append("items", 2)
len3 = store.append("items", 3)
print(len1)
print(len2)
print(len3)`,
			"1\n2\n3\n",
		},
		{
			"append various types",
			`store = datastore("test_append_types")
store.append("mixed", 42)
store.append("mixed", "text")
store.append("mixed", true)
store.append("mixed", {key = "value"})
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

// TestDatastore_ErrorOnAppendNonArray tests error when appending to non-array
func TestDatastore_ErrorOnAppendNonArray(t *testing.T) {
	code := `store = datastore("test_append_error")
store.set("scalar", 42)
store.append("scalar", "item")
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

	// Append to array
	store.Append("items", "first")

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

// TestDatastore_AppendReturnsLength tests append returns array length
func TestDatastore_AppendReturnsLength(t *testing.T) {
	code := `store = datastore("test_append_return")
len1 = store.append("arr", "x")
len2 = store.append("arr", "y")
len3 = store.append("arr", "z")
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
