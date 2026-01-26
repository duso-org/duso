# http_server()

Create an HTTP server that listens for incoming requests and runs handler scripts. Available in `duso` CLI only.

## Signature

```duso
http_server([config])
```

## Parameters

- `config` (optional, object) - Configuration object with options:
  - `port` (number) - Port to listen on (default: 8080)
  - `address` (string) - Bind address (default: "0.0.0.0")
  - `https` (boolean) - Enable HTTPS (default: false)
  - `cert_file` (string) - Path to TLS certificate (required if https=true)
  - `key_file` (string) - Path to TLS private key (required if https=true)
  - `timeout` (number) - Request timeout in seconds (default: 30)

## Returns

HTTP server object with methods

## Methods

- `route(method, path [, handler])` - Register a route
  - `method` - HTTP method: `"GET"`, `"POST"`, `"DELETE"`, etc., or `"*"`/`nil` for all methods
  - `path` - URL path (supports prefix matching)
  - `handler` - (optional) Path to handler script. If omitted, uses current script
- `start()` - Start the server (blocks until Ctrl+C, then returns)

## Examples

Minimal self-referential server:

```duso
ctx = context()

if ctx == nil then
  // Server setup mode
  server = http_server({port = 8080})
  server.route("GET", "/")
  server.start()
end

// Handler mode - only runs when ctx != nil
req = ctx.request()
ctx.response({
  "status" = 200,
  "body" = "hello world",
  "headers" = {"Content-Type" = "text/plain"}
})
```

Multiple routes with different handlers:

```duso
server = http_server({port = 8080})
server.route("GET", "/", "handlers/home.du")
server.route("GET", "/api/users", "handlers/users.du")
server.route("POST", "/api/users", "handlers/create_user.du")
server.route("DELETE", "/api/users", "handlers/delete_user.du")

print("Server listening on http://localhost:8080")
server.start()
print("Server stopped")
```

Handling requests in a handler script:

```duso
// handlers/users.du
ctx = context()
req = ctx.request()

// req contains: method, path, headers, query, body
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

Multiple methods on the same route:

```duso
server = http_server({port = 8080})
server.route(["GET", "POST"], "/api/data")
server.start()
```

## Request Context

Inside a handler script, call `context()` to access request data:

```duso
ctx = context()
req = ctx.request()

// req object contains:
// - method: HTTP method (e.g., "GET", "POST")
// - path: Request path (e.g., "/api/users")
// - headers: Object with request headers
// - query: Object with query parameters
// - body: Request body as string
```

## Sending Responses

Use `ctx.response()` to send an HTTP response:

```duso
ctx.response({
  "status" = 200,
  "body" = "response body",
  "headers" = {"Content-Type" = "text/plain"}
})
```

The response object supports:
- `status` - HTTP status code (default: 200)
- `body` - Response body as string
- `headers` - Object with response headers

## Routing

Routes support prefix matching, with the most specific route taking priority:

```duso
server.route("GET", "/api")        // Matches /api, /api/users, /api/users/123
server.route("GET", "/api/users")  // More specific, matches /api/users and /api/users/123
```

When a request matches multiple routes, the longest (most specific) path is used.

## Self-Referential Scripts

A single script file can be both the server and its handlers using the gate pattern:

```duso
ctx = context()

if ctx == nil then
  // Server setup: ctx is nil when not in a request handler
  server = http_server({port = 8080})
  server.route("GET", "/")
  server.start()
end

// Handler code: only runs when ctx != nil
ctx.request()  // Access request data
ctx.response({...})  // Send response
```

This pattern enables a complete server in a single script file, perfect for simple applications.

## Concurrency

Each incoming request runs in a separate goroutine with a fresh evaluator instance, providing true concurrent request handling with no shared state.

## Notes

- Server blocks until it receives a Ctrl+C signal (SIGINT) or termination signal (SIGTERM)
- Routes can be registered after `start()` is called (thread-safe)
- Script execution continues after `server.start()` returns, allowing cleanup code
- Handler scripts are loaded from disk for each request (use caching/preprocessing if performance is critical)
- Available in `duso` CLI only

## See Also

- [context() - Access request context in handlers](./nil.md)
- [http_client() - Make HTTP requests](./http_client.md)
