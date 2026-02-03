# file_exists()

Check if a file or directory exists.

## Signature

```duso
file_exists(path)
```

## Parameters

- `path` (string) - File or directory path to check

## Returns

Boolean: `true` if the path exists, `false` otherwise

## Examples

Check before reading:

```duso
if file_exists("config.json") then
  config = parse_json(load("config.json"))
else
  print("Config file not found")
end
```

Guard file operations:

```duso
if file_exists("output.txt") then
  remove_file("output.txt")
end
save("output.txt", "new data")
```

Conditional processing:

```duso
for file in ["data1.txt", "data2.txt", "data3.txt"] do
  if file_exists(file) then
    content = load(file)
    process(content)
  end
end
```

## See Also

- [file_type() - Get file type](/docs/reference/file_type.md)
- [list_dir() - List directory contents](/docs/reference/list_dir.md)
