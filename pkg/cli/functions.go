// Package cli provides CLI-specific functions for Duso scripts.
//
// These functions extend the core language with file I/O and Claude API integration.
// They are NOT part of the core language and are only available when using the duso CLI.
//
// Embedded Go applications can optionally register these functions if they wish,
// or implement their own versions with different behavior.
package cli

import (
	"fmt"
	"os"
	"path/filepath"
)

// FileIOContext holds context for file I/O operations (script directory, etc.)
type FileIOContext struct {
	ScriptDir string
}

// NewLoadFunction creates a load(filename) function that reads files.
//
// load() reads the contents of a file relative to the script's directory.
// It's only available in the CLI environment.
//
// Example:
//     content = load("data.txt")
//     data = parse_json(load("config.json"))
func NewLoadFunction(ctx FileIOContext) func(map[string]any) (any, error) {
	return func(args map[string]any) (any, error) {
		filename, ok := args["0"].(string)
		if !ok {
			// Check for named argument "filename"
			if f, ok := args["filename"]; ok {
				filename = fmt.Sprintf("%v", f)
			} else {
				return nil, fmt.Errorf("load() requires a filename argument")
			}
		}

		fullPath := filepath.Join(ctx.ScriptDir, filename)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return nil, fmt.Errorf("cannot load '%s': %w", filename, err)
		}

		return string(content), nil
	}
}

// NewSaveFunction creates a save(filename, content) function that writes files.
//
// save() writes content to a file relative to the script's directory.
// It's only available in the CLI environment.
//
// Example:
//     save("output.txt", "Hello, World!")
//     save("data.json", format_json(myObject))
func NewSaveFunction(ctx FileIOContext) func(map[string]any) (any, error) {
	return func(args map[string]any) (any, error) {
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

		fullPath := filepath.Join(ctx.ScriptDir, filename)

		// Create parent directories if needed
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return nil, fmt.Errorf("cannot create directory: %w", err)
		}

		err := os.WriteFile(fullPath, []byte(content), 0644)
		if err != nil {
			return nil, fmt.Errorf("cannot save to '%s': %w", filename, err)
		}

		return nil, nil
	}
}

// NewIncludeFunction creates an include(filename) function that executes other scripts.
//
// include() loads and executes another .du script file in the current environment.
// Variables and functions defined in the included script are available after include().
// It's only available in the CLI environment.
//
// Example:
//     include("helpers.du")
//     result = helper_function()  // Now available
//
// Note: The includeExecutor function must be provided to actually execute the included script.
func NewIncludeFunction(ctx FileIOContext, includeExecutor func(string) error) func(map[string]any) (any, error) {
	return func(args map[string]any) (any, error) {
		filename, ok := args["0"].(string)
		if !ok {
			// Check for named argument "filename"
			if f, ok := args["filename"]; ok {
				filename = fmt.Sprintf("%v", f)
			} else {
				return nil, fmt.Errorf("include() requires a filename argument")
			}
		}

		fullPath := filepath.Join(ctx.ScriptDir, filename)
		source, err := os.ReadFile(fullPath)
		if err != nil {
			return nil, fmt.Errorf("cannot include '%s': %w", filename, err)
		}

		// Execute the included script in the current environment
		if err := includeExecutor(string(source)); err != nil {
			return nil, fmt.Errorf("error in included script '%s': %w", filename, err)
		}

		return nil, nil
	}
}
