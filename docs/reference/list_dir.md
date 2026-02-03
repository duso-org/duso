# list_dir()

List the contents of a directory.

## Signature

```duso
list_dir(path)
```

## Parameters

- `path` (string) - Directory path relative to the script's directory

## Returns

Array of objects with `name` (string) and `is_dir` (boolean) for each entry

## Examples

List files in current directory:

```duso
entries = list_dir(".")
for entry in entries do
  if entry.is_dir then
    print("  📁 " + entry.name)
  else
    print("  📄 " + entry.name)
  end
end
```

Check for specific files:

```duso
entries = list_dir("src")
has_main = false
for entry in entries do
  if entry.name == "main.du" then
    has_main = true
  end
end
print("Has main.du: " + has_main)
```

## See Also

- [file_exists() - Check if file exists](/docs/reference/file_exists.md)
- [file_type() - Get file type](/docs/reference/file_type.md)
