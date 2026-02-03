package cli

import (
	"os"
	"path/filepath"
	"testing"
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

	result, err := fn(map[string]any{"0": "."})
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

	result, err := fn(map[string]any{"0": "."})
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

	_, err := fn(map[string]any{"0": "nonexistent"})
	if err == nil {
		t.Errorf("expected error for nonexistent directory, got nil")
	}
}

func TestListDir_MissingArg(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewListDirFunction(ctx)

	_, err := fn(map[string]any{})
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

	_, err := fn(map[string]any{"0": "newdir"})
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

	_, err := fn(map[string]any{"0": "a/b/c"})
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
	_, err := fn(map[string]any{"0": "existing"})
	if err != nil {
		t.Fatalf("expected no error for existing directory, got: %v", err)
	}
}

func TestMakeDir_MissingArg(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewMakeDirFunction(ctx)

	_, err := fn(map[string]any{})
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

	_, err := fn(map[string]any{"0": "test.txt"})
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

	_, err := fn(map[string]any{"0": "nonexistent.txt"})
	if err == nil {
		t.Errorf("expected error for nonexistent file, got nil")
	}
}

func TestRemoveFile_MissingArg(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewRemoveFileFunction(ctx)

	_, err := fn(map[string]any{})
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

	_, err := fn(map[string]any{"0": "testdir"})
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

	_, err := fn(map[string]any{"0": "testdir"})
	if err == nil {
		t.Errorf("expected error for non-empty directory, got nil")
	}
}

func TestRemoveDir_Nonexistent(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewRemoveDirFunction(ctx)

	_, err := fn(map[string]any{"0": "nonexistent"})
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

	_, err := fn(map[string]any{"0": "old.txt", "1": "new.txt"})
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

	_, err := fn(map[string]any{"0": "test.txt", "1": "subdir/test.txt"})
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

	_, err := fn(map[string]any{"0": "nonexistent.txt", "1": "new.txt"})
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
			_, err := fn(tt.args)
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

	result, err := fn(map[string]any{"0": "test.txt"})
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

	result, err := fn(map[string]any{"0": "testdir"})
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

	_, err := fn(map[string]any{"0": "nonexistent"})
	if err == nil {
		t.Errorf("expected error for nonexistent path, got nil")
	}
}

func TestFileType_MissingArg(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(t, tempDir)

	ctx := FileIOContext{ScriptDir: tempDir}
	fn := NewFileTypeFunction(ctx)

	_, err := fn(map[string]any{})
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

	result, err := fn(map[string]any{"0": "test.txt"})
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

	result, err := fn(map[string]any{"0": "testdir"})
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

	result, err := fn(map[string]any{"0": "nonexistent"})
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

	_, err := fn(map[string]any{})
	if err == nil {
		t.Errorf("expected error for missing argument, got nil")
	}
}

// ============================================================================
// current_dir() TESTS
// ============================================================================

func TestCurrentDir(t *testing.T) {
	fn := NewCurrentDirFunction()

	result, err := fn(map[string]any{})
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

	_, err := fn(map[string]any{"0": "test.txt", "1": "hello"})
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

	_, err := fn(map[string]any{"0": "test.txt", "1": " world"})
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

	_, err := fn(map[string]any{"0": "a/b/test.txt", "1": "content"})
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
			_, err := fn(tt.args)
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

	_, err := fn(map[string]any{"0": "src.txt", "1": "dst.txt"})
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

	_, err := fn(map[string]any{"0": "src.txt", "1": "dir/a/b/dst.txt"})
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

	_, err := fn(map[string]any{"0": "src.txt", "1": "dst.txt"})
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

	_, err := fn(map[string]any{"0": "nonexistent.txt", "1": "dst.txt"})
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
			_, err := fn(tt.args)
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

	_, err := fn(map[string]any{"0": "src.txt", "1": "dst.txt"})
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

	_, err := fn(map[string]any{"0": "src.txt", "1": "dir/dst.txt"})
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

	_, err := fn(map[string]any{"0": "nonexistent.txt", "1": "dst.txt"})
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
			_, err := fn(tt.args)
			if err == nil {
				t.Errorf("expected error, got nil")
			}
		})
	}
}
