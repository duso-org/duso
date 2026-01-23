package script

import (
	"strings"
)

// Interpreter is the public API for executing Duso scripts.
//
// CORE INTERPRETER - This is suitable for both embedded Go applications and CLI usage.
// It uses only the core language runtime with no external dependencies.
//
// To extend with CLI features (file I/O, Claude API), see pkg/cli/register.go
type Interpreter struct {
	evaluator *Evaluator
	output    strings.Builder
	verbose   bool
}

// NewInterpreter creates a new interpreter instance.
//
// This creates a minimal interpreter with only the core Duso language features.
// Use this in embedded Go applications, then optionally register custom functions
// with RegisterFunction() or CLI features with pkg/cli.RegisterFunctions().
func NewInterpreter(verbose bool) *Interpreter {
	return &Interpreter{
		verbose: verbose,
	}
}

// RegisterFunction registers a custom Go function callable from Duso scripts.
//
// This is how embedded applications extend Duso with domain-specific functionality.
// For CLI-specific functions (load, save, include, claude, conversation), see pkg/cli.
func (i *Interpreter) RegisterFunction(name string, fn GoFunction) error {
	if i.evaluator == nil {
		i.evaluator = NewEvaluator(&i.output)
	}
	i.evaluator.RegisterFunction(name, fn)
	return nil
}

// RegisterObject registers an object with methods (e.g., "agents" with methods like "classify")
func (i *Interpreter) RegisterObject(name string, methods map[string]GoFunction) error {
	if i.evaluator == nil {
		i.evaluator = NewEvaluator(&i.output)
	}
	i.evaluator.RegisterObject(name, methods)

	// Create a wrapper object that allows method calls
	objMethods := make(map[string]Value)
	for methodName, fn := range methods {
		objMethods[methodName] = NewGoFunction(fn)
	}

	// Register as an object in the environment
	objVal := NewObject(make(map[string]Value))
	i.evaluator.env.Define(name, objVal)

	// Actually, we need to handle object method calls differently
	// For now, register each method as "object.method"
	for methodName, fn := range methods {
		fullName := name + "." + methodName
		i.evaluator.RegisterFunction(fullName, fn)
	}

	return nil
}

// Execute executes script source code
func (i *Interpreter) Execute(source string) (string, error) {
	if i.evaluator == nil {
		i.evaluator = NewEvaluator(&i.output)
	}

	// Tokenize
	lexer := NewLexer(source)
	tokens := lexer.Tokenize()

	if i.verbose {
		// Uncomment for debugging
		// for _, tok := range tokens {
		//     fmt.Printf("%v\n", tok)
		// }
	}

	// Parse
	parser := NewParser(tokens)
	program, err := parser.Parse()
	if err != nil {
		return "", err
	}

	// Evaluate
	_, err = i.evaluator.Eval(program)
	if err != nil {
		return "", err
	}

	return i.GetOutput(), nil
}

// ExecuteFile executes a script file
func (i *Interpreter) ExecuteFile(path string) (string, error) {
	// Note: We don't have file I/O here - that's handled by the caller
	// This is a placeholder for future implementation
	return "", nil
}

// GetOutput returns the captured output from print() calls
func (i *Interpreter) GetOutput() string {
	return i.output.String()
}

// Reset clears the output buffer and resets the environment
func (i *Interpreter) Reset() {
	i.output.Reset()
	i.evaluator = nil
}

// RegisterConversationAPI registers the conversation() and claude() functions for Claude API access.
//
// DEPRECATED - CLI FEATURE: These functions are Claude API-specific, not part of the core language.
// For normal CLI usage, use pkg/cli.RegisterFunctions() instead, which handles all CLI features.
//
// This method is kept for backward compatibility. New code should use:
//     cli.RegisterFunctions(interp, cli.RegisterOptions{ScriptDir: "."})
//
// If you're embedding Duso and want Claude support, you can either:
// 1. Call this method directly, or
// 2. Use pkg/cli.RegisterFunctions() for all CLI features at once
func (i *Interpreter) RegisterConversationAPI() {
	if i.evaluator == nil {
		i.evaluator = NewEvaluator(&i.output)
	}
	RegisterConversationAPI(i.evaluator.env)
}
