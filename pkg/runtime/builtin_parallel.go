package runtime

import (
	"fmt"
	"sync"
)

// builtinParallel executes functions concurrently with isolated evaluators
// Each function runs in its own Evaluator instance with parent scope access (read-only)
// This enables true parallelism with no evaluator contention or shared mutable state
// Accepts: array of functions, object of functions, or varargs of functions
// Returns: results in same structure as input
// Error handling: all run regardless, errors become nil
func builtinParallel(evaluator *Evaluator, args map[string]any) (any, error) {
	if evaluator == nil {
		return nil, fmt.Errorf("parallel() requires evaluator context")
	}

	// Case 1: Single array argument parallel([fn1, fn2, fn3])
	if arrPtr, ok := args["0"].(*[]Value); ok && len(args) == 1 {
		// Convert *[]Value to []any for parallelArrayWithEval
		arr := make([]any, len(*arrPtr))
		for i, v := range *arrPtr {
			arr[i] = &ValueRef{Val: v}
		}
		return parallelArrayWithEval(evaluator, arr)
	}

	// Case 2: Single object argument parallel({a = fn1, b = fn2})
	if obj, ok := args["0"].(map[string]any); ok && len(args) == 1 {
		return parallelObjectWithEval(evaluator, obj)
	}

	// Case 3: Varargs parallel(fn1, fn2, fn3)
	varargs := make([]any, 0)
	for i := 0; ; i++ {
		key := fmt.Sprintf("%d", i)
		if val, ok := args[key]; ok {
			varargs = append(varargs, val)
		} else {
			break
		}
	}

	if len(varargs) > 0 {
		return parallelArrayWithEval(evaluator, varargs)
	}

	return nil, fmt.Errorf("parallel() requires an array, object, or functions as arguments")
}

// parallelArrayWithEval executes an array of functions in parallel with isolated evaluators
func parallelArrayWithEval(evaluator *Evaluator, functions []any) (any, error) {
	results := make([]any, len(functions))

	var wg sync.WaitGroup
	for i, fnArg := range functions {
		wg.Add(1)
		go func(index int, fn any) {
			defer wg.Done()

			// Create a child evaluator for this block with parent scope access
			childEval := NewEvaluator()
			parentEnv := evaluator.GetEnv()
			childEnv := NewChildEnvironment(parentEnv)
			childEnv.SetParallelContext(true)
			childEval.SetEnvironment(childEnv)
			childEval.SetParallelContext(true) // Block parent scope writes

			// Call the function in the child evaluator
			fnVal := InterfaceToValue(fn)
			result, err := childEval.CallFunction(fnVal, make(map[string]Value))

			if err != nil {
				// Error handling: Option B - errors become nil
				results[index] = nil
			} else {
				results[index] = ValueToInterface(result)
			}
		}(i, fnArg)
	}
	wg.Wait()

	return results, nil
}

// parallelObjectWithEval executes an object of functions in parallel with isolated evaluators
func parallelObjectWithEval(evaluator *Evaluator, functions map[string]any) (any, error) {
	results := make(map[string]any)
	var mu sync.Mutex

	var wg sync.WaitGroup
	for key, fnArg := range functions {
		wg.Add(1)
		go func(k string, fn any) {
			defer wg.Done()

			// Create a child evaluator for this block with parent scope access
			childEval := NewEvaluator()
			parentEnv := evaluator.GetEnv()
			childEnv := NewChildEnvironment(parentEnv)
			childEnv.SetParallelContext(true)
			childEval.SetEnvironment(childEnv)
			childEval.SetParallelContext(true) // Block parent scope writes

			// Call the function in the child evaluator
			fnVal := InterfaceToValue(fn)
			result, err := childEval.CallFunction(fnVal, make(map[string]Value))

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				// Error handling: Option B - errors become nil
				results[k] = nil
			} else {
				results[k] = ValueToInterface(result)
			}
		}(key, fnArg)
	}
	wg.Wait()

	return results, nil
}

