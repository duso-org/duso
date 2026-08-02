package runtime

import (
	"fmt"
	"strings"

	"github.com/duso-org/duso/pkg/script"
)

// JSON builtins. The encoding and decoding themselves live in json_encode.go
// and json_decode.go, which work on script.Value directly; these functions are
// just the two calling conventions on top of them.

// builtinParseJSON parses a JSON string into Duso objects/arrays
func builtinParseJSON(evaluator *Evaluator, args map[string]any) (any, error) {
	jsonStr, ok := args["0"].(string)
	if !ok {
		return nil, fmt.Errorf("parse_json() requires a string as first argument")
	}

	result, err := decodeJSON([]byte(jsonStr))
	if err != nil {
		return nil, fmt.Errorf("parse_json() failed to parse JSON: %v", err)
	}
	return result, nil
}

// fastParseJSON is the []Value form of parse_json (see builtin_fast.go)
func fastParseJSON(evaluator *Evaluator, args []Value) (Value, error) {
	if len(args) < 1 || !args[0].IsString() {
		return script.NewNil(), fmt.Errorf("parse_json() requires a string as first argument")
	}

	result, err := decodeJSON([]byte(args[0].AsString()))
	if err != nil {
		return script.NewNil(), fmt.Errorf("parse_json() failed to parse JSON: %v", err)
	}
	return result, nil
}

// builtinFormatJSON converts a Duso value to JSON string
func builtinFormatJSON(evaluator *Evaluator, args map[string]any) (any, error) {
	value, ok := args["0"]
	if !ok {
		return nil, fmt.Errorf("format_json() requires at least one argument")
	}

	var indent string
	if indentArg, ok := args["1"]; ok {
		switch i := indentArg.(type) {
		case float64:
			indent = indentSpaces(i)
		case string:
			indent = i
		}
	}

	result, err := encodeJSON(value, indent)
	if err != nil {
		return nil, fmt.Errorf("format_json() failed to serialize: %v", err)
	}
	return string(result), nil
}

// fastFormatJSON is the []Value form of format_json (see builtin_fast.go)
func fastFormatJSON(evaluator *Evaluator, args []Value) (Value, error) {
	if len(args) < 1 {
		return script.NewNil(), fmt.Errorf("format_json() requires at least one argument")
	}

	var indent string
	if len(args) > 1 {
		switch {
		case args[1].IsNumber():
			indent = indentSpaces(args[1].AsNumber())
		case args[1].IsString():
			indent = args[1].AsString()
		}
	}

	result, err := encodeValueJSON(args[0], indent)
	if err != nil {
		return script.NewNil(), fmt.Errorf("format_json() failed to serialize: %v", err)
	}
	return script.NewString(string(result)), nil
}

// indentSpaces turns a numeric indent argument into that many spaces.
func indentSpaces(n float64) string {
	count := int(n)
	if count <= 0 {
		return ""
	}
	return strings.Repeat(" ", count)
}
