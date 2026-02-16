package runtime

import "fmt"

// builtinTemplate creates a reusable template function from a template string.
// template(template_string) returns a function that evaluates the template with provided named args.
func builtinTemplate(evaluator *Evaluator, args map[string]any) (any, error) {
	templateStr, ok := args["0"].(string)
	if !ok {
		return nil, fmt.Errorf("template() requires a string argument")
	}

	// Return a function that evaluates the template with provided args
	templateFn := func(templateEval *Evaluator, templateArgs map[string]any) (any, error) {
		// Convert map[string]any to map[string]Value for bindings
		bindings := make(map[string]Value)
		for key, val := range templateArgs {
			// Skip numeric positional keys (those are internal)
			if _, err := fmt.Sscanf(key, "%d", new(int)); err == nil {
				continue
			}

			// Convert Go value to Duso Value
			bindings[key] = InterfaceToValue(val)
		}

		// Use public API to evaluate template
		result, err := templateEval.EvaluateTemplate(templateStr, bindings)
		if err != nil {
			return nil, fmt.Errorf("template evaluation error: %w", err)
		}

		return result, nil
	}

	return NewGoFunction(templateFn), nil
}
