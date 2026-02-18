package cli

import (
	"bufio"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/duso-org/duso/pkg/core"
	"github.com/duso-org/duso/pkg/runtime"
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
// - /STORE/ prefix to read from virtual datastore
func readFile(path string) ([]byte, error) {
	// Check if this is a /STORE/ path
	if core.HasPathPrefix(path, "STORE") {
		return readFromStore(path)
	}

	// Check if this is an embedded path
	if core.HasPathPrefix(path, "EMBED") {
		// Remove the /EMBED/ prefix and read from embedded filesystem
		embeddedPath := core.TrimPathPrefix(path, "EMBED")
		return EmbeddedFileRead(embeddedPath)
	}

	// Normal filesystem read
	return os.ReadFile(path)
}

// ReadEmbeddedFile reads a file from the embedded filesystem.
// Path should start with /EMBED/ for embedded files.
func ReadEmbeddedFile(path string) ([]byte, error) {
	return readFile(path)
}

// ReadScriptWithFallback reads a script file with fallback logic.
// Tries in order:
// 1. Local file at the given path
// 2. /STORE/ virtual filesystem at /STORE/{path}
// 3. Embedded file at /EMBED/{path}
// 4. Embedded file at /EMBED/{scriptDir}/{path} (for relative imports)
func ReadScriptWithFallback(scriptPath string, scriptDir string) ([]byte, error) {
	// Try 1: Local file
	if data, err := readFile(scriptPath); err == nil {
		return data, nil
	}

	// Try 2: /STORE/ virtual filesystem
	if data, err := readFile("/STORE/" + scriptPath); err == nil {
		return data, nil
	}

	// Try 3: Embedded file at /EMBED/{path}
	if data, err := readFile("/EMBED/" + scriptPath); err == nil {
		return data, nil
	}

	// Try 4: Embedded file at /EMBED/{scriptDir}/{path}
	if scriptDir != "" && scriptDir != "." {
		if data, err := readFile("/EMBED/" + scriptDir + "/" + scriptPath); err == nil {
			return data, nil
		}
	}

	// All attempts failed - return error from first attempt
	return readFile(scriptPath)
}

// writeFile is a wrapper for os.WriteFile.
// Supports writing to:
// - Normal filesystem
// - /STORE/ virtual filesystem backed by datastore
// Note: Does not support /EMBED/ writes (binary is read-only)
func writeFile(path string, data []byte, perm os.FileMode) error {
	if core.HasPathPrefix(path, "EMBED") {
		return fmt.Errorf("cannot write to /EMBED/: embedded filesystem is read-only")
	}

	if core.HasPathPrefix(path, "STORE") {
		return writeToStore(path, data)
	}

	return os.WriteFile(path, data, perm)
}

// fileExists checks if a file or directory exists, supporting disk, /EMBED/, and /STORE/ paths.
func fileExists(path string) bool {
	if core.HasPathPrefix(path, "STORE") {
		// Check in /STORE/ virtual filesystem
		_, err := readFromStore(path)
		return err == nil
	}

	if core.HasPathPrefix(path, "EMBED") {
		// Check in embedded filesystem
		embeddedPath := core.TrimPathPrefix(path, "EMBED")
		_, err := EmbeddedStat(embeddedPath)
		return err == nil
	}

	// Check on disk filesystem
	_, err := os.Stat(path)
	return err == nil
}

// getFileMtime returns the modification time of a file in Unix seconds.
// Supports disk, /EMBED/, and /STORE/ paths. Returns 0 if the file cannot be accessed.
func getFileMtime(path string) int64 {
	if core.HasPathPrefix(path, "STORE") {
		// /STORE/ files don't have real mtimes, return current time
		// This allows them to be treated as "fresh" for cache purposes
		return 0 // Will cause re-reading, which is fine
	}

	if core.HasPathPrefix(path, "EMBED") {
		// Check in embedded filesystem
		embeddedPath := core.TrimPathPrefix(path, "EMBED")
		info, err := EmbeddedStat(embeddedPath)
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

// readFromStore reads a file from the /STORE/ virtual filesystem backed by datastore.
// Path should include the /STORE/ prefix.
func readFromStore(path string) ([]byte, error) {
	// Strip /STORE/ prefix to get the datastore key
	key := core.TrimPathPrefix(path, "STORE")
	if key == "" {
		return nil, fmt.Errorf("invalid /STORE/ path: %s", path)
	}

	// Get the datastore
	store := runtime.GetDatastore("vfs", nil)

	// Get the value from datastore
	value, err := store.Get(key)
	if err != nil {
		return nil, fmt.Errorf("file not found in /STORE/: %s", key)
	}

	// Convert value to bytes
	if value == nil {
		return nil, fmt.Errorf("file not found in /STORE/: %s", key)
	}

	str, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("invalid file format in /STORE/%s: expected string", key)
	}

	return []byte(str), nil
}

// writeToStore writes a file to the /STORE/ virtual filesystem backed by datastore.
// Path should include the /STORE/ prefix.
func writeToStore(path string, data []byte) error {
	// Strip /STORE/ prefix to get the datastore key
	key := core.TrimPathPrefix(path, "STORE")
	if key == "" {
		return fmt.Errorf("invalid /STORE/ path: %s", path)
	}

	// Get the datastore
	store := runtime.GetDatastore("vfs", nil)

	// Set the value in datastore
	return store.Set(key, string(data))
}

// appendToStore appends to a file in the /STORE/ virtual filesystem.
// Path should include the /STORE/ prefix.
func appendToStore(path string, data []byte) error {
	// Strip /STORE/ prefix to get the datastore key
	key := core.TrimPathPrefix(path, "STORE")
	if key == "" {
		return fmt.Errorf("invalid /STORE/ path: %s", path)
	}

	// Get the datastore
	store := runtime.GetDatastore("vfs", nil)

	// Get existing value
	existing, err := store.Get(key)
	if err != nil {
		// File doesn't exist, create it
		return store.Set(key, string(data))
	}

	// Append to existing value
	var existingStr string
	if existing != nil {
		if s, ok := existing.(string); ok {
			existingStr = s
		}
	}

	return store.Set(key, existingStr+string(data))
}

// listFromStore lists files in a /STORE/ directory.
// Supports wildcard patterns like /STORE/*.du or /STORE/foo/*.txt
func listFromStore(pattern string) ([]map[string]any, error) {
	// Strip /STORE/ prefix
	pattern = core.TrimPathPrefix(pattern, "STORE")
	if pattern == "" {
		pattern = "*"
	}

	// Get all keys from the store (datastore doesn't provide direct enumeration,
	// so we need to implement a workaround)
	// For now, we'll use a prefix-based approach
	// This requires keys to be stored in a way that allows prefix matching

	// Extract the prefix and pattern parts
	// e.g., "*.du" -> prefix="", pattern="*.du"
	// e.g., "foo/*.du" -> prefix="foo/", pattern="*.du"
	baseDir := filepath.Dir(pattern)
	if baseDir == "." {
		baseDir = ""
	}
	_ = baseDir // Reserved for future use when implementing key enumeration

	// We need to scan all keys and filter them
	// Since Go's datastore doesn't provide enumeration, we'll use a special marker
	// For now, return empty until we implement key enumeration
	// TODO: Implement key enumeration in datastore if needed for listFromStore

	return []map[string]any{}, nil
}

// hasWildcard checks if a pattern contains wildcard characters (* or ?)
func hasWildcard(pattern string) bool {
	return strings.ContainsAny(pattern, "*?")
}

// validatePattern checks if pattern is valid and rejects unsupported syntax like **
func validatePattern(pattern string) error {
	if strings.Contains(pattern, "**") {
		return fmt.Errorf("** (recursive wildcard) is not supported\nAvailable patterns: * (match anything), ? (match single character)")
	}
	return nil
}

// ExpandGlob expands glob patterns for filesystem, /EMBED/, and /STORE/ paths.
// Returns a list of matching file paths.
func ExpandGlob(pattern string) ([]string, error) {
	// Validate pattern first
	if err := validatePattern(pattern); err != nil {
		return nil, err
	}

	// Handle /EMBED/ patterns (read-only, uses embed.FS)
	if core.HasPathPrefix(pattern, "EMBED") {
		embeddedPattern := core.TrimPathPrefix(pattern, "EMBED")

		matches, err := EmbeddedGlob(embeddedPattern)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern: %w", err)
		}

		// Add /EMBED/ prefix back to results
		for i, match := range matches {
			matches[i] = "/EMBED/" + match
		}
		return matches, nil
	}

	// Handle /STORE/ patterns (datastore backed)
	if core.HasPathPrefix(pattern, "STORE") {
		storePattern := core.TrimPathPrefix(pattern, "STORE")

		// Get the vfs datastore
		store := runtime.GetDatastore("vfs", nil)

		// Get all keys from datastore
		allKeys := store.Keys()

		// Filter keys that match the pattern
		var matches []string
		for _, key := range allKeys {
			matched, err := filepath.Match(storePattern, key)
			if err != nil {
				return nil, fmt.Errorf("invalid pattern: %w", err)
			}
			if matched {
				matches = append(matches, "/STORE/"+key)
			}
		}
		return matches, nil
	}

	// Use filepath.Glob for regular filesystem patterns
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid pattern: %w", err)
	}
	return matches, nil
}

// matchGlob matches a file name against a glob pattern (*, ?)
func matchGlob(pattern, name string) (bool, error) {
	return filepath.Match(pattern, name)
}

// Global reader for input() to preserve buffer state across multiple calls
// (unlike creating a new Scanner each time)
var stdinReader *bufio.Reader

func init() {
	stdinReader = bufio.NewReader(os.Stdin)
}

// readInputLine reads a line from stdin with optional prompt
// Used as the InputReader capability for input() builtin
func readInputLine(prompt string) (string, error) {
	if prompt != "" {
		fmt.Fprint(os.Stdout, prompt)
		os.Stdout.Sync()
	}

	// Use the global reader to read full lines including spaces
	// (unlike fmt.Scanln which only reads words)
	line, err := stdinReader.ReadString('\n')
	if err != nil && err.Error() != "EOF" {
		return "", err
	}

	// Remove the trailing newline
	if len(line) > 0 && line[len(line)-1] == '\n' {
		line = line[:len(line)-1]
	}

	return line, nil
}
