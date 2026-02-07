package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/duso-org/duso/pkg/script"
)

// MockModuleResolver implements ModuleResolver for testing
type MockModuleResolver struct {
	modules map[string]string // filename -> file path
}

func (m *MockModuleResolver) ResolveModule(moduleName string) (string, []string, error) {
	if path, ok := m.modules[moduleName]; ok {
		return path, nil, nil
	}
	return "", []string{}, fmt.Errorf("module not found: %s", moduleName)
}

// MockCircularDetector implements CircularDetector for testing
type MockCircularDetector struct {
	stack map[string]bool // Tracks modules in current require stack
}

func NewMockCircularDetector() *MockCircularDetector {
	return &MockCircularDetector{
		stack: make(map[string]bool),
	}
}

func (d *MockCircularDetector) Push(path string) error {
	if d.stack[path] {
		return fmt.Errorf("circular dependency detected: %s", path)
	}
	d.stack[path] = true
	return nil
}

func (d *MockCircularDetector) Pop() {
	// In real implementation, we'd track stack properly
	// For testing, this is good enough
}

// TestRequire_BasicModule tests loading a basic module that exports a value
func TestRequire_BasicModule(t *testing.T) {
	tmpDir := t.TempDir()
	modulePath := filepath.Join(tmpDir, "math.du")

	// Create a simple module that exports a number
	moduleCode := `42`
	if err := os.WriteFile(modulePath, []byte(moduleCode), 0644); err != nil {
		t.Fatalf("Failed to create module: %v", err)
	}

	interp := script.NewInterpreter(false)
	interp.ScriptLoader = func(path string) ([]byte, error) {
		return os.ReadFile(path)
	}

	resolver := &MockModuleResolver{
		modules: map[string]string{
			"math": modulePath,
		},
	}
	detector := NewMockCircularDetector()

	fn := NewRequireFunction(resolver, detector, interp)
	args := map[string]any{"0": "math"}

	result, err := fn(nil, args)
	if err != nil {
		t.Fatalf("require() failed: %v", err)
	}

	// Module should export 42
	if resultVal, ok := result.(float64); !ok || resultVal != 42 {
		t.Errorf("Expected 42, got %v (type %T)", result, result)
	}
}

// TestRequire_ModuleReturnsObject tests module that exports an object with functions
func TestRequire_ModuleReturnsObject(t *testing.T) {
	tmpDir := t.TempDir()
	modulePath := filepath.Join(tmpDir, "helpers.du")

	// Create a module that exports an object
	moduleCode := `{
  greet = function(name)
    return "Hello, " + name
  end,
  multiply = function(a, b)
    return a * b
  end
}`
	if err := os.WriteFile(modulePath, []byte(moduleCode), 0644); err != nil {
		t.Fatalf("Failed to create module: %v", err)
	}

	interp := script.NewInterpreter(false)
	interp.ScriptLoader = func(path string) ([]byte, error) {
		return os.ReadFile(path)
	}

	resolver := &MockModuleResolver{
		modules: map[string]string{
			"helpers": modulePath,
		},
	}
	detector := NewMockCircularDetector()

	fn := NewRequireFunction(resolver, detector, interp)
	args := map[string]any{"0": "helpers"}

	result, err := fn(nil, args)
	if err != nil {
		t.Fatalf("require() failed: %v", err)
	}

	// Module should return an object with functions
	obj, ok := result.(map[string]any)
	if !ok {
		t.Errorf("Expected object, got %T", result)
		return
	}

	if _, hasGreet := obj["greet"]; !hasGreet {
		t.Errorf("Module should export 'greet' function")
	}

	if _, hasMultiply := obj["multiply"]; !hasMultiply {
		t.Errorf("Module should export 'multiply' function")
	}
}

// TestRequire_CacheHit tests that modules are cached (only executed once)
func TestRequire_CacheHit(t *testing.T) {
	tmpDir := t.TempDir()
	modulePath := filepath.Join(tmpDir, "counter.du")

	// Create module with side effect: set a value in datastore
	moduleCode := `
store = datastore("test_require_cache")
store.set("loaded", true)
{value = 42}
`
	if err := os.WriteFile(modulePath, []byte(moduleCode), 0644); err != nil {
		t.Fatalf("Failed to create module: %v", err)
	}

	interp := script.NewInterpreter(false)
	interp.ScriptLoader = func(path string) ([]byte, error) {
		return os.ReadFile(path)
	}

	resolver := &MockModuleResolver{
		modules: map[string]string{
			"counter": modulePath,
		},
	}
	detector := NewMockCircularDetector()

	fn := NewRequireFunction(resolver, detector, interp)
	args := map[string]any{"0": "counter"}

	// First require
	result1, err := fn(nil, args)
	if err != nil {
		t.Fatalf("First require() failed: %v", err)
	}

	// Get datastore state - should have loaded=true
	ds := script.GetDatastore("test_require_cache", nil)
	loaded1, _ := ds.Get("loaded")

	// Second require - should return cached value
	result2, err := fn(nil, args)
	if err != nil {
		t.Fatalf("Second require() failed: %v", err)
	}

	// Both should be identical objects
	if fmt.Sprintf("%v", result1) != fmt.Sprintf("%v", result2) {
		t.Errorf("Cache hit: results should be identical")
	}

	// Verify module executed (loaded value exists)
	if loaded1 == nil {
		t.Errorf("Expected module to execute and set 'loaded' value")
	}

	ds.Clear()
}

// TestRequire_RelativePath tests require() with relative path
func TestRequire_RelativePath(t *testing.T) {
	tmpDir := t.TempDir()
	modulePath := filepath.Join(tmpDir, "relative_module.du")

	moduleCode := `"from relative path"`
	if err := os.WriteFile(modulePath, []byte(moduleCode), 0644); err != nil {
		t.Fatalf("Failed to create module: %v", err)
	}

	interp := script.NewInterpreter(false)
	interp.ScriptLoader = func(path string) ([]byte, error) {
		return os.ReadFile(path)
	}

	resolver := &MockModuleResolver{
		modules: map[string]string{
			"./relative_module": modulePath,
		},
	}
	detector := NewMockCircularDetector()

	fn := NewRequireFunction(resolver, detector, interp)
	args := map[string]any{"0": "./relative_module"}

	result, err := fn(nil, args)
	if err != nil {
		t.Fatalf("require() with relative path failed: %v", err)
	}

	if result != "from relative path" {
		t.Errorf("Expected 'from relative path', got %v", result)
	}
}

// TestRequire_AbsolutePath tests require() with absolute path
func TestRequire_AbsolutePath(t *testing.T) {
	tmpDir := t.TempDir()
	modulePath := filepath.Join(tmpDir, "absolute_module.du")

	moduleCode := `"absolute path content"`
	if err := os.WriteFile(modulePath, []byte(moduleCode), 0644); err != nil {
		t.Fatalf("Failed to create module: %v", err)
	}

	interp := script.NewInterpreter(false)
	interp.ScriptLoader = func(path string) ([]byte, error) {
		return os.ReadFile(path)
	}

	resolver := &MockModuleResolver{
		modules: map[string]string{
			modulePath: modulePath,
		},
	}
	detector := NewMockCircularDetector()

	fn := NewRequireFunction(resolver, detector, interp)
	args := map[string]any{"0": modulePath}

	result, err := fn(nil, args)
	if err != nil {
		t.Fatalf("require() with absolute path failed: %v", err)
	}

	if result != "absolute path content" {
		t.Errorf("Expected 'absolute path content', got %v", result)
	}
}

// TestRequire_NonexistentModule tests error on missing module
func TestRequire_NonexistentModule(t *testing.T) {
	interp := script.NewInterpreter(false)
	interp.ScriptLoader = func(path string) ([]byte, error) {
		return os.ReadFile(path)
	}

	resolver := &MockModuleResolver{
		modules: map[string]string{},
	}
	detector := NewMockCircularDetector()

	fn := NewRequireFunction(resolver, detector, interp)
	args := map[string]any{"0": "nonexistent"}

	_, err := fn(nil, args)
	if err == nil {
		t.Errorf("Expected error for nonexistent module, got nil")
	}
}

// TestRequire_SyntaxErrorInModule tests error on module with syntax error
func TestRequire_SyntaxErrorInModule(t *testing.T) {
	tmpDir := t.TempDir()
	modulePath := filepath.Join(tmpDir, "syntax_error.du")

	// Create module with syntax error
	moduleCode := `if x then
  print("missing end")`
	if err := os.WriteFile(modulePath, []byte(moduleCode), 0644); err != nil {
		t.Fatalf("Failed to create module: %v", err)
	}

	interp := script.NewInterpreter(false)
	interp.ScriptLoader = func(path string) ([]byte, error) {
		return os.ReadFile(path)
	}

	resolver := &MockModuleResolver{
		modules: map[string]string{
			"syntax": modulePath,
		},
	}
	detector := NewMockCircularDetector()

	fn := NewRequireFunction(resolver, detector, interp)
	args := map[string]any{"0": "syntax"}

	_, err := fn(nil, args)
	if err == nil {
		t.Errorf("Expected error for module with syntax error, got nil")
	}
}

// TestRequire_RuntimeErrorInModule tests error on module with runtime error
func TestRequire_RuntimeErrorInModule(t *testing.T) {
	tmpDir := t.TempDir()
	modulePath := filepath.Join(tmpDir, "runtime_error.du")

	// Create module that throws runtime error
	moduleCode := `x = 1 / 0`
	if err := os.WriteFile(modulePath, []byte(moduleCode), 0644); err != nil {
		t.Fatalf("Failed to create module: %v", err)
	}

	interp := script.NewInterpreter(false)
	interp.ScriptLoader = func(path string) ([]byte, error) {
		return os.ReadFile(path)
	}

	resolver := &MockModuleResolver{
		modules: map[string]string{
			"error": modulePath,
		},
	}
	detector := NewMockCircularDetector()

	fn := NewRequireFunction(resolver, detector, interp)
	args := map[string]any{"0": "error"}

	_, err := fn(nil, args)
	if err == nil {
		t.Errorf("Expected error for module with runtime error, got nil")
	}
}

// TestRequire_CircularDependency tests error on circular dependencies (A -> B -> A)
func TestRequire_CircularDependency(t *testing.T) {
	tmpDir := t.TempDir()

	// Create two modules that require each other
	moduleAPath := filepath.Join(tmpDir, "module_a.du")
	moduleBPath := filepath.Join(tmpDir, "module_b.du")

	moduleACode := `b = require("module_b")\n{a = 1}`
	moduleBCode := `a = require("module_a")\n{b = 2}`

	if err := os.WriteFile(moduleAPath, []byte(moduleACode), 0644); err != nil {
		t.Fatalf("Failed to create module A: %v", err)
	}
	if err := os.WriteFile(moduleBPath, []byte(moduleBCode), 0644); err != nil {
		t.Fatalf("Failed to create module B: %v", err)
	}

	interp := script.NewInterpreter(false)
	interp.ScriptLoader = func(path string) ([]byte, error) {
		return os.ReadFile(path)
	}

	// Note: The detector needs proper implementation to catch circular deps
	// For now, test that circular detection mechanism is called
	resolver := &MockModuleResolver{
		modules: map[string]string{
			"module_a": moduleAPath,
			"module_b": moduleBPath,
		},
	}
	detector := NewMockCircularDetector()

	fn := NewRequireFunction(resolver, detector, interp)
	args := map[string]any{"0": "module_a"}

	_, err := fn(nil, args)
	// Should detect circular dependency
	if err == nil {
		t.Errorf("Expected error for circular dependency, got nil")
	}
}

// TestRequire_NoArguments tests error on missing arguments
func TestRequire_NoArguments(t *testing.T) {
	interp := script.NewInterpreter(false)
	interp.ScriptLoader = func(path string) ([]byte, error) {
		return os.ReadFile(path)
	}

	resolver := &MockModuleResolver{}
	detector := NewMockCircularDetector()

	fn := NewRequireFunction(resolver, detector, interp)
	args := map[string]any{} // No arguments

	_, err := fn(nil, args)
	if err == nil {
		t.Errorf("Expected error for missing arguments, got nil")
	}
}

// TestRequire_NamedArgument tests require() with named 'filename' argument
func TestRequire_NamedArgument(t *testing.T) {
	tmpDir := t.TempDir()
	modulePath := filepath.Join(tmpDir, "named.du")

	moduleCode := `"named argument module"`
	if err := os.WriteFile(modulePath, []byte(moduleCode), 0644); err != nil {
		t.Fatalf("Failed to create module: %v", err)
	}

	interp := script.NewInterpreter(false)
	interp.ScriptLoader = func(path string) ([]byte, error) {
		return os.ReadFile(path)
	}

	resolver := &MockModuleResolver{
		modules: map[string]string{
			"named": modulePath,
		},
	}
	detector := NewMockCircularDetector()

	fn := NewRequireFunction(resolver, detector, interp)
	args := map[string]any{"filename": "named"} // Named argument

	result, err := fn(nil, args)
	if err != nil {
		t.Fatalf("require() with named argument failed: %v", err)
	}

	if result != "named argument module" {
		t.Errorf("Expected 'named argument module', got %v", result)
	}
}

// TestRequire_MultipleModules tests loading multiple different modules
func TestRequire_MultipleModules(t *testing.T) {
	tmpDir := t.TempDir()
	module1Path := filepath.Join(tmpDir, "mod1.du")
	module2Path := filepath.Join(tmpDir, "mod2.du")
	module3Path := filepath.Join(tmpDir, "mod3.du")

	if err := os.WriteFile(module1Path, []byte(`1`), 0644); err != nil {
		t.Fatalf("Failed to create module 1: %v", err)
	}
	if err := os.WriteFile(module2Path, []byte(`"two"`), 0644); err != nil {
		t.Fatalf("Failed to create module 2: %v", err)
	}
	if err := os.WriteFile(module3Path, []byte(`{three = 3}`), 0644); err != nil {
		t.Fatalf("Failed to create module 3: %v", err)
	}

	interp := script.NewInterpreter(false)
	interp.ScriptLoader = func(path string) ([]byte, error) {
		return os.ReadFile(path)
	}

	resolver := &MockModuleResolver{
		modules: map[string]string{
			"mod1": module1Path,
			"mod2": module2Path,
			"mod3": module3Path,
		},
	}
	detector := NewMockCircularDetector()

	fn := NewRequireFunction(resolver, detector, interp)

	// Load module 1
	result1, err := fn(nil, map[string]any{"0": "mod1"})
	if err != nil {
		t.Fatalf("require(mod1) failed: %v", err)
	}
	if result1 != float64(1) {
		t.Errorf("Expected 1, got %v", result1)
	}

	// Load module 2
	result2, err := fn(nil, map[string]any{"0": "mod2"})
	if err != nil {
		t.Fatalf("require(mod2) failed: %v", err)
	}
	if result2 != "two" {
		t.Errorf("Expected 'two', got %v", result2)
	}

	// Load module 3
	result3, err := fn(nil, map[string]any{"0": "mod3"})
	if err != nil {
		t.Fatalf("require(mod3) failed: %v", err)
	}
	obj, ok := result3.(map[string]any)
	if !ok {
		t.Errorf("Expected object, got %T", result3)
	} else if val, hasThree := obj["three"]; !hasThree || val != float64(3) {
		t.Errorf("Expected {three = 3}, got %v", result3)
	}
}

// TestRequire_ScriptLoaderError tests error when ScriptLoader is missing
func TestRequire_ScriptLoaderError(t *testing.T) {
	interp := script.NewInterpreter(false)
	interp.ScriptLoader = nil // No ScriptLoader capability

	resolver := &MockModuleResolver{
		modules: map[string]string{
			"test": "/tmp/test.du",
		},
	}
	detector := NewMockCircularDetector()

	fn := NewRequireFunction(resolver, detector, interp)
	args := map[string]any{"0": "test"}

	_, err := fn(nil, args)
	if err == nil {
		t.Errorf("Expected error when ScriptLoader is missing, got nil")
	}
}
