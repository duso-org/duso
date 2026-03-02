# decode_base64()

Decode a base64-encoded string back to its original form.

## Signature

```duso
decode_base64(string)
```

## Parameters

- `string` (string) - The base64-encoded string to decode

## Returns

Decoded string

## Examples

Decode a simple base64 string:

```duso
decoded = decode_base64("aGVsbG8gd29ybGQ=")
print(decoded)                  // "hello world"
```

Decode authentication credentials:

```duso
encoded_auth = "dXNlcjpwYXNzd29yZA=="
credentials = decode_base64(encoded_auth)
print(credentials)              // "user:password"
```

Decode data from file:

```duso
encoded_content = load("data.b64")
original = decode_base64(encoded_content)
print(original)
```

Handle decoding errors:

```duso
try
  result = decode_base64("invalid!!!base64")
catch (err)
  print("Decode failed: " + err)
end
```

Round-trip encoding and decoding:

```duso
original = "secret message"
encoded = encode_base64(original)
decoded = decode_base64(encoded)
print(decoded == original)      // true
```

## Notes

- Follows RFC 4648 standard base64 decoding
- Handles padding characters (`=`) automatically
- Throws error if input is not valid base64
- Returns the original string as UTF-8 text

## See Also

- [encode_base64() - Encode to base64](/docs/reference/encode_base64.md)
- [parse_json() - Parse JSON string](/docs/reference/parse_json.md)
- [load() - Read file](/docs/reference/load.md)
