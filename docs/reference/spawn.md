# spawn()

Spawn a script in a background goroutine with optional context. Fire-and-forget execution. Available in `duso` CLI only.

## Signature

```duso
spawn(script_path [, context] [, io])
```

## Parameters

- `script_path` (string) - Path to script file to spawn
- `context` (optional, object) - Context object passed to spawned script
- `io` (optional, object) - I/O routing configuration for capturing process output/errors/exit codes

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

## I/O Routing

Capture process output, errors, and exit codes to a datastore queue for inspection by agents:

```duso
spawn("worker.du", {data = [1, 2, 3]}, io = {
  datastore = "logs",
  queue = "worker-1",
  out = true,   // Capture print() output (default: true)
  err = true,   // Capture errors and exceptions (default: true)
  exit = true   // Capture exit code (default: true)
})

sleep(0.5)
logs = datastore("logs").get("worker-1")
// logs = [
//   {pid: 1, out: "message\n"},
//   {pid: 1, err: "error with stack trace\n..."},
//   {pid: 1, exit: 42}
// ]
```

**I/O Config Fields:**

- `datastore` (string) - Name of the datastore to use for the queue
- `queue` (string) - Key in the datastore where I/O events are appended
- `out` (boolean, default: true) - Route `print()` and `write()` to queue
- `err` (boolean, default: true) - Route runtime errors to queue (with full stack trace)
- `exit` (boolean, default: true) - Route exit code from `exit()` to queue

**Queue Format:**

```
{pid: 1, out: "output message\n"}
{pid: 1, err: "file:line: Error message\n\nCall stack:\n  at func (file:line)..."}
{pid: 1, exit: 42}
```

Errors include full call stack with file locations and function names, matching debug output.

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
