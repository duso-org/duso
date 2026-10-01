# fetch()

Make HTTP requests using a JavaScript-style fetch API. Returns immediately (synchronous from script perspective, but Go handles async HTTP internally).

`fetch(url [, options])`

## Parameters

- `url` (string) - URL to request
- `options` (optional, object) - Request configuration with properties:
  - `method` (string) - HTTP method (GET, POST, PUT, DELETE, etc.), default "GET"
  - `headers` (object) - Request headers
  - `body` (string) - Request body
  - `timeout` (number) - Request timeout in seconds
  - `follow_redirects` (boolean) - Follow 3xx redirects, default `true`. Set to `false` to get the redirect response itself, with its `Location` and `Set-Cookie` headers

## Returns

Response object with properties and methods:
- `status` (number) - HTTP status code
- `proto` (string) - HTTP protocol version of the response (e.g., `"HTTP/1.1"`, `"HTTP/2.0"`)
- `ok` (boolean) - true if status < 400
- `body` (string) - Response body as string
- `headers` (object) - Response headers
- `json()` - Method to parse body as JSON, returns parsed object/array
- `text()` - Method to return body as string (same as .body)

## Examples

Simple GET request:

```duso
response = fetch("https://api.example.com/data")
if response.ok then
  print("Success: " + response.status)
end
```

POST request with JSON:

```duso
data = {name = "Alice", age = 30}
response = fetch("https://api.example.com/users", {
  method = "POST",
  headers = {"Content-Type" = "application/json"},
  body = format_json(data)
})

if response.ok then
  result = response.json()
  print("Created user: " + result.id)
else
  print("Error: " + response.status)
end
```

Using timeout:

```duso
response = fetch("https://slow-api.example.com/data", {
  timeout = 10
})
```

Handling errors:

```duso
response = fetch("https://api.example.com/users/999")
if not response.ok then
  print("Error: " + response.status)
  print("Response: " + response.body)
else
  data = response.json()
  print(data)
end
```

Multi-step workflow:

```duso
// Create
create_resp = fetch("https://api.example.com/items", {
  method = "POST",
  headers = {"Content-Type" = "application/json"},
  body = format_json({name = "New Item"})
})

created = create_resp.json()

// Retrieve
get_resp = fetch("https://api.example.com/items/" + created.id)
item = get_resp.json()

// Update
update_resp = fetch("https://api.example.com/items/" + created.id, {
  method = "PUT",
  headers = {"Content-Type" = "application/json"},
  body = format_json({name = "Updated Item"})
})

// Delete
delete_resp = fetch("https://api.example.com/items/" + created.id, {
  method = "DELETE"
})

print("Delete status: " + delete_resp.status)
```

Inspect a redirect instead of following it:

```duso
r = fetch("https://example.com/login", {
  method = "POST",
  body = "user=alice&pass=secret",
  follow_redirects = false
})
print(r.status)                    // 302
print(r.headers.Location)          // "/dashboard"
print(r.headers["Set-Cookie"])     // session cookie(s)
```

## Notes

- Requires network access
- HTTP/2 is negotiated automatically over TLS (`https://` URLs) whenever the server supports it — no configuration needed. Check `response.proto` to see which protocol was used. Plaintext HTTP/2 (h2c) is not supported; `http://` requests always use HTTP/1.1
- Connection pooling is handled automatically by Go's HTTP client
- Response status codes include all HTTP statuses (2xx, 3xx, 4xx, 5xx). Check `.ok` or `.status` to determine success
- Redirects are followed automatically (up to 10) and no cookie jar is kept, so cookies set by an intermediate redirect are not seen or resent. Use `follow_redirects = false` to inspect a redirect: `.status` is the 3xx code and `.ok` is `true` (it means status < 400). Multiple `Set-Cookie` headers come back as an array
- Headers are case-sensitive in the response object (Go normalizes them to canonical form)
- For header names with special characters, use quoted keys: `headers = {["Content-Type"] = "application/json"}`

## See Also

- [format_json() - Convert to JSON](/docs/reference/format_json.md)
- [parse_json() - Parse JSON](/docs/reference/parse_json.md)
- [http_server() - Create HTTP server](/docs/reference/http_server.md)
