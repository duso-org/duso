package runtime

import (
	"fmt"
	"os"

	"github.com/duso-org/duso/pkg/script"
)

// NewInputFunction creates an input(prompt) builtin that reads from user input.
//
// input() reads a line from user input with an optional prompt.
// Returns the line without the trailing newline, or empty string on EOF.
//
// Example:
//
//	name = input("Enter your name: ")
//	age = input("How old are you? ")
func NewInputFunction(interp *script.Interpreter) func(*script.Evaluator, map[string]any) (any, error) {
	return func(evaluator *script.Evaluator, args map[string]any) (any, error) {
		// Get optional prompt argument
		var prompt string
		if promptArg, ok := args["0"]; ok {
			prompt = fmt.Sprintf("%v", promptArg)
		}

		// Use InputReader capability if available
		if interp.InputReader != nil {
			line, err := interp.InputReader(prompt)
			return line, err
		}

		// Fallback: check if stdin is disabled
		if evaluator != nil && evaluator.NoStdin {
			fmt.Println("warning: stdin disabled, input() returned ''")
			return "", nil
		}

		// Fallback to direct stdin reading
		if prompt != "" {
			fmt.Fprint(os.Stdout, prompt)
		}

		var line string
		fmt.Scanln(&line)
		return line, nil
	}
}
