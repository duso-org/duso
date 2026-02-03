# current_dir()

Get the current working directory.

## Signature

```duso
current_dir()
```

## Parameters

None

## Returns

String: the absolute path of the current working directory

## Examples

Print current directory:

```duso
print("Working in: " + current_dir())
```

Build absolute paths:

```duso
cwd = current_dir()
config_path = cwd + "/config.json"
print("Config at: " + config_path)
```

Log script location:

```duso
wd = current_dir()
now = format_time(now(), "iso")
append_file("log.txt", now + " - Running in: " + wd + "\n")
```

## Future

This function is designed with future extensibility in mind. In future versions, duso may support setting the working directory for a script's scope (similar to changing directory in a shell, but isolated to that script).

## See Also

- [file_exists() - Check if file exists](/docs/reference/file_exists.md)
- [list_dir() - List directory contents](/docs/reference/list_dir.md)
