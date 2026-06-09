# save()

Write content to a file. Available in `duso` CLI only.


`save(filename, content)`

```

## Parameters

- `filename` (string) - Path to write to. See [Files, Modules, and Paths](/docs/files-and-modules.md#path-roots) for how paths are resolved.
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

- [Files, Modules, and Paths](/docs/files-and-modules.md) - Path roots, file operations overview
- [load() - Read file](/docs/reference/load.md)
- [format_json() - Convert to JSON](/docs/reference/format_json.md)
