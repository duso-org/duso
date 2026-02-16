package runtime

import (
	"fmt"
	"regexp"
)

// builtinContains checks if string contains substring
func builtinContains(evaluator *Evaluator, args map[string]any) (any, error) {
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

// Regex functions

// builtinFind finds all matches of a pattern in a string
func builtinFind(evaluator *Evaluator, args map[string]any) (any, error) {
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
func builtinReplace(evaluator *Evaluator, args map[string]any) (any, error) {
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

	fn := InterfaceToValue(replacement)
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

		// Call the replacement function with (text, pos, len) using public API
		fnArgs := map[string]Value{
			"0": NewString(text),
			"1": NewNumber(float64(match[0])), // Original position in original string
			"2": NewNumber(float64(len(text))),
		}
		replacementResult, err := evaluator.CallFunction(fn, fnArgs)
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
