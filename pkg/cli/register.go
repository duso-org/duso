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
	NoFiles   bool   // Disable filesystem access (only allow /STORE/ and /EMBED/)
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
	ctx := FileIOContext{ScriptDir: opts.ScriptDir, NoFiles: opts.NoFiles}

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
	interp.RegisterFunction("include", NewIncludeFunction(resolver, detector, interp))

	// Register require(moduleName) - loads modules in isolated scope with caching
	// With path resolution and circular dependency detection
	interp.RegisterFunction("require", NewRequireFunction(resolver, detector, interp))

	// Register doc(moduleName) - displays module documentation
	// Uses same path resolution as require()
	interp.RegisterFunction("doc", NewDocFunction(resolver))

	// Register env(varname) - reads environment variables
	interp.RegisterFunction("env", NewEnvFunction())

	// Register fetch(url, options) - make HTTP requests (JavaScript-style fetch API)
	interp.RegisterFunction("fetch", NewFetchFunction())

	// Register http_server(config) - creates stateful HTTP server
	interp.RegisterFunction("http_server", NewHTTPServerFunction(interp))

	// Register context() - provides request-scoped context in HTTP handlers
	interp.RegisterFunction("context", NewContextFunction())

	// Register spawn(script, context) - spawns script in background goroutine
	interp.RegisterFunction("spawn", NewSpawnFunction(interp))

	// Register run(script, context) - runs script synchronously
	interp.RegisterFunction("run", NewRunFunction(interp))

	// File operations (all CLI-only)
	interp.RegisterFunction("list_dir", NewListDirFunction(ctx))
	interp.RegisterFunction("list_files", NewListFilesFunction(ctx))
	interp.RegisterFunction("make_dir", NewMakeDirFunction(ctx))
	interp.RegisterFunction("remove_file", NewRemoveFileFunction(ctx))
	interp.RegisterFunction("remove_dir", NewRemoveDirFunction(ctx))
	interp.RegisterFunction("rename_file", NewRenameFileFunction(ctx))
	interp.RegisterFunction("file_type", NewFileTypeFunction(ctx))
	interp.RegisterFunction("file_exists", NewFileExistsFunction(ctx))
	interp.RegisterFunction("current_dir", NewCurrentDirFunction())
	interp.RegisterFunction("append_file", NewAppendFileFunction(ctx))
	interp.RegisterFunction("copy_file", NewCopyFileFunction(ctx))
	interp.RegisterFunction("move_file", NewMoveFileFunction(ctx))

	// Register datastore(namespace, config) - thread-safe in-memory key/value store
	interp.RegisterFunction("datastore", NewDatastoreFunction())

	// If in debug mode, register the console debug handler and start the listener
	if opts.DebugMode {
		handler := NewConsoleDebugHandler(interp)
		interp.RegisterDebugHandler(handler)

		// Start the debug event listener goroutine
		// This listens for debug events from all sources (main, spawn, run, HTTP)
		// and delegates to the registered handler
		go func() {
			for event := range interp.GetDebugEventChan() {
				if event != nil {
					// Call the registered handler
					debugHandler := interp.GetDebugHandler()
					if debugHandler != nil {
						debugHandler(event)
					}
				}
			}
		}()
	}

	return nil
}
