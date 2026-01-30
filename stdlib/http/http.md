# HTTP Module

HTTP client for making requests and managing connections.

## Quick Start

```duso
http = require("http")

// Simple request
response = http.fetch("https://api.example.com/data")
data = parse_json(response)

// Reusable client
client = http.client({base_url = "https://api.example.com"})
resp1 = client.get("/users")
resp2 = client.get("/posts")
client.close()
```

## http.fetch()

Make a single HTTP request and get the response body.

### Signature

```duso
http.fetch(url [, options])
```

### Parameters

- `url` (string) - Full URL to request
- `options` (optional, object):
  - `method` (string) - HTTP method, default "GET"
  - `body` (string) - Request body
  - `headers` (object) - Custom headers
  - `timeout` (number) - Timeout in seconds

### Returns

String containing response body

### Examples

GET request:

```duso
body = http.fetch("https://api.example.com/data")
print(body)
```

POST request with JSON:

```duso
body = http.fetch("https://api.example.com/submit", {
  method = "POST",
  body = format_json({name = "Alice", age = 30}),
  headers = {["Content-Type"] = "application/json"}
})
```

Custom timeout:

```duso
body = http.fetch("https://slow-api.example.com/data", {
  timeout = 60
})
```

## http.client()

Create a reusable HTTP client with shared configuration.

### Signature

```duso
client = http.client([config])
```

### Configuration

- `base_url` (string) - Base URL for relative requests
- `timeout` (number) - Timeout in seconds for all requests
- `headers` (object) - Default headers for all requests
- `auth` (string) - Authorization header value

### Returns

Client object

### Client Methods

#### get(path [, options])

Make GET request.

```duso
response = client.get("/data", {headers = {Accept = "application/json"}})
```

Returns response object `{status, body, headers}`

#### post(path, body [, options])

Make POST request.

```duso
response = client.post("/submit", format_json({name = "Alice"}), {
  headers = {["Content-Type"] = "application/json"}
})
```

Returns response object

#### put(path, body [, options])

Make PUT request.

```duso
response = client.put("/items/123", format_json({name = "Updated Alice"}))
```

Returns response object

#### delete(path [, options])

Make DELETE request.

```duso
response = client.delete("/items/123")
```

Returns response object

#### send(request)

Make custom HTTP request with full control.

```duso
response = client.send({
  method = "PATCH",
  url = "/items/123",
  body = format_json({status = "active"}),
  headers = {["X-Custom"] = "value"}
})
```

Request fields:
- `method` (string) - HTTP method
- `url` (string) - URL (relative or absolute)
- `body` (optional, string) - Request body
- `headers` (optional, object) - Request headers

Response fields:
- `status` (number) - HTTP status code
- `body` (string) - Response body
- `headers` (object) - Response headers

#### close()

Close idle connections.

```duso
client.close()
```

## Examples

Basic GET with JSON parsing:

```duso
http = require("http")
body = http.fetch("https://jsonplaceholder.typicode.com/posts/1")
data = parse_json(body)
print("Title: " + data.title)
print("User ID: " + data.userId)
```

Authenticated API client:

```duso
http = require("http")
client = http.client({
  base_url = "https://api.example.com",
  auth = "Bearer your-api-key-here",
  timeout = 30
})

users = parse_json(client.get("/users").body)
posts = parse_json(client.get("/posts").body)
client.close()
```

Error handling:

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

Multi-step workflow:

```duso
http = require("http")
client = http.client({
  base_url = "https://api.example.com",
  headers = {["Content-Type"] = "application/json"}
})

// Create
createResp = client.post("/items", format_json({name = "New Item"}))
createdItem = parse_json(createResp.body)

// Retrieve
getResp = client.get("/items/" + createdItem.id)

// Update
updateResp = client.put("/items/" + createdItem.id, format_json({name = "Updated"}))

// Delete
deleteResp = client.delete("/items/" + createdItem.id)
print("Status: " + deleteResp.status)

client.close()
```

## See Also

- [format_json() - Convert to JSON](/docs/reference/format_json.md)
- [parse_json() - Parse JSON](/docs/reference/parse_json.md)
