# run()

Execute a script synchronously and return its result. Available in `duso` CLI only.

## Signature

```duso
run(script_path [, context [, timeout]])
```

## Parameters

- `script_path` (string) - Path to script file to execute (positional or named `script`)
- `context` (optional, object) - Context object passed to script (positional or named `context`)
- `timeout` (optional, number) - Timeout in seconds (positional as 3rd arg or named `timeout`)

## Returns

The value passed to `exit()` by the script, `nil` if script completes without calling `exit()`, or an error string if timeout exceeded

## Examples

Simple synchronous execution:

```duso
result = run("worker.du")
print("Worker finished with: " + format_json(result))
```

Run with context data and timeout:

```duso
// Positional: script, context, timeout
result = run("processor.du", {data = [1, 2, 3]}, 10)
print("Processed: " + format_json(result))
```

All named arguments:

```duso
result = run(script = "worker.du", timeout = 5)
result = run(script = "processor.du", context = {data = 42}, timeout = 10)
```

## Script Behavior

The executed script can be a standalone script or a gate pattern script:

```duso
ctx = context()

if ctx then
  // Running via run() - has context
  data = ctx.request()  // Can access context if needed
  // Process work...
  exit({status = "done", value = 42})
else
  // Running standalone
  // Can spawn other scripts
  run("child.du", {})
end
```

## Concurrency

- `run()` blocks until the script completes
- Script runs in a separate goroutine with a fresh evaluator
- Parent script waits synchronously for completion
- Each spawned script gets all registered functions (builtins)

## Call Stack

Run scripts can access their call stack:

```duso
ctx = context()
if ctx then
  stack = ctx.callstack()
  // Shows: reason = "run", parent frame = whatever called run()
end
```

## Returning Values

Scripts use `exit()` to return values:

```duso
// child.du
exit({result = 42, status = "ok"})
```

Parent receives the exit value:

```duso
value = run("child.du")
print(value.result)  // 42
```

## Error Handling

If the script errors before calling `exit()`:

```duso
// child.du
print(undefined_var)  // Error!
```

The error is returned as a value (not thrown):

```duso
result = run("child.du")
// result is an error object
```

## Notes

- Synchronous: `run()` always blocks until script completes
- Fire-and-forget equivalent: use `spawn()` instead
- Script output (print) goes to stdout
- Script file is read fresh each time (use caching if performance critical)

## See Also

- [spawn() - Run script asynchronously](/docs/reference/spawn.md)
- [context() - Access handler context](/docs/reference/context.md)
- [exit() - Return value from script](/docs/reference/exit.md)
