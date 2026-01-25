// builtins.go - Duso core built-in functions
//
// This file implements the standard library of built-in functions available in all Duso scripts.
// These are the functions that come "out of the box" with the language.
//
// CORE LANGUAGE COMPONENT: All functions registered here are part of the minimal core language.
// They are available in both embedded applications (without any setup) and the CLI.
//
// Built-in function categories:
// - I/O: print(), input()
// - Collections: len(), append(), keys(), values()
// - Type: type(), tonumber(), tostring(), tobool()
// - Strings: upper(), lower(), substr(), trim(), split(), join(), contains(), replace()
// - Math: abs(), floor(), ceil(), round(), min(), max(), sqrt(), pow(), clamp()
// - Functional: map(), filter(), reduce()
// - Arrays: sort()
// - JSON: parse_json(), format_json()
// - Utility: range()
// - Date/Time: now(), format_time(), parse_time()
// - System: exit()
//
// Optional features (like file I/O or Claude API) are NOT registered here.
// Those are registered by pkg/cli via RegisterFunctions() or by custom code.
package script

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Builtins struct {
	output    *strings.Builder
	evaluator *Evaluator
}

// NewBuiltins creates a new builtins handler
func NewBuiltins(output *strings.Builder, evaluator *Evaluator) *Builtins {
	return &Builtins{output: output, evaluator: evaluator}
}

// RegisterBuiltins adds built-in functions to an environment
func (b *Builtins) RegisterBuiltins(env *Environment) {
	// Core functions
	env.Define("print", NewGoFunction(b.builtinPrint))
	env.Define("input", NewGoFunction(b.builtinInput))
	env.Define("len", NewGoFunction(b.builtinLen))
	env.Define("append", NewGoFunction(b.builtinAppend))
	env.Define("type", NewGoFunction(b.builtinType))

	// Type conversion
	env.Define("tonumber", NewGoFunction(b.builtinToNumber))
	env.Define("tostring", NewGoFunction(b.builtinToString))
	env.Define("tobool", NewGoFunction(b.builtinToBool))

	// String functions
	env.Define("upper", NewGoFunction(b.builtinUpper))
	env.Define("lower", NewGoFunction(b.builtinLower))
	env.Define("substr", NewGoFunction(b.builtinSubstr))
	env.Define("trim", NewGoFunction(b.builtinTrim))
	env.Define("split", NewGoFunction(b.builtinSplit))
	env.Define("join", NewGoFunction(b.builtinJoin))
	env.Define("contains", NewGoFunction(b.builtinContains))
	env.Define("replace", NewGoFunction(b.builtinReplace))

	// Math functions
	env.Define("floor", NewGoFunction(b.builtinFloor))
	env.Define("ceil", NewGoFunction(b.builtinCeil))
	env.Define("round", NewGoFunction(b.builtinRound))
	env.Define("abs", NewGoFunction(b.builtinAbs))
	env.Define("min", NewGoFunction(b.builtinMin))
	env.Define("max", NewGoFunction(b.builtinMax))
	env.Define("sqrt", NewGoFunction(b.builtinSqrt))
	env.Define("pow", NewGoFunction(b.builtinPow))
	env.Define("clamp", NewGoFunction(b.builtinClamp))

	// Array/Object functions
	env.Define("keys", NewGoFunction(b.builtinKeys))
	env.Define("values", NewGoFunction(b.builtinValues))
	env.Define("sort", NewGoFunction(b.builtinSort))
	env.Define("map", NewGoFunction(b.builtinMap))
	env.Define("filter", NewGoFunction(b.builtinFilter))
	env.Define("reduce", NewGoFunction(b.builtinReduce))

	// JSON functions
	env.Define("parse_json", NewGoFunction(b.builtinParseJSON))
	env.Define("format_json", NewGoFunction(b.builtinFormatJSON))

	// Utility functions
	env.Define("range", NewGoFunction(b.builtinRange))

	// Date/time functions
	env.Define("now", NewGoFunction(b.builtinNow))
	env.Define("format_time", NewGoFunction(b.builtinFormatTime))
	env.Define("parse_time", NewGoFunction(b.builtinParseTime))

	// System functions
	env.Define("exit", NewGoFunction(b.builtinExit))
	env.Define("throw", NewGoFunction(b.builtinThrow))
	env.Define("breakpoint", NewGoFunction(b.builtinBreakpoint))
	env.Define("watch", NewGoFunction(b.builtinWatch))

	// Concurrency functions
	env.Define("parallel", NewGoFunction(b.builtinParallel))
}

// builtinPrint prints values to output
func (b *Builtins) builtinPrint(args map[string]any) (any, error) {
	var parts []string
	for i := 0; ; i++ {
		key := fmt.Sprintf("%d", i)
		if val, ok := args[key]; ok {
			parts = append(parts, fmt.Sprintf("%v", val))
		} else {
			break
		}
	}

	output := strings.Join(parts, " ")
	// Print directly to stdout for immediate interactive feedback
	fmt.Println(output)

	return nil, nil
}

// builtinInput reads a line from stdin with optional prompt
func (b *Builtins) builtinInput(args map[string]any) (any, error) {
	// Optional prompt argument
	if prompt, ok := args["0"]; ok {
		fmt.Fprint(os.Stdout, prompt)
	}

	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err == io.EOF {
		return nil, nil // EOF returns nil
	}
	if err != nil {
		return nil, fmt.Errorf("input() error: %v", err)
	}

	// Remove the trailing newline
	if len(line) > 0 && line[len(line)-1] == '\n' {
		line = line[:len(line)-1]
	}
	// Also remove carriage return if on Windows
	if len(line) > 0 && line[len(line)-1] == '\r' {
		line = line[:len(line)-1]
	}

	return line, nil
}

// builtinLen returns the length of an array, object, or string
func (b *Builtins) builtinLen(args map[string]any) (any, error) {
	if arg, ok := args["0"]; ok {
		switch v := arg.(type) {
		case []any:
			return float64(len(v)), nil
		case map[string]any:
			return float64(len(v)), nil
		case string:
			return float64(len(v)), nil
		default:
			return nil, fmt.Errorf("len() requires array, object, or string")
		}
	}
	return nil, fmt.Errorf("len() requires an argument")
}

// builtinAppend adds an element to an array
func (b *Builtins) builtinAppend(args map[string]any) (any, error) {
	arr, ok := args["0"].([]any)
	if !ok {
		return nil, fmt.Errorf("append() requires an array as first argument")
	}

	val, ok := args["1"]
	if !ok {
		return nil, fmt.Errorf("append() requires a second argument")
	}

	return append(arr, val), nil
}

// builtinType returns the type of a value
func (b *Builtins) builtinType(args map[string]any) (any, error) {
	if arg, ok := args["0"]; ok {
		switch arg.(type) {
		case nil:
			return "nil", nil
		case float64:
			return "number", nil
		case string:
			return "string", nil
		case bool:
			return "boolean", nil
		case []any:
			return "array", nil
		case map[string]any:
			return "object", nil
		default:
			return "unknown", nil
		}
	}
	return nil, fmt.Errorf("type() requires an argument")
}

// Type conversion functions

// builtinToNumber converts a value to number
func (b *Builtins) builtinToNumber(args map[string]any) (any, error) {
	if arg, ok := args["0"]; ok {
		switch v := arg.(type) {
		case float64:
			return v, nil
		case string:
			var n float64
			_, err := fmt.Sscanf(v, "%f", &n)
			if err != nil {
				return 0.0, nil // Return 0 on parse error like Lua
			}
			return n, nil
		case bool:
			if v {
				return 1.0, nil
			}
			return 0.0, nil
		default:
			return 0.0, nil
		}
	}
	return nil, fmt.Errorf("tonumber() requires an argument")
}

// builtinToString converts a value to string
func (b *Builtins) builtinToString(args map[string]any) (any, error) {
	if arg, ok := args["0"]; ok {
		// Special handling for numbers: if it's a whole number, format as integer
		if num, ok := arg.(float64); ok {
			if isInteger(num) {
				return fmt.Sprintf("%d", int64(num)), nil
			}
			return fmt.Sprintf("%v", num), nil
		}
		return fmt.Sprintf("%v", arg), nil
	}
	return nil, fmt.Errorf("tostring() requires an argument")
}

// builtinToBool converts a value to boolean
func (b *Builtins) builtinToBool(args map[string]any) (any, error) {
	if arg, ok := args["0"]; ok {
		switch v := arg.(type) {
		case bool:
			return v, nil
		case nil:
			return false, nil
		case float64:
			return v != 0, nil
		case string:
			return v != "", nil
		case []any:
			return true, nil // Arrays are always truthy
		case map[string]any:
			return true, nil // Objects are always truthy
		default:
			return false, nil
		}
	}
	return nil, fmt.Errorf("tobool() requires an argument")
}

// String functions

// builtinUpper converts string to uppercase
func (b *Builtins) builtinUpper(args map[string]any) (any, error) {
	if arg, ok := args["0"]; ok {
		if s, ok := arg.(string); ok {
			return strings.ToUpper(s), nil
		}
		return nil, fmt.Errorf("upper() requires a string")
	}
	return nil, fmt.Errorf("upper() requires an argument")
}

// builtinLower converts string to lowercase
func (b *Builtins) builtinLower(args map[string]any) (any, error) {
	if arg, ok := args["0"]; ok {
		if s, ok := arg.(string); ok {
			return strings.ToLower(s), nil
		}
		return nil, fmt.Errorf("lower() requires a string")
	}
	return nil, fmt.Errorf("lower() requires an argument")
}

// builtinSubstr extracts substring: substr(str, start [, length])
func (b *Builtins) builtinSubstr(args map[string]any) (any, error) {
	s, ok := args["0"].(string)
	if !ok {
		return nil, fmt.Errorf("substr() requires a string as first argument")
	}

	start, ok := args["1"].(float64)
	if !ok {
		return nil, fmt.Errorf("substr() requires a number as second argument")
	}
	startIdx := int(start)

	// Handle negative index (from end)
	if startIdx < 0 {
		startIdx = len(s) + startIdx
	}
	if startIdx < 0 {
		startIdx = 0
	}
	if startIdx >= len(s) {
		return "", nil
	}

	// If length provided, use it; otherwise take to end
	if length, ok := args["2"].(float64); ok {
		endIdx := startIdx + int(length)
		if endIdx > len(s) {
			endIdx = len(s)
		}
		return s[startIdx:endIdx], nil
	}

	return s[startIdx:], nil
}

// builtinTrim removes whitespace from both ends
func (b *Builtins) builtinTrim(args map[string]any) (any, error) {
	if arg, ok := args["0"]; ok {
		if s, ok := arg.(string); ok {
			return strings.TrimSpace(s), nil
		}
		return nil, fmt.Errorf("trim() requires a string")
	}
	return nil, fmt.Errorf("trim() requires an argument")
}

// builtinSplit splits string by separator
func (b *Builtins) builtinSplit(args map[string]any) (any, error) {
	s, ok := args["0"].(string)
	if !ok {
		return nil, fmt.Errorf("split() requires a string as first argument")
	}

	sep, ok := args["1"].(string)
	if !ok {
		return nil, fmt.Errorf("split() requires a string separator as second argument")
	}

	parts := strings.Split(s, sep)
	result := make([]any, len(parts))
	for i, p := range parts {
		result[i] = p
	}
	return result, nil
}

// builtinJoin joins array elements with separator
func (b *Builtins) builtinJoin(args map[string]any) (any, error) {
	arr, ok := args["0"].([]any)
	if !ok {
		return nil, fmt.Errorf("join() requires an array as first argument")
	}

	sep, ok := args["1"].(string)
	if !ok {
		return nil, fmt.Errorf("join() requires a string separator as second argument")
	}

	parts := make([]string, len(arr))
	for i, item := range arr {
		parts[i] = fmt.Sprintf("%v", item)
	}
	return strings.Join(parts, sep), nil
}

// builtinContains checks if string contains substring
func (b *Builtins) builtinContains(args map[string]any) (any, error) {
	s, ok := args["0"].(string)
	if !ok {
		return nil, fmt.Errorf("contains() requires a string as first argument")
	}

	substr, ok := args["1"].(string)
	if !ok {
		return nil, fmt.Errorf("contains() requires a string as second argument")
	}

	exact := false
	if e, ok := args["2"].(bool); ok {
		exact = e
	}

	if exact {
		return strings.Contains(s, substr), nil
	}
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr)), nil
}

// builtinReplace replaces all instances of old with new
func (b *Builtins) builtinReplace(args map[string]any) (any, error) {
	s, ok := args["0"].(string)
	if !ok {
		return nil, fmt.Errorf("replace() requires a string as first argument")
	}

	old, ok := args["1"].(string)
	if !ok {
		return nil, fmt.Errorf("replace() requires a string as second argument")
	}

	new, ok := args["2"].(string)
	if !ok {
		return nil, fmt.Errorf("replace() requires a string as third argument")
	}

	exact := false
	if e, ok := args["3"].(bool); ok {
		exact = e
	}

	if exact {
		return strings.ReplaceAll(s, old, new), nil
	}

	// Case-insensitive replace: find matches ignoring case, but preserve original text
	lower := strings.ToLower(s)
	oldLower := strings.ToLower(old)

	var result strings.Builder
	lastIdx := 0

	for {
		idx := strings.Index(lower[lastIdx:], oldLower)
		if idx == -1 {
			result.WriteString(s[lastIdx:])
			break
		}

		actualIdx := lastIdx + idx
		result.WriteString(s[lastIdx:actualIdx])
		result.WriteString(new)
		lastIdx = actualIdx + len(old)
	}

	return result.String(), nil
}

// Math functions

// builtinFloor rounds down to nearest integer
func (b *Builtins) builtinFloor(args map[string]any) (any, error) {
	if arg, ok := args["0"].(float64); ok {
		return math.Floor(arg), nil
	}
	return nil, fmt.Errorf("floor() requires a number")
}

// builtinCeil rounds up to nearest integer
func (b *Builtins) builtinCeil(args map[string]any) (any, error) {
	if arg, ok := args["0"].(float64); ok {
		return math.Ceil(arg), nil
	}
	return nil, fmt.Errorf("ceil() requires a number")
}

// builtinRound rounds to nearest integer
func (b *Builtins) builtinRound(args map[string]any) (any, error) {
	if arg, ok := args["0"].(float64); ok {
		return math.Round(arg), nil
	}
	return nil, fmt.Errorf("round() requires a number")
}

// builtinAbs returns absolute value
func (b *Builtins) builtinAbs(args map[string]any) (any, error) {
	if arg, ok := args["0"].(float64); ok {
		return math.Abs(arg), nil
	}
	return nil, fmt.Errorf("abs() requires a number")
}

// builtinMin returns minimum of arguments
func (b *Builtins) builtinMin(args map[string]any) (any, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("min() requires at least one argument")
	}

	var min float64
	var set bool

	for i := 0; ; i++ {
		key := fmt.Sprintf("%d", i)
		val, ok := args[key].(float64)
		if !ok {
			break
		}
		if !set || val < min {
			min = val
			set = true
		}
	}

	if !set {
		return nil, fmt.Errorf("min() requires numeric arguments")
	}
	return min, nil
}

// builtinMax returns maximum of arguments
func (b *Builtins) builtinMax(args map[string]any) (any, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("max() requires at least one argument")
	}

	var max float64
	var set bool

	for i := 0; ; i++ {
		key := fmt.Sprintf("%d", i)
		val, ok := args[key].(float64)
		if !ok {
			break
		}
		if !set || val > max {
			max = val
			set = true
		}
	}

	if !set {
		return nil, fmt.Errorf("max() requires numeric arguments")
	}
	return max, nil
}

// builtinSqrt returns square root
func (b *Builtins) builtinSqrt(args map[string]any) (any, error) {
	if arg, ok := args["0"].(float64); ok {
		return math.Sqrt(arg), nil
	}
	return nil, fmt.Errorf("sqrt() requires a number")
}

// builtinPow returns x^y
func (b *Builtins) builtinPow(args map[string]any) (any, error) {
	x, ok := args["0"].(float64)
	if !ok {
		return nil, fmt.Errorf("pow() requires a number as first argument")
	}

	y, ok := args["1"].(float64)
	if !ok {
		return nil, fmt.Errorf("pow() requires a number as second argument")
	}

	return math.Pow(x, y), nil
}

// builtinClamp clamps value between min and max
func (b *Builtins) builtinClamp(args map[string]any) (any, error) {
	val, ok := args["0"].(float64)
	if !ok {
		return nil, fmt.Errorf("clamp() requires a number as first argument")
	}

	min, ok := args["1"].(float64)
	if !ok {
		return nil, fmt.Errorf("clamp() requires a number as second argument (min)")
	}

	max, ok := args["2"].(float64)
	if !ok {
		return nil, fmt.Errorf("clamp() requires a number as third argument (max)")
	}

	if val < min {
		return min, nil
	}
	if val > max {
		return max, nil
	}
	return val, nil
}

// Array/Object functions

// builtinKeys returns array of object keys or array indices
func (b *Builtins) builtinKeys(args map[string]any) (any, error) {
	if arg, ok := args["0"].(map[string]any); ok {
		keys := make([]any, 0, len(arg))
		for k := range arg {
			keys = append(keys, k)
		}
		return keys, nil
	}
	return nil, fmt.Errorf("keys() requires an object")
}

// builtinValues returns array of object values or array items
func (b *Builtins) builtinValues(args map[string]any) (any, error) {
	if arg, ok := args["0"].(map[string]any); ok {
		values := make([]any, 0, len(arg))
		for _, v := range arg {
			values = append(values, v)
		}
		return values, nil
	}
	return nil, fmt.Errorf("values() requires an object")
}

// callComparisonFunction calls a comparison function with two values and returns a boolean
func (b *Builtins) callComparisonFunction(fn Value, valA, valB Value) (bool, error) {
	if !fn.IsFunction() {
		return false, fmt.Errorf("comparison must be a function")
	}

	// Handle script functions
	if scriptFn, ok := fn.Data.(*ScriptFunction); ok {
		fnEnv := NewFunctionEnvironment(scriptFn.Closure)

		// Define parameters with their defaults
		for _, param := range scriptFn.Parameters {
			var defaultVal Value = NewNil()
			if param.Default != nil {
				prevEnv := b.evaluator.env
				b.evaluator.env = scriptFn.Closure
				val, err := b.evaluator.Eval(param.Default)
				b.evaluator.env = prevEnv
				if err != nil {
					return false, err
				}
				defaultVal = val
			}
			fnEnv.Define(param.Name, defaultVal)
			fnEnv.MarkParameter(param.Name)
		}

		// Override with provided arguments
		if len(scriptFn.Parameters) >= 1 {
			fnEnv.Define(scriptFn.Parameters[0].Name, valA)
		}
		if len(scriptFn.Parameters) >= 2 {
			fnEnv.Define(scriptFn.Parameters[1].Name, valB)
		}

		// Execute the function
		prevEnv := b.evaluator.env
		b.evaluator.env = fnEnv

		var result Value
		for _, stmt := range scriptFn.Body {
			val, err := b.evaluator.Eval(stmt)
			if returnVal, ok := err.(*ReturnValue); ok {
				result = returnVal.Value
				break
			}
			if err != nil {
				b.evaluator.env = prevEnv
				return false, err
			}
			result = val
		}

		b.evaluator.env = prevEnv

		// Convert result to boolean
		return result.IsTruthy(), nil
	}

	// Handle Go functions
	if goFn, ok := fn.Data.(GoFunction); ok {
		argMap := map[string]any{
			"0": valueToInterface(valA),
			"1": valueToInterface(valB),
		}
		res, err := goFn(argMap)
		if err != nil {
			return false, err
		}

		// Convert result to boolean
		resValue := interfaceToValue(res)
		return resValue.IsTruthy(), nil
	}

	return false, fmt.Errorf("invalid function type for comparison")
}

// builtinSort sorts an array with optional comparison function
func (b *Builtins) builtinSort(args map[string]any) (any, error) {
	arr, ok := args["0"].([]any)
	if !ok {
		return nil, fmt.Errorf("sort() requires an array as first argument")
	}

	// Make a copy to avoid modifying original
	result := make([]any, len(arr))
	copy(result, arr)

	// Check if comparison function provided
	if compareFnArg, hasCompareFn := args["1"]; hasCompareFn {
		// Custom comparison function
		if b.evaluator == nil {
			return nil, fmt.Errorf("sort() with comparison function requires evaluator context")
		}

		// Convert the argument back to a Value
		compareFn := interfaceToValue(compareFnArg)

		// Sort using the comparison function
		sortErr := error(nil)
		sort.Slice(result, func(i, j int) bool {
			if sortErr != nil {
				return false
			}

			vi := interfaceToValue(result[i])
			vj := interfaceToValue(result[j])

			less, err := b.callComparisonFunction(compareFn, vi, vj)
			if err != nil {
				sortErr = err
				return false
			}
			return less
		})

		if sortErr != nil {
			return nil, sortErr
		}

		return result, nil
	}

	// Default sort: compare by value
	sort.Slice(result, func(i, j int) bool {
		vi, vj := result[i], result[j]

		// Handle numeric comparison
		if ni, okI := vi.(float64); okI {
			if nj, okJ := vj.(float64); okJ {
				return ni < nj
			}
		}

		// Handle string comparison
		if si, okI := vi.(string); okI {
			if sj, okJ := vj.(string); okJ {
				return si < sj
			}
		}

		// Mixed types or unsupported - compare as strings
		return fmt.Sprintf("%v", vi) < fmt.Sprintf("%v", vj)
	})

	return result, nil
}

// builtinMap applies a function to each element of an array
func (b *Builtins) builtinMap(args map[string]any) (any, error) {
	arr, ok := args["0"].([]any)
	if !ok {
		return nil, fmt.Errorf("map() requires an array as first argument")
	}

	fnArg, ok := args["1"]
	if !ok {
		return nil, fmt.Errorf("map() requires a function as second argument")
	}

	if b.evaluator == nil {
		return nil, fmt.Errorf("map() requires evaluator context")
	}

	fn := interfaceToValue(fnArg)

	result := make([]any, 0, len(arr))
	for _, item := range arr {
		itemVal := interfaceToValue(item)
		retVal, err := b.callUserFunction(fn, []Value{itemVal})
		if err != nil {
			return nil, fmt.Errorf("error in map function: %w", err)
		}
		result = append(result, ValueToInterface(retVal))
	}

	return result, nil
}

// builtinFilter keeps only array elements that match a predicate
func (b *Builtins) builtinFilter(args map[string]any) (any, error) {
	arr, ok := args["0"].([]any)
	if !ok {
		return nil, fmt.Errorf("filter() requires an array as first argument")
	}

	fnArg, ok := args["1"]
	if !ok {
		return nil, fmt.Errorf("filter() requires a function as second argument")
	}

	if b.evaluator == nil {
		return nil, fmt.Errorf("filter() requires evaluator context")
	}

	fn := interfaceToValue(fnArg)

	result := make([]any, 0, len(arr))
	for _, item := range arr {
		itemVal := interfaceToValue(item)
		retVal, err := b.callUserFunction(fn, []Value{itemVal})
		if err != nil {
			return nil, fmt.Errorf("error in filter function: %w", err)
		}
		if retVal.IsTruthy() {
			result = append(result, item)
		}
	}

	return result, nil
}

// builtinReduce combines all array elements into a single value
func (b *Builtins) builtinReduce(args map[string]any) (any, error) {
	arr, ok := args["0"].([]any)
	if !ok {
		return nil, fmt.Errorf("reduce() requires an array as first argument")
	}

	fnArg, ok := args["1"]
	if !ok {
		return nil, fmt.Errorf("reduce() requires a function as second argument")
	}

	if b.evaluator == nil {
		return nil, fmt.Errorf("reduce() requires evaluator context")
	}

	fn := interfaceToValue(fnArg)

	// Get initial value (third argument)
	accumulator := NewNil()
	if initVal, ok := args["2"]; ok {
		accumulator = interfaceToValue(initVal)
	}

	// Iterate through array
	for _, item := range arr {
		itemVal := interfaceToValue(item)
		retVal, err := b.callUserFunction(fn, []Value{accumulator, itemVal})
		if err != nil {
			return nil, fmt.Errorf("error in reduce function: %w", err)
		}
		accumulator = retVal
	}

	return ValueToInterface(accumulator), nil
}

// callUserFunction calls a user function with the given arguments
func (b *Builtins) callUserFunction(fn Value, args []Value) (Value, error) {
	if !fn.IsFunction() {
		return NewNil(), fmt.Errorf("expected function")
	}

	// Handle script functions
	if scriptFn, ok := fn.Data.(*ScriptFunction); ok {
		fnEnv := NewFunctionEnvironment(scriptFn.Closure)

		// Define parameters with their defaults
		for i, param := range scriptFn.Parameters {
			var defaultVal Value = NewNil()
			if param.Default != nil {
				prevEnv := b.evaluator.env
				b.evaluator.env = scriptFn.Closure
				val, err := b.evaluator.Eval(param.Default)
				b.evaluator.env = prevEnv
				if err != nil {
					return NewNil(), err
				}
				defaultVal = val
			}
			fnEnv.Define(param.Name, defaultVal)
			fnEnv.MarkParameter(param.Name)

			// Override with provided arguments
			if i < len(args) {
				fnEnv.Define(param.Name, args[i])
			}
		}

		// Execute the function
		prevEnv := b.evaluator.env
		b.evaluator.env = fnEnv

		var result Value
		for _, stmt := range scriptFn.Body {
			val, err := b.evaluator.Eval(stmt)
			if returnVal, ok := err.(*ReturnValue); ok {
				result = returnVal.Value
				break
			}
			if err != nil {
				b.evaluator.env = prevEnv
				return NewNil(), err
			}
			result = val
		}

		b.evaluator.env = prevEnv
		return result, nil
	}

	// Handle Go functions
	if goFn, ok := fn.Data.(GoFunction); ok {
		argMap := make(map[string]any)
		for i, arg := range args {
			argMap[fmt.Sprintf("%d", i)] = ValueToInterface(arg)
		}
		ret, err := goFn(argMap)
		if err != nil {
			return NewNil(), err
		}
		return interfaceToValue(ret), nil
	}

	return NewNil(), fmt.Errorf("not a callable function")
}

// Utility functions

// builtinRange creates an array of numbers in range
func (b *Builtins) builtinRange(args map[string]any) (any, error) {
	start, ok := args["0"].(float64)
	if !ok {
		return nil, fmt.Errorf("range() requires a number as first argument")
	}

	end, ok := args["1"].(float64)
	if !ok {
		return nil, fmt.Errorf("range() requires a number as second argument")
	}

	step := 1.0
	if s, ok := args["2"].(float64); ok {
		step = s
	}

	if step == 0 {
		return nil, fmt.Errorf("range() step cannot be zero")
	}

	var result []any
	if step > 0 {
		for i := start; i <= end; i += step {
			result = append(result, i)
		}
	} else {
		for i := start; i >= end; i += step {
			result = append(result, i)
		}
	}

	return result, nil
}

// System functions

// builtinExit stops execution and returns values to host
func (b *Builtins) builtinExit(args map[string]any) (any, error) {
	// Collect all arguments as return values
	values := make([]any, 0)
	for i := 0; ; i++ {
		key := fmt.Sprintf("%d", i)
		if val, ok := args[key]; ok {
			values = append(values, val)
		} else {
			break
		}
	}

	return nil, &ExitExecution{Values: values}
}

// builtinThrow throws an error with message and call stack
func (b *Builtins) builtinThrow(args map[string]any) (any, error) {
	message := ""
	if msg, ok := args["0"].(string); ok {
		message = msg
	} else if msg, ok := args["message"]; ok {
		message = fmt.Sprintf("%v", msg)
	} else {
		message = "unknown error"
	}

	// Create DusoError with call stack
	err := &DusoError{
		Message:   message,
		FilePath:  b.evaluator.ctx.FilePath,
		CallStack: b.evaluator.ctx.CallStack,
	}

	return nil, err
}

// formatArgs converts arguments to space-separated string (like print would output)
func (b *Builtins) formatArgs(args map[string]any) string {
	var parts []string
	for i := 0; ; i++ {
		key := fmt.Sprintf("%d", i)
		if val, ok := args[key]; ok {
			parts = append(parts, fmt.Sprintf("%v", val))
		} else {
			break
		}
	}
	return strings.Join(parts, " ")
}

// builtinBreakpoint signals a debug breakpoint with call stack captured
// Optional arguments are printed before the breakpoint (like print())
func (b *Builtins) builtinBreakpoint(args map[string]any) (any, error) {
	// Only trigger breakpoint if debug mode is enabled
	if !b.evaluator.DebugMode {
		return nil, nil
	}

	// If arguments provided, print them with prefix
	if len(args) > 0 {
		output := b.formatArgs(args)
		fmt.Println("BREAKPOINT: " + output)
	}

	// Capture call stack and current environment for debug display
	// Clone the call stack so it can't be modified
	callStack := make([]CallFrame, len(b.evaluator.ctx.CallStack))
	copy(callStack, b.evaluator.ctx.CallStack)

	err := &BreakpointError{
		FilePath:  b.evaluator.ctx.FilePath,
		CallStack: callStack,
		Env:       b.evaluator.env, // Capture current environment for scope access
	}
	return nil, err
}

// builtinWatch evaluates expressions and breaks if values change
// Each argument is a string expression to watch
func (b *Builtins) builtinWatch(args map[string]any) (any, error) {
	var triggered []string // Collect which watches triggered

	// Process each watch expression
	for i := 0; ; i++ {
		key := fmt.Sprintf("%d", i)
		exprStr, ok := args[key]
		if !ok {
			break
		}

		// Expression must be a string
		expr, ok := exprStr.(string)
		if !ok {
			return nil, fmt.Errorf("watch() requires string expressions, got %v", exprStr)
		}

		// Parse and evaluate the expression
		lexer := NewLexer(expr)
		tokens := lexer.Tokenize()
		parser := NewParser(tokens)
		node, err := parser.parseExpression()
		if err != nil {
			return nil, fmt.Errorf("watch() parse error in '{{%s}}': %v", expr, err)
		}

		val, err := b.evaluator.Eval(node)
		if err != nil {
			return nil, fmt.Errorf("watch() evaluation error in '{{%s}}': %v", expr, err)
		}

		// Check if value changed from cached
		cachedVal, exists := b.evaluator.watchCache[expr]
		if !exists || !b.valuesEqual(val, cachedVal) {
			// Value changed or first time seeing it
			b.evaluator.watchCache[expr] = val
			triggered = append(triggered, fmt.Sprintf("WATCH: %s = %v", expr, val.String()))
		}
	}

	// If any watches triggered and debug mode is enabled, print them and break
	if len(triggered) > 0 && b.evaluator.DebugMode {
		for _, msg := range triggered {
			fmt.Println(msg)
		}

		// Trigger breakpoint with call stack
		callStack := make([]CallFrame, len(b.evaluator.ctx.CallStack))
		copy(callStack, b.evaluator.ctx.CallStack)

		err := &BreakpointError{
			FilePath:  b.evaluator.ctx.FilePath,
			CallStack: callStack,
			Env:       b.evaluator.env,
		}
		return nil, err
	}

	return nil, nil
}

// valuesEqual checks if two values are equal (for watch caching)
func (b *Builtins) valuesEqual(v1, v2 Value) bool {
	if v1.Type != v2.Type {
		return false
	}

	switch v1.Type {
	case VAL_NIL:
		return true
	case VAL_NUMBER:
		return v1.AsNumber() == v2.AsNumber()
	case VAL_STRING:
		return v1.AsString() == v2.AsString()
	case VAL_BOOL:
		return v1.AsBool() == v2.AsBool()
	case VAL_ARRAY:
		arr1 := v1.AsArray()
		arr2 := v2.AsArray()
		if len(arr1) != len(arr2) {
			return false
		}
		for i := range arr1 {
			if !b.valuesEqual(arr1[i], arr2[i]) {
				return false
			}
		}
		return true
	case VAL_OBJECT:
		obj1 := v1.AsObject()
		obj2 := v2.AsObject()
		if len(obj1) != len(obj2) {
			return false
		}
		for k, v := range obj1 {
			v2Val, ok := obj2[k]
			if !ok || !b.valuesEqual(v, v2Val) {
				return false
			}
		}
		return true
	case VAL_FUNCTION:
		// Functions are compared by reference
		return v1.Data == v2.Data
	default:
		return false
	}
}

// Date/time functions

// translateDateFormat converts standard date format (YYYY-MM-DD) to Go's format
func translateDateFormat(format string) string {
	replacements := map[string]string{
		"YYYY": "2006",
		"YY":   "06",
		"MM":   "01",
		"DD":   "02",
		"HH":   "15",
		"mm":   "04",
		"ss":   "05",
	}

	result := format
	for standard, goFormat := range replacements {
		result = strings.ReplaceAll(result, standard, goFormat)
	}
	return result
}

// builtinNow returns current Unix timestamp (seconds)
func (b *Builtins) builtinNow(args map[string]any) (any, error) {
	return float64(time.Now().Unix()), nil
}

// builtinFormatTime formats a Unix timestamp to string
func (b *Builtins) builtinFormatTime(args map[string]any) (any, error) {
	var timestamp float64
	var ok bool

	// Accept either number or string that parses as number
	arg := args["0"]
	if num, isNum := arg.(float64); isNum {
		timestamp = num
		ok = true
	} else if str, isStr := arg.(string); isStr {
		// Try to parse string as number (e.g., JSON timestamp from string)
		num, err := strconv.ParseFloat(str, 64)
		if err == nil {
			timestamp = num
			ok = true
		}
	}

	if !ok {
		return nil, fmt.Errorf("format_time() requires a number or numeric string as first argument")
	}

	format := "2006-01-02 15:04:05" // default

	if formatArg, ok := args["1"].(string); ok {
		switch formatArg {
		case "iso":
			format = "2006-01-02T15:04:05Z"
		case "date":
			format = "2006-01-02"
		case "time":
			format = "15:04:05"
		case "long_date":
			format = "January 2, 2006"
		case "long_date_dow":
			format = "Mon January 2, 2006"
		case "short_date":
			format = "Jan 2, 2006"
		case "short_date_dow":
			format = "Mon Jan 2, 2006"
		default:
			// User provided custom format
			format = translateDateFormat(formatArg)
		}
	}

	t := time.Unix(int64(timestamp), 0).UTC()
	return t.Format(format), nil
}

// builtinParseTime parses a date string to Unix timestamp
func (b *Builtins) builtinParseTime(args map[string]any) (any, error) {
	dateStr, ok := args["0"].(string)
	if !ok {
		return nil, fmt.Errorf("parse_time() requires a string as first argument")
	}

	// If format provided, use it
	if formatArg, ok := args["1"].(string); ok {
		format := translateDateFormat(formatArg)
		t, err := time.Parse(format, dateStr)
		if err != nil {
			return nil, fmt.Errorf("parse_time() failed to parse %q with format %q: %v", dateStr, formatArg, err)
		}
		return float64(t.Unix()), nil
	}

	// No format hint: try common patterns
	commonFormats := []string{
		"2006-01-02T15:04:05Z",        // ISO with Z
		"2006-01-02T15:04:05",         // ISO without Z
		"2006-01-02 15:04:05",         // Default
		"2006-01-02",                  // Date only
		"January 2, 2006",             // Long date
		"Mon January 2, 2006",         // Long date with day of week
		"Jan 2, 2006",                 // Short date
		"Mon Jan 2, 2006",             // Short date with day of week
	}

	for _, format := range commonFormats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return float64(t.Unix()), nil
		}
	}

	return nil, fmt.Errorf("parse_time() could not parse %q - try providing a format", dateStr)
}

// builtinParseJSON parses a JSON string into Duso objects/arrays
func (b *Builtins) builtinParseJSON(args map[string]any) (any, error) {
	jsonStr, ok := args["0"].(string)
	if !ok {
		return nil, fmt.Errorf("parse_json() requires a string as first argument")
	}

	var result any
	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return nil, fmt.Errorf("parse_json() failed to parse JSON: %v", err)
	}

	// Convert JSON types to Duso-friendly types
	return b.jsonToValue(result), nil
}

// jsonToValue recursively converts JSON-unmarshaled values to Duso values
func (b *Builtins) jsonToValue(v any) any {
	switch val := v.(type) {
	case map[string]any:
		// Convert JSON object to Duso object
		obj := make(map[string]any)
		for k, v := range val {
			obj[k] = b.jsonToValue(v)
		}
		return obj
	case []any:
		// Convert JSON array to Duso array
		arr := make([]any, len(val))
		for i, v := range val {
			arr[i] = b.jsonToValue(v)
		}
		return arr
	case nil:
		return nil
	case bool:
		return val
	case float64:
		return val
	case string:
		return val
	default:
		return fmt.Sprintf("%v", val)
	}
}

// builtinFormatJSON converts a Duso value to JSON string
func (b *Builtins) builtinFormatJSON(args map[string]any) (any, error) {
	if _, ok := args["0"]; !ok {
		return nil, fmt.Errorf("format_json() requires at least one argument")
	}

	value := args["0"]

	// Check if indent is specified
	var indent string
	if indentArg, ok := args["1"]; ok {
		switch i := indentArg.(type) {
		case float64:
			// Create indent string (spaces)
			indentNum := int(i)
			if indentNum < 0 {
				indentNum = 0
			}
			indent = strings.Repeat(" ", indentNum)
		case string:
			indent = i
		}
	}

	// Convert Duso value to JSON-marshable format
	jsonValue := b.valueToJSON(value)

	var result []byte
	var err error

	if indent != "" {
		result, err = json.MarshalIndent(jsonValue, "", indent)
	} else {
		result, err = json.Marshal(jsonValue)
	}

	if err != nil {
		return nil, fmt.Errorf("format_json() failed to serialize: %v", err)
	}

	return string(result), nil
}

// valueToJSON recursively converts Duso values to JSON-marshable values
func (b *Builtins) valueToJSON(v any) any {
	switch val := v.(type) {
	case map[string]any:
		// Convert Duso object to JSON object
		obj := make(map[string]any)
		for k, v := range val {
			obj[k] = b.valueToJSON(v)
		}
		return obj
	case []any:
		// Convert Duso array to JSON array
		arr := make([]any, len(val))
		for i, v := range val {
			arr[i] = b.valueToJSON(v)
		}
		return arr
	case nil:
		return nil
	case bool:
		return val
	case float64:
		return val
	case string:
		return val
	case *ScriptFunction:
		// Functions can't be serialized
		return "[Function]"
	default:
		return fmt.Sprintf("%v", val)
	}
}

// builtinParallel executes functions concurrently with isolated evaluators
// Each function runs in its own Evaluator instance with parent scope access (read-only)
// This enables true parallelism with no evaluator contention or shared mutable state
// Accepts: array of functions, object of functions, or varargs of functions
// Returns: results in same structure as input
// Error handling: all run regardless, errors become nil
func (b *Builtins) builtinParallel(args map[string]any) (any, error) {
	if b.evaluator == nil {
		return nil, fmt.Errorf("parallel() requires evaluator context")
	}

	// Case 1: Single array argument parallel([fn1, fn2, fn3])
	if arr, ok := args["0"].([]any); ok && len(args) == 1 {
		return b.parallelArrayWithEval(arr)
	}

	// Case 2: Single object argument parallel({a = fn1, b = fn2})
	if obj, ok := args["0"].(map[string]any); ok && len(args) == 1 {
		return b.parallelObjectWithEval(obj)
	}

	// Case 3: Varargs parallel(fn1, fn2, fn3)
	varargs := make([]any, 0)
	for i := 0; ; i++ {
		key := fmt.Sprintf("%d", i)
		if val, ok := args[key]; ok {
			varargs = append(varargs, val)
		} else {
			break
		}
	}

	if len(varargs) > 0 {
		return b.parallelArrayWithEval(varargs)
	}

	return nil, fmt.Errorf("parallel() requires an array, object, or functions as arguments")
}

// parallelArrayWithEval executes an array of functions in parallel with isolated evaluators
func (b *Builtins) parallelArrayWithEval(functions []any) (any, error) {
	results := make([]any, len(functions))

	var wg sync.WaitGroup
	for i, fnArg := range functions {
		wg.Add(1)
		go func(index int, fn any) {
			defer wg.Done()

			// Create a child evaluator for this block with parent scope access
			childEval := NewEvaluator(&strings.Builder{})
			childEval.env.parent = b.evaluator.env
			childEval.env.evaluator = childEval // Set evaluator reference for parallel context check
			childEval.isParallelContext = true   // Block parent scope writes

			// Call the function in the child evaluator
			fnVal := interfaceToValue(fn)
			result, err := callUserFunctionInEvaluator(childEval, fnVal, []Value{})

			if err != nil {
				// Error handling: Option B - errors become nil
				results[index] = nil
			} else {
				results[index] = ValueToInterface(result)
			}
		}(i, fnArg)
	}
	wg.Wait()

	return results, nil
}

// parallelObjectWithEval executes an object of functions in parallel with isolated evaluators
func (b *Builtins) parallelObjectWithEval(functions map[string]any) (any, error) {
	results := make(map[string]any)
	var mu sync.Mutex

	var wg sync.WaitGroup
	for key, fnArg := range functions {
		wg.Add(1)
		go func(k string, fn any) {
			defer wg.Done()

			// Create a child evaluator for this block with parent scope access
			childEval := NewEvaluator(&strings.Builder{})
			childEval.env.parent = b.evaluator.env
			childEval.env.evaluator = childEval // Set evaluator reference for parallel context check
			childEval.isParallelContext = true   // Block parent scope writes

			// Call the function in the child evaluator
			fnVal := interfaceToValue(fn)
			result, err := callUserFunctionInEvaluator(childEval, fnVal, []Value{})

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				// Error handling: Option B - errors become nil
				results[k] = nil
			} else {
				results[k] = ValueToInterface(result)
			}
		}(key, fnArg)
	}
	wg.Wait()

	return results, nil
}

// callUserFunctionInEvaluator calls a user function in a specific evaluator context
// Similar to callUserFunction but uses the provided evaluator
func callUserFunctionInEvaluator(eval *Evaluator, fn Value, args []Value) (Value, error) {
	if !fn.IsFunction() {
		return NewNil(), fmt.Errorf("expected function")
	}

	// Handle script functions
	if scriptFn, ok := fn.Data.(*ScriptFunction); ok {
		fnEnv := NewFunctionEnvironment(scriptFn.Closure)
		fnEnv.evaluator = eval // Set evaluator reference

		// Define parameters with their defaults
		for i, param := range scriptFn.Parameters {
			var defaultVal Value = NewNil()
			if param.Default != nil {
				prevEnv := eval.env
				eval.env = scriptFn.Closure
				val, err := eval.Eval(param.Default)
				eval.env = prevEnv
				if err != nil {
					return NewNil(), err
				}
				defaultVal = val
			}
			fnEnv.Define(param.Name, defaultVal)
			fnEnv.MarkParameter(param.Name)

			// Override with provided arguments
			if i < len(args) {
				fnEnv.Define(param.Name, args[i])
			}
		}

		// Execute the function
		prevEnv := eval.env
		eval.env = fnEnv

		var result Value
		for _, stmt := range scriptFn.Body {
			val, err := eval.Eval(stmt)
			if returnVal, ok := err.(*ReturnValue); ok {
				result = returnVal.Value
				break
			}
			if err != nil {
				eval.env = prevEnv
				return NewNil(), err
			}
			result = val
		}

		eval.env = prevEnv
		return result, nil
	}

	// Handle Go functions
	if goFn, ok := fn.Data.(GoFunction); ok {
		argMap := make(map[string]any)
		for i, arg := range args {
			argMap[fmt.Sprintf("%d", i)] = ValueToInterface(arg)
		}
		ret, err := goFn(argMap)
		if err != nil {
			return NewNil(), err
		}
		return interfaceToValue(ret), nil
	}

	return NewNil(), fmt.Errorf("not a callable function")
}
