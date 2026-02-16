package runtime

import "fmt"

// builtinLen returns the length of an array, object, or string (returns 0 for nil)
func builtinLen(evaluator *Evaluator, args map[string]any) (any, error) {
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
func builtinType(evaluator *Evaluator, args map[string]any) (any, error) {
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
func builtinToNumber(evaluator *Evaluator, args map[string]any) (any, error) {
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
func builtinToString(evaluator *Evaluator, args map[string]any) (any, error) {
	if arg, ok := args["0"]; ok {
		// Special handling for numbers: if it's a whole number, format as integer
		if num, ok := arg.(float64); ok {
			if IsInteger(num) {
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
func builtinToBool(evaluator *Evaluator, args map[string]any) (any, error) {
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
