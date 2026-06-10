// registry.go - Global builtin function registry
//
// This file maintains the global registry of builtin functions.
// The registry is populated once at startup by the host (runtime package or CLI).
// Each evaluator gets a copy of the map for lock-free function lookups.
package script

// globalBuiltins holds all registered builtin functions
// Written once at startup, then only read via CopyBuiltins()
var globalBuiltins = make(map[string]GoFunction)

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
