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

import "fmt"

// Environment represents a scope for variables
type Environment struct {
	variables        map[string]Value
	parent           *Environment
	self             Value // For method calls - provides context for variable lookup
	isFunctionScope  bool  // If true, assignments don't walk up past this scope
}

// NewEnvironment creates a new root environment
func NewEnvironment() *Environment {
	return &Environment{
		variables: make(map[string]Value),
		parent:    nil,
		self:      NewNil(),
	}
}

// NewChildEnvironment creates a child environment with a parent scope
func NewChildEnvironment(parent *Environment) *Environment {
	return &Environment{
		variables:       make(map[string]Value),
		parent:          parent,
		self:            NewNil(),
		isFunctionScope: false,
	}
}

// NewChildEnvironmentWithSelf creates a child environment with a parent scope and self
func NewChildEnvironmentWithSelf(parent *Environment, self Value) *Environment {
	return &Environment{
		variables:       make(map[string]Value),
		parent:          parent,
		self:            self,
		isFunctionScope: false,
	}
}

// NewFunctionEnvironment creates a function scope that blocks variable assignment walk-up
func NewFunctionEnvironment(parent *Environment) *Environment {
	return &Environment{
		variables:       make(map[string]Value),
		parent:          parent,
		self:            NewNil(),
		isFunctionScope: true,
	}
}

// NewFunctionEnvironmentWithSelf creates a function scope with self binding
func NewFunctionEnvironmentWithSelf(parent *Environment, self Value) *Environment {
	return &Environment{
		variables:       make(map[string]Value),
		parent:          parent,
		self:            self,
		isFunctionScope: true,
	}
}

// Define creates a new variable in the current scope
func (e *Environment) Define(name string, value Value) {
	e.variables[name] = value
}

// Get retrieves a variable, walking up the parent chain if necessary
func (e *Environment) Get(name string) (Value, error) {
	if val, ok := e.variables[name]; ok {
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

// Set updates a variable, walking up the parent chain to find and modify existing variables
func (e *Environment) Set(name string, value Value) error {
	if _, ok := e.variables[name]; ok {
		e.variables[name] = value
		return nil
	}

	// Walk up to parent scope (even through function boundaries) to find existing variable
	if e.parent != nil {
		return e.parent.Set(name, value)
	}

	// If not found in any scope, define it in current scope (create locally)
	e.variables[name] = value
	return nil
}

// SetLocal updates a variable only in the current scope
func (e *Environment) SetLocal(name string, value Value) error {
	if _, ok := e.variables[name]; ok {
		e.variables[name] = value
		return nil
	}
	return fmt.Errorf("variable %s not defined in current scope", name)
}
