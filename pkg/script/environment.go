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
//
// STORAGE: most scopes (function frames, loop bodies) hold only a handful of
// variables, so the first smallScopeSize variables live in inline arrays on the
// struct — creating and using such a scope costs a single allocation and lookups
// are a short linear scan (with Go's pointer-equality fast path for the shared
// AST identifier strings). Scopes that outgrow the inline slots spill to the
// overflow map.
//
// CONCURRENCY INVARIANT: Environments are deliberately unsynchronized. An env tree
// is only ever touched by one goroutine, with one exception: parallel() branches
// read parent scopes concurrently. That is safe because parallel() blocks the
// parent goroutine on wg.Wait() (no writer exists while readers run), and branch
// writes stop at the branch's own function env via isParallelContext. Anything
// that would share an env tree across goroutines in a new way must revisit this.
package script

import (
	"fmt"
)

// smallScopeSize is the number of variable slots stored inline in an Environment.
const smallScopeSize = 6

// Environment represents a scope for variables
type Environment struct {
	names    [smallScopeSize]string
	vals     [smallScopeSize]Value
	n        int              // live inline slots
	overflow map[string]Value // nil until a scope outgrows the inline slots

	parent          *Environment
	self            Value // For method calls - provides context for variable lookup
	isFunctionScope bool  // If true, assignments don't walk up past this scope

	// Parameter tracking optimization: store common single-letter parameters as bit flags
	// instead of a map to save memory (~250 bytes per env) while keeping lookups fast
	paramFlags uint64          // Bit flags for common parameter names (n, x, y, i, j, k, etc.)
	parameters map[string]bool // Map for uncommon parameter names (lazy allocated)

	isParallelContext bool // If true, assignments don't walk up to parent scope (for parallel() blocks)
}

// paramNameToFlag converts common parameter names to bit positions
// Returns (flag bit, isCommon). Common names use bits 0-31, uncommon go to map.
func paramNameToFlag(name string) (uint32, bool) {
	switch name {
	case "n":
		return 1 << 0, true // fib, recursion
	case "x":
		return 1 << 1, true // math, general
	case "y":
		return 1 << 2, true // math, general
	case "i":
		return 1 << 3, true // loops
	case "j":
		return 1 << 4, true // loops
	case "k":
		return 1 << 5, true // loops
	case "v":
		return 1 << 6, true // values
	case "val":
		return 1 << 7, true // values
	case "item":
		return 1 << 8, true // iteration
	case "key":
		return 1 << 9, true // iteration
	case "a":
		return 1 << 10, true // arrays
	case "b":
		return 1 << 11, true // binary operations
	case "fn":
		return 1 << 12, true // functions
	case "f":
		return 1 << 13, true // functions
	case "arg":
		return 1 << 14, true // arguments
	case "args":
		return 1 << 15, true // arguments
	default:
		return 0, false // Use map for other names
	}
}

// NewEnvironment creates a new root environment
func NewEnvironment() *Environment {
	return &Environment{self: NewNil()}
}

// NewChildEnvironment creates a child environment with a parent scope
func NewChildEnvironment(parent *Environment) *Environment {
	return &Environment{
		parent: parent,
		self:   NewNil(),
	}
}

// NewChildEnvironmentWithSelf creates a child environment with a parent scope and self
func NewChildEnvironmentWithSelf(parent *Environment, self Value) *Environment {
	return &Environment{
		parent: parent,
		self:   self,
	}
}

// NewFunctionEnvironment creates a function scope that blocks variable assignment walk-up
func NewFunctionEnvironment(parent *Environment) *Environment {
	return &Environment{
		parent:          parent,
		self:            NewNil(),
		isFunctionScope: true,
	}
}

// NewFunctionEnvironmentWithSelf creates a function scope with self binding
func NewFunctionEnvironmentWithSelf(parent *Environment, self Value) *Environment {
	return &Environment{
		parent:          parent,
		self:            self,
		isFunctionScope: true,
	}
}

// lookupLocal finds a variable in this scope only (inline slots, then overflow)
func (e *Environment) lookupLocal(name string) (Value, bool) {
	for i := 0; i < e.n; i++ {
		if e.names[i] == name {
			return e.vals[i], true
		}
	}
	if e.overflow != nil {
		if v, ok := e.overflow[name]; ok {
			return v, true
		}
	}
	return Value{}, false
}

// updateLocal updates a variable in this scope only, returning false if absent
func (e *Environment) updateLocal(name string, value Value) bool {
	for i := 0; i < e.n; i++ {
		if e.names[i] == name {
			e.vals[i] = value
			return true
		}
	}
	if e.overflow != nil {
		if _, ok := e.overflow[name]; ok {
			e.overflow[name] = value
			return true
		}
	}
	return false
}

// Define creates a new variable in the current scope
func (e *Environment) Define(name string, value Value) {
	if e.updateLocal(name, value) {
		return
	}
	if e.n < smallScopeSize {
		e.names[e.n] = name
		e.vals[e.n] = value
		e.n++
		return
	}
	if e.overflow == nil {
		e.overflow = make(map[string]Value)
	}
	e.overflow[name] = value
}

// Get retrieves a variable, walking up the parent chain if necessary
func (e *Environment) Get(name string) (Value, error) {
	// Check if accessing "self" directly
	if name == "self" && !e.self.IsNil() {
		return e.self, nil
	}

	if val, ok := e.lookupLocal(name); ok {
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
	if e.updateLocal(name, value) {
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
		e.Define(name, value)
		return nil
	}

	// Walk up to parent scope to find existing variable
	if e.parent != nil {
		return e.parent.Set(name, value)
	}

	// If not found in any scope, define it in current scope (create locally)
	e.Define(name, value)
	return nil
}

// SetLocal updates a variable only in the current scope
func (e *Environment) SetLocal(name string, value Value) error {
	if e.updateLocal(name, value) {
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
