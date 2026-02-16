package runtime

import (
	"fmt"
	"os"
)

// builtinEnv reads environment variables from the OS.
//
// env(varname) returns the value of an environment variable as a string,
// or empty string if the variable is not set.
//
// Example:
//
//	key = env("ANTHROPIC_API_KEY")
//	debug = env("DEBUG_MODE")
func builtinEnv(evaluator *Evaluator, args map[string]any) (any, error) {
	varname, ok := args["0"].(string)
	if !ok {
		// Check for named argument "varname"
		if v, ok := args["varname"]; ok {
			varname = fmt.Sprintf("%v", v)
		} else {
			return nil, fmt.Errorf("env() requires a variable name argument")
		}
	}

	return os.Getenv(varname), nil
}
