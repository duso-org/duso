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
// load() reads the contents of a file using centralized path resolution:
// 1. Absolute paths and /STORE/, /EMBED/ → used as-is
// 2. Relative paths → tries cwd, script dir, /STORE/, /EMBED/ in order
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

		// Get script directory from current invocation frame
		var scriptDir string
		gid := script.GetGoroutineID()
		if ctx, ok := script.GetRequestContext(gid); ok && ctx.Frame != nil && ctx.Frame.Filename != "" {
			scriptDir = filepath.Dir(ctx.Frame.Filename)
		}
		if scriptDir == "" {
			scriptDir = interp.GetScriptDir()
		}

		// Use host-provided FileReader capability
		if interp.FileReader == nil {
			return nil, fmt.Errorf("load() requires FileReader capability (not provided by host)")
		}

		// Try paths in order: absolute/virtual as-is, then cwd, scriptDir, /STORE/, /EMBED/
		if filepath.IsAbs(filename) || strings.HasPrefix(filename, "/") {
			// Absolute or virtual path - try as-is
			if err := checkFilesAllowed(filename, interp.NoFiles); err != nil {
				return nil, err
			}
			content, err := interp.FileReader(filename)
			if err != nil {
				return nil, fmt.Errorf("cannot load '%s': %w", filename, err)
			}
			return string(content), nil
		}

		// For relative paths, try candidates in order:
		// 1. scriptDir (where the script lives)
		// 2. /STORE/ (virtual filesystem)
		// 3. /EMBED/ (embedded resources)
		// Note: cwd is NOT used here - only at CLI invocation time to find the initial script
		candidates := []string{
			filepath.Join(scriptDir, filename), // script directory (primary)
			filepath.Join("/STORE", filename),  // virtual filesystem fallback
			filepath.Join("/EMBED", filename),  // embedded resources fallback
		}

		var lastErr error
		for _, candidate := range candidates {
			if err := checkFilesAllowed(candidate, interp.NoFiles); err != nil {
				continue
			}
			content, err := interp.FileReader(candidate)
			if err == nil {
				return string(content), nil
			}
			lastErr = err
		}

		// All candidates failed
		if lastErr != nil {
			return nil, fmt.Errorf("cannot load '%s': %w", filename, lastErr)
		}
		return nil, fmt.Errorf("cannot load '%s': file not found", filename)
	}
}

// NewSaveFunction creates a save(filename, content) builtin that writes files.
//
// save() writes content to a file using centralized path resolution:
// 1. Absolute paths and /STORE/, /EMBED/ → used as-is
// 2. Relative paths → tries cwd, script dir, /STORE/, /EMBED/ in order
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

		// Get script directory from current invocation frame
		var scriptDir string
		gid := script.GetGoroutineID()
		if ctx, ok := script.GetRequestContext(gid); ok && ctx.Frame != nil && ctx.Frame.Filename != "" {
			scriptDir = filepath.Dir(ctx.Frame.Filename)
		}
		if scriptDir == "" {
			scriptDir = interp.GetScriptDir()
		}

		// Use host-provided FileWriter capability
		if interp.FileWriter == nil {
			return nil, fmt.Errorf("save() requires FileWriter capability (not provided by host)")
		}

		// Resolve path: absolute/virtual as-is, else use scriptDir
		// (save goes to script's directory, not cwd)
		var fullPath string
		if filepath.IsAbs(filename) || strings.HasPrefix(filename, "/") {
			fullPath = filename
		} else {
			// Relative path - use script directory
			fullPath = filepath.Join(scriptDir, filename)
		}

		// Check if file operations are allowed
		if err := checkFilesAllowed(fullPath, interp.NoFiles); err != nil {
			return nil, err
		}

		// For /STORE/, don't create parent directories (they're implicit)
		// For regular filesystem, create parent directories if needed
		if !strings.HasPrefix(fullPath, "/STORE/") && !strings.HasPrefix(fullPath, "/EMBED/") {
			if mkErr := os.MkdirAll(filepath.Dir(fullPath), 0755); mkErr != nil {
				return nil, fmt.Errorf("cannot create directory: %w", mkErr)
			}
		}

		saveErr := interp.FileWriter(fullPath, content)
		if saveErr != nil {
			return nil, fmt.Errorf("cannot save to '%s': %w", filename, saveErr)
		}

		return nil, nil
	}
}
