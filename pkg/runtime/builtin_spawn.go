package runtime

import (
	"context"
	"fmt"
	"os"
	"time"

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
		var parentFrame *script.InvocationFrame
		if ctx, ok := script.GetRequestContext(gid); ok {
			parentFrame = ctx.Frame
		}

		// Resolve relative paths relative to the calling script's directory
		resolvedPath := scriptPath
		if parentFrame != nil && parentFrame.Filename != "" {
			resolvedPath = script.ResolveScriptPath(scriptPath, parentFrame.Filename)
		}

		// Read and validate script file BEFORE spawning (to catch errors early)
		// Use host-provided script loader capability
		fileBytes, err := interp.ScriptLoader(resolvedPath)
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

		// Get unique process ID and increment spawn counter
		pid := IncrementSpawnProcs()

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
			// The getter returns the data passed to spawn() by the caller
			SetContextGetter(spawnedGid, func() any {
				ctx, ok := script.GetRequestContext(spawnedGid)
				if !ok {
					return nil
				}
				return ctx.Data
			})
			defer ClearContextGetter(spawnedGid)

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

		return float64(pid), nil
	}
}

// NewRunFunction creates the run(script, context) builtin.
//
// run() executes a script synchronously in a spawned goroutine and blocks until
// the script calls exit() or completes. Returns the value passed to exit().
//
// Example:
//
//	result = run("worker.du", {data = [1, 2, 3]})
//	print("Result: " + format_json(result))
func NewRunFunction(interp *script.Interpreter) func(*script.Evaluator, map[string]any) (any, error) {
	return func(evaluator *script.Evaluator, args map[string]any) (any, error) {
		// Get script path (positional "0" or named "script")
		var scriptPath string
		if sp, ok := args["script"]; ok {
			if spStr, ok := sp.(string); ok {
				scriptPath = spStr
			} else {
				return nil, fmt.Errorf("run() script must be a string")
			}
		} else if sp, ok := args["0"]; ok {
			if spStr, ok := sp.(string); ok {
				scriptPath = spStr
			} else {
				return nil, fmt.Errorf("run() script path must be a string")
			}
		} else {
			return nil, fmt.Errorf("run() requires script path argument")
		}

		// Get context data (positional "1" or named "context", optional) - can be any Duso value
		var contextData any
		if cd, ok := args["context"]; ok {
			contextData = cd
		} else if cd, ok := args["1"]; ok {
			contextData = cd
		}

		// Get timeout in seconds (positional "2" or named "timeout", optional)
		var timeoutSecs float64
		if tm, ok := args["timeout"]; ok {
			if tmNum, ok := tm.(float64); ok {
				timeoutSecs = tmNum
			}
		} else if tm, ok := args["2"]; ok {
			if tmNum, ok := tm.(float64); ok {
				timeoutSecs = tmNum
			}
		}

		// Get current invocation frame (if in context)
		gid := GetGoroutineID()
		var parentFrame *script.InvocationFrame
		if ctx, ok := GetRequestContext(gid); ok {
			parentFrame = ctx.Frame
		}

		// Create timeout context if specified
		var timeoutCtx context.Context
		var cancel context.CancelFunc
		if timeoutSecs > 0 {
			timeoutCtx, cancel = context.WithTimeout(context.Background(), time.Duration(timeoutSecs)*time.Second)
			defer cancel()
		} else {
			timeoutCtx, cancel = context.WithCancel(context.Background())
			defer cancel()
		}

		// Increment run counter
		IncrementRunProcs()

		// Resolve relative paths relative to the calling script's directory
		resolvedPath := scriptPath
		if parentFrame != nil && parentFrame.Filename != "" {
			resolvedPath = script.ResolveScriptPath(scriptPath, parentFrame.Filename)
		}

		// Read script file - use host-provided script loader capability
		fileBytes, err := interp.ScriptLoader(resolvedPath)
		if err != nil {
			return nil, fmt.Errorf("run: failed to read %s: %w", scriptPath, err)
		}

		// Tokenize and parse
		lexer := script.NewLexer(string(fileBytes))
		tokens := lexer.Tokenize()
		parser := script.NewParserWithFile(tokens, scriptPath)
		program, err := parser.Parse()
		if err != nil {
			return nil, fmt.Errorf("run: failed to parse %s: %w", scriptPath, err)
		}

		// Create invocation frame for spawned script
		frame := &script.InvocationFrame{
			Filename: scriptPath,
			Line:     1,
			Col:      1,
			Reason:   "run",
			Details:  map[string]any{},
			Parent:   parentFrame,
		}

		// Create spawned context
		spawnedCtx := &script.RequestContext{
			Frame: frame,
		}

		// Execute script in goroutine and collect results
		resultChan := make(chan *script.ScriptExecutionResult, 1)
		done := make(chan bool, 1)

		go func() {
			// Register spawned context in THIS goroutine
			spawnedGid := script.GetGoroutineID()
			// Deep copy context data to isolate from parent scope
			contextDataCopy := script.DeepCopyAny(contextData)
			script.SetRequestContextWithData(spawnedGid, spawnedCtx, contextDataCopy)
			defer script.ClearRequestContext(spawnedGid)

			// Set up context getter for context() builtin
			// The getter returns the data passed to run() by the caller
			SetContextGetter(spawnedGid, func() any {
				ctx, ok := script.GetRequestContext(spawnedGid)
				if !ok {
					return nil
				}
				return ctx.Data
			})
			defer ClearContextGetter(spawnedGid)

			// Execute script (synchronously within the goroutine)
			result := script.ExecuteScript(
				program,
				interp,
				frame,
				spawnedCtx,
				timeoutCtx,
			)
			resultChan <- result
			done <- true
		}()

		// Wait for execution to complete
		<-done

		// Get the result
		result := <-resultChan
		if result != nil {
			// Return error if any, otherwise return the value
			if result.Error != nil {
				return nil, fmt.Errorf("run: error executing %s: %w", scriptPath, result.Error)
			}
			return result.Value, nil
		}
		return nil, nil
	}
}
