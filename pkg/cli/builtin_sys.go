package cli

import (
	"fmt"

	"github.com/duso-org/duso/pkg/runtime"
	"github.com/duso-org/duso/pkg/script"
)

// GetSysFlag retrieves a flag from the sys datastore with type preservation and default fallback.
// Usage: GetSysFlag("-no-color", false), GetSysFlag("-v", false), etc.
// Returns the value from datastore, or defaultVal if not found or type mismatch.
func GetSysFlag[T any](key string, defaultVal T) T {
	ds := runtime.GetDatastore("sys", nil)
	val, _ := ds.Get(key)
	if typed, ok := val.(T); ok && val != nil {
		return typed
	}
	return defaultVal
}

// builtinSys retrieves values from the sys datastore
// Usage: sys("-debug"), sys("-no-color"), sys("any-key"), etc.
// Returns the value, or nil if the key was not found
func builtinSys(evaluator *script.Evaluator, args map[string]any) (any, error) {
	// Get the key from positional or named argument
	var key string

	if k, ok := args["0"].(string); ok {
		key = k
	} else if k, ok := args["key"]; ok {
		key = fmt.Sprintf("%v", k)
	} else {
		return nil, fmt.Errorf("sys() requires a string key")
	}

	// Get from sys datastore
	sysDs := runtime.GetDatastore("sys", nil)
	val, _ := sysDs.Get(key)

	// Return the value (or nil if not found)
	return val, nil
}
