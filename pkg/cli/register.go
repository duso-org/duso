// Package cli provides CLI-specific functions for Duso scripts.
// This file contains the main registration function.
package cli

import (
	"os"
	"path/filepath"

	"github.com/duso-org/duso/pkg/script"
)

// RegisterOptions configures how CLI functions are registered.
type RegisterOptions struct {
	ScriptDir string // Directory relative to which files are loaded/saved
	DebugMode bool   // Enable debug mode (breakpoint() pauses execution)
}

// NewModuleResolver creates a ModuleResolver from RegisterOptions.
// This is used internally by RegisterFunctions and can also be used by the CLI
// to handle doc() lookups before script execution.
func NewModuleResolver(opts RegisterOptions) *ModuleResolver {
	// Parse DUSO_LIB environment variable (colon-separated list of directories)
	dusoPath := []string{}
	if dusoPathEnv := os.Getenv("DUSO_LIB"); dusoPathEnv != "" {
		dusoPath = filepath.SplitList(dusoPathEnv)
	}

	return &ModuleResolver{
		ScriptDir: opts.ScriptDir,
		DusoPath:  dusoPath,
	}
}

// RegisterFunctions registers all CLI-specific functions (load, save, include, require)
// in the given interpreter.
//
// This is called automatically by the duso CLI in cmd/duso/main.go.
// Embedded Go applications can optionally call this to enable CLI features,
// or implement their own versions of these functions.
//
// Provides module loading via:
// - include(filename): Loads and executes scripts in current scope (variables leak)
// - require(moduleName): Loads modules in isolated scope (variables isolated, returns exports)
//
// Example (CLI usage - automatic):
//     // cmd/duso/main.go already calls this for you
//     interp := script.NewInterpreter(false)
//     cli.RegisterFunctions(interp, cli.RegisterOptions{ScriptDir: "/path/to/script"})
//
// Example (embedded usage - optional):
//     interp := script.NewInterpreter(false)
//     // Enable file I/O (optional)
//     cli.RegisterFunctions(interp, cli.RegisterOptions{ScriptDir: "."})
//     // Now scripts can use: load(), save(), include(), require()
func RegisterFunctions(interp *script.Interpreter, opts RegisterOptions) error {
	ctx := FileIOContext{ScriptDir: opts.ScriptDir}

	// Create module resolver for path resolution (both require and include)
	resolver := NewModuleResolver(opts)

	// Create circular dependency detector (both require and include)
	detector := &CircularDetector{
		stack: []string{},
	}

	// Register load(filename) - reads files
	interp.RegisterFunction("load", NewLoadFunction(ctx))

	// Register save(filename, content) - writes files
	interp.RegisterFunction("save", NewSaveFunction(ctx))

	// Register enhanced include(filename) - loads and executes scripts in current scope
	// With path resolution and circular dependency detection
	interp.RegisterFunction("include", NewIncludeFunction(resolver, detector, func(source string) error {
		_, err := interp.Execute(source)
		return err
	}))

	// Register require(moduleName) - loads modules in isolated scope with caching
	// With path resolution and circular dependency detection
	interp.RegisterFunction("require", NewRequireFunction(resolver, detector, interp))

	// Register doc(moduleName) - displays module documentation
	// Uses same path resolution as require()
	interp.RegisterFunction("doc", NewDocFunction(resolver))

	// Register markdown(text) - renders markdown to ANSI for terminal display
	interp.RegisterFunction("markdown", NewMarkdownFunction())

	// Register env(varname) - reads environment variables
	interp.RegisterFunction("env", NewEnvFunction())

	// Register http_client(config) - creates stateful HTTP client
	interp.RegisterFunction("http_client", NewHTTPClientFunction())

	// Register breakpoint() - pauses execution in debug mode, no-op otherwise
	interp.RegisterFunction("breakpoint", NewBreakpointFunction(opts.DebugMode))

	return nil
}
