// value.go - Duso runtime type system
//
// This file defines the core value types and runtime representation for all Duso data.
// Every value computed during script execution is represented as a Value struct.
//
// CORE LANGUAGE COMPONENT: This is part of the minimal core runtime.
// All values in Duso scripts map to one of these types:
// - NIL: Absence of a value (uninitialized variables)
// - NUMBER: Floating-point numbers (no integer type)
// - STRING: Text values
// - BOOL: True/false
// - ARRAY: Ordered lists of values (indexed by numbers)
// - OBJECT: Maps/tables (key-value pairs with string keys)
// - FUNCTION: Callable functions (either Go or Duso functions)
//
// This type system is simple and dynamically typed to match Duso's design goal
// of being easy to embed and learn.
package script

import (
	"fmt"
	"strconv"
	"strings"
)

// CodeValue represents pre-parsed code (source + AST + optional metadata)
type CodeValue struct {
	Source   string
	Program  *Program              // parsed AST, immutable
	Metadata map[string]Value      // optional user metadata from parse(src, meta)
}

// ErrorValue represents a first-class error value (message + stack trace string)
type ErrorValue struct {
	Message Value  // the value passed to throw(), or runtime error message string
	Stack   string // formatted string: file:line:col + call stack
}

type ValueType int

const (
	VAL_NIL ValueType = iota
	VAL_NUMBER
	VAL_STRING
	VAL_BOOL
	VAL_ARRAY
	VAL_OBJECT
	VAL_FUNCTION
	VAL_CODE   // pre-parsed code (source + AST + metadata)
	VAL_ERROR  // first-class error value (message + stack string)
)

// String returns a human-readable name for the ValueType
func (vt ValueType) String() string {
	switch vt {
	case VAL_NIL:
		return "nil"
	case VAL_NUMBER:
		return "number"
	case VAL_STRING:
		return "string"
	case VAL_BOOL:
		return "bool"
	case VAL_ARRAY:
		return "array"
	case VAL_OBJECT:
		return "object"
	case VAL_FUNCTION:
		return "function"
	case VAL_CODE:
		return "code"
	case VAL_ERROR:
		return "error"
	default:
		return "unknown"
	}
}

type GoFunction func(evaluator *Evaluator, args map[string]any) (any, error)

// ValueRef wraps a Value so it can pass through the any interface without losing type info
type ValueRef struct {
	Val Value
}

type Value struct {
	Type ValueType
	Data any
}

// Constructors
func NewNil() Value {
	return Value{Type: VAL_NIL, Data: nil}
}

func NewNumber(n float64) Value {
	return Value{Type: VAL_NUMBER, Data: n}
}

func NewString(s string) Value {
	return Value{Type: VAL_STRING, Data: s}
}

func NewBool(b bool) Value {
	return Value{Type: VAL_BOOL, Data: b}
}

func NewArray(elements []Value) Value {
	return Value{Type: VAL_ARRAY, Data: &elements}
}

func NewObject(obj map[string]Value) Value {
	return Value{Type: VAL_OBJECT, Data: obj}
}

func NewFunction(fn *ScriptFunction) Value {
	return Value{Type: VAL_FUNCTION, Data: fn}
}

func NewGoFunction(fn GoFunction) Value {
	return Value{Type: VAL_FUNCTION, Data: fn}
}

func NewCode(src string, prog *Program, meta map[string]Value) Value {
	return Value{Type: VAL_CODE, Data: &CodeValue{Source: src, Program: prog, Metadata: meta}}
}

func NewErrorValue(msg Value, stack string) Value {
	return Value{Type: VAL_ERROR, Data: &ErrorValue{Message: msg, Stack: stack}}
}

type ScriptFunction struct {
	Name       string
	FilePath   string        // File where function was defined (for error reporting)
	Parameters []*Parameter
	Body       []Node
	Closure    *Environment
}

// Type checking
func (v Value) IsNil() bool {
	return v.Type == VAL_NIL
}

func (v Value) IsNumber() bool {
	return v.Type == VAL_NUMBER
}

func (v Value) IsString() bool {
	return v.Type == VAL_STRING
}

func (v Value) IsBool() bool {
	return v.Type == VAL_BOOL
}

func (v Value) IsArray() bool {
	return v.Type == VAL_ARRAY
}

func (v Value) IsObject() bool {
	return v.Type == VAL_OBJECT
}

func (v Value) IsFunction() bool {
	return v.Type == VAL_FUNCTION
}

func (v Value) IsCode() bool {
	return v.Type == VAL_CODE
}

func (v Value) IsError() bool {
	return v.Type == VAL_ERROR
}

// Getters
func (v Value) AsNumber() float64 {
	if v.Type == VAL_NUMBER {
		return v.Data.(float64)
	}
	return 0
}

func (v Value) AsString() string {
	if v.Type == VAL_STRING {
		return v.Data.(string)
	}
	return ""
}

func (v Value) AsBool() bool {
	if v.Type == VAL_BOOL {
		return v.Data.(bool)
	}
	return false
}

func (v Value) AsArray() []Value {
	if v.Type == VAL_ARRAY {
		arrPtr := v.Data.(*[]Value)
		return *arrPtr
	}
	return nil
}

// AsArrayPtr returns a pointer to the array for in-place mutations
func (v Value) AsArrayPtr() *[]Value {
	if v.Type == VAL_ARRAY {
		return v.Data.(*[]Value)
	}
	return nil
}

func (v Value) AsObject() map[string]Value {
	if v.Type == VAL_OBJECT {
		return v.Data.(map[string]Value)
	}
	return nil
}

func (v Value) AsCode() *CodeValue {
	if v.Type == VAL_CODE {
		return v.Data.(*CodeValue)
	}
	return nil
}

func (v Value) AsErrorVal() *ErrorValue {
	if v.Type == VAL_ERROR {
		return v.Data.(*ErrorValue)
	}
	return nil
}

// Truthiness
func (v Value) IsTruthy() bool {
	switch v.Type {
	case VAL_NIL:
		return false
	case VAL_BOOL:
		return v.Data.(bool)
	case VAL_NUMBER:
		return v.Data.(float64) != 0
	case VAL_STRING:
		return v.Data.(string) != ""
	case VAL_ARRAY:
		arr := v.AsArray()
		return len(arr) > 0
	case VAL_OBJECT:
		return len(v.Data.(map[string]Value)) > 0
	default:
		return true // Functions, code, error remain truthy
	}
}

// String representation
func (v Value) String() string {
	switch v.Type {
	case VAL_NIL:
		return "nil"
	case VAL_NUMBER:
		n := v.Data.(float64)
		if n == float64(int64(n)) {
			return fmt.Sprintf("%d", int64(n))
		}
		return strconv.FormatFloat(n, 'f', -1, 64)
	case VAL_STRING:
		return v.Data.(string)
	case VAL_BOOL:
		if v.Data.(bool) {
			return "true"
		}
		return "false"
	case VAL_ARRAY:
		arr := v.AsArray()
		var builder strings.Builder
		builder.WriteString("[")
		for i, item := range arr {
			if i > 0 {
				builder.WriteString(", ")
			}
			builder.WriteString(item.String())
		}
		builder.WriteString("]")
		return builder.String()
	case VAL_OBJECT:
		obj := v.Data.(map[string]Value)
		var builder strings.Builder
		builder.WriteString("{")
		first := true
		for k, v := range obj {
			if !first {
				builder.WriteString(", ")
			}
			builder.WriteString(k)
			builder.WriteString(" = ")
			builder.WriteString(v.String())
			first = false
		}
		builder.WriteString("}")
		return builder.String()
	case VAL_FUNCTION:
		return "<function>"
	case VAL_CODE:
		return "<code>"
	case VAL_ERROR:
		ev := v.AsErrorVal()
		if ev != nil {
			return ev.Message.String()
		}
		return "<error>"
	default:
		return "unknown"
	}
}

// ValueToInterface converts a Value to interface{} for Go interop.
// This is used to convert script values to Go types for external functions.
// For arrays, returns *[]Value directly to allow in-place mutations by builtins.
func ValueToInterface(v Value) any {
	switch v.Type {
	case VAL_NIL:
		return nil
	case VAL_NUMBER:
		return v.Data.(float64)
	case VAL_STRING:
		return v.Data.(string)
	case VAL_BOOL:
		return v.Data.(bool)
	case VAL_ARRAY:
		// Return pointer directly for in-place mutations
		return v.Data.(*[]Value)
	case VAL_OBJECT:
		obj := v.Data.(map[string]Value)
		result := make(map[string]any)
		for k, val := range obj {
			result[k] = ValueToInterface(val)
		}
		return result
	case VAL_FUNCTION:
		return &ValueRef{Val: v} // Wrap function so it survives the any conversion
	case VAL_CODE:
		return &ValueRef{Val: v} // Wrap code value so it survives the any conversion
	case VAL_ERROR:
		return &ValueRef{Val: v} // Wrap error value so it survives the any conversion
	default:
		return nil
	}
}

// valueToInterface is the internal version - kept for backward compatibility
func valueToInterface(v Value) any {
	return ValueToInterface(v)
}

// interfaceToValue is the internal version - kept for backward compatibility
func interfaceToValue(i any) Value {
	return InterfaceToValue(i)
}

// InterfaceToValue converts Go any to script values.
// This is used to convert Go values to script Values for builtins.
func InterfaceToValue(i any) Value {
	if i == nil {
		return NewNil()
	}

	// If it's a ValueRef, unwrap it
	if vr, ok := i.(*ValueRef); ok {
		return vr.Val
	}

	// If it's already a Value, return it directly
	if v, ok := i.(Value); ok {
		return v
	}

	switch v := i.(type) {
	case float64:
		return NewNumber(v)
	case int:
		return NewNumber(float64(v))
	case int64:
		return NewNumber(float64(v))
	case string:
		return NewString(v)
	case bool:
		return NewBool(v)
	case []any:
		arr := make([]Value, len(v))
		for i, item := range v {
			arr[i] = InterfaceToValue(item)
		}
		return NewArray(arr)
	case *[]Value:
		// Already a pointer array, just wrap it
		return Value{Type: VAL_ARRAY, Data: v}
	case map[string]any:
		obj := make(map[string]Value)
		for k, val := range v {
			obj[k] = InterfaceToValue(val)
		}
		return NewObject(obj)
	default:
		return NewNil()
	}
}

// DeepCopy creates a deep copy of a Value, recursively copying arrays and objects
func DeepCopy(v Value) Value {
	switch v.Type {
	case VAL_ARRAY:
		arr := v.AsArray()
		newArr := make([]Value, len(arr))
		for i, elem := range arr {
			newArr[i] = DeepCopy(elem)
		}
		return NewArray(newArr)
	case VAL_OBJECT:
		obj := v.AsObject()
		newObj := make(map[string]Value, len(obj))
		for k, val := range obj {
			newObj[k] = DeepCopy(val)
		}
		return NewObject(newObj)
	case VAL_CODE, VAL_ERROR:
		// Code and error values are immutable, return as-is
		return v
	default:
		// Primitives are immutable
		return v
	}
}

// DeepCopyAny performs deep copy on any type (for scope boundaries)
func DeepCopyAny(val any) any {
	switch v := val.(type) {
	case []any:
		newArr := make([]any, len(v))
		for i, elem := range v {
			newArr[i] = DeepCopyAny(elem)
		}
		return newArr
	case map[string]any:
		newObj := make(map[string]any, len(v))
		for k, elem := range v {
			newObj[k] = DeepCopyAny(elem)
		}
		return newObj
	default:
		return v
	}
}
