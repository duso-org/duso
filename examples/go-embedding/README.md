# Go Embedding Examples

These examples demonstrate how to embed Duso in Go applications.

Each file is a complete, self-contained example that can be run independently.

## Examples

### hello-world

**Simplest embedding**

Demonstrates:
- Creating an interpreter
- Executing a script
- Handling errors

```bash
go run ./hello-world
```

### custom-functions

**Register Go functions callable from Duso**

Demonstrates:
- Registering custom functions with `RegisterFunction()`
- Calling from Duso with named arguments
- Returning values
- Multiple functions

```bash
go run ./custom-functions
```

### config-dsl

**Use Duso as a configuration language**

Demonstrates:
- Defining configuration objects in Duso
- Scripts with nested data structures
- Real-world pattern: config files, app settings

```bash
go run ./config-dsl
```

### task-scripting

**Orchestrate workflows with Duso**

Demonstrates:
- Multiple registered Go functions
- Workflows in Duso that use those functions
- Control flow (if/then/else)
- Real-world pattern: ETL, task orchestration, agents

```bash
go run ./task-scripting
```

## Run Examples

Run directly:

```bash
go run ./hello-world
go run ./custom-functions
go run ./config-dsl
go run ./task-scripting
```

Or build to `bin/` and run:

```bash
go build -o bin/hello-world ./hello-world
go build -o bin/custom-functions ./custom-functions
go build -o bin/config-dsl ./config-dsl
go build -o bin/task-scripting ./task-scripting

./bin/hello-world
./bin/custom-functions
./bin/config-dsl
./bin/task-scripting
```

## Which One Should I Read?

- **Learning Go embedding?** → Start with `hello-world/`
- **Adding custom functions?** → Read `custom-functions/`
- **Building a config system?** → Read `config-dsl/`
- **Orchestrating workflows?** → Read `task-scripting/`

## Key Concepts

### Creating an Interpreter

```go
interp := script.NewInterpreter(false)
```

### Executing Scripts

```go
_, err := interp.Execute(`
    x = 5
    print(x)
`)
if err != nil {
    log.Fatal(err)
}
```

### Registering Functions

```go
interp.RegisterFunction("myFunc", func(args map[string]any) (any, error) {
    value := args["param"].(float64)
    return value * 2, nil
})
```

## Common Patterns

### Pattern: Configuration

Store app settings in Duso scripts.

```go
// Load config.du with app settings
config := loadConfigFromDuso("config.du")
startServer(config)
```

### Pattern: Custom DSL

Create a domain-specific language on top of Duso.

```go
// Register domain functions
registerDatabaseFunctions(interp)
registerAPIFunctions(interp)

// User scripts use your DSL
interp.Execute(userScript)
```

### Pattern: Custom DSL with Go Functions

Let scripts orchestrate custom Go functions.

```go
// Register domain-specific functions
interp.RegisterFunction("query", queryDB)
interp.RegisterFunction("process", processData)

// Script uses those functions
interp.Execute(`
    data = query("SELECT * FROM users")
    result = process(data)
    print(result)
`)
```

### Pattern: Workflow Orchestration

Coordinate multi-step processes.

```go
// Register services
interp.RegisterFunction("fetchData", ...)
interp.RegisterFunction("processData", ...)
interp.RegisterFunction("saveResults", ...)

// User writes orchestration logic
interp.Execute(orchestrationScript)
```

## Next Steps

- [Embedding Documentation](/docs/embedding/) - Full guides
- [API Reference](/docs/embedding/API_REFERENCE.md) - Complete API
- [Custom Functions Guide](/docs/embedding/CUSTOM_FUNCTIONS.md) - More patterns
- [Language Spec](/docs/language-spec.md) - Language reference
