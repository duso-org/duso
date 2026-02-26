# Registering Custom Functions

Learn how to add your domain logic to Duso through custom Go functions.

## Basic Function Registration

The simplest custom function:

```go
interp.RegisterFunction("add", func(args map[string]any) (any, error) {
    a := args["a"].(float64)
    b := args["b"].(float64)
    return a + b, nil
})

// Call from Duso:
interp.Execute(`result = add(a = 5, b = 3)`)
```

## Handling Arguments

### Named Arguments (Recommended)

```go
interp.RegisterFunction("createUser", func(args map[string]any) (any, error) {
    name := args["name"].(string)
    email := args["email"].(string)
    age := int(args["age"].(float64))

    // Create user...
    return map[string]any{
        "id": 123,
        "name": name,
    }, nil
})

// Called with named arguments:
interp.Execute(`
    user = createUser(name = "Alice", email = "alice@example.com", age = 30)
`)
```

### Optional Arguments

```go
interp.RegisterFunction("greet", func(args map[string]any) (any, error) {
    name := "World"
    if val, ok := args["name"]; ok {
        name = val.(string)
    }

    greeting := "Hello, " + name
    return greeting, nil
})

// Can be called with or without arguments:
interp.Execute(`
    a = greet()              // "Hello, World"
    b = greet(name = "Bob")  // "Hello, Bob"
`)
```

### Type Checking

```go
interp.RegisterFunction("formatNumber", func(args map[string]any) (any, error) {
    value, ok := args["value"]
    if !ok {
        return nil, fmt.Errorf("missing required argument: value")
    }

    num, ok := value.(float64)
    if !ok {
        return nil, fmt.Errorf("value must be a number, got %T", value)
    }

    return fmt.Sprintf("%.2f", num), nil
})
```

## Return Values

### Simple Values

```go
// Number
interp.RegisterFunction("random", func(args map[string]any) (any, error) {
    return rand.Float64(), nil
})

// String
interp.RegisterFunction("timestamp", func(args map[string]any) (any, error) {
    return time.Now().String(), nil
})

// Boolean
interp.RegisterFunction("isValid", func(args map[string]any) (any, error) {
    return true, nil
})
```

### Arrays

```go
interp.RegisterFunction("getNumbers", func(args map[string]any) (any, error) {
    return []any{1.0, 2.0, 3.0, 4.0, 5.0}, nil
})

// Use in Duso:
interp.Execute(`
    nums = getNumbers()
    for n in nums do
        print(n)
    end
`)
```

### Objects

```go
interp.RegisterFunction("getUserInfo", func(args map[string]any) (any, error) {
    userID := int(args["id"].(float64))

    return map[string]any{
        "id": float64(userID),
        "name": "Alice",
        "email": "alice@example.com",
        "active": true,
    }, nil
})

// Use in Duso:
interp.Execute(`
    user = getUserInfo(id = 1)
    print(user.name)  // "Alice"
`)
```

## Error Handling

Return errors from functions:

```go
interp.RegisterFunction("divide", func(args map[string]any) (any, error) {
    a := args["a"].(float64)
    b := args["b"].(float64)

    if b == 0 {
        return nil, fmt.Errorf("division by zero")
    }

    return a / b, nil
})

// Catch errors in Duso:
interp.Execute(`
    try
        result = divide(a = 10, b = 0)
    catch (err)
        print("Error: " + err)
    end
`)
```

## Real-World Examples

### Database Query

```go
interp.RegisterFunction("queryDB", func(args map[string]any) (any, error) {
    sql := args["sql"].(string)

    // Execute SQL...
    rows, err := db.Query(sql)
    if err != nil {
        return nil, err
    }

    // Convert to Duso array of objects
    var results []any
    for rows.Next() {
        var id int
        var name string
        rows.Scan(&id, &name)
        push(results, map[string]any{
            "id": float64(id),
            "name": name,
        })
    }

    return results, nil
})

// Use:
interp.Execute(`
    users = queryDB(sql = "SELECT id, name FROM users")
    for user in users do
        print(user.name)
    end
`)
```

### HTTP Request

```go
interp.RegisterFunction("httpGet", func(args map[string]any) (any, error) {
    url := args["url"].(string)

    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    return map[string]any{
        "status": float64(resp.StatusCode),
        "body": string(body),
    }, nil
})

// Use:
interp.Execute(`
    response = httpGet(url = "https://api.example.com/data")
    data = parse_json(response.body)
`)
```

### File Operations

```go
interp.RegisterFunction("readFile", func(args map[string]any) (any, error) {
    filename := args["filename"].(string)

    content, err := os.ReadFile(filename)
    if err != nil {
        return nil, fmt.Errorf("cannot read file: %w", err)
    }

    return string(content), nil
})

interp.RegisterFunction("writeFile", func(args map[string]any) (any, error) {
    filename := args["filename"].(string)
    content := args["content"].(string)

    err := os.WriteFile(filename, []byte(content), 0644)
    if err != nil {
        return nil, fmt.Errorf("cannot write file: %w", err)
    }

    return true, nil
})

// Use:
interp.Execute(`
    data = readFile(filename = "input.txt")
    writeFile(filename = "output.txt", content = upper(data))
`)
```

## Organizing Functions

### Register Once, Use Many

```go
// In your setup code:
registerDatabaseFunctions(interp)
registerHTTPFunctions(interp)
registerFileOperations(interp)

// Then execute any number of scripts:
for {
    script := userScript()
    interp.Execute(script)
}
```

### Function Collections

```go
func registerDatabaseFunctions(interp *script.Interpreter) {
    interp.RegisterFunction("dbQuery", dbQuery)
    interp.RegisterFunction("dbInsert", dbInsert)
    interp.RegisterFunction("dbUpdate", dbUpdate)
    interp.RegisterFunction("dbDelete", dbDelete)
}

func registerHTTPFunctions(interp *script.Interpreter) {
    interp.RegisterFunction("httpGet", httpGet)
    interp.RegisterFunction("httpPost", httpPost)
    interp.RegisterFunction("httpDelete", httpDelete)
}
```

### Namespaced Functions (Object Pattern)

```go
// Create a namespace using an object
databaseAPI := map[string]func(map[string]any) (any, error){
    "query": dbQuery,
    "insert": dbInsert,
    "update": dbUpdate,
}

interp.RegisterFunction("db", func(args map[string]any) (any, error) {
    return databaseAPI, nil
})

// Use from Duso:
interp.Execute(`
    users = db()["query"](sql = "SELECT * FROM users")
`)
```

## Best Practices

1. **Consistent Naming** - Use snake_case (Duso convention)
2. **Type Assertions** - Always validate argument types
3. **Error Messages** - Return clear, helpful error messages
4. **Documentation** - Document what your functions do and their arguments
5. **Consistency** - Return consistent types from your functions
6. **Testing** - Write tests for your functions (both from Go and Duso)

## Testing Custom Functions

```go
func TestCustomFunction(t *testing.T) {
    interp := script.NewInterpreter(false)
    interp.RegisterFunction("add", func(args map[string]any) (any, error) {
        a := args["a"].(float64)
        b := args["b"].(float64)
        return a + b, nil
    })

    result, err := interp.Call("add", 5, 3)
    assert.NoError(t, err)
    assert.Equal(t, 8.0, result)

    // Or test through Duso
    interp.Execute(`result = add(a = 5, b = 3)`)
    val := interp.GetVariable("result")
    assert.Equal(t, 8.0, val)
}
```

## See Also

- [API Reference](/docs/embedding/api-reference.md) - RegisterFunction details
- [Patterns & Use Cases](/docs/embedding/patterns.md) - Common applications
- [Examples](/docs/embedding/examples.md) - Complete application examples
