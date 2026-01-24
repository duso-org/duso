package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
// Supports both direct modules (http.du) and directory-based modules (http/http.du).
// Returns: (resolved absolute path, list of paths searched, error)
func (r *ModuleResolver) ResolveModule(moduleName string) (string, []string, error) {
	var searchedPaths []string

	// Create search path list, always ending with embedded modules (stdlib, then contrib)
	searchPaths := append([]string{}, r.DusoPath...)
	searchPaths = append(searchPaths, "/EMBED/stdlib")
	searchPaths = append(searchPaths, "/EMBED/contrib")

	// Helper to check if a file path exists and add to searchedPaths
	// Only returns true for regular files, not directories
	// Works with both disk and /EMBED/ paths
	checkPath := func(path string) (string, bool) {
		searchedPaths = append(searchedPaths, path)
		if fileExists(path) {
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

	// Step 2: Try relative to script directory
	if r.ScriptDir != "" {
		relPath := filepath.Join(r.ScriptDir, moduleName)

		if resolved, found := tryResolve(relPath); found {
			return resolved, searchedPaths, nil
		}
	}

	// Step 3: Try each search path (DUSO_LIB directories + embedded stdlib)
	for _, dir := range searchPaths {
		if dir == "" {
			continue
		}

		expandedDir := expandHome(dir)
		pathInDir := filepath.Join(expandedDir, moduleName)

		if resolved, found := tryResolve(pathInDir); found {
			return resolved, searchedPaths, nil
		}
	}

	// Step 4: Not found - return error with searched paths
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
