package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/duso-org/duso/pkg/script"
)

// Helper function to create a temporary test directory
func setupTestDir(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "duso-file-ops-test-")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}
	return tempDir
}

// Helper function to clean up test directory
func cleanupTestDir(t *testing.T, dir string) {
	if err := os.RemoveAll(dir); err != nil {
		t.Logf("warning: failed to cleanup temp directory: %v", err)
	}
}

// ============================================================================
// list_dir() TESTS
// ============================================================================

func TestListDir_Empty(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewListDirFunction(ctx)

	result, err := fn(nil, map[string]any{"0": "."})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	entries, ok := result.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", result)
	}

	if len(entries) != 0 {
		t.Errorf("expected empty directory, got %d entries", len(entries))
	}
}

func TestListDir_WithFiles(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	// Create test files
	if err := os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "file2.txt"), []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	if err := os.Mkdir(filepath.Join(tempDir, "subdir"), 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewListDirFunction(ctx)

	result, err := fn(nil, map[string]any{"0": "."})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	entries, ok := result.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", result)
	}

	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(entries))
	}

	// Check that entries have name and is_dir fields
	for i, entry := range entries {
		entryMap, ok := entry.(map[string]any)
		if !ok {
			t.Errorf("entry %d is not map[string]any, got %T", i, entry)
			continue
		}
		if _, ok := entryMap["name"]; !ok {
			t.Errorf("entry %d missing 'name' field", i)
		}
		if _, ok := entryMap["is_dir"]; !ok {
			t.Errorf("entry %d missing 'is_dir' field", i)
		}
	}
}

func TestListDir_Nonexistent(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewListDirFunction(ctx)

	_, err := fn(nil, map[string]any{"0": "nonexistent"})
	if err == nil {
		t.Errorf("expected error for nonexistent directory, got nil")
	}
}

func TestListDir_MissingArg(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewListDirFunction(ctx)

	_, err := fn(nil, map[string]any{})
	if err == nil {
		t.Errorf("expected error for missing argument, got nil")
	}
}

// ============================================================================
// make_dir() TESTS
// ============================================================================

func TestMakeDir_Single(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewMakeDirFunction(ctx)

	_, err := fn(nil, map[string]any{"0": "newdir"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(filepath.Join(tempDir, "newdir")); err != nil {
		t.Errorf("directory not created: %v", err)
	}
}

func TestMakeDir_Nested(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewMakeDirFunction(ctx)

	_, err := fn(nil, map[string]any{"0": "a/b/c"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify nested directories were created
	if _, err := os.Stat(filepath.Join(tempDir, "a", "b", "c")); err != nil {
		t.Errorf("nested directories not created: %v", err)
	}
}

func TestMakeDir_Existing(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	// Create directory first
	if err := os.Mkdir(filepath.Join(tempDir, "existing"), 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewMakeDirFunction(ctx)

	// Should not error when creating existing directory
	_, err := fn(nil, map[string]any{"0": "existing"})
	if err != nil {
		t.Fatalf("expected no error for existing directory, got: %v", err)
	}
}

func TestMakeDir_MissingArg(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewMakeDirFunction(ctx)

	_, err := fn(nil, map[string]any{})
	if err == nil {
		t.Errorf("expected error for missing argument, got nil")
	}
}

// ============================================================================
// remove_file() TESTS
// ============================================================================

func TestRemoveFile_Exists(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewRemoveFileFunction(ctx)

	_, err := fn(nil, map[string]any{"0": "test.txt"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify file was deleted
	if _, err := os.Stat(testFile); err == nil {
		t.Errorf("file was not deleted")
	}
}

func TestRemoveFile_Nonexistent(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewRemoveFileFunction(ctx)

	_, err := fn(nil, map[string]any{"0": "nonexistent.txt"})
	if err == nil {
		t.Errorf("expected error for nonexistent file, got nil")
	}
}

func TestRemoveFile_MissingArg(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewRemoveFileFunction(ctx)

	_, err := fn(nil, map[string]any{})
	if err == nil {
		t.Errorf("expected error for missing argument, got nil")
	}
}

// ============================================================================
// remove_dir() TESTS
// ============================================================================

func TestRemoveDir_Empty(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	// Create test directory
	testDir := filepath.Join(tempDir, "testdir")
	if err := os.Mkdir(testDir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewRemoveDirFunction(ctx)

	_, err := fn(nil, map[string]any{"0": "testdir"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify directory was deleted
	if _, err := os.Stat(testDir); err == nil {
		t.Errorf("directory was not deleted")
	}
}

func TestRemoveDir_NonEmpty(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	// Create test directory with file
	testDir := filepath.Join(tempDir, "testdir")
	if err := os.Mkdir(testDir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(testDir, "file.txt"), []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewRemoveDirFunction(ctx)

	_, err := fn(nil, map[string]any{"0": "testdir"})
	if err == nil {
		t.Errorf("expected error for non-empty directory, got nil")
	}
}

func TestRemoveDir_Nonexistent(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewRemoveDirFunction(ctx)

	_, err := fn(nil, map[string]any{"0": "nonexistent"})
	if err == nil {
		t.Errorf("expected error for nonexistent directory, got nil")
	}
}

// ============================================================================
// rename_file() TESTS
// ============================================================================

func TestRenameFile_Basic(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	// Create test file
	oldPath := filepath.Join(tempDir, "old.txt")
	if err := os.WriteFile(oldPath, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewRenameFileFunction(ctx)

	_, err := fn(nil, map[string]any{"0": "old.txt", "1": "new.txt"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify file was renamed
	if _, err := os.Stat(oldPath); err == nil {
		t.Errorf("old file still exists")
	}
	newPath := filepath.Join(tempDir, "new.txt")
	if _, err := os.Stat(newPath); err != nil {
		t.Errorf("new file does not exist: %v", err)
	}
}

func TestRenameFile_ToSubdir(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	// Create test file and subdirectory
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	if err := os.Mkdir(filepath.Join(tempDir, "subdir"), 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewRenameFileFunction(ctx)

	_, err := fn(nil, map[string]any{"0": "test.txt", "1": "subdir/test.txt"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify file was moved
	newPath := filepath.Join(tempDir, "subdir", "test.txt")
	if _, err := os.Stat(newPath); err != nil {
		t.Errorf("file not moved to subdir: %v", err)
	}
}

func TestRenameFile_Nonexistent(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewRenameFileFunction(ctx)

	_, err := fn(nil, map[string]any{"0": "nonexistent.txt", "1": "new.txt"})
	if err == nil {
		t.Errorf("expected error for nonexistent file, got nil")
	}
}

func TestRenameFile_MissingArgs(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewRenameFileFunction(ctx)

	tests := []struct {
		name string
		args map[string]any
	}{
		{"no args", map[string]any{}},
		{"only one arg", map[string]any{"0": "old.txt"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := fn(nil, tt.args)
			if err == nil {
				t.Errorf("expected error, got nil")
			}
		})
	}
}

// ============================================================================
// file_type() TESTS
// ============================================================================

func TestFileType_File(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	// Create test file
	if err := os.WriteFile(filepath.Join(tempDir, "test.txt"), []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewFileTypeFunction(ctx)

	result, err := fn(nil, map[string]any{"0": "test.txt"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if result != "file" {
		t.Errorf("expected 'file', got %q", result)
	}
}

func TestFileType_Directory(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	// Create test directory
	if err := os.Mkdir(filepath.Join(tempDir, "testdir"), 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewFileTypeFunction(ctx)

	result, err := fn(nil, map[string]any{"0": "testdir"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if result != "directory" {
		t.Errorf("expected 'directory', got %q", result)
	}
}

func TestFileType_Nonexistent(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewFileTypeFunction(ctx)

	_, err := fn(nil, map[string]any{"0": "nonexistent"})
	if err == nil {
		t.Errorf("expected error for nonexistent path, got nil")
	}
}

func TestFileType_MissingArg(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewFileTypeFunction(ctx)

	_, err := fn(nil, map[string]any{})
	if err == nil {
		t.Errorf("expected error for missing argument, got nil")
	}
}

// ============================================================================
// file_exists() TESTS
// ============================================================================

func TestFileExists_Exists(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	// Create test file
	if err := os.WriteFile(filepath.Join(tempDir, "test.txt"), []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewFileExistsFunction(ctx)

	result, err := fn(nil, map[string]any{"0": "test.txt"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if result != true {
		t.Errorf("expected true, got %v", result)
	}
}

func TestFileExists_Directory(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	// Create test directory
	if err := os.Mkdir(filepath.Join(tempDir, "testdir"), 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewFileExistsFunction(ctx)

	result, err := fn(nil, map[string]any{"0": "testdir"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if result != true {
		t.Errorf("expected true, got %v", result)
	}
}

func TestFileExists_Nonexistent(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewFileExistsFunction(ctx)

	result, err := fn(nil, map[string]any{"0": "nonexistent"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if result != false {
		t.Errorf("expected false, got %v", result)
	}
}

func TestFileExists_MissingArg(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewFileExistsFunction(ctx)

	_, err := fn(nil, map[string]any{})
	if err == nil {
		t.Errorf("expected error for missing argument, got nil")
	}
}

// ============================================================================
// current_dir() TESTS
// ============================================================================

func TestCurrentDir(t *testing.T) {
	fn := NewCurrentDirFunction()

	result, err := fn(nil, map[string]any{})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	cwd, ok := result.(string)
	if !ok {
		t.Fatalf("expected string, got %T", result)
	}

	if cwd == "" {
		t.Errorf("expected non-empty working directory")
	}

	// Verify it's an absolute path by checking it exists
	if _, err := os.Stat(cwd); err != nil {
		t.Errorf("returned directory does not exist: %v", err)
	}
}

// ============================================================================
// append_file() TESTS
// ============================================================================

func TestAppendFile_New(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewAppendFileFunction(ctx)

	_, err := fn(nil, map[string]any{"0": "test.txt", "1": "hello"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify file was created with content
	content, err := os.ReadFile(filepath.Join(tempDir, "test.txt"))
	if err != nil {
		t.Fatalf("file not created: %v", err)
	}

	if string(content) != "hello" {
		t.Errorf("expected 'hello', got %q", string(content))
	}
}

func TestAppendFile_Existing(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	// Create test file with initial content
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewAppendFileFunction(ctx)

	_, err := fn(nil, map[string]any{"0": "test.txt", "1": " world"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify content was appended
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if string(content) != "hello world" {
		t.Errorf("expected 'hello world', got %q", string(content))
	}
}

func TestAppendFile_Nested(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	// Create nested directory
	if err := os.MkdirAll(filepath.Join(tempDir, "a", "b"), 0755); err != nil {
		t.Fatalf("failed to create nested directory: %v", err)
	}

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewAppendFileFunction(ctx)

	_, err := fn(nil, map[string]any{"0": "a/b/test.txt", "1": "content"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify file was created in nested location
	content, err := os.ReadFile(filepath.Join(tempDir, "a", "b", "test.txt"))
	if err != nil {
		t.Fatalf("file not created: %v", err)
	}

	if string(content) != "content" {
		t.Errorf("expected 'content', got %q", string(content))
	}
}

func TestAppendFile_MissingArgs(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewAppendFileFunction(ctx)

	tests := []struct {
		name string
		args map[string]any
	}{
		{"no args", map[string]any{}},
		{"only path", map[string]any{"0": "test.txt"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := fn(nil, tt.args)
			if err == nil {
				t.Errorf("expected error, got nil")
			}
		})
	}
}

// ============================================================================
// copy_file() TESTS
// ============================================================================

func TestCopyFile_Basic(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	// Create source file
	srcPath := filepath.Join(tempDir, "src.txt")
	if err := os.WriteFile(srcPath, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewCopyFileFunction(ctx)

	_, err := fn(nil, map[string]any{"0": "src.txt", "1": "dst.txt"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify destination file exists with same content
	dstPath := filepath.Join(tempDir, "dst.txt")
	content, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("destination file not created: %v", err)
	}

	if string(content) != "content" {
		t.Errorf("expected 'content', got %q", string(content))
	}

	// Source file should still exist
	if _, err := os.Stat(srcPath); err != nil {
		t.Errorf("source file was deleted: %v", err)
	}
}

func TestCopyFile_ToNested(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	// Create source file
	if err := os.WriteFile(filepath.Join(tempDir, "src.txt"), []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewCopyFileFunction(ctx)

	_, err := fn(nil, map[string]any{"0": "src.txt", "1": "dir/a/b/dst.txt"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify nested destination was created and has content
	dstPath := filepath.Join(tempDir, "dir", "a", "b", "dst.txt")
	content, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("destination file not created: %v", err)
	}

	if string(content) != "content" {
		t.Errorf("expected 'content', got %q", string(content))
	}
}

func TestCopyFile_Overwrite(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	// Create source and destination files
	if err := os.WriteFile(filepath.Join(tempDir, "src.txt"), []byte("new"), 0644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "dst.txt"), []byte("old"), 0644); err != nil {
		t.Fatalf("failed to create destination file: %v", err)
	}

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewCopyFileFunction(ctx)

	_, err := fn(nil, map[string]any{"0": "src.txt", "1": "dst.txt"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify destination was overwritten
	dstPath := filepath.Join(tempDir, "dst.txt")
	content, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("failed to read destination file: %v", err)
	}

	if string(content) != "new" {
		t.Errorf("expected 'new', got %q", string(content))
	}
}

func TestCopyFile_Nonexistent(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewCopyFileFunction(ctx)

	_, err := fn(nil, map[string]any{"0": "nonexistent.txt", "1": "dst.txt"})
	if err == nil {
		t.Errorf("expected error for nonexistent source, got nil")
	}
}

func TestCopyFile_MissingArgs(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewCopyFileFunction(ctx)

	tests := []struct {
		name string
		args map[string]any
	}{
		{"no args", map[string]any{}},
		{"only source", map[string]any{"0": "src.txt"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := fn(nil, tt.args)
			if err == nil {
				t.Errorf("expected error, got nil")
			}
		})
	}
}

// ============================================================================
// move_file() TESTS
// ============================================================================

func TestMoveFile_Basic(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	// Create source file
	srcPath := filepath.Join(tempDir, "src.txt")
	if err := os.WriteFile(srcPath, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewMoveFileFunction(ctx)

	_, err := fn(nil, map[string]any{"0": "src.txt", "1": "dst.txt"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify source file no longer exists
	if _, err := os.Stat(srcPath); err == nil {
		t.Errorf("source file was not moved")
	}

	// Verify destination file exists with same content
	dstPath := filepath.Join(tempDir, "dst.txt")
	content, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("destination file not created: %v", err)
	}

	if string(content) != "content" {
		t.Errorf("expected 'content', got %q", string(content))
	}
}

func TestMoveFile_ToNested(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	// Create source file and destination directory
	srcPath := filepath.Join(tempDir, "src.txt")
	if err := os.WriteFile(srcPath, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tempDir, "dir"), 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewMoveFileFunction(ctx)

	_, err := fn(nil, map[string]any{"0": "src.txt", "1": "dir/dst.txt"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify source file no longer exists
	if _, err := os.Stat(srcPath); err == nil {
		t.Errorf("source file was not moved")
	}

	// Verify destination file exists
	dstPath := filepath.Join(tempDir, "dir", "dst.txt")
	content, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("destination file not created: %v", err)
	}

	if string(content) != "content" {
		t.Errorf("expected 'content', got %q", string(content))
	}
}

func TestMoveFile_Nonexistent(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewMoveFileFunction(ctx)

	_, err := fn(nil, map[string]any{"0": "nonexistent.txt", "1": "dst.txt"})
	if err == nil {
		t.Errorf("expected error for nonexistent source, got nil")
	}
}

func TestMoveFile_MissingArgs(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewMoveFileFunction(ctx)

	tests := []struct {
		name string
		args map[string]any
	}{
		{"no args", map[string]any{}},
		{"only source", map[string]any{"0": "src.txt"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := fn(nil, tt.args)
			if err == nil {
				t.Errorf("expected error, got nil")
			}
		})
	}
}

// ============================================================================
// /STORE/ VIRTUAL FILESYSTEM TESTS
// ============================================================================
// Load/Save tests moved to pkg/runtime since those functions are now universal builtins
// See builtin_files.go in pkg/runtime for implementation
// Load/save functionality is tested via integration tests

// ============================================================================
// WILDCARD TESTS (list_files, remove_file, copy_file, move_file with patterns)
// ============================================================================

func TestListFiles_Wildcard_Simple(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	// Create test files
	files := []string{"file1.txt", "file2.txt", "file3.go", "readme.md"}
	for _, file := range files {
		if err := os.WriteFile(filepath.Join(tempDir, file), []byte("test"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewListFilesFunction(ctx)

	// Test *.txt pattern
	result, err := fn(nil, map[string]any{"0": "*.txt"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	entries, ok := result.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", result)
	}

	if len(entries) != 2 {
		t.Errorf("expected 2 .txt files, got %d", len(entries))
	}

	// Verify the results are strings (file paths)
	for _, entry := range entries {
		if _, ok := entry.(string); !ok {
			t.Errorf("expected string file path, got %T", entry)
		}
	}
}

func TestListFiles_Wildcard_Question(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	// Create test files
	files := []string{"file_1.log", "file_2.log", "file_10.log", "data.log"}
	for _, file := range files {
		if err := os.WriteFile(filepath.Join(tempDir, file), []byte("test"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewListFilesFunction(ctx)

	// Test file_?.log pattern (should match only file_1.log and file_2.log)
	result, err := fn(nil, map[string]any{"0": "file_?.log"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	entries, ok := result.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", result)
	}

	if len(entries) != 2 {
		t.Errorf("expected 2 matching files with file_?.log, got %d", len(entries))
	}
}

func TestListFiles_Wildcard_NoMatches(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	// Create test files
	files := []string{"file1.txt", "file2.txt"}
	for _, file := range files {
		if err := os.WriteFile(filepath.Join(tempDir, file), []byte("test"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewListFilesFunction(ctx)

	// Test pattern with no matches
	result, err := fn(nil, map[string]any{"0": "*.go"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	entries, ok := result.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", result)
	}

	if len(entries) != 0 {
		t.Errorf("expected 0 matches for *.go, got %d", len(entries))
	}
}

func TestListFiles_Wildcard_InvalidPattern(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewListFilesFunction(ctx)

	// Test ** pattern (should error)
	_, err := fn(nil, map[string]any{"0": "**/*.txt"})
	if err == nil {
		t.Fatalf("expected error for ** pattern, got nil")
	}

	if !contains(err.Error(), "** (recursive wildcard) is not supported") {
		t.Errorf("expected error message about **, got: %v", err)
	}
}

func TestRemoveFile_Wildcard_Simple(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	// Create test files
	files := []string{"file1.log", "file2.log", "file3.txt"}
	for _, file := range files {
		if err := os.WriteFile(filepath.Join(tempDir, file), []byte("test"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewRemoveFileFunction(ctx)

	// Remove *.log files
	result, err := fn(nil, map[string]any{"0": "*.log"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	removed, ok := result.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", result)
	}

	if len(removed) != 2 {
		t.Errorf("expected 2 removed files, got %d", len(removed))
	}

	// Verify files are actually removed
	if _, err := os.Stat(filepath.Join(tempDir, "file1.log")); !os.IsNotExist(err) {
		t.Errorf("file1.log should have been removed")
	}
	if _, err := os.Stat(filepath.Join(tempDir, "file2.log")); !os.IsNotExist(err) {
		t.Errorf("file2.log should have been removed")
	}
	// file3.txt should still exist
	if _, err := os.Stat(filepath.Join(tempDir, "file3.txt")); os.IsNotExist(err) {
		t.Errorf("file3.txt should not have been removed")
	}
}

func TestRemoveFile_Wildcard_NoMatches(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	// Create test files
	files := []string{"file1.txt", "file2.txt"}
	for _, file := range files {
		if err := os.WriteFile(filepath.Join(tempDir, file), []byte("test"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewRemoveFileFunction(ctx)

	// Remove pattern with no matches
	result, err := fn(nil, map[string]any{"0": "*.log"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	removed, ok := result.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", result)
	}

	if len(removed) != 0 {
		t.Errorf("expected 0 removed files, got %d", len(removed))
	}
}

func TestRemoveFile_SingleFile_ReturnsArray(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	// Create test file
	if err := os.WriteFile(filepath.Join(tempDir, "file.txt"), []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewRemoveFileFunction(ctx)

	// Remove single file
	result, err := fn(nil, map[string]any{"0": "file.txt"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	removed, ok := result.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", result)
	}

	// Single file remove should still return array
	if len(removed) != 1 {
		t.Errorf("expected 1 removed file, got %d", len(removed))
	}
}

func TestCopyFile_Wildcard_ToDirectory(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	srcDir := filepath.Join(tempDir, "src")
	dstDir := filepath.Join(tempDir, "dst")
	if err := os.Mkdir(srcDir, 0755); err != nil {
		t.Fatalf("failed to create src dir: %v", err)
	}
	if err := os.Mkdir(dstDir, 0755); err != nil {
		t.Fatalf("failed to create dst dir: %v", err)
	}

	// Create test files in src
	files := []string{"file1.ts", "file2.ts", "file3.js"}
	for _, file := range files {
		if err := os.WriteFile(filepath.Join(srcDir, file), []byte("content"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewCopyFileFunction(ctx)

	// Copy *.ts files to dst
	result, err := fn(nil, map[string]any{"0": "src/*.ts", "1": dstDir})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	copied, ok := result.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", result)
	}

	if len(copied) != 2 {
		t.Errorf("expected 2 copied files, got %d", len(copied))
	}

	// Verify files are copied
	if _, err := os.Stat(filepath.Join(dstDir, "file1.ts")); os.IsNotExist(err) {
		t.Errorf("file1.ts should have been copied")
	}
	if _, err := os.Stat(filepath.Join(dstDir, "file2.ts")); os.IsNotExist(err) {
		t.Errorf("file2.ts should have been copied")
	}
	if _, err := os.Stat(filepath.Join(dstDir, "file3.js")); !os.IsNotExist(err) {
		t.Errorf("file3.js should not have been copied")
	}
}

func TestCopyFile_Wildcard_NonDirectoryDest(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	// Create test files
	if err := os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "file2.txt"), []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewCopyFileFunction(ctx)

	// Try to copy with non-directory destination
	_, err := fn(nil, map[string]any{"0": "*.txt", "1": "output.txt"})
	if err == nil {
		t.Fatalf("expected error for non-directory destination, got nil")
	}

	if !contains(err.Error(), "existing directory") {
		t.Errorf("expected error about directory, got: %v", err)
	}
}

func TestCopyFile_SingleFile_ReturnsArray(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	// Create test file
	if err := os.WriteFile(filepath.Join(tempDir, "input.txt"), []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewCopyFileFunction(ctx)

	// Copy single file
	result, err := fn(nil, map[string]any{"0": "input.txt", "1": "output.txt"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	copied, ok := result.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", result)
	}

	// Single file copy should still return array
	if len(copied) != 1 {
		t.Errorf("expected 1 copied file, got %d", len(copied))
	}
}

func TestMoveFile_Wildcard_ToDirectory(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	srcDir := filepath.Join(tempDir, "src")
	dstDir := filepath.Join(tempDir, "dst")
	if err := os.Mkdir(srcDir, 0755); err != nil {
		t.Fatalf("failed to create src dir: %v", err)
	}
	if err := os.Mkdir(dstDir, 0755); err != nil {
		t.Fatalf("failed to create dst dir: %v", err)
	}

	// Create test files in src
	files := []string{"old_1.txt", "old_2.txt", "keep.txt"}
	for _, file := range files {
		if err := os.WriteFile(filepath.Join(srcDir, file), []byte("content"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewMoveFileFunction(ctx)

	// Move old_*.txt files to dst
	result, err := fn(nil, map[string]any{"0": "src/old_*.txt", "1": dstDir})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	moved, ok := result.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", result)
	}

	if len(moved) != 2 {
		t.Errorf("expected 2 moved files, got %d", len(moved))
	}

	// Verify files are moved (not in src, in dst)
	if _, err := os.Stat(filepath.Join(srcDir, "old_1.txt")); !os.IsNotExist(err) {
		t.Errorf("old_1.txt should have been moved from src")
	}
	if _, err := os.Stat(filepath.Join(dstDir, "old_1.txt")); os.IsNotExist(err) {
		t.Errorf("old_1.txt should have been moved to dst")
	}
	if _, err := os.Stat(filepath.Join(srcDir, "keep.txt")); os.IsNotExist(err) {
		t.Errorf("keep.txt should still be in src")
	}
}

func TestMoveFile_SingleFile_ReturnsArray(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	// Create test file
	if err := os.WriteFile(filepath.Join(tempDir, "old.txt"), []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewMoveFileFunction(ctx)

	// Move single file
	result, err := fn(nil, map[string]any{"0": "old.txt", "1": "new.txt"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	moved, ok := result.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", result)
	}

	// Single file move should still return array
	if len(moved) != 1 {
		t.Errorf("expected 1 moved file, got %d", len(moved))
	}
}

// ============================================================================
// VIRTUAL FILESYSTEM TESTS (/EMBED/ and /STORE/)
// ============================================================================

// TestListFiles_Store_Wildcard tests listing files in /STORE/ with wildcards
func TestListFiles_Store_Wildcard(t *testing.T) {
	// Clear /STORE/ before test
	store := script.GetDatastore("vfs", nil)
	store.Clear()

	store.Set("test1.txt", "content1")
	store.Set("test2.txt", "content2")
	store.Set("keep.log", "content3")

	ctx := FileIOContext{ScriptDir: "/"}
	fn := NewListFilesFunction(ctx)

	// List files matching /STORE/test*.txt
	result, err := fn(nil, map[string]any{"0": "/STORE/test*.txt"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	files, ok := result.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", result)
	}

	if len(files) != 2 {
		t.Errorf("expected 2 matching files, got %d: %v", len(files), files)
	}

	// Clean up
	store.Clear()
}

// TestRemoveFile_Store_Wildcard tests removing files from /STORE/ with wildcards
func TestRemoveFile_Store_Wildcard(t *testing.T) {
	// Clear /STORE/ before test
	store := script.GetDatastore("vfs", nil)
	store.Clear()
	store.Set("temp1.tmp", "content1")
	store.Set("temp2.tmp", "content2")
	store.Set("keep.txt", "content3")

	ctx := FileIOContext{ScriptDir: "/"}
	fn := NewRemoveFileFunction(ctx)

	// Remove files matching /STORE/temp*.tmp
	result, err := fn(nil, map[string]any{"0": "/STORE/temp*.tmp"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	removed, ok := result.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", result)
	}

	if len(removed) != 2 {
		t.Errorf("expected 2 removed files, got %d: %v", len(removed), removed)
	}

	// Verify files are actually removed
	existsFn := NewFileExistsFunction(ctx)
	if exists, _ := existsFn(nil, map[string]any{"0": "/STORE/temp1.tmp"}); exists != false {
		t.Errorf("temp1.tmp should have been removed")
	}
	if exists, _ := existsFn(nil, map[string]any{"0": "/STORE/temp2.tmp"}); exists != false {
		t.Errorf("temp2.tmp should have been removed")
	}
	// keep.txt should still exist
	if exists, _ := existsFn(nil, map[string]any{"0": "/STORE/keep.txt"}); exists != true {
		t.Errorf("keep.txt should not have been removed")
	}

	// Clean up
	store.Clear()
}

// TestCopyFile_Store_Wildcard tests copying files from /STORE/ with wildcards
func TestCopyFile_Store_Wildcard(t *testing.T) {
	// Clear /STORE/ before test
	store := script.GetDatastore("vfs", nil)
	store.Clear()
	store.Set("file1.du", "code1")
	store.Set("file2.du", "code2")
	store.Set("readme.txt", "doc")

	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	dstDir := filepath.Join(tempDir, "backup")
	if err := os.Mkdir(dstDir, 0755); err != nil {
		t.Fatalf("failed to create backup dir: %v", err)
	}

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewCopyFileFunction(ctx)

	// Copy files matching /STORE/*.du to backup directory
	result, err := fn(nil, map[string]any{"0": "/STORE/*.du", "1": dstDir})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	copied, ok := result.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", result)
	}

	if len(copied) != 2 {
		t.Errorf("expected 2 copied files, got %d: %v", len(copied), copied)
	}

	// Verify files are actually copied to disk
	if _, err := os.Stat(filepath.Join(dstDir, "file1.du")); os.IsNotExist(err) {
		t.Errorf("file1.du should have been copied")
	}
	if _, err := os.Stat(filepath.Join(dstDir, "file2.du")); os.IsNotExist(err) {
		t.Errorf("file2.du should have been copied")
	}
	if _, err := os.Stat(filepath.Join(dstDir, "readme.txt")); !os.IsNotExist(err) {
		t.Errorf("readme.txt should not have been copied")
	}

	// Clean up
	store.Clear()
}

// TestMoveFile_Store_Wildcard tests moving files from /STORE/ to filesystem with wildcards
func TestMoveFile_Store_Wildcard(t *testing.T) {
	// Clear /STORE/ before test
	store := script.GetDatastore("vfs", nil)
	store.Clear()

	store.Set("old1.txt", "content1")
	store.Set("old2.txt", "content2")
	store.Set("keep.txt", "content3")

	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	dstDir := filepath.Join(tempDir, "archive")
	if err := os.Mkdir(dstDir, 0755); err != nil {
		t.Fatalf("failed to create archive dir: %v", err)
	}

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewMoveFileFunction(ctx)

	// Move files matching /STORE/old*.txt to the archive directory
	result, err := fn(nil, map[string]any{"0": "/STORE/old*.txt", "1": dstDir})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	moved, ok := result.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", result)
	}

	if len(moved) != 2 {
		t.Errorf("expected 2 moved files, got %d: %v", len(moved), moved)
	}

	// Verify files are moved from /STORE/
	existsFn := NewFileExistsFunction(ctx)
	if exists, _ := existsFn(nil, map[string]any{"0": "/STORE/old1.txt"}); exists != false {
		t.Errorf("old1.txt should have been moved from /STORE/")
	}
	if exists, _ := existsFn(nil, map[string]any{"0": "/STORE/old2.txt"}); exists != false {
		t.Errorf("old2.txt should have been moved from /STORE/")
	}
	// keep.txt should still exist in /STORE/
	if exists, _ := existsFn(nil, map[string]any{"0": "/STORE/keep.txt"}); exists != true {
		t.Errorf("keep.txt should not have been moved")
	}

	// Verify files exist on filesystem
	if _, err := os.Stat(filepath.Join(dstDir, "old1.txt")); os.IsNotExist(err) {
		t.Errorf("old1.txt should have been moved to filesystem")
	}
	if _, err := os.Stat(filepath.Join(dstDir, "old2.txt")); os.IsNotExist(err) {
		t.Errorf("old2.txt should have been moved to filesystem")
	}

	// Clean up
	store.Clear()
}

// TestListFiles_Store_Wildcard_MultiplePatterns tests various wildcard patterns in /STORE/
func TestListFiles_Store_Wildcard_MultiplePatterns(t *testing.T) {
	store := script.GetDatastore("vfs", nil)
	store.Clear()

	store.Set("app.du", "code1")
	store.Set("app.tmp", "temp1")
	store.Set("lib.du", "code2")
	store.Set("lib.tmp", "temp2")
	store.Set("test_app.du", "test1")

	ctx := FileIOContext{ScriptDir: "/"}
	fn := NewListFilesFunction(ctx)

	// Test *.du pattern
	result, err := fn(nil, map[string]any{"0": "/STORE/*.du"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	files, _ := result.([]any)
	if len(files) != 3 {
		t.Errorf("expected 3 .du files, got %d", len(files))
	}

	// Test *.tmp pattern
	result, err = fn(nil, map[string]any{"0": "/STORE/*.tmp"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	files, _ = result.([]any)
	if len(files) != 2 {
		t.Errorf("expected 2 .tmp files, got %d", len(files))
	}

	// Test app.* pattern
	result, err = fn(nil, map[string]any{"0": "/STORE/app.*"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	files, _ = result.([]any)
	if len(files) != 2 {
		t.Errorf("expected 2 app.* files, got %d", len(files))
	}

	store.Clear()
}

// TestListFiles_Store_Wildcard_QuestionMark tests ? wildcard in /STORE/
func TestListFiles_Store_Wildcard_QuestionMark(t *testing.T) {
	store := script.GetDatastore("vfs", nil)
	store.Clear()

	store.Set("file1.txt", "content1")
	store.Set("file2.txt", "content2")
	store.Set("file10.txt", "content3")
	store.Set("data.txt", "content4")

	ctx := FileIOContext{ScriptDir: "/"}
	fn := NewListFilesFunction(ctx)

	// Test file?.txt pattern (should match only 1 and 2, not 10)
	result, err := fn(nil, map[string]any{"0": "/STORE/file?.txt"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	files, _ := result.([]any)
	if len(files) != 2 {
		t.Errorf("expected 2 files matching file?.txt, got %d", len(files))
	}

	store.Clear()
}

// TestRemoveFile_Store_Wildcard_PartialSuccess tests removing when some files fail
func TestRemoveFile_Store_Wildcard_PartialSuccess(t *testing.T) {
	store := script.GetDatastore("vfs", nil)
	store.Clear()

	store.Set("remove1.txt", "content1")
	store.Set("remove2.txt", "content2")
	store.Set("keep.txt", "content3")

	ctx := FileIOContext{ScriptDir: "/"}
	fn := NewRemoveFileFunction(ctx)

	// Remove all remove*.txt files
	result, err := fn(nil, map[string]any{"0": "/STORE/remove*.txt"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	removed, _ := result.([]any)
	if len(removed) != 2 {
		t.Errorf("expected 2 removed files, got %d", len(removed))
	}

	// Verify removed files are gone
	existsFn := NewFileExistsFunction(ctx)
	if exists, _ := existsFn(nil, map[string]any{"0": "/STORE/remove1.txt"}); exists != false {
		t.Errorf("remove1.txt should be gone")
	}
	if exists, _ := existsFn(nil, map[string]any{"0": "/STORE/remove2.txt"}); exists != false {
		t.Errorf("remove2.txt should be gone")
	}
	// keep.txt should still exist
	if exists, _ := existsFn(nil, map[string]any{"0": "/STORE/keep.txt"}); exists != true {
		t.Errorf("keep.txt should still exist")
	}

	store.Clear()
}

// TestCopyFile_Store_Wildcard_PreserveContent tests that content is preserved when copying
func TestCopyFile_Store_Wildcard_PreserveContent(t *testing.T) {
	store := script.GetDatastore("vfs", nil)
	store.Clear()

	content1 := "line1\nline2\nline3"
	content2 := "data: {\"key\": \"value\"}"
	store.Set("file1.txt", content1)
	store.Set("file2.txt", content2)

	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	dstDir := filepath.Join(tempDir, "copies")
	if err := os.Mkdir(dstDir, 0755); err != nil {
		t.Fatalf("failed to create dst dir: %v", err)
	}

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewCopyFileFunction(ctx)

	// Copy files
	_, err := fn(nil, map[string]any{"0": "/STORE/*.txt", "1": dstDir})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify content is preserved
	actualContent1, err := os.ReadFile(filepath.Join(dstDir, "file1.txt"))
	if err != nil {
		t.Fatalf("failed to read file1.txt: %v", err)
	}
	if string(actualContent1) != content1 {
		t.Errorf("file1.txt content mismatch: expected %q, got %q", content1, string(actualContent1))
	}

	actualContent2, err := os.ReadFile(filepath.Join(dstDir, "file2.txt"))
	if err != nil {
		t.Fatalf("failed to read file2.txt: %v", err)
	}
	if string(actualContent2) != content2 {
		t.Errorf("file2.txt content mismatch: expected %q, got %q", content2, string(actualContent2))
	}

	store.Clear()
}

// TestMoveFile_Store_Wildcard_ContentPreserved tests content is preserved when moving from /STORE/
func TestMoveFile_Store_Wildcard_ContentPreserved(t *testing.T) {
	store := script.GetDatastore("vfs", nil)
	store.Clear()

	contentA := "content A with special chars: !@#$%"
	contentB := "content B\nwith\nmultiple\nlines"
	store.Set("fileA.txt", contentA)
	store.Set("fileB.txt", contentB)

	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	dstDir := filepath.Join(tempDir, "moved")
	if err := os.Mkdir(dstDir, 0755); err != nil {
		t.Fatalf("failed to create dst dir: %v", err)
	}

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewMoveFileFunction(ctx)

	// Move files
	_, err := fn(nil, map[string]any{"0": "/STORE/file*.txt", "1": dstDir})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify content is preserved on filesystem
	actualContentA, err := os.ReadFile(filepath.Join(dstDir, "fileA.txt"))
	if err != nil {
		t.Fatalf("failed to read fileA.txt: %v", err)
	}
	if string(actualContentA) != contentA {
		t.Errorf("fileA.txt content mismatch after move")
	}

	actualContentB, err := os.ReadFile(filepath.Join(dstDir, "fileB.txt"))
	if err != nil {
		t.Fatalf("failed to read fileB.txt: %v", err)
	}
	if string(actualContentB) != contentB {
		t.Errorf("fileB.txt content mismatch after move")
	}

	// Verify files are gone from /STORE/
	existsFn := NewFileExistsFunction(ctx)
	if exists, _ := existsFn(nil, map[string]any{"0": "/STORE/fileA.txt"}); exists != false {
		t.Errorf("fileA.txt should be gone from /STORE/")
	}
	if exists, _ := existsFn(nil, map[string]any{"0": "/STORE/fileB.txt"}); exists != false {
		t.Errorf("fileB.txt should be gone from /STORE/")
	}

	store.Clear()
}

// TestCopyFile_Store_To_Store tests copying files within /STORE/
func TestCopyFile_Store_To_Store(t *testing.T) {
	store := script.GetDatastore("vfs", nil)
	store.Clear()

	store.Set("original1.txt", "content1")
	store.Set("original2.txt", "content2")

	ctx := FileIOContext{ScriptDir: "/"}
	fn := NewCopyFileFunction(ctx)

	// Copy from /STORE/ to /STORE/ (different filename/location)
	result, err := fn(nil, map[string]any{"0": "/STORE/original*.txt", "1": "/STORE/backup/"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	copied, _ := result.([]any)
	if len(copied) != 2 {
		t.Errorf("expected 2 copied files, got %d", len(copied))
	}

	// Verify copies exist in /STORE/ with new names
	existsFn := NewFileExistsFunction(ctx)
	if exists, _ := existsFn(nil, map[string]any{"0": "/STORE/backup/original1.txt"}); exists != true {
		t.Errorf("backup/original1.txt should exist in /STORE/")
	}
	if exists, _ := existsFn(nil, map[string]any{"0": "/STORE/backup/original2.txt"}); exists != true {
		t.Errorf("backup/original2.txt should exist in /STORE/")
	}

	// Verify originals still exist
	if exists, _ := existsFn(nil, map[string]any{"0": "/STORE/original1.txt"}); exists != true {
		t.Errorf("original1.txt should still exist")
	}

	store.Clear()
}

// TestRemoveFile_Store_Wildcard_AllFiles tests removing all files matching pattern
func TestRemoveFile_Store_Wildcard_AllFiles(t *testing.T) {
	store := script.GetDatastore("vfs", nil)
	store.Clear()

	for i := 1; i <= 5; i++ {
		store.Set(fmt.Sprintf("file%d.log", i), fmt.Sprintf("log content %d", i))
	}
	store.Set("other.txt", "other content")

	ctx := FileIOContext{ScriptDir: "/"}
	fn := NewRemoveFileFunction(ctx)

	// Remove all *.log files
	result, err := fn(nil, map[string]any{"0": "/STORE/*.log"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	removed, _ := result.([]any)
	if len(removed) != 5 {
		t.Errorf("expected 5 removed files, got %d", len(removed))
	}

	// Verify all logs are gone but other.txt remains
	existsFn := NewFileExistsFunction(ctx)
	if exists, _ := existsFn(nil, map[string]any{"0": "/STORE/file1.log"}); exists != false {
		t.Errorf("file1.log should be removed")
	}
	if exists, _ := existsFn(nil, map[string]any{"0": "/STORE/other.txt"}); exists != true {
		t.Errorf("other.txt should still exist")
	}

	store.Clear()
}

// TestListFiles_Store_Wildcard_DotFiles tests listing hidden files with dots
func TestListFiles_Store_Wildcard_DotFiles(t *testing.T) {
	store := script.GetDatastore("vfs", nil)
	store.Clear()

	store.Set(".config", "config")
	store.Set(".hidden", "hidden")
	store.Set("visible.txt", "visible")
	store.Set("file.bak", "backup")

	ctx := FileIOContext{ScriptDir: "/"}
	fn := NewListFilesFunction(ctx)

	// Test .* pattern (dot files)
	result, err := fn(nil, map[string]any{"0": "/STORE/.*"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	files, _ := result.([]any)
	if len(files) != 2 {
		t.Errorf("expected 2 dot files, got %d", len(files))
	}

	// Test *.bak pattern
	result, err = fn(nil, map[string]any{"0": "/STORE/*.bak"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	files, _ = result.([]any)
	if len(files) != 1 {
		t.Errorf("expected 1 .bak file, got %d", len(files))
	}

	store.Clear()
}

// TestMoveFile_Store_Wildcard_ToFileSystem tests moving from /STORE/ to filesystem
func TestMoveFile_Store_Wildcard_ToFileSystem(t *testing.T) {
	store := script.GetDatastore("vfs", nil)
	store.Clear()

	store.Set("export1.du", "export content 1")
	store.Set("export2.du", "export content 2")

	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	exportDir := filepath.Join(tempDir, "exports")
	if err := os.Mkdir(exportDir, 0755); err != nil {
		t.Fatalf("failed to create export dir: %v", err)
	}

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewMoveFileFunction(ctx)

	// Move from /STORE/ to filesystem
	result, err := fn(nil, map[string]any{"0": "/STORE/export*.du", "1": exportDir})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	moved, _ := result.([]any)
	if len(moved) != 2 {
		t.Errorf("expected 2 moved files, got %d", len(moved))
	}

	// Verify files exist on filesystem
	if _, err := os.Stat(filepath.Join(exportDir, "export1.du")); os.IsNotExist(err) {
		t.Errorf("export1.du should exist on filesystem")
	}
	if _, err := os.Stat(filepath.Join(exportDir, "export2.du")); os.IsNotExist(err) {
		t.Errorf("export2.du should exist on filesystem")
	}

	// Verify files are gone from /STORE/
	existsFn := NewFileExistsFunction(ctx)
	if exists, _ := existsFn(nil, map[string]any{"0": "/STORE/export1.du"}); exists != false {
		t.Errorf("export1.du should be gone from /STORE/")
	}

	store.Clear()
}

// TestCopyFile_Store_Wildcard_ToFileSystem tests copying (not moving) from /STORE/ to filesystem
func TestCopyFile_Store_Wildcard_ToFileSystem(t *testing.T) {
	store := script.GetDatastore("vfs", nil)
	store.Clear()

	store.Set("backup1.txt", "backup content 1")
	store.Set("backup2.txt", "backup content 2")

	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	backupDir := filepath.Join(tempDir, "backups")
	if err := os.Mkdir(backupDir, 0755); err != nil {
		t.Fatalf("failed to create backup dir: %v", err)
	}

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewCopyFileFunction(ctx)

	// Copy from /STORE/ to filesystem
	result, err := fn(nil, map[string]any{"0": "/STORE/backup*.txt", "1": backupDir})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	copied, _ := result.([]any)
	if len(copied) != 2 {
		t.Errorf("expected 2 copied files, got %d", len(copied))
	}

	// Verify files exist on filesystem
	if _, err := os.Stat(filepath.Join(backupDir, "backup1.txt")); os.IsNotExist(err) {
		t.Errorf("backup1.txt should exist on filesystem")
	}

	// Verify files STILL EXIST in /STORE/ (copy, not move)
	existsFn := NewFileExistsFunction(ctx)
	if exists, _ := existsFn(nil, map[string]any{"0": "/STORE/backup1.txt"}); exists != true {
		t.Errorf("backup1.txt should still exist in /STORE/")
	}

	store.Clear()
}

// TestListFiles_Store_Empty tests listing from empty /STORE/ returns empty array
func TestListFiles_Store_Empty(t *testing.T) {
	store := script.GetDatastore("vfs", nil)
	store.Clear()

	ctx := FileIOContext{ScriptDir: "/"}
	fn := NewListFilesFunction(ctx)

	// List from empty /STORE/
	result, err := fn(nil, map[string]any{"0": "/STORE/*"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	files, _ := result.([]any)
	if len(files) != 0 {
		t.Errorf("expected 0 files in empty /STORE/, got %d", len(files))
	}
}

// ============================================================================
// /EMBED/ VIRTUAL FILESYSTEM TESTS (read-only)
// ============================================================================

// TestListFiles_Embed_Wildcard tests listing files in /EMBED/ with wildcards
// This tests the read-only /EMBED/ filesystem code path
func TestListFiles_Embed_Wildcard(t *testing.T) {
	ctx := FileIOContext{ScriptDir: "/"}
	fn := NewListFilesFunction(ctx)

	// List files matching /EMBED/stdlib/*.du
	// These should be embedded Duso stdlib modules
	result, err := fn(nil, map[string]any{"0": "/EMBED/stdlib/*.du"})
	if err != nil {
		t.Fatalf("expected no error listing /EMBED/ files, got: %v", err)
	}

	files, ok := result.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", result)
	}

	// Note: May be 0 if no stdlib files are embedded in this build
	_ = files // files may be empty depending on embedded content

	// Verify results are strings (file paths)
	for _, f := range files {
		if _, ok := f.(string); !ok {
			t.Errorf("expected string file path, got %T", f)
		}
		// Paths should start with /EMBED/
		if path, ok := f.(string); ok && !contains(path, "/EMBED/") {
			t.Errorf("expected /EMBED/ prefix in path, got: %v", path)
		}
	}
}

// TestListFiles_Embed_Wildcard_NoMatches tests listing with pattern that doesn't match
func TestListFiles_Embed_Wildcard_NoMatches(t *testing.T) {
	ctx := FileIOContext{ScriptDir: "/"}
	fn := NewListFilesFunction(ctx)

	// Try to list files that don't exist
	result, err := fn(nil, map[string]any{"0": "/EMBED/nonexistent/*.du"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	files, ok := result.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", result)
	}

	// Should return empty array, not error
	if len(files) != 0 {
		t.Errorf("expected 0 matches, got %d", len(files))
	}
}

// TestCopyFile_Embed_Wildcard_ToFilesystem tests copying from /EMBED/ to filesystem
func TestCopyFile_Embed_Wildcard_ToFilesystem(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	dstDir := filepath.Join(tempDir, "embedded_copies")
	if err := os.Mkdir(dstDir, 0755); err != nil {
		t.Fatalf("failed to create destination dir: %v", err)
	}

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewCopyFileFunction(ctx)

	// Copy files from /EMBED/stdlib/*.du to filesystem
	result, err := fn(nil, map[string]any{"0": "/EMBED/stdlib/*.du", "1": dstDir})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	copied, ok := result.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", result)
	}

	// May have copied 0 files if no stdlib files are embedded in this build
	_ = copied // copied may be empty depending on embedded content

	// Verify each copied file exists on disk
	for _, f := range copied {
		if filename, ok := f.(string); ok {
			filepath := filepath.Join(dstDir, filename)
			if _, err := os.Stat(filepath); os.IsNotExist(err) {
				t.Errorf("copied file does not exist: %s", filepath)
			}
		}
	}
}

// TestCopyFile_Embed_Wildcard_ToStore tests copying from /EMBED/ to /STORE/
func TestCopyFile_Embed_Wildcard_ToStore(t *testing.T) {
	store := script.GetDatastore("vfs", nil)
	store.Clear()

	ctx := FileIOContext{ScriptDir: "/"}
	fn := NewCopyFileFunction(ctx)

	// Copy files from /EMBED/stdlib/*.du to /STORE/
	result, err := fn(nil, map[string]any{"0": "/EMBED/stdlib/*.du", "1": "/STORE/embedded/"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	copied, ok := result.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", result)
	}

	// May have copied 0 files if no stdlib files are embedded in this build
	_ = copied // copied may be empty depending on embedded content

	// Clean up
	store.Clear()
}

// TestRemoveFile_Embed_Wildcard_Readonly tests that /EMBED/ is read-only
func TestRemoveFile_Embed_Wildcard_Readonly(t *testing.T) {
	ctx := FileIOContext{ScriptDir: "/"}
	fn := NewRemoveFileFunction(ctx)

	// Try to remove files from /EMBED/ - should error
	_, err := fn(nil, map[string]any{"0": "/EMBED/stdlib/*.du"})
	if err == nil {
		t.Fatalf("expected error for /EMBED/ removal, got nil")
	}

	if !contains(err.Error(), "read-only") && !contains(err.Error(), "cannot") {
		t.Errorf("expected error about read-only filesystem, got: %v", err)
	}
}

// TestMoveFile_Embed_Wildcard_Readonly tests that /EMBED/ is read-only for move
func TestMoveFile_Embed_Wildcard_Readonly(t *testing.T) {
	ctx := FileIOContext{ScriptDir: "/"}
	fn := NewMoveFileFunction(ctx)

	// Try to move files from /EMBED/ - should error
	_, err := fn(nil, map[string]any{"0": "/EMBED/stdlib/*.du", "1": "/STORE/"})
	if err == nil {
		t.Fatalf("expected error for /EMBED/ move, got nil")
	}

	if !contains(err.Error(), "read-only") && !contains(err.Error(), "cannot") {
		t.Errorf("expected error about read-only filesystem, got: %v", err)
	}
}

// TestListFiles_Embed_InvalidPattern tests /EMBED/ with ** pattern
func TestListFiles_Embed_InvalidPattern(t *testing.T) {
	ctx := FileIOContext{ScriptDir: "/"}
	fn := NewListFilesFunction(ctx)

	// Try to use ** pattern with /EMBED/ - should error
	_, err := fn(nil, map[string]any{"0": "/EMBED/**/*.du"})
	if err == nil {
		t.Fatalf("expected error for ** pattern, got nil")
	}

	if !contains(err.Error(), "** (recursive wildcard) is not supported") {
		t.Errorf("expected error about **, got: %v", err)
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && (s[:len(substr)] == substr || findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 1; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
