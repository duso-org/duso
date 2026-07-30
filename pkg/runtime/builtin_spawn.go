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

	// Get current invocation frame and parent interpreter (if in context)
	callerCtx, callerCtxOk := script.CurrentRequestContext(evaluator)
	var parentFrame *script.InvocationFrame
	var parentInterp *script.Interpreter
	if callerCtxOk && callerCtx != nil {
		parentFrame = callerCtx.Frame
		parentInterp = callerCtx.Interpreter
	}
	if parentInterp == nil {
		return nil, fmt.Errorf("spawn: no interpreter available in context")
	}

	if program == nil {
		// Resolve relative paths relative to the calling script's directory
		resolvedPath := scriptPath
		if parentFrame != nil && parentFrame.Filename != "" {
			resolvedPath = script.ResolveScriptPath(scriptPath, parentFrame.Filename)
		}

		// Parse with caching to avoid re-parsing the same script on repeated spawns
		// This is critical for workloads that spawn the same worker script many times
		var err error
		program, err = parentInterp.ParseScript(resolvedPath)
		if err != nil {
			return nil, fmt.Errorf("spawn: failed to parse %s: %w", scriptPath, err)
		}

		scriptPath = resolvedPath
	} else {
		scriptPath = "<dynamic>"
	}

	pid := runInBackground(program, scriptPath, contextData, parentFrame, ioConfig, "spawn")
	return float64(pid), nil
}

// runInBackground is the shared fire-and-forget execution path used by both
// spawn() and schedule() - a fresh evaluator, a PID registered in spawnedProcs
// so kill() can cancel it, goroutine-local context for context(), and the same
// error/panic routing. parentFrame may be nil (schedule() has no calling script
// frame to chain from, unlike spawn()). Returns the allocated PID.
func runInBackground(program *script.Program, scriptPath string, contextData any, parentFrame *script.InvocationFrame, ioConfig *script.IOConfig, reason string) int64 {
	pid := IncrementSpawnProcs()

	procCtx, cancel := context.WithCancel(context.Background())

	procMutex.Lock()
	spawnedProcs[pid] = cancel
	procMutex.Unlock()

	go func() {
		defer func() {
			procMutex.Lock()
			delete(spawnedProcs, pid)
			procMutex.Unlock()
		}()
		defer func() {
			if r := recover(); r != nil {
				panicMsg := fmt.Sprintf("panic in script: %v", r)
				if ioConfig != nil && ioConfig.Err {
					globalInterpreter.AppendToIOQueue("err", panicMsg, ioConfig.PID)
				} else {
					fmt.Fprintf(os.Stderr, "%s: %s\n", reason, panicMsg)
				}
			}
		}()

		frame := &script.InvocationFrame{
			Filename: scriptPath,
			Line:     1,
			Col:      1,
			Reason:   reason,
			Details:  map[string]any{},
			Parent:   parentFrame,
		}

		spawnedEval := script.NewEvaluator()

		var outputWriter func(string) error
		if ioConfig != nil {
			ioConfig.PID = int(pid)
			outputWriter = func(msg string) error {
				return globalInterpreter.AppendToIOQueue("out", msg, ioConfig.PID)
			}
		} else {
			outputWriter = globalInterpreter.OutputWriter
		}

		spawnedCtx := &script.RequestContext{
			Frame:        frame,
			ProcessCtx:   procCtx,
			Interpreter:  globalInterpreter,
			Evaluator:    spawnedEval,
			IOConfig:     ioConfig,
			OutputWriter: outputWriter,
		}

		spawnedGid := script.GetGoroutineID()
		contextDataCopy := script.DeepCopyAny(contextData)
		script.SetRequestContextWithData(spawnedGid, spawnedCtx, contextDataCopy)
		defer script.ClearRequestContext(spawnedGid)

		SetContextGetter(spawnedGid, func() any {
			ctx, ok := script.GetRequestContext(spawnedGid)
			if !ok {
				return nil
			}
			return ctx.Data
		})
		defer ClearContextGetter(spawnedGid)

		// procCtx is the cancellation context, so kill() actually interrupts the
		// per-statement execution loop.
		result := script.ExecuteScript(program, globalInterpreter, frame, spawnedCtx, procCtx)

		if result != nil {
			if result.Error != nil {
				var errorMsg string
				if dusoErr, ok := result.Error.(*script.DusoError); ok {
					errorMsg = script.FormatErrorWithStack(dusoErr)
				} else {
					errorMsg = result.Error.Error()
				}

				if ioConfig != nil && ioConfig.Err {
					globalInterpreter.AppendToIOQueue("err", errorMsg, ioConfig.PID)
				} else {
					fmt.Fprintf(os.Stderr, "%s: error in %s: %s\n", reason, scriptPath, errorMsg)
				}
			}

			if ioConfig != nil && ioConfig.Exit && result.Value != nil {
				globalInterpreter.AppendToIOQueue("exit", result.Value, ioConfig.PID)
			}
		}
	}()

	return pid
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
	var parentFrame *script.InvocationFrame
	if ctx, ok := script.CurrentRequestContext(evaluator); ok {
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

	// Create fresh evaluator for this execution's environment
	runEval := script.NewEvaluator()

	// Set up output writer - route to datastore if configured, otherwise to global
	var outputWriter func(string) error
	if ioConfig != nil {
		ioConfig.PID = int(pid)
		outputWriter = func(msg string) error {
			return globalInterpreter.AppendToIOQueue("out", msg, ioConfig.PID)
		}
	} else {
		outputWriter = globalInterpreter.OutputWriter
	}

	// Create spawned context with per-execution state
	spawnedCtx := &script.RequestContext{
		Frame:        frame,
		ProcessCtx:   timeoutCtx,
		Interpreter:  globalInterpreter,
		Evaluator:    runEval,
		IOConfig:     ioConfig,
		OutputWriter: outputWriter,
	}

	// Execute script in goroutine and collect results
	resultChan := make(chan *script.ScriptExecutionResult, 1)
	done := make(chan bool, 1)

	go func() {
		defer func() {
			// Capture panic and convert to error result
			if r := recover(); r != nil {
				// Format panic as error message
				panicMsg := fmt.Sprintf("panic in script: %v", r)
				dusoErr := &script.DusoError{
					Message:  panicMsg,
					FilePath: scriptPath,
				}
				result := &script.ScriptExecutionResult{
					Error: dusoErr,
				}
				resultChan <- result
				done <- true
				return
			}
		}()
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
			globalInterpreter,
			frame,
			spawnedCtx,
			timeoutCtx,
		)

		// Handle I/O routing inside the goroutine
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
