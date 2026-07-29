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
	"strings"

	"github.com/duso-org/duso/pkg/core"
	"github.com/duso-org/duso/pkg/runtime"
	"github.com/duso-org/duso/pkg/script"
)

// checkFilesAllowed enforces the -no-files sandbox. When enabled, only
// /STORE/ and /EMBED/ paths are accepted; disk paths are rejected.
func checkFilesAllowed(path string) error {
	sysDs := runtime.GetDatastore("duso_sys", nil)
	noFilesVal, _ := sysDs.Get("-no-files")
	noFiles := false
	if noFilesVal != nil {
		if b, ok := noFilesVal.(bool); ok {
			noFiles = b
		}
	}

	if !noFiles {
		return nil
	}

	if core.HasPathPrefix(path, "STORE") || core.HasPathPrefix(path, "EMBED") {
		return nil
	}

	return fmt.Errorf("filesystem access disabled (use -no-files to enable)")
}

// appDir returns the entry-script's directory (or "" if unset). Used to
// produce caller-friendly relative paths in list/copy/move return values.
func appDir() string {
	if globalInterpreter == nil {
		return ""
	}
	return globalInterpreter.GetScriptDir()
}

// isDatastorePath checks if a path is a datastore path (/namespace/key format).
// Returns (isDatastore, namespace, key).
// Special case: /STORE/ maps to "vfs" namespace.
func isDatastorePath(path string) (bool, string, string) {
	if !core.IsAbsoluteOrSpecial(path) {
		return false, "", ""
	}

	// Special case for /STORE/ (maps to "vfs" namespace)
	if core.HasPathPrefix(path, "STORE") {
		key := core.TrimPathPrefix(path, "STORE")
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

// builtinDoc displays documentation.
// TODO: Needs ModuleResolver - convert to use RequestContext or pass via closure later
func builtinDoc(evaluator *script.Evaluator, args map[string]any) (any, error) {
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
		if globalResolver == nil {
		searchPaths := []string{"."}
		searchPaths = append(searchPaths, "/EMBED/")

		for _, basePath := range searchPaths {
			candidates := []string{
				core.Join(basePath, "docs/reference", name+".md"),
				core.Join(basePath, "stdlib", name, name+".md"),
				core.Join(basePath, "contrib", name+".md"),
			}
			for _, docPath := range candidates {
				if content, err := readFile(docPath); err == nil {
					output := fmt.Sprintf("Documentation from: %s\n\n%s", docPath, string(content))
					return output, nil
				}
			}
		}
		return nil, nil
	}

	fullPath, _, err := globalResolver.ResolveModule(name)
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
		searchPaths = append(searchPaths, globalResolver.DusoPath...)
		searchPaths = append(searchPaths, "/EMBED/")

		for _, basePath := range searchPaths {
			// Try docs/reference, stdlib/{name}, and contrib patterns
			candidates := []string{
				core.Join(basePath, "docs/reference", name+".md"),
				core.Join(basePath, "stdlib", name, name+".md"),
				core.Join(basePath, "contrib", name+".md"),
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

// builtinListDir lists directory contents. Path is resolved via ResolvePath
// (bare → appDir; /HERE/, /CWD/, /EMBED/, /STORE/, absolute as documented).
func builtinListDir(evaluator *script.Evaluator, args map[string]any) (any, error) {
	path, ok := args["0"].(string)
	if !ok {
		return nil, fmt.Errorf("list_dir() requires a path argument")
	}

	resolved := ResolvePath(path)
	entries, err := ListDirVFS(resolved)
	if err != nil {
		return nil, fmt.Errorf("cannot list directory '%s': %s", path, describeFileError(err, resolved))
	}

	result := make([]any, len(entries))
	for i, entry := range entries {
		result[i] = entry
	}
	return result, nil
}

// builtinListFiles lists files matching a wildcard pattern. Pattern is
// resolved via ResolvePath; results from a bare pattern come back relative
// to appDir so they round-trip cleanly into other file builtins.
func builtinListFiles(evaluator *script.Evaluator, args map[string]any) (any, error) {
	pattern, ok := args["0"].(string)
	if !ok {
		return nil, fmt.Errorf("list_files() requires a pattern argument")
	}

	resolved := ResolvePath(pattern)
	matches, err := ExpandGlob(resolved)
	if err != nil {
		return nil, err
	}

	// Bare input → re-relativize matches against appDir so callers see paths
	// in the same shape they wrote. Explicit prefixes / absolute paths stay
	// as-is so the result keeps the caller's intent visible.
	if !core.IsAbsoluteOrSpecial(pattern) {
		base := appDir()
		if base != "" {
			for i, match := range matches {
				if rel, err := core.Rel(base, match); err == nil {
					matches[i] = rel
				}
			}
		}
	}

	result := make([]any, len(matches))
	for i, path := range matches {
		result[i] = path
	}
	return result, nil
}

// builtinMakeDir creates directories. Path is resolved via ResolvePath.
func builtinMakeDir(evaluator *script.Evaluator, args map[string]any) (any, error) {
	path, ok := args["0"].(string)
	if !ok {
		return nil, fmt.Errorf("make_dir() requires a path argument")
	}

	resolved := ResolvePath(path)

	// /STORE/ is a virtual filesystem backed by datastore; directories don't need explicit creation
	// Just mark the directory exists by creating a marker entry if needed
	if core.HasPathPrefix(resolved, "STORE") {
		dirMarker := core.TrimPathPrefix(resolved, "STORE") + "/.dir"
		store := runtime.GetDatastore("duso_vfs", nil)
		return nil, store.Set(dirMarker, "")
	}

	if core.HasPathPrefix(resolved, "EMBED") {
		return nil, fmt.Errorf("cannot write to /EMBED/: embedded filesystem is read-only")
	}

	if err := os.MkdirAll(resolved, 0755); err != nil {
		return nil, fmt.Errorf("cannot create directory '%s': %s", path, describeFileError(err, resolved))
	}
	return nil, nil
}

// builtinRemoveFile deletes files matching a pattern. Path is resolved via
// ResolvePath; /EMBED/ is rejected (read-only).
func builtinRemoveFile(evaluator *script.Evaluator, args map[string]any) (any, error) {
	path, ok := args["0"].(string)
	if !ok {
		return nil, fmt.Errorf("remove_file() requires a path argument")
	}

	fullPath := ResolvePath(path)

	if core.HasPathPrefix(fullPath, "EMBED") {
		return nil, fmt.Errorf("cannot write to /EMBED/: embedded filesystem is read-only")
	}

	// Wildcard expansion: walk matches and skip silently on per-file errors.
	if hasWildcard(fullPath) {
		matches, err := ExpandGlob(fullPath)
		if err != nil {
			return nil, err
		}

		noFiles := GetSysFlag("-no-files", false)
		base := appDir()
		bareInput := !core.IsAbsoluteOrSpecial(path)

		removed := []string{}
		for _, match := range matches {
			if noFiles && !core.HasPathPrefix(match, "STORE") && !core.HasPathPrefix(match, "EMBED") {
				continue
			}

			var removeErr error
			if core.HasPathPrefix(match, "STORE") {
				key := core.TrimPathPrefix(match, "STORE")
				store := runtime.GetDatastore("duso_vfs", nil)
				_, removeErr = store.Delete(key)
			} else {
				removeErr = os.Remove(match)
			}

			if removeErr == nil {
				resultPath := match
				if bareInput && base != "" {
					if rel, err := core.Rel(base, match); err == nil {
						resultPath = rel
					}
				}
				removed = append(removed, resultPath)
			}
		}

		result := make([]any, len(removed))
		for i, p := range removed {
			result[i] = p
		}
		return result, nil
	}

	if err := checkFilesAllowed(fullPath); err != nil {
		return nil, err
	}

	if core.HasPathPrefix(fullPath, "STORE") {
		key := core.TrimPathPrefix(fullPath, "STORE")
		store := runtime.GetDatastore("duso_vfs", nil)
		if _, err := store.Delete(key); err != nil {
			return nil, fmt.Errorf("cannot remove file '%s': %s", path, describeFileError(err, fullPath))
		}
		return []any{path}, nil
	}

	if err := os.Remove(fullPath); err != nil {
		return nil, fmt.Errorf("cannot remove file '%s': %s", path, describeFileError(err, fullPath))
	}
	return []any{path}, nil
}

// builtinRemoveDir removes empty directories. Path is resolved via ResolvePath.
func builtinRemoveDir(evaluator *script.Evaluator, args map[string]any) (any, error) {
	path, ok := args["0"].(string)
	if !ok {
		return nil, fmt.Errorf("remove_dir() requires a path argument")
	}

	resolved := ResolvePath(path)

	// /STORE/ is a virtual filesystem; remove the directory marker if it exists
	if core.HasPathPrefix(resolved, "STORE") {
		dirMarker := core.TrimPathPrefix(resolved, "STORE") + "/.dir"
		store := runtime.GetDatastore("duso_vfs", nil)
		_, _ = store.Delete(dirMarker) // Ignore error if marker doesn't exist
		return nil, nil
	}

	if core.HasPathPrefix(resolved, "EMBED") {
		return nil, fmt.Errorf("cannot write to /EMBED/: embedded filesystem is read-only")
	}

	if err := os.Remove(resolved); err != nil {
		return nil, fmt.Errorf("cannot remove directory '%s': %s", path, describeFileError(err, resolved))
	}
	return nil, nil
}

// builtinRenameFile renames a file. Both paths resolved via ResolvePath.
func builtinRenameFile(evaluator *script.Evaluator, args map[string]any) (any, error) {
	oldPath, ok := args["0"].(string)
	if !ok {
		return nil, fmt.Errorf("rename_file() requires two path arguments")
	}

	newPath, ok := args["1"].(string)
	if !ok {
		return nil, fmt.Errorf("rename_file() requires two path arguments")
	}

	oldFull := ResolvePath(oldPath)
	newFull := ResolvePath(newPath)

	if err := os.Rename(oldFull, newFull); err != nil {
		return nil, fmt.Errorf("cannot rename '%s' to '%s': %s", oldPath, newPath, describeFileError(err, oldFull))
	}
	return nil, nil
}

// builtinFileType returns file type. Path is resolved via ResolvePath.
func builtinFileType(evaluator *script.Evaluator, args map[string]any) (any, error) {
	path, ok := args["0"].(string)
	if !ok {
		return nil, fmt.Errorf("file_type() requires a path argument")
	}

	resolved := ResolvePath(path)

	// /STORE/ is a virtual filesystem; check for directory marker or regular file
	if core.HasPathPrefix(resolved, "STORE") {
		key := core.TrimPathPrefix(resolved, "STORE")
		store := runtime.GetDatastore("duso_vfs", nil)

		// Check if it's a directory (has .dir marker)
		dirMarker := key + "/.dir"
		if val, _ := store.Get(dirMarker); val != nil {
			return "directory", nil
		}

		// Check if it's a regular file
		if val, _ := store.Get(key); val != nil {
			return "file", nil
		}

		return nil, fmt.Errorf("cannot stat '%s': %s", path, describeFileError(fmt.Errorf("no such file"), resolved))
	}

	// /EMBED/ paths
	if core.HasPathPrefix(resolved, "EMBED") {
		embeddedPath := core.TrimPathPrefix(resolved, "EMBED")
		info, err := EmbeddedStat(embeddedPath)
		if err != nil {
			return nil, fmt.Errorf("cannot stat '%s': %s", path, describeFileError(err, resolved))
		}
		if info.IsDir() {
			return "directory", nil
		}
		return "file", nil
	}

	// Regular filesystem
	info, err := os.Stat(resolved)
	if err != nil {
		return nil, fmt.Errorf("cannot stat '%s': %s", path, describeFileError(err, resolved))
	}

	if info.IsDir() {
		return "directory", nil
	}
	return "file", nil
}

// builtinFileExists checks if a file exists. Path is resolved via ResolvePath.
func builtinFileExists(evaluator *script.Evaluator, args map[string]any) (any, error) {
	path, ok := args["0"].(string)
	if !ok {
		return nil, fmt.Errorf("file_exists() requires a path argument")
	}

	return fileExists(ResolvePath(path)), nil
}

// builtinCurrentDir returns the working directory.
func builtinCurrentDir(evaluator *script.Evaluator, args map[string]any) (any, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("cannot get current directory: %w", err)
	}
	return wd, nil
}

// builtinAppendFile appends content to a file.
func builtinAppendFile(evaluator *script.Evaluator, args map[string]any) (any, error) {
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

	resolved := ResolvePath(path)

	if core.HasPathPrefix(resolved, "STORE") {
		if err := appendToStore(resolved, []byte(content)); err != nil {
			return nil, fmt.Errorf("cannot append to '%s': %s", path, describeFileError(err, resolved))
		}
		return nil, nil
	}

	file, err := os.OpenFile(resolved, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("cannot open '%s': %s", path, describeFileError(err, resolved))
	}
	defer file.Close()

	if _, err := file.WriteString(content); err != nil {
		return nil, fmt.Errorf("cannot append to '%s': %s", path, describeFileError(err, resolved))
	}
	return nil, nil
}

// builtinCopyFile copies a file from source to destination. Both paths are
// resolved via ResolvePath; wildcard sources require an existing directory
// destination.
func builtinCopyFile(evaluator *script.Evaluator, args map[string]any) (any, error) {
	src, ok := args["0"].(string)
	if !ok {
		return nil, fmt.Errorf("copy_file() requires source and destination arguments")
	}

	dst, ok := args["1"].(string)
	if !ok {
		return nil, fmt.Errorf("copy_file() requires source and destination arguments")
	}

	fullSrc := ResolvePath(src)
	fullDst := ResolvePath(dst)

	if hasWildcard(fullSrc) {
		if !core.HasPathPrefix(fullDst, "STORE") && !core.HasPathPrefix(fullDst, "EMBED") {
			info, err := os.Stat(fullDst)
			if err != nil || !info.IsDir() {
				return nil, fmt.Errorf("copy_file() with wildcard source requires destination to be an existing directory")
			}
		}

		matches, err := ExpandGlob(fullSrc)
		if err != nil {
			return nil, err
		}

		noFiles := GetSysFlag("-no-files", false)
		base := appDir()
		bareDst := !core.IsAbsoluteOrSpecial(dst)

		copied := []string{}
		for _, match := range matches {
			content, err := readFile(match)
			if err != nil {
				continue
			}

			basename := core.Base(match)
			dstPath := core.Join(fullDst, basename)

			if noFiles && !core.HasPathPrefix(dstPath, "STORE") && !core.HasPathPrefix(dstPath, "EMBED") {
				continue
			}

			if err := writeFile(dstPath, content, 0644); err == nil {
				resultPath := dstPath
				if bareDst && base != "" {
					if rel, err := core.Rel(base, dstPath); err == nil {
						resultPath = rel
					}
				}
				copied = append(copied, resultPath)
			}
		}

		result := make([]any, len(copied))
		for i, p := range copied {
			result[i] = p
		}
		return result, nil
	}

	if err := checkFilesAllowed(fullDst); err != nil {
		return nil, err
	}

	content, err := readFile(fullSrc)
	if err != nil {
		return nil, fmt.Errorf("cannot copy_file '%s': %s", src, describeFileError(err, fullSrc))
	}

	finalDst := fullDst
	if !core.HasPathPrefix(fullDst, "STORE") && !core.HasPathPrefix(fullDst, "EMBED") {
		if info, err := os.Stat(fullDst); err == nil && info.IsDir() {
			finalDst = core.Join(fullDst, core.Base(fullSrc))
		}
	}

	if !core.HasPathPrefix(finalDst, "STORE") && !core.HasPathPrefix(finalDst, "EMBED") {
		if err := os.MkdirAll(core.Dir(finalDst), 0755); err != nil {
			return nil, fmt.Errorf("cannot copy_file '%s' to '%s': %s", src, dst, describeFileError(err, finalDst))
		}
	}

	if err := writeFile(finalDst, content, 0644); err != nil {
		return nil, fmt.Errorf("cannot copy_file '%s' to '%s': %s", src, dst, describeFileError(err, finalDst))
	}
	return []any{dst}, nil
}

// builtinMoveFile moves a file from source to destination. Both paths are
// resolved via ResolvePath; /EMBED/ source is rejected.
func builtinMoveFile(evaluator *script.Evaluator, args map[string]any) (any, error) {
	src, ok := args["0"].(string)
	if !ok {
		return nil, fmt.Errorf("move_file() requires source and destination arguments")
	}

	dst, ok := args["1"].(string)
	if !ok {
		return nil, fmt.Errorf("move_file() requires source and destination arguments")
	}

	fullSrc := ResolvePath(src)
	fullDst := ResolvePath(dst)

	if core.HasPathPrefix(fullSrc, "EMBED") {
		return nil, fmt.Errorf("cannot write to /EMBED/: embedded filesystem is read-only")
	}

	if hasWildcard(fullSrc) {
		info, err := os.Stat(fullDst)
		if err != nil || !info.IsDir() {
			return nil, fmt.Errorf("move_file() with wildcard source requires destination to be an existing directory")
		}

		matches, err := ExpandGlob(fullSrc)
		if err != nil {
			return nil, err
		}

		noFiles := GetSysFlag("-no-files", false)
		base := appDir()
		bareDst := !core.IsAbsoluteOrSpecial(dst)

		moved := []string{}
		for _, match := range matches {
			basename := core.Base(match)
			dstPath := core.Join(fullDst, basename)

			if noFiles && !core.HasPathPrefix(dstPath, "STORE") && !core.HasPathPrefix(dstPath, "EMBED") {
				continue
			}

			var moveErr error
			if core.HasPathPrefix(match, "STORE") {
				content, err := readFile(match)
				if err != nil {
					continue
				}

				if err := writeFile(dstPath, content, 0644); err != nil {
					continue
				}

				srcKey := core.TrimPathPrefix(match, "STORE")
				store := runtime.GetDatastore("duso_vfs", nil)
				_, moveErr = store.Delete(srcKey)
			} else {
				moveErr = os.Rename(match, dstPath)
			}

			if moveErr == nil {
				resultPath := dstPath
				if bareDst && base != "" {
					if rel, err := core.Rel(base, dstPath); err == nil {
						resultPath = rel
					}
				}
				moved = append(moved, resultPath)
			}
		}

		result := make([]any, len(moved))
		for i, p := range moved {
			result[i] = p
		}
		return result, nil
	}

	if err := checkFilesAllowed(fullDst); err != nil {
		return nil, err
	}

	finalDst := fullDst
	if info, err := os.Stat(fullDst); err == nil && info.IsDir() {
		finalDst = core.Join(fullDst, core.Base(fullSrc))
	}

	if core.HasPathPrefix(fullSrc, "STORE") {
		content, err := readFile(fullSrc)
		if err != nil {
			return nil, fmt.Errorf("cannot move_file '%s': %s", src, describeFileError(err, fullSrc))
		}

		if err := writeFile(finalDst, content, 0644); err != nil {
			return nil, fmt.Errorf("cannot move_file '%s' to '%s': %s", src, dst, describeFileError(err, finalDst))
		}

		srcKey := core.TrimPathPrefix(fullSrc, "STORE")
		store := runtime.GetDatastore("duso_vfs", nil)
		if _, err := store.Delete(srcKey); err != nil {
			return nil, fmt.Errorf("cannot move_file '%s': %s", src, describeFileError(err, fullSrc))
		}

		return []any{dst}, nil
	}

	if err := os.Rename(fullSrc, finalDst); err != nil {
		return nil, fmt.Errorf("cannot move_file '%s' to '%s': %s", src, dst, describeFileError(err, finalDst))
	}
	return []any{dst}, nil
}
