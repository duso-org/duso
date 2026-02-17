# sys()

Access values from the sys datastore. Provides a convenient interface to read system information including CLI flags, configuration, and other runtime data. Available in `duso` CLI only.

## Signature

```duso
sys(key)
```

## Parameters

- `key` (string) - The key to retrieve from sys datastore

## Returns

The value associated with the key, or nil if the key doesn't exist. Return type depends on what was stored (bool, string, number, object, etc.)

## Examples

Check CLI flags:

```duso
if sys("-debug") then
  print("Debug mode enabled")
end

if sys("-no-color") then
  print("Colors disabled")
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

Check multiple CLI options:

```duso
verbose = sys("-verbose") or false
no_files = sys("-no-files") or false
no_stdin = sys("-no-stdin") or false

if verbose then
  print("Running in verbose mode")
end

if no_files then
  print("File system access disabled")
end
```

## Notes

- The sys datastore contains CLI flags (stored with leading hyphen: `-debug`, `-no-color`) and other system information
- The `-config` flag is parsed into an object for convenient access to configuration values
- Boolean flags return `true` or `nil` (not `false`) when not set
- This is the recommended way to access system information and CLI options from scripts

## See Also

- [datastore() - Key-value store for coordination](/docs/reference/datastore.md)
- [env() - Read environment variables](/docs/reference/env.md)
- [context() - Access request context](/docs/reference/context.md)
