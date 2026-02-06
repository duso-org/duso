# spawn()

Spawn a script in a background goroutine with optional context. Fire-and-forget execution. Available in `duso` CLI only.

## Signature

```duso
spawn(script_path [, context])
```

## Parameters

- `script_path` (string) - Path to script file to spawn
- `context` (optional, object) - Context object passed to spawned script

## Returns

Numeric process ID (number) - A unique identifier for the spawned process. Returns immediately (fire-and-forget).

## Examples

Simple spawn without context:

```duso
pid = spawn("worker.du")
print("Spawned worker with PID: " + pid)
```

Spawn with context data:

```duso
pid = spawn("processor.du", {data = [1, 2, 3], timeout = 30})
print("Process started: " + pid)
```

Track spawned processes:

```duso
pids = []
push(pids, spawn("worker1.du"))
push(pids, spawn("worker2.du"))
push(pids, spawn("worker3.du"))
print("Spawned " + len(pids) + " workers with PIDs: " + format_json(pids))
```

## Spawned Script Context

The spawned script can check for context and access the call stack:

```duso
ctx = context()

if ctx then
  // Has context from spawn()
  stack = ctx.callstack()
  print("Spawned from: " + stack[0].filename)
else
  // Standalone execution
  spawn("child.du", {})
end
```

## Concurrency

- `spawn()` returns immediately (fire-and-forget)
- Spawned script runs in a separate goroutine
- Each spawned script gets a fresh evaluator with all registered functions
- No direct communication with parent script (use callbacks or side effects like file I/O)

## Call Stack

Spawned scripts can access their call stack via `context().callstack()`, showing the chain of spawns that led to execution:

```duso
ctx = context()
stack = ctx.callstack()

// Example output for nested spawns:
// [
//   {filename = "grandchild.du", line = 1, col = 1, reason = "spawn"},
//   {filename = "child.du", line = 5, col = 3, reason = "spawn"},
//   {filename = "parent.du", line = 2, col = 1, reason = "spawn"}
// ]
```

## Notes

- Fire-and-forget: spawn() returns immediately with a process ID
- Process ID (PID) is a unique identifier that can be stored for later reference
- PIDs are assigned sequentially (1, 2, 3, ...) for each spawned process
- No direct way to kill or wait for a process yet (planned for future releases)
- Script lifecycle is managed by the script itself (no external control)
- All registered Go functions (builtins) are available in spawned scripts
- Spawned scripts run with full access to the runtime (file I/O, HTTP, etc.)

## See Also

- [context() - Access runtime context](/docs/reference/context.md)
- [run() - Run script synchronously](/docs/reference/run.md)
- [exit() - Return value from script](/docs/reference/exit.md)
- [parallel() - Run multiple functions concurrently with results](/docs/reference/parallel.md)
