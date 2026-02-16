package runtime

import (
	"github.com/duso-org/duso/pkg/core"
	"github.com/duso-org/duso/pkg/script"
)

// Type aliases to avoid script. prefix throughout runtime builtins
type (
	Evaluator = script.Evaluator
	Environment = script.Environment
	Value = script.Value
	ValueRef = script.ValueRef
	GoFunction = script.GoFunction
	ScriptFunction = script.ScriptFunction
	CallFrame = script.CallFrame
	DusoError = script.DusoError
	BreakpointError = script.BreakpointError
)

// Value type constants
const (
	VAL_NIL = script.VAL_NIL
	VAL_NUMBER = script.VAL_NUMBER
	VAL_STRING = script.VAL_STRING
	VAL_BOOL = script.VAL_BOOL
	VAL_ARRAY = script.VAL_ARRAY
	VAL_OBJECT = script.VAL_OBJECT
	VAL_FUNCTION = script.VAL_FUNCTION
)

// Value constructors
var (
	NewNil = script.NewNil
	NewNumber = script.NewNumber
	NewString = script.NewString
	NewBool = script.NewBool
	NewArray = script.NewArray
	NewObject = script.NewObject
	NewGoFunction = script.NewGoFunction
	NewEnvironment = script.NewEnvironment
	NewEvaluator = script.NewEvaluator
	NewChildEnvironment = script.NewChildEnvironment
)

// Value conversion functions
var (
	InterfaceToValue = script.InterfaceToValue
	ValueToInterface = script.ValueToInterface
)

// Core utility functions
var (
	IsInteger = core.IsInteger
	DeepCopyAny = script.DeepCopyAny
)

// Exception types
type (
	ExitExecution = script.ExitExecution
)

// Registry functions
var (
	RegisterBuiltin = script.RegisterBuiltin
	CopyBuiltins = script.CopyBuiltins
)
