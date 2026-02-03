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
    push(results, result)
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

## Relationship to CLI and Runtime

The CLI (`cmd/duso/main.go`) is a **user** of `pkg/script/` and `pkg/runtime/`. It:
1. Creates an interpreter
2. Registers runtime functions (HTTP, datastore, concurrency) using `pkg/runtime/`
3. Registers CLI functions (file I/O, Claude) using `pkg/cli/`
4. Executes user scripts
5. Displays results

**When you embed:**
- **Minimal embedding**: Use only `pkg/script/` for pure language features
- **With runtime features**: Add `pkg/runtime/` for HTTP, datastore, and concurrency
- **With CLI features**: Optionally add `pkg/cli/` for file I/O and Claude integration

## What's Embeddable vs CLI-Only

**✅ Embeddable (in `pkg/runtime/`):**
- `http_server()` - Create HTTP servers
- `datastore()` - Thread-safe coordination
- `spawn()`, `run()` - Background execution
- `context()` - Request context management
- `parallel()` - Concurrent execution (in `pkg/script/`)

**❌ CLI-Only (in `pkg/cli/`):**
- `fetch()` - Make HTTP requests
- `load()`, `save()` - File I/O
- `include()`, `require()` - Module loading
- `claude()`, `conversation()` - Claude API
- `env()` - Environment variables
- `doc()` - Documentation lookup

## Troubleshooting

**"How do I add HTTP or datastore to my embedded app?"**
→ These are in `pkg/runtime/` and are fully embeddable. Import `pkg/runtime/` and create instances directly, or use `pkg/cli/register.go` as a reference for exposing them as script functions.

**"How do I add file I/O to my embedded app?"**
→ File I/O is CLI-specific (in `pkg/cli/functions.go`). You can:
1. Call `cli.RegisterFunctions()` if you want standard file I/O
2. Implement your own `load()` and `save()` with custom access control
3. Copy patterns from `pkg/cli/functions.go` and register your own

**"How do I use Claude in an embedded app?"**
→ Claude integration is CLI-specific (in `pkg/cli/`). You can:
1. Call `cli.RegisterFunctions()` to enable Claude
2. Implement your own wrapper around `pkg/anthropic/` for custom behavior

**"Can I prevent certain operations?"**
→ Simply don't register dangerous functions. The core evaluator has:
- ✅ No I/O (unless you add it)
- ✅ No network (unless you add it)
- ✅ No shell access (unless you add it)
- ✅ No file access (unless you add it)
- ✅ No Claude access (unless you add it)

The `pkg/runtime/` features (HTTP, datastore) are available to use safely in controlled contexts.

## See Also

- [**Learning Duso**](/docs/learning-duso.md) - Complete language guide
- [**Built-in Functions Reference**](/docs/reference/index.md) - All function reference
- [**API Reference**](API_REFERENCE.md) - Full Go API docs
- [**go-embedding examples**](/go-embedding/) - Complete code examples
