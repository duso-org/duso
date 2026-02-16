// Package cli provides CLI-specific functions for Duso scripts.
// This file contains the main registration function.
package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/duso-org/duso/pkg/script"
)

// Package-level globals set by RegisterFunctions for use by builtins
var (
	globalResolver  *ModuleResolver
	globalDetector  *CircularDetector
	globalInterpreter *script.Interpreter
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
	// Store the global interpreter reference for builtins to access
	globalInterpreter = interp

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

	// OutputWriter - outputs messages (no automatic newline)
	// Callers are responsible for adding newlines if needed
	// Automatically clears any active busy() spinner before writing
	if stdinServer != nil {
		interp.OutputWriter = stdinServer.GetOutputWriter()
	} else {
		interp.OutputWriter = func(msg string) error {
			ClearBusySpinner()
			_, err := fmt.Fprint(os.Stdout, msg)
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

	// Register CLI-specific builtins to global registry
	// TODO: Register CLI builtins (list_dir, make_dir, copy_file, etc.) via script.RegisterBuiltin()
	// when they are refactored from factories to standalone functions
	RegisterCLIBuiltins(resolver, detector)

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

// RegisterCLIBuiltins registers CLI-specific builtins to the global script registry.
func RegisterCLIBuiltins(resolver *ModuleResolver, detector *CircularDetector) {
	// Store resolver and detector for use by builtins that need them
	globalResolver = resolver
	globalDetector = detector

	// Console functions (CLI versions override runtime versions)
	script.RegisterBuiltin("print", builtinPrint)
	script.RegisterBuiltin("error", builtinError)
	script.RegisterBuiltin("write", builtinWrite)
	script.RegisterBuiltin("debug", builtinDebug)
	script.RegisterBuiltin("input", builtinInput)

	script.RegisterBuiltin("busy", builtinBusy)
	script.RegisterBuiltin("require", builtinRequire)
	script.RegisterBuiltin("include", builtinInclude)
	script.RegisterBuiltin("load", builtinLoad)
	script.RegisterBuiltin("save", builtinSave)
	script.RegisterBuiltin("list_dir", builtinListDir)
	script.RegisterBuiltin("list_files", builtinListFiles)
	script.RegisterBuiltin("make_dir", builtinMakeDir)
	script.RegisterBuiltin("remove_dir", builtinRemoveDir)
	script.RegisterBuiltin("rename_file", builtinRenameFile)
	script.RegisterBuiltin("file_type", builtinFileType)
	script.RegisterBuiltin("file_exists", builtinFileExists)
	script.RegisterBuiltin("current_dir", builtinCurrentDir)
	script.RegisterBuiltin("append_file", builtinAppendFile)
	script.RegisterBuiltin("copy_file", builtinCopyFile)
	script.RegisterBuiltin("move_file", builtinMoveFile)
	script.RegisterBuiltin("remove_file", builtinRemoveFile)
	script.RegisterBuiltin("doc", builtinDoc)
}
