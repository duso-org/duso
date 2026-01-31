# http_client()

Create an HTTP client for making requests. Available in `duso` CLI only.

## Signature

```duso
http_client([options])
```

## Parameters

- `options` (optional, object) - Configuration object

## Returns

HTTP client object with methods

## Methods

- `fetch(url)` - Make GET request
- `post(url, body)` - Make POST request
- `get(url)` - Make GET request
- `request(method, url [, body])` - Make custom request

## Examples

Simple GET request:

```duso
http = http_client()
response = http.fetch("https://api.example.com/data")
data = parse_json(response)
```

POST request:

```duso
http = http_client()
payload = format_json({name = "Alice", age = 30})
response = http.post("https://api.example.com/users", payload)
```

## Notes

- Low-level HTTP functionality
- For higher-level HTTP operations, use stdlib http module
- Requires network access

## See Also

- [require() - Load http module](/docs/reference/require.md)
- [parse_json() - Parse response](/docs/reference/parse_json.md)
- [format_json() - Format request body](/docs/reference/format_json.md)
