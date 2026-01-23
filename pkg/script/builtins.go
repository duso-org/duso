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
// - Collections: len(), append()
// - Type: type(), tonumber(), tostring(), tobool()
// - Strings: upper(), lower(), substr(), trim(), split(), join(), contains(), replace()
// - Math: abs(), floor(), ceil(), round(), min(), max(), sqrt()
// - Arrays: map(), filter(), reduce(), sort()
// - JSON: parse_json(), format_json()
// - Misc: range(), assert()
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

		// Define the two parameters
		if len(scriptFn.Parameters) >= 1 {
			fnEnv.Define(scriptFn.Parameters[0], valA)
		}
		if len(scriptFn.Parameters) >= 2 {
			fnEnv.Define(scriptFn.Parameters[1], valB)
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
