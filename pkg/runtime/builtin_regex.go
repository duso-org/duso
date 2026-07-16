package runtime

import (
	"fmt"
	"regexp"
	"unicode/utf8"

	"github.com/duso-org/duso/pkg/script"
)

// unwrapValue extracts a Value from various input types (raw Value, ValueRef, or any)
func unwrapValue(v any) script.Value {
	// If it's a ValueRef, unwrap it
	if vr, ok := v.(*script.ValueRef); ok {
		return vr.Val
	}
	// If it's already a Value, return it
	if val, ok := v.(script.Value); ok {
		return val
	}
	// Otherwise convert it
	return script.InterfaceToValue(v)
}

// builtinToRegex compiles a string pattern into a Regex value
func builtinToRegex(evaluator *Evaluator, args map[string]any) (any, error) {
	pattern, ok := GetArg(args, 0, "pattern").(string)
	if !ok {
		return nil, fmt.Errorf("toregex() requires a string pattern as first argument")
	}

	compiled, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("toregex() invalid pattern: %v", err)
	}

	return script.NewRegex(pattern, compiled), nil
}

// builtinContains checks if string contains substring or matches pattern
func builtinContains(evaluator *Evaluator, args map[string]any) (any, error) {
	s, ok := args["0"].(string)
	if !ok {
		return nil, fmt.Errorf("contains() requires a string as first argument")
	}

	patternArg := args["1"]
	if patternArg == nil {
		return nil, fmt.Errorf("contains() requires a pattern (string or regex) as second argument")
	}

	ignoreCase := false
	// Check positional argument first
	if ic, ok := args["2"].(bool); ok {
		ignoreCase = ic
	} else if ic, ok := args["ignore_case"].(bool); ok {
		// Fall back to named parameter
		ignoreCase = ic
	}

	var re *regexp.Regexp
	var err error

	// Unwrap the pattern argument (may be wrapped in ValueRef)
	patternVal := unwrapValue(patternArg)

	// Handle regex value or string literal
	if patternVal.IsRegex() {
		regex := patternVal.AsRegex()
		if regex == nil {
			return nil, fmt.Errorf("contains() invalid regex value")
		}
		re = regex.Compiled
	} else if patternVal.IsString() {
		// Treat string as literal (escape special regex characters)
		pattern := regexp.QuoteMeta(patternVal.AsString())
		// Add case-insensitive flag if needed
		if ignoreCase {
			pattern = "(?i)" + pattern
		}
		re, err = regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("contains() invalid pattern: %v", err)
		}
	} else {
		return nil, fmt.Errorf("contains() requires a pattern (string or regex) as second argument")
	}

	return re.MatchString(s), nil
}

// builtinStartsWith checks if string starts with a prefix or matches pattern
func builtinStartsWith(evaluator *Evaluator, args map[string]any) (any, error) {
	s, ok := args["0"].(string)
	if !ok {
		return nil, fmt.Errorf("starts_with() requires a string as first argument")
	}

	patternArg := args["1"]
	if patternArg == nil {
		return nil, fmt.Errorf("starts_with() requires a pattern (string or regex) as second argument")
	}

	ignoreCase := false
	// Check positional argument first
	if ic, ok := args["2"].(bool); ok {
		ignoreCase = ic
	} else if ic, ok := args["ignore_case"].(bool); ok {
		// Fall back to named parameter
		ignoreCase = ic
	}

	var re *regexp.Regexp
	var err error

	// Unwrap the pattern argument (may be wrapped in ValueRef)
	patternVal := unwrapValue(patternArg)

	// Handle regex value or string literal
	if patternVal.IsRegex() {
		regex := patternVal.AsRegex()
		if regex == nil {
			return nil, fmt.Errorf("starts_with() invalid regex value")
		}
		// Wrap regex with ^ anchor
		wrappedPattern := "^" + regex.Pattern
		re, err = regexp.Compile(wrappedPattern)
		if err != nil {
			return nil, fmt.Errorf("starts_with() invalid regex: %v", err)
		}
	} else if patternVal.IsString() {
		// Treat string as literal prefix
		pattern := "(?i)^" + regexp.QuoteMeta(patternVal.AsString())
		if !ignoreCase {
			pattern = "^" + regexp.QuoteMeta(patternVal.AsString())
		}
		re, err = regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("starts_with() invalid pattern: %v", err)
		}
	} else {
		return nil, fmt.Errorf("starts_with() requires a pattern (string or regex) as second argument")
	}

	return re.MatchString(s), nil
}

// builtinEndsWith checks if string ends with a suffix or matches pattern
func builtinEndsWith(evaluator *Evaluator, args map[string]any) (any, error) {
	s, ok := args["0"].(string)
	if !ok {
		return nil, fmt.Errorf("ends_with() requires a string as first argument")
	}

	patternArg := args["1"]
	if patternArg == nil {
		return nil, fmt.Errorf("ends_with() requires a pattern (string or regex) as second argument")
	}

	ignoreCase := false
	// Check positional argument first
	if ic, ok := args["2"].(bool); ok {
		ignoreCase = ic
	} else if ic, ok := args["ignore_case"].(bool); ok {
		// Fall back to named parameter
		ignoreCase = ic
	}

	var re *regexp.Regexp
	var err error

	// Unwrap the pattern argument (may be wrapped in ValueRef)
	patternVal := unwrapValue(patternArg)

	// Handle regex value or string literal
	if patternVal.IsRegex() {
		regex := patternVal.AsRegex()
		if regex == nil {
			return nil, fmt.Errorf("ends_with() invalid regex value")
		}
		// Wrap regex with $ anchor
		wrappedPattern := regex.Pattern + "$"
		re, err = regexp.Compile(wrappedPattern)
		if err != nil {
			return nil, fmt.Errorf("ends_with() invalid regex: %v", err)
		}
	} else if patternVal.IsString() {
		// Treat string as literal suffix
		pattern := "(?i)" + regexp.QuoteMeta(patternVal.AsString()) + "$"
		if !ignoreCase {
			pattern = regexp.QuoteMeta(patternVal.AsString()) + "$"
		}
		re, err = regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("ends_with() invalid pattern: %v", err)
		}
	} else {
		return nil, fmt.Errorf("ends_with() requires a pattern (string or regex) as second argument")
	}

	return re.MatchString(s), nil
}

// Regex functions

// builtinFind finds all matches of a pattern in a string
func builtinFind(evaluator *Evaluator, args map[string]any) (any, error) {
	s, ok := GetArg(args, 0, "str").(string)
	if !ok {
		return nil, fmt.Errorf("find() requires a string as first argument")
	}

	patternArg := GetArg(args, 1, "pattern")
	if patternArg == nil {
		return nil, fmt.Errorf("find() requires a pattern (string or regex) as second argument")
	}

	ignoreCase := false
	if ic, ok := GetArg(args, 2, "ignore_case").(bool); ok {
		ignoreCase = ic
	}

	var re *regexp.Regexp
	var err error

	// Unwrap the pattern argument (may be wrapped in ValueRef)
	patternVal := unwrapValue(patternArg)

	// Handle regex value or string literal
	if patternVal.IsRegex() {
		// Use compiled regex directly
		regex := patternVal.AsRegex()
		if regex == nil {
			return nil, fmt.Errorf("find() invalid regex value")
		}
		re = regex.Compiled
	} else if patternVal.IsString() {
		// Treat string as literal (escape special regex characters)
		pattern := regexp.QuoteMeta(patternVal.AsString())
		// Add case-insensitive flag if needed
		if ignoreCase {
			pattern = "(?i)" + pattern
		}
		re, err = regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("find() invalid pattern: %v", err)
		}
	} else {
		return nil, fmt.Errorf("find() requires a pattern (string or regex) as second argument")
	}

	// Find all matches and convert byte positions to character positions
	matches := re.FindAllStringIndex(s, -1)
	var result []Value
	for _, match := range matches {
		byteStart := match[0]
		byteEnd := match[1]
		text := s[byteStart:byteEnd]

		// Convert byte positions to character positions
		charPos := utf8.RuneCountInString(s[:byteStart])
		charLen := utf8.RuneCountInString(text)

		matchObj := make(map[string]Value)
		matchObj["text"] = NewString(text)
		matchObj["pos"] = NewNumber(float64(charPos))
		matchObj["len"] = NewNumber(float64(charLen))

		result = append(result, NewObject(matchObj))
	}

	return NewArray(result), nil
}

// builtinReplace replaces matches of a pattern with a string or function result
func builtinReplace(evaluator *Evaluator, args map[string]any) (any, error) {
	s, ok := GetArg(args, 0, "str").(string)
	if !ok {
		return nil, fmt.Errorf("replace() requires a string as first argument")
	}

	patternArg := GetArg(args, 1, "pattern")
	if patternArg == nil {
		return nil, fmt.Errorf("replace() requires a pattern (string or regex) as second argument")
	}

	replacement := GetArg(args, 2, "replacement")
	if replacement == nil {
		return nil, fmt.Errorf("replace() requires a replacement (string or function) as third argument")
	}

	ignoreCase := false
	if ic, ok := GetArg(args, 3, "ignore_case").(bool); ok {
		ignoreCase = ic
	}

	var re *regexp.Regexp
	var err error

	// Unwrap the pattern argument (may be wrapped in ValueRef)
	patternVal := unwrapValue(patternArg)

	// Handle regex value or string literal
	if patternVal.IsRegex() {
		// Use compiled regex directly
		regex := patternVal.AsRegex()
		if regex == nil {
			return nil, fmt.Errorf("replace() invalid regex value")
		}
		re = regex.Compiled
	} else if patternVal.IsString() {
		// Treat string as literal (escape special regex characters)
		pattern := regexp.QuoteMeta(patternVal.AsString())
		// Add case-insensitive flag if needed
		if ignoreCase {
			pattern = "(?i)" + pattern
		}
		re, err = regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("replace() invalid pattern: %v", err)
		}
	} else {
		return nil, fmt.Errorf("replace() requires a pattern (string or regex) as second argument")
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

	fn := InterfaceToValue(replacement)
	if fn.Type != VAL_FUNCTION {
		return nil, fmt.Errorf("replace() requires replacement to be a string or function")
	}

	// Find all matches and replace with function results
	matches := re.FindAllStringIndex(s, -1)
	result := s
	offset := 0 // Track offset as we replace

	for _, match := range matches {
		byteStart := match[0] + offset
		byteEnd := match[1] + offset
		text := result[byteStart:byteEnd]

		// Convert byte position to character position in original string
		charPos := utf8.RuneCountInString(s[:match[0]])
		charLen := utf8.RuneCountInString(text)

		// Call the replacement function with (text, pos, len) using public API
		fnArgs := map[string]Value{
			"0": NewString(text),
			"1": NewNumber(float64(charPos)),
			"2": NewNumber(float64(charLen)),
		}
		replacementResult, err := evaluator.CallFunction(fn, fnArgs)
		if err != nil {
			return nil, fmt.Errorf("replace() function error: %v", err)
		}

		replacementText := replacementResult.AsString()

		// Replace in result
		result = result[:byteStart] + replacementText + result[byteEnd:]
		offset += len(replacementText) - (byteEnd - byteStart)
	}

	return result, nil
}
