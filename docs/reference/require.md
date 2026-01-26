# require()

Load a module in an isolated scope and return its exports. Available in `duso` CLI only.

## Signature

```duso
require(moduleName)
```

## Parameters

- `moduleName` (string) - Name of the module to load

## Returns

Module's exported value (usually an object with functions and data)

## Examples

Load HTTP module:

```duso
http = require("http")
response = http.fetch("https://example.com")
```

Load custom module:

```duso
math = require("mylib-math")
result = math.add(5, 3)
```

Access module properties:

```duso
config = require("config-module")
print(config.version)
print(config.settings)
```

## Notes

- Executes in isolated scope (variables don't leak)
- Results are cached (subsequent calls return cached value)
- Best for reusable modules and libraries

## Module Search

Searches in order:
1. User-provided paths
2. Relative to script directory
3. DUSO_LIB environment variable
4. Built-in stdlib modules
5. Built-in contrib modules

## See Also

- [include() - Include script](./include.md)
- [load() - Read file as string](./load.md)
