# copy_file()

Copy a file from source to destination.

## Signature

```duso
copy_file(src, dst)
```

## Parameters

- `src` (string) - Source file path (can be from `/EMBED/` for embedded files)
- `dst` (string) - Destination file path

## Returns

nil

## Details

- Supports copying from `/EMBED/` (embedded files in binary)
- Creates parent directories if needed
- Overwrites destination file if it exists

## Examples

Copy file:

```duso
copy_file("template.txt", "output.txt")
```

Copy from embedded templates (like in `-init`):

```duso
copy_file("/EMBED/examples/init/hello/hello.du", "hello.du")
copy_file("/EMBED/examples/init/hello/README.md", "README.md")
```

Create backup:

```duso
if file_exists("data.json") then
  copy_file("data.json", "data.backup.json")
end
```

Duplicate project files:

```duso
copy_file("src/main.du", "src/main.du.bak")
copy_file("config.json", "config.default.json")
```

## See Also

- [move_file() - Move file](/docs/reference/move_file.md)
- [load() - Read file](/docs/reference/load.md)
- [save() - Write file](/docs/reference/save.md)
