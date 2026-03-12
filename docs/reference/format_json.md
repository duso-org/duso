# format_json()

Convert a Duso value to a JSON string.

## Signature

```duso
format_json(value [, indent])
```

## Parameters

- `value` - Any Duso value to convert
- `indent` (optional, number) - Number of spaces for indentation. 0 for compact, 2 or 4 for pretty-printed

## Returns

JSON string

## Examples

Compact JSON:

```duso
data = {name = "Alice", age = 30}
json = format_json(data)
print(json)                     // {"name":"Alice","age":30}
```

Pretty-printed JSON:

```duso
data = {name = "Alice", skills = ["Go", "Duso"]}
json = format_json(data, 2)
print(json)
```

JSON with binary data:

```duso
image = load_binary("photo.png")
data = {file = image, size = 1024}
json = format_json(data)
print(json)                     // {"file":"<binary: photo.png (1024 bytes)>","size":1024}
```

Save to file:

```duso
config = {timeout = 30, retries = 3}
save("config.json", format_json(config, 2))
```

## Serialization of Non-JSON Types

Some Duso types don't have direct JSON equivalents and are stringified:

- **Binary data**: Stringified as `<binary: filename (size)>` - use `encode_base64()` to encode binary data for transmission
- **Functions**: Stringified as `<function>` - functions cannot be serialized to JSON
- **Errors**: Stringified as `<error: message>` - error values are converted to string representation
- **Code values**: Stringified as their source code text

## Notes

- JSON does not support binary data directly. To transmit binary in JSON, use `encode_base64()` first
- Functions and errors are stringified for debugging, but don't carry their full semantics
- To preserve binary data with metadata, consider creating an object with encoded data and metadata fields

## See Also

- [parse_json() - Parse JSON string](/docs/reference/parse_json.md)
- [encode_base64() - Encode to base64](/docs/reference/encode_base64.md)
- [save() - Write file](/docs/reference/save.md)
