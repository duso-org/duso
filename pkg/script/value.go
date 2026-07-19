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
	"regexp"
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

// BinaryValue represents immutable binary data (e.g., files, images)
type BinaryValue struct {
	Data     *[]byte           // Pointer to immutable binary data
	Metadata map[string]Value  // filename, content_type, size, etc.
}

// RegexValue represents a compiled regular expression pattern
type RegexValue struct {
	Pattern string        // Original pattern source
	Compiled *regexp.Regexp // Compiled regex
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
	VAL_CODE    // pre-parsed code (source + AST + metadata)
	VAL_ERROR   // first-class error value (message + stack string)
	VAL_BINARY  // immutable binary data (files, images, etc.)
	VAL_REGEX   // compiled regular expression pattern
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
	case VAL_BINARY:
		return "binary"
	case VAL_REGEX:
		return "regex"
	default:
		return "unknown"
	}
}

type GoFunction func(evaluator *Evaluator, args map[string]any) (any, error)

// GoFunctionFast is the fast-path builtin signature: evaluated positional args
// in, Value out, no interface{} marshalling. See RegisterBuiltinFast.
type GoFunctionFast func(evaluator *Evaluator, args []Value) (Value, error)

// argKeys caches positional-argument map keys so hot call paths avoid fmt.Sprintf
var argKeys = [32]string{
	"0", "1", "2", "3", "4", "5", "6", "7",
	"8", "9", "10", "11", "12", "13", "14", "15",
	"16", "17", "18", "19", "20", "21", "22", "23",
	"24", "25", "26", "27", "28", "29", "30", "31",
}

// ArgKey returns the args-map key for positional argument i
func ArgKey(i int) string {
	if i >= 0 && i < len(argKeys) {
		return argKeys[i]
	}
	return strconv.Itoa(i)
}

// ValueRef wraps a Value so it can pass through the any interface without losing type info
type ValueRef struct {
	Val Value
}

type Value struct {
	Type ValueType
	Num  float64 // inline storage for VAL_NUMBER — keeps arithmetic off the heap
	Data any
}

// Constructors
func NewNil() Value {
	return Value{Type: VAL_NIL, Data: nil}
}

func NewNumber(n float64) Value {
	return Value{Type: VAL_NUMBER, Num: n}
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

func NewBinary(data []byte) Value {
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)
	return Value{Type: VAL_BINARY, Data: &BinaryValue{Data: &dataCopy, Metadata: make(map[string]Value)}}
}

func NewRegex(pattern string, compiled *regexp.Regexp) Value {
	return Value{Type: VAL_REGEX, Data: &RegexValue{Pattern: pattern, Compiled: compiled}}
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

func (v Value) IsBinary() bool {
	return v.Type == VAL_BINARY
}

func (v Value) IsRegex() bool {
	return v.Type == VAL_REGEX
}

// Getters
func (v Value) AsNumber() float64 {
	if v.Type == VAL_NUMBER {
		return v.Num
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

func (v Value) AsBinary() *BinaryValue {
	if v.Type == VAL_BINARY {
		return v.Data.(*BinaryValue)
	}
	return nil
}

func (v Value) AsRegex() *RegexValue {
	if v.Type == VAL_REGEX {
		return v.Data.(*RegexValue)
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
		return v.Num != 0
	case VAL_STRING:
		return v.Data.(string) != ""
	case VAL_ARRAY:
		arr := v.AsArray()
		return len(arr) > 0
	case VAL_OBJECT:
		return len(v.Data.(map[string]Value)) > 0
	case VAL_BINARY:
		bin := v.AsBinary()
		return bin != nil && bin.Data != nil && len(*bin.Data) > 0
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
		n := v.Num
		if n == float64(int64(n)) {
			return strconv.FormatInt(int64(n), 10)
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
			builder.WriteString(ValueToDusoString(item))
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
			builder.WriteString("=")
			builder.WriteString(ValueToDusoString(v))
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
	case VAL_BINARY:
		bin := v.AsBinary()
		if bin != nil && bin.Data != nil {
			size := len(*bin.Data)
			filename := ""
			if fn, ok := bin.Metadata["filename"]; ok && fn.IsString() {
				filename = fn.AsString()
			}
			if filename != "" {
				return fmt.Sprintf("<binary: %s (%d bytes)>", filename, size)
			}
			return fmt.Sprintf("<binary: %d bytes>", size)
		}
		return "<binary>"
	case VAL_REGEX:
		regex := v.AsRegex()
		if regex != nil {
			return fmt.Sprintf("~%s~", regex.Pattern)
		}
		return "<regex>"
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
		return v.Num
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
	case VAL_BINARY:
		return &ValueRef{Val: v} // Wrap binary value so it survives the any conversion
	case VAL_REGEX:
		return &ValueRef{Val: v} // Wrap regex value so it survives the any conversion
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
	case VAL_CODE, VAL_ERROR, VAL_BINARY:
		// Code, error, and binary values are immutable (binary is just a pointer)
		// Return as-is to share the same underlying data
		return v
	default:
		// Primitives are immutable
		return v
	}
}

// DeepCopyAny performs deep copy on any type (for scope boundaries)
func DeepCopyAny(val any) any {
	switch v := val.(type) {
	case *[]Value:
		// Convert *[]Value to []any (script arrays to Go arrays)
		arr := *v
		newArr := make([]any, len(arr))
		for i, elem := range arr {
			newArr[i] = DeepCopyAny(ValueToInterface(elem))
		}
		return newArr
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
