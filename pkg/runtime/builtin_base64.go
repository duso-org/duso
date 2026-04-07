package runtime

import (
	"encoding/base64"
	"fmt"

	"github.com/duso-org/duso/pkg/script"
)

// builtinEncodeBase64 encodes a string or binary to base64
func builtinEncodeBase64(evaluator *Evaluator, args map[string]any) (any, error) {
	if _, ok := args["0"]; !ok {
		return nil, fmt.Errorf("encode_base64() requires a string or binary argument")
	}

	var data []byte

	// Try as binary first
	if val, ok := args["0"].(script.Value); ok && val.IsBinary() {
		binVal := val.AsBinary()
		if binVal != nil && binVal.Data != nil {
			data = *binVal.Data
		}
	} else if val, ok := args["0"].(*script.ValueRef); ok && val.Val.IsBinary() {
		// Handle ValueRef wrapper
		binVal := val.Val.AsBinary()
		if binVal != nil && binVal.Data != nil {
			data = *binVal.Data
		}
	} else if str, ok := args["0"].(string); ok {
		// Handle string
		data = []byte(str)
	} else {
		// Fallback to stringify
		data = []byte(fmt.Sprintf("%v", args["0"]))
	}

	encoded := base64.StdEncoding.EncodeToString(data)
	return encoded, nil
}

// builtinDecodeBase64 decodes a base64 string to string or binary
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

	// Check for optional type parameter (named or positional)
	returnType := "string" // default
	if val, ok := args["type"]; ok {
		if typeStr, ok := val.(string); ok {
			returnType = typeStr
		}
	} else if val, ok := args["1"]; ok {
		if typeStr, ok := val.(string); ok {
			returnType = typeStr
		}
	}

	// Return based on specified type
	switch returnType {
	case "binary":
		return script.NewBinary(decoded), nil
	case "string":
		return string(decoded), nil
	default:
		return nil, fmt.Errorf("decode_base64() type parameter must be 'string' or 'binary', got '%s'", returnType)
	}
}
