# Go Embedding Examples

These examples demonstrate how to embed Duso in Go applications.

Each file is a complete, self-contained example that can be run independently.

## Examples

### 01-hello-world.go

**Simplest embedding**

Demonstrates:
- Creating an interpreter
- Executing a script
- Handling errors

```bash
go run 01-hello-world.go
```

### 02-custom-functions.go

**Register Go functions callable from Duso**

Demonstrates:
- Registering custom functions with `RegisterFunction()`
- Calling from Duso with named arguments
- Returning values
- Multiple functions

```bash
go run 02-custom-functions.go
```

### 03-config-dsl.go

**Use Duso as a configuration language**

Demonstrates:
- Loading configuration scripts
- Accessing nested object values
- Configuration validation
- Real-world pattern: config files

```bash
go run 03-config-dsl.go
```

### 04-task-scripting.go

**Orchestrate workflows with Duso**

Demonstrates:
- Multiple registered functions
- Defining workflows in Duso
- Calling Duso functions from Go with `Call()`
- Control flow in scripts
- Real-world pattern: ETL, task orchestration, agents

```bash
go run 04-task-scripting.go
```

## Build All

```bash
go build -o 01-hello-world 01-hello-world.go
go build -o 02-custom-functions 02-custom-functions.go
go build -o 03-config-dsl 03-config-dsl.go
go build -o 04-task-scripting 04-task-scripting.go

./01-hello-world
./02-custom-functions
./03-config-dsl
./04-task-scripting
```

## Which One Should I Read?

- **Learning Go embedding?** → Start with `01-hello-world.go`
- **Adding custom functions?** → Read `02-custom-functions.go`
- **Building a config system?** → Read `03-config-dsl.go`
- **Orchestrating workflows?** → Read `04-task-scripting.go`

## Key Concepts

### Creating an Interpreter

```go
interp := script.NewInterpreter(false)
```

### Executing Scripts

```go
result, err := interp.Execute(`
    x = 5
    print(x)
`)
```

### Registering Functions

```go
interp.RegisterFunction("myFunc", func(args map[string]any) (any, error) {
    value := args["param"].(float64)
    return value * 2, nil
})
```

### Calling Duso Functions

```go
interp.Execute(`function add(a, b) return a + b end`)
result, _ := interp.Call("add", 5, 3)
```

### Getting Variables

```go
interp.Execute(`x = 42`)
value := interp.GetVariable("x")
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

### Pattern: Plugin System

Let users extend your app with scripts.

```go
// User writes: function onEventHook(event) ... end
interp.Execute(userPluginScript)

// Call from Go
interp.Call("onEventHook", eventData)
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

- [Embedding Documentation](../../docs/embedding/) - Full guides
- [API Reference](../../docs/embedding/API_REFERENCE.md) - Complete API
- [Custom Functions Guide](../../docs/embedding/CUSTOM_FUNCTIONS.md) - More patterns
- [Language Spec](../../docs/language-spec.md) - Language reference
