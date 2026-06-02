# HTTP Functions

Duso provides built-in HTTP client and server functionality for building networked applications.

## Client

- [fetch()](/docs/reference/fetch.md) - Make HTTP requests (GET, POST, etc.)

## Server

- [http_server()](/docs/reference/http_server.md) - Create HTTP server for handling requests

## Quick Examples

### Making HTTP Requests

`fetch()` supports all HTTP methods (GET, POST, PUT, DELETE, PATCH, etc.), custom headers, request bodies, timeouts, and automatic response parsing. It's production-ready for API integration, with full control over request/response handling and error conditions.

```duso
// GET request
response = fetch("https://api.github.com/users/github")
print(response.status)

// POST request with JSON
response = fetch("https://httpbin.org/post", {
  method = "POST",
  body = format_json({name = "Alice"})
})
print(response.ok)
```

### Simple HTTP Server

`http_server()` is a production-ready HTTP/HTTPS server with advanced features: routing with path parameters, static file serving, CORS, JWT authentication, WebSocket support, file uploads, resource limits, custom timeouts, and access logging. The example below shows the self-referential pattern where the same script handles both server setup (when `context()` is nil) and request handling (when context is available).

```duso
ctx = context()

if ctx == nil then
  // Server setup: runs once to configure and start the server
  server = http_server({port = 8080})
  server.route("GET", "/")
  server.start()
end

// Handler: runs for each GET / request
ctx.response().text("Hello World")
```
