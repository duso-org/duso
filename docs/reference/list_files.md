# list_files()

List files matching a wildcard pattern.

## Signature

```duso
list_files(pattern)
```

## Parameters

- `pattern` (string) - Wildcard pattern to match files, relative to the script's directory

## Returns

Array of file paths (strings) that match the pattern. Returns empty array if no files match.

## Details

- Supports wildcard patterns with `*` (match any characters) and `?` (match single character)
- Does not support `**` (recursive wildcard)
- Pattern is matched against file names
- Returns relative paths if the input pattern was relative, absolute paths if pattern was absolute
- Works with regular filesystem, `/EMBED/` (embedded files), and `/STORE/` (virtual filesystem)
- Use `list_dir()` for plain directory listing without pattern matching

## Examples

List all Duso scripts:

```duso
scripts = list_files("*.du")
for script in scripts do
  print(script)
end
```

Find backup files:

```duso
backups = list_files("*.bak")
print("Found " + len(backups) + " backup files")
```

Match single character:

```duso
temp_files = list_files("temp_?.txt")
// Matches: temp_1.txt, temp_2.txt, temp_a.txt
// Does not match: temp_10.txt (two characters)
```

List files in subdirectory:

```duso
data_files = list_files("data/*.json")
```

List all files in current directory:

```duso
all_files = list_files("*")
```

From embedded filesystem:

```duso
examples = list_files("/EMBED/examples/*.du")
```

From virtual filesystem:

```duso
generated = list_files("/STORE/*.txt")
```

## See Also

- [list_dir() - List directory contents](/docs/reference/list_dir.md)
- [copy_file() - Copy files](/docs/reference/copy_file.md)
- [move_file() - Move files](/docs/reference/move_file.md)
- [remove_file() - Delete files](/docs/reference/remove_file.md)
