# sys()

Access system information, CLI flags, and runtime data. Provides a convenient interface to query how the duso process was invoked. Available in `duso` CLI only.

`sys(key)`

## Parameters

- `key` (string) - The key to retrieve from sys datastore

## Returns

The value associated with the key, or nil if the key doesn't exist. Return type depends on what was stored (bool, string, object, array, etc.)

## Available Keys

**System Information**
- `sys("version")` - Duso version string (e.g., "1.6.17")
- `sys("args")` - Array of command-line arguments passed to duso

**CLI Flags** (stored with leading hyphen)
- `sys("-debug")` - Boolean, true if run via `duso debug script.du` subcommand
- `sys("-no-color")` - Boolean, true if `-no-color` flag passed
- `sys("-no-files")` - Boolean, true if `-no-files` flag passed
- `sys("-no-stdin")` - Boolean, true if `-no-stdin` flag passed
- `sys("-config")` - Object containing parsed config (if `-config` flag passed)

**Notes on Flags**
- Boolean flags return `true` when set, `nil` when not set (never `false`)
- Unknown flags are stored as-is with their flag name as key
- Use leading hyphen to check flag status: `sys("-flagname")`

## Examples

Check duso version:

```duso
version = sys("version")
print("Running duso " + version)
```

Access command-line arguments:

```duso
args = sys("args")
print("Arguments: " + format_json(args))
// args is an array of strings passed to duso
```

Check boolean CLI flags:

```duso
if sys("-debug") then
  print("Debug mode enabled (duso debug subcommand)")
end

if sys("-no-color") then
  print("Colors disabled")
end

if sys("-no-files") then
  print("File system access disabled")
end
```

Access configuration passed via `-config`:

```duso
config = sys("-config")
if config then
  port = config.port or 8080
  timeout = config.timeout or 30
  print("Server config: port={{port}}, timeout={{timeout}}")
end
```

Pass configuration via CLI:

```bash
duso -config 'port=8080, timeout=30' script.du
```

Check multiple CLI options:

```duso
if sys("-no-color") then
  print("Colors disabled")
end

if sys("-no-files") then
  print("File system access disabled")
end

if sys("-no-stdin") then
  print("Standard input disabled")
end
```

## Notes

- The sys datastore is read-only and contains system information and CLI flags captured at startup
- CLI flags are stored with a leading hyphen in the key name (e.g., `"-debug"`)
- The `-config` flag is automatically parsed from its string format into an object for convenient access
- Boolean flags follow the Lua convention: `true` when set, `nil` when not set (never `false`)
- The `"args"` array contains the full list of command-line arguments passed to duso
- Unknown custom flags are stored as-is with their flag name as key, following the same convention
- This is the recommended way to access system information and CLI options from scripts

## See Also

- [datastore() - Key-value store for coordination](/docs/reference/datastore.md)
- [env() - Read environment variables](/docs/reference/env.md)
- [context() - Access request context](/docs/reference/context.md)
