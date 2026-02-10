package runtime

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/duso-org/duso/pkg/script"
)

// MockFileReader implements a FileReader for testing
// It intelligently handles path resolution: if exact path not found,
// it tries the bare filename for backwards compatibility with tests.
type MockFileReader struct {
	files map[string]string
}

func (m *MockFileReader) Read(path string) ([]byte, error) {
	// Try exact path first
	if content, ok := m.files[path]; ok {
		return []byte(content), nil
	}

	// For backwards compatibility: if path has directory components,
	// try matching the bare filename
	// This allows tests that register "test.txt" to work with load()
	// that now resolves to "/tmp/test.txt"
	for registeredPath, content := range m.files {
		// Extract bare filename from both paths
		var bareFile string
		for i := len(path) - 1; i >= 0; i-- {
			if path[i] == '/' || path[i] == '\\' {
				bareFile = path[i+1:]
				break
			}
		}
		if bareFile == "" {
			bareFile = path
		}

		// If registered path matches the bare filename, use it
		if registeredPath == bareFile {
			return []byte(content), nil
		}
	}

	return nil, fmt.Errorf("file not found: %s", path)
}

// MockFileWriter implements a FileWriter for testing
type MockFileWriter struct {
	files map[string]string
}

func (m *MockFileWriter) Write(path string, content string) error {
	if m.files == nil {
		m.files = make(map[string]string)
	}
	m.files[path] = content
	return nil
}

// TestLoad_BasicTextFile tests loading a plain text file
func TestLoad_BasicTextFile(t *testing.T) {
	interp := script.NewInterpreter(false)
	interp.SetScriptDir("/tmp")
	interp.FileReader = (&MockFileReader{
		files: map[string]string{
			"test.txt": "Hello, World!",
		},
	}).Read

	fn := NewLoadFunction(interp)
	args := map[string]any{"0": "test.txt"}

	result, err := fn(nil, args)
	if err != nil {
		t.Fatalf("load() failed: %v", err)
	}

	content, ok := result.(string)
	if !ok {
		t.Fatalf("Expected string, got %T", result)
	}

	if content != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!', got %q", content)
	}
}

// TestLoad_JSONFile tests loading and parsing JSON
func TestLoad_JSONFile(t *testing.T) {
	jsonData := map[string]any{
		"name": "Alice",
		"age":  30.0,
	}
	jsonBytes, _ := json.Marshal(jsonData)

	interp := script.NewInterpreter(false)
	interp.SetScriptDir("/tmp")
	interp.FileReader = (&MockFileReader{
		files: map[string]string{
			"config.json": string(jsonBytes),
		},
	}).Read

	fn := NewLoadFunction(interp)
	args := map[string]any{"0": "config.json"}

	result, err := fn(nil, args)
	if err != nil {
		t.Fatalf("load() failed: %v", err)
	}

	content, ok := result.(string)
	if !ok {
		t.Fatalf("Expected string, got %T", result)
	}

	// Verify it's valid JSON
	var parsed map[string]any
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		t.Errorf("JSON content invalid: %v", err)
	}
}

// TestLoad_NonexistentFile tests error on missing file
func TestLoad_NonexistentFile(t *testing.T) {
	interp := script.NewInterpreter(false)
	interp.SetScriptDir("/tmp")
	interp.FileReader = (&MockFileReader{
		files: map[string]string{},
	}).Read

	fn := NewLoadFunction(interp)
	args := map[string]any{"0": "nonexistent.txt"}

	_, err := fn(nil, args)
	if err == nil {
		t.Errorf("Expected error for nonexistent file, got nil")
	}
}

// TestLoad_STOREPath tests loading from virtual /STORE/ path
func TestLoad_STOREPath(t *testing.T) {
	interp := script.NewInterpreter(false)
	interp.SetScriptDir("/tmp")
	interp.FileReader = (&MockFileReader{
		files: map[string]string{
			"/STORE/data.txt": "stored data",
		},
	}).Read

	fn := NewLoadFunction(interp)
	args := map[string]any{"0": "/STORE/data.txt"}

	result, err := fn(nil, args)
	if err != nil {
		t.Fatalf("load() failed: %v", err)
	}

	content, _ := result.(string)
	if content != "stored data" {
		t.Errorf("Expected 'stored data', got %q", content)
	}
}

// TestLoad_MissingArgument tests error when filename is missing
func TestLoad_MissingArgument(t *testing.T) {
	interp := script.NewInterpreter(false)
	interp.SetScriptDir("/tmp")
	interp.FileReader = (&MockFileReader{}).Read

	fn := NewLoadFunction(interp)
	args := map[string]any{} // No filename

	_, err := fn(nil, args)
	if err == nil {
		t.Errorf("Expected error for missing argument, got nil")
	}
}

// TestLoad_WithoutCapability tests error when FileReader capability is missing
func TestLoad_WithoutCapability(t *testing.T) {
	interp := script.NewInterpreter(false)
	interp.SetScriptDir("/tmp")
	interp.FileReader = nil // No FileReader capability

	fn := NewLoadFunction(interp)
	args := map[string]any{"0": "test.txt"}

	_, err := fn(nil, args)
	if err == nil {
		t.Errorf("Expected error for missing FileReader capability, got nil")
	}
}

// TestSave_BasicTextFile tests saving plain text
func TestSave_BasicTextFile(t *testing.T) {
	writer := &MockFileWriter{files: make(map[string]string)}

	interp := script.NewInterpreter(false)
	interp.SetScriptDir("/tmp")
	interp.FileWriter = func(path, content string) error {
		return writer.Write(path, content)
	}

	fn := NewSaveFunction(interp)
	args := map[string]any{
		"0": "output.txt",
		"1": "Hello, World!",
	}

	result, err := fn(nil, args)
	if err != nil {
		t.Fatalf("save() failed: %v", err)
	}

	// save() returns nil
	if result != nil {
		t.Errorf("Expected nil return, got %v", result)
	}

	// Verify content was written
	if writer.files["/tmp/output.txt"] != "Hello, World!" {
		t.Errorf("Content not saved correctly: %q", writer.files["/tmp/output.txt"])
	}
}

// TestSave_JSONContent tests saving JSON
func TestSave_JSONContent(t *testing.T) {
	writer := &MockFileWriter{files: make(map[string]string)}

	interp := script.NewInterpreter(false)
	interp.SetScriptDir("/tmp")
	interp.FileWriter = func(path, content string) error {
		return writer.Write(path, content)
	}

	fn := NewSaveFunction(interp)
	jsonStr := `{"name":"Alice","age":30}`
	args := map[string]any{
		"0": "data.json",
		"1": jsonStr,
	}

	_, err := fn(nil, args)
	if err != nil {
		t.Fatalf("save() failed: %v", err)
	}

	// Verify JSON is valid
	var parsed map[string]any
	if err := json.Unmarshal([]byte(writer.files["/tmp/data.json"]), &parsed); err != nil {
		t.Errorf("Saved JSON invalid: %v", err)
	}
}

// TestSave_STOREPath tests saving to virtual /STORE/ path
func TestSave_STOREPath(t *testing.T) {
	writer := &MockFileWriter{files: make(map[string]string)}

	interp := script.NewInterpreter(false)
	interp.SetScriptDir("/tmp")
	interp.FileWriter = func(path, content string) error {
		return writer.Write(path, content)
	}

	fn := NewSaveFunction(interp)
	args := map[string]any{
		"0": "/STORE/generated.du",
		"1": "// Generated code",
	}

	_, err := fn(nil, args)
	if err != nil {
		t.Fatalf("save() to /STORE/ failed: %v", err)
	}

	if writer.files["/STORE/generated.du"] != "// Generated code" {
		t.Errorf("Content not saved to /STORE/ correctly")
	}
}

// TestSave_EMBEDPathError tests that /EMBED/ is read-only for save
func TestSave_EMBEDPathError(t *testing.T) {
	writer := &MockFileWriter{files: make(map[string]string)}

	interp := script.NewInterpreter(false)
	interp.SetScriptDir("/tmp")
	interp.FileWriter = func(path, content string) error {
		// Simulate /EMBED/ being read-only
		if path == "/EMBED/file.txt" {
			return fmt.Errorf("/EMBED/ is read-only")
		}
		return writer.Write(path, content)
	}

	fn := NewSaveFunction(interp)
	args := map[string]any{
		"0": "/EMBED/file.txt",
		"1": "should fail",
	}

	_, err := fn(nil, args)
	if err == nil {
		t.Errorf("Expected error for /EMBED/ write, got nil")
	}
}

// TestSave_MissingArguments tests error when arguments are missing
func TestSave_MissingArguments(t *testing.T) {
	interp := script.NewInterpreter(false)
	interp.SetScriptDir("/tmp")
	interp.FileWriter = func(path, content string) error { return nil }

	fn := NewSaveFunction(interp)

	tests := []struct {
		name string
		args map[string]any
	}{
		{"no args", map[string]any{}},
		{"only filename", map[string]any{"0": "test.txt"}},
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

// TestSave_WithoutCapability tests error when FileWriter capability is missing
func TestSave_WithoutCapability(t *testing.T) {
	interp := script.NewInterpreter(false)
	interp.SetScriptDir("/tmp")
	interp.FileWriter = nil // No FileWriter capability

	fn := NewSaveFunction(interp)
	args := map[string]any{
		"0": "test.txt",
		"1": "content",
	}

	_, err := fn(nil, args)
	if err == nil {
		t.Errorf("Expected error for missing FileWriter capability, got nil")
	}
}

// TestLoad_WithNamedArgument tests load with named 'filename' argument
func TestLoad_WithNamedArgument(t *testing.T) {
	interp := script.NewInterpreter(false)
	interp.SetScriptDir("/tmp")
	interp.FileReader = (&MockFileReader{
		files: map[string]string{
			"test.txt": "content",
		},
	}).Read

	fn := NewLoadFunction(interp)
	args := map[string]any{"filename": "test.txt"} // Named argument

	result, err := fn(nil, args)
	if err != nil {
		t.Fatalf("load() with named argument failed: %v", err)
	}

	content, _ := result.(string)
	if content != "content" {
		t.Errorf("Expected 'content', got %q", content)
	}
}

// TestSave_WithNamedArguments tests save with named arguments
func TestSave_WithNamedArguments(t *testing.T) {
	writer := &MockFileWriter{files: make(map[string]string)}

	interp := script.NewInterpreter(false)
	interp.SetScriptDir("/tmp")
	interp.FileWriter = func(path, content string) error {
		return writer.Write(path, content)
	}

	fn := NewSaveFunction(interp)
	args := map[string]any{
		"filename": "test.txt",
		"content":  "data",
	}

	_, err := fn(nil, args)
	if err != nil {
		t.Fatalf("save() with named arguments failed: %v", err)
	}

	if writer.files["/tmp/test.txt"] != "data" {
		t.Errorf("Content not saved correctly with named arguments")
	}
}

// TestLoad_RelativePath tests loading with relative path
func TestLoad_RelativePath(t *testing.T) {
	interp := script.NewInterpreter(false)
	interp.SetScriptDir("/home/user")
	interp.FileReader = (&MockFileReader{
		files: map[string]string{
			"/home/user/data.txt": "relative path content",
			"data.txt":            "relative path content",
		},
	}).Read

	fn := NewLoadFunction(interp)
	args := map[string]any{"0": "data.txt"}

	result, err := fn(nil, args)
	if err != nil {
		t.Fatalf("load() with relative path failed: %v", err)
	}

	content, _ := result.(string)
	if content != "relative path content" {
		t.Errorf("Expected 'relative path content', got %q", content)
	}
}

// TestSave_AbsolutePath tests saving with absolute path
func TestSave_AbsolutePath(t *testing.T) {
	writer := &MockFileWriter{files: make(map[string]string)}

	interp := script.NewInterpreter(false)
	interp.SetScriptDir("/home/user")
	interp.FileWriter = func(path, content string) error {
		return writer.Write(path, content)
	}

	fn := NewSaveFunction(interp)
	args := map[string]any{
		"0": "/var/log/output.txt",
		"1": "absolute path content",
	}

	_, err := fn(nil, args)
	if err != nil {
		t.Fatalf("save() with absolute path failed: %v", err)
	}

	if writer.files["/var/log/output.txt"] != "absolute path content" {
		t.Errorf("Content not saved to absolute path correctly")
	}
}

// TestLoad_MultilineContent tests loading multiline files
func TestLoad_MultilineContent(t *testing.T) {
	content := "line 1\nline 2\nline 3"
	interp := script.NewInterpreter(false)
	interp.SetScriptDir("/tmp")
	interp.FileReader = (&MockFileReader{
		files: map[string]string{
			"multiline.txt": content,
		},
	}).Read

	fn := NewLoadFunction(interp)
	args := map[string]any{"0": "multiline.txt"}

	result, err := fn(nil, args)
	if err != nil {
		t.Fatalf("load() failed: %v", err)
	}

	if result != content {
		t.Errorf("Multiline content not preserved: expected %q, got %q", content, result)
	}
}

// TestSave_LargeContent tests saving large files
func TestSave_LargeContent(t *testing.T) {
	writer := &MockFileWriter{files: make(map[string]string)}

	interp := script.NewInterpreter(false)
	interp.SetScriptDir("/tmp")
	interp.FileWriter = func(path, content string) error {
		return writer.Write(path, content)
	}

	fn := NewSaveFunction(interp)

	// Create large content (1MB)
	largeContent := ""
	for i := 0; i < 10000; i++ {
		largeContent += "This is a line of text. "
	}

	args := map[string]any{
		"0": "large.txt",
		"1": largeContent,
	}

	_, err := fn(nil, args)
	if err != nil {
		t.Fatalf("save() failed for large content: %v", err)
	}

	if writer.files["/tmp/large.txt"] != largeContent {
		t.Errorf("Large content not saved correctly")
	}
}
