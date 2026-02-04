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

// NewLoadFunction creates a load(filename) function that reads files.
//
// load() reads the contents of a file. Supports:
// - Relative paths (relative to script directory)
// - Absolute paths
// - /STORE/ virtual filesystem paths
// - /EMBED/ embedded files
//
// It's only available in the CLI environment.
//
// Example:
//
//	content = load("data.txt")
//	data = parse_json(load("config.json"))
//	code = load("/STORE/generated.du")  // Load from virtual filesystem
func NewLoadFunction(ctx FileIOContext) func(map[string]any) (any, error) {
	return func(args map[string]any) (any, error) {
		filename, ok := args["0"].(string)
		if !ok {
			// Check for named argument "filename"
			if f, ok := args["filename"]; ok {
				filename = fmt.Sprintf("%v", f)
			} else {
				return nil, fmt.Errorf("load() requires a filename argument")
			}
		}

		// Determine the full path to check permissions
		var fullPath string
		if filepath.IsAbs(filename) || strings.HasPrefix(filename, "/") {
			fullPath = filename
		} else {
			fullPath = filepath.Join(ctx.ScriptDir, filename)
		}

		// Check if file operations are allowed
		if err := ctx.checkFilesAllowed(fullPath); err != nil {
			return nil, err
		}

		// Try to load as specified first (supports /STORE/, /EMBED/, absolute, home paths)
		content, err := readFile(filename)
		if err != nil {
			// Fallback: try with script directory prepended (for relative paths)
			fallbackPath := filepath.Join(ctx.ScriptDir, filename)
			content, err = readFile(fallbackPath)
			if err != nil {
				return nil, fmt.Errorf("cannot load '%s': %w", filename, err)
			}
		}

		return string(content), nil
	}
}

// NewSaveFunction creates a save(filename, content) function that writes files.
//
// save() writes content to a file. Supports:
// - Relative paths (relative to script directory)
// - Absolute paths
// - /STORE/ virtual filesystem paths
// - /EMBED/ paths (read-only, will error)
//
// It's only available in the CLI environment.
//
// Example:
//
//	save("output.txt", "Hello, World!")
//	save("data.json", format_json(myObject))
//	save("/STORE/generated.du", code)  // Save to virtual filesystem
func NewSaveFunction(ctx FileIOContext) func(map[string]any) (any, error) {
	return func(args map[string]any) (any, error) {
		filename, ok := args["0"].(string)
		if !ok {
			// Check for named argument "filename"
			if f, ok := args["filename"]; ok {
				filename = fmt.Sprintf("%v", f)
			} else {
				return nil, fmt.Errorf("save() requires filename and content arguments")
			}
		}

		content, ok := args["1"].(string)
		if !ok {
			// Check for named argument "content"
			if c, ok := args["content"]; ok {
				content = fmt.Sprintf("%v", c)
			} else {
				return nil, fmt.Errorf("save() requires filename and content arguments")
			}
		}

		// Determine the full path
		var fullPath string
		if filepath.IsAbs(filename) || strings.HasPrefix(filename, "/") {
			// Absolute path or virtual filesystem
			fullPath = filename
		} else {
			// Relative path - prefix with script directory
			fullPath = filepath.Join(ctx.ScriptDir, filename)
		}

		// Check if file operations are allowed
		if err := ctx.checkFilesAllowed(fullPath); err != nil {
			return nil, err
		}

		// For /STORE/, don't create parent directories (they're implicit)
		// For regular filesystem, create parent directories if needed
		if !strings.HasPrefix(fullPath, "/STORE/") && !strings.HasPrefix(fullPath, "/EMBED/") {
			if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
				return nil, fmt.Errorf("cannot create directory: %w", err)
			}
		}

		err := writeFile(fullPath, []byte(content), 0644)
		if err != nil {
			return nil, fmt.Errorf("cannot save to '%s': %w", filename, err)
		}

		return nil, nil
	}
}

// NewIncludeFunction creates an include(filename) function that executes other scripts.
//
// include() loads and executes another .du script file in the current environment.
// Variables and functions defined in the included script are available after include().
// It's only available in the CLI environment.
//
// Unlike require(), include() executes in the current scope (not isolated),
// and results are not cached. However, the AST is cached globally with mtime validation
// for efficient reloading during development.
//
// Example:
//
//	include("helpers.du")
//	result = helper_function()  // Now available
//
// This function supports path resolution: user-provided paths, relative to script dir, and DUSO_LIB.
func NewIncludeFunction(resolver *ModuleResolver, detector *CircularDetector, interp *script.Interpreter) func(map[string]any) (any, error) {
	return func(args map[string]any) (any, error) {
		filename, ok := args["0"].(string)
		if !ok {
			// Check for named argument "filename"
			if f, ok := args["filename"]; ok {
				filename = fmt.Sprintf("%v", f)
			} else {
				return nil, fmt.Errorf("include() requires a filename argument")
			}
		}

		// Resolve module path using standard resolution algorithm
		fullPath, searchedPaths, err := resolver.ResolveModule(filename)
		if err != nil {
			return nil, fmt.Errorf("cannot find module '%s'\nSearched:\n  %s",
				filename, strings.Join(searchedPaths, "\n  "))
		}

		// Check for circular dependency
		if err := detector.Push(fullPath); err != nil {
			return nil, err
		}
		defer detector.Pop()

		// Set file path context for error reporting
		prevPath := interp.GetFilePath()
		interp.SetFilePath(fullPath)
		defer interp.SetFilePath(prevPath)

		// Parse script file (AST is cached with mtime validation)
		program, err := interp.ParseScriptFile(fullPath, readFile, getFileMtime)
		if err != nil {
			return nil, fmt.Errorf("cannot include '%s': %w", fullPath, err)
		}

		// Execute in current environment (no isolation)
		_, err = interp.EvalProgram(program)
		if err != nil {
			return nil, fmt.Errorf("error in included script '%s': %w", fullPath, err)
		}

		return nil, nil
	}
}

// NewRequireFunction creates a require(moduleName) function that loads modules.
//
// require() loads a module in an isolated scope and returns its exports.
// Unlike include(), require():
// - Executes the module in its own isolated scope
// - Returns the last expression value (the module's exports)
// - Caches results - subsequent requires return cached value without re-executing
//
// The AST is cached globally with mtime validation for hot reload during development.
// The module result is cached per-interpreter to allow concurrent evaluators to get
// fresh module instances while reusing the parsed AST.
//
// Example:
//
//	math = require("math")
//	result = math.add(2, 3)  // Calls function from module
//
// This function supports path resolution: user-provided paths, relative to script dir, and DUSO_LIB.
func NewRequireFunction(resolver *ModuleResolver, detector *CircularDetector, interp *script.Interpreter) func(map[string]any) (any, error) {
	return func(args map[string]any) (any, error) {
		filename, ok := args["0"].(string)
		if !ok {
			// Check for named argument "filename"
			if f, ok := args["filename"]; ok {
				filename = fmt.Sprintf("%v", f)
			} else {
				return nil, fmt.Errorf("require() requires a filename argument")
			}
		}

		// Resolve module path using standard resolution algorithm
		fullPath, searchedPaths, err := resolver.ResolveModule(filename)
		if err != nil {
			return nil, fmt.Errorf("cannot find module '%s'\nSearched:\n  %s",
				filename, strings.Join(searchedPaths, "\n  "))
		}

		// Check module cache (absolute path as key)
		// This caches the result value, not the AST
		if cached, ok := interp.GetModuleCache(fullPath); ok {
			return script.ValueToInterface(cached), nil
		}

		// Check for circular dependency
		if err := detector.Push(fullPath); err != nil {
			return nil, err
		}
		defer detector.Pop()

		// Set file path context for error reporting
		prevPath := interp.GetFilePath()
		interp.SetFilePath(fullPath)
		defer interp.SetFilePath(prevPath)

		// Parse script file (AST is cached with mtime validation)
		program, err := interp.ParseScriptFile(fullPath, readFile, getFileMtime)
		if err != nil {
			return nil, fmt.Errorf("cannot require '%s': %w", fullPath, err)
		}

		// Execute in isolated scope using ExecuteModuleProgram to reuse evaluator logic
		value, err := interp.ExecuteModuleProgram(program)
		if err != nil {
			return nil, fmt.Errorf("error in module '%s': %w", fullPath, err)
		}

		// Cache the result
		interp.SetModuleCache(fullPath, value)

		// Convert to interface{} for return
		return script.ValueToInterface(value), nil
	}
}

// NewEnvFunction creates an env(varname) function that reads environment variables.
//
// env() reads the value of an environment variable from the OS.
// It's only available in the CLI environment (not in embedded contexts without explicit opt-in).
//
// Example:
//
//	key = env("ANTHROPIC_API_KEY")
//	debug = env("DEBUG_MODE")
//
// Returns the value as a string, or empty string if the variable is not set.
func NewEnvFunction() func(map[string]any) (any, error) {
	return func(args map[string]any) (any, error) {
		varname, ok := args["0"].(string)
		if !ok {
			// Check for named argument "varname"
			if v, ok := args["varname"]; ok {
				varname = fmt.Sprintf("%v", v)
			} else {
				return nil, fmt.Errorf("env() requires a variable name argument")
			}
		}

		return os.Getenv(varname), nil
	}
}

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
func NewDocFunction(resolver *ModuleResolver) func(map[string]any) (any, error) {
	return func(args map[string]any) (any, error) {
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
func NewListDirFunction(ctx FileIOContext) func(map[string]any) (any, error) {
	return func(args map[string]any) (any, error) {
		path, ok := args["0"].(string)
		if !ok {
			return nil, fmt.Errorf("list_dir() requires a path argument")
		}

		fullPath := filepath.Join(ctx.ScriptDir, path)

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
func NewListFilesFunction(ctx FileIOContext) func(map[string]any) (any, error) {
	return func(args map[string]any) (any, error) {
		pattern, ok := args["0"].(string)
		if !ok {
			return nil, fmt.Errorf("list_files() requires a pattern argument")
		}

		// Resolve full pattern path
		var fullPattern string
		if filepath.IsAbs(pattern) || strings.HasPrefix(pattern, "/") {
			fullPattern = pattern
		} else {
			fullPattern = filepath.Join(ctx.ScriptDir, pattern)
		}

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
func NewMakeDirFunction(ctx FileIOContext) func(map[string]any) (any, error) {
	return func(args map[string]any) (any, error) {
		path, ok := args["0"].(string)
		if !ok {
			return nil, fmt.Errorf("make_dir() requires a path argument")
		}

		fullPath := filepath.Join(ctx.ScriptDir, path)
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
func NewRemoveFileFunction(ctx FileIOContext) func(map[string]any) (any, error) {
	return func(args map[string]any) (any, error) {
		path, ok := args["0"].(string)
		if !ok {
			return nil, fmt.Errorf("remove_file() requires a path argument")
		}

		// Determine the full path
		var fullPath string
		if filepath.IsAbs(path) || strings.HasPrefix(path, "/") {
			fullPath = path
		} else {
			fullPath = filepath.Join(ctx.ScriptDir, path)
		}

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
func NewRemoveDirFunction(ctx FileIOContext) func(map[string]any) (any, error) {
	return func(args map[string]any) (any, error) {
		path, ok := args["0"].(string)
		if !ok {
			return nil, fmt.Errorf("remove_dir() requires a path argument")
		}

		fullPath := filepath.Join(ctx.ScriptDir, path)
		if err := os.Remove(fullPath); err != nil {
			return nil, fmt.Errorf("cannot remove directory '%s': %w", path, err)
		}
		return nil, nil
	}
}

// NewRenameFileFunction creates a rename_file(old, new) function.
func NewRenameFileFunction(ctx FileIOContext) func(map[string]any) (any, error) {
	return func(args map[string]any) (any, error) {
		oldPath, ok := args["0"].(string)
		if !ok {
			return nil, fmt.Errorf("rename_file() requires two path arguments")
		}

		newPath, ok := args["1"].(string)
		if !ok {
			return nil, fmt.Errorf("rename_file() requires two path arguments")
		}

		oldFull := filepath.Join(ctx.ScriptDir, oldPath)
		newFull := filepath.Join(ctx.ScriptDir, newPath)

		if err := os.Rename(oldFull, newFull); err != nil {
			return nil, fmt.Errorf("cannot rename '%s' to '%s': %w", oldPath, newPath, err)
		}
		return nil, nil
	}
}

// NewFileTypeFunction creates a file_type(path) function that returns file type.
func NewFileTypeFunction(ctx FileIOContext) func(map[string]any) (any, error) {
	return func(args map[string]any) (any, error) {
		path, ok := args["0"].(string)
		if !ok {
			return nil, fmt.Errorf("file_type() requires a path argument")
		}

		fullPath := filepath.Join(ctx.ScriptDir, path)
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
func NewFileExistsFunction(ctx FileIOContext) func(map[string]any) (any, error) {
	return func(args map[string]any) (any, error) {
		path, ok := args["0"].(string)
		if !ok {
			return nil, fmt.Errorf("file_exists() requires a path argument")
		}

		// Determine the full path
		var fullPath string
		if filepath.IsAbs(path) || strings.HasPrefix(path, "/") {
			fullPath = path
		} else {
			fullPath = filepath.Join(ctx.ScriptDir, path)
		}

		// Check if file operations are allowed (for non-virtual paths)
		if err := ctx.checkFilesAllowed(fullPath); err != nil {
			return nil, err
		}

		return fileExists(fullPath), nil
	}
}

// NewCurrentDirFunction creates a current_dir() function that returns the working directory.
func NewCurrentDirFunction() func(map[string]any) (any, error) {
	return func(args map[string]any) (any, error) {
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
func NewAppendFileFunction(ctx FileIOContext) func(map[string]any) (any, error) {
	return func(args map[string]any) (any, error) {
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

		// Determine the full path
		var fullPath string
		if filepath.IsAbs(path) || strings.HasPrefix(path, "/") {
			fullPath = path
		} else {
			fullPath = filepath.Join(ctx.ScriptDir, path)
		}

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
// - Relative paths (relative to script directory)
// - Wildcard patterns in source (*, ?)
// - When source has wildcards, destination must be a directory
//
// Example:
//
//	copy_file("template.txt", "output.txt")
//	copy_file("src/*.ts", "dist/")
//	copy_file("/EMBED/stdlib/module.du", "/STORE/module.du")
func NewCopyFileFunction(ctx FileIOContext) func(map[string]any) (any, error) {
	return func(args map[string]any) (any, error) {
		src, ok := args["0"].(string)
		if !ok {
			return nil, fmt.Errorf("copy_file() requires source and destination arguments")
		}

		dst, ok := args["1"].(string)
		if !ok {
			return nil, fmt.Errorf("copy_file() requires source and destination arguments")
		}

		// Determine the full source path
		var fullSrc string
		if filepath.IsAbs(src) || strings.HasPrefix(src, "/") {
			fullSrc = src
		} else {
			fullSrc = filepath.Join(ctx.ScriptDir, src)
		}

		// Determine the full destination path
		var fullDst string
		if filepath.IsAbs(dst) || strings.HasPrefix(dst, "/") {
			fullDst = dst
		} else {
			fullDst = filepath.Join(ctx.ScriptDir, dst)
		}

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
// - Relative paths (relative to script directory)
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
func NewMoveFileFunction(ctx FileIOContext) func(map[string]any) (any, error) {
	return func(args map[string]any) (any, error) {
		src, ok := args["0"].(string)
		if !ok {
			return nil, fmt.Errorf("move_file() requires source and destination arguments")
		}

		dst, ok := args["1"].(string)
		if !ok {
			return nil, fmt.Errorf("move_file() requires source and destination arguments")
		}

		// Determine the full source path
		var fullSrc string
		if filepath.IsAbs(src) || strings.HasPrefix(src, "/") {
			fullSrc = src
		} else {
			fullSrc = filepath.Join(ctx.ScriptDir, src)
		}

		// Determine the full destination path
		var fullDst string
		if filepath.IsAbs(dst) || strings.HasPrefix(dst, "/") {
			fullDst = dst
		} else {
			fullDst = filepath.Join(ctx.ScriptDir, dst)
		}

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
