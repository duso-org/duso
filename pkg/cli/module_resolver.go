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
// 3. DUSO_PATH environment variable directories
type ModuleResolver struct {
	ScriptDir string   // Directory of the currently executing script
	DusoPath  []string // Parsed from DUSO_PATH env variable
}

// ResolveModule finds a module file using the standard resolution algorithm.
// Returns: (resolved absolute path, list of paths searched, error)
func (r *ModuleResolver) ResolveModule(moduleName string) (string, []string, error) {
	var searchedPaths []string

	// Helper to check if a path exists and add to searchedPaths
	checkPath := func(path string) (string, bool) {
		searchedPaths = append(searchedPaths, path)
		if _, err := os.Stat(path); err == nil {
			return path, true
		}
		return "", false
	}

	// Step 1: User-provided filespec (absolute or ~/...)
	if filepath.IsAbs(moduleName) || strings.HasPrefix(moduleName, "~") {
		expandedPath := expandHome(moduleName)

		// Try exact path
		if resolved, found := checkPath(expandedPath); found {
			return resolved, searchedPaths, nil
		}

		// Try with .du extension if no extension present
		if !strings.HasSuffix(expandedPath, ".du") {
			withDu := expandedPath + ".du"
			if resolved, found := checkPath(withDu); found {
				return resolved, searchedPaths, nil
			}
		}
	}

	// Step 2: Try relative to script directory
	if r.ScriptDir != "" {
		relPath := filepath.Join(r.ScriptDir, moduleName)

		// Try exact path
		if resolved, found := checkPath(relPath); found {
			return resolved, searchedPaths, nil
		}

		// Try with .du extension if no extension present
		if !strings.HasSuffix(relPath, ".du") {
			withDu := relPath + ".du"
			if resolved, found := checkPath(withDu); found {
				return resolved, searchedPaths, nil
			}
		}
	}

	// Step 3: Try each DUSO_PATH directory
	for _, dusoDir := range r.DusoPath {
		if dusoDir == "" {
			continue
		}

		expandedDir := expandHome(dusoDir)
		pathInDuso := filepath.Join(expandedDir, moduleName)

		// Try exact path
		if resolved, found := checkPath(pathInDuso); found {
			return resolved, searchedPaths, nil
		}

		// Try with .du extension if no extension present
		if !strings.HasSuffix(pathInDuso, ".du") {
			withDu := pathInDuso + ".du"
			if resolved, found := checkPath(withDu); found {
				return resolved, searchedPaths, nil
			}
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
