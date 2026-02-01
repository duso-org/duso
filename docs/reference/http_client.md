# http_client()

Create a stateful HTTP client for making requests. Available in `duso` CLI only.

## Signature

```duso
http_client([options])
```

## Parameters

- `options` (optional, object) - Configuration object with properties:
  - `base_url` (string) - Base URL for relative requests
  - `timeout` (number) - Request timeout in seconds
  - `headers` (object) - Default headers applied to all requests

## Returns

HTTP client object with methods for executing requests

## Methods

### send(request)

Execute an HTTP request. Returns a response object.

**Parameters:**
- `request` (object) - Request configuration with properties:
  - `method` (string) - HTTP method (GET, POST, PUT, DELETE, etc.)
  - `url` (string) - Request URL (absolute or relative to base_url)
  - `body` (string, optional) - Request body
  - `headers` (object, optional) - Additional headers for this request
  - `query` (object, optional) - Query parameters

**Returns:** Response object with properties:
- `status` (number) - HTTP status code
- `body` (string) - Response body
- `headers` (object) - Response headers

### close()

Close idle connections. Call when finished making requests.

## Examples

Basic GET request:

```duso
client = http_client()
response = client.send({
  method = "GET",
  url = "https://api.example.com/data"
})
print(response.body)
```

POST request with body:

```duso
client = http_client()
response = client.send({
  method = "POST",
  url = "https://api.example.com/users",
  body = "name=Alice&age=30",
  headers = {["Content-Type"] = "application/x-www-form-urlencoded"}
})
print(response.status)
```

Reusing client with base_url and headers:

```duso
client = http_client({
  base_url = "https://api.example.com",
  timeout = 30,
  headers = {Authorization = "Bearer token123"}
})

// Multiple requests reuse connection pool and auth header
response1 = client.send({method = "GET", url = "/users/1"})
response2 = client.send({method = "GET", url = "/users/2"})
response3 = client.send({method = "POST", url = "/posts", body = "data"})

client.close()
```

## Notes

- Available only in `duso` CLI (not in HTTP server handlers)
- Requires network access
- `send()` is a low-level primitive that requires manual header and body handling
- For convenient form/JSON submission, use the `require("http")` stdlib module
- Connection pooling: creating one client and reusing it for multiple requests is more efficient than creating a new client for each request

## See Also

- [require() - Load http module for convenience wrappers](/docs/reference/require.md)
