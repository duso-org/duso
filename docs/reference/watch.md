# watch()

Monitor a file, directory, or glob pattern for changes.

## Signature

```duso
watch(path [, timeout = 30])
```

## Parameters

- `path` (string) - File path, directory, or glob pattern to watch
  - Absolute or relative paths (relative paths are resolved from script directory)
  - Supports `*` and `?` glob patterns
  - Cannot watch `/EMBED/` (read-only embedded filesystem)
- `timeout` (number, optional) - Maximum seconds to block waiting for changes (default: 30)

## Returns

- `true` if files matching the path have changed since the last call
- `false` if no changes detected before timeout expires
- First call always returns `false` (establishes baseline state)

## Behavior

- **Stateful**: Maintains memory of watched paths across calls. State is unique per (scriptPath, watchPath) pair.
- **Blocking**: Blocks for up to `timeout` seconds, checking every 1 second for changes
- **Efficient**: For directories without wildcards, only monitors subdirectory modification times (not individual files)
- **Wildcard support**: Glob patterns with `*` (any characters) and `?` (single character) are expanded and monitored

## How It Works

The function computes a hash (signature) of the watched files' modification times:
- **Directories**: Hashes all subdirectory paths and their modification times (detects any file/subdirectory changes within)
- **Files**: Hashes the file's path and modification time
- **Glob patterns**: Expands the pattern and hashes all matching files' paths and modification times

Changes to files automatically update their parent directory's modification time on all major filesystems, making directory watching very efficient.

## Examples

Watch a source directory and rebuild on changes:

```duso
server = http_server({port = 8080})
server.static("/", "./public")

// Start the server
server.route("GET", "/")
server.start()

// Watch for source changes and trigger rebuild
while true then
  if watch("./src", timeout = 30) then
    print("Source files changed, rebuilding...")
    rebuild()
  end
end
```

Watch for changes to specific file types:

```duso
if watch("*.md", timeout = 10) then
  print("Markdown files changed!")
  regenerate_docs()
end
```

Watch a single file:

```duso
if watch("config.json", timeout = 5) then
  print("Configuration updated")
  reload_config()
end
```

Establish baseline and check for changes:

```duso
print("Establishing baseline...")
watch("./data")

print("Waiting for data changes (up to 60 seconds)...")
if watch("./data", timeout = 60) then
  print("Data files changed!")
  process_changes()
else
  print("No changes detected")
end
```

## State Management

Each unique (script_path, watch_path) combination maintains its own state:

```duso
// Script: main.du
watch("./src")   // State: main.du:./src

watch("./docs")  // State: main.du:./docs

// Script: other.du (different script)
watch("./src")   // State: other.du:./src (different from main.du:./src)
```

## Error Cases

- **Path not found**: Returns error if the initial path doesn't exist (except for glob patterns)
- **Permission denied**: Returns error if path is inaccessible
- **Read-only filesystem**: Returns error if trying to watch `/EMBED/` (use `/STORE/` or regular filesystem instead)

```duso
if err = watch("/EMBED/") then
  // Error: cannot watch /EMBED/: embedded filesystem is read-only
end
```

## Performance

- **Directories**: Very efficient - only checks subdirectory modification times, not individual files
- **Large directories**: Fast even with thousands of files
- **Glob patterns**: Scans matching files (slower than directory-only mode, but still efficient)
- **Polling interval**: 1 second (fixed, for consistent responsiveness)

## Notes

- Only detects changes to the actual filesystem (modifications made outside Duso)
- State is per-process; multiple processes watching the same path maintain separate state
- Timeout of `0` is allowed and will return immediately with current vs. previous state without blocking
- `/STORE/` paths are not yet supported (planned for future versions)

## See Also

- [load() - Read files](/docs/reference/load.md)
- [save() - Write files](/docs/reference/save.md)
- [http_server() - HTTP server with file serving](/docs/reference/http_server.md)
