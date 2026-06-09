# env()

Read an environment variable. Available in `duso` CLI only.


`env(varname)`

```

## Parameters

- `varname` (string) - Name of the environment variable

## Returns

String value of the variable, or empty string if not set

## Examples

Read API key:

```duso
api_key = env("API_KEY")
if api_key == "" then
  print("Error: API_KEY not set")
  exit(1)
end
```

Read configuration:

```duso
timeout = tonumber(env("TIMEOUT") or "30")
is_debug = env("DEBUG") == "true"
```

Pass settings:

```bash
# From shell:
export DATABASE_URL="postgres://localhost"
export LOG_LEVEL="debug"
```

```duso
db_url = env("DATABASE_URL")
log_level = env("LOG_LEVEL")
```

## See Also

- [print() - Output text](/docs/reference/print.md)
- [tonumber() - Convert to number](/docs/reference/tonumber.md)
