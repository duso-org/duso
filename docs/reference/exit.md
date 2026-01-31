# exit()

Terminate the current script and optionally return a value to the caller.

## Signature

```duso
exit([value])
```

## Parameters

- `value` (optional, any type) - Value to return to the caller. Meaning depends on context.

## Returns

Never returns (terminates script execution)

## Context-Dependent Behavior

The value passed to `exit()` has different meanings depending on how the script was invoked:

### HTTP Request Handler

The value becomes the HTTP response:

```duso
ctx = context()
req = ctx.request()

exit({
  "status" = 200,
  "body" = "response body",
  "headers" = {"Content-Type" = "text/plain"}
})
```

If no `exit()` is called, the response is 204 No Content.

### run() Script

The value becomes the return value from `run()`:

```duso
// worker.du
exit({status = "done", value = 42})
```

```duso
// main script
result = run("worker.du")
print(result.value)  // 42
```

### spawn() Script

The script terminates (value is ignored):

```duso
// spawned.du
exit({status = "completed"})  // Value is not captured
```

### Main Script (CLI)

Exits Duso (value is ignored):

```duso
print("Doing work...")
exit()  // Duso exits with status 0
```

## Examples

Returning data from a worker:

```duso
// worker.du
data = [1, 2, 3, 4, 5]
sum = 0
for item in data do
  sum = sum + item
end
exit(sum)  // Returns 15
```

```duso
// main
result = run("worker.du")
print("Sum: " + result)  // 15
```

HTTP handler returning JSON:

```duso
ctx = context()
if ctx then
  users = [
    {id = 1, name = "Alice"},
    {id = 2, name = "Bob"}
  ]

  exit({
    "status" = 200,
    "body" = format_json(users),
    "headers" = {"Content-Type" = "application/json"}
  })
end
```

Self-referential server cleanup:

```duso
ctx = context()

if ctx == nil then
  server = http_server({port = 8080})
  server.route("GET", "/")
  server.start()

  // Cleanup after server stops
  print("Server shutdown complete")
  exit(0)
end

// Handler code
req = ctx.request()
exit({status = 200, body = "Hello"})
```

## Notes

- `exit()` terminates the current script immediately
- No code after `exit()` will execute
- If called in the main script, Duso exits (value is ignored)
- If called in a spawned script (`spawn()`), the value is lost
- If called in an HTTP handler, the value must be a map with response structure
- If called in a `run()` script, the value becomes the return value for the caller
- Calling `exit()` without a value returns `nil` to the caller

## See Also

- [context() - Access handler context](/docs/reference/context.md)
- [run() - Execute script and get result](/docs/reference/run.md)
- [http_server() - Create HTTP servers](/docs/reference/http_server.md)
