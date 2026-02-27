package runtime

import (
	"github.com/duso-org/duso/pkg/script"
)

// builtinParse parses source code string and returns a code value or error value (never throws).
// parse(source_string, metadata?)
//
// Example:
//
//	code_val = parse("return 1 + 1")
//	if is_error(code_val) then
//	  print("Parse failed: " + code_val.message)
//	else
//	  result = run(code_val)
//	  print(result)
//	end
func builtinParse(evaluator *script.Evaluator, args map[string]any) (any, error) {
	source, ok := args["0"].(string)
	if !ok {
		return script.NewErrorValue(script.NewString("parse() requires a string"), ""), nil
	}

	// Optional metadata object (positional arg 1 or named "metadata")
	// ValueToInterface converts objects to map[string]any, so we need to convert back
	var meta map[string]script.Value
	if m, ok := args["1"]; ok {
		if mv, ok := m.(map[string]any); ok {
			// Convert map[string]any back to map[string]script.Value
			meta = make(map[string]script.Value)
			for k, v := range mv {
				meta[k] = script.InterfaceToValue(v)
			}
		}
	} else if m, ok := args["metadata"]; ok {
		if mv, ok := m.(map[string]any); ok {
			// Convert map[string]any back to map[string]script.Value
			meta = make(map[string]script.Value)
			for k, v := range mv {
				meta[k] = script.InterfaceToValue(v)
			}
		}
	}

	// Parse — catch error, return as error value (never throw)
	lexer := script.NewLexer(source)
	tokens := lexer.Tokenize()
	parser := script.NewParser(tokens)  // no file path — dynamic code
	program, err := parser.Parse()
	if err != nil {
		errVal := script.NewErrorValue(script.NewString(err.Error()), err.Error())
		return errVal, nil  // return error value, don't throw
	}

	return script.NewCode(source, program, meta), nil
}
