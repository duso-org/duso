package script

// FunctionCaller is an interface for invoking Duso functions and accessing evaluation context.
// This interface decouples Builtins from directly depending on Evaluator, allowing
// callback-based builtins (map, filter, reduce, etc.) to work without circular dependencies.
type FunctionCaller interface {
	// CallFunction calls a Duso function with the given arguments
	CallFunction(fn Value, args map[string]Value) (Value, error)

	// EvalTemplateLiteral evaluates a template string and returns the result
	EvalTemplateLiteral(template string) (string, error)

	// GetEnvironment returns the current evaluation environment
	GetEnvironment() *Environment

	// IsParallelContext returns true if executing in a parallel() block
	IsParallelContext() bool
}
