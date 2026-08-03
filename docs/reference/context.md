# context()

Get the data passed into this script, or `nil` when it was run directly.

`context()`

## Parameters

None

## Returns

The context object, or `nil` if the script was invoked with no context.

## The context is the data

`spawn()` and `run()` pass an object to the script they start. That object *is* the
context — read its fields directly.

```duso
// parent.du
result = run("worker.du", {job_id = "job-42", retries = 3})
print("Worker said: " + result.status)
```

```duso
// worker.du
ctx = context()

job_id = ctx.job_id            // "job-42"
retries = ctx.retries or 1     // `or` supplies a default when a field is omitted

exit({status = "done"})        // becomes run()'s return value
```

## Detecting Context Availability

`context()` returns `nil` when the script was run directly, which lets one script
work both ways:

```duso
ctx = context()

if ctx == nil then
  print("Running standalone")
else
  print("Started by another script")
end
```

This is the **gate pattern** — a script that sets things up when run directly, and
does the work when invoked with context:

```duso
ctx = context()

if ctx == nil then
  // Standalone: launch the workers
  for i = 1, 4 do
    spawn("worker.du", {worker_id = i})
  end
else
  // Invoked with context: be a worker
  store = datastore("swarm")
  store.increment("done")
end
```

## Returning a value

`exit(value)` ends the script and hands `value` back to whatever started it:

```duso
exit({status = "done", processed = 42})
```

For `run()` that is the return value. For `spawn()` nothing is waiting for it, so
coordinate through a [datastore()](/docs/reference/datastore.md) instead.

## HTTP handlers

[http_server()](/docs/reference/http_server.md) pre-populates the context of a route
handler with `request()` and `response()`:

```duso
ctx = context()
req = ctx.request()
ctx.response().json({path = req.path})
```

See [http_server()](/docs/reference/http_server.md) for what those provide.

## Notes

- Returns `nil` when the script was invoked with no context
- Each request, spawn, and run gets its own context instance
- Context data is deep-copied across the boundary, and functions are stripped to
  `nil` — closures cannot survive into another script instance

## See Also

- [exit() - Return a value from a script](/docs/reference/exit.md)
- [run() - Execute a script synchronously](/docs/reference/run.md)
- [spawn() - Execute a script asynchronously](/docs/reference/spawn.md)
- [http_server() - Create HTTP servers](/docs/reference/http_server.md)
