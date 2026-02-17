# Debugger Stepping Implementation Plan

## Context

Duso has an interactive debugger that can be triggered by:
- `breakpoint()` builtin
- **Exceptions in -debug mode**
- **Uncaught throw() in -debug mode**

When debug mode is triggered, users enter a REPL where they can inspect variables and evaluate expressions. Currently, the only execution control is `c` (continue).

## Goal

Add stepping commands:
- **step/s** - Step into (execute next line, any call depth)
- **next/n** - Step over (execute next line at same call depth, skip functions)
- **finish/f** - Step out (continue until current function returns)
- **kill <pid>** - Terminate a runaway spawned process
- **list/l** - Re-display source context

## Design Approach

### Core Mechanism: StepState Flag

Add a `StepState` struct to the Evaluator that acts as a "break on next line" flag:

```go
type StepMode int
const (
    StepModeNone StepMode = 0  // Not stepping
    StepModeInto StepMode = 1  // Step into
    StepModeOver StepMode = 2  // Step over (same/lower depth)
    StepModeOut  StepMode = 3  // Step out (higher depth)
)

type StepState struct {
    Mode        StepMode
    TargetDepth int       // For StepOver/StepOut
    LastPos     Position  // Prevent breaking on same line twice
}
```

When stepping is active, inject checks before statement evaluation that throw a BreakpointError if the step condition is met.

### Injection Points

Add `checkStepBreak(pos)` calls in three statement evaluation loops:
1. **evalProgram** (line 307) - Top-level statements
2. **callScriptFunction** (line 1112) - Function body statements
3. **evalBlock** (line 1449) - Block statements (if/while/for/try)

### REPL Command Handling

Modify `openConsoleDebugREPL` to:
- Accept commands: `s`, `n`, `f`, `l`, `kill <pid>`, `c`
- Return a `*StepState` indicating how to resume
- Thread StepState through DebugManager → ExecuteScript → Evaluator

### Process Registry for kill

Add a global process registry in `pkg/runtime/metrics.go`:
- Map of PID → ProcessHandle (contains context.CancelFunc)
- RegisterProcess() called by spawn()
- KillProcess(pid) cancels the context
- UnregisterProcess() on completion

## Implementation Phases

### Phase 1: Core Infrastructure

**Files**: `pkg/script/evaluator.go`, `pkg/script/script.go`, `pkg/script/debug_manager.go`

1. Add StepMode enum and StepState struct
2. Add `stepState *StepState` field to Evaluator
3. Add `checkStepBreak(pos Position) error` method:
   ```go
   func (e *Evaluator) checkStepBreak(pos Position) error {
       if e.stepState == nil { return nil }

       // Don't break on same position twice
       if e.stepState.LastPos == pos { return nil }

       // Check depth condition based on mode
       shouldBreak := false
       depth := e.ctx.Depth()
       switch e.stepState.Mode {
       case StepModeInto: shouldBreak = true
       case StepModeOver: shouldBreak = depth <= e.stepState.TargetDepth
       case StepModeOut:  shouldBreak = depth < e.stepState.TargetDepth
       }

       if shouldBreak {
           e.stepState = nil  // Clear after single step
           return &BreakpointError{
               FilePath:  e.ctx.FilePath,
               Position:  pos,
               CallStack: e.ctx.CallStack,
               Env:       e.env,
               Message:   "",
           }
       }
       return nil
   }
   ```

4. Update DebugHandler signature:
   ```go
   type DebugHandler func(*DebugEvent) *StepState
   ```

5. Update DebugManager.Wait() to return *StepState

### Phase 2: Injection Points

**File**: `pkg/script/evaluator.go`

Add `checkStepBreak()` before `e.Eval(stmt)` in three loops:

```go
// In evalProgram, callScriptFunction, evalBlock:
if err := e.checkStepBreak(stmt.GetPos()); err != nil {
    return NewNil(), err
}
val, err := e.Eval(stmt)
```

### Phase 3: REPL Commands

**File**: `pkg/cli/debug.go`

Update `openConsoleDebugREPL()` to return `*StepState`:

```go
func openConsoleDebugREPL(interp *Interpreter, bpErr *BreakpointError, msg string) *StepState {
    // ... display context ...

    for {
        line := readInput("debug> ")

        switch line {
        case "c":
            return nil  // Continue without stepping

        case "s", "step":
            return &StepState{
                Mode: StepModeInto,
                LastPos: bpErr.Position,
            }

        case "n", "next":
            depth := len(bpErr.CallStack)
            return &StepState{
                Mode: StepModeOver,
                TargetDepth: depth,
                LastPos: bpErr.Position,
            }

        case "f", "finish":
            depth := len(bpErr.CallStack)
            return &StepState{
                Mode: StepModeOut,
                TargetDepth: depth,
                LastPos: bpErr.Position,
            }

        case "l", "list":
            showSourceContext(...)
            continue

        case starts with "kill ":
            pid := parsePID(line)
            killProcess(pid)
            continue

        default:
            evalInEnvironment(line)
            continue
        }
    }
}
```

### Phase 4: Process Registry

**File**: `pkg/runtime/metrics.go`

```go
type ProcessHandle struct {
    PID        int64
    ScriptPath string
    StartTime  time.Time
    CancelFunc context.CancelFunc
}

type ProcessRegistry struct {
    processes map[int64]*ProcessHandle
    mu        sync.RWMutex
}

var processRegistry = &ProcessRegistry{
    processes: make(map[int64]*ProcessHandle),
}

func RegisterProcess(pid int64, path string, cancel context.CancelFunc) {
    processRegistry.mu.Lock()
    defer processRegistry.mu.Unlock()
    processRegistry.processes[pid] = &ProcessHandle{
        PID: pid, ScriptPath: path,
        StartTime: time.Now(), CancelFunc: cancel,
    }
}

func UnregisterProcess(pid int64) {
    processRegistry.mu.Lock()
    defer processRegistry.mu.Unlock()
    delete(processRegistry.processes, pid)
}

func KillProcess(pid int64) error {
    processRegistry.mu.RLock()
    handle := processRegistry.processes[pid]
    processRegistry.mu.RUnlock()

    if handle == nil {
        return fmt.Errorf("process %d not found", pid)
    }

    handle.CancelFunc()
    UnregisterProcess(pid)
    return nil
}
```

**File**: `pkg/runtime/builtin_spawn.go`

```go
func spawn(...) {
    pid := IncrementSpawnProcs()
    spawnCtx, cancel := context.WithCancel(context.Background())
    RegisterProcess(pid, scriptPath, cancel)

    go func() {
        defer UnregisterProcess(pid)
        result := ExecuteScript(..., spawnCtx, nil)
        // ...
    }()

    return float64(pid), nil
}
```

### Phase 5: ExecuteScript Integration

**File**: `pkg/script/http_server_value.go`

Update signature:
```go
func ExecuteScript(
    program Node,
    interpreter *Interpreter,
    invocationFrame *InvocationFrame,
    requestContext *RequestContext,
    timeoutCtx context.Context,
    initialStepState *StepState,  // NEW
) (*ScriptExecutionResult, *StepState)  // NEW return value
```

Thread StepState through the loop:
```go
childEval.stepState = initialStepState

for _, stmt := range prog.Statements {
    _, execErr := childEval.Eval(stmt)

    if bpErr, ok := execErr.(*BreakpointError); ok {
        // Queue debug event for ALL processes (main and spawned)
        // DebugManager serializes them
        stepState := debugManager.Wait(debugEvent, interpreter)
        childEval.stepState = stepState  // Update for next iteration
        continue
    }
    // ... existing error handling for exceptions/throw
}

return result, childEval.stepState
```

Update all callers (spawn, run, HTTP handlers) to pass `nil` for initial step state.

## Critical Files to Modify

1. **pkg/script/evaluator.go** (~197 lines changes)
   - StepMode, StepState types
   - stepState field
   - checkStepBreak() method
   - 3 injection points

2. **pkg/cli/debug.go** (~50 lines changes)
   - Update REPL to handle new commands
   - Return StepState
   - Add killProcess() helper

3. **pkg/runtime/metrics.go** (~80 lines new code)
   - ProcessRegistry
   - Register/Unregister/Kill functions

4. **pkg/runtime/builtin_spawn.go** (~20 lines changes)
   - Create cancellable context
   - Register/unregister processes

5. **pkg/script/http_server_value.go** (~30 lines changes)
   - ExecuteScript signature
   - Thread StepState

6. **pkg/script/script.go** (~5 lines changes)
   - DebugHandler signature

7. **pkg/script/debug_manager.go** (~20 lines changes)
   - Wait() returns StepState
   - Thread through processor

## Verification Tests

### Test 1: Basic Stepping
```bash
duso -debug test.du
```

Script:
```duso
function foo(x)
    y = x + 1
    return y
end

breakpoint()
z = foo(5)
print(z)
```

Commands to test:
- `n` → step over foo(), should break at print(z)
- `s` → step into foo(), should break at y = x + 1
- `f` → finish foo(), should break at print(z)
- `l` → re-display context
- `c` → continue to end

### Test 2: Exception in Debug Mode
```duso
x = 1
y = x / 0  // Division by zero
```

Run with `duso -debug test.du` - should enter REPL on exception, allow stepping.

### Test 3: Kill Command
```duso
pid = spawn("loop.du")  // Infinite loop
sleep(1)
breakpoint()
```

At breakpoint: `kill 1` → should terminate spawned process

### Test 4: Uncaught throw()
```duso
throw("error!")
```

Run with `duso -debug test.du` - should enter REPL, allow stepping after continue.

## Design Rationale

1. **Reuse BreakpointError** - Stepping is just "automatic breakpoints"
2. **Single-step model** - Each command triggers one break, clears state (like GDB)
3. **Depth-based tracking** - Use existing CallStack.Depth() for next/finish
4. **Position deduplication** - Track LastPos to prevent breaking twice on same line
5. **Context cancellation** - Clean goroutine termination
6. **Remove Parent==nil filter** - Allow spawned processes to queue debug events (that's what the queue is for!)

## Notes

- Exceptions and throw() already trigger debug mode when -debug is active
- Stepping works the same way regardless of what triggered the debug session
- The StepState is cleared after each single step (one break per command)
- kill command uses context cancellation, not unsafe goroutine termination
- Process registry is thread-safe with RWMutex
