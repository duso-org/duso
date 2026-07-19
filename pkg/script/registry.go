// registry.go - Global builtin function registry
//
// This file maintains the global registry of builtin functions.
// The registry is populated once at startup by the host (runtime package or CLI).
// Each evaluator gets a copy of the map for lock-free function lookups.
package script

// globalBuiltins holds all registered builtin functions
// Written once at startup, then only read via CopyBuiltins()
var globalBuiltins = make(map[string]GoFunction)

// globalFastBuiltins holds fast-path variants of hot builtins. A fast builtin
// takes evaluated args as []Value directly, skipping the map[string]any
// marshalling round trip. A name registered here MUST also be registered as a
// regular builtin with identical semantics — the fast form is used only for
// direct positional calls; named args, indirect calls, and CallFunction take
// the map path.
var globalFastBuiltins = make(map[string]GoFunctionFast)

// RegisterBuiltinFast registers a fast-path variant for an existing builtin.
func RegisterBuiltinFast(name string, fn GoFunctionFast) {
	globalFastBuiltins[name] = fn
}

// CopyFastBuiltins returns a copy of the fast builtin registry.
func CopyFastBuiltins() map[string]GoFunctionFast {
	copy := make(map[string]GoFunctionFast, len(globalFastBuiltins))
	for name, fn := range globalFastBuiltins {
		copy[name] = fn
	}
	return copy
}

// RegisterBuiltin registers a builtin function in the global registry.
// This is called by the host (runtime package or CLI) during initialization.
func RegisterBuiltin(name string, fn GoFunction) {
	globalBuiltins[name] = fn
}

// CopyBuiltins returns a copy of the global builtin registry.
// Called once per evaluator so it can have lock-free lookups.
func CopyBuiltins() map[string]GoFunction {
	copy := make(map[string]GoFunction, len(globalBuiltins))
	for name, fn := range globalBuiltins {
		copy[name] = fn
	}
	return copy
}

// GetBuiltin retrieves a single builtin function by name, or nil if not found.
func GetBuiltin(name string) GoFunction {
	return globalBuiltins[name]
}

// GetAllBuiltinNames returns a sorted list of all registered builtin function names.
func GetAllBuiltinNames() []string {
	names := make([]string, 0, len(globalBuiltins))
	for name := range globalBuiltins {
		names = append(names, name)
	}
	return names
}
