// Package cli provides CLI-specific functions for Duso scripts.
// This file contains the main registration function.
package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/duso-org/duso/pkg/runtime"
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
// The optional stdinServer parameter enables HTTP stdin/stdout transport.
// If provided, script input/output is exposed over HTTP instead of the console.
//
// Example (CLI usage - automatic):
//     // cmd/duso/main.go already calls this for you
//     interp := script.NewInterpreter(false)
//     cli.RegisterFunctions(interp, cli.RegisterOptions{ScriptDir: "/path/to/script"}, nil)
//
// Example (embedded usage - optional):
//     interp := script.NewInterpreter(false)
//     // Enable file I/O (optional)
//     cli.RegisterFunctions(interp, cli.RegisterOptions{ScriptDir: "."}, nil)
//     // Now scripts can use: load(), save(), include(), require()
//
// Example (with HTTP stdin/stdout):
//     server := cli.NewStdinHTTPServer(9999, "localhost")
//     go server.Start()
//     cli.RegisterFunctions(interp, opts, server)
func RegisterFunctions(interp *script.Interpreter, opts RegisterOptions, stdinServer *StdinHTTPServer) error {
	ctx := FileIOContext{ScriptDir: opts.ScriptDir, NoFiles: opts.NoFiles}

	// Create module resolver for path resolution (both require and include)
	resolver := NewModuleResolver(opts)

	// Create circular dependency detector (both require and include)
	detector := &CircularDetector{
		stack: []string{},
	}

	// Set up host-provided capabilities for builtins
	// These are used by spawn/run and file operations

	// ScriptLoader - wraps existing ReadScriptWithFallback for spawn/run
	interp.ScriptLoader = func(path string) ([]byte, error) {
		return ReadScriptWithFallback(path, opts.ScriptDir)
	}

	// FileReader - reads files using the underlying file I/O utility
	interp.FileReader = readFile

	// FileWriter - writes files using the underlying file I/O utility
	interp.FileWriter = func(path, content string) error {
		return writeFile(path, []byte(content), 0644)
	}

	// FileStatter - gets file modification time for caching
	interp.FileStatter = getFileMtime

	// OutputWriter - outputs messages for print/error/debug
	if stdinServer != nil {
		interp.OutputWriter = stdinServer.GetOutputWriter()
	} else {
		interp.OutputWriter = func(msg string) error {
			_, err := fmt.Fprintln(os.Stdout, msg)
			return err
		}
	}

	// InputReader - reads input from user with prompt
	if stdinServer != nil {
		interp.InputReader = stdinServer.GetInputReader()
	} else {
		interp.InputReader = readInputLine
	}

	// EnvReader - reads environment variables
	interp.EnvReader = func(varname string) string {
		return os.Getenv(varname)
	}

	// NoFiles - restriction on file access
	interp.NoFiles = opts.NoFiles

	// Register load(filename) - reads files
	// Implemented in pkg/runtime with capability injection
	interp.RegisterFunction("load", runtime.NewLoadFunction(interp))

	// Register save(filename, content) - writes files
	// Implemented in pkg/runtime with capability injection
	interp.RegisterFunction("save", runtime.NewSaveFunction(interp))

	// Register enhanced include(filename) - loads and executes scripts in current scope
	// With path resolution and circular dependency detection
	// Implemented in pkg/runtime with capability injection
	interp.RegisterFunction("include", runtime.NewIncludeFunction(resolver, detector, interp))

	// Register require(moduleName) - loads modules in isolated scope with caching
	// With path resolution and circular dependency detection
	// Implemented in pkg/runtime with capability injection
	interp.RegisterFunction("require", runtime.NewRequireFunction(resolver, detector, interp))

	// Register doc(moduleName) - displays module documentation
	// Uses same path resolution as require()
	interp.RegisterFunction("doc", NewDocFunction(resolver))

	// Register env(varname) - reads environment variables
	// Implemented in pkg/runtime with capability injection
	interp.RegisterFunction("env", runtime.NewEnvFunction(interp))

	// Register fetch(url, options) - make HTTP requests (JavaScript-style fetch API)
	// Implemented in pkg/runtime
	interp.RegisterFunction("fetch", runtime.NewFetchFunction(interp))

	// Register http_server(config) - creates stateful HTTP server
	// Implemented in pkg/runtime with capability injection
	interp.RegisterFunction("http_server", runtime.NewHTTPServerFunction(interp))

	// Register context() - provides request-scoped context in HTTP handlers
	// Implemented in pkg/runtime
	interp.RegisterFunction("context", runtime.NewContextFunction())

	// Register spawn(script, context) - spawns script in background goroutine
	// Implemented in pkg/runtime with capability injection
	interp.RegisterFunction("spawn", runtime.NewSpawnFunction(interp))

	// Register run(script, context) - runs script synchronously
	// Implemented in pkg/runtime with capability injection
	interp.RegisterFunction("run", runtime.NewRunFunction(interp))

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

	// Register output functions - override core versions to use OutputWriter capability
	// Implemented in pkg/runtime with capability injection
	interp.RegisterFunction("print", runtime.NewPrintFunction(interp))
	interp.RegisterFunction("error", runtime.NewErrorFunction(interp))
	interp.RegisterFunction("debug", runtime.NewDebugFunction(interp))

	// Register input() - override core version to use InputReader capability
	// Implemented in pkg/runtime with capability injection
	interp.RegisterFunction("input", runtime.NewInputFunction(interp))

	// Register datastore(namespace, config) - thread-safe in-memory key/value store
	// Implemented in pkg/runtime
	interp.RegisterFunction("datastore", runtime.NewDatastoreFunction())

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
