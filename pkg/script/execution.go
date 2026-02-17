package script

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

// ScriptExecutionResult holds the result of script execution
type ScriptExecutionResult struct {
	Value any   // The exit value or nil
	Error error // Any error that occurred
}

// ExecuteScript executes a parsed script with proper exception handling.
// Used by run(), spawn(), and HTTP handlers to unify script execution and error handling.
func ExecuteScript(
	program Node,
	interpreter *Interpreter,
	invocationFrame *InvocationFrame,
	requestContext *RequestContext,
	timeoutCtx context.Context,
) *ScriptExecutionResult {
	// Create fresh evaluator
	childEval := NewEvaluator()

	// Set the script filename for error reporting
	if invocationFrame != nil && invocationFrame.Filename != "" {
		childEval.ctx.FilePath = invocationFrame.Filename
	}

	// Copy custom functions from interpreter
	if interpreter != nil {
		parentEval := interpreter.GetEvaluator()

		// Copy custom registered functions
		for name, fn := range parentEval.GetGoFunctions() {
			childEval.RegisterFunction(name, fn)
		}
	}

	// Execute statements one-by-one so breakpoints can pause and resume mid-execution
	prog, ok := program.(*Program)
	if !ok {
		// Not a program, evaluate as single node
		_, err := childEval.Eval(program)
		select {
		case <-timeoutCtx.Done():
			return &ScriptExecutionResult{Value: nil, Error: fmt.Errorf("timeout exceeded")}
		default:
		}
		if err != nil {
			return &ScriptExecutionResult{Value: nil, Error: err}
		}
		return &ScriptExecutionResult{Value: nil, Error: nil}
	}

	// Execute statements one-by-one
	for _, stmt := range prog.Statements {
		// Check for timeout
		select {
		case <-timeoutCtx.Done():
			return &ScriptExecutionResult{
				Value: nil,
				Error: fmt.Errorf("timeout exceeded"),
			}
		default:
		}

		_, execErr := childEval.Eval(stmt)
		if execErr != nil {
			// Check for BreakpointError
			if bpErr, ok := execErr.(*BreakpointError); ok {
				debugEvent := &DebugEvent{
					Error:           bpErr,
					FilePath:        bpErr.FilePath,
					Position:        bpErr.Position,
					CallStack:       bpErr.CallStack,
					InvocationStack: invocationFrame,
					Env:             bpErr.Env,
					Message:         bpErr.Message,
				}
				if interpreter != nil {
					// Queue debug event for all goroutines
					debugManager := GetDebugManager()
					debugManager.Wait(debugEvent, interpreter)
				}
				// Continue to next statement after breakpoint
				continue
			}

			// Check for ExitExecution
			if exitErr, ok := execErr.(*ExitExecution); ok {
				var exitValue any
				if len(exitErr.Values) > 0 {
					exitValue = exitErr.Values[0]
				}
				return &ScriptExecutionResult{
					Value: exitValue,
					Error: nil,
				}
			}

			// In debug mode, other errors trigger debug queue
			// Check if debug mode is enabled by calling sys("-debug")
			debugMode := false
			if sysBuiltin := GetBuiltin("sys"); sysBuiltin != nil {
				result, _ := sysBuiltin(childEval, map[string]any{"0": "-debug"})
				if b, ok := result.(bool); ok {
					debugMode = b
				}
			}

			if debugMode {
				debugEvent := &DebugEvent{
					Error:           execErr,
					Message:         execErr.Error(),
					FilePath:        invocationFrame.Filename,
					InvocationStack: invocationFrame,
					Env:             childEval.GetEnv(),
				}
				if dusoErr, ok := execErr.(*DusoError); ok {
					debugEvent.FilePath = dusoErr.FilePath
					debugEvent.Position = dusoErr.Position
					debugEvent.CallStack = dusoErr.CallStack
				}
				if interpreter != nil {
					// Queue debug event for all goroutines
					debugManager := GetDebugManager()
					debugManager.Wait(debugEvent, interpreter)
				}
				continue
			}

			// Regular error - return it
			return &ScriptExecutionResult{
				Value: nil,
				Error: execErr,
			}
		}
	}

	return &ScriptExecutionResult{
		Value: nil,
		Error: nil,
	}
}

// InvocationFrame represents a single level in the call stack
type InvocationFrame struct {
	Filename string           // Script filename
	Line     int              // Line number where invocation happened
	Col      int              // Column number
	Reason   string           // "http_route", "spawn", etc.
	Details  map[string]any   // Additional context (method, path, etc.)
	Parent   *InvocationFrame // Previous frame in chain
}

// RequestContext holds context data for any spawned/invoked script
// Used for spawn() calls, run() calls, and HTTP handlers
type RequestContext struct {
	Data     any              // Generic context data (spawn/run data or HTTP request/response functions)
	Frame    *InvocationFrame // Root invocation frame for this context
	ExitChan chan any         // Channel to receive exit value from script
	closed   bool
	mutex    sync.Mutex
}

// Global goroutine-local storage for request contexts
var (
	requestContexts = make(map[uint64]*RequestContext)
	contextMutex    sync.RWMutex
)

// GetGoroutineID extracts the current goroutine ID from the stack trace
func GetGoroutineID() uint64 {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	stackTrace := string(buf[:n])

	// Parse "goroutine 123 [running]:"
	lines := strings.Split(stackTrace, "\n")
	if len(lines) > 0 {
		line := lines[0]
		if strings.HasPrefix(line, "goroutine ") {
			parts := strings.Fields(line)
			if len(parts) > 1 {
				if id, err := strconv.ParseUint(parts[1], 10, 64); err == nil {
					return id
				}
			}
		}
	}
	return 0
}

// setRequestContext stores a request context in goroutine-local storage
func setRequestContext(gid uint64, ctx *RequestContext) {
	contextMutex.Lock()
	defer contextMutex.Unlock()
	requestContexts[gid] = ctx
}

// SetRequestContextWithData stores a request context with optional spawned context data
func SetRequestContextWithData(gid uint64, ctx *RequestContext, spawnedData any) {
	contextMutex.Lock()
	defer contextMutex.Unlock()

	// Store data in the Data field (generic context)
	ctx.Data = spawnedData

	requestContexts[gid] = ctx
}

// GetRequestContext retrieves a request context from goroutine-local storage
func GetRequestContext(gid uint64) (*RequestContext, bool) {
	contextMutex.RLock()
	defer contextMutex.RUnlock()
	ctx, ok := requestContexts[gid]
	return ctx, ok
}

// ClearRequestContext removes a request context from goroutine-local storage
func ClearRequestContext(gid uint64) {
	contextMutex.Lock()
	defer contextMutex.Unlock()
	delete(requestContexts, gid)
}

// clearRequestContext removes a request context from goroutine-local storage
func clearRequestContext(gid uint64) {
	contextMutex.Lock()
	defer contextMutex.Unlock()
	delete(requestContexts, gid)
}
