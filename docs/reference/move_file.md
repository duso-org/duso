# move_file()

Move a file from source to destination.

## Signature

```duso
move_file(src, dst)
```

## Parameters

- `src` (string) - Source file path
- `dst` (string) - Destination file path

## Returns

nil

## Details

- Atomically moves file (same semantics as `rename_file`)
- Creates parent directories at destination if needed
- Cannot move from `/EMBED/` (embedded files are read-only)

## Examples

Move file to different directory:

```duso
move_file("temp/data.json", "processed/data.json")
```

Archive old files:

```duso
timestamp = format_time(now(), "YYYY-MM-DD")
move_file("current.log", "archive/log_" + timestamp + ".log")
```

Organize files:

```duso
entries = list_dir("inbox")
for entry in entries do
  if ends_with(entry.name, ".txt") then
    move_file("inbox/" + entry.name, "documents/" + entry.name)
  end
end
```

## See Also

- [rename_file() - Rename file](/docs/reference/rename_file.md)
- [copy_file() - Copy file](/docs/reference/copy_file.md)
- [make_dir() - Create directory](/docs/reference/make_dir.md)
