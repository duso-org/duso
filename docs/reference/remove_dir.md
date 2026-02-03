# remove_dir()

Remove an empty directory.

## Signature

```duso
remove_dir(path)
```

## Parameters

- `path` (string) - Directory path to remove (must be empty), relative to the script's directory

## Returns

nil

## Examples

Remove empty directory:

```duso
make_dir("temp")
// ... do work ...
remove_dir("temp")
```

Clean up after processing:

```duso
if file_exists("staging/work") then
  remove_dir("staging/work")
  print("Staging directory removed")
end
```

## See Also

- [make_dir() - Create directory](/docs/reference/make_dir.md)
- [remove_file() - Remove file](/docs/reference/remove_file.md)
