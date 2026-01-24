// Package cli provides CLI-specific functions for Duso scripts.
//
// These functions extend the core language with file I/O, environment access, and module loading.
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

	"github.com/duso-org/duso/pkg/markdown"
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

		// Try to load as specified first (supports /EMBED/, absolute, home paths)
		content, err := readFile(filename)
		if err != nil {
			// Fallback: try with script directory prepended (for relative paths)
			fallbackPath := filepath.Join(ctx.ScriptDir, filename)
			content, err = readFile(fallbackPath)
			if err != nil {
				return nil, fmt.Errorf("cannot load '%s': %w", filename, err)
			}
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

		err := writeFile(fullPath, []byte(content), 0644)
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
// This function supports path resolution: user-provided paths, relative to script dir, and DUSO_LIB.
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
		source, err := readFile(fullPath)
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
// This function supports path resolution: user-provided paths, relative to script dir, and DUSO_LIB.
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
		source, err := readFile(fullPath)
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

// NewEnvFunction creates an env(varname) function that reads environment variables.
//
// env() reads the value of an environment variable from the OS.
// It's only available in the CLI environment (not in embedded contexts without explicit opt-in).
//
// Example:
//     key = env("ANTHROPIC_API_KEY")
//     debug = env("DEBUG_MODE")
//
// Returns the value as a string, or empty string if the variable is not set.
func NewEnvFunction() func(map[string]any) (any, error) {
	return func(args map[string]any) (any, error) {
		varname, ok := args["0"].(string)
		if !ok {
			// Check for named argument "varname"
			if v, ok := args["varname"]; ok {
				varname = fmt.Sprintf("%v", v)
			} else {
				return nil, fmt.Errorf("env() requires a variable name argument")
			}
		}

		return os.Getenv(varname), nil
	}
}

// NewDocFunction creates a doc(name) function that displays documentation.
//
// doc() searches for documentation in this order:
// 1. Module documentation (.du files with matching .md, using require() resolution)
// 2. Reference documentation (docs/reference/*.md for builtins and CLI functions)
//
// It's only available in the CLI environment.
//
// Example:
//     docs = doc("http")      // Module docs
//     docs = doc("split")     // Builtin reference docs
//     print(markdown(docs))
//
// The function prints the full path to the documentation file before the content,
// which helps with debugging version issues.
// Returns nil if the documentation is not found.
func NewDocFunction(resolver *ModuleResolver) func(map[string]any) (any, error) {
	return func(args map[string]any) (any, error) {
		name, ok := args["0"].(string)
		if !ok {
			// Check for named argument "name"
			if n, ok := args["name"]; ok {
				name = fmt.Sprintf("%v", n)
			} else {
				return nil, fmt.Errorf("doc() requires a name argument")
			}
		}

		// First, try to find as a module (same resolution as require())
		fullPath, _, err := resolver.ResolveModule(name)
		if err == nil && fullPath != "" {
			// Convert .du extension to .md
			docPath := strings.TrimSuffix(fullPath, ".du") + ".md"
			content, err := readFile(docPath)
			if err == nil {
				output := fmt.Sprintf("Documentation from: %s\n\n%s", docPath, string(content))
				return output, nil
			}
		}

		// If not a module, try reference documentation in docs/reference/
		refPath := "/EMBED/docs/reference/" + name + ".md"
		content, err := readFile(refPath)
		if err == nil {
			output := fmt.Sprintf("Documentation from: %s\n\n%s", refPath, string(content))
			return output, nil
		}

		// Not found anywhere
		return nil, nil
	}
}

// NewBreakpointFunction creates a breakpoint() function for debugging.
// In debug mode, this pauses execution and drops into an interactive session.
// In regular mode, it's a no-op.
//
// TODO: Implement full debug mode with step (n), next (s), and continue (c) commands
func NewBreakpointFunction(debugMode bool) func(map[string]any) (any, error) {
	return func(args map[string]any) (any, error) {
		if debugMode {
			// TODO: Signal to pause execution and enter debug REPL
			// For now, no-op in all modes
			fmt.Fprintf(os.Stderr, "[breakpoint] (debug mode not yet implemented)\n")
		}
		return nil, nil
	}
}

// NewMarkdownFunction creates a markdown(text) function that renders markdown to ANSI-formatted output.
//
// markdown() takes a markdown string and returns it formatted with ANSI color codes for terminal display.
// This is useful for rendering documentation, Claude responses, or any markdown content to the console.
// It's only available in the CLI environment.
//
// Example:
//     docs = doc("split")
//     print(markdown(docs))
//
//     response = claude("explain closures")
//     print(markdown(response))
func NewMarkdownFunction() func(map[string]any) (any, error) {
	return NewMarkdownFunctionWithOptions(false)
}

// NewMarkdownFunctionWithOptions creates a markdown function with optional color disabling.
func NewMarkdownFunctionWithOptions(noColor bool) func(map[string]any) (any, error) {
	return func(args map[string]any) (any, error) {
		text, ok := args["0"].(string)
		if !ok {
			// Check for named argument "text"
			if t, ok := args["text"]; ok {
				text = fmt.Sprintf("%v", t)
			} else {
				return nil, fmt.Errorf("markdown() requires a text argument")
			}
		}

		if noColor {
			return text, nil
		}

		formatted := markdown.ToANSI(text)
		return formatted, nil
	}
}
