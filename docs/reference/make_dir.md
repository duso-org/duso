# make_dir()

Create a directory (including parent directories if needed).

## Signature

```duso
make_dir(path)
```

## Parameters

- `path` (string) - Directory path to create, relative to the script's directory

## Returns

nil

## Examples

Create a single directory:

```duso
make_dir("output")
save("output/result.txt", "data")
```

Create nested directories:

```duso
make_dir("data/processed/2026")
save("data/processed/2026/results.json", format_json(results))
```

Create project structure:

```duso
make_dir("src")
make_dir("tests")
make_dir("docs")
print("Project structure created")
```

## See Also

- [remove_dir() - Remove empty directory](/docs/reference/remove_dir.md)
- [file_exists() - Check if path exists](/docs/reference/file_exists.md)
