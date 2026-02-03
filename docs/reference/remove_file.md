# remove_file()

Delete a file.

## Signature

```duso
remove_file(path)
```

## Parameters

- `path` (string) - File path to delete, relative to the script's directory

## Returns

nil

## Examples

Delete a file:

```duso
remove_file("temp.txt")
print("Temp file deleted")
```

Clean up old logs:

```duso
for i = 1, 5 do
  path = "logs/old_" + i + ".log"
  if file_exists(path) then
    remove_file(path)
  end
end
```

## See Also

- [file_exists() - Check if file exists](/docs/reference/file_exists.md)
- [remove_dir() - Remove directory](/docs/reference/remove_dir.md)
