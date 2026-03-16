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
  - `max_body_size` (number) - Max request body size in bytes (default: 10485760 = 10MB). Returns 413 Payload Too Large if exceeded.
  - `max_header_size` (number) - Max per-header size in bytes (default: 8192 = 8KB). Returns 431 Request Header Fields Too Large if exceeded.
  - `max_headers` (number) - Max number of headers in request (default: 100)
  - `max_form_fields` (number) - Max number of form fields in multipart/form-data (default: 1000)
  - `idle_timeout` (number) - Idle connection timeout in seconds (default: 120)
  - `access_log` (boolean) - Enable access logging to stderr in Apache Combined Log Format (default: true)
  - `directory` (boolean) - Enable directory listing when no default file is found (default: false)
  - `default` (string or array) - Default file(s) to serve in directories (default: ["index.html"]). Can be a single filename, comma-separated list, or array of filenames. Set to nil or empty to disable defaults.
  - `cache_control` (string) - Cache-Control header for dynamic responses. Used by response helpers (html(), json(), text()) unless handler sets custom headers (default: "no-cache, no-store, must-revalidate").
  - `static_cache_control` (string) - Cache-Control header for static file responses (default: "public, max-age=3600"). Set to empty string to disable.
  - `cors` (object) - CORS configuration (optional):
    - `enabled` (boolean) - Enable CORS (default: false)
    - `origins` (string or array) - Allowed origins: `"*"` for all, or array of specific origins (default: [])
    - `methods` (string or array) - Allowed HTTP methods (default: [])
    - `headers` (string or array) - Allowed request headers (default: [])
    - `credentials` (boolean) - Allow credentials in CORS requests (default: false)
    - `max_age` (number) - Max age for preflight cache in seconds (default: 0)
  - `jwt` (object) - JWT configuration (optional):
    - `enabled` (boolean) - Enable JWT verification (default: false)
    - `secret` (string) - Secret key for HS256 signing/verification (required if enabled)
    - `required` (boolean) - Require valid JWT token for all requests (default: false)

## Returns

HTTP server object with methods

## Methods

- `route(method, path [, handler])` - Register a route with a handler script
  - `method` - HTTP method: string, array of strings, `"*"`, or `nil` for all methods
    - Valid methods: `"GET"`, `"POST"`, `"PUT"`, `"DELETE"`, `"PATCH"`, `"HEAD"`, `"OPTIONS"`, `"TRACE"`, `"CONNECT"`
    - Case-insensitive (e.g., `"get"`, `"Get"`, `"GET"` all work)
    - Array example: `["GET", "POST"]` to handle both methods on same path
    - `"*"` or `nil` matches all HTTP methods
  - `path` - URL path pattern:
    - Exact match: `/api` matches only `/api`
    - Parameterized: `/users/:id` matches `/users/123` etc. with params
    - Wildcard: `/api/*` matches `/api/` and everything under it (prefix match)
  - `handler` - (optional) Path to handler script. If omitted, uses current script
- `static(path, directory)` - Serve static files from a directory
  - `path` - URL path prefix (e.g., `"/"` or `"/public"`)
  - `directory` - Directory path to serve files from (e.g., `"./public"` or `"."`)
- `start()` - Start the server (blocks until Ctrl+C, then returns)

## Access Logging

By default, the server logs all HTTP requests to stderr in Apache Combined Log Format, a standard format used by web servers like Apache and Nginx. This makes logs stream-friendly and compatible with log aggregation tools.

Log format:
```
remotehost rfc931 authuser [timestamp] "request line" status bytes_sent "referrer" "user-agent"
```

Example:
```
127.0.0.1 - - [16/Mar/2026 10:45:09 -0700] "GET /api/users HTTP/1.1" 200 1234 "-" "curl/7.64.1"
```

Access logging can be disabled by setting `access_log = false` in the configuration. Log lines are truncated if they would exceed 4KB to ensure atomic writes to stderr.

## Resource Limits

The server enforces strict resource limits for incoming requests to prevent abuse:

- **Body size**: Requests exceeding `max_body_size` are rejected with 413 Payload Too Large before the handler script runs
- **Header size**: If a single header exceeds `max_header_size`, the server will enforce this via http.Server (standard behavior)
- **Header count**: Requests with more headers than `max_headers` are rejected with 431 Request Header Fields Too Large before the handler runs
- **Form fields**: Requests with more form fields than `max_form_fields` are rejected with 400 Bad Request before the handler runs

These limits are enforced at the HTTP level, not in the handler script, so they provide true DOS protection.

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

Server with custom cache control:

```duso
// Default cache_control is "no-cache, no-store, must-revalidate" (no browser caching)
// Override with custom value:
server = http_server({
  port = 8080,
  cache_control = "public, max-age=3600"  // Cache for 1 hour
})
server.route("GET", "/", "handlers/index.du")
server.start()
```

The `cache_control` setting applies to all responses from `response.html()`, `response.json()`, and `response.text()` helpers. Handlers can override by setting custom Cache-Control headers in their response.

Server with resource limits:

```duso
server = http_server({
  port = 8080,
  max_body_size = 50 * 1024 * 1024,      // 50MB max request body
  max_header_size = 16 * 1024,           // 16KB max per header
  max_headers = 100,                     // 100 headers max
  max_form_fields = 1000,                // 1000 form fields max
  idle_timeout = 60                      // 60s idle timeout
})
server.route("POST", "/upload", "handlers/upload.du")
server.start()
```

These limits protect against malformed or malicious requests:
- `max_body_size` - Returns 413 Payload Too Large if exceeded (early rejection via Content-Length check)
- `max_header_size` - Returns 431 Request Header Fields Too Large if exceeded
- `max_form_fields` - Returns 400 Bad Request if exceeded (enforced at HTTP level before handler)
- `idle_timeout` - Closes connections that have been idle for longer than specified

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

API server with CORS and JWT authentication:

```duso
server = http_server({
  port = 8080,
  cors = {
    "enabled" = true,
    "origins" = ["https://example.com", "https://app.example.com"],
    "methods" = ["GET", "POST", "DELETE"],
    "headers" = ["Content-Type", "Authorization"],
    "credentials" = true,
    "max_age" = 86400
  },
  jwt = {
    "enabled" = true,
    "secret" = "your-secret-key-here",
    "required" = false
  }
})

server.route("POST", "/auth/login", "handlers/login.du")
server.route("GET", "/api/profile", "handlers/profile.du")
server.route("DELETE", "/api/session", "handlers/logout.du")

server.start()
```

Handler that signs a JWT token:

```duso
// handlers/login.du
ctx = context()
req = ctx.request()
resp = ctx.response()

// Verify credentials (simplified example)
body = req.body
user_id = "user123"

// Sign token valid for 24 hours
token = resp.sign_jwt({
  "sub" = user_id,
  "iat" = time()
}, 86400)

resp.json({
  "access_token" = token,
  "token_type" = "Bearer",
  "expires_in" = 86400
})
```

Handler that verifies JWT token:

```duso
// handlers/profile.du
ctx = context()
req = ctx.request()
resp = ctx.response()

claims = req.jwt_claims

if claims == nil then
  resp.error(401, "Authorization required")
else
  user_id = claims["sub"]

  resp.json({
    "id" = user_id,
    "name" = "John Doe",
    "email" = "john@example.com"
  })
end
```

## CORS (Cross-Origin Resource Sharing)

Enable CORS to allow cross-origin requests from web browsers:

```duso
server = http_server({
  port = 8080,
  cors = {
    "enabled" = true,
    "origins" = "*",                                    // or ["https://example.com", "https://app.example.com"]
    "methods" = ["GET", "POST", "PUT", "DELETE"],
    "headers" = ["Content-Type", "Authorization"],
    "credentials" = false,
    "max_age" = 3600
  }
})
server.route("GET", "/api")
server.start()
```

CORS handles:
- Setting `Access-Control-Allow-Origin` header based on configured origins
- Returning `Access-Control-Allow-Methods`, `Access-Control-Allow-Headers`, `Access-Control-Allow-Credentials`, and `Access-Control-Max-Age` headers
- Responding to preflight OPTIONS requests with 204 No Content

**CORS options:**
- `enabled` - Enable CORS support (default: false)
- `origins` - Allowed origins: `"*"` for all origins, or array of specific origins
- `methods` - Allowed HTTP methods (e.g., `["GET", "POST"]`)
- `headers` - Allowed request headers (e.g., `["Content-Type", "Authorization"]`)
- `credentials` - Allow requests with credentials like cookies (default: false)
- `max_age` - Cache time for preflight responses in seconds (default: 0)

## JWT (JSON Web Tokens)

Enable JWT to sign and verify bearer tokens in Authorization headers:

```duso
server = http_server({
  port = 8080,
  jwt = {
    "enabled" = true,
    "secret" = "your-secret-key",
    "required" = false
  }
})
server.route("POST", "/login")
server.route("GET", "/api/profile")
server.start()
```

### Signing Tokens

Create JWT tokens in handler responses:

```duso
// handlers/login.du
ctx = context()
resp = ctx.response()

// Sign a token that expires in 1 hour (3600 seconds)
token = resp.sign_jwt({
  "sub" = "user123",
  "email" = "user@example.com",
  "role" = "admin"
}, 3600)

resp.json({"token" = token}, 200)
```

### Verifying Tokens

Access verified claims in incoming requests:

```duso
// handlers/profile.du
ctx = context()
req = ctx.request()

// req.jwt_claims contains the verified token claims (or nil if no valid token)
if req.jwt_claims == nil then
  resp = ctx.response()
  resp.error(401, "No valid JWT token")
else
  claims = req.jwt_claims
  user_id = claims["sub"]

  ctx.response().json({
    "user_id" = user_id,
    "email" = claims["email"],
    "role" = claims["role"]
  })
end
```

**JWT Implementation:**
- Uses HS256 (HMAC-SHA256) algorithm
- No external dependencies (stdlib only)
- Tokens are signed with the configured secret
- Signature verified on incoming requests with constant-time comparison
- Expiration (`exp` claim) automatically validated
- Token extracted from `Authorization: Bearer <token>` header

**JWT options:**
- `enabled` - Enable JWT support (default: false)
- `secret` - Secret key for HS256 signing and verification (required if enabled)
- `required` - Require valid JWT on all requests (return 401 if missing/invalid). If false, requests without valid tokens continue with `jwt_claims = nil` (default: false)

**sign_jwt(claims, expires_in)** - Response helper to create signed tokens
- `claims` - Object with token claims (e.g., `{sub = "user123", role = "admin"}`)
- `expires_in` - Token lifetime in seconds (e.g., 3600 for 1 hour, default: 3600)
- Returns: Signed JWT token string

**Request JWT Claims**
- `request().jwt_claims` - Object with verified token claims, or `nil` if no valid token
- Claims are only included if a valid Bearer token is present in Authorization header
- Standard JWT claims like `exp` (expiration) are automatically validated

## Request Context

Inside a handler script, call `context()` to access request data:

```duso
ctx = context()
req = ctx.request()
resp = ctx.response()
```

### Request Object Properties

- `method` - HTTP method (e.g., `"GET"`, `"POST"`)
- `path` - Request path (e.g., `"/api/users"`)
- `headers` - Object with request headers
- `query` - Object with query parameters (from URL `?name=value`)
- `form` - Object with form data (POST/PUT submissions)
- `body` - Request body as raw string
- `params` - Object with path parameters (from route `/users/:id`)
- `jwt_claims` - Object with verified JWT claims (if enabled), or nil

### Accessing Query Parameters

URL: `?name=Alice&age=30`

```duso
ctx = context()
req = ctx.request()

name = req.query.name      // "Alice"
age = req.query.age        // "30"
```

Multiple values: `?tag=js&tag=web`

```duso
tags = req.query.tag       // "js" if one value, ["js", "web"] if multiple
```

### Accessing Path Parameters

Route: `server.route("GET", "/users/:id/tokens/:token")`
Request: `GET /users/123/tokens/abc-xyz`

```duso
ctx = context()
req = ctx.request()

user_id = req.params.id    // "123"
token = req.params.token   // "abc-xyz"
```

### Accessing Form Data

POST/PUT with form submission:

```duso
ctx = context()
req = ctx.request()

username = req.form.username
password = req.form.password

// Multiple values (checkboxes, multi-select)
roles = req.form.roles     // "admin" or ["admin", "user"]
```

### Accessing JSON Body

```duso
ctx = context()
req = ctx.request()

body = parse_json(req.body)
name = body.name
email = body.email
```

### Accessing Headers

```duso
ctx = context()
req = ctx.request()

auth = req.headers["Authorization"]
content_type = req.headers["Content-Type"]
```

### Accessing HTTP Method

```duso
ctx = context()
req = ctx.request()

if req.method == "POST" then
  print("handling POST")
end
```

### Accessing JWT Claims

```duso
ctx = context()
req = ctx.request()

if req.jwt_claims == nil then
  ctx.response().error(401, "Authorization required")
end

user_id = req.jwt_claims.sub
email = req.jwt_claims.email
```

### Complete Handler Example

Route: `server.route("POST", "/api/users/:id")`

```duso
ctx = context()
req = ctx.request()
resp = ctx.response()

// Extract path parameter
user_id = req.params.id

// Parse request body (check content type)
if req.headers["Content-Type"] contains "application/json" then
  data = parse_json(req.body)
else
  data = req.form
end

name = data.name
email = data.email

// Verify JWT authorization
if req.jwt_claims == nil then
  resp.error(401, "Authorization required")
end

// Check user owns this resource
if req.jwt_claims.sub != user_id then
  resp.error(403, "Forbidden")
end

// Update and return response
updated = {id = user_id, name = name, email = email}
resp.json(updated, 200)
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

Routes use exact matching by default. Use `/*` for wildcard (prefix) matching:

```duso
server.route("GET", "/api")        // Matches ONLY /api (exact)
server.route("GET", "/api/users")  // Matches ONLY /api/users (exact)
server.route("GET", "/api/*")      // Matches /api/, /api/users, /api/users/123 (wildcard)
```

When a request matches multiple routes, the most specific route is used:
- Parameterized routes (`:id`) match before wildcards (`/*`)
- Longer exact routes match before shorter ones
- Exact routes match before wildcard routes

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
