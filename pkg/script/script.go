package script

import (
	"strings"
)

// Interpreter is the public API for executing Duso scripts.
//
// CORE INTERPRETER - This is suitable for both embedded Go applications and CLI usage.
// It uses only the core language runtime with no external dependencies.
//
// To extend with CLI features (file I/O, module loading), see pkg/cli/register.go
type Interpreter struct {
	evaluator   *Evaluator
	output      strings.Builder
	verbose     bool
	moduleCache map[string]Value // Cache for require() results, keyed by absolute path
}

// NewInterpreter creates a new interpreter instance.
//
// This creates a minimal interpreter with only the core Duso language features.
// Use this in embedded Go applications, then optionally register custom functions
// with RegisterFunction() or CLI features with pkg/cli.RegisterFunctions().
func NewInterpreter(verbose bool) *Interpreter {
	return &Interpreter{
		verbose:     verbose,
		moduleCache: make(map[string]Value),
	}
}

// RegisterFunction registers a custom Go function callable from Duso scripts.
//
// This is how embedded applications extend Duso with domain-specific functionality.
// For CLI-specific functions (load, save, include), see pkg/cli.
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

// ExecuteNode executes a single AST node.
// Used by debugger for statement-by-statement execution.
// Maintains evaluator state between calls.
func (i *Interpreter) ExecuteNode(node Node) error {
	if i.evaluator == nil {
		i.evaluator = NewEvaluator(&i.output)
	}
	_, err := i.evaluator.Eval(node)
	return err
}

// ExecuteFile executes a script file
func (i *Interpreter) ExecuteFile(path string) (string, error) {
	// Note: We don't have file I/O here - that's handled by the caller
	// This is a placeholder for future implementation
	return "", nil
}

// ExecuteModule executes script source in an isolated module scope and returns the result value.
// This is used by require() to load modules in isolation. The module's variables
// don't leak into the caller's scope. The last expression value (or explicit return) is the export.
func (i *Interpreter) ExecuteModule(source string) (Value, error) {
	if i.evaluator == nil {
		i.evaluator = NewEvaluator(&i.output)
	}

	// Tokenize
	lexer := NewLexer(source)
	tokens := lexer.Tokenize()

	// Parse
	parser := NewParser(tokens)
	program, err := parser.Parse()
	if err != nil {
		return NewNil(), err
	}

	// Evaluate in isolated scope
	return i.evaluator.EvalModule(program)
}

// GetOutput returns the captured output from print() calls
func (i *Interpreter) GetOutput() string {
	return i.output.String()
}

// GetModuleCache retrieves a cached module value by absolute path.
// Used by require() to implement module caching.
func (i *Interpreter) GetModuleCache(path string) (Value, bool) {
	val, ok := i.moduleCache[path]
	return val, ok
}

// SetModuleCache stores a module value in the cache by absolute path.
// Used by require() to cache module results so they're only loaded once.
func (i *Interpreter) SetModuleCache(path string, value Value) {
	i.moduleCache[path] = value
}

// Reset clears the output buffer and resets the environment
func (i *Interpreter) Reset() {
	i.output.Reset()
	i.evaluator = nil
}
