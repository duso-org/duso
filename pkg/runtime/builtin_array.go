package runtime

import (
	"fmt"
	"strings"
)

// Array/Object functions

// builtinKeys returns array of object keys or array indices
func builtinKeys(evaluator *Evaluator, args map[string]any) (any, error) {
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
func builtinValues(evaluator *Evaluator, args map[string]any) (any, error) {
	if arg, ok := args["0"].(map[string]any); ok {
		values := make([]any, 0, len(arg))
		for _, v := range arg {
			values = append(values, v)
		}
		return values, nil
	}
	return nil, fmt.Errorf("values() requires an object")
}

// builtinPush appends items to the end of an array, returns new length
func builtinPush(evaluator *Evaluator, args map[string]any) (any, error) {
	arrPtr, ok := args["0"].(*[]Value)
	if !ok {
		return nil, fmt.Errorf("push() requires an array as first argument")
	}

	// Fast path: single item (common case in loops)
	if itemArg, ok := args["1"]; ok {
		if _, ok := args["2"]; !ok {
			// Only one item to push - avoid temporary slice allocation
			*arrPtr = append(*arrPtr, InterfaceToValue(itemArg))
			return float64(len(*arrPtr)), nil
		}
	}

	// Slow path: multiple items
	var items []Value
	i := 1
	for {
		key := fmt.Sprintf("%d", i)
		if itemArg, ok := args[key]; ok {
			items = append(items, InterfaceToValue(itemArg))
			i++
		} else {
			break
		}
	}

	*arrPtr = append(*arrPtr, items...)
	return float64(len(*arrPtr)), nil
}

// builtinPop removes and returns the last element of an array
func builtinPop(evaluator *Evaluator, args map[string]any) (any, error) {
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
func builtinShift(evaluator *Evaluator, args map[string]any) (any, error) {
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
func builtinUnshift(evaluator *Evaluator, args map[string]any) (any, error) {
	arrPtr, ok := args["0"].(*[]Value)
	if !ok {
		return nil, fmt.Errorf("unshift() requires an array as first argument")
	}

	// Fast path: single item (common case)
	if itemArg, ok := args["1"]; ok {
		if _, ok := args["2"]; !ok {
			// Only one item to unshift
			item := InterfaceToValue(itemArg)
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
			items = append(items, InterfaceToValue(itemArg))
			i++
		} else {
			break
		}
	}

	*arrPtr = append(items, *arrPtr...)
	return float64(len(*arrPtr)), nil
}

// builtinSplit splits string by separator
func builtinSplit(evaluator *Evaluator, args map[string]any) (any, error) {
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
func builtinJoin(evaluator *Evaluator, args map[string]any) (any, error) {
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

// builtinRange creates an array of numbers in range
func builtinRange(evaluator *Evaluator, args map[string]any) (any, error) {
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

