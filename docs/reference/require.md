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

## Scope

`require()` executes in an **isolated scope**. Variables and functions defined in the module don't leak into your script. Instead, the module returns a value (usually an object) that you assign to a variable:

```duso
http = require("http")    // Module exports, assigned to http variable
db = require("db")        // Module exports, assigned to db variable
// Variables inside http.du and db.du are not visible here
```

## Module Search Order

When you call `require("modulename")`, Duso searches for the module in this order and uses the first match:

1. **Local files** - Relative to your script directory (`./modulename.du`)
2. **$DUSO_LIB paths** - If `$DUSO_LIB` environment variable is set, search those directories
3. **Embedded stdlib** - Built-in standard library modules
4. **Embedded contrib** - Built-in contributed modules

## Comparison with include()

| Feature | require() | include() |
|---------|-----------|----------|
| Scope | Isolated | Current scope |
| Variables | Don't leak | Leak into parent |
| Returns | Module value | nil |
| Use for | Libraries, reusable code | Configuration, helpers |

## See Also

- [include() - Include script](/docs/reference/include.md)
- [load() - Read file as string](/docs/reference/load.md)
