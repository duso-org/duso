package runtime

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"strings"
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

	// Get data - support both positional (1) and named (data)
	data := ""
	if d, ok := args["data"]; ok {
		if ds, ok := d.(string); ok {
			data = ds
		}
	} else if d, ok := args["1"]; ok {
		if ds, ok := d.(string); ok {
			data = ds
		}
	}

	if data == "" {
		return nil, fmt.Errorf("hash() requires a data string argument")
	}

	// Normalize algorithm name (lowercase)
	algo = strings.ToLower(algo)

	// Compute hash based on algorithm
	var hashBytes []byte
	switch algo {
	case "sha256":
		h := sha256.Sum256([]byte(data))
		hashBytes = h[:]
	case "sha512":
		h := sha512.Sum512([]byte(data))
		hashBytes = h[:]
	case "sha1":
		h := sha1.Sum([]byte(data))
		hashBytes = h[:]
	case "md5":
		h := md5.Sum([]byte(data))
		hashBytes = h[:]
	default:
		return nil, fmt.Errorf("hash() unsupported algorithm: %s (supported: sha256, sha512, sha1, md5)", algo)
	}

	// Convert to hex string
	hexString := fmt.Sprintf("%x", hashBytes)
	return hexString, nil
}
