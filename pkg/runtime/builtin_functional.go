package runtime

import (
	"fmt"
	"sort"
)

// builtinMap applies a function to each element of an array (returns new array)
func builtinMap(evaluator *Evaluator, args map[string]any) (any, error) {
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

	fn := InterfaceToValue(fnArg)

	arr := *arrPtr
	result := make([]Value, len(arr))
	for i, item := range arr {
		// Use public CallFunction API
		fnArgs := map[string]Value{"0": item}
		retVal, err := evaluator.CallFunction(fn, fnArgs)
		if err != nil {
			return nil, fmt.Errorf("error in map function: %w", err)
		}
		result[i] = retVal
	}

	return &result, nil
}

// builtinFilter keeps only array elements that match a predicate (returns new array)
func builtinFilter(evaluator *Evaluator, args map[string]any) (any, error) {
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

	fn := InterfaceToValue(fnArg)

	arr := *arrPtr
	result := make([]Value, 0, len(arr))
	for _, item := range arr {
		// Use public CallFunction API
		fnArgs := map[string]Value{"0": item}
		retVal, err := evaluator.CallFunction(fn, fnArgs)
		if err != nil {
			return nil, fmt.Errorf("error in filter function: %w", err)
		}
		if retVal.IsTruthy() {
			result = append(result, item)
		}
	}

	return &result, nil
}

// builtinReduce combines all array elements into a single value
func builtinReduce(evaluator *Evaluator, args map[string]any) (any, error) {
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

	fn := InterfaceToValue(fnArg)

	// Get initial value (third argument)
	accumulator := NewNil()
	if initVal, ok := args["2"]; ok {
		accumulator = InterfaceToValue(initVal)
	}

	// Iterate through array
	arr := *arrPtr
	for _, item := range arr {
		// Use public CallFunction API with two arguments
		fnArgs := map[string]Value{
			"0": accumulator,
			"1": item,
		}
		retVal, err := evaluator.CallFunction(fn, fnArgs)
		if err != nil {
			return nil, fmt.Errorf("error in reduce function: %w", err)
		}
		accumulator = retVal
	}

	return accumulator, nil
}

// builtinSort sorts an array with optional comparison function
func builtinSort(evaluator *Evaluator, args map[string]any) (any, error) {
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
		compareFn := InterfaceToValue(compareFnArg)

		// Sort using the comparison function
		sortErr := error(nil)
		sort.Slice(result, func(i, j int) bool {
			if sortErr != nil {
				return false
			}

			// Use public CallFunction API
			fnArgs := map[string]Value{
				"0": result[i],
				"1": result[j],
			}
			less, err := evaluator.CallFunction(compareFn, fnArgs)
			if err != nil {
				sortErr = err
				return false
			}
			return less.IsTruthy()
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
