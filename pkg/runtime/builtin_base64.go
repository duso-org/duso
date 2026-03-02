package runtime

import (
	"encoding/base64"
	"fmt"
)

// builtinEncodeBase64 encodes a string to base64
func builtinEncodeBase64(evaluator *Evaluator, args map[string]any) (any, error) {
	if _, ok := args["0"]; !ok {
		return nil, fmt.Errorf("encode_base64() requires a string argument")
	}

	input, ok := args["0"].(string)
	if !ok {
		input = fmt.Sprintf("%v", args["0"])
	}

	encoded := base64.StdEncoding.EncodeToString([]byte(input))
	return encoded, nil
}

// builtinDecodeBase64 decodes a base64 string
func builtinDecodeBase64(evaluator *Evaluator, args map[string]any) (any, error) {
	if _, ok := args["0"]; !ok {
		return nil, fmt.Errorf("decode_base64() requires a string argument")
	}

	input, ok := args["0"].(string)
	if !ok {
		return nil, fmt.Errorf("decode_base64() requires a string argument")
	}

	decoded, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return nil, fmt.Errorf("decode_base64() failed to decode: %v", err)
	}

	return string(decoded), nil
}
