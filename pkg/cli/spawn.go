package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/duso-org/duso/pkg/script"
)

// NewSpawnFunction creates the spawn(script, context) builtin.
//
// spawn() runs a script in a background goroutine with an optional context object.
// The spawned script receives the context via context() builtin.
// This is fire-and-forget: spawn() returns immediately without waiting.
//
// Example:
//
//	spawn("worker.du", {data = [1, 2, 3]})
//	print("worker running in background")
func NewSpawnFunction(interp *script.Interpreter) func(map[string]any) (any, error) {
	return func(args map[string]any) (any, error) {
		// Get script path
		var scriptPath string
		if sp, ok := args["0"]; ok {
			if spStr, ok := sp.(string); ok {
				scriptPath = spStr
			} else {
				return nil, fmt.Errorf("spawn() script path must be a string")
			}
		} else {
			return nil, fmt.Errorf("spawn() requires script path argument")
		}

		// Get context data (optional) - can be any Duso value
		var contextData any
		if cd, ok := args["1"]; ok {
			contextData = cd
		}

		// Get current invocation frame (if in context)
		gid := script.GetGoroutineID()
		var parentFrame *script.InvocationFrame
		if ctx, ok := script.GetRequestContext(gid); ok {
			parentFrame = ctx.Frame
		}

		// Spawn goroutine
		go func() {
			// Create invocation frame for spawned script
			frame := &script.InvocationFrame{
				Filename: scriptPath,
				Line:     1,
				Col:      1,
				Reason:   "spawn",
				Details:  map[string]any{},
				Parent:   parentFrame,
			}

			// Create spawned context
			spawnedCtx := &script.RequestContext{
				Frame: frame,
			}

			// Register spawned context in goroutine-local storage
			spawnedGid := script.GetGoroutineID()
			script.SetRequestContextWithData(spawnedGid, spawnedCtx, contextData)
			defer script.ClearRequestContext(spawnedGid)

			// Read script file
			fileBytes, err := readFile(scriptPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "spawn: failed to read %s: %v\n", scriptPath, err)
				return
			}
			source := string(fileBytes)

			// Tokenize and parse
			lexer := script.NewLexer(source)
			tokens := lexer.Tokenize()
			parser := script.NewParser(tokens)
			program, err := parser.Parse()
			if err != nil {
				fmt.Fprintf(os.Stderr, "spawn: failed to parse %s: %v\n", scriptPath, err)
				return
			}

			// Create fresh evaluator
			childEval := script.NewEvaluator(&strings.Builder{})

			// Copy registered functions from parent evaluator
			if interp != nil && interp.GetEvaluator() != nil {
				parentEval := interp.GetEvaluator()
				for name, fn := range parentEval.GetGoFunctions() {
					childEval.RegisterFunction(name, fn)
				}
			}

			// Execute spawned script
			_, err = childEval.Eval(program)
			if err != nil {
				// Check if exit() was called
				if _, ok := err.(*script.ExitExecution); ok {
					// Script called exit() - that's normal
				} else {
					// Regular error
					fmt.Fprintf(os.Stderr, "spawn: error executing %s: %v\n", scriptPath, err)
				}
			}
		}()

		return nil, nil
	}
}
