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
	"strings"

	"github.com/duso-org/duso/pkg/script"
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
// Unlike require(), include() executes in the current scope (not isolated),
// and results are not cached.
//
// Example:
//     include("helpers.du")
//     result = helper_function()  // Now available
//
// This function supports path resolution: user-provided paths, relative to script dir, and DUSO_PATH.
func NewIncludeFunction(resolver *ModuleResolver, detector *CircularDetector, includeExecutor func(string) error) func(map[string]any) (any, error) {
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

		// Resolve module path using standard resolution algorithm
		fullPath, searchedPaths, err := resolver.ResolveModule(filename)
		if err != nil {
			return nil, fmt.Errorf("cannot find module '%s'\nSearched:\n  %s",
				filename, strings.Join(searchedPaths, "\n  "))
		}

		// Check for circular dependency
		if err := detector.Push(fullPath); err != nil {
			return nil, err
		}
		defer detector.Pop()

		// Read file
		source, err := os.ReadFile(fullPath)
		if err != nil {
			return nil, fmt.Errorf("cannot include '%s': %w", fullPath, err)
		}

		// Execute in current environment (no isolation)
		if err := includeExecutor(string(source)); err != nil {
			return nil, fmt.Errorf("error in included script '%s': %w", fullPath, err)
		}

		return nil, nil
	}
}

// NewRequireFunction creates a require(moduleName) function that loads modules.
//
// require() loads a module in an isolated scope and returns its exports.
// Unlike include(), require():
// - Executes the module in its own isolated scope
// - Returns the last expression value (the module's exports)
// - Caches results - subsequent requires return cached value without re-executing
//
// Example:
//     math = require("math")
//     result = math.add(2, 3)  // Calls function from module
//
// This function supports path resolution: user-provided paths, relative to script dir, and DUSO_PATH.
func NewRequireFunction(resolver *ModuleResolver, detector *CircularDetector, interp *script.Interpreter) func(map[string]any) (any, error) {
	return func(args map[string]any) (any, error) {
		filename, ok := args["0"].(string)
		if !ok {
			// Check for named argument "filename"
			if f, ok := args["filename"]; ok {
				filename = fmt.Sprintf("%v", f)
			} else {
				return nil, fmt.Errorf("require() requires a filename argument")
			}
		}

		// Resolve module path using standard resolution algorithm
		fullPath, searchedPaths, err := resolver.ResolveModule(filename)
		if err != nil {
			return nil, fmt.Errorf("cannot find module '%s'\nSearched:\n  %s",
				filename, strings.Join(searchedPaths, "\n  "))
		}

		// Check module cache (absolute path as key)
		if cached, ok := interp.GetModuleCache(fullPath); ok {
			return script.ValueToInterface(cached), nil
		}

		// Check for circular dependency
		if err := detector.Push(fullPath); err != nil {
			return nil, err
		}
		defer detector.Pop()

		// Read file
		source, err := os.ReadFile(fullPath)
		if err != nil {
			return nil, fmt.Errorf("cannot require '%s': %w", fullPath, err)
		}

		// Execute in isolated scope
		value, err := interp.ExecuteModule(string(source))
		if err != nil {
			return nil, fmt.Errorf("error in module '%s': %w", fullPath, err)
		}

		// Cache the result
		interp.SetModuleCache(fullPath, value)

		// Convert to interface{} for return
		return script.ValueToInterface(value), nil
	}
}
