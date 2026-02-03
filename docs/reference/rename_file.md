# rename_file()

Rename or move a file.

## Signature

```duso
rename_file(old_path, new_path)
```

## Parameters

- `old_path` (string) - Current file path
- `new_path` (string) - New file path

## Returns

nil

## Examples

Rename a file:

```duso
rename_file("output.txt", "final_output.txt")
```

Move file to different directory:

```duso
rename_file("temp/data.json", "processed/data.json")
```

Add timestamp to file:

```duso
timestamp = format_time(now(), "YYYY-MM-DD-HH-mm-ss")
rename_file("report.pdf", "report_" + timestamp + ".pdf")
```

## See Also

- [move_file() - Move file](/docs/reference/move_file.md)
- [file_exists() - Check if file exists](/docs/reference/file_exists.md)
