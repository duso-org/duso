package runtime

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// builtinHashPassword hashes a password using bcrypt
// Usage: hash_password(password [, cost])
func builtinHashPassword(evaluator *Evaluator, args map[string]any) (any, error) {
	// Get password - support both positional (0) and named (password)
	password := ""
	if p, ok := args["password"]; ok {
		if ps, ok := p.(string); ok {
			password = ps
		}
	} else if p, ok := args["0"]; ok {
		if ps, ok := p.(string); ok {
			password = ps
		}
	}

	if password == "" {
		return nil, fmt.Errorf("hash_password() requires a password string argument")
	}

	// Get optional cost parameter (default: 10, range: 4-31)
	cost := 10
	if costArg, ok := args["cost"]; ok {
		switch c := costArg.(type) {
		case float64:
			cost = int(c)
		}
	} else if costArg, ok := args["1"]; ok {
		switch c := costArg.(type) {
		case float64:
			cost = int(c)
		}
	}

	// Clamp cost to valid bcrypt range
	if cost < 4 {
		cost = 4
	}
	if cost > 31 {
		cost = 31
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return nil, fmt.Errorf("hash_password() failed: %v", err)
	}

	return string(hash), nil
}

// builtinVerifyPassword verifies a password against its hash
// Usage: verify_password(password, hash)
// Returns: true if password matches, false otherwise (never throws on mismatch)
func builtinVerifyPassword(evaluator *Evaluator, args map[string]any) (any, error) {
	// Get password - support both positional (0) and named (password)
	password := ""
	if p, ok := args["password"]; ok {
		if ps, ok := p.(string); ok {
			password = ps
		}
	} else if p, ok := args["0"]; ok {
		if ps, ok := p.(string); ok {
			password = ps
		}
	}

	if password == "" {
		return nil, fmt.Errorf("verify_password() requires password string argument")
	}

	// Get hash - support both positional (1) and named (hash)
	hash := ""
	if h, ok := args["hash"]; ok {
		if hs, ok := h.(string); ok {
			hash = hs
		}
	} else if h, ok := args["1"]; ok {
		if hs, ok := h.(string); ok {
			hash = hs
		}
	}

	if hash == "" {
		return nil, fmt.Errorf("verify_password() requires hash string argument")
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		// Password does not match - return false, don't throw
		return false, nil
	}

	return true, nil
}
