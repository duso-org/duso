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

## HTTP Context

When called from an HTTP request handler, `context()` returns an object with request handling methods:

### Methods

- `request()` - Get request data
- `response(data)` - Send HTTP response

### request() Returns

Object with:
- `method` - HTTP method (e.g., "GET", "POST")
- `path` - Request path (e.g., "/api/users")
- `headers` - Object with request headers
- `query` - Object with query parameters
- `body` - Request body as string

### response(data)

Send an HTTP response. Data object supports:
- `status` - HTTP status code (default: 200)
- `body` - Response body as string
- `headers` - Object with response headers

### Example

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

ctx.response({
  "status" = 200,
  "body" = format_json(users),
  "headers" = {"Content-Type" = "application/json"}
})
```

## Future: Script Spawning Context

In upcoming versions, `context()` will also return parameter objects when scripts are spawned with context. The same gate pattern will work:

```duso
ctx = context()

if ctx == nil then
  // Standalone: spawn another script with context
  spawn("child.du", {data = [1, 2, 3]})
else
  // Child mode: process parameters
  data = ctx.data
  print("Received: " + format_json(data))
end
```

## Notes

- Returns `nil` when called outside of a handler or spawned context
- Enables flexible scripts that work both standalone and as handlers
- Each request/spawn gets its own context instance
- Context data varies depending on how the script was invoked

## See Also

- [http_server() - Create HTTP servers](./http_server.md)
