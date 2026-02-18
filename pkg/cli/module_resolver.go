package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/duso-org/duso/pkg/core"
)

// ModuleResolver handles finding module files using the standard search order:
// 1. User-provided filespec (absolute or ~/...)
// 2. Relative to script directory
// 3. DUSO_LIB environment variable directories
type ModuleResolver struct {
	ScriptDir string   // Directory of the currently executing script
	DusoPath  []string // Parsed from DUSO_LIB env variable
}

// ResolveModule finds a module file using the standard resolution algorithm.
// Search order: . (current dir) → $DUSO_LIB paths → /EMBED
// Supports both direct modules (http.du) and directory-based modules (http/http.du).
// Returns: (resolved absolute path, list of paths searched, error)
func (r *ModuleResolver) ResolveModule(moduleName string) (string, []string, error) {
	var searchedPaths []string

	// Build search path list in the correct order:
	// 1. Current directory (.)
	// 2. DUSO_LIB directories (from environment variable, in order)
	// 3. Embedded modules (stdlib, then contrib)
	searchPaths := []string{"."}
	searchPaths = append(searchPaths, r.DusoPath...)
	searchPaths = append(searchPaths, "/EMBED/stdlib")
	searchPaths = append(searchPaths, "/EMBED/contrib")

	// Helper to check if a path is a directory
	isDir := func(path string) bool {
		if core.HasPathPrefix(path, "EMBED") {
			embeddedPath := core.TrimPathPrefix(path, "EMBED")
			stat, err := EmbeddedStat(embeddedPath)
			return err == nil && stat.IsDir()
		}
		stat, err := os.Stat(path)
		return err == nil && stat.IsDir()
	}

	// Helper to check if a file path exists and add to searchedPaths
	// Only returns true for regular files, not directories
	// Works with both disk and /EMBED/ paths
	checkPath := func(path string) (string, bool) {
		searchedPaths = append(searchedPaths, path)
		if fileExists(path) && !isDir(path) {
			return path, true
		}
		return "", false
	}

	// Helper to try resolving a path, with optional directory-based fallback
	// For "http", tries: "http", "http.du", "http/http.du"
	tryResolve := func(basePath string) (string, bool) {
		// Try exact path
		if resolved, found := checkPath(basePath); found {
			return resolved, true
		}

		// Try with .du extension if no extension present
		if !strings.HasSuffix(basePath, ".du") {
			withDu := basePath + ".du"
			if resolved, found := checkPath(withDu); found {
				return resolved, true
			}

			// Try directory-based module: basePath/baseName.du
			// For "http" -> "http/http.du"
			// For "http/cache" -> "http/cache/cache.du"
			baseName := filepath.Base(basePath)
			dirBased := filepath.Join(basePath, baseName+".du")
			if resolved, found := checkPath(dirBased); found {
				return resolved, true
			}
		}

		return "", false
	}

	// Step 1: User-provided filespec (absolute or ~/...)
	if filepath.IsAbs(moduleName) || strings.HasPrefix(moduleName, "~") {
		expandedPath := expandHome(moduleName)

		if resolved, found := tryResolve(expandedPath); found {
			return resolved, searchedPaths, nil
		}
	}

	// Step 2: Try each search path with three variants:
	// - Direct: dir/moduleName
	// - Stdlib: dir/stdlib/moduleName
	// - Contrib: dir/contrib/moduleName
	for _, dir := range searchPaths {
		if dir == "" {
			continue
		}

		expandedDir := expandHome(dir)

		// Try direct path
		pathInDir := filepath.Join(expandedDir, moduleName)
		if resolved, found := tryResolve(pathInDir); found {
			return resolved, searchedPaths, nil
		}

		// Try in stdlib subdirectory
		stdlibPath := filepath.Join(expandedDir, "stdlib", moduleName)
		if resolved, found := tryResolve(stdlibPath); found {
			return resolved, searchedPaths, nil
		}

		// Try in contrib subdirectory
		contribPath := filepath.Join(expandedDir, "contrib", moduleName)
		if resolved, found := tryResolve(contribPath); found {
			return resolved, searchedPaths, nil
		}
	}

	// Step 3: Not found - return error with searched paths
	return "", searchedPaths, fmt.Errorf("module not found: %s", moduleName)
}

// expandHome expands ~ to the user's home directory
func expandHome(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	if path == "~" {
		return home
	}

	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:])
	}

	return path
}
