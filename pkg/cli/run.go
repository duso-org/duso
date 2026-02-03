package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/duso-org/duso/pkg/runtime"
	"github.com/duso-org/duso/pkg/script"
)

// NewRunFunction creates the run(script, context) builtin.
//
// run() executes a script synchronously in a spawned goroutine and blocks until
// the script calls exit() or completes. Returns the value passed to exit().
//
// Example:
//
//	result = run("worker.du", {data = [1, 2, 3]})
//	print("Result: " + format_json(result))
func NewRunFunction(interp *script.Interpreter) func(map[string]any) (any, error) {
	return func(args map[string]any) (any, error) {
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
		gid := runtime.GetGoroutineID()
		var parentFrame *script.InvocationFrame
		if ctx, ok := runtime.GetRequestContext(gid); ok {
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
		runtime.IncrementRunProcs()

		// Read script file (try local first, then embedded)
		fileBytes, err := ReadScriptWithFallback(scriptPath, interp.GetScriptDir())
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
			// The getter returns the RequestContext stored in script's goroutine-local storage
			runtime.SetContextGetter(spawnedGid, func() any {
				ctx, ok := script.GetRequestContext(spawnedGid)
				if !ok {
					return nil
				}
				return ctx
			})
			defer runtime.ClearContextGetter(spawnedGid)

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

			// Execute script (synchronously within the goroutine)
			result := script.ExecuteScript(
				program,
				spawnedEvaluator,
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
