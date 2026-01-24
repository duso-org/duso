# include()

Load and execute another Duso script in the current scope. Available in `duso` CLI only.

## Signature

```duso
include(filename)
```

## Parameters

- `filename` (string) - Path to the script, relative to script directory

## Returns

`nil`

## Examples

Include helper functions:

```duso
include("helpers.du")
result = my_helper_function()
```

Include configuration:

```duso
include("config.du")
// Variables from config.du are now available
print(api_url)
print(timeout)
```

Compose scripts:

```duso
include("database.du")
include("api.du")
db_result = query_database()
api_result = call_api()
```

## Notes

- Executes in current scope (variables leak into parent)
- Results are NOT cached (file re-executed each time)
- Best for configuration and shared utilities

## See Also

- [require() - Load module](./require.md)
- [load() - Read file as string](./load.md)
- [Modules](../language-spec.md#modules)
