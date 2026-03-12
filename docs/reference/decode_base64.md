# decode_base64()

Decode a base64-encoded string to binary data.

## Signature

```duso
decode_base64(string)
```

## Parameters

- `string` (string) - The base64-encoded string to decode

## Returns

Binary data

## Examples

Decode base64 to binary:

```duso
decoded = decode_base64("aGVsbG8gd29ybGQ=")
print(type(decoded))            // "binary"
print(decoded)                  // <binary: 11 bytes>
```

Decode and save as file:

```duso
encoded_data = load("data.b64")
binary_data = decode_base64(encoded_data)
save_binary(binary_data, "output.dat")
```

Decode image from base64:

```duso
image_b64 = load("photo.b64")
image = decode_base64(image_b64)
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
decoded = decode_base64(encoded)
print(type(decoded))            // "binary"
```

## Notes

- Follows RFC 4648 standard base64 decoding
- Handles padding characters (`=`) automatically
- Throws error if input is not valid base64
- Returns binary data (use with `save_binary()` to write to files)

## See Also

- [encode_base64() - Encode to base64](/docs/reference/encode_base64.md)
- [parse_json() - Parse JSON string](/docs/reference/parse_json.md)
- [load() - Read file](/docs/reference/load.md)
