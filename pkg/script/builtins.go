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
// - Collections: len(), keys(), values()
// - Type: type(), tonumber(), tostring(), tobool()
// - Strings: upper(), lower(), substr(), trim(), split(), join(), contains(), replace()
// - Math: abs(), floor(), ceil(), round(), min(), max(), sqrt(), pow(), clamp()
// - Trigonometry: sin(), cos(), tan(), asin(), acos(), atan(), atan2()
// - Exponential/Logarithmic: exp(), log() (base 10), ln() (natural log)
// - Constants: pi
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
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	mathrand "math/rand"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Builtins struct {
	// Completely stateless - all methods accept evaluator as parameter
	// This allows unlimited concurrent handler execution without locks
}

// NewBuiltins creates a new stateless builtins handler
// Since Builtins is completely stateless, there's only one global instance
func NewBuiltins() *Builtins {
	return &Builtins{}
}

// RegisterBuiltins adds built-in functions to an environment
func (b *Builtins) RegisterBuiltins(env *Environment) {
	// Core functions - wrap each method to pass evaluator at call time
	env.Define("print", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinPrint(eval, args)
	}))
	env.Define("input", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinInput(eval, args)
	}))
	env.Define("len", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinLen(eval, args)
	}))
	env.Define("type", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinType(eval, args)
	}))

	// Type conversion
	env.Define("tonumber", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinToNumber(eval, args)
	}))
	env.Define("tostring", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinToString(eval, args)
	}))
	env.Define("tobool", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinToBool(eval, args)
	}))

	// String functions
	env.Define("upper", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinUpper(eval, args)
	}))
	env.Define("lower", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinLower(eval, args)
	}))
	env.Define("substr", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinSubstr(eval, args)
	}))
	env.Define("trim", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinTrim(eval, args)
	}))
	env.Define("split", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinSplit(eval, args)
	}))
	env.Define("join", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinJoin(eval, args)
	}))
	env.Define("contains", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinContains(eval, args)
	}))
	env.Define("find", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinFind(eval, args)
	}))
	env.Define("replace", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinReplace(eval, args)
	}))
	env.Define("template", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinTemplate(eval, args)
	}))

	// Math functions
	env.Define("floor", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinFloor(eval, args)
	}))
	env.Define("ceil", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinCeil(eval, args)
	}))
	env.Define("round", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinRound(eval, args)
	}))
	env.Define("abs", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinAbs(eval, args)
	}))
	env.Define("min", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinMin(eval, args)
	}))
	env.Define("max", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinMax(eval, args)
	}))
	env.Define("sqrt", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinSqrt(eval, args)
	}))
	env.Define("pow", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinPow(eval, args)
	}))
	env.Define("clamp", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinClamp(eval, args)
	}))

	// Trigonometric functions
	env.Define("sin", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinSin(eval, args)
	}))
	env.Define("cos", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinCos(eval, args)
	}))
	env.Define("tan", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinTan(eval, args)
	}))
	env.Define("asin", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinAsin(eval, args)
	}))
	env.Define("acos", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinAcos(eval, args)
	}))
	env.Define("atan", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinAtan(eval, args)
	}))
	env.Define("atan2", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinAtan2(eval, args)
	}))

	// Exponential and logarithmic functions
	env.Define("exp", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinExp(eval, args)
	}))
	env.Define("log", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinLog(eval, args)
	}))
	env.Define("ln", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinLn(eval, args)
	}))
	env.Define("pi", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinPi(eval, args)
	}))

	// Array/Object functions
	env.Define("keys", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinKeys(eval, args)
	}))
	env.Define("values", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinValues(eval, args)
	}))
	env.Define("sort", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinSort(eval, args)
	}))
	env.Define("map", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinMap(eval, args)
	}))
	env.Define("filter", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinFilter(eval, args)
	}))
	env.Define("push", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinPush(eval, args)
	}))
	env.Define("pop", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinPop(eval, args)
	}))
	env.Define("shift", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinShift(eval, args)
	}))
	env.Define("unshift", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinUnshift(eval, args)
	}))
	env.Define("reduce", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinReduce(eval, args)
	}))
	env.Define("deep_copy", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinDeepCopy(eval, args)
	}))

	// JSON functions
	env.Define("parse_json", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinParseJSON(eval, args)
	}))
	env.Define("format_json", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinFormatJSON(eval, args)
	}))

	// Utility functions
	env.Define("range", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinRange(eval, args)
	}))
	env.Define("random", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinRandom(eval, args)
	}))
	env.Define("uuid", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinUUID(eval, args)
	}))

	// Date/time functions
	env.Define("now", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinNow(eval, args)
	}))
	env.Define("format_time", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinFormatTime(eval, args)
	}))
	env.Define("parse_time", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinParseTime(eval, args)
	}))
	env.Define("sleep", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinSleep(eval, args)
	}))

	// System functions
	env.Define("exit", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinExit(eval, args)
	}))
	env.Define("throw", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinThrow(eval, args)
	}))
	env.Define("breakpoint", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinBreakpoint(eval, args)
	}))
	env.Define("watch", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinWatch(eval, args)
	}))

	// Concurrency functions
	env.Define("parallel", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinParallel(eval, args)
	}))

	// Coordination & state
	env.Define("datastore", NewGoFunction(func(eval *Evaluator, args map[string]any) (any, error) {
		return b.builtinDatastore(eval, args)
	}))
}

// builtinPrint prints values to output
func (b *Builtins) builtinPrint(evaluator *Evaluator, args map[string]any) (any, error) {
	var parts []string
	for i := 0; ; i++ {
		key := fmt.Sprintf("%d", i)
		if val, ok := args[key]; ok {
			// Convert back to script Value to get proper string representation
			scriptVal := interfaceToValue(val)
			parts = append(parts, scriptVal.String())
		} else {
			break
		}
	}

	output := strings.Join(parts, " ")
	fmt.Println(output)
	return nil, nil
}

// builtinInput reads a line from stdin with optional prompt
func (b *Builtins) builtinInput(evaluator *Evaluator, args map[string]any) (any, error) {
	// Optional prompt argument
	if prompt, ok := args["0"]; ok {
		fmt.Fprint(os.Stdout, prompt)
	}

	if evaluator != nil && evaluator.NoStdin {
		fmt.Println("warning: stdin disabled, input() returned ''")
		return "", nil
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

// builtinLen returns the length of an array, object, or string (returns 0 for nil)
func (b *Builtins) builtinLen(evaluator *Evaluator, args map[string]any) (any, error) {
	if arg, ok := args["0"]; ok {
		switch v := arg.(type) {
		case nil:
			return float64(0), nil
		case *[]Value:
			return float64(len(*v)), nil
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

// builtinType returns the type of a value
func (b *Builtins) builtinType(evaluator *Evaluator, args map[string]any) (any, error) {
	if arg, ok := args["0"]; ok {
		// Check for ValueRef wrapper first (used for functions)
		if vr, ok := arg.(*ValueRef); ok {
			return vr.Val.Type.String(), nil
		}

		switch arg.(type) {
		case nil:
			return "nil", nil
		case float64:
			return "number", nil
		case string:
			return "string", nil
		case bool:
			return "boolean", nil
		case *[]Value:
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
func (b *Builtins) builtinToNumber(evaluator *Evaluator, args map[string]any) (any, error) {
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
func (b *Builtins) builtinToString(evaluator *Evaluator, args map[string]any) (any, error) {
	if arg, ok := args["0"]; ok {
		// Special handling for numbers: if it's a whole number, format as integer
		if num, ok := arg.(float64); ok {
			if isInteger(num) {
				return fmt.Sprintf("%d", int64(num)), nil
			}
			return fmt.Sprintf("%v", num), nil
		}
		// Special handling for arrays
		if arrPtr, ok := arg.(*[]Value); ok {
			val := Value{Type: VAL_ARRAY, Data: arrPtr}
			return val.String(), nil
		}
		return fmt.Sprintf("%v", arg), nil
	}
	return nil, fmt.Errorf("tostring() requires an argument")
}

// builtinToBool converts a value to boolean
func (b *Builtins) builtinToBool(evaluator *Evaluator, args map[string]any) (any, error) {
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
		case *[]Value:
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

// builtinUpper converts string to uppercase, coercing input to string if needed
func (b *Builtins) builtinUpper(evaluator *Evaluator, args map[string]any) (any, error) {
	if arg, ok := args["0"]; ok {
		var s string
		if strVal, ok := arg.(string); ok {
			s = strVal
		} else {
			// Coerce to string using tostring logic
			if num, ok := arg.(float64); ok {
				if isInteger(num) {
					s = fmt.Sprintf("%d", int64(num))
				} else {
					s = fmt.Sprintf("%v", num)
				}
			} else {
				s = fmt.Sprintf("%v", arg)
			}
		}
		return strings.ToUpper(s), nil
	}
	return nil, fmt.Errorf("upper() requires an argument")
}

// builtinLower converts string to lowercase, coercing input to string if needed
func (b *Builtins) builtinLower(evaluator *Evaluator, args map[string]any) (any, error) {
	if arg, ok := args["0"]; ok {
		var s string
		if strVal, ok := arg.(string); ok {
			s = strVal
		} else {
			// Coerce to string using tostring logic
			if num, ok := arg.(float64); ok {
				if isInteger(num) {
					s = fmt.Sprintf("%d", int64(num))
				} else {
					s = fmt.Sprintf("%v", num)
				}
			} else {
				s = fmt.Sprintf("%v", arg)
			}
		}
		return strings.ToLower(s), nil
	}
	return nil, fmt.Errorf("lower() requires an argument")
}

// builtinSubstr extracts substring: substr(str, start [, length])
func (b *Builtins) builtinSubstr(evaluator *Evaluator, args map[string]any) (any, error) {
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
func (b *Builtins) builtinTrim(evaluator *Evaluator, args map[string]any) (any, error) {
	if arg, ok := args["0"]; ok {
		if s, ok := arg.(string); ok {
			return strings.TrimSpace(s), nil
		}
		return nil, fmt.Errorf("trim() requires a string")
	}
	return nil, fmt.Errorf("trim() requires an argument")
}

// builtinSplit splits string by separator
func (b *Builtins) builtinSplit(evaluator *Evaluator, args map[string]any) (any, error) {
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
func (b *Builtins) builtinJoin(evaluator *Evaluator, args map[string]any) (any, error) {
	arrPtr, ok := args["0"].(*[]Value)
	if !ok {
		return nil, fmt.Errorf("join() requires an array as first argument")
	}

	sep, ok := args["1"].(string)
	if !ok {
		return nil, fmt.Errorf("join() requires a string separator as second argument")
	}

	arr := *arrPtr
	parts := make([]string, len(arr))
	for i, item := range arr {
		parts[i] = item.String()
	}
	return strings.Join(parts, sep), nil
}

// builtinContains checks if string contains substring
func (b *Builtins) builtinContains(evaluator *Evaluator, args map[string]any) (any, error) {
	s, ok := args["0"].(string)
	if !ok {
		return nil, fmt.Errorf("contains() requires a string as first argument")
	}

	pattern, ok := args["1"].(string)
	if !ok {
		return nil, fmt.Errorf("contains() requires a string as second argument")
	}

	ignoreCase := false
	if ic, ok := args["ignore_case"].(bool); ok {
		ignoreCase = ic
	}

	// Add case-insensitive flag if needed
	if ignoreCase {
		pattern = "(?i)" + pattern
	}

	// Compile as regex
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("contains() invalid regex: %v", err)
	}

	return re.MatchString(s), nil
}

// builtinReplace replaces all instances of old with new

// builtinTemplate creates a reusable template function from a template string
// template(template_string) returns a function that evaluates the template with provided named args
func (b *Builtins) builtinTemplate(evaluator *Evaluator, args map[string]any) (any, error) {
	templateStr, ok := args["0"].(string)
	if !ok {
		return nil, fmt.Errorf("template() requires a string argument")
	}

	// Check if the string contains template expressions
	if !strings.Contains(templateStr, "{{") {
		return nil, fmt.Errorf("template() requires a string with {{expressions}}. Use raw \"...\" to pass a template string without evaluation")
	}

	// Return a function that evaluates the template with provided args
	templateFn := func(templateEval *Evaluator, templateArgs map[string]any) (any, error) {
		// Create a fresh environment with ONLY the provided arguments
		// This means undefined variables will render as {{varname}}
		templateEnv := NewEnvironment()

		// Add all provided arguments to the template environment (skip positional args)
		for key, val := range templateArgs {
			// Skip numeric positional keys
			if _, err := strconv.Atoi(key); err == nil {
				continue
			}

			// Convert Go value to Duso Value
			var dusoVal Value
			switch v := val.(type) {
			case Value:
				dusoVal = v
			case float64:
				dusoVal = NewNumber(v)
			case string:
				dusoVal = NewString(v)
			case bool:
				dusoVal = NewBool(v)
			case []Value:
				dusoVal = NewArray(v)
			case map[string]Value:
				dusoVal = NewObject(v)
			case map[string]any:
				// Convert Go map to Duso object
				obj := make(map[string]Value)
				for k, v := range v {
					if dv, ok := v.(Value); ok {
						obj[k] = dv
					} else {
						obj[k] = NewString(fmt.Sprintf("%v", v))
					}
				}
				dusoVal = NewObject(obj)
			case []any:
				// Convert Go array to Duso array
				arr := make([]Value, len(v))
				for i, elem := range v {
					if dv, ok := elem.(Value); ok {
						arr[i] = dv
					} else {
						arr[i] = NewString(fmt.Sprintf("%v", elem))
					}
				}
				dusoVal = NewArray(arr)
			case nil:
				dusoVal = NewNil()
			default:
				dusoVal = NewString(fmt.Sprintf("%v", v))
			}

			templateEnv.Define(key, dusoVal)
		}

		// Save current environment and switch to template environment
		prevEnv := templateEval.env
		templateEval.env = templateEnv
		defer func() { templateEval.env = prevEnv }()

		// Parse the template string
		tempParser := &Parser{filePath: "<template>"}
		templateNode, err := tempParser.ParseTemplateString(templateStr, NoPos)
		if err != nil {
			return nil, fmt.Errorf("template() parse error: %w", err)
		}

		// Evaluate the template
		var result Value
		switch n := templateNode.(type) {
		case *TemplateLiteral:
			result, err = templateEval.evalTemplateLiteral(n)
		case *StringLiteral:
			result = NewString(n.Value)
		default:
			val, err := templateEval.Eval(n)
			if err != nil {
				return nil, err
			}
			result = val
		}

		if err != nil {
			return nil, err
		}

		return result.AsString(), nil
	}

	return NewGoFunction(templateFn), nil
}

// Regex functions

// builtinFind finds all matches of a pattern in a string
func (b *Builtins) builtinFind(evaluator *Evaluator, args map[string]any) (any, error) {
	s, ok := args["0"].(string)
	if !ok {
		return nil, fmt.Errorf("find() requires a string as first argument")
	}

	pattern, ok := args["1"].(string)
	if !ok {
		return nil, fmt.Errorf("find() requires a string pattern as second argument")
	}

	ignoreCase := false
	if ic, ok := args["ignore_case"].(bool); ok {
		ignoreCase = ic
	}

	// Add case-insensitive flag if needed
	if ignoreCase {
		pattern = "(?i)" + pattern
	}

	// Compile regex
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("find() invalid regex: %v", err)
	}

	// Find all matches
	matches := re.FindAllStringIndex(s, -1)
	var result []Value
	for _, match := range matches {
		start := match[0]
		end := match[1]
		text := s[start:end]

		matchObj := make(map[string]Value)
		matchObj["text"] = NewString(text)
		matchObj["pos"] = NewNumber(float64(start))
		matchObj["len"] = NewNumber(float64(len(text)))

		result = append(result, NewObject(matchObj))
	}

	return NewArray(result), nil
}

// builtinReplace replaces matches of a pattern with a string or function result
func (b *Builtins) builtinReplace(evaluator *Evaluator, args map[string]any) (any, error) {
	s, ok := args["0"].(string)
	if !ok {
		return nil, fmt.Errorf("replace() requires a string as first argument")
	}

	pattern, ok := args["1"].(string)
	if !ok {
		return nil, fmt.Errorf("replace() requires a string pattern as second argument")
	}

	replacement, ok := args["2"]
	if !ok {
		return nil, fmt.Errorf("replace() requires a replacement (string or function) as third argument")
	}

	ignoreCase := false
	if ic, ok := args["ignore_case"].(bool); ok {
		ignoreCase = ic
	}

	// Add case-insensitive flag if needed
	if ignoreCase {
		pattern = "(?i)" + pattern
	}

	// Compile regex
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("replace() invalid regex: %v", err)
	}

	// Handle string replacement
	if replacementStr, ok := replacement.(string); ok {
		result := re.ReplaceAllString(s, replacementStr)
		return result, nil
	}

	// Handle function replacement
	if evaluator == nil {
		return nil, fmt.Errorf("replace() requires evaluator context for function replacement")
	}

	fn := interfaceToValue(replacement)
	if fn.Type != VAL_FUNCTION {
		return nil, fmt.Errorf("replace() requires replacement to be a string or function")
	}

	// Find all matches and replace with function results
	matches := re.FindAllStringIndex(s, -1)
	result := s
	offset := 0 // Track offset as we replace

	for _, match := range matches {
		start := match[0] + offset
		end := match[1] + offset
		text := result[start:end]

		// Call the replacement function with (text, pos, len)
		args := []Value{
			NewString(text),
			NewNumber(float64(match[0])), // Original position in original string
			NewNumber(float64(len(text))),
		}
		replacementResult, err := b.callUserFunction(evaluator, fn, args)
		if err != nil {
			return nil, fmt.Errorf("replace() function error: %v", err)
		}

		replacementText := replacementResult.AsString()

		// Replace in result
		result = result[:start] + replacementText + result[end:]
		offset += len(replacementText) - (end - start)
	}

	return result, nil
}

// Math functions

// builtinFloor rounds down to nearest integer
func (b *Builtins) builtinFloor(evaluator *Evaluator, args map[string]any) (any, error) {
	if arg, ok := args["0"].(float64); ok {
		return math.Floor(arg), nil
	}
	return nil, fmt.Errorf("floor() requires a number")
}

// builtinCeil rounds up to nearest integer
func (b *Builtins) builtinCeil(evaluator *Evaluator, args map[string]any) (any, error) {
	if arg, ok := args["0"].(float64); ok {
		return math.Ceil(arg), nil
	}
	return nil, fmt.Errorf("ceil() requires a number")
}

// builtinRound rounds to nearest integer
func (b *Builtins) builtinRound(evaluator *Evaluator, args map[string]any) (any, error) {
	if arg, ok := args["0"].(float64); ok {
		return math.Round(arg), nil
	}
	return nil, fmt.Errorf("round() requires a number")
}

// builtinAbs returns absolute value
func (b *Builtins) builtinAbs(evaluator *Evaluator, args map[string]any) (any, error) {
	if arg, ok := args["0"].(float64); ok {
		return math.Abs(arg), nil
	}
	return nil, fmt.Errorf("abs() requires a number")
}

// builtinMin returns minimum of arguments
// minMaxHelper computes min/max of numeric arguments
func (b *Builtins) minMaxHelper(args map[string]any, isMin bool) (any, error) {
	if len(args) == 0 {
		name := "min()"
		if !isMin {
			name = "max()"
		}
		return nil, fmt.Errorf("%s requires at least one argument", name)
	}

	var result float64
	var set bool

	for i := 0; ; i++ {
		key := fmt.Sprintf("%d", i)
		val, ok := args[key].(float64)
		if !ok {
			break
		}
		if !set {
			result = val
			set = true
		} else if isMin && val < result {
			result = val
		} else if !isMin && val > result {
			result = val
		}
	}

	if !set {
		name := "min()"
		if !isMin {
			name = "max()"
		}
		return nil, fmt.Errorf("%s requires numeric arguments", name)
	}
	return result, nil
}

func (b *Builtins) builtinMin(evaluator *Evaluator, args map[string]any) (any, error) {
	return b.minMaxHelper(args, true)
}

// builtinMax returns maximum of arguments
func (b *Builtins) builtinMax(evaluator *Evaluator, args map[string]any) (any, error) {
	return b.minMaxHelper(args, false)
}

// builtinSqrt returns square root
func (b *Builtins) builtinSqrt(evaluator *Evaluator, args map[string]any) (any, error) {
	if arg, ok := args["0"].(float64); ok {
		return math.Sqrt(arg), nil
	}
	return nil, fmt.Errorf("sqrt() requires a number")
}

// builtinPow returns x^y
func (b *Builtins) builtinPow(evaluator *Evaluator, args map[string]any) (any, error) {
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
func (b *Builtins) builtinClamp(evaluator *Evaluator, args map[string]any) (any, error) {
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

// Trigonometric functions

// builtinSin returns sine of angle in radians
func (b *Builtins) builtinSin(evaluator *Evaluator, args map[string]any) (any, error) {
	if arg, ok := args["0"].(float64); ok {
		return math.Sin(arg), nil
	}
	return nil, fmt.Errorf("sin() requires a number (angle in radians)")
}

// builtinCos returns cosine of angle in radians
func (b *Builtins) builtinCos(evaluator *Evaluator, args map[string]any) (any, error) {
	if arg, ok := args["0"].(float64); ok {
		return math.Cos(arg), nil
	}
	return nil, fmt.Errorf("cos() requires a number (angle in radians)")
}

// builtinTan returns tangent of angle in radians
func (b *Builtins) builtinTan(evaluator *Evaluator, args map[string]any) (any, error) {
	if arg, ok := args["0"].(float64); ok {
		return math.Tan(arg), nil
	}
	return nil, fmt.Errorf("tan() requires a number (angle in radians)")
}

// builtinAsin returns arcsine in radians (inverse of sine)
func (b *Builtins) builtinAsin(evaluator *Evaluator, args map[string]any) (any, error) {
	if arg, ok := args["0"].(float64); ok {
		return math.Asin(arg), nil
	}
	return nil, fmt.Errorf("asin() requires a number between -1 and 1")
}

// builtinAcos returns arccosine in radians (inverse of cosine)
func (b *Builtins) builtinAcos(evaluator *Evaluator, args map[string]any) (any, error) {
	if arg, ok := args["0"].(float64); ok {
		return math.Acos(arg), nil
	}
	return nil, fmt.Errorf("acos() requires a number between -1 and 1")
}

// builtinAtan returns arctangent in radians (inverse of tangent)
func (b *Builtins) builtinAtan(evaluator *Evaluator, args map[string]any) (any, error) {
	if arg, ok := args["0"].(float64); ok {
		return math.Atan(arg), nil
	}
	return nil, fmt.Errorf("atan() requires a number")
}

// builtinAtan2 returns arctangent of y/x in radians, handling quadrants correctly
func (b *Builtins) builtinAtan2(evaluator *Evaluator, args map[string]any) (any, error) {
	y, ok := args["0"].(float64)
	if !ok {
		return nil, fmt.Errorf("atan2() requires a number as first argument (y)")
	}

	x, ok := args["1"].(float64)
	if !ok {
		return nil, fmt.Errorf("atan2() requires a number as second argument (x)")
	}

	return math.Atan2(y, x), nil
}

// Exponential and logarithmic functions

// builtinExp returns e^x
func (b *Builtins) builtinExp(evaluator *Evaluator, args map[string]any) (any, error) {
	if arg, ok := args["0"].(float64); ok {
		return math.Exp(arg), nil
	}
	return nil, fmt.Errorf("exp() requires a number")
}

// builtinLog returns logarithm base 10
func (b *Builtins) builtinLog(evaluator *Evaluator, args map[string]any) (any, error) {
	if arg, ok := args["0"].(float64); ok {
		return math.Log10(arg), nil
	}
	return nil, fmt.Errorf("log() requires a number")
}

// builtinLn returns natural logarithm (base e)
func (b *Builtins) builtinLn(evaluator *Evaluator, args map[string]any) (any, error) {
	if arg, ok := args["0"].(float64); ok {
		return math.Log(arg), nil
	}
	return nil, fmt.Errorf("ln() requires a number")
}

// builtinPi returns the mathematical constant pi
func (b *Builtins) builtinPi(evaluator *Evaluator, args map[string]any) (any, error) {
	return math.Pi, nil
}

// Array/Object functions

// builtinKeys returns array of object keys or array indices
func (b *Builtins) builtinKeys(evaluator *Evaluator, args map[string]any) (any, error) {
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
func (b *Builtins) builtinValues(evaluator *Evaluator, args map[string]any) (any, error) {
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
func (b *Builtins) callComparisonFunction(evaluator *Evaluator, fn Value, valA, valB Value) (bool, error) {
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
				prevEnv := evaluator.env
				evaluator.env = scriptFn.Closure
				val, err := evaluator.Eval(param.Default)
				evaluator.env = prevEnv
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
		prevEnv := evaluator.env
		evaluator.env = fnEnv

		var result Value
		for _, stmt := range scriptFn.Body {
			val, err := evaluator.Eval(stmt)
			if returnVal, ok := err.(*ReturnValue); ok {
				result = returnVal.Value
				break
			}
			if err != nil {
				evaluator.env = prevEnv
				return false, err
			}
			result = val
		}

		evaluator.env = prevEnv

		// Convert result to boolean
		return result.IsTruthy(), nil
	}

	// Handle Go functions
	if goFn, ok := fn.Data.(GoFunction); ok {
		argMap := map[string]any{
			"0": valueToInterface(valA),
			"1": valueToInterface(valB),
		}
		res, err := goFn(evaluator, argMap)
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
func (b *Builtins) builtinSort(evaluator *Evaluator, args map[string]any) (any, error) {
	arrPtr, ok := args["0"].(*[]Value)
	if !ok {
		return nil, fmt.Errorf("sort() requires an array as first argument")
	}

	// Make a copy to avoid modifying original
	arr := *arrPtr
	result := make([]Value, len(arr))
	copy(result, arr)

	// Check if comparison function provided
	if compareFnArg, hasCompareFn := args["1"]; hasCompareFn {
		// Custom comparison function
		if evaluator == nil {
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

			less, err := b.callComparisonFunction(evaluator, compareFn, result[i], result[j])
			if err != nil {
				sortErr = err
				return false
			}
			return less
		})

		if sortErr != nil {
			return nil, sortErr
		}

		return &result, nil
	}

	// Default sort: compare by value
	sort.Slice(result, func(i, j int) bool {
		vi, vj := result[i], result[j]

		// Handle numeric comparison
		if vi.IsNumber() && vj.IsNumber() {
			return vi.AsNumber() < vj.AsNumber()
		}

		// Handle string comparison
		if vi.IsString() && vj.IsString() {
			return vi.AsString() < vj.AsString()
		}

		// Mixed types or unsupported - compare as strings
		return vi.String() < vj.String()
	})

	return &result, nil
}

// builtinMap applies a function to each element of an array (returns new array)
func (b *Builtins) builtinMap(evaluator *Evaluator, args map[string]any) (any, error) {
	arrPtr, ok := args["0"].(*[]Value)
	if !ok {
		return nil, fmt.Errorf("map() requires an array as first argument")
	}

	fnArg, ok := args["1"]
	if !ok {
		return nil, fmt.Errorf("map() requires a function as second argument")
	}

	if evaluator == nil {
		return nil, fmt.Errorf("map() requires evaluator context")
	}

	fn := interfaceToValue(fnArg)

	arr := *arrPtr
	result := make([]Value, len(arr))
	for i, item := range arr {
		retVal, err := b.callUserFunction(evaluator, fn, []Value{item})
		if err != nil {
			return nil, fmt.Errorf("error in map function: %w", err)
		}
		result[i] = retVal
	}

	return &result, nil
}

// builtinFilter keeps only array elements that match a predicate (returns new array)
func (b *Builtins) builtinFilter(evaluator *Evaluator, args map[string]any) (any, error) {
	arrPtr, ok := args["0"].(*[]Value)
	if !ok {
		return nil, fmt.Errorf("filter() requires an array as first argument")
	}

	fnArg, ok := args["1"]
	if !ok {
		return nil, fmt.Errorf("filter() requires a function as second argument")
	}

	if evaluator == nil {
		return nil, fmt.Errorf("filter() requires evaluator context")
	}

	fn := interfaceToValue(fnArg)

	arr := *arrPtr
	result := make([]Value, 0, len(arr))
	for _, item := range arr {
		retVal, err := b.callUserFunction(evaluator, fn, []Value{item})
		if err != nil {
			return nil, fmt.Errorf("error in filter function: %w", err)
		}
		if retVal.IsTruthy() {
			result = append(result, item)
		}
	}

	return &result, nil
}

// builtinPush appends items to the end of an array, returns new length
func (b *Builtins) builtinPush(evaluator *Evaluator, args map[string]any) (any, error) {
	arrPtr, ok := args["0"].(*[]Value)
	if !ok {
		return nil, fmt.Errorf("push() requires an array as first argument")
	}

	// Fast path: single item (common case in loops)
	if itemArg, ok := args["1"]; ok {
		if _, ok := args["2"]; !ok {
			// Only one item to push - avoid temporary slice allocation
			*arrPtr = append(*arrPtr, interfaceToValue(itemArg))
			return float64(len(*arrPtr)), nil
		}
	}

	// Slow path: multiple items
	var items []Value
	i := 1
	for {
		key := fmt.Sprintf("%d", i)
		if itemArg, ok := args[key]; ok {
			items = append(items, interfaceToValue(itemArg))
			i++
		} else {
			break
		}
	}

	*arrPtr = append(*arrPtr, items...)
	return float64(len(*arrPtr)), nil
}

// builtinPop removes and returns the last element of an array
func (b *Builtins) builtinPop(evaluator *Evaluator, args map[string]any) (any, error) {
	arrPtr, ok := args["0"].(*[]Value)
	if !ok {
		return nil, fmt.Errorf("pop() requires an array as first argument")
	}

	arr := *arrPtr
	if len(arr) == 0 {
		return nil, nil
	}

	last := arr[len(arr)-1]
	*arrPtr = arr[:len(arr)-1]
	return last, nil
}

// builtinShift removes and returns the first element of an array
func (b *Builtins) builtinShift(evaluator *Evaluator, args map[string]any) (any, error) {
	arrPtr, ok := args["0"].(*[]Value)
	if !ok {
		return nil, fmt.Errorf("shift() requires an array as first argument")
	}

	arr := *arrPtr
	if len(arr) == 0 {
		return nil, nil
	}

	first := arr[0]
	*arrPtr = arr[1:]
	return first, nil
}

// builtinUnshift prepends items to the beginning of an array, returns new length
func (b *Builtins) builtinUnshift(evaluator *Evaluator, args map[string]any) (any, error) {
	arrPtr, ok := args["0"].(*[]Value)
	if !ok {
		return nil, fmt.Errorf("unshift() requires an array as first argument")
	}

	// Fast path: single item (common case)
	if itemArg, ok := args["1"]; ok {
		if _, ok := args["2"]; !ok {
			// Only one item to unshift
			item := interfaceToValue(itemArg)
			newArr := make([]Value, len(*arrPtr)+1)
			newArr[0] = item
			copy(newArr[1:], *arrPtr)
			*arrPtr = newArr
			return float64(len(*arrPtr)), nil
		}
	}

	// Slow path: multiple items
	var items []Value
	i := 1
	for {
		key := fmt.Sprintf("%d", i)
		if itemArg, ok := args[key]; ok {
			items = append(items, interfaceToValue(itemArg))
			i++
		} else {
			break
		}
	}

	*arrPtr = append(items, *arrPtr...)
	return float64(len(*arrPtr)), nil
}

// builtinReduce combines all array elements into a single value
func (b *Builtins) builtinReduce(evaluator *Evaluator, args map[string]any) (any, error) {
	arrPtr, ok := args["0"].(*[]Value)
	if !ok {
		return nil, fmt.Errorf("reduce() requires an array as first argument")
	}

	fnArg, ok := args["1"]
	if !ok {
		return nil, fmt.Errorf("reduce() requires a function as second argument")
	}

	if evaluator == nil {
		return nil, fmt.Errorf("reduce() requires evaluator context")
	}

	fn := interfaceToValue(fnArg)

	// Get initial value (third argument)
	accumulator := NewNil()
	if initVal, ok := args["2"]; ok {
		accumulator = interfaceToValue(initVal)
	}

	// Iterate through array
	arr := *arrPtr
	for _, item := range arr {
		retVal, err := b.callUserFunction(evaluator, fn, []Value{accumulator, item})
		if err != nil {
			return nil, fmt.Errorf("error in reduce function: %w", err)
		}
		accumulator = retVal
	}

	return accumulator, nil
}

// callUserFunction calls a user function with the given arguments
func (b *Builtins) callUserFunction(evaluator *Evaluator, fn Value, args []Value) (Value, error) {
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
				prevEnv := evaluator.env
				evaluator.env = scriptFn.Closure
				val, err := evaluator.Eval(param.Default)
				evaluator.env = prevEnv
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
		prevEnv := evaluator.env
		evaluator.env = fnEnv

		var result Value
		for _, stmt := range scriptFn.Body {
			val, err := evaluator.Eval(stmt)
			if returnVal, ok := err.(*ReturnValue); ok {
				result = returnVal.Value
				break
			}
			if err != nil {
				evaluator.env = prevEnv
				return NewNil(), err
			}
			result = val
		}

		evaluator.env = prevEnv
		return result, nil
	}

	// Handle Go functions
	if goFn, ok := fn.Data.(GoFunction); ok {
		argMap := make(map[string]any)
		for i, arg := range args {
			argMap[fmt.Sprintf("%d", i)] = ValueToInterface(arg)
		}
		ret, err := goFn(evaluator, argMap)
		if err != nil {
			return NewNil(), err
		}
		return interfaceToValue(ret), nil
	}

	return NewNil(), fmt.Errorf("not a callable function")
}

// Utility functions

// builtinRange creates an array of numbers in range
func (b *Builtins) builtinRange(evaluator *Evaluator, args map[string]any) (any, error) {
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

// builtinRandom returns a random float between 0 and 1
func (b *Builtins) builtinRandom(evaluator *Evaluator, args map[string]any) (any, error) {
	rng := mathrand.New(mathrand.NewSource(time.Now().UnixNano()))
	return rng.Float64(), nil
}

// builtinUUID generates a UUID v7 (RFC 9562)
// UUID v7 is time-sorted with 48-bit Unix timestamp in milliseconds followed by random data
func (b *Builtins) builtinUUID(evaluator *Evaluator, args map[string]any) (any, error) {
	buf := make([]byte, 16)

	// 48-bit timestamp (Unix epoch in milliseconds)
	binary.BigEndian.PutUint64(buf[0:8], uint64(time.Now().UnixMilli()))

	// Truncate timestamp to 6 bytes, shifting because PutUint64 writes 8 bytes
	copy(buf[0:6], buf[2:8])

	// 10 bytes random data
	if _, err := rand.Read(buf[6:16]); err != nil {
		return nil, fmt.Errorf("uuid() failed to generate random bytes: %v", err)
	}

	// Version 7: set version bits to 0111 in the 7th byte
	buf[6] = (buf[6] & 0x0f) | 0x70

	// Variant: set variant bits to 10 in the 9th byte
	buf[8] = (buf[8] & 0x3f) | 0x80

	// Format as UUID string: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	return fmt.Sprintf("%x-%x-%x-%x-%x", buf[0:4], buf[4:6], buf[6:8], buf[8:10], buf[10:16]), nil
}

// System functions

// builtinExit stops execution and returns values to host
func (b *Builtins) builtinExit(evaluator *Evaluator, args map[string]any) (any, error) {
	// Collect all arguments as return values
	values := make([]any, 0)
	for i := 0; ; i++ {
		key := fmt.Sprintf("%d", i)
		if val, ok := args[key]; ok {
			// Deep copy to isolate return values from parent scope
			values = append(values, DeepCopyAny(val))
		} else {
			break
		}
	}

	return nil, &ExitExecution{Values: values}
}

// builtinThrow throws an error with message and call stack
func (b *Builtins) builtinThrow(evaluator *Evaluator, args map[string]any) (any, error) {
	// Accept any value type, no deep copy at throw time
	// Will be deep copied only if it crosses process boundaries (run())
	var value any = "unknown error"

	if msg, ok := args["0"]; ok {
		value = msg
	} else if msg, ok := args["message"]; ok {
		value = msg
	}

	// Create DusoError with call stack, storing the original value
	err := &DusoError{
		Message: value,
	}

	if evaluator != nil && evaluator.ctx != nil {
		err.FilePath = evaluator.ctx.FilePath
		err.CallStack = evaluator.ctx.CallStack
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
// Optional arguments are passed as a debug message (not printed directly)
func (b *Builtins) builtinBreakpoint(evaluator *Evaluator, args map[string]any) (any, error) {
	// Only trigger breakpoint if debug mode is enabled
	if evaluator == nil || !evaluator.DebugMode {
		return nil, nil
	}

	// If arguments provided, format them as a debug message
	var message string
	if len(args) > 0 {
		message = "BREAKPOINT: " + b.formatArgs(args)
	}

	// Capture call stack and current environment for debug display
	// Clone the call stack so it can't be modified
	callStack := make([]CallFrame, len(evaluator.ctx.CallStack))
	copy(callStack, evaluator.ctx.CallStack)

	err := &BreakpointError{
		FilePath:  evaluator.ctx.FilePath,
		CallStack: callStack,
		Env:       evaluator.env, // Capture current environment for scope access
		Message:   message,         // Pass message to debug handler
	}
	return nil, err
}

// builtinWatch evaluates expressions and breaks if values change
// Each argument is a string expression to watch
func (b *Builtins) builtinWatch(evaluator *Evaluator, args map[string]any) (any, error) {
	if evaluator == nil {
		return nil, fmt.Errorf("watch() requires evaluator context")
	}

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

		val, err := evaluator.Eval(node)
		if err != nil {
			return nil, fmt.Errorf("watch() evaluation error in '{{%s}}': %v", expr, err)
		}

		// Check if value changed from cached
		cachedVal, exists := evaluator.watchCache[expr]
		if !exists || !b.valuesEqual(val, cachedVal) {
			// Value changed or first time seeing it
			evaluator.watchCache[expr] = val
			triggered = append(triggered, fmt.Sprintf("WATCH: %s = %v", expr, val.String()))
		}
	}

	// If any watches triggered and debug mode is enabled, create breakpoint with messages
	if len(triggered) > 0 && evaluator.DebugMode {
		// Combine all triggered messages
		message := strings.Join(triggered, "\n")

		// Trigger breakpoint with call stack
		callStack := make([]CallFrame, len(evaluator.ctx.CallStack))
		copy(callStack, evaluator.ctx.CallStack)

		err := &BreakpointError{
			FilePath:  evaluator.ctx.FilePath,
			CallStack: callStack,
			Env:       evaluator.env,
			Message:   message, // Pass all watch messages to debug handler
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
func (b *Builtins) builtinNow(evaluator *Evaluator, args map[string]any) (any, error) {
	return float64(time.Now().UnixMilli()), nil
}

// builtinFormatTime formats a Unix timestamp to string
func (b *Builtins) builtinFormatTime(evaluator *Evaluator, args map[string]any) (any, error) {
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
func (b *Builtins) builtinParseTime(evaluator *Evaluator, args map[string]any) (any, error) {
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
		"2006-01-02T15:04:05Z", // ISO with Z
		"2006-01-02T15:04:05",  // ISO without Z
		"2006-01-02 15:04:05",  // Default
		"2006-01-02",           // Date only
		"January 2, 2006",      // Long date
		"Mon January 2, 2006",  // Long date with day of week
		"Jan 2, 2006",          // Short date
		"Mon Jan 2, 2006",      // Short date with day of week
	}

	for _, format := range commonFormats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return float64(t.Unix()), nil
		}
	}

	return nil, fmt.Errorf("parse_time() could not parse %q - try providing a format", dateStr)
}

// builtinSleep pauses execution for the specified duration in seconds (default: 1)
func (b *Builtins) builtinSleep(evaluator *Evaluator, args map[string]any) (any, error) {
	seconds := 1.0 // Default to 1 second
	if arg, ok := args["0"]; ok {
		num, ok := arg.(float64)
		if !ok {
			return nil, fmt.Errorf("sleep() requires a number (seconds)")
		}
		if num < 0 {
			return nil, fmt.Errorf("sleep() duration cannot be negative")
		}
		seconds = num
	}
	time.Sleep(time.Duration(seconds * float64(time.Second)))
	return nil, nil
}

// builtinParseJSON parses a JSON string into Duso objects/arrays
func (b *Builtins) builtinParseJSON(evaluator *Evaluator, args map[string]any) (any, error) {
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
func (b *Builtins) builtinFormatJSON(evaluator *Evaluator, args map[string]any) (any, error) {
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
	case *[]Value:
		// Handle new array pointer type - convert to JSON array
		arr := make([]any, len(*val))
		for i, v := range *val {
			arr[i] = b.valueToJSON(ValueToInterface(v))
		}
		return arr
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
func (b *Builtins) builtinParallel(evaluator *Evaluator, args map[string]any) (any, error) {
	if evaluator == nil {
		return nil, fmt.Errorf("parallel() requires evaluator context")
	}

	// Case 1: Single array argument parallel([fn1, fn2, fn3])
	if arrPtr, ok := args["0"].(*[]Value); ok && len(args) == 1 {
		// Convert *[]Value to []any for parallelArrayWithEval
		arr := make([]any, len(*arrPtr))
		for i, v := range *arrPtr {
			arr[i] = &ValueRef{Val: v}
		}
		return b.parallelArrayWithEval(evaluator, arr)
	}

	// Case 2: Single object argument parallel({a = fn1, b = fn2})
	if obj, ok := args["0"].(map[string]any); ok && len(args) == 1 {
		return b.parallelObjectWithEval(evaluator, obj)
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
		return b.parallelArrayWithEval(evaluator, varargs)
	}

	return nil, fmt.Errorf("parallel() requires an array, object, or functions as arguments")
}

// parallelArrayWithEval executes an array of functions in parallel with isolated evaluators
func (b *Builtins) parallelArrayWithEval(evaluator *Evaluator, functions []any) (any, error) {
	results := make([]any, len(functions))

	var wg sync.WaitGroup
	for i, fnArg := range functions {
		wg.Add(1)
		go func(index int, fn any) {
			defer wg.Done()

			// Create a child evaluator for this block with parent scope access
			childEval := NewEvaluator()
			childEval.env.parent = evaluator.env
			childEval.env.isParallelContext = true
			childEval.isParallelContext = true // Block parent scope writes

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
func (b *Builtins) parallelObjectWithEval(evaluator *Evaluator, functions map[string]any) (any, error) {
	results := make(map[string]any)
	var mu sync.Mutex

	var wg sync.WaitGroup
	for key, fnArg := range functions {
		wg.Add(1)
		go func(k string, fn any) {
			defer wg.Done()

			// Create a child evaluator for this block with parent scope access
			childEval := NewEvaluator()
			childEval.env.parent = evaluator.env
			childEval.env.isParallelContext = true
			childEval.isParallelContext = true // Block parent scope writes

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
		ret, err := goFn(eval, argMap)
		if err != nil {
			return NewNil(), err
		}
		return interfaceToValue(ret), nil
	}

	return NewNil(), fmt.Errorf("not a callable function")
}

// builtinDatastore creates a thread-safe namespaced key/value store
func (b *Builtins) builtinDatastore(evaluator *Evaluator, args map[string]any) (any, error) {
	// Get namespace from first positional or named argument
	var namespace string

	if ns, ok := args["0"]; ok {
		// Positional argument
		namespace = fmt.Sprintf("%v", ns)
	} else if ns, ok := args["namespace"]; ok {
		// Named argument
		namespace = fmt.Sprintf("%v", ns)
	} else {
		return nil, fmt.Errorf("datastore() requires a namespace argument")
	}

	// Get config from second positional or named argument (optional)
	var config map[string]any

	if cfg, ok := args["1"]; ok {
		// Positional argument
		if cfgMap, ok := cfg.(map[string]any); ok {
			config = cfgMap
		}
	} else if cfg, ok := args["config"]; ok {
		// Named argument
		if cfgMap, ok := cfg.(map[string]any); ok {
			config = cfgMap
		}
	}

	// sys datastore is read-only and rejects any config
	if namespace == "sys" {
		if len(config) > 0 {
			return nil, fmt.Errorf("datastore(\"sys\") does not accept configuration options")
		}
	}

	// Get or create the datastore
	store := GetDatastore(namespace, config)

	// Create set(key, value) method
	setFn := NewGoFunction(func(setEval *Evaluator, setArgs map[string]any) (any, error) {
		if namespace == "sys" {
			return nil, fmt.Errorf("datastore(\"sys\") is read-only")
		}
		key, ok := setArgs["0"].(string)
		if !ok {
			return nil, fmt.Errorf("set() requires key (string) and value arguments")
		}
		value, ok := setArgs["1"]
		if !ok {
			return nil, fmt.Errorf("set() requires key and value arguments")
		}
		return nil, store.Set(key, value)
	})

	// Create get(key) method
	getFn := NewGoFunction(func(getEval *Evaluator, getArgs map[string]any) (any, error) {
		key, ok := getArgs["0"].(string)
		if !ok {
			return nil, fmt.Errorf("get() requires a key (string) argument")
		}
		return store.Get(key)
	})

	// Create increment(key, delta) method
	incrementFn := NewGoFunction(func(incEval *Evaluator, incArgs map[string]any) (any, error) {
		if namespace == "sys" {
			return nil, fmt.Errorf("datastore(\"sys\") is read-only")
		}
		key, ok := incArgs["0"].(string)
		if !ok {
			return nil, fmt.Errorf("increment() requires key (string) and delta arguments")
		}
		delta, ok := incArgs["1"].(float64)
		if !ok {
			return nil, fmt.Errorf("increment() requires a numeric delta argument")
		}
		return store.Increment(key, delta)
	})

	// Create push(key, item) method
	pushFn := NewGoFunction(func(appEval *Evaluator, appArgs map[string]any) (any, error) {
		if namespace == "sys" {
			return nil, fmt.Errorf("datastore(\"sys\") is read-only")
		}
		key, ok := appArgs["0"].(string)
		if !ok {
			return nil, fmt.Errorf("push() requires a key (string) argument")
		}
		item, ok := appArgs["1"]
		if !ok {
			return nil, fmt.Errorf("push() requires an item argument")
		}
		return store.Push(key, item)
	})

	// Create wait(key [, expectedValue]) method
	waitFn := NewGoFunction(func(waitEval *Evaluator, waitArgs map[string]any) (any, error) {
		key, ok := waitArgs["0"].(string)
		if !ok {
			return nil, fmt.Errorf("wait() requires a key (string) argument")
		}

		// Check if expectedValue provided
		expectedValue, hasExpectedValue := waitArgs["1"]

		// Check for timeout (optional)
		timeout := time.Duration(0)
		if timeoutArg, ok := waitArgs["2"]; ok {
			if timeoutSecs, ok := timeoutArg.(float64); ok {
				timeout = time.Duration(timeoutSecs) * time.Second
			}
		} else if timeoutArg, ok := waitArgs["timeout"]; ok {
			if timeoutSecs, ok := timeoutArg.(float64); ok {
				timeout = time.Duration(timeoutSecs) * time.Second
			}
		}

		value, err := store.Wait(key, expectedValue, hasExpectedValue, timeout)
		return value, err
	})

	// Create wait_for(key, predicate [, timeout]) method
	waitForFn := NewGoFunction(func(wfEval *Evaluator, wfArgs map[string]any) (any, error) {
		key, ok := wfArgs["0"].(string)
		if !ok {
			return nil, fmt.Errorf("wait_for() requires a key (string) argument")
		}

		predicateArg, ok := wfArgs["1"]
		if !ok {
			return nil, fmt.Errorf("wait_for() requires a predicate function argument")
		}

		// Extract GoFunction from the argument
		var predicateFn GoFunction

		if goFn, ok := predicateArg.(GoFunction); ok {
			// Direct GoFunction
			predicateFn = goFn
		} else if vr, ok := predicateArg.(*ValueRef); ok {
			// Wrapped in ValueRef - extract the function
			if vr.Val.IsFunction() {
				if goFn, ok := vr.Val.Data.(GoFunction); ok {
					predicateFn = goFn
				} else {
					return nil, fmt.Errorf("wait_for() predicate must be a Go function")
				}
			} else {
				return nil, fmt.Errorf("wait_for() predicate must be a function")
			}
		} else {
			return nil, fmt.Errorf("wait_for() predicate must be a function")
		}

		// Check for timeout (optional)
		timeout := time.Duration(0)
		if timeoutArg, ok := wfArgs["2"]; ok {
			if timeoutSecs, ok := timeoutArg.(float64); ok {
				timeout = time.Duration(timeoutSecs) * time.Second
			}
		} else if timeoutArg, ok := wfArgs["timeout"]; ok {
			if timeoutSecs, ok := timeoutArg.(float64); ok {
				timeout = time.Duration(timeoutSecs) * time.Second
			}
		}

		value, err := store.WaitFor(wfEval, key, predicateFn, timeout)
		return value, err
	})

	// Create delete(key) method
	deleteFn := NewGoFunction(func(delEval *Evaluator, delArgs map[string]any) (any, error) {
		if namespace == "sys" {
			return nil, fmt.Errorf("datastore(\"sys\") is read-only")
		}
		key, ok := delArgs["0"].(string)
		if !ok {
			return nil, fmt.Errorf("delete() requires a key (string) argument")
		}
		return nil, store.Delete(key)
	})

	// Create clear() method
	clearFn := NewGoFunction(func(clearEval *Evaluator, clearArgs map[string]any) (any, error) {
		if namespace == "sys" {
			return nil, fmt.Errorf("datastore(\"sys\") is read-only")
		}
		return nil, store.Clear()
	})

	// Create save() method
	saveFn := NewGoFunction(func(saveEval *Evaluator, saveArgs map[string]any) (any, error) {
		if namespace == "sys" {
			return nil, fmt.Errorf("datastore(\"sys\") is read-only")
		}
		return nil, store.Save()
	})

	// Create load() method
	loadFn := NewGoFunction(func(loadEval *Evaluator, loadArgs map[string]any) (any, error) {
		if namespace == "sys" {
			return nil, fmt.Errorf("datastore(\"sys\") is read-only")
		}
		return nil, store.Load()
	})

	// Return store object with methods
	return map[string]any{
		"set":       setFn,
		"get":       getFn,
		"increment": incrementFn,
		"push":      pushFn,
		"wait":      waitFn,
		"wait_for":  waitForFn,
		"delete":    deleteFn,
		"clear":     clearFn,
		"save":      saveFn,
		"load":      loadFn,
	}, nil
}

// builtinDeepCopy creates a deep copy of a value (recursively copies arrays and objects)
func (b *Builtins) builtinDeepCopy(evaluator *Evaluator, args map[string]any) (any, error) {
	val, ok := args["0"]
	if !ok {
		return nil, fmt.Errorf("deep_copy() requires 1 argument")
	}

	scriptVal := interfaceToValue(val)
	return b.deepCopyValue(scriptVal), nil
}

// deepCopyValue recursively deep copies a value
// Functions are excluded from deep copy (they don't work out of scope)
func (b *Builtins) deepCopyValue(v Value) Value {
	switch v.Type {
	case VAL_ARRAY:
		arr := v.AsArray()
		newArr := make([]Value, len(arr))
		for i, item := range arr {
			newArr[i] = b.deepCopyValue(item)
		}
		return NewArray(newArr)

	case VAL_OBJECT:
		obj := v.AsObject()
		newObj := make(map[string]Value)
		for k, item := range obj {
			// Skip functions - they don't work out of scope
			if item.IsFunction() {
				continue
			}
			newObj[k] = b.deepCopyValue(item)
		}
		return NewObject(newObj)

	case VAL_FUNCTION:
		// Functions are not copied (they don't work out of scope)
		return NewNil()

	default:
		// Primitives (number, string, bool, nil) are immutable, return as-is
		return v
	}
}
