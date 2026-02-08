package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/duso-org/duso/pkg/script"
)

// checkFilesAllowed checks if file operations are allowed for the given path.
// If NoFiles is enabled, only /STORE/ and /EMBED/ paths are allowed.
func checkFilesAllowed(path string, noFiles bool) error {
	if !noFiles {
		return nil // Files are allowed
	}

	// NoFiles is enabled - only allow /STORE/ and /EMBED/
	if strings.HasPrefix(path, "/STORE/") || strings.HasPrefix(path, "/EMBED/") {
		return nil
	}

	return fmt.Errorf("filesystem access disabled (use -no-files=false to enable)")
}

// NewLoadFunction creates a load(filename) builtin that reads files.
//
// load() reads the contents of a file. Supports:
// - Relative paths (relative to script directory)
// - Absolute paths
// - /STORE/ virtual filesystem paths
// - /EMBED/ embedded files
//
// Uses host-provided FileReader capability for actual I/O.
//
// Example:
//
//	content = load("data.txt")
//	data = parse_json(load("config.json"))
//	code = load("/STORE/generated.du")  // Load from virtual filesystem
func NewLoadFunction(interp *script.Interpreter) func(*script.Evaluator, map[string]any) (any, error) {
	return func(evaluator *script.Evaluator, args map[string]any) (any, error) {
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
		// Get script directory from current invocation frame (for relative path resolution)
		var scriptDir string
		gid := script.GetGoroutineID()
		if ctx, ok := script.GetRequestContext(gid); ok && ctx.Frame != nil && ctx.Frame.Filename != "" {
			scriptDir = filepath.Dir(ctx.Frame.Filename)
		}
		if scriptDir == "" {
			// Fallback to interpreter's script dir (set at startup)
			scriptDir = interp.GetScriptDir()
		}

		var fullPath string
		if filepath.IsAbs(filename) || strings.HasPrefix(filename, "/") {
			fullPath = filename
		} else {
			fullPath = filepath.Join(scriptDir, filename)
		}

		// Check if file operations are allowed
		if err := checkFilesAllowed(fullPath, interp.NoFiles); err != nil {
			return nil, err
		}

		// Use host-provided FileReader capability
		if interp.FileReader == nil {
			return nil, fmt.Errorf("load() requires FileReader capability (not provided by host)")
		}

		// Try to load as specified first (supports /STORE/, /EMBED/, absolute, home paths)
		content, err := interp.FileReader(filename)
		if err != nil {
			// Fallback: try with script directory prepended (for relative paths)
			fallbackPath := filepath.Join(scriptDir, filename)
			content, err = interp.FileReader(fallbackPath)
			if err != nil {
				return nil, fmt.Errorf("cannot load '%s': %w", filename, err)
			}
		}

		return string(content), nil
	}
}

// NewSaveFunction creates a save(filename, content) builtin that writes files.
//
// save() writes content to a file. Supports:
// - Relative paths (relative to script directory)
// - Absolute paths
// - /STORE/ virtual filesystem paths
// - /EMBED/ paths (read-only, will error)
//
// Uses host-provided FileWriter capability for actual I/O.
//
// Example:
//
//	save("output.txt", "Hello, World!")
//	save("data.json", format_json(myObject))
//	save("/STORE/generated.du", code)  // Save to virtual filesystem
func NewSaveFunction(interp *script.Interpreter) func(*script.Evaluator, map[string]any) (any, error) {
	return func(evaluator *script.Evaluator, args map[string]any) (any, error) {
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
		// Get script directory from current invocation frame (for relative path resolution)
		var scriptDir string
		gid := script.GetGoroutineID()
		if ctx, ok := script.GetRequestContext(gid); ok && ctx.Frame != nil && ctx.Frame.Filename != "" {
			scriptDir = filepath.Dir(ctx.Frame.Filename)
		}
		if scriptDir == "" {
			// Fallback to interpreter's script dir (set at startup)
			scriptDir = interp.GetScriptDir()
		}

		var fullPath string
		if filepath.IsAbs(filename) || strings.HasPrefix(filename, "/") {
			// Absolute path or virtual filesystem
			fullPath = filename
		} else {
			// Relative path - prefix with script directory
			fullPath = filepath.Join(scriptDir, filename)
		}

		// Check if file operations are allowed
		if err := checkFilesAllowed(fullPath, interp.NoFiles); err != nil {
			return nil, err
		}

		// Use host-provided FileWriter capability
		if interp.FileWriter == nil {
			return nil, fmt.Errorf("save() requires FileWriter capability (not provided by host)")
		}

		// For /STORE/, don't create parent directories (they're implicit)
		// For regular filesystem, create parent directories if needed
		if !strings.HasPrefix(fullPath, "/STORE/") && !strings.HasPrefix(fullPath, "/EMBED/") {
			if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
				return nil, fmt.Errorf("cannot create directory: %w", err)
			}
		}

		err := interp.FileWriter(fullPath, content)
		if err != nil {
			return nil, fmt.Errorf("cannot save to '%s': %w", filename, err)
		}

		return nil, nil
	}
}
