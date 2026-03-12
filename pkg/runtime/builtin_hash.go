package runtime

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"strings"

	"github.com/duso-org/duso/pkg/script"
)

// builtinHash computes a hash of data using the specified algorithm
// Usage: hash(algo, data)
// Algorithms: "sha256" (default), "sha512", "sha1", "md5"
// Returns: hex-encoded hash string
func builtinHash(evaluator *Evaluator, args map[string]any) (any, error) {
	// Get algo - support both positional (0) and named (algo)
	algo := ""
	if a, ok := args["algo"]; ok {
		if as, ok := a.(string); ok {
			algo = as
		}
	} else if a, ok := args["0"]; ok {
		if as, ok := a.(string); ok {
			algo = as
		}
	}

	if algo == "" {
		return nil, fmt.Errorf("hash() requires an algorithm string argument")
	}

	// Get data - support both positional (1) and named (data), accept string or binary
	var dataBytes []byte
	var dataArg any

	if d, ok := args["data"]; ok {
		dataArg = d
	} else if d, ok := args["1"]; ok {
		dataArg = d
	}

	if dataArg == nil {
		return nil, fmt.Errorf("hash() requires a data argument (string or binary)")
	}

	// Handle binary data
	if val, ok := dataArg.(script.Value); ok && val.IsBinary() {
		binVal := val.AsBinary()
		if binVal != nil && binVal.Data != nil {
			dataBytes = *binVal.Data
		}
	} else if val, ok := dataArg.(*script.ValueRef); ok && val.Val.IsBinary() {
		// Handle ValueRef wrapper
		binVal := val.Val.AsBinary()
		if binVal != nil && binVal.Data != nil {
			dataBytes = *binVal.Data
		}
	} else if str, ok := dataArg.(string); ok {
		// Handle string
		dataBytes = []byte(str)
	} else {
		// Fallback to stringify
		dataBytes = []byte(fmt.Sprintf("%v", dataArg))
	}

	if len(dataBytes) == 0 {
		return nil, fmt.Errorf("hash() requires a non-empty data argument")
	}

	// Normalize algorithm name (lowercase)
	algo = strings.ToLower(algo)

	// Compute hash based on algorithm
	var hashBytes []byte
	switch algo {
	case "sha256":
		h := sha256.Sum256(dataBytes)
		hashBytes = h[:]
	case "sha512":
		h := sha512.Sum512(dataBytes)
		hashBytes = h[:]
	case "sha1":
		h := sha1.Sum(dataBytes)
		hashBytes = h[:]
	case "md5":
		h := md5.Sum(dataBytes)
		hashBytes = h[:]
	default:
		return nil, fmt.Errorf("hash() unsupported algorithm: %s (supported: sha256, sha512, sha1, md5)", algo)
	}

	// Convert to hex string
	hexString := fmt.Sprintf("%x", hashBytes)
	return hexString, nil
}
