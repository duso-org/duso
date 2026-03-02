# encode_base64()

Encode a string to base64.

## Signature

```duso
encode_base64(string)
```

## Parameters

- `string` (string) - The string to encode

## Returns

Base64-encoded string

## Examples

Encode a simple string:

```duso
encoded = encode_base64("hello world")
print(encoded)                  // "aGVsbG8gd29ybGQ="
```

Encode authentication credentials:

```duso
api_key = "sk_test_123"
credentials = encode_base64(api_key + ":")
auth_header = "Basic " + credentials
```

Encode for HTTP requests:

```duso
data = "user:password"
response = fetch("https://api.example.com/data", {
  method = "GET",
  headers = {
    "Authorization" = "Basic " + encode_base64(data)
  }
})
```

Save encoded data to file:

```duso
content = "sensitive information"
encoded = encode_base64(content)
save("data.b64", encoded)
```

## Notes

- Follows RFC 4648 standard base64 encoding
- Output includes padding characters (`=`) as needed
- Works with any string input
- Non-string inputs are converted to strings first

## See Also

- [decode_base64() - Decode base64 string](/docs/reference/decode_base64.md)
- [format_json() - Convert to JSON](/docs/reference/format_json.md)
- [fetch() - Make HTTP requests](/docs/reference/fetch.md)
