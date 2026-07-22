# http_server()

Create an HTTP server that listens for incoming requests and runs handler scripts. Available in `duso` CLI only.

`http_server([config])`

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
  - `max_websocket_connections` (number) - Max concurrent WebSocket connections (default: 0 = unlimited). Returns 503 if exceeded.
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
    - `secret` (string) - Secret key for HS256 signing/verification (optional, required for HS256)
    - `rs256_private_key` (string) - PEM-encoded RSA private key for RS256 signing (optional)
    - `rs256_public_key` (string) - PEM-encoded RSA public key for RS256 verification (optional)
    - `required` (boolean) - Require valid JWT token for all requests (default: false)
  - `uploads` (object) - File upload configuration (optional):
    - `enabled` (boolean) - Enable file uploads via multipart/form-data (default: false)
    - `max_size` (number) - Max file size in KB per uploaded file (default: 10240 = 10MB)
    - `timeout` (number) - Upload timeout in seconds (reserved for future use)
  - `websocket` (object) - WebSocket configuration (optional). See [websocket()](/docs/reference/websocket.md) for all options including `idle_timeout`, `max_message_size`, `max_messages_per_second`, and queue sizes.

## Returns

HTTP server object with methods

## Methods

- `route(method, path [, handler])` - Register a route with a handler script
  - `method` - HTTP method: string, array of strings, `"*"`, or `nil` for all methods
    - Valid methods: `"GET"`, `"POST"`, `"PUT"`, `"DELETE"`, `"PATCH"`, `"HEAD"`, `"OPTIONS"`, `"TRACE"`, `"CONNECT"`, `"WS"` (WebSocket)
    - Case-insensitive (e.g., `"get"`, `"Get"`, `"GET"` all work)
    - Array example: `["GET", "POST"]` to handle both methods on same path
    - `"WS"` for WebSocket upgrade requests (see WebSocket section below)
    - `"*"` or `nil` matches all HTTP methods (but not WebSocket)
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

## JWT Authentication

The server supports both **HS256** (HMAC) and **RS256** (RSA) JWT algorithms, allowing flexible authentication scenarios:

- **HS256**: Symmetric key (secret string) for local/internal authentication
- **RS256**: Asymmetric keys (RSA public/private) for sharing with partners or remote servers

### Verification

Call `verify_jwt([options])` on the request object to verify JWT tokens. The function auto-detects the algorithm from the token header and uses configured defaults, with optional overrides:

```duso
ctx = context()
req = ctx.request()

// Verify with configured defaults
claims = req.verify_jwt()
if claims == nil then
  exit({status = 401})
end

// Verify with custom key (for partner tokens)
claims = req.verify_jwt({
  public_key = load("partner-public.pem")
})

// Or custom HS256 secret
claims = req.verify_jwt({
  secret = "alternative-secret"
})
```

Returns the claims map if valid, or `nil` if verification fails (invalid signature, expired, etc).

### Signing

Use `sign_jwt(claims [, options])` in handlers to create tokens:

```duso
// Sign with HS256 (default)
token = sign_jwt({user_id = 123})

// Sign with RS256
token = sign_jwt({user_id = 123}, {algorithm = "RS256"})

// Override key or set expiration
token = sign_jwt({user_id = 123}, {
  algorithm = "RS256",
  private_key = alternative_key,
  expires_in = 7200
})
```

### Configuration Examples

**HS256 only** (internal use):
```duso
jwt = {
  enabled = true,
  secret = "your-secret-key",
  required = false
}
```

**Both HS256 and RS256** (mixed authentication):
```duso
private_key = load("keys/private.pem")
public_key = load("keys/public.pem")

jwt = {
  enabled = true,
  secret = "hs256-secret",
  rs256_private_key = private_key,
  rs256_public_key = public_key
}
```

The developer controls key management (loading, caching, rotation) by passing PEM-encoded key strings.

## Examples

Quick static file server (one-liner):

```bash
duso eval 'http_server().start()'
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
ct = req.headers["Content-Type"]
if ct and contains(ct, "application/json") then
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

// Send binary data (image from datastore, etc)
image = datastore("images").get("avatar_" + user_id)
resp.binary(image, "image/png", 200)

// Generic response with custom headers
resp.response("Custom body", 200, {"X-Custom" = "Header"})
```

**Response wrapper methods:**
- `json(data, [status])` - Send JSON response with `Content-Type: application/json`
- `text(data, [status])` - Send plain text with `Content-Type: text/plain`
- `html(data, [status])` - Send HTML with `Content-Type: text/html`
- `binary(data, content_type, [status])` - Send binary data (images, archives, etc.) with custom Content-Type
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

## WebSocket

Duso supports WebSocket connections for real-time bidirectional communication. Use the `"WS"` method with `route()` to register WebSocket endpoints.

### Registering a WebSocket Route

```duso
server = http_server({port = 8080})
server.route("WS", "/ws", "handlers/ws.du")
server.route("WS", "/game/:id", "handlers/game.du")
server.start()
```

WebSocket routes work like HTTP routes but with different semantics:
- The HTTP upgrade request can have any HTTP method (typically GET)
- The server upgrades the connection to WebSocket automatically
- Each connection spawns a persistent handler session that runs until the client disconnects

### WebSocket Handler Pattern

Unlike HTTP handlers that respond to single requests, WebSocket handlers manage a persistent connection:

```duso
ctx = context()
conn = ctx.connection()
req = ctx.request()

// Access request data (headers, path params, JWT claims)
if req.jwt_claims then
  user_id = req.jwt_claims.sub
end

conn.accept()  // Accept the WebSocket connection

// Message loop: blocks until message received, timeout, or disconnect
while true do
  msg = conn.read(timeout=30)  // Wait up to 30 seconds for message

  if msg == nil then    // nil means client disconnected or timeout
    break
  end

  // Process message
  response = "Echo: " + msg
  conn.write(response)   // Send message to client
end
```

### Connection Methods

The `ctx.connection()` object provides WebSocket methods:

- `accept()` - Accept the WebSocket connection (complete the upgrade)
- `read([timeout])` - Block until a message is received. Returns `nil` on disconnect or timeout.
  - `timeout` (optional, number) - Wait timeout in seconds. If omitted, uses `default_read_timeout` from config.
  - Supports both positional: `read(5)` and named: `read(timeout=5)` arguments.
- `write(message)` - Queue a message to the connected client. Returns bytes queued (number) or `nil` if queue is full.
- `close()` - Explicitly close the WebSocket connection
- `is_connected()` - Check if connection is still open (returns boolean)
- `id` - Unique connection ID (string). Use with `send_websocket()` for broadcasting.

### Message Queuing

Incoming and outgoing messages are queued in buffers to handle backpressure:

- **Read queue**: Incoming messages from the client are buffered. Default: 100 messages.
- **Write queue**: Outgoing messages to the client are buffered. Default: 100 messages.

If `write()` returns `nil`, the write queue is full—the client is receiving slower than you're sending. Handle gracefully:

```duso
bytes = conn.write(msg)
if bytes == nil then
  // Queue overflow - client too slow
  print("Failed to send to client")
  conn.close()
end
```

Configure WebSocket per server (all options optional, defaults provided):

```duso
server = http_server({
  port = 8080,
  max_websocket_connections = 1000,  // Server-level limit
  websocket = {
    read_queue_size = 100,
    write_queue_size = 100,
    read_timeout = 30,
    idle_timeout = 300,                // Disconnect after 5 min idle (0 = disabled)
    max_message_size = 65536,          // 64KB (0 = unlimited)
    max_messages_per_second = 0        // Disabled by default
  }
})
```

### Broadcasting with send_websocket()

Send messages to any connection by ID without direct access to the connection object:

```duso
// Inside handler or anywhere else
bytes = send_websocket(conn.id, "broadcast message")
if bytes == nil then
  print("Failed to queue broadcast")
end
```

Use with `datastore()` to coordinate across handlers:

```duso
// handlers/chat.du
ctx = context()
conn = ctx.connection()
conn.accept()

store = datastore("chat_room")
store.push("connections", conn.id)  // Register this connection

while true do
  msg = conn.read()
  if msg == nil then break end
  
  // Broadcast to all connections
  all_conns = store.get("connections")
  for cid in all_conns do
    send_websocket(cid, "Message: " + msg)
  end
end
```

### Request Context

WebSocket handlers have access to the same request context as HTTP handlers:

```duso
ctx = context()
conn = ctx.connection()
req = ctx.request()

// Full request access from the WebSocket upgrade request
method = req.method          // Always "GET" for WebSocket upgrades
path = req.path              // "/ws" or "/game/123" etc.
params = req.params          // Path parameters (:id from route pattern)
headers = req.headers        // All HTTP headers from upgrade request
jwt_claims = req.jwt_claims  // Verified JWT claims (if JWT enabled)
```

### Example: Chat Server

```duso
// server.du
ctx = context()

if ctx == nil then
  server = http_server({port = 8080})
  server.route("WS", "/chat", "handlers/chat.du")
  server.start()
end

// handlers/chat.du - Chat handler (runs for each connection)
ctx = context()
conn = ctx.connection()

conn.accept()
print("Client connected")

while true do
  msg = conn.read()
  if msg == nil then
    print("Client disconnected")
    break
  end

  // Broadcast to all connected clients would require coordination
  // For now, just echo back
  conn.write("You said: " + msg)
end
```

### Lifecycle

Each WebSocket connection:
1. **Upgrade Request** - Client sends HTTP upgrade request to registered path
2. **Handler Spawn** - Server spawns handler script in a new goroutine
3. **Accept** - Handler calls `conn.accept()` to complete the upgrade
4. **Message Loop** - Handler receives/sends messages until disconnect
5. **Cleanup** - Handler exits when client disconnects (receive returns nil)

### Notes on WebSocket

- Each connection lives until the client disconnects
- No automatic timeout (uses socket-level `idle_timeout` if configured)
- Handler script runs in its own goroutine (no blocking issues with other requests)
- Request context (headers, JWT, params) is from the initial upgrade request
- Each connection has a unique ID (`conn.id`) that can be used with `send_websocket()` for broadcasting
- For coordination between multiple WebSocket connections, use `datastore()` to store connection IDs
- WebSocket upgrade requests cannot be matched with `"*"` (all methods) - must explicitly register as `"WS"`
- Messages are queued (not dropped) to handle slow clients, but queues have limits to prevent memory exhaustion

## File Uploads

Enable multipart/form-data file uploads with the `uploads` configuration:

```duso
server = http_server({
  port = 3000,
  uploads = {
    enabled = true,
    max_size = 10240  // 10MB per file (in KB)
  }
})

server.route("POST", "/upload", "upload_handler.du")
server.start()
```

In handler scripts, access uploaded files via `req.files`:

```duso
ctx = context()
req = ctx.request()
res = ctx.response()

// Access single file
file = req.files.avatar
if file then
  print(file.filename)      // "avatar.png"
  print(file.content_type)  // "image/png"
  print(file.size)          // bytes
  
  if type(file.data) == "binary" then
    save_binary(file.data, "/STORE/uploads/" + file.filename)
  elseif type(file.data) == "string" then
    // Text files (JSON, XML, etc.)
    parsed = parse_json(file.data)
  end
end

// Multiple files on same field
for f in req.files.attachments do
  print(f.filename)
end
```

### MIME Type Handling

- **Text MIME types** (`text/*`, `application/json`, `application/xml`, etc.) → `file.data` is a string
- **Binary MIME types** (images, archives, etc.) → `file.data` is a binary value

### File Object Properties

- `data` - File content (binary or string)
- `filename` - Original filename from client
- `content_type` - MIME type (detected from upload header or file extension)
- `size` - File size in bytes

### Limits and Errors

- Uploads are disabled by default (set `enabled = true` to activate)
- Max file size is enforced; oversized files are ignored (silently skipped)
- Form field limits (`max_form_fields`) also apply to file uploads
- If no files are uploaded, `req.files` is an empty map (never nil)

## See Also

- [context() - Access request context in handlers](/docs/reference/context.md)
- [fetch() - Make HTTP requests](/docs/reference/fetch.md)
