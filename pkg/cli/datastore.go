package cli

import (
	"fmt"
	"time"

	"github.com/duso-org/duso/pkg/script"
)

// NewDatastoreFunction creates the datastore(namespace, config) builtin.
//
// datastore() returns a namespaced thread-safe key/value store with methods:
//   - .set(key, value) - Store a value
//   - .get(key) - Retrieve a value
//   - .increment(key, delta) - Atomically increment a number
//   - .append(key, item) - Atomically append to an array
//   - .wait(key [, expectedValue]) - Block until key changes or equals value
//   - .wait_for(key, predicate) - Block until predicate returns true
//   - .delete(key) - Remove a key
//   - .clear() - Remove all keys
//   - .save() - Explicitly save to disk (if configured)
//   - .load() - Explicitly load from disk (if configured)
//
// Configuration options:
//   - persist (string) - Path to JSON file for persistence
//   - persist_interval (number) - Auto-save interval in seconds
//
// Multiple scripts can share the same namespace to coordinate work.
//
// Example:
//
//	store = datastore("myapp", {persist = "data.json", persist_interval = 60})
//	store.set("status", "running")
//	store.increment("counter", 1)
//	store.append("items", {id = 1})
//	store.wait("counter", 10)  // Block until counter reaches 10
func NewDatastoreFunction() func(map[string]any) (any, error) {
	return func(args map[string]any) (any, error) {
		// Get namespace from first positional or named argument
		var namespace string

		if ns, ok := args["0"]; ok {
			// Positional argument
			namespace = fmt.Sprintf("%v", ns)
		} else if ns, ok := args["namespace"]; ok {
			// Named argument
			namespace = fmt.Sprintf("%v", ns)
		} else {
			return nil, fmt.Errorf("datastore() requires a namespace argument")
		}

		// Get config from second positional or named argument (optional)
		var config map[string]any

		if cfg, ok := args["1"]; ok {
			// Positional argument
			if cfgMap, ok := cfg.(map[string]any); ok {
				config = cfgMap
			}
		} else if cfg, ok := args["config"]; ok {
			// Named argument
			if cfgMap, ok := cfg.(map[string]any); ok {
				config = cfgMap
			}
		}

		// sys datastore is read-only and rejects any config
		if namespace == "sys" {
			if len(config) > 0 {
				return nil, fmt.Errorf("datastore(\"sys\") does not accept configuration options")
			}
		}

		// Get or create the datastore
		store := script.GetDatastore(namespace, config)

		// Create set(key, value) method
		setFn := script.NewGoFunction(func(setArgs map[string]any) (any, error) {
			if namespace == "sys" {
				return nil, fmt.Errorf("datastore(\"sys\") is read-only")
			}
			key, ok := setArgs["0"].(string)
			if !ok {
				return nil, fmt.Errorf("set() requires key (string) and value arguments")
			}
			value, ok := setArgs["1"]
			if !ok {
				return nil, fmt.Errorf("set() requires key and value arguments")
			}
			return nil, store.Set(key, value)
		})

		// Create get(key) method
		getFn := script.NewGoFunction(func(getArgs map[string]any) (any, error) {
			key, ok := getArgs["0"].(string)
			if !ok {
				return nil, fmt.Errorf("get() requires a key (string) argument")
			}
			return store.Get(key)
		})

		// Create increment(key, delta) method
		incrementFn := script.NewGoFunction(func(incArgs map[string]any) (any, error) {
			if namespace == "sys" {
				return nil, fmt.Errorf("datastore(\"sys\") is read-only")
			}
			key, ok := incArgs["0"].(string)
			if !ok {
				return nil, fmt.Errorf("increment() requires key (string) and delta arguments")
			}
			delta, ok := incArgs["1"].(float64)
			if !ok {
				return nil, fmt.Errorf("increment() requires a numeric delta argument")
			}
			return store.Increment(key, delta)
		})

		// Create append(key, item) method
		appendFn := script.NewGoFunction(func(appArgs map[string]any) (any, error) {
			if namespace == "sys" {
				return nil, fmt.Errorf("datastore(\"sys\") is read-only")
			}
			key, ok := appArgs["0"].(string)
			if !ok {
				return nil, fmt.Errorf("append() requires a key (string) argument")
			}
			item, ok := appArgs["1"]
			if !ok {
				return nil, fmt.Errorf("append() requires an item argument")
			}
			return store.Append(key, item)
		})

		// Create wait(key [, expectedValue]) method
		waitFn := script.NewGoFunction(func(waitArgs map[string]any) (any, error) {
			key, ok := waitArgs["0"].(string)
			if !ok {
				return nil, fmt.Errorf("wait() requires a key (string) argument")
			}

			// Check if expectedValue provided
			expectedValue, hasExpectedValue := waitArgs["1"]

			// Check for timeout (optional)
			timeout := time.Duration(0)
			if timeoutArg, ok := waitArgs["2"]; ok {
				if timeoutSecs, ok := timeoutArg.(float64); ok {
					timeout = time.Duration(timeoutSecs) * time.Second
				}
			} else if timeoutArg, ok := waitArgs["timeout"]; ok {
				if timeoutSecs, ok := timeoutArg.(float64); ok {
					timeout = time.Duration(timeoutSecs) * time.Second
				}
			}

			value, err := store.Wait(key, expectedValue, hasExpectedValue, timeout)
			return value, err
		})

		// Create wait_for(key, predicate [, timeout]) method
		waitForFn := script.NewGoFunction(func(wfArgs map[string]any) (any, error) {
			key, ok := wfArgs["0"].(string)
			if !ok {
				return nil, fmt.Errorf("wait_for() requires a key (string) argument")
			}

			predicateArg, ok := wfArgs["1"]
			if !ok {
				return nil, fmt.Errorf("wait_for() requires a predicate function argument")
			}

			// Extract GoFunction from the argument
			// It might be: GoFunction directly, or wrapped in ValueRef or Value
			var predicateFn script.GoFunction

			if goFn, ok := predicateArg.(script.GoFunction); ok {
				// Direct GoFunction
				predicateFn = goFn
			} else if vr, ok := predicateArg.(*script.ValueRef); ok {
				// Wrapped in ValueRef - extract the function
				if vr.Val.IsFunction() {
					if goFn, ok := vr.Val.Data.(script.GoFunction); ok {
						predicateFn = goFn
					} else {
						return nil, fmt.Errorf("wait_for() predicate must be a Go function (script functions not yet supported)")
					}
				} else {
					return nil, fmt.Errorf("wait_for() predicate must be a function")
				}
			} else {
				return nil, fmt.Errorf("wait_for() predicate must be a function")
			}

			// Check for timeout (optional)
			timeout := time.Duration(0)
			if timeoutArg, ok := wfArgs["2"]; ok {
				if timeoutSecs, ok := timeoutArg.(float64); ok {
					timeout = time.Duration(timeoutSecs) * time.Second
				}
			} else if timeoutArg, ok := wfArgs["timeout"]; ok {
				if timeoutSecs, ok := timeoutArg.(float64); ok {
					timeout = time.Duration(timeoutSecs) * time.Second
				}
			}

			value, err := store.WaitFor(key, predicateFn, timeout)
		return value, err
		})

		// Create delete(key) method
		deleteFn := script.NewGoFunction(func(delArgs map[string]any) (any, error) {
			if namespace == "sys" {
				return nil, fmt.Errorf("datastore(\"sys\") is read-only")
			}
			key, ok := delArgs["0"].(string)
			if !ok {
				return nil, fmt.Errorf("delete() requires a key (string) argument")
			}
			return nil, store.Delete(key)
		})

		// Create clear() method
		clearFn := script.NewGoFunction(func(clearArgs map[string]any) (any, error) {
			if namespace == "sys" {
				return nil, fmt.Errorf("datastore(\"sys\") is read-only")
			}
			return nil, store.Clear()
		})

		// Create save() method
		saveFn := script.NewGoFunction(func(saveArgs map[string]any) (any, error) {
			if namespace == "sys" {
				return nil, fmt.Errorf("datastore(\"sys\") is read-only")
			}
			return nil, store.Save()
		})

		// Create load() method
		loadFn := script.NewGoFunction(func(loadArgs map[string]any) (any, error) {
			if namespace == "sys" {
				return nil, fmt.Errorf("datastore(\"sys\") is read-only")
			}
			return nil, store.Load()
		})

		// Return store object with methods
		return map[string]any{
			"set":       setFn,
			"get":       getFn,
			"increment": incrementFn,
			"append":    appendFn,
			"wait":      waitFn,
			"wait_for":  waitForFn,
			"delete":    deleteFn,
			"clear":     clearFn,
			"save":      saveFn,
			"load":      loadFn,
		}, nil
	}
}
