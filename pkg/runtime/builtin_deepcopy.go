package runtime

import "fmt"

// builtinDeepCopy creates a deep copy of a value (recursively copies arrays and objects)
func builtinDeepCopy(evaluator *Evaluator, args map[string]any) (any, error) {
	val, ok := args["0"]
	if !ok {
		return nil, fmt.Errorf("deep_copy() requires 1 argument")
	}

	scriptVal := InterfaceToValue(val)
	return deepCopyValue(scriptVal), nil
}

// deepCopyValue recursively deep copies a value
// Functions are excluded from deep copy (they don't work out of scope)
func deepCopyValue(v Value) Value {
	switch v.Type {
	case VAL_ARRAY:
		arr := v.AsArray()
		newArr := make([]Value, len(arr))
		for i, item := range arr {
			newArr[i] = deepCopyValue(item)
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
			newObj[k] = deepCopyValue(item)
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
