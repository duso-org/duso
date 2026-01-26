package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

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

		// Get context data (positional "1" or named "context", optional)
		var contextData map[string]any
		if cd, ok := args["context"]; ok {
			if cdMap, ok := cd.(map[string]any); ok {
				contextData = cdMap
			}
		} else if cd, ok := args["1"]; ok {
			if cdMap, ok := cd.(map[string]any); ok {
				contextData = cdMap
			}
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
		gid := script.GetGoroutineID()
		var parentFrame *script.InvocationFrame
		if ctx, ok := script.GetRequestContext(gid); ok {
			parentFrame = ctx.Frame
		}

		// Create channel for result
		resultChan := make(chan any, 1)

		// Create timeout context if specified
		var ctx context.Context
		var cancel context.CancelFunc
		if timeoutSecs > 0 {
			ctx, cancel = context.WithTimeout(context.Background(), time.Duration(timeoutSecs)*time.Second)
			defer cancel()
		} else {
			ctx, cancel = context.WithCancel(context.Background())
			defer cancel()
		}

		// Spawn goroutine (synchronously wait for it)
		done := make(chan bool, 1)

		go func() {
			defer func() { done <- true }()

			// Create invocation frame for spawned script
			frame := &script.InvocationFrame{
				Filename: scriptPath,
				Line:     1,
				Col:      1,
				Reason:   "run",
				Details: map[string]any{
					"context": contextData != nil,
				},
				Parent: parentFrame,
			}

			// Create spawned context with result channel
			spawnedCtx := &script.RequestContext{
				Frame:    frame,
				ExitChan: resultChan,
			}

			// Register spawned context in goroutine-local storage
			spawnedGid := script.GetGoroutineID()
			script.SetRequestContextWithData(spawnedGid, spawnedCtx, contextData)
			defer script.ClearRequestContext(spawnedGid)

			// Read script file
			fileBytes, err := readFile(scriptPath)
			if err != nil {
				select {
				case resultChan <- fmt.Errorf("run: failed to read %s: %w", scriptPath, err):
				case <-ctx.Done():
					resultChan <- fmt.Errorf("run: timeout exceeded")
				}
				return
			}
			source := string(fileBytes)

			// Tokenize and parse
			lexer := script.NewLexer(source)
			tokens := lexer.Tokenize()
			parser := script.NewParser(tokens)
			program, err := parser.Parse()
			if err != nil {
				select {
				case resultChan <- fmt.Errorf("run: failed to parse %s: %w", scriptPath, err):
				case <-ctx.Done():
					resultChan <- fmt.Errorf("run: timeout exceeded")
				}
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

			// Check for timeout before processing result
			select {
			case <-ctx.Done():
				resultChan <- fmt.Errorf("run: timeout exceeded after %v seconds", timeoutSecs)
				return
			default:
			}

			if err != nil {
				// Check if exit() was called
				if exitErr, ok := err.(*script.ExitExecution); ok {
					// Script called exit() - send the value(s)
					if len(exitErr.Values) > 0 {
						resultChan <- exitErr.Values[0]
					} else {
						resultChan <- nil
					}
				} else {
					// Regular error
					resultChan <- fmt.Errorf("run: error executing %s: %w", scriptPath, err)
				}
			}
		}()

		// Wait for goroutine to finish
		<-done

		// Check for result (script called exit())
		select {
		case result := <-resultChan:
			return result, nil
		default:
			// Script completed without exit() - return nil
			return nil, nil
		}
	}
}
