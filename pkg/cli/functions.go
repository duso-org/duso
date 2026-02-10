// Package cli provides CLI-specific functions for Duso scripts.
//
// These functions extend the core language with file I/O, environment access, and module loading.
// They are NOT part of the core language and are only available when using the duso CLI.
//
// Embedded Go applications can optionally register these functions if they wish,
// or implement their own versions with different behavior.
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/duso-org/duso/pkg/script"
)

// FileIOContext holds context for file I/O operations (script directory, etc.)
type FileIOContext struct {
	ScriptDir string
	NoFiles   bool // If true, restrict to /STORE/ and /EMBED/ only
}

// checkFilesAllowed checks if file operations are allowed for the given path.
// If NoFiles is enabled, only /STORE/ and /EMBED/ paths are allowed.
func (ctx *FileIOContext) checkFilesAllowed(path string) error {
	if !ctx.NoFiles {
		return nil // Files are allowed
	}

	// NoFiles is enabled - only allow /STORE/ and /EMBED/
	if strings.HasPrefix(path, "/STORE/") || strings.HasPrefix(path, "/EMBED/") {
		return nil
	}

	return fmt.Errorf("filesystem access disabled (use -no-files=false to enable)")
}

// ResolvePath resolves relative paths to scriptDir for file operations.
func (ctx *FileIOContext) ResolvePath(filespec string) string {
	if filepath.IsAbs(filespec) || strings.HasPrefix(filespec, "/") {
		return filespec
	}
	return filepath.Join(ctx.ScriptDir, filespec)
}

// isDatastorePath checks if a path is a datastore path (/namespace/key format).
// Returns (isDatastore, namespace, key).
// Special case: /STORE/ maps to "vfs" namespace.
func isDatastorePath(path string) (bool, string, string) {
	if !strings.HasPrefix(path, "/") {
		return false, "", ""
	}

	// Special case for /STORE/ (maps to "vfs" namespace)
	if strings.HasPrefix(path, "/STORE/") {
		key := strings.TrimPrefix(path, "/STORE/")
		return true, "vfs", key
	}

	// General /namespace/key format (e.g., /test_remove_store/file.txt)
	if strings.Count(path, "/") >= 2 {
		parts := strings.SplitN(path[1:], "/", 2) // Skip leading /
		if len(parts) == 2 {
			namespace := parts[0]
			key := parts[1]
			// Verify this looks like a datastore path (namespace contains no slashes)
			if !strings.Contains(namespace, "/") {
				return true, namespace, key
			}
		}
	}

	return false, "", ""
}

// Load and Save functions have been moved to pkg/runtime/builtin_files.go
// and are registered via capability injection in register.go

// Include and require functions have been moved to pkg/runtime/builtin_require.go
// and are registered via capability injection in register.go

// env() function has been moved to pkg/runtime/builtin_env.go
// and is registered via capability injection in register.go

// NewDocFunction creates a doc(name) function that displays documentation.
//
// doc() searches for documentation in this order:
// 1. Module documentation (.du files with matching .md, using require() resolution)
// 2. Reference documentation (docs/reference/*.md for builtins and CLI functions)
//
// It's only available in the CLI environment.
//
// Example:
//
//	docs = doc("http")      // Module docs
//	docs = doc("split")     // Builtin reference docs
//	print(markdown(docs))
//
// The function prints the full path to the documentation file before the content,
// which helps with debugging version issues.
// Returns nil if the documentation is not found.
func NewDocFunction(resolver *ModuleResolver) func(*script.Evaluator, map[string]any) (any, error) {
	return func(evaluator *script.Evaluator, args map[string]any) (any, error) {
		name, ok := args["0"].(string)
		if !ok {
			// Check for named argument "name"
			if n, ok := args["name"]; ok {
				name = fmt.Sprintf("%v", n)
			} else {
				// Default to index if no name provided
				name = "index"
			}
		}

		// First, try to find as a module (same resolution as require())
		fullPath, _, err := resolver.ResolveModule(name)
		if err == nil && fullPath != "" {
			// Convert .du extension to .md
			docPath := strings.TrimSuffix(fullPath, ".du") + ".md"
			content, err := readFile(docPath)
			if err == nil {
				output := fmt.Sprintf("Documentation from: %s\n\n%s", docPath, string(content))
				return output, nil
			}
		}

		// If not a module, try reference documentation using same resolution as require()
		searchPaths := []string{"."}
		searchPaths = append(searchPaths, resolver.DusoPath...)
		searchPaths = append(searchPaths, "/EMBED/")

		for _, basePath := range searchPaths {
			// Try docs/reference, stdlib/{name}, and contrib patterns
			candidates := []string{
				filepath.Join(basePath, "docs/reference", name+".md"),
				filepath.Join(basePath, "stdlib", name, name+".md"),
				filepath.Join(basePath, "contrib", name+".md"),
			}
			for _, docPath := range candidates {
				if content, err := readFile(docPath); err == nil {
					output := fmt.Sprintf("Documentation from: %s\n\n%s", docPath, string(content))
					return output, nil
				}
			}
		}

		// Not found anywhere
		return nil, nil
	}
}

// NewListDirFunction creates a list_dir(path) function that lists directory contents.
func NewListDirFunction(ctx FileIOContext) func(*script.Evaluator, map[string]any) (any, error) {
	return func(evaluator *script.Evaluator, args map[string]any) (any, error) {
		path, ok := args["0"].(string)
		if !ok {
			return nil, fmt.Errorf("list_dir() requires a path argument")
		}

		fullPath := ctx.ResolvePath(path)

		entries, err := os.ReadDir(fullPath)
		if err != nil {
			return nil, fmt.Errorf("cannot list directory '%s': %w", path, err)
		}

		result := make([]any, len(entries))
		for i, entry := range entries {
			result[i] = map[string]any{
				"name":   entry.Name(),
				"is_dir": entry.IsDir(),
			}
		}
		return result, nil
	}
}

// NewListFilesFunction creates a list_files(pattern) function that lists files matching a wildcard pattern.
// Supports wildcard patterns (* and ?). For plain directory listing, use list_dir().
func NewListFilesFunction(ctx FileIOContext) func(*script.Evaluator, map[string]any) (any, error) {
	return func(evaluator *script.Evaluator, args map[string]any) (any, error) {
		pattern, ok := args["0"].(string)
		if !ok {
			return nil, fmt.Errorf("list_files() requires a pattern argument")
		}

		// Resolve full pattern path using centralized resolution
		fullPattern := ctx.ResolvePath(pattern)

		// Expand glob pattern
		matches, err := expandGlob(fullPattern)
		if err != nil {
			return nil, err
		}

		// Convert to relative paths if input was relative
		if !filepath.IsAbs(pattern) && !strings.HasPrefix(pattern, "/") {
			for i, match := range matches {
				rel, err := filepath.Rel(ctx.ScriptDir, match)
				if err == nil {
					matches[i] = rel
				}
			}
		}

		// Convert to []any for Duso compatibility
		result := make([]any, len(matches))
		for i, path := range matches {
			result[i] = path
		}
		return result, nil
	}
}

// NewMakeDirFunction creates a make_dir(path) function that creates directories.
func NewMakeDirFunction(ctx FileIOContext) func(*script.Evaluator, map[string]any) (any, error) {
	return func(evaluator *script.Evaluator, args map[string]any) (any, error) {
		path, ok := args["0"].(string)
		if !ok {
			return nil, fmt.Errorf("make_dir() requires a path argument")
		}

		fullPath := ctx.ResolvePath(path)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			return nil, fmt.Errorf("cannot create directory '%s': %w", path, err)
		}
		return nil, nil
	}
}

// NewRemoveFileFunction creates a remove_file(path) function that deletes files.
//
// remove_file() removes a file or files matching a pattern. Supports:
// - Relative paths (relative to script directory)
// - Absolute paths
// - /STORE/ virtual filesystem paths
// - Wildcard patterns (* and ?)
//
// Example:
//
//	remove_file("temp.txt")
//	remove_file("logs/*.log")
//	remove_file("/STORE/generated.du")
func NewRemoveFileFunction(ctx FileIOContext) func(*script.Evaluator, map[string]any) (any, error) {
	return func(evaluator *script.Evaluator, args map[string]any) (any, error) {
		path, ok := args["0"].(string)
		if !ok {
			return nil, fmt.Errorf("remove_file() requires a path argument")
		}

		// Resolve the full path using centralized resolution
		fullPath := ctx.ResolvePath(path)

		// /EMBED/ is read-only - reject any remove attempts
		if strings.HasPrefix(fullPath, "/EMBED/") {
			return nil, fmt.Errorf("cannot write to /EMBED/: embedded filesystem is read-only")
		}

		// Check for wildcards in the path
		if hasWildcard(fullPath) {
			// Expand the pattern
			matches, err := expandGlob(fullPath)
			if err != nil {
				return nil, err
			}

			// Remove each matched file
			removed := []string{}
			for _, match := range matches {
				// Check permissions
				if err := ctx.checkFilesAllowed(match); err != nil {
					continue // Skip files that aren't allowed
				}

				// Try to remove the file
				var removeErr error
				if strings.HasPrefix(match, "/STORE/") {
					key := strings.TrimPrefix(match, "/STORE/")
					store := script.GetDatastore("vfs", nil)
					removeErr = store.Delete(key)
				} else {
					removeErr = os.Remove(match)
				}

				if removeErr == nil {
					// Success: add to results (use relative path if possible)
					resultPath := match
					if !filepath.IsAbs(path) && !strings.HasPrefix(path, "/") {
						if rel, err := filepath.Rel(ctx.ScriptDir, match); err == nil {
							resultPath = rel
						}
					}
					removed = append(removed, resultPath)
				}
				// Errors are silently skipped (per requirements)
			}

			// Convert to []any
			result := make([]any, len(removed))
			for i, p := range removed {
				result[i] = p
			}
			return result, nil
		}

		// No wildcards: single file remove
		// Check if file operations are allowed
		if err := ctx.checkFilesAllowed(fullPath); err != nil {
			return nil, err
		}

		// Handle /STORE/ paths differently
		if strings.HasPrefix(fullPath, "/STORE/") {
			key := strings.TrimPrefix(fullPath, "/STORE/")
			store := script.GetDatastore("vfs", nil)
			if err := store.Delete(key); err != nil {
				return nil, err
			}
			return []any{path}, nil
		}

		// Regular filesystem remove
		if err := os.Remove(fullPath); err != nil {
			return nil, fmt.Errorf("cannot remove file '%s': %w", path, err)
		}
		return []any{path}, nil
	}
}

// NewRemoveDirFunction creates a remove_dir(path) function that removes empty directories.
func NewRemoveDirFunction(ctx FileIOContext) func(*script.Evaluator, map[string]any) (any, error) {
	return func(evaluator *script.Evaluator, args map[string]any) (any, error) {
		path, ok := args["0"].(string)
		if !ok {
			return nil, fmt.Errorf("remove_dir() requires a path argument")
		}

		fullPath := ctx.ResolvePath(path)
		if err := os.Remove(fullPath); err != nil {
			return nil, fmt.Errorf("cannot remove directory '%s': %w", path, err)
		}
		return nil, nil
	}
}

// NewRenameFileFunction creates a rename_file(old, new) function.
func NewRenameFileFunction(ctx FileIOContext) func(*script.Evaluator, map[string]any) (any, error) {
	return func(evaluator *script.Evaluator, args map[string]any) (any, error) {
		oldPath, ok := args["0"].(string)
		if !ok {
			return nil, fmt.Errorf("rename_file() requires two path arguments")
		}

		newPath, ok := args["1"].(string)
		if !ok {
			return nil, fmt.Errorf("rename_file() requires two path arguments")
		}

		oldFull := ctx.ResolvePath(oldPath)
		newFull := ctx.ResolvePath(newPath)

		if err := os.Rename(oldFull, newFull); err != nil {
			return nil, fmt.Errorf("cannot rename '%s' to '%s': %w", oldPath, newPath, err)
		}
		return nil, nil
	}
}

// NewFileTypeFunction creates a file_type(path) function that returns file type.
func NewFileTypeFunction(ctx FileIOContext) func(*script.Evaluator, map[string]any) (any, error) {
	return func(evaluator *script.Evaluator, args map[string]any) (any, error) {
		path, ok := args["0"].(string)
		if !ok {
			return nil, fmt.Errorf("file_type() requires a path argument")
		}

		fullPath := ctx.ResolvePath(path)
		info, err := os.Stat(fullPath)
		if err != nil {
			return nil, fmt.Errorf("cannot stat '%s': %w", path, err)
		}

		if info.IsDir() {
			return "directory", nil
		}
		return "file", nil
	}
}

// NewFileExistsFunction creates a file_exists(path) function.
//
// file_exists() checks if a file exists. Supports:
// - Relative paths (relative to script directory)
// - Absolute paths
// - /STORE/ virtual filesystem paths
// - /EMBED/ embedded files
//
// Returns true if the file exists, false otherwise.
func NewFileExistsFunction(ctx FileIOContext) func(*script.Evaluator, map[string]any) (any, error) {
	return func(evaluator *script.Evaluator, args map[string]any) (any, error) {
		path, ok := args["0"].(string)
		if !ok {
			return nil, fmt.Errorf("file_exists() requires a path argument")
		}

		// Resolve the full path using centralized resolution
		fullPath := ctx.ResolvePath(path)

		// Check if file operations are allowed (for non-virtual paths)
		if err := ctx.checkFilesAllowed(fullPath); err != nil {
			return nil, err
		}

		return fileExists(fullPath), nil
	}
}

// NewCurrentDirFunction creates a current_dir() function that returns the working directory.
func NewCurrentDirFunction() func(*script.Evaluator, map[string]any) (any, error) {
	return func(evaluator *script.Evaluator, args map[string]any) (any, error) {
		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("cannot get current directory: %w", err)
		}
		return wd, nil
	}
}

// NewAppendFileFunction creates an append_file(path, content) function.
//
// append_file() appends content to a file. Supports:
// - Relative paths (relative to script directory)
// - Absolute paths
// - /STORE/ virtual filesystem paths
//
// Example:
//
//	append_file("log.txt", "New log entry\n")
//	append_file("/STORE/output.txt", result)
func NewAppendFileFunction(ctx FileIOContext) func(*script.Evaluator, map[string]any) (any, error) {
	return func(evaluator *script.Evaluator, args map[string]any) (any, error) {
		path, ok := args["0"].(string)
		if !ok {
			return nil, fmt.Errorf("append_file() requires path and content arguments")
		}

		content, ok := args["1"].(string)
		if !ok {
			if c, ok := args["content"]; ok {
				content = fmt.Sprintf("%v", c)
			} else {
				return nil, fmt.Errorf("append_file() requires path and content arguments")
			}
		}

		// Resolve the full path using centralized resolution
		fullPath := ctx.ResolvePath(path)

		// Check if file operations are allowed
		if err := ctx.checkFilesAllowed(fullPath); err != nil {
			return nil, err
		}

		// Handle /STORE/ paths differently
		if strings.HasPrefix(fullPath, "/STORE/") {
			return nil, appendToStore(fullPath, []byte(content))
		}

		// Regular filesystem append
		file, err := os.OpenFile(fullPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("cannot open '%s': %w", path, err)
		}
		defer file.Close()

		if _, err := file.WriteString(content); err != nil {
			return nil, fmt.Errorf("cannot append to '%s': %w", path, err)
		}
		return nil, nil
	}
}

// NewCopyFileFunction creates a copy_file(src, dst) function.
//
// copy_file() copies a file from source to destination. Supports:
// - /EMBED/ (read-only source)
// - /STORE/ (source and/or destination)
// - Regular filesystem
// - Relative paths (resolved via centralized path resolution)
// - Wildcard patterns in source (*, ?)
// - When source has wildcards, destination must be a directory
//
// Example:
//
//	copy_file("template.txt", "output.txt")
//	copy_file("src/*.ts", "dist/")
//	copy_file("/EMBED/stdlib/module.du", "/STORE/module.du")
func NewCopyFileFunction(ctx FileIOContext) func(*script.Evaluator, map[string]any) (any, error) {
	return func(evaluator *script.Evaluator, args map[string]any) (any, error) {
		src, ok := args["0"].(string)
		if !ok {
			return nil, fmt.Errorf("copy_file() requires source and destination arguments")
		}

		dst, ok := args["1"].(string)
		if !ok {
			return nil, fmt.Errorf("copy_file() requires source and destination arguments")
		}

		// Resolve paths using centralized resolution
		fullSrc := ctx.ResolvePath(src)
		fullDst := ctx.ResolvePath(dst)

		// Check for wildcards in source
		if hasWildcard(fullSrc) {
			// For wildcard operations, destination MUST be a directory
			// Special handling for /STORE/ (always valid) and /EMBED/ (read-only)
			if !strings.HasPrefix(fullDst, "/STORE/") && !strings.HasPrefix(fullDst, "/EMBED/") {
				info, err := os.Stat(fullDst)
				if err != nil || !info.IsDir() {
					return nil, fmt.Errorf("copy_file() with wildcard source requires destination to be an existing directory")
				}
			}

			// Expand the source pattern
			matches, err := expandGlob(fullSrc)
			if err != nil {
				return nil, err
			}

			// Copy each matched file
			copied := []string{}
			for _, match := range matches {
				// Read source file
				content, err := readFile(match)
				if err != nil {
					continue // Skip on read error
				}

				// Determine destination filename
				basename := filepath.Base(match)
				dstPath := filepath.Join(fullDst, basename)

				// Check if file operations are allowed
				if err := ctx.checkFilesAllowed(dstPath); err != nil {
					continue // Skip on permission error
				}

				// Write destination file
				if err := writeFile(dstPath, content, 0644); err == nil {
					// Success: add to results (use relative path if possible)
					resultPath := dstPath
					if !filepath.IsAbs(dst) && !strings.HasPrefix(dst, "/") {
						if rel, err := filepath.Rel(ctx.ScriptDir, dstPath); err == nil {
							resultPath = rel
						}
					}
					copied = append(copied, resultPath)
				}
				// Errors are silently skipped (per requirements)
			}

			// Convert to []any
			result := make([]any, len(copied))
			for i, p := range copied {
				result[i] = p
			}
			return result, nil
		}

		// No wildcards: single file copy
		// Check if file operations are allowed (for destination if not virtual)
		if err := ctx.checkFilesAllowed(fullDst); err != nil {
			return nil, err
		}

		// Support reading from /EMBED/ and /STORE/
		content, err := readFile(fullSrc)
		if err != nil {
			return nil, fmt.Errorf("cannot read '%s': %w", src, err)
		}

		// Create parent directories (not needed for /STORE/)
		if !strings.HasPrefix(fullDst, "/STORE/") && !strings.HasPrefix(fullDst, "/EMBED/") {
			if err := os.MkdirAll(filepath.Dir(fullDst), 0755); err != nil {
				return nil, fmt.Errorf("cannot create directory: %w", err)
			}
		}

		if err := writeFile(fullDst, content, 0644); err != nil {
			return nil, fmt.Errorf("cannot write to '%s': %w", dst, err)
		}
		return []any{dst}, nil
	}
}

// NewMoveFileFunction creates a move_file(src, dst) function.
//
// move_file() moves (renames) a file from source to destination. Supports:
// - Relative paths (resolved via centralized path resolution)
// - Absolute paths
// - /STORE/ virtual filesystem paths
// - Wildcard patterns in source (*, ?)
// - When source has wildcards, destination must be a directory
//
// Example:
//
//	move_file("old.txt", "new.txt")
//	move_file("old_*.txt", "archive/")
//	move_file("/STORE/temp.du", "/STORE/final.du")
func NewMoveFileFunction(ctx FileIOContext) func(*script.Evaluator, map[string]any) (any, error) {
	return func(evaluator *script.Evaluator, args map[string]any) (any, error) {
		src, ok := args["0"].(string)
		if !ok {
			return nil, fmt.Errorf("move_file() requires source and destination arguments")
		}

		dst, ok := args["1"].(string)
		if !ok {
			return nil, fmt.Errorf("move_file() requires source and destination arguments")
		}

		// Resolve paths using centralized resolution
		fullSrc := ctx.ResolvePath(src)
		fullDst := ctx.ResolvePath(dst)

		// /EMBED/ is read-only - reject any move attempts
		if strings.HasPrefix(fullSrc, "/EMBED/") {
			return nil, fmt.Errorf("cannot write to /EMBED/: embedded filesystem is read-only")
		}

		// Check for wildcards in source
		if hasWildcard(fullSrc) {
			// For wildcard operations, destination MUST be a directory
			info, err := os.Stat(fullDst)
			if err != nil || !info.IsDir() {
				return nil, fmt.Errorf("move_file() with wildcard source requires destination to be an existing directory")
			}

			// Expand the source pattern
			matches, err := expandGlob(fullSrc)
			if err != nil {
				return nil, err
			}

			// Move each matched file
			moved := []string{}
			for _, match := range matches {
				// Determine destination filename
				basename := filepath.Base(match)
				dstPath := filepath.Join(fullDst, basename)

				// Check if file operations are allowed
				if err := ctx.checkFilesAllowed(dstPath); err != nil {
					continue // Skip on permission error
				}

				// Move the file (for /STORE/, this is copy+delete)
				var moveErr error
				if strings.HasPrefix(match, "/STORE/") {
					// Read from /STORE/
					content, err := readFile(match)
					if err != nil {
						continue // Skip on read error
					}

					// Write to destination
					if err := writeFile(dstPath, content, 0644); err != nil {
						continue // Skip on write error
					}

					// Delete from /STORE/
					srcKey := strings.TrimPrefix(match, "/STORE/")
					store := script.GetDatastore("vfs", nil)
					moveErr = store.Delete(srcKey)
				} else {
					// Regular filesystem move
					moveErr = os.Rename(match, dstPath)
				}

				if moveErr == nil {
					// Success: add to results (use relative path if possible)
					resultPath := dstPath
					if !filepath.IsAbs(dst) && !strings.HasPrefix(dst, "/") {
						if rel, err := filepath.Rel(ctx.ScriptDir, dstPath); err == nil {
							resultPath = rel
						}
					}
					moved = append(moved, resultPath)
				}
				// Errors are silently skipped (per requirements)
			}

			// Convert to []any
			result := make([]any, len(moved))
			for i, p := range moved {
				result[i] = p
			}
			return result, nil
		}

		// No wildcards: single file move
		// Check if file operations are allowed
		if err := ctx.checkFilesAllowed(fullDst); err != nil {
			return nil, err
		}

		// Handle /STORE/ source paths differently (copy from store, write to dest, delete from store)
		if strings.HasPrefix(fullSrc, "/STORE/") {
			// Read from /STORE/
			content, err := readFile(fullSrc)
			if err != nil {
				return nil, fmt.Errorf("cannot read '%s': %w", src, err)
			}

			// Write to destination
			if err := writeFile(fullDst, content, 0644); err != nil {
				return nil, fmt.Errorf("cannot write to '%s': %w", dst, err)
			}

			// Delete from /STORE/
			srcKey := strings.TrimPrefix(fullSrc, "/STORE/")
			store := script.GetDatastore("vfs", nil)
			if err := store.Delete(srcKey); err != nil {
				return nil, err
			}

			return []any{dst}, nil
		}

		// Regular filesystem move
		if err := os.Rename(fullSrc, fullDst); err != nil {
			return nil, fmt.Errorf("cannot move '%s' to '%s': %w", src, dst, err)
		}
		return []any{dst}, nil
	}
}
