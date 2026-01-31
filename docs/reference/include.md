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

## Scope

`include()` executes in your **current scope**. All variables and functions defined in the included script become available in your script—they "leak" into your namespace:

```duso
include("config.du")
print(api_url)   // Variables from config.du are now available
print(timeout)
```

This makes `include()` useful for configuration, helpers, and composing multiple scripts together.

## See Also

- [require() - Load module](/docs/reference/require.md)
- [load() - Read file as string](/docs/reference/load.md)
