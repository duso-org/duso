# context()

Get runtime context information. Returns contextual data from the surrounding environment (HTTP request, spawned script parameters, etc.) or `nil` if no context is available.

## Signature

```duso
context()
```

## Parameters

None

## Returns

Context object (varies by source) or `nil` if no context available

## Usage Patterns

### Detecting Context Availability

The most common pattern is to check whether context exists:

```duso
ctx = context()

if ctx == nil then
  // No context - script is running standalone
  print("Running standalone")
else
  // Context exists - script is running as a handler
  print("Running with context")
end
```

This enables **gate pattern** scripts that work both standalone and as handlers:

```duso
ctx = context()

if ctx == nil then
  // Standalone mode: set up server or spawn child script
  server = http_server({port = 8080})
  server.route("GET", "/")
  server.start()
else
  // Handler mode: process the context
  // ...
end
```

## HTTP Request Context

When called from an HTTP request handler, `context()` returns an object with request handling methods:

### Methods

- `request()` - Get request data
- `callstack()` - Get invocation call stack

### request() Returns

Object with:
- `method` - HTTP method (e.g., "GET", "POST")
- `path` - Request path (e.g., "/api/users")
- `headers` - Object with request headers
- `query` - Object with query parameters
- `body` - Request body as string

### callstack() Returns

Array of invocation frames showing the call path:

```duso
stack = ctx.callstack()
// [
//   {filename = "server.du", line = 8, col = 1, reason = "http_route", method = "GET", path = "/"}
// ]
```

Each frame has:
- `filename` - Script filename
- `line` - Line number
- `col` - Column number
- `reason` - "http_route", "run", or "spawn"
- Additional fields depending on context (method, path for HTTP routes)

### Sending Responses

Use `exit()` to send an HTTP response:

```duso
ctx = context()

if ctx == nil then
  server = http_server({port = 8080})
  server.route("GET", "/api/users")
  server.start()
end

// Handler code
req = ctx.request()

users = [
  {id = 1, name = "Alice"},
  {id = 2, name = "Bob"}
]

exit({
  "status" = 200,
  "body" = format_json(users),
  "headers" = {"Content-Type" = "application/json"}
})
```

### Example

Complete self-referential HTTP server:

```duso
ctx = context()

if ctx == nil then
  server = http_server({port = 8080})
  server.route("GET", "/")
  server.start()
  print("Server stopped")
  exit(0)
end

// Handler code - only runs when ctx != nil
req = ctx.request()

exit({
  "status" = 200,
  "body" = "Hello from " + req.path,
  "headers" = {"Content-Type" = "text/plain"}
})
```

## Spawned Script Context

When a script is spawned with `spawn()` or `run()`, the script receives context with callstack information:

```duso
ctx = context()

if ctx then
  // Script was spawned or run
  stack = ctx.callstack()
  for frame in stack do
    print(frame.filename + ":" + frame.line + " (" + frame.reason + ")")
  end
else
  // Standalone: spawn child script
  result = run("child.du", {data = [1, 2, 3]})
end
```

## Notes

- Returns `nil` when called outside of a handler or spawned context
- Enables flexible scripts that work both standalone and as handlers
- Each request/spawn gets its own context instance
- Use `exit(value)` to return from handlers (becomes HTTP response or run() return value)
- Use `context().callstack()` for debugging and error reporting

## See Also

- [exit() - Return value from script](./exit.md)
- [http_server() - Create HTTP servers](./http_server.md)
- [run() - Execute script synchronously](./run.md)
- [spawn() - Execute script asynchronously](./spawn.md)
