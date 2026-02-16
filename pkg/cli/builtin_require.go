package cli

import (
	"fmt"
	"strings"

	"github.com/duso-org/duso/pkg/script"
)

// builtinRequire loads a module in an isolated scope and returns its exports.
//
// require() loads a module in an isolated scope and returns its exports.
// Unlike include(), require():
// - Executes the module in its own isolated scope
// - Returns the last expression value (the module's exports)
// - Caches results - subsequent requires return cached value without re-executing
//
// The AST is cached globally with mtime validation for hot reload during development.
// The module result is cached per-interpreter to allow concurrent evaluators to get
// fresh module instances while reusing the parsed AST.
//
// Example:
//
//	math = require("math")
//	result = math.add(2, 3)  // Calls function from module
//
// This function supports path resolution: user-provided paths, relative to script dir, and DUSO_LIB.
func builtinRequire(evaluator *Evaluator, args map[string]any) (any, error) {
	filename, ok := args["0"].(string)
	if !ok {
		// Check for named argument "filename"
		if f, ok := args["filename"]; ok {
			filename = fmt.Sprintf("%v", f)
		} else {
			return nil, fmt.Errorf("require() requires a filename argument")
		}
	}

	// Get the global interpreter (set by RegisterFunctions)
	if globalInterpreter == nil {
		return nil, fmt.Errorf("require() requires interpreter context")
	}
	interp := globalInterpreter

	// Resolve module path using standard resolution algorithm
	fullPath, searchedPaths, err := globalResolver.ResolveModule(filename)
	if err != nil {
		return nil, fmt.Errorf("cannot find module '%s'\nSearched:\n  %s",
			filename, strings.Join(searchedPaths, "\n  "))
	}

	// Check module cache (absolute path as key)
	// This caches the result value, not the AST
	if cached, ok := interp.GetModuleCache(fullPath); ok {
		return script.ValueToInterface(cached), nil
	}

	// Check for circular dependency
	if err := globalDetector.Push(fullPath); err != nil {
		return nil, err
	}
	defer globalDetector.Pop()

	// Set file path context for error reporting
	prevPath := interp.GetFilePath()
	interp.SetFilePath(fullPath)
	defer interp.SetFilePath(prevPath)

	// Parse the module script (with AST caching and mtime validation)
	readFileFunc := func(path string) ([]byte, error) {
		if interp.ScriptLoader == nil {
			return nil, fmt.Errorf("require() requires ScriptLoader capability")
		}
		return interp.ScriptLoader(path)
	}
	getMtimeFunc := func(path string) int64 {
		if interp.FileStatter == nil {
			return 0
		}
		return interp.FileStatter(path)
	}

	program, err := interp.ParseScriptFile(fullPath, readFileFunc, getMtimeFunc)
	if err != nil {
		return nil, fmt.Errorf("cannot parse module '%s': %w", fullPath, err)
	}

	// Execute in isolated scope using ExecuteModuleProgram to reuse evaluator logic
	value, err := interp.ExecuteModuleProgram(program)
	if err != nil {
		return nil, fmt.Errorf("error in module '%s': %w", fullPath, err)
	}

	// Cache the result
	interp.SetModuleCache(fullPath, value)

	// Convert to interface{} for return
	return script.ValueToInterface(value), nil
}

// builtinInclude loads and executes another .du script file in the current environment.
//
// include() loads and executes another .du script file in the current environment.
// Variables and functions defined in the included script are available after include().
// It's only available in the CLI environment.
//
// Unlike require(), include() executes in the current scope (not isolated),
// and results are not cached. However, the AST is cached globally with mtime validation
// for efficient reloading during development.
//
// Example:
//
//	include("helpers.du")
//	result = helper_function()  // Now available
//
// This function supports path resolution: user-provided paths, relative to script dir, and DUSO_LIB.
func builtinInclude(evaluator *Evaluator, args map[string]any) (any, error) {
	filename, ok := args["0"].(string)
	if !ok {
		// Check for named argument "filename"
		if f, ok := args["filename"]; ok {
			filename = fmt.Sprintf("%v", f)
		} else {
			return nil, fmt.Errorf("include() requires a filename argument")
		}
	}

	// Get the global interpreter (set by RegisterFunctions)
	if globalInterpreter == nil {
		return nil, fmt.Errorf("include() requires interpreter context")
	}
	interp := globalInterpreter

	// Resolve module path using standard resolution algorithm
	fullPath, searchedPaths, err := globalResolver.ResolveModule(filename)
	if err != nil {
		return nil, fmt.Errorf("cannot find module '%s'\nSearched:\n  %s",
			filename, strings.Join(searchedPaths, "\n  "))
	}

	// Check for circular dependency
	if err := globalDetector.Push(fullPath); err != nil {
		return nil, err
	}
	defer globalDetector.Pop()

	// Set file path context for error reporting
	prevPath := interp.GetFilePath()
	interp.SetFilePath(fullPath)
	defer interp.SetFilePath(prevPath)

	// Parse the included script (with AST caching and mtime validation)
	readFileFunc := func(path string) ([]byte, error) {
		if interp.ScriptLoader == nil {
			return nil, fmt.Errorf("include() requires ScriptLoader capability")
		}
		return interp.ScriptLoader(path)
	}
	getMtimeFunc := func(path string) int64 {
		if interp.FileStatter == nil {
			return 0
		}
		return interp.FileStatter(path)
	}

	program, err := interp.ParseScriptFile(fullPath, readFileFunc, getMtimeFunc)
	if err != nil {
		return nil, fmt.Errorf("cannot parse include '%s': %w", fullPath, err)
	}

	// Execute in current environment (no isolation)
	_, err = interp.EvalProgram(program)
	if err != nil {
		return nil, fmt.Errorf("error in included script '%s': %w", fullPath, err)
	}

	return nil, nil
}
