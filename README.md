# Duso - Embeddable Scripting Language

A lightweight, Go-based scripting language designed for **embedding in applications** and **command-line agent orchestration**.

Write scripts once, run them anywhere—embedded in your Go app, or standalone via the CLI.

## Two Paths, One Language

### 🔨 **Embedding in Go Applications**

Use Duso as a scripting layer in your Go code. Configure behavior, extend functionality, or let users write plugins—all without external dependencies.

```go
package main

import "github.com/duso-org/duso/pkg/script"

func main() {
    interp := script.NewInterpreter(false)
    interp.RegisterFunction("myFunc", func(args map[string]any) (any, error) {
        // Custom Go logic here
        return "result", nil
    })

    result, _ := interp.Execute(`
        x = 5
        y = myFunc()
        print(x)
    `)
}
```

→ **[Embedding Guide](docs/embedding/)** - Full documentation for Go developers

### 🎯 **CLI - Script Writing & Agent Orchestration**

Write scripts from the command line with built-in file I/O and Claude API integration.

```bash
duso examples/core/basic.du
duso examples/cli/comedy_writer.du
```

```duso
// File I/O
config = load("config.du")
results = parse_json(config)

// Claude integration
agent = conversation(system = "You are a helpful assistant")
response = agent.prompt("Analyze this data")

save("output.json", format_json(results))
```

→ **[CLI User Guide](docs/cli/)** - Complete guide for script writers

## Key Features

- **Zero external dependencies** - Runs on Go stdlib only
- **Full lexical scoping** - Closures, `var` keyword, implicit locals
- **Objects as blueprints** - Constructor pattern with field overrides
- **String templates** - Embed expressions with `{{expr}}` syntax
- **Multiline strings** - Clean syntax with `"""..."""`
- **Exception handling** - `try/catch` blocks
- **Claude integration** (CLI) - `claude()` and `conversation()` functions
- **File I/O** (CLI) - `load()`, `save()`, `include()` functions
- **Extensible** - Register custom Go functions from host applications

## Quick Start

### Using the CLI

```bash
go build -o duso cmd/duso/main.go

# Run a script
./duso examples/core/basic.du

# Write your own script
echo 'print("Hello, {{name}}")' > hello.du
./duso hello.du
```

### Embedding in Go

```go
import "github.com/duso-org/duso/pkg/script"

interp := script.NewInterpreter(false)
result, _ := interp.Execute(`
    name = "World"
    message = "Hello, {{name}}!"
    print(message)
`)
```

## Examples

**Learning the Language:**
- Start with `examples/core/basic.du` for language fundamentals
- Explore `examples/core/` for all language features (works everywhere)
- Check `examples/cli/` for CLI-specific features (file I/O, Claude)

→ **[Examples Directory](examples/README.md)** - Full guide to all examples

## Project Structure

```
cmd/duso/              - CLI application entry point
pkg/script/            - Core language runtime (lexer, parser, evaluator)
pkg/anthropic/         - Claude API client (CLI-provided)
pkg/cli/               - CLI-specific functions (file I/O, Claude integration)
examples/
  ├── core/            - Language feature demonstrations (embeddable)
  └── cli/             - CLI-specific examples (file I/O, Claude)
docs/
  ├── embedding/       - Guide for Go developers embedding Duso
  └── cli/             - Guide for CLI script writers
vscode/                - VSCode syntax highlighting extension
```

## Documentation

**For Everyone:**
- [**Language Specification**](docs/language-spec.md) - Complete syntax, types, operators, and semantics

**For Go Developers (Embedding):**
- [**Embedding Guide**](docs/embedding/README.md) - How to use Duso in Go applications
- [**API Reference**](docs/embedding/API_REFERENCE.md) - Go API documentation
- [**Custom Functions**](docs/embedding/CUSTOM_FUNCTIONS.md) - Registering functions from Go

**For Script Writers (CLI):**
- [**CLI User Guide**](docs/cli/README.md) - Using the duso command
- [**File I/O**](docs/cli/FILE_IO.md) - load(), save(), include()
- [**Claude Integration**](docs/cli/CLAUDE_INTEGRATION.md) - conversation(), claude()

**Additional:**
- [**Implementation Notes**](docs/implementation-notes.md) - Design decisions and architecture
- [**Contributing**](CONTRIBUTING.md) - Guidelines for contributors

## Language Example

```duso
// Variables and types (all features work everywhere)
name = "Alice"
skills = ["Go", "Python", "Rust"]
config = {timeout = 30, retries = 3}

// Functions with closures
function makeGreeter(greeting)
  function greet(person)
    return greeting + ", " + person
  end
  return greet
end

sayHello = makeGreeter("Hello")
print(sayHello(name))

// Objects as constructors
Request = {method = "GET", timeout = 30}
req = Request(method = "POST")
print(req.method)

// String templates
message = "User {{name}} has {{len(skills)}} skills"
print(message)

// Error handling
try
  result = 1 / 0
catch (error)
  print("Error: " + error)
end

// CLI-only features (when running with duso command)
data = load("config.json")
agent = conversation(system = "You are a helpful assistant")
response = agent.prompt("Help me with this data")
```

## VSCode Extension

Syntax highlighting for `.du` files available in `vscode/` directory. See extension documentation for installation.

## Development

```bash
# Build the CLI
go build -o duso cmd/duso/main.go

# Run tests (if available)
go test ./...

# Run examples
./duso examples/core/basic.du
```

## License

MIT License - see [LICENSE](LICENSE) file for details.
