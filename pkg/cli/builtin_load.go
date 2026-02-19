package cli

import (
	"fmt"
	"strings"

	"github.com/duso-org/duso/pkg/core"
	"github.com/duso-org/duso/pkg/script"
)

// builtinLoad reads a file and returns its contents as a string.
//
// load(filename) reads a file using centralized path resolution:
// 1. Absolute paths and /STORE/, /EMBED/ → used as-is
// 2. Relative paths → tries script dir, /STORE/, /EMBED/ in order
//
// Example:
//
//	content = load("data.txt")
//	data = parse_json(load("config.json"))
//	code = load("/STORE/generated.du")
func builtinLoad(evaluator *script.Evaluator, args map[string]any) (any, error) {
	filename, ok := args["0"].(string)
	if !ok {
		if f, ok := args["filename"]; ok {
			filename = fmt.Sprintf("%v", f)
		} else {
			return nil, fmt.Errorf("load() requires a filename argument")
		}
	}

	// Get script directory from request context
	scriptDir := ""
	gid := script.GetGoroutineID()
	if ctx, ok := script.GetRequestContext(gid); ok && ctx.Frame != nil && ctx.Frame.Filename != "" {
		scriptDir = core.Dir(ctx.Frame.Filename)
	}

	// For absolute/virtual paths, use as-is
	if core.IsAbsolute(filename) || strings.HasPrefix(filename, "/") {
		content, err := readFile(filename)
		if err != nil {
			return nil, fmt.Errorf("cannot load '%s': %w", filename, err)
		}
		return string(content), nil
	}

	// For relative paths, try candidates in order: scriptDir, /STORE/, /EMBED/
	candidates := []string{
		core.Join(scriptDir, filename),
		core.Join("/STORE", filename),
		core.Join("/EMBED", filename),
	}

	var lastErr error
	for _, candidate := range candidates {
		content, err := readFile(candidate)
		if err == nil {
			return string(content), nil
		}
		lastErr = err
	}

	if lastErr != nil {
		return nil, fmt.Errorf("cannot load '%s': %w", filename, lastErr)
	}
	return nil, fmt.Errorf("cannot load '%s': file not found", filename)
}

// builtinSave writes a string to a file.
//
// save(filename, content) writes content to a file using centralized path resolution:
// 1. Absolute paths and /STORE/ → used as-is
// 2. Relative paths → written to script dir
//
// Example:
//
//	save("output.txt", "Hello, World!")
//	save("data.json", format_json(myObject))
//	save("/STORE/generated.du", code)
func builtinSave(evaluator *script.Evaluator, args map[string]any) (any, error) {
	filename, ok := args["0"].(string)
	if !ok {
		if f, ok := args["filename"]; ok {
			filename = fmt.Sprintf("%v", f)
		} else {
			return nil, fmt.Errorf("save() requires filename and content arguments")
		}
	}

	content, ok := args["1"].(string)
	if !ok {
		if c, ok := args["content"]; ok {
			content = fmt.Sprintf("%v", c)
		} else {
			return nil, fmt.Errorf("save() requires filename and content arguments")
		}
	}

	// Get script directory from request context
	scriptDir := ""
	gid := script.GetGoroutineID()
	if ctx, ok := script.GetRequestContext(gid); ok && ctx.Frame != nil && ctx.Frame.Filename != "" {
		scriptDir = core.Dir(ctx.Frame.Filename)
	}

	// Resolve path: absolute/virtual as-is, else use scriptDir
	var fullPath string
	if core.IsAbsolute(filename) || strings.HasPrefix(filename, "/") {
		fullPath = filename
	} else {
		fullPath = core.Join(scriptDir, filename)
	}

	// Write the file
	if err := writeFile(fullPath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("cannot save to '%s': %w", filename, err)
	}

	return nil, nil
}
