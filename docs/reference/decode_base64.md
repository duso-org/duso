# decode_base64()

Decode a base64-encoded string to either string or binary data.

## Signature

```duso
decode_base64(string)
decode_base64(string, type)
decode_base64(string, type="string" | "binary")
```

## Parameters

- `string` (string) - The base64-encoded string to decode
- `type` (string, optional) - Return type: `"string"` or `"binary"`. Defaults to `"string"` if not specified

## Returns

- String if `type="string"` (default)
- Binary data if `type="binary"`

## Examples

Decode base64 to string (default):

```duso
decoded = decode_base64("aGVsbG8gd29ybGQ=")
print(decoded)                  // "hello world"
print(type(decoded))            // "string"
```

Explicitly decode to string:

```duso
text = decode_base64("aGVsbG8gd29ybGQ=", "string")
// or with named parameter
text = decode_base64("aGVsbG8gd29ybGQ=", type="string")
print(text)                     // "hello world"
```

Decode base64 to binary:

```duso
decoded = decode_base64("aGVsbG8gd29ybGQ=", "binary")
// or with named parameter
decoded = decode_base64("aGVsbG8gd29ybGQ=", type="binary")
print(type(decoded))            // "binary"
print(decoded)                  // <binary: 11 bytes>
```

Decode and save as file:

```duso
encoded_data = load("data.b64")
binary_data = decode_base64(encoded_data, "binary")
save_binary(binary_data, "output.dat")
```

Decode image from base64:

```duso
image_b64 = load("photo.b64")
image = decode_base64(image_b64, type="binary")
save_binary(image, "photo.png")
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
original = load_binary("file.dat")
encoded = encode_base64(original)
decoded = decode_base64(encoded, "binary")
print(type(decoded))            // "binary"
```

## Notes

- Follows RFC 4648 standard base64 decoding
- Handles padding characters (`=`) automatically
- Throws error if input is not valid base64
- Defaults to returning string; use `type="binary"` to get binary data
- When returning binary, use with `save_binary()` to write to files
- The `type` parameter accepts `"string"` or `"binary"` only

## See Also

- [encode_base64() - Encode to base64](/docs/reference/encode_base64.md)
- [parse_json() - Parse JSON string](/docs/reference/parse_json.md)
- [load() - Read file](/docs/reference/load.md)
