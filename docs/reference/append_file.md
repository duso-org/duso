# append_file()

Append content to a file (create if doesn't exist).

## Signature

```duso
append_file(path, content)
```

## Parameters

- `path` (string) - File path (relative to script directory)
- `content` (string) - Content to append

## Returns

nil

## Examples

Write logs:

```duso
timestamp = format_time(now(), "iso")
append_file("debug.log", timestamp + " - Script started\n")
// ... do work ...
append_file("debug.log", timestamp + " - Script completed\n")
```

Build report line by line:

```duso
append_file("report.txt", "# Analysis Report\n\n")
append_file("report.txt", "Generated: " + format_time(now(), "date") + "\n\n")
append_file("report.txt", "Results:\n")

for item in items do
  append_file("report.txt", "  - " + item + "\n")
end
```

Accumulate data:

```duso
for i = 1, 100 do
  result = compute(i)
  append_file("results.csv", i + "," + result + "\n")
end
```

## See Also

- [save() - Write/overwrite file](/docs/reference/save.md)
- [load() - Read file](/docs/reference/load.md)
