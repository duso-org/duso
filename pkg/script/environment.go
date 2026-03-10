// environment.go - Duso variable scoping system
//
// This file implements the lexical scoping and variable lookup system for Duso.
// An Environment is a single scope level, with optional parent scopes forming a scope chain.
//
// CORE LANGUAGE COMPONENT: This is part of the minimal core runtime.
// Scope management is essential for:
// - Local variables in functions
// - Nested scopes (if blocks, function bodies, loops)
// - Variable shadowing (redefining in inner scopes)
// - Function closure support
//
// The scoping model is simple and dynamically typed:
// - Variables are stored as Value structs
// - Each environment has an optional parent environment
// - Variable lookup walks up the scope chain
// - Function scopes prevent assignments from walking to parent (local declarations)
// - The "self" value provides context for method calls
package script

import (
	"fmt"
	// "sync"  // Unused - locks removed after fixing parallel context propagation
)

// Environment represents a scope for variables
type Environment struct {
	variables        map[string]Value
	// mu               sync.RWMutex // Protects concurrent access to variables
	// REMOVED: No concurrent access to same environment - each evaluator/goroutine has isolated chain.
	// parallel() creates read-only parent access with isParallelContext flag preventing writes.
	parent           *Environment
	self             Value // For method calls - provides context for variable lookup
	isFunctionScope  bool  // If true, assignments don't walk up past this scope

	// Parameter tracking optimization: store common single-letter parameters as bit flags
	// instead of a map to save memory (~250 bytes per env) while keeping lookups fast
	paramFlags       uint64 // Bit flags for common parameter names (n, x, y, i, j, k, etc.)
	parameters       map[string]bool // Map for uncommon parameter names (lazy allocated)

	isParallelContext bool // If true, assignments don't walk up to parent scope (for parallel() blocks)
}

// paramNameToFlag converts common parameter names to bit positions
// Returns (flag bit, isCommon). Common names use bits 0-31, uncommon go to map.
func paramNameToFlag(name string) (uint32, bool) {
	switch name {
	case "n":      return 1 << 0, true   // fib, recursion
	case "x":      return 1 << 1, true   // math, general
	case "y":      return 1 << 2, true   // math, general
	case "i":      return 1 << 3, true   // loops
	case "j":      return 1 << 4, true   // loops
	case "k":      return 1 << 5, true   // loops
	case "v":      return 1 << 6, true   // values
	case "val":    return 1 << 7, true   // values
	case "item":   return 1 << 8, true   // iteration
	case "key":    return 1 << 9, true   // iteration
	case "a":      return 1 << 10, true  // arrays
	case "b":      return 1 << 11, true  // binary operations
	case "fn":     return 1 << 12, true  // functions
	case "f":      return 1 << 13, true  // functions
	case "arg":    return 1 << 14, true  // arguments
	case "args":   return 1 << 15, true  // arguments
	default:       return 0, false       // Use map for other names
	}
}

// NewEnvironment creates a new root environment
func NewEnvironment() *Environment {
	return &Environment{
		variables:  make(map[string]Value),
		parent:     nil,
		self:       NewNil(),
		parameters: make(map[string]bool),
	}
}

// NewChildEnvironment creates a child environment with a parent scope
func NewChildEnvironment(parent *Environment) *Environment {
	return &Environment{
		variables:       make(map[string]Value),
		parent:          parent,
		self:            NewNil(),
		isFunctionScope: false,
		parameters:      make(map[string]bool),
	}
}

// NewChildEnvironmentWithSelf creates a child environment with a parent scope and self
func NewChildEnvironmentWithSelf(parent *Environment, self Value) *Environment {
	return &Environment{
		variables:       make(map[string]Value),
		parent:          parent,
		self:            self,
		isFunctionScope: false,
		parameters:      make(map[string]bool),
	}
}

// NewFunctionEnvironment creates a function scope that blocks variable assignment walk-up
func NewFunctionEnvironment(parent *Environment) *Environment {
	return &Environment{
		variables:       make(map[string]Value),
		parent:          parent,
		self:            NewNil(),
		isFunctionScope: true,
		parameters:      make(map[string]bool),
	}
}

// NewFunctionEnvironmentWithSelf creates a function scope with self binding
func NewFunctionEnvironmentWithSelf(parent *Environment, self Value) *Environment {
	return &Environment{
		variables:       make(map[string]Value),
		parent:          parent,
		self:            self,
		isFunctionScope: true,
		parameters:      make(map[string]bool),
	}
}

// Define creates a new variable in the current scope
func (e *Environment) Define(name string, value Value) {
	// e.mu.Lock()
	e.variables[name] = value
	// e.mu.Unlock()
}

// Get retrieves a variable, walking up the parent chain if necessary
func (e *Environment) Get(name string) (Value, error) {
	// e.mu.RLock()
	val, ok := e.variables[name]
	// e.mu.RUnlock()
	if ok {
		return val, nil
	}

	// If self exists and is an object, check its properties
	if !e.self.IsNil() && e.self.IsObject() {
		objMap := e.self.AsObject()
		if val, ok := objMap[name]; ok {
			return val, nil
		}
	}

	if e.parent != nil {
		return e.parent.Get(name)
	}

	return NewNil(), fmt.Errorf("undefined variable: %s", name)
}

// Set updates a variable, checking self properties first, then walking up the parent chain
// Parallel context blocks assignment walk-up to parent: parent scope becomes read-only
func (e *Environment) Set(name string, value Value) error {
	// e.mu.Lock()
	_, ok := e.variables[name]
	// e.mu.Unlock()

	if ok {
		// e.mu.Lock()
		e.variables[name] = value
		// e.mu.Unlock()
		return nil
	}

	// If self exists and is an object, check and update its properties
	if !e.self.IsNil() && e.self.IsObject() {
		objMap := e.self.AsObject()
		if _, ok := objMap[name]; ok {
			objMap[name] = value
			return nil
		}
	}

	// If we're in a parallel context, don't allow walks to parent scope
	// Create locally instead to prevent race conditions
	if e.isParallelContext {
		// e.mu.Lock()
		e.variables[name] = value
		// e.mu.Unlock()
		return nil
	}

	// Walk up to parent scope to find existing variable
	if e.parent != nil {
		return e.parent.Set(name, value)
	}

	// If not found in any scope, define it in current scope (create locally)
	// e.mu.Lock()
	e.variables[name] = value
	// e.mu.Unlock()
	return nil
}

// SetLocal updates a variable only in the current scope
func (e *Environment) SetLocal(name string, value Value) error {
	// e.mu.Lock()
	// defer e.mu.Unlock()

	if _, ok := e.variables[name]; ok {
		e.variables[name] = value
		return nil
	}
	return fmt.Errorf("variable %s not defined in current scope", name)
}

// MarkParameter marks a name as a function parameter (can't be shadowed with var)
// Uses bit flags for common names, map for uncommon (memory optimization)
func (e *Environment) MarkParameter(name string) {
	if flag, isCommon := paramNameToFlag(name); isCommon {
		e.paramFlags |= uint64(flag)
	} else {
		// Lazy allocate map only for uncommon parameter names
		if e.parameters == nil {
			e.parameters = make(map[string]bool)
		}
		e.parameters[name] = true
	}
}

// IsParameter checks if a name is a function parameter
// Fast path for common names (bit test), slow path for uncommon (map lookup)
func (e *Environment) IsParameter(name string) bool {
	if flag, isCommon := paramNameToFlag(name); isCommon {
		return (e.paramFlags & uint64(flag)) != 0
	}
	if e.parameters == nil {
		return false
	}
	return e.parameters[name]
}

// SetParallelContext marks this environment as part of a parallel() block
// When true, assignments don't walk up to parent scope (parent scope is read-only)
func (e *Environment) SetParallelContext(isParallel bool) {
	e.isParallelContext = isParallel
}
