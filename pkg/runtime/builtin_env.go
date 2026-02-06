package runtime

import (
	"fmt"
	"os"

	"github.com/duso-org/duso/pkg/script"
)

// NewEnvFunction creates an env(varname) builtin that reads environment variables.
//
// env() reads the value of an environment variable from the OS (or host-provided source).
// Returns the value as a string, or empty string if the variable is not set.
//
// Hosts can provide their own environment via the EnvReader capability for custom behavior
// (e.g., sandboxed environment, filtered variables, etc.)
//
// Example:
//
//	key = env("ANTHROPIC_API_KEY")
//	debug = env("DEBUG_MODE")
func NewEnvFunction(interp *script.Interpreter) func(*script.Evaluator, map[string]any) (any, error) {
	return func(evaluator *script.Evaluator, args map[string]any) (any, error) {
		varname, ok := args["0"].(string)
		if !ok {
			// Check for named argument "varname"
			if v, ok := args["varname"]; ok {
				varname = fmt.Sprintf("%v", v)
			} else {
				return nil, fmt.Errorf("env() requires a variable name argument")
			}
		}

		// Use EnvReader capability if available, otherwise fall back to os.Getenv
		if interp.EnvReader != nil {
			return interp.EnvReader(varname), nil
		}
		return os.Getenv(varname), nil
	}
}
