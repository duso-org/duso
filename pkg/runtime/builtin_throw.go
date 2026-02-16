package runtime

import (
	"fmt"
	"strings"
)

// builtinThrow throws an error with message and call stack
func builtinThrow(evaluator *Evaluator, args map[string]any) (any, error) {
	// Accept any value type, no deep copy at throw time
	// Will be deep copied only if it crosses process boundaries (run())
	var value any = "unknown error"

	if msg, ok := args["0"]; ok {
		value = msg
	} else if msg, ok := args["message"]; ok {
		value = msg
	}

	// Create DusoError with call stack, storing the original value
	err := &DusoError{
		Message: value,
	}

	if evaluator != nil {
		ctx := evaluator.GetContext()
		if ctx != nil {
			err.FilePath = ctx.FilePath
			err.CallStack = ctx.CallStack
		}
	}

	return nil, err
}

// builtinBreakpoint signals a debug breakpoint with call stack captured
// Optional arguments are passed as a debug message (not printed directly)
func builtinBreakpoint(evaluator *Evaluator, args map[string]any) (any, error) {
	// Only trigger breakpoint if debug mode is enabled
	if evaluator == nil || !evaluator.DebugMode {
		return nil, nil
	}

	// If arguments provided, format them as a debug message
	var message string
	if len(args) > 0 {
		message = "BREAKPOINT: " + formatArgsThrow(args)
	}

	ctx := evaluator.GetContext()
	if ctx == nil {
		return nil, fmt.Errorf("breakpoint() requires execution context")
	}

	// Capture call stack and current environment for debug display
	// Clone the call stack so it can't be modified
	callStack := make([]CallFrame, len(ctx.CallStack))
	copy(callStack, ctx.CallStack)

	env := evaluator.GetEnv()
	err := &BreakpointError{
		FilePath:  ctx.FilePath,
		CallStack: callStack,
		Env:       env, // Capture current environment for scope access
		Message:   message,       // Pass message to debug handler
	}
	return nil, err
}

// builtinWatch evaluates expressions and breaks if values change
// Each argument is a string expression to watch
func builtinWatch(evaluator *Evaluator, args map[string]any) (any, error) {
	if evaluator == nil {
		return nil, fmt.Errorf("watch() requires evaluator context")
	}

	var triggered []string // Collect which watches triggered
	watchCache := evaluator.GetWatchCache()

	// Process each watch expression
	for i := 0; ; i++ {
		key := fmt.Sprintf("%d", i)
		exprStr, ok := args[key]
		if !ok {
			break
		}

		// Expression must be a string
		expr, ok := exprStr.(string)
		if !ok {
			return nil, fmt.Errorf("watch() requires string expressions, got %v", exprStr)
		}

		// Parse and evaluate the expression using public API
		node, err := evaluator.ParseExpression(expr)
		if err != nil {
			return nil, fmt.Errorf("watch() parse error in '{{%s}}': %v", expr, err)
		}

		val, err := evaluator.Eval(node)
		if err != nil {
			return nil, fmt.Errorf("watch() evaluation error in '{{%s}}': %v", expr, err)
		}

		// Check if value changed from cached
		cachedVal, exists := watchCache[expr]
		if !exists || !valuesEqualThrow(val, cachedVal) {
			// Value changed or first time seeing it
			watchCache[expr] = val
			triggered = append(triggered, fmt.Sprintf("WATCH: %s = %v", expr, val.String()))
		}
	}

	// If any watches triggered and debug mode is enabled, create breakpoint with messages
	if len(triggered) > 0 && evaluator.DebugMode {
		// Combine all triggered messages
		message := strings.Join(triggered, "\n")

		ctx := evaluator.GetContext()
		if ctx == nil {
			return nil, fmt.Errorf("watch() requires execution context")
		}

		// Trigger breakpoint with call stack
		callStack := make([]CallFrame, len(ctx.CallStack))
		copy(callStack, ctx.CallStack)

		env := evaluator.GetEnv()
		err := &BreakpointError{
			FilePath:  ctx.FilePath,
			CallStack: callStack,
			Env:       env,
			Message:   message, // Pass all watch messages to debug handler
		}
		return nil, err
	}

	return nil, nil
}

// formatArgsThrow converts arguments to space-separated string (like print would output)
func formatArgsThrow(args map[string]any) string {
	var parts []string
	for i := 0; ; i++ {
		key := fmt.Sprintf("%d", i)
		if val, ok := args[key]; ok {
			parts = append(parts, fmt.Sprintf("%v", val))
		} else {
			break
		}
	}
	return strings.Join(parts, " ")
}

// valuesEqualThrow checks if two values are equal (for watch caching)
func valuesEqualThrow(v1, v2 Value) bool {
	if v1.Type != v2.Type {
		return false
	}

	switch v1.Type {
	case VAL_NIL:
		return true
	case VAL_NUMBER:
		return v1.AsNumber() == v2.AsNumber()
	case VAL_STRING:
		return v1.AsString() == v2.AsString()
	case VAL_BOOL:
		return v1.AsBool() == v2.AsBool()
	case VAL_ARRAY:
		arr1 := v1.AsArray()
		arr2 := v2.AsArray()
		if len(arr1) != len(arr2) {
			return false
		}
		for i := range arr1 {
			if !valuesEqualThrow(arr1[i], arr2[i]) {
				return false
			}
		}
		return true
	case VAL_OBJECT:
		obj1 := v1.AsObject()
		obj2 := v2.AsObject()
		if len(obj1) != len(obj2) {
			return false
		}
		for k, v := range obj1 {
			v2Val, ok := obj2[k]
			if !ok || !valuesEqualThrow(v, v2Val) {
				return false
			}
		}
		return true
	case VAL_FUNCTION:
		// Functions are compared by reference
		return v1.Data == v2.Data
	default:
		return false
	}
}
