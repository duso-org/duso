# parse_json()

Parse a JSON string into Duso objects and arrays.

## Signature

```duso
parse_json(string)
```

## Parameters

- `string` (string) - A valid JSON string

## Returns

Parsed Duso value (JSON objects become Duso objects, JSON arrays become Duso arrays, etc.)

## Examples

Parse a JSON object:

```duso
json_str = '{"name": "Alice", "age": 30}'
data = parse_json(json_str)
print(data.name)                // Output: "Alice"
print(data.age)                 // Output: 30
```

Parse a JSON array:

```duso
json_str = '[1, 2, 3, 4, 5]'
numbers = parse_json(json_str)
print(numbers[0])               // Output: 1
print(len(numbers))             // Output: 5
```

Nested structures:

```duso
json_str = '{"users": [{"name": "Alice"}, {"name": "Bob"}]}'
data = parse_json(json_str)
print(data.users[0].name)       // Output: "Alice"
print(data.users[1].name)       // Output: "Bob"
```

With load():

```duso
json_content = load("config.json")
config = parse_json(json_content)
timeout = config.server.timeout
retries = config.server.retries
```

## Common Patterns

Parsing API responses:
```duso
response_json = load("api_response.json")
response = parse_json(response_json)
for user in response.data do
  print(user.id + ": " + user.name)
end
```

Extracting nested values with error handling:
```duso
try
  data = parse_json(json_str)
  value = data.deeply.nested.field
  print(value)
catch (error)
  print("Failed to parse: " + error)
end
```

## See Also

- [format_json() - Convert values to JSON](/docs/reference/format_json.md)
- [load() - Read files](/docs/reference/load.md)
