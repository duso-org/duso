# Embedding API Reference

Complete API documentation for embedding Duso in Go applications.

## Core Interface

### `NewInterpreter(verbose bool) *Interpreter`

Creates a new Duso interpreter instance.

```go
interp := script.NewInterpreter(false)  // false = no debug output
```

**Parameters:**
- `verbose` (bool) - Enable verbose logging for debugging

**Returns:** Pointer to a new Interpreter

---

## Interpreter Methods

### `Execute(source string) (any, error)`

Executes a Duso script and returns the result (or last value).

```go
result, err := interp.Execute(`
    x = 5
    y = 10
    x + y
`)
if err != nil {
    fmt.Println("Error:", err)
} else {
    fmt.Println("Result:", result)  // 15
}
```

**Parameters:**
- `source` (string) - Duso script code

**Returns:**
- `any` - Result of executing the script (usually nil unless script returns)
- `error` - Error if parsing or execution failed

---

### `RegisterFunction(name string, fn GoFunction) error`

Registers a custom Go function callable from Duso scripts.

```go
interp.RegisterFunction("add", func(args map[string]any) (any, error) {
    a := args["a"].(float64)
    b := args["b"].(float64)
    return a + b, nil
})

// Now callable from Duso as: result = add(a = 5, b = 3)
```

**Parameters:**
- `name` (string) - Function name in Duso
- `fn` (GoFunction) - Function implementation

**Returns:**
- `error` - Error if registration fails

**Function Signature:**
```go
type GoFunction func(args map[string]any) (any, error)
```

The `args` map contains named arguments from the Duso call:
- Key: parameter name
- Value: passed value (any type)

---

### `GetVariable(name string) any`

Retrieves a variable from the script's global scope.

```go
interp.Execute(`
    user = {name = "Alice", age = 30}
`)
user := interp.GetVariable("user")
```

**Parameters:**
- `name` (string) - Variable name

**Returns:**
- `any` - Variable value, or nil if not found

---

### `SetVariable(name string, value any) error`

Sets a variable in the global scope (before or after execution).

```go
interp.SetVariable("apiKey", "secret123")
interp.Execute(`
    result = callAPI(key = apiKey)
`)
```

**Parameters:**
- `name` (string) - Variable name
- `value` (any) - Value to set

**Returns:**
- `error` - Error if setting fails

---

### `Call(functionName string, args ...any) (any, error)`

Calls a Duso function (defined in scripts) from Go.

```go
interp.Execute(`
    function multiply(a, b)
        return a * b
    end
`)

result, err := interp.Call("multiply", 6, 7)  // 42
```

**Parameters:**
- `functionName` (string) - Name of function defined in Duso
- `args` (...any) - Positional arguments

**Returns:**
- `any` - Function return value
- `error` - Error if function not found or execution fails

---

## Value Types

Duso values map to Go types as follows:

| Duso Type | Go Type | Example |
|-----------|---------|---------|
| nil | nil | nil |
| number | float64 | 42.5 |
| string | string | "hello" |
| boolean | bool | true |
| array | []any | []any{1, 2, 3} |
| object | map[string]any | map[string]any{"x": 1} |
| function | (not accessible) | (callable only within Duso) |

---

## Type Conversion

When working with values from Duso, use type assertions:

```go
// From Duso to Go
obj := interp.GetVariable("config")
configMap := obj.(map[string]any)
host := configMap["host"].(string)
port := int(configMap["port"].(float64))

// From Go to Duso
interp.SetVariable("limit", float64(100))
interp.SetVariable("options", map[string]any{
    "verbose": true,
    "timeout": 30.0,
})
```

---

## Error Handling

Duso errors include context about where they occurred:

```go
_, err := interp.Execute(`
    x = undefined_var
`)
if err != nil {
    fmt.Println(err)
    // Output: undefined variable=undefined_var
}
```

Common error types:
- **Parse errors** - Syntax issues (line number included)
- **Undefined variable** - Variable not found
- **Type errors** - Operation on incompatible types
- **Division by zero** - Arithmetic error
- **Index out of bounds** - Array/object access error

---

## Registering Objects

Register complex data structures:

```go
// Register multiple related functions as an "object"
databaseFunctions := map[string]script.GoFunction{
    "query": func(args map[string]any) (any, error) {
        // SQL query implementation
    },
    "insert": func(args map[string]any) (any, error) {
        // SQL insert implementation
    },
}

interp.RegisterFunction("db", func(args map[string]any) (any, error) {
    return databaseFunctions, nil
})

// Use from Duso:
// result = db().query(sql = "SELECT ...")
```

---

## Performance Considerations

1. **Interpreter Reuse** - Create one interpreter and reuse it for multiple scripts (it's efficient)
2. **Variable Scope** - Variables persist between Execute() calls in the same interpreter
3. **Function Registration** - Register functions once, call many times
4. **Script Complexity** - Simple scripts are very fast; tree-walking interpreter for complex logic

---

## Complete Example

```go
package main

import (
    "fmt"
    "github.com/duso-org/duso/pkg/script"
)

func main() {
    interp := script.NewInterpreter(false)

    // Register custom function
    interp.RegisterFunction("getUserAge", func(args map[string]any) (any, error) {
        userID := int(args["id"].(float64))
        // Simulate database query
        ages := map[int]float64{1: 25, 2: 30, 3: 35}
        return ages[userID], nil
    })

    // Execute script
    script := `
        age = getUserAge(id = 2)
        message = "User is {{age}} years old"
        result = {status = "ok", message = message}
    `
    interp.Execute(script)

    // Get result
    result := interp.GetVariable("result")
    fmt.Printf("%+v\n", result)
    // Output: map[message:User is 30 years old status:ok]
}
```

---

## See Also

- [Custom Functions Guide](CUSTOM_FUNCTIONS.md) - More function registration patterns
- [Patterns & Use Cases](PATTERNS.md) - Common application patterns
- [Language Spec](/docs/language-spec.md) - Language syntax reference
