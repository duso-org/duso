package runtime

import (
	"fmt"
	"strings"
)

// builtinUpper converts string to uppercase, coercing input to string if needed
func builtinUpper(evaluator *Evaluator, args map[string]any) (any, error) {
	if arg, ok := args["0"]; ok {
		var s string
		if strVal, ok := arg.(string); ok {
			s = strVal
		} else {
			// Coerce to string using tostring logic
			if num, ok := arg.(float64); ok {
				if IsInteger(num) {
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
func builtinLower(evaluator *Evaluator, args map[string]any) (any, error) {
	if arg, ok := args["0"]; ok {
		var s string
		if strVal, ok := arg.(string); ok {
			s = strVal
		} else {
			// Coerce to string using tostring logic
			if num, ok := arg.(float64); ok {
				if IsInteger(num) {
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
func builtinSubstr(evaluator *Evaluator, args map[string]any) (any, error) {
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
func builtinTrim(evaluator *Evaluator, args map[string]any) (any, error) {
	if arg, ok := args["0"]; ok {
		if s, ok := arg.(string); ok {
			return strings.TrimSpace(s), nil
		}
		return nil, fmt.Errorf("trim() requires a string")
	}
	return nil, fmt.Errorf("trim() requires an argument")
}

// builtinRepeat repeats a string: repeat(str, count)
func builtinRepeat(evaluator *Evaluator, args map[string]any) (any, error) {
	s, ok := args["0"].(string)
	if !ok {
		return nil, fmt.Errorf("repeat() requires a string as first argument")
	}

	count, ok := args["1"].(float64)
	if !ok {
		return nil, fmt.Errorf("repeat() requires a number as second argument")
	}

	n := int(count)
	if n < 0 {
		return nil, fmt.Errorf("repeat() count must be non-negative")
	}

	return strings.Repeat(s, n), nil
}

