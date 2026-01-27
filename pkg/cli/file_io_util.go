package cli

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"strings"
)

// embeddedFS holds the embedded stdlib and docs directories
// Initialized by SetEmbeddedFS() from cmd/duso/main.go
var embeddedFS embed.FS

// SetEmbeddedFS sets the embedded filesystem for use by file I/O functions.
// Called from cmd/duso/main.go during initialization.
func SetEmbeddedFS(fs embed.FS) {
	embeddedFS = fs
}

// readFile is a wrapper for os.ReadFile that supports:
// - Normal filesystem paths
// - /EMBED/ prefix to read from embedded stdlib/docs directories
func readFile(path string) ([]byte, error) {
	// Check if this is an embedded path
	if strings.HasPrefix(path, "/EMBED/") {
		// Remove the /EMBED/ prefix and read from embedded filesystem
		embeddedPath := strings.TrimPrefix(path, "/EMBED/")
		return embeddedFS.ReadFile(embeddedPath)
	}

	// Normal filesystem read
	return os.ReadFile(path)
}

// ReadEmbeddedFile reads a file from the embedded filesystem.
// Path should start with /EMBED/ for embedded files.
func ReadEmbeddedFile(path string) ([]byte, error) {
	return readFile(path)
}

// ReadScriptWithFallback reads a script file with fallback logic for embedded files.
// Tries in order:
// 1. Local file at the given path
// 2. Embedded file at /EMBED/{path}
// 3. Embedded file at /EMBED/{scriptDir}/{path} (for relative imports)
func ReadScriptWithFallback(scriptPath string, scriptDir string) ([]byte, error) {
	// Try 1: Local file
	if data, err := readFile(scriptPath); err == nil {
		return data, nil
	}

	// Try 2: Embedded file at /EMBED/{path}
	if data, err := readFile("/EMBED/" + scriptPath); err == nil {
		return data, nil
	}

	// Try 3: Embedded file at /EMBED/{scriptDir}/{path}
	if scriptDir != "" && scriptDir != "." {
		if data, err := readFile("/EMBED/" + scriptDir + "/" + scriptPath); err == nil {
			return data, nil
		}
	}

	// All attempts failed - return error from first attempt
	return readFile(scriptPath)
}

// writeFile is a wrapper for os.WriteFile.
//
// Currently a pass-through to os.WriteFile.
// Future: May add logic to handle virtual filesystems or other destinations.
// Note: Does not support /EMBED/ writes (binary is read-only)
func writeFile(path string, data []byte, perm os.FileMode) error {
	if strings.HasPrefix(path, "/EMBED/") {
		return fmt.Errorf("cannot write to /EMBED/: embedded filesystem is read-only")
	}

	return os.WriteFile(path, data, perm)
}

// fileExists checks if a file exists, supporting both disk and /EMBED/ paths.
// Returns true only for regular files, not directories.
func fileExists(path string) bool {
	if strings.HasPrefix(path, "/EMBED/") {
		// Check in embedded filesystem
		embeddedPath := strings.TrimPrefix(path, "/EMBED/")
		info, err := fs.Stat(embeddedFS, embeddedPath)
		return err == nil && !info.IsDir()
	}

	// Check on disk filesystem
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// getFileMtime returns the modification time of a file in Unix seconds.
// Supports both disk and /EMBED/ paths. Returns 0 if the file cannot be accessed.
func getFileMtime(path string) int64 {
	if strings.HasPrefix(path, "/EMBED/") {
		// Check in embedded filesystem
		embeddedPath := strings.TrimPrefix(path, "/EMBED/")
		info, err := fs.Stat(embeddedFS, embeddedPath)
		if err != nil {
			return 0
		}
		return info.ModTime().Unix()
	}

	// Check on disk filesystem
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.ModTime().Unix()
}
