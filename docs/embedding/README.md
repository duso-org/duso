# Embedding Duso in Go Applications

This guide is for Go developers who want to embed Duso as a scripting layer in their applications.

## What is Embedding?

Embedding means using Duso as a **library** in your Go code, rather than as a standalone CLI tool. Your Go application controls:

- Which functions are available in Duso scripts
- How scripts are loaded and executed
- What happens with script results
- When and how to call Duso code

## Why Embed Duso?

**Configuration Language**
```go
// Load app config from a Duso script
configScript := `
  server = {host = "localhost", port = 8080}
  database = {url = "localhost:5432", timeout = 30}
`
interp.Execute(configScript)
appConfig = interp.GetVariable("server")
```

**User Scripts & Plugins**
```go
// Let users extend your app with scripts
userScript := loadUserPlugin("~/.myapp/plugin.du")
interp.Execute(userScript)
result := interp.Call("onUserAction", userData)
```

**Workflow Orchestration**
```go
// Coordinate complex operations
orchestrationScript := `
  results = []
  for item in items do
    result = processItem(item)  // Your custom Go function
    results = append(results, result)
  end
  return format_json(results)
`
```

**DSLs (Domain-Specific Languages)**
```go
// Build a custom DSL on top of Duso
scriptingEngine := duso.NewEngine()
scriptingEngine.RegisterFunction("queryDatabase", queryDB)
scriptingEngine.RegisterFunction("sendEmail", sendEmail)
// Now your DSL can do: queryDatabase(...) and sendEmail(...)
```

## Key Principles

1. **Core Language Only** - Embedded Duso uses `pkg/script/` only
2. **No CLI Dependencies** - No file I/O, no Claude API (unless you add them)
3. **Zero External Dependencies** - Duso itself has no external deps (Go stdlib only)
4. **Extensible** - Register custom functions easily from Go
5. **Lightweight** - Minimal overhead for scripting capability

## What's Available When Embedded?

### Always Available
- Full Duso language syntax (variables, functions, objects, loops, etc.)
- All core built-in functions (math, string, array, JSON, date/time)
- No external dependencies

### NOT Available (unless you add them)
- File I/O (`load()`, `save()`, `include()`)
- Claude integration (`claude()`, `conversation()`)
- Any custom functions you haven't registered

## Quick Start

→ **[Getting Started](GETTING_STARTED.md)** - 5-minute tutorial

## Deep Dives

- [**API Reference**](API_REFERENCE.md) - Complete Go API documentation
- [**Custom Functions**](CUSTOM_FUNCTIONS.md) - How to add your own functions
- [**Patterns**](PATTERNS.md) - Common use cases and design patterns
- [**Examples**](EXAMPLES.md) - Full example applications

## Common Tasks

**Executing a script:**
```go
import "github.com/duso-org/duso/pkg/script"

interp := script.NewInterpreter(false)
result, err := interp.Execute(`print("hello")`)
```

**Registering a function:**
```go
interp.RegisterFunction("myFunc", func(args map[string]any) (any, error) {
    // args contains named arguments from the Duso script
    return "result", nil
})
```

**Getting values from scripts:**
```go
interp.Execute(`x = 42`)
value := interp.GetVariable("x")  // Returns 42
```

**Calling functions from Go:**
```go
interp.Execute(`function add(a, b) return a + b end`)
result, _ := interp.Call("add", 5, 3)  // Returns 8
```

## Next Steps

1. **[Getting Started Tutorial](GETTING_STARTED.md)** - Build your first embedded app
2. **[API Reference](API_REFERENCE.md)** - Understand the full API
3. **[Custom Functions](CUSTOM_FUNCTIONS.md)** - Add your domain logic
4. **[Examples](EXAMPLES.md)** - See complete applications

## Architecture Overview

```
Your Go App
    ↓
pkg/script/ (Duso core)
    ├── Lexer      - Tokenizes code
    ├── Parser     - Builds AST
    ├── Evaluator  - Executes AST
    ├── Builtins   - Core functions
    └── Value      - Runtime types
```

Your app imports `pkg/script` and calls its public API. That's it! No CLI code, no external dependencies.

## Relationship to CLI

The CLI (`cmd/duso/main.go`) is a **user** of `pkg/script/`. It:
1. Creates an interpreter
2. Registers CLI functions (file I/O, Claude) using `pkg/cli/`
3. Executes user scripts
4. Displays results

When you embed, you do steps 1-4 yourself, with your own functions instead of CLI functions.

## Troubleshooting

**"How do I add file I/O to my embedded app?"**
→ Copy the implementation from `pkg/cli/functions.go` and register it

**"How do I use Claude in an embedded app?"**
→ Check `pkg/cli/` for the implementation, register it with your API key

**"Can I prevent certain operations?"**
→ Simply don't register dangerous functions. The evaluator has no I/O, network, or shell capabilities unless you add them.

## See Also

- [**Learning Duso**](/docs/learning_duso.md) - Complete language guide
- [**Built-in Functions Reference**](/docs/reference/index.md) - All function reference
- [**API Reference**](API_REFERENCE.md) - Full Go API docs
- [**go-embedding examples**](../../examples/go-embedding/) - Complete code examples
