package runtime

import (
	"fmt"
	"strings"

	"github.com/duso-org/duso/pkg/script"
)

// ModuleResolver interface for module path resolution.
// This allows different hosts to provide their own resolution logic.
type ModuleResolver interface {
	ResolveModule(moduleName string) (string, []string, error)
}

// CircularDetector interface for detecting circular module dependencies.
type CircularDetector interface {
	Push(path string) error
	Pop()
}

// NewRequireFunction creates a require(moduleName) function that loads modules.
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
func NewRequireFunction(resolver ModuleResolver, detector CircularDetector, interp *script.Interpreter) func(*script.Evaluator, map[string]any) (any, error) {
	return func(evaluator *script.Evaluator, args map[string]any) (any, error) {
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
		// This caches the result value, not the AST
		if cached, ok := interp.GetModuleCache(fullPath); ok {
			return script.ValueToInterface(cached), nil
		}

		// Check for circular dependency
		if err := detector.Push(fullPath); err != nil {
			return nil, err
		}
		defer detector.Pop()

		// Set file path context for error reporting
		prevPath := interp.GetFilePath()
		interp.SetFilePath(fullPath)
		defer interp.SetFilePath(prevPath)

		// Read script file using host's ScriptLoader capability
		if interp.ScriptLoader == nil {
			return nil, fmt.Errorf("require() requires ScriptLoader capability (not provided by host)")
		}

		fileBytes, err := interp.ScriptLoader(fullPath)
		if err != nil {
			return nil, fmt.Errorf("cannot require '%s': %w", fullPath, err)
		}

		// Parse the module script
		lexer := script.NewLexer(string(fileBytes))
		tokens := lexer.Tokenize()
		parser := script.NewParserWithFile(tokens, fullPath)
		program, err := parser.Parse()
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
}

// NewIncludeFunction creates an include(filename) function that executes other scripts.
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
func NewIncludeFunction(resolver ModuleResolver, detector CircularDetector, interp *script.Interpreter) func(*script.Evaluator, map[string]any) (any, error) {
	return func(evaluator *script.Evaluator, args map[string]any) (any, error) {
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

		// Set file path context for error reporting
		prevPath := interp.GetFilePath()
		interp.SetFilePath(fullPath)
		defer interp.SetFilePath(prevPath)

		// Read script file using host's ScriptLoader capability
		if interp.ScriptLoader == nil {
			return nil, fmt.Errorf("include() requires ScriptLoader capability (not provided by host)")
		}

		fileBytes, err := interp.ScriptLoader(fullPath)
		if err != nil {
			return nil, fmt.Errorf("cannot include '%s': %w", fullPath, err)
		}

		// Parse the included script
		lexer := script.NewLexer(string(fileBytes))
		tokens := lexer.Tokenize()
		parser := script.NewParserWithFile(tokens, fullPath)
		program, err := parser.Parse()
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
}
