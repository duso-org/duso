package runtime

import (
	"os"
	"path/filepath"
	"strings"
)

// ResolveFilePath resolves a file path using a consistent search order.
// This is the SINGLE source of truth for all file path resolution in Duso.
//
// For absolute or virtual paths (starting with / or /EMBED/ or /STORE/):
//   Returns the path as-is
//
// For relative paths, tries in order:
//   1. cwd + filespec (current working directory, for development)
//   2. scriptDir + filespec (directory of executing script)
//   3. /STORE/ + filespec (virtual filesystem)
//   4. /EMBED/ + filespec (embedded resources)
//
// Returns the first path that exists, or the scriptDir-based path as fallback
// (letting the actual file operation handle the error).
//
// scriptDir should be the FULL path where the script was sourced from,
// including /STORE/ or /EMBED/ if applicable.
func ResolveFilePath(filespec string, scriptDir string, runtimeCwd string) string {
	// If already absolute or virtual, return as-is
	if filepath.IsAbs(filespec) || strings.HasPrefix(filespec, "/") {
		return filespec
	}

	// For relative paths, build candidate list
	candidates := []string{
		filepath.Join(runtimeCwd, filespec),      // cwd
		filepath.Join(scriptDir, filespec),       // script directory
		filepath.Join("/STORE", filespec),        // virtual filesystem
		filepath.Join("/EMBED", filespec),        // embedded resources
	}

	// Return the first candidate that exists, or scriptDir-based as fallback
	for _, candidate := range candidates {
		if pathExists(candidate) {
			return candidate
		}
	}

	// Return scriptDir-based path as fallback (let actual operation fail with proper error)
	return filepath.Join(scriptDir, filespec)
}

// pathExists checks if a path exists, handling both real filesystem and virtual paths.
// Does NOT try to stat /EMBED/ (it's compile-time embedded); let file operations handle it.
func pathExists(path string) bool {
	// Check real filesystem
	_, err := os.Stat(path)
	if err == nil {
		return true
	}

	// Check /STORE/ virtual filesystem
	if strings.HasPrefix(path, "/STORE/") {
		key := strings.TrimPrefix(path, "/STORE/")
		store := GetDatastore("vfs", nil)
		if store != nil {
			val, _ := store.Get(key)
			return val != nil
		}
	}

	return false
}
