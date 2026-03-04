package runtime

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/duso-org/duso/pkg/script"
)

// builtinParseJSON parses a JSON string into Duso objects/arrays
func builtinParseJSON(evaluator *Evaluator, args map[string]any) (any, error) {
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
	return jsonToValue(result), nil
}

// jsonToValue recursively converts JSON-unmarshaled values to Duso values
func jsonToValue(v any) any {
	switch val := v.(type) {
	case map[string]any:
		// Convert JSON object to Duso object
		obj := make(map[string]any)
		for k, v := range val {
			obj[k] = jsonToValue(v)
		}
		return obj
	case []any:
		// Convert JSON array to Duso array
		arr := make([]any, len(val))
		for i, v := range val {
			arr[i] = jsonToValue(v)
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
func builtinFormatJSON(evaluator *Evaluator, args map[string]any) (any, error) {
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
	jsonValue := valueToJSON(value)

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
// Functions, errors, and code types (when alone) return nil to skip them
func valueToJSON(v any) any {
	// Handle ValueRef wrappers (used for code and error types)
	if ref, ok := v.(*script.ValueRef); ok {
		return valueToJSON(ref.Val)
	}

	// Handle Value structs
	if val, ok := v.(script.Value); ok {
		switch val.Type {
		case script.VAL_FUNCTION:
			return nil // Skip functions
		case script.VAL_ERROR:
			return nil // Skip errors
		case script.VAL_CODE:
			// Code types return only the source string
			if code, ok := val.Data.(*script.CodeValue); ok {
				return code.Source
			}
			return nil
		default:
			// Recurse on the wrapped data
			return valueToJSON(val.Data)
		}
	}

	switch val := v.(type) {
	case *[]Value:
		// Handle array of Value structs
		arr := make([]any, 0)
		for _, item := range *val {
			if converted := valueToJSON(item); converted != nil {
				arr = append(arr, converted)
			}
		}
		return arr
	case []any:
		// Handle generic array
		arr := make([]any, 0)
		for _, item := range val {
			if converted := valueToJSON(item); converted != nil {
				arr = append(arr, converted)
			}
		}
		return arr
	case map[string]any:
		// Handle generic object
		obj := make(map[string]any)
		for k, item := range val {
			if converted := valueToJSON(item); converted != nil {
				obj[k] = converted
			}
		}
		return obj
	case map[string]Value:
		// Handle Value object
		obj := make(map[string]any)
		for k, item := range val {
			if converted := valueToJSON(item); converted != nil {
				obj[k] = converted
			}
		}
		return obj
	case nil:
		return nil
	case bool:
		return val
	case float64:
		return val
	case string:
		return val
	case *ScriptFunction:
		return nil // Skip functions
	case error:
		return nil // Skip errors
	default:
		// Unknown type - stringify it
		return fmt.Sprintf("%v", val)
	}
}
