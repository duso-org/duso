# HTTP Module

HTTP client for making requests and managing connections.

**Status**: Stable (v1.0)
**Requires**: Duso 0.1.0+

## Overview

The HTTP module provides two interfaces for making HTTP requests:

- **Simple**: `http.fetch()` for quick one-liners
- **Stateful**: `http.client()` for reusable clients with shared config

Both are wrappers around the built-in `http_client()` Go function, which provides the low-level power layer.

## Quick Start

```duso
http = require("http")

-- Simple request
response = http.fetch("https://api.example.com/data")
data = parse_json(response)

-- Reusable client
client = http.client({base_url = "https://api.example.com"})
resp1 = client.get("/users")
resp2 = client.get("/posts")
client.close()
```

## Simple Interface: fetch()

Make a single HTTP request and get the response body.

```duso
body = http.fetch(url [, options])
```

**Parameters:**
- `url` (string) - Full URL to request
- `options` (object, optional):
  - `method` (string) - HTTP method, default "GET"
  - `body` (string) - Request body
  - `headers` (object) - Custom headers
  - `timeout` (number) - Timeout in seconds

**Returns:**
- `string` - Response body

**Examples:**

```duso
http = require("http")

-- GET request
body = http.fetch("https://api.example.com/data")
print(body)

-- POST request with JSON
body = http.fetch("https://api.example.com/submit", {
  method = "POST",
  body = format_json({name = "Alice", age = 30}),
  headers = {["Content-Type"] = "application/json"}
})

-- Custom timeout
body = http.fetch("https://slow-api.example.com/data", {
  timeout = 60  -- 60 seconds
})
```

## Stateful Interface: client()

Create a reusable HTTP client with shared configuration.

```duso
client_obj = http.client([config])
```

**Configuration:**
- `base_url` (string) - Base URL for relative requests
- `timeout` (number) - Timeout in seconds for all requests
- `headers` (object) - Default headers for all requests
- `auth` (string) - Authorization header (convenience, sets Authorization header)

**Returns:**
- Client object with methods: `get()`, `post()`, `put()`, `delete()`, `send()`, `close()`

### Client Methods

#### get(path [, options])

```duso
response = client.get("/data", {
  headers = {Accept = "application/json"},
  query = {page = 1}  -- Currently passed through, see limitations
})
```

**Returns:** Response object `{status, body, headers}`

#### post(path, body [, options])

```duso
response = client.post("/submit",
  format_json({name = "Alice"}),
  {headers = {["Content-Type"] = "application/json"}}
)
```

**Returns:** Response object

#### put(path, body [, options])

```duso
response = client.put("/items/123",
  format_json({name = "Updated Alice"})
)
```

**Returns:** Response object

#### delete(path [, options])

```duso
response = client.delete("/items/123")
```

**Returns:** Response object

#### send(request)

Direct access to the underlying HTTP client for custom requests.

```duso
response = client.send({
  method = "PATCH",
  url = "/items/123",
  body = format_json({status = "active"}),
  headers = {["X-Custom"] = "value"}
})
```

**Request object fields:**
- `method` (string) - HTTP method (GET, POST, PUT, DELETE, etc)
- `url` (string) - Request URL (relative or absolute)
- `body` (string, optional) - Request body
- `headers` (object, optional) - Request headers

**Response object fields:**
- `status` (number) - HTTP status code (e.g., 200, 404)
- `body` (string) - Response body
- `headers` (object) - Response headers (lowercase keys)

#### close()

Close idle connections and cleanup.

```duso
client.close()
```

## Examples

### Basic GET with JSON Parsing

```duso
http = require("http")

response = http.fetch("https://jsonplaceholder.typicode.com/posts/1")
data = parse_json(response)

print("Title: " + data.title)
print("User ID: " + data.userId)
```

### Authenticated API Client

```duso
http = require("http")

client = http.client({
  base_url = "https://api.example.com",
  auth = "Bearer your-api-key-here",
  timeout = 30
})

-- Make multiple requests with shared auth
users = parse_json(client.get("/users").body)
posts = parse_json(client.get("/posts").body)

client.close()
```

### Error Handling

```duso
http = require("http")

try
  response = http.fetch("https://api.example.com/data")
  data = parse_json(response)

  if response.status ~= 200 then
    print("Request failed with status: " + response.status)
  end
catch (e)
  print("Error: " + e)
end
```

Wait, response from fetch() is just the body (string), not the full response object. For full control, use client:

```duso
http = require("http")

try
  client = http.client()
  response = client.send({
    method = "GET",
    url = "https://api.example.com/data"
  })

  if response.status == 200 then
    data = parse_json(response.body)
    print(data)
  else
    print("Error: " + response.status)
  end
catch (e)
  print("Request failed: " + e)
end
```

### Multi-Step Workflow

```duso
http = require("http")

client = http.client({
  base_url = "https://api.example.com",
  headers = {["Content-Type"] = "application/json"}
})

-- Create a resource
createResp = client.post("/items",
  format_json({name = "New Item"})
)
createdItem = parse_json(createResp.body)

-- Get the resource
getResp = client.get("/items/" + createdItem.id)
retrieved = parse_json(getResp.body)

-- Update it
updateResp = client.put("/items/" + createdItem.id,
  format_json({name = "Updated Item"})
)

-- Delete it
deleteResp = client.delete("/items/" + createdItem.id)

print("Status: " + deleteResp.status)

client.close()
```

## Design Rationale

### Why Two Interfaces?

**fetch()** handles the 90% case: simple requests where you just need the body.

**client()** handles the 10% case: complex workflows where you need to:
- Reuse connection config
- Check status codes
- Access response headers
- Manage multiple requests efficiently

Both are equally valid. Choose based on your use case.

### Why Wrap http_client()?

The HTTP module is implemented as a Duso wrapper around the built-in `http_client()` Go function. This design provides:

1. **Community ownership** - The Duso module can evolve without requiring Go changes
2. **Rapid iteration** - Add features, fix bugs in Duso code (no recompile)
3. **Minimal overhead** - Network latency dominates, Duso overhead is negligible
4. **Power access** - Advanced users can call `http_client()` directly for ultimate control

This is the Duso philosophy: Go provides the engine, Duso provides the API surface.

### Performance Considerations

- Connections are reused by the underlying Go `net/http.Client`
- Each `http.client()` creates a new client pool - reuse the same client for multiple requests
- `fetch()` creates a new client for each request - fine for occasional use, but not tight loops
- For performance-critical code with many requests, use `http.client()`

### Limitations

- Query parameters are passed through to the request object but not automatically encoded
- Use client.send() with full URL for complex requests
- SSL verification is always enabled (can't be disabled for security reasons)

## See Also

- [Duso Language Reference](../../docs/language-spec.md)
- [File I/O Guide](../../docs/cli/FILE_IO.md)
- [Claude Integration](../../docs/cli/CLAUDE_INTEGRATION.md)
