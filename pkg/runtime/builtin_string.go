package runtime

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
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
					s = strconv.FormatInt(int64(num), 10)
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
					s = strconv.FormatInt(int64(num), 10)
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

// builtinSubstr extracts substring: substr(str, pos [, length])
func builtinSubstr(evaluator *Evaluator, args map[string]any) (any, error) {
	s, ok := GetArg(args, 0, "str").(string)
	if !ok {
		return nil, fmt.Errorf("substr() requires a string as first argument")
	}

	startVal := GetArg(args, 1, "pos")
	start, ok := startVal.(float64)
	if !ok {
		return nil, fmt.Errorf("substr() requires a number as second argument")
	}
	startIdx := int(start)

	// Convert string to runes for character-based indexing
	runes := []rune(s)
	runeLen := len(runes)

	// Handle negative index (from end)
	if startIdx < 0 {
		startIdx = runeLen + startIdx
	}
	if startIdx < 0 {
		startIdx = 0
	}
	if startIdx >= runeLen {
		return "", nil
	}

	// If length provided, use it; otherwise take to end
	lengthVal := GetArg(args, 2, "length")
	if lengthVal != nil {
		if length, ok := lengthVal.(float64); ok {
			endIdx := startIdx + int(length)
			if endIdx > runeLen {
				endIdx = runeLen
			}
			return string(runes[startIdx:endIdx]), nil
		}
	}

	return string(runes[startIdx:]), nil
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

// builtinPadLeft pads a string on the left: pad_left(str, width [, char])
func builtinPadLeft(evaluator *Evaluator, args map[string]any) (any, error) {
	arg := GetArg(args, 0, "str")
	var s string
	if strVal, ok := arg.(string); ok {
		s = strVal
	} else {
		// Coerce to string using tostring logic
		if num, ok := arg.(float64); ok {
			if IsInteger(num) {
				s = strconv.FormatInt(int64(num), 10)
			} else {
				s = fmt.Sprintf("%v", num)
			}
		} else {
			s = fmt.Sprintf("%v", arg)
		}
	}

	widthVal := GetArg(args, 1, "width")
	width, ok := widthVal.(float64)
	if !ok {
		return nil, fmt.Errorf("pad_left() requires a number as second argument")
	}

	padChar := " "
	if charVal := GetArg(args, 2, "char"); charVal != nil {
		var char string
		if strVal, ok := charVal.(string); ok {
			char = strVal
		} else {
			if num, ok := charVal.(float64); ok {
				if IsInteger(num) {
					char = strconv.FormatInt(int64(num), 10)
				} else {
					char = fmt.Sprintf("%v", num)
				}
			} else {
				char = fmt.Sprintf("%v", charVal)
			}
		}
		if utf8.RuneCountInString(char) != 1 {
			return nil, fmt.Errorf("pad_left() pad character must be a single character")
		}
		padChar = char
	}

	w := int(width)
	currentLen := utf8.RuneCountInString(s)
	if currentLen >= w {
		return s, nil
	}

	padding := strings.Repeat(padChar, w-currentLen)
	return padding + s, nil
}

// builtinPadRight pads a string on the right: pad_right(str, width [, char])
func builtinPadRight(evaluator *Evaluator, args map[string]any) (any, error) {
	arg := GetArg(args, 0, "str")
	var s string
	if strVal, ok := arg.(string); ok {
		s = strVal
	} else {
		// Coerce to string using tostring logic
		if num, ok := arg.(float64); ok {
			if IsInteger(num) {
				s = strconv.FormatInt(int64(num), 10)
			} else {
				s = fmt.Sprintf("%v", num)
			}
		} else {
			s = fmt.Sprintf("%v", arg)
		}
	}

	widthVal := GetArg(args, 1, "width")
	width, ok := widthVal.(float64)
	if !ok {
		return nil, fmt.Errorf("pad_right() requires a number as second argument")
	}

	padChar := " "
	if charVal := GetArg(args, 2, "char"); charVal != nil {
		var char string
		if strVal, ok := charVal.(string); ok {
			char = strVal
		} else {
			if num, ok := charVal.(float64); ok {
				if IsInteger(num) {
					char = strconv.FormatInt(int64(num), 10)
				} else {
					char = fmt.Sprintf("%v", num)
				}
			} else {
				char = fmt.Sprintf("%v", charVal)
			}
		}
		if utf8.RuneCountInString(char) != 1 {
			return nil, fmt.Errorf("pad_right() pad character must be a single character")
		}
		padChar = char
	}

	w := int(width)
	currentLen := utf8.RuneCountInString(s)
	if currentLen >= w {
		return s, nil
	}

	padding := strings.Repeat(padChar, w-currentLen)
	return s + padding, nil
}
