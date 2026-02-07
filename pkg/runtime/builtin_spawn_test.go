package runtime

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/duso-org/duso/pkg/script"
)

// TestSpawn_BasicExecution tests spawning a script that modifies datastore
func TestSpawn_BasicExecution(t *testing.T) {
	// Create a temp script file
	scriptContent := `
store = datastore("test_spawn_basic")
store.set("executed", true)
`
	tmpDir := t.TempDir()
	scriptPath := tmpDir + "/worker.du"
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0644); err != nil {
		t.Fatalf("Failed to create temp script: %v", err)
	}

	// Create interpreter with script loader
	interp := &script.Interpreter{
		ScriptLoader: func(path string) ([]byte, error) {
			return os.ReadFile(path)
		},
	}

	fn := NewSpawnFunction(interp)
	args := map[string]any{"0": scriptPath}

	pid, err := fn(nil, args)
	if err != nil {
		t.Fatalf("spawn() failed: %v", err)
	}

	// Verify spawn returned a PID
	if pidVal, ok := pid.(float64); !ok || pidVal == 0 {
		t.Errorf("Expected positive PID, got %v", pid)
	}

	// Give spawned goroutine time to execute
	time.Sleep(100 * time.Millisecond)

	// Verify script executed by checking datastore
	ds := script.GetDatastore("test_spawn_basic", nil)
	executed, _ := ds.Get("executed")
	if executed == nil {
		t.Errorf("Expected executed=true in datastore, got nil (script didn't run)")
	}
	defer ds.Clear()
}

// TestSpawn_MultipleSpawns tests spawning multiple scripts concurrently
func TestSpawn_MultipleSpawns(t *testing.T) {
	scriptTemplate := `
store = datastore("test_spawn_multi")
id = %d
store.increment("count", 1)
`
	tmpDir := t.TempDir()
	ds := script.GetDatastore("test_spawn_multi", nil)
	defer ds.Clear()

	// Create interpreter with script loader
	interp := &script.Interpreter{
		ScriptLoader: func(path string) ([]byte, error) {
			return os.ReadFile(path)
		},
	}

	// Spawn 5 scripts
	for i := 0; i < 5; i++ {
		scriptPath := fmt.Sprintf("%s/worker_%d.du", tmpDir, i)
		scriptContent := fmt.Sprintf(scriptTemplate, i)
		if err := os.WriteFile(scriptPath, []byte(scriptContent), 0644); err != nil {
			t.Fatalf("Failed to create temp script: %v", err)
		}

		fn := NewSpawnFunction(interp)
		args := map[string]any{"0": scriptPath}

		_, err := fn(nil, args)
		if err != nil {
			t.Fatalf("spawn() failed: %v", err)
		}
	}

	// Wait for all spawned scripts to execute
	time.Sleep(300 * time.Millisecond)

	// Verify all scripts executed
	count, _ := ds.Get("count")
	if countVal, ok := count.(float64); !ok || countVal < 5 {
		t.Errorf("Expected count >= 5, got %v", count)
	}
}

// TestSpawn_ReturnValue tests that spawn() returns immediately
func TestSpawn_ReturnValue(t *testing.T) {
	scriptContent := `
store = datastore("test_spawn_return")
store.set("key", "value")
`
	tmpDir := t.TempDir()
	scriptPath := tmpDir + "/worker.du"
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0644); err != nil {
		t.Fatalf("Failed to create temp script: %v", err)
	}

	interp := &script.Interpreter{
		ScriptLoader: func(path string) ([]byte, error) {
			return os.ReadFile(path)
		},
	}

	fn := NewSpawnFunction(interp)
	args := map[string]any{"0": scriptPath}

	startTime := time.Now()
	pid, err := fn(nil, args)
	elapsed := time.Since(startTime)

	if err != nil {
		t.Fatalf("spawn() failed: %v", err)
	}

	// spawn() should return almost immediately (not wait for script)
	if elapsed > 100*time.Millisecond {
		t.Errorf("spawn() took too long (should be async): %v", elapsed)
	}

	// Verify it returned a PID
	if pidVal, ok := pid.(float64); !ok || pidVal == 0 {
		t.Errorf("Expected positive PID, got %v", pid)
	}
}


// TestSpawn_DatastoreSharing tests that parent and child share datastore namespace
func TestSpawn_DatastoreSharing(t *testing.T) {
	scriptContent := `
store = datastore("test_spawn_share")
store.set("child_key", "child_value")
`
	tmpDir := t.TempDir()
	scriptPath := tmpDir + "/worker.du"
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0644); err != nil {
		t.Fatalf("Failed to create temp script: %v", err)
	}

	interp := &script.Interpreter{
		ScriptLoader: func(path string) ([]byte, error) {
			return os.ReadFile(path)
		},
	}

	fn := NewSpawnFunction(interp)
	args := map[string]any{"0": scriptPath}

	_, err := fn(nil, args)
	if err != nil {
		t.Fatalf("spawn() failed: %v", err)
	}

	// Wait for script to execute
	time.Sleep(100 * time.Millisecond)

	// Verify parent and child share the same datastore
	ds := script.GetDatastore("test_spawn_share", nil)
	defer ds.Clear()

	childVal, _ := ds.Get("child_key")
	if childVal == nil {
		t.Errorf("Expected child_key in parent's datastore, got nil (datastores not shared)")
	}
}

// TestSpawn_ScriptError tests that script errors don't crash spawn
func TestSpawn_ScriptError(t *testing.T) {
	scriptContent := `
x = 1 / 0
`
	tmpDir := t.TempDir()
	scriptPath := tmpDir + "/error_worker.du"
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0644); err != nil {
		t.Fatalf("Failed to create temp script: %v", err)
	}

	interp := &script.Interpreter{
		ScriptLoader: func(path string) ([]byte, error) {
			return os.ReadFile(path)
		},
	}

	fn := NewSpawnFunction(interp)
	args := map[string]any{"0": scriptPath}

	// spawn() should succeed even though the script will error
	pid, err := fn(nil, args)
	if err != nil {
		t.Fatalf("spawn() failed: %v", err)
	}

	// Verify it returned a PID
	if pidVal, ok := pid.(float64); !ok || pidVal == 0 {
		t.Errorf("Expected positive PID, got %v", pid)
	}

	// Wait for script to execute and error
	time.Sleep(100 * time.Millisecond)
	// No panic should occur
}

// TestSpawn_MissingScript tests error on nonexistent script
func TestSpawn_MissingScript(t *testing.T) {
	interp := &script.Interpreter{
		ScriptLoader: func(path string) ([]byte, error) {
			return os.ReadFile(path)
		},
	}

	fn := NewSpawnFunction(interp)
	args := map[string]any{"0": "/nonexistent/path/worker.du"}

	_, err := fn(nil, args)
	if err == nil {
		t.Errorf("Expected error for missing script, got nil")
	}
}

// TestSpawn_InvalidArguments tests error on wrong argument types
func TestSpawn_InvalidArguments(t *testing.T) {
	interp := &script.Interpreter{}

	fn := NewSpawnFunction(interp)

	tests := []struct {
		name string
		args map[string]any
	}{
		{"no_args", map[string]any{}},
		{"non_string_path", map[string]any{"0": 123}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := fn(nil, tc.args)
			if err == nil {
				t.Errorf("Expected error for %s, got nil", tc.name)
			}
		})
	}
}

// TestSpawn_SyntaxErrorInScript tests error on script with syntax errors
func TestSpawn_SyntaxErrorInScript(t *testing.T) {
	scriptContent := `
if x then
  print("missing end")
`
	tmpDir := t.TempDir()
	scriptPath := tmpDir + "/syntax_error.du"
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0644); err != nil {
		t.Fatalf("Failed to create temp script: %v", err)
	}

	interp := &script.Interpreter{
		ScriptLoader: func(path string) ([]byte, error) {
			return os.ReadFile(path)
		},
	}

	fn := NewSpawnFunction(interp)
	args := map[string]any{"0": scriptPath}

	_, err := fn(nil, args)
	if err == nil {
		t.Errorf("Expected error for syntax error in script, got nil")
	}
}

// TestSpawn_GoroutineCleanup tests that spawned goroutines clean up properly
func TestSpawn_GoroutineCleanup(t *testing.T) {
	scriptContent := `
store = datastore("test_spawn_cleanup")
store.set("done", true)
`
	tmpDir := t.TempDir()
	scriptPath := tmpDir + "/worker.du"
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0644); err != nil {
		t.Fatalf("Failed to create temp script: %v", err)
	}

	interp := &script.Interpreter{
		ScriptLoader: func(path string) ([]byte, error) {
			return os.ReadFile(path)
		},
	}

	initialGoroutines := runtime.NumGoroutine()

	fn := NewSpawnFunction(interp)
	args := map[string]any{"0": scriptPath}

	_, err := fn(nil, args)
	if err != nil {
		t.Fatalf("spawn() failed: %v", err)
	}

	// Wait for script to complete
	time.Sleep(200 * time.Millisecond)

	// Allow a small window for goroutine cleanup
	time.Sleep(50 * time.Millisecond)

	finalGoroutines := runtime.NumGoroutine()

	// Goroutine count should be same (spawned goroutine should have exited)
	// Allow ±2 for any background activity
	if finalGoroutines > initialGoroutines+2 {
		t.Errorf("Goroutine leak detected: started with %d, ended with %d",
			initialGoroutines, finalGoroutines)
	}
}

// TestSpawn_NoMemoryLeak tests spawning many scripts doesn't leak memory
func TestSpawn_NoMemoryLeak(t *testing.T) {
	scriptTemplate := `
store = datastore("test_spawn_leak")
store.increment("count", 1)
`
	tmpDir := t.TempDir()
	scriptPath := tmpDir + "/worker.du"
	if err := os.WriteFile(scriptPath, []byte(scriptTemplate), 0644); err != nil {
		t.Fatalf("Failed to create temp script: %v", err)
	}

	interp := &script.Interpreter{
		ScriptLoader: func(path string) ([]byte, error) {
			return os.ReadFile(path)
		},
	}

	initialGoroutines := runtime.NumGoroutine()
	var wg sync.WaitGroup

	// Spawn 100 scripts
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fn := NewSpawnFunction(interp)
			args := map[string]any{"0": scriptPath}
			fn(nil, args) // Ignore errors for this stress test
		}()
	}

	wg.Wait()

	// Wait for all spawned scripts to complete
	time.Sleep(500 * time.Millisecond)

	finalGoroutines := runtime.NumGoroutine()

	// Should not have significant goroutine leak
	// Allow up to 10 additional goroutines for cleanup buffers
	if finalGoroutines > initialGoroutines+10 {
		t.Errorf("Potential memory leak: started with %d goroutines, ended with %d",
			initialGoroutines, finalGoroutines)
	}
}
