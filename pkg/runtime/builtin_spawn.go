package runtime

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/duso-org/duso/pkg/script"
)

var (
	spawnProcsCounter int64 // Counter for unique spawn() process IDs
	runProcsCounter   int64 // Counter for run() process calls

	// Track spawned goroutines by PID for kill() support
	spawnedProcs = make(map[int64]context.CancelFunc)
	procMutex    sync.RWMutex
)

// IncrementSpawnProcs returns the next unique spawn process ID
func IncrementSpawnProcs() int64 {
	return atomic.AddInt64(&spawnProcsCounter, 1)
}

// IncrementRunProcs increments the run process counter
func IncrementRunProcs() {
	atomic.AddInt64(&runProcsCounter, 1)
}

// parseIOConfig parses an I/O config object into an IOConfig struct.
// Returns nil if ioConfigObj is nil or not a map.
func parseIOConfig(ioConfigObj any) *script.IOConfig {
	if ioConfigObj == nil {
		return nil
	}

	ioMap, ok := ioConfigObj.(map[string]any)
	if !ok {
		return nil
	}

	ioConfig := &script.IOConfig{}

	// Extract datastore
	if ds, ok := ioMap["datastore"]; ok {
		if dsStr, ok := ds.(string); ok {
			ioConfig.Datastore = dsStr
		}
	}

	// Extract queue
	if q, ok := ioMap["queue"]; ok {
		if qStr, ok := q.(string); ok {
			ioConfig.Queue = qStr
		}
	}

	// Extract boolean flags (default: all true)
	ioConfig.Out = true  // Default: capture print() output
	ioConfig.Err = true  // Default: capture errors
	ioConfig.Exit = true // Default: capture exit code

	if out, ok := ioMap["out"]; ok {
		if outBool, ok := out.(bool); ok {
			ioConfig.Out = outBool
		}
	}

	if err, ok := ioMap["err"]; ok {
		if errBool, ok := err.(bool); ok {
			ioConfig.Err = errBool
		}
	}

	if exit, ok := ioMap["exit"]; ok {
		if exitBool, ok := exit.(bool); ok {
			ioConfig.Exit = exitBool
		}
	}

	return ioConfig
}

// builtinSpawn runs a script in a background goroutine with an optional context object.
//
// spawn() runs a script in a background goroutine with an optional context object.
// The spawned script receives the context via context() builtin.
// This is fire-and-forget: spawn() returns immediately without waiting.
//
// Example:
//
//	spawn("worker.du", {data = [1, 2, 3]})
//	print("worker running in background")
func builtinSpawn(evaluator *Evaluator, args map[string]any) (any, error) {
	// Get script path or code value
	var scriptPath string
	var program *script.Program
	if sp, ok := args["0"]; ok {
		switch v := sp.(type) {
		case string:
			scriptPath = v
		case *script.ValueRef:
			if v.Val.IsCode() {
				program = v.Val.AsCode().Program
			} else {
				return nil, fmt.Errorf("spawn() arg must be a string path or code value")
			}
		default:
			return nil, fmt.Errorf("spawn() arg must be a string path or code value")
		}
	} else {
		return nil, fmt.Errorf("spawn() requires script path argument")
	}

	// Get context data (optional, named "context" or positional "1") - can be any Duso value
	var contextData any
	if cd, ok := args["context"]; ok {
		contextData = cd
	} else if cd, ok := args["1"]; ok {
		contextData = cd
	}

	// Get I/O config (optional, named "io")
	var ioConfig *script.IOConfig
	if ioCfg, ok := args["io"]; ok {
		ioConfig = parseIOConfig(ioCfg)
	}

	// Get current invocation frame (if in context)
	gid := script.GetGoroutineID()
	var parentFrame *script.InvocationFrame
	if ctx, ok := script.GetRequestContext(gid); ok {
		parentFrame = ctx.Frame
	}

	// Parse script from file if not already provided as code value
	if program == nil {
		// Resolve relative paths relative to the calling script's directory
		resolvedPath := scriptPath
		if parentFrame != nil && parentFrame.Filename != "" {
			resolvedPath = script.ResolveScriptPath(scriptPath, parentFrame.Filename)
		}

		// Parse with caching to avoid re-parsing the same script on repeated spawns
		// This is critical for workloads that spawn the same worker script many times
		var err error
		program, err = globalInterpreter.ParseScript(resolvedPath)
		if err != nil {
			return nil, fmt.Errorf("spawn: failed to parse %s: %w", scriptPath, err)
		}

		scriptPath = resolvedPath
	} else {
		scriptPath = "<dynamic>"
	}

	// Get unique process ID and increment spawn counter
	pid := IncrementSpawnProcs()

	// Create cancellable context for this spawned process
	procCtx, cancel := context.WithCancel(context.Background())

	// Register the cancel function for kill() support
	procMutex.Lock()
	spawnedProcs[pid] = cancel
	procMutex.Unlock()

	// Spawn goroutine (fire-and-forget)
	go func() {
		defer func() {
			// Clean up PID tracking when goroutine exits
			procMutex.Lock()
			delete(spawnedProcs, pid)
			procMutex.Unlock()
		}()
		// Create invocation frame for spawned script
		// Use scriptPath as Filename so scriptDir is correct
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

		// Set up I/O config on the spawned interpreter
		var savedOutputWriter func(string) error
		if ioConfig != nil {
			ioConfig.PID = int(pid)
			globalInterpreter.IOConfig = ioConfig

			// Save original handlers
			savedOutputWriter = globalInterpreter.OutputWriter

			// Replace OutputWriter with I/O routing version
			if ioConfig.Out {
				globalInterpreter.OutputWriter = func(msg string) error {
					return globalInterpreter.AppendToIOQueue("out", msg, ioConfig.PID)
				}
			}

			defer func() {
				globalInterpreter.IOConfig = nil
				globalInterpreter.OutputWriter = savedOutputWriter
			}()
		}

		// Execute script with cancellable context
		result := script.ExecuteScript(
			program,
			globalInterpreter,
			frame,
			spawnedCtx,
			procCtx,
		)

		// Handle errors and exit values
		if result != nil {
			// Route error to queue if configured, otherwise to stderr
			if result.Error != nil {
				var errorMsg string
				if dusoErr, ok := result.Error.(*script.DusoError); ok {
					errorMsg = script.FormatErrorWithStack(dusoErr)
				} else {
					errorMsg = result.Error.Error()
				}

				if ioConfig != nil && ioConfig.Err {
					// Route to datastore queue
					globalInterpreter.AppendToIOQueue("err", errorMsg, ioConfig.PID)
				} else {
					// Log to stderr
					fmt.Fprintf(os.Stderr, "spawn: error in %s: %s\n", scriptPath, errorMsg)
				}
			}

			// Route exit value to queue if configured
			if ioConfig != nil && ioConfig.Exit && result.Value != nil {
				globalInterpreter.AppendToIOQueue("exit", result.Value, ioConfig.PID)
			}
		}
	}()

	return float64(pid), nil
}

// builtinRun executes a script synchronously in a spawned goroutine and blocks until
// the script calls exit() or completes. Returns the value passed to exit().
//
// run() executes a script synchronously in a spawned goroutine and blocks until
// the script calls exit() or completes. Returns the value passed to exit().
//
// Example:
//
//	result = run("worker.du", {data = [1, 2, 3]})
//	print("Result: " + format_json(result))
func builtinRun(evaluator *Evaluator, args map[string]any) (any, error) {
	// Get script path or code value (positional "0" or named "script")
	var scriptPath string
	var program *script.Program
	var sp any
	if scriptArg, ok := args["script"]; ok {
		sp = scriptArg
	} else if scriptArg, ok := args["0"]; ok {
		sp = scriptArg
	} else {
		return nil, fmt.Errorf("run() requires script path argument")
	}

	// Handle string path or code value
	switch v := sp.(type) {
	case string:
		scriptPath = v
	case *script.ValueRef:
		if v.Val.IsCode() {
			program = v.Val.AsCode().Program
		} else {
			return nil, fmt.Errorf("run() script arg must be a string path or code value")
		}
	default:
		return nil, fmt.Errorf("run() script arg must be a string path or code value")
	}

	// Get context data (optional, named "context" or positional "1") - can be any Duso value
	var contextData any
	if cd, ok := args["context"]; ok {
		contextData = cd
	} else if cd, ok := args["1"]; ok {
		contextData = cd
	}

	// Get I/O config (optional, named "io")
	var ioConfig *script.IOConfig
	if ioCfg, ok := args["io"]; ok {
		ioConfig = parseIOConfig(ioCfg)
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

	// Get unique process ID and increment counter (shared with spawn)
	pid := IncrementSpawnProcs()

	// Parse script from file if not already provided as code value
	if program == nil {
		// Resolve relative paths relative to the calling script's directory
		resolvedPath := scriptPath
		if parentFrame != nil && parentFrame.Filename != "" {
			resolvedPath = script.ResolveScriptPath(scriptPath, parentFrame.Filename)
		}

		// Parse with caching to avoid re-parsing the same script on repeated run() calls
		var err error
		program, err = globalInterpreter.ParseScript(resolvedPath)
		if err != nil {
			return nil, fmt.Errorf("run: failed to parse %s: %w", scriptPath, err)
		}

		scriptPath = resolvedPath
	} else {
		scriptPath = "<dynamic>"
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

		// Set up I/O config on the spawned interpreter
		var savedOutputWriter func(string) error
		if ioConfig != nil {
			ioConfig.PID = int(pid)
			globalInterpreter.IOConfig = ioConfig

			// Save original handler
			savedOutputWriter = globalInterpreter.OutputWriter

			// Replace OutputWriter with I/O routing version
			if ioConfig.Out {
				globalInterpreter.OutputWriter = func(msg string) error {
					return globalInterpreter.AppendToIOQueue("out", msg, ioConfig.PID)
				}
			}

			defer func() {
				globalInterpreter.IOConfig = nil
				globalInterpreter.OutputWriter = savedOutputWriter
			}()
		}

		// Execute script (synchronously within the goroutine)
		result := script.ExecuteScript(
			program,
			globalInterpreter,
			frame,
			spawnedCtx,
			timeoutCtx,
		)

		// Handle I/O routing inside the goroutine (before defer clears IOConfig)
		if result != nil {
			// Route error to queue if configured
			if result.Error != nil {
				var errorMsg string
				if dusoErr, ok := result.Error.(*script.DusoError); ok {
					errorMsg = script.FormatErrorWithStack(dusoErr)
				} else {
					errorMsg = result.Error.Error()
				}

				if ioConfig != nil && ioConfig.Err {
					globalInterpreter.AppendToIOQueue("err", errorMsg, ioConfig.PID)
				}
			}

			// Route exit value to queue if configured
			if ioConfig != nil && ioConfig.Exit && result.Value != nil {
				globalInterpreter.AppendToIOQueue("exit", result.Value, ioConfig.PID)
			}
		}

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
			// For DusoError, deep copy the Message value at the process boundary
			if dusoErr, ok := result.Error.(*script.DusoError); ok {
				dusoErr.Message = script.DeepCopyAny(dusoErr.Message)
			}
			return nil, result.Error
		}

		return result.Value, nil
	}
	return nil, nil
}
