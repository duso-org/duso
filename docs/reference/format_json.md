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

Save to file:

```duso
config = {timeout = 30, retries = 3}
save("config.json", format_json(config, 2))
```

## See Also

- [parse_json() - Parse JSON string](./parse_json.md)
- [save() - Write file](./save.md)
- [JSON functions](../language-spec.md#json-functions)
