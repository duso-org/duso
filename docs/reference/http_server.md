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
  - `timeout` (number) - Socket read/write timeout in seconds (default: 30)
  - `request_handler_timeout` (number) - Handler script execution timeout in seconds (default: 30)
  - `directory` (boolean) - Enable directory listing when no default file is found (default: false)
  - `default` (string or array) - Default file(s) to serve in directories (default: ["index.html"]). Can be a single filename, comma-separated list, or array of filenames. Set to nil or empty to disable defaults.

## Returns

HTTP server object with methods

## Methods

- `route(method, path [, handler])` - Register a route with a handler script
  - `method` - HTTP method: `"GET"`, `"POST"`, `"DELETE"`, etc., or `"*"`/`nil` for all methods
  - `path` - URL path (supports prefix matching)
  - `handler` - (optional) Path to handler script. If omitted, uses current script
- `static(path, directory)` - Serve static files from a directory
  - `path` - URL path prefix (e.g., `"/"` or `"/public"`)
  - `directory` - Directory path to serve files from (e.g., `"./public"` or `"."`)
- `start()` - Start the server (blocks until Ctrl+C, then returns)

## Examples

Quick static file server (one-liner):

```bash
duso -c 'http_server().start()'
```

This starts a server on `http://localhost:8080` that serves all files from the current directory. Perfect for quick testing and local development.

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
exit({
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

Server with handler timeout:

```duso
server = http_server({
  port = 8080,
  request_handler_timeout = 5  // 5-second timeout per request
})
server.route("GET", "/fast", "handlers/fast.du")
server.route("GET", "/slow", "handlers/slow.du")

server.start()
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

exit({
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

Serving static files:

```duso
server = http_server({port = 8080})
server.static("/", "./public")
server.start()
```

This serves all files from the `./public` directory at the root path. Requests to `/index.html`, `/style.css`, `/images/logo.png` etc. will be served directly from disk with appropriate MIME types detected automatically.

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

You can send responses in two ways:

### Using Response Convenience Wrappers

Access response helpers via `context().response()`:

```duso
ctx = context()
resp = ctx.response()

// Send JSON response
resp.json({id = 1, name = "Alice"}, 200)

// Send plain text
resp.text("Hello, World!", 200)

// Send HTML
resp.html("<h1>Welcome</h1>", 200)

// Send error
resp.error(404, "Not Found")

// Send redirect
resp.redirect("https://example.com", 302)

// Send file
resp.file("./public/index.html", 200)

// Generic response with custom headers
resp.response("Custom body", 200, {"X-Custom" = "Header"})
```

**Response wrapper methods:**
- `json(data, [status])` - Send JSON response with `Content-Type: application/json`
- `text(data, [status])` - Send plain text with `Content-Type: text/plain`
- `html(data, [status])` - Send HTML with `Content-Type: text/html`
- `error(status, [message])` - Send error response with JSON body
- `redirect(url, [status])` - Send redirect (default status: 302)
- `file(path, [status])` - Serve file from filesystem
- `response(data, status, [headers])` - Generic response with custom headers

All methods have optional status parameter (default: 200). Calling any response method immediately sends the response and exits the handler.

### Using exit()

Alternatively, use `exit()` with a response object:

```duso
exit({
  "status" = 200,
  "body" = "response body",
  "headers" = {"Content-Type" = "text/plain"}
})
```

The response object supports:
- `status` - HTTP status code (default: 200)
- `body` - Response body as string
- `headers` - Object with response headers

If the handler doesn't call `exit()` or use a response wrapper, the response will be 204 No Content.

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
req = ctx.request()  // Access request data
ctx.response().json({message = "Hello"}, 200)  // Send response
```

The `ctx.response()` method returns an object with convenience wrappers for different response types (json, text, html, error, redirect, file). Use any of these to send a response, or fall back to `exit()` for full control.

This pattern enables a complete server in a single script file, perfect for simple applications.

## Concurrency

Each incoming request runs in a separate goroutine with a fresh evaluator instance, providing true concurrent request handling with no shared state.

## Static File Routes

Static routes registered with `static()` behave differently from handler routes:
- Files are served directly from the filesystem
- Content type is determined automatically based on file extension
- Missing files return 404 responses
- No handler script execution or timeout applies
- Efficient for serving assets, HTML, CSS, JavaScript, images, etc.

## Notes

- Server blocks until it receives a Ctrl+C signal (SIGINT) or termination signal (SIGTERM)
- Routes can be registered after `start()` is called (thread-safe)
- Both handler routes (`route()`) and static routes (`static()`) can be used together
- Script execution continues after `server.start()` returns, allowing cleanup code
- Handler scripts are loaded from disk for each request (use caching/preprocessing if performance is critical)
- Use `exit()` to send responses (becomes HTTP response with status/body/headers)
- If handler doesn't call `exit()`, response is 204 No Content
- If handler exceeds `request_handler_timeout`, response is 504 Gateway Timeout
- `timeout` (socket level) and `request_handler_timeout` (handler script) are independent
- Available in `duso` CLI only

## See Also

- [context() - Access request context in handlers](/docs/reference/context.md)
- [fetch() - Make HTTP requests](/docs/reference/fetch.md)
