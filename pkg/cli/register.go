// Package cli provides CLI-specific functions for Duso scripts.
// This file contains the main registration function.
package cli

import (
	"github.com/duso-org/duso/pkg/script"
)

// RegisterOptions configures how CLI functions are registered.
type RegisterOptions struct {
	ScriptDir string // Directory relative to which files are loaded/saved
}

// RegisterFunctions registers all CLI-specific functions (load, save, include, claude, conversation)
// in the given interpreter.
//
// This is called automatically by the duso CLI in cmd/duso/main.go.
// Embedded Go applications can optionally call this to enable CLI features,
// or implement their own versions of these functions.
//
// Example (CLI usage - automatic):
//     // cmd/duso/main.go already calls this for you
//     interp := script.NewInterpreter(false)
//     cli.RegisterFunctions(interp, cli.RegisterOptions{ScriptDir: "/path/to/script"})
//
// Example (embedded usage - optional):
//     interp := script.NewInterpreter(false)
//     // Enable file I/O and Claude integration (optional)
//     cli.RegisterFunctions(interp, cli.RegisterOptions{ScriptDir: "."})
//     // Now scripts can use: load(), save(), include(), claude(), conversation()
func RegisterFunctions(interp *script.Interpreter, opts RegisterOptions) error {
	ctx := FileIOContext{ScriptDir: opts.ScriptDir}

	// Register load(filename) - reads files
	interp.RegisterFunction("load", NewLoadFunction(ctx))

	// Register save(filename, content) - writes files
	interp.RegisterFunction("save", NewSaveFunction(ctx))

	// Register include(filename) - loads and executes scripts
	// This requires access to Execute, so we pass a closure
	interp.RegisterFunction("include", NewIncludeFunction(ctx, func(source string) error {
		_, err := interp.Execute(source)
		return err
	}))

	// Register claude() and conversation() - Claude API functions
	// Use the existing RegisterConversationAPI method on the Interpreter
	interp.RegisterConversationAPI()

	return nil
}
