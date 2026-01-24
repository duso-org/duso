# save()

Write content to a file. Available in `duso` CLI only.

## Signature

```duso
save(filename, content)
```

## Parameters

- `filename` (string) - Path to write to, relative to script directory
- `content` (string) - Content to write

## Returns

`nil`

## Examples

Save text:

```duso
save("output.txt", "Hello, World!")
```

Save JSON:

```duso
data = {name = "Alice", age = 30}
save("data.json", format_json(data, 2))
```

Save processed results:

```duso
input_text = load("input.txt")
processed = upper(input_text)
save("output.txt", processed)
```

## Notes

- Creates parent directories if needed
- Overwrites file if it exists

## See Also

- [load() - Read file](./load.md)
- [format_json() - Convert to JSON](./format_json.md)
- [File I/O](../cli/FILE_IO.md)
