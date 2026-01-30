# Getting Started with Embedded Duso

Learn to embed Duso in your Go application in 5 minutes.

## Step 1: Import Duso

```go
package main

import (
    "fmt"
    "github.com/duso-org/duso/pkg/script"
)

func main() {
    // Create an interpreter
    interp := script.NewInterpreter(false)

    // Run a Duso script
    result, err := interp.Execute(`
        x = 5
        y = 10
        print(x + y)
    `)

    if err != nil {
        fmt.Println("Error:", err)
    }
}
```

**Run it:**
```bash
go mod tidy
go run main.go
# Output: 15
```

## Step 2: Get Values Back

Scripts can set variables that you read from Go:

```go
interp := script.NewInterpreter(false)
interp.Execute(`
    name = "Alice"
    age = 30
    skills = ["Go", "Python"]
`)

// Read variables back
name := interp.GetVariable("name")
age := interp.GetVariable("age")
skills := interp.GetVariable("skills")

fmt.Println(name, age, skills)
// Output: Alice 30 [Go Python]
```

## Step 3: Register Custom Functions

Your Go code can provide functions to Duso scripts:

```go
interp := script.NewInterpreter(false)

// Register a custom function
interp.RegisterFunction("greet", func(args map[string]any) (any, error) {
    // Extract the "name" argument
    name, ok := args["name"].(string)
    if !ok {
        name = "World"
    }
    return "Hello, " + name, nil
})

// Use it in Duso
result, _ := interp.Execute(`
    message = greet(name = "Bob")
    print(message)
`)
// Output: Hello, Bob
```

## Step 4: Call Functions from Go

You can define functions in Duso and call them from Go:

```go
interp := script.NewInterpreter(false)

// Define a function in Duso
interp.Execute(`
    function double(x)
        return x * 2
    end
`)

// Call it from Go
result, _ := interp.Call("double", 21)
fmt.Println(result)  // Output: 42
```

## Step 5: Load Script Files

Read Duso scripts from files:

```go
import "os"

interp := script.NewInterpreter(false)

// Load script from file
scriptContent, err := os.ReadFile("script.du")
if err != nil {
    fmt.Println("Error:", err)
    return
}

result, err := interp.Execute(string(scriptContent))
if err != nil {
    fmt.Println("Error:", err)
}
```

## Complete Example

Here's a configuration system using Duso:

```go
package main

import (
    "fmt"
    "github.com/duso-org/duso/pkg/script"
)

type AppConfig struct {
    Host     string
    Port     int
    LogLevel string
}

func main() {
    // Create interpreter
    interp := script.NewInterpreter(false)

    // Script to configure the app
    configScript := `
        config = {
            host = "localhost",
            port = 8080,
            logLevel = "info"
        }
    `

    // Execute config script
    _, err := interp.Execute(configScript)
    if err != nil {
        fmt.Println("Config error:", err)
        return
    }

    // Read config back into Go
    configObj := interp.GetVariable("config")
    configMap, ok := configObj.(map[string]any)
    if !ok {
        fmt.Println("Config is not an object")
        return
    }

    // Convert to Go struct
    appConfig := AppConfig{
        Host:     configMap["host"].(string),
        Port:     int(configMap["port"].(float64)),
        LogLevel: configMap["logLevel"].(string),
    }

    fmt.Printf("Config: %+v\n", appConfig)
    // Output: Config: {Host:localhost Port:8080 LogLevel:info}
}
```

## Next Steps

- **[API Reference](API_REFERENCE.md)** - See all available methods
- **[Custom Functions](CUSTOM_FUNCTIONS.md)** - Build powerful extensions
- **[Examples](EXAMPLES.md)** - Full application examples
- **[Language Spec](/docs/learning-duso.md)** - Learn the language syntax

## Common Patterns

### Configuration Language

```go
// Load user config from Duso DSL
configScript := loadConfigFile("app.du")
interp.Execute(configScript)
config := interp.GetVariable("settings")
```

### Plugin System

```go
// User writes plugins in Duso
userPlugin := loadUserFile("~/.myapp/plugin.du")
interp.Execute(userPlugin)

// Call plugin functions
result, _ := interp.Call("onDataReceived", data)
```

### Data Transformation

```go
// Use Duso to transform/process data
interp.RegisterFunction("parseJSON", parseJSON)
interp.RegisterFunction("database", dbQuery)

// User writes transformation logic
transformScript := `
    raw = database("SELECT * FROM users")
    processed = map(raw, function(user)
        return {id = user.id, name = upper(user.name)}
    end)
`
```

### Workflow Orchestration

```go
// Duso calls your Go functions in sequence
interp.RegisterFunction("fetchData", fetchData)
interp.RegisterFunction("processData", processData)
interp.RegisterFunction("storeData", storeData)

orchestration := `
    data = fetchData()
    results = processData(data)
    storeData(results)
`
```

## Tips

1. **Error Handling** - Always check the error returned by `Execute()`
2. **Type Conversion** - Values from Duso might need type assertion in Go
3. **Function Names** - Duso uses snake_case by convention
4. **Testing** - Write tests that verify both Duso scripts and Go functions
