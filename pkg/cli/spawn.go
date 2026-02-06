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
func NewSpawnFunction(interp *script.Interpreter) func(*script.Evaluator, map[string]any) (any, error) {
	return func(evaluator *script.Evaluator, args map[string]any) (any, error) {
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
		fmt.Fprintf(os.Stderr, "[SPAWN-CALLER] gid=%d\n", gid)
		var parentFrame *script.InvocationFrame
		if ctx, ok := script.GetRequestContext(gid); ok {
			parentFrame = ctx.Frame
			fmt.Fprintf(os.Stderr, "[SPAWN-CALLER] found context for gid=%d: %+v\n", gid, ctx)
		} else {
			fmt.Fprintf(os.Stderr, "[SPAWN-CALLER] NO context found for gid=%d\n", gid)
		}

		// Debug: show what directory we think we're in
		scriptDirBeingUsed := interp.GetScriptDir()
		parentInfo := "none"
		if parentFrame != nil && parentFrame.Filename != "" {
			parentInfo = parentFrame.Filename
		}
		fmt.Fprintf(os.Stderr, "[SPAWN] gid=%d scriptPath=%q parent=%q scriptDir=%q\n", gid, scriptPath, parentInfo, scriptDirBeingUsed)

		// Read and validate script file BEFORE spawning (to catch errors early)
		fileBytes, err := ReadScriptWithFallback(scriptPath, interp.GetScriptDir())
		if err != nil {
			return nil, fmt.Errorf("spawn: failed to read %s: %w", scriptPath, err)
		}

		// Tokenize and parse BEFORE spawning (to catch parse errors early)
		lexer := script.NewLexer(string(fileBytes))
		tokens := lexer.Tokenize()
		parser := script.NewParserWithFile(tokens, scriptPath)
		program, err := parser.Parse()
		if err != nil {
			return nil, fmt.Errorf("spawn: failed to parse %s: %w", scriptPath, err)
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
			fmt.Fprintf(os.Stderr, "[SPAWN-GOROUTINE] storing context for gid=%d frame=%q\n", spawnedGid, frame.Filename)
			// Deep copy context data to isolate from parent scope
			contextDataCopy := script.DeepCopyAny(contextData)
			script.SetRequestContextWithData(spawnedGid, spawnedCtx, contextDataCopy)
			defer script.ClearRequestContext(spawnedGid)

			// Set up context getter for context() builtin
			// The getter returns the data passed to spawn() by the caller
			runtime.SetContextGetter(spawnedGid, func() any {
				ctx, ok := script.GetRequestContext(spawnedGid)
				if !ok {
					return nil
				}
				return ctx.Data
			})
			defer runtime.ClearContextGetter(spawnedGid)

			// Execute script (fire-and-forget, no timeout)
			result := script.ExecuteScript(
				program,
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
