package runtime

import (
	"fmt"
	"os"
	"strings"

	"github.com/duso-org/duso/pkg/script"
)

// NewPrintFunction creates a print() builtin that uses OutputWriter capability.
//
// print() outputs values to the host's output stream (stdout by default in CLI).
// Multiple arguments are separated by spaces.
//
// Example:
//
//	print("Hello, World!")
//	print("Value:", 42, "Status:", true)
func NewPrintFunction(interp *script.Interpreter) func(*script.Evaluator, map[string]any) (any, error) {
	return func(evaluator *script.Evaluator, args map[string]any) (any, error) {
		var parts []string
		for i := 0; ; i++ {
			key := fmt.Sprintf("%d", i)
			if val, ok := args[key]; ok {
				parts = append(parts, fmt.Sprintf("%v", val))
			} else {
				break
			}
		}

		output := strings.Join(parts, " ")

		// Use OutputWriter capability if available, otherwise fall back to fmt.Println
		if interp.OutputWriter != nil {
			return nil, interp.OutputWriter(output)
		}
		fmt.Println(output)
		return nil, nil
	}
}

// NewErrorFunction creates an error() builtin for error output.
//
// error() outputs an error message to stderr via the OutputWriter capability.
// Used for printing error messages and warnings.
//
// Example:
//
//	error("File not found:", filename)
//	error("Warning: deprecated function")
func NewErrorFunction(interp *script.Interpreter) func(*script.Evaluator, map[string]any) (any, error) {
	return func(evaluator *script.Evaluator, args map[string]any) (any, error) {
		var parts []string
		for i := 0; ; i++ {
			key := fmt.Sprintf("%d", i)
			if val, ok := args[key]; ok {
				parts = append(parts, fmt.Sprintf("%v", val))
			} else {
				break
			}
		}

		output := strings.Join(parts, " ")

		// Use OutputWriter capability if available, otherwise fall back to stderr
		if interp.OutputWriter != nil {
			return nil, interp.OutputWriter(output)
		}
		fmt.Fprintln(os.Stderr, output)
		return nil, nil
	}
}

// NewDebugFunction creates a debug() builtin for debug output.
//
// debug() outputs debug messages to stdout when debug mode is enabled.
// Messages are prefixed with "[DEBUG]" for easy identification.
//
// Example:
//
//	debug("Processing item:", item)
//	debug("State:", state)
func NewDebugFunction(interp *script.Interpreter) func(*script.Evaluator, map[string]any) (any, error) {
	return func(evaluator *script.Evaluator, args map[string]any) (any, error) {
		var parts []string
		for i := 0; ; i++ {
			key := fmt.Sprintf("%d", i)
			if val, ok := args[key]; ok {
				parts = append(parts, fmt.Sprintf("%v", val))
			} else {
				break
			}
		}

		output := "[DEBUG] " + strings.Join(parts, " ")

		// Use OutputWriter capability if available, otherwise fall back to fmt.Println
		if interp.OutputWriter != nil {
			return nil, interp.OutputWriter(output)
		}
		fmt.Println(output)
		return nil, nil
	}
}
