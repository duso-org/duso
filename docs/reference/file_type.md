# file_type()

Get the type of a file or directory.

## Signature

```duso
file_type(path)
```

## Parameters

- `path` (string) - File or directory path

## Returns

String: either `"file"` or `"directory"`

## Examples

Check file type:

```duso
type = file_type("data.txt")
if type == "file" then
  print("It's a file")
end
```

Filter directories from listing:

```duso
entries = list_dir(".")
for entry in entries do
  if entry.is_dir then
    print("Directory: " + entry.name)
  end
end
```

Process only files:

```duso
entries = list_dir("src")
for entry in entries do
  if file_type("src/" + entry.name) == "file" then
    print("Processing: " + entry.name)
  end
end
```

## See Also

- [list_dir() - List directory contents](/docs/reference/list_dir.md)
- [file_exists() - Check if path exists](/docs/reference/file_exists.md)
