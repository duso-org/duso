package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/duso-org/duso/pkg/runtime"
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
		gid := runtime.GetGoroutineID()
		var parentFrame *script.InvocationFrame
		if ctx, ok := runtime.GetRequestContext(gid); ok {
			parentFrame = ctx.Frame
		}

		// Increment spawn counter
		runtime.IncrementSpawnProcs()

		// Spawn goroutine (fire-and-forget)
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
			// Deep copy context data to isolate from parent scope
			contextDataCopy := script.DeepCopyAny(contextData)
			script.SetRequestContextWithData(spawnedGid, spawnedCtx, contextDataCopy)
			defer script.ClearRequestContext(spawnedGid)

			// Set up context getter for context() builtin
			// The getter returns the RequestContext stored in script's goroutine-local storage
			runtime.SetContextGetter(spawnedGid, func() any {
				ctx, ok := script.GetRequestContext(spawnedGid)
				if !ok {
					return nil
				}
				return ctx
			})
			defer runtime.ClearContextGetter(spawnedGid)

			// Read script file (try local first, then embedded)
			fileBytes, err := ReadScriptWithFallback(scriptPath, interp.GetScriptDir())
			if err != nil {
				fmt.Fprintf(os.Stderr, "spawn: failed to read %s: %v\n", scriptPath, err)
				return
			}

			// Tokenize and parse
			lexer := script.NewLexer(string(fileBytes))
			tokens := lexer.Tokenize()
			parser := script.NewParserWithFile(tokens, scriptPath)
			program, err := parser.Parse()
			if err != nil {
				fmt.Fprintf(os.Stderr, "spawn: failed to parse %s: %v\n", scriptPath, err)
				return
			}

			// Create a fresh evaluator for the spawned script
			spawnedEvaluator := script.NewEvaluator()

			// Copy registered functions and settings from parent evaluator
			parentEval := interp.GetEvaluator()
			for name, fn := range parentEval.GetGoFunctions() {
				spawnedEvaluator.RegisterFunction(name, fn)
			}
			// Copy debug and stdin settings from parent
			spawnedEvaluator.DebugMode = parentEval.DebugMode
			spawnedEvaluator.NoStdin = parentEval.NoStdin

			// Execute script (fire-and-forget, no timeout)
			result := script.ExecuteScript(
				program,
				spawnedEvaluator,
				interp,
				frame,
				spawnedCtx,
				context.Background(),
			)

			// Log any errors to stderr
			if result != nil && result.Error != nil {
				fmt.Fprintf(os.Stderr, "spawn: error in %s: %v\n", scriptPath, result.Error)
			}
		}()

		return nil, nil
	}
}
