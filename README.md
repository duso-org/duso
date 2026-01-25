![](docs/duso-logo.png)

# Duso - Scripted Intelligence

A non-nonsense modern app ecosystem designed for building and deploying zero-install single-binary intelligence applications.

Includes the duso scriptng language, runtime, and customizable embedded applicatin scripts and modules.

Or, embed the duso interpreter in your own go-based applications as an alternative language that was purpose-built to work smoothly with AI coding assistants and human developers alike. It's like a "best of" some of your favorite scripting languages. Simple, practical, powerful.

## Four Tiers of Usage

Choose your level of customization—from zero setup to full Go integration.

```
duso script layers:
 ┌────────────────────────────────────────────┐ 
 │    duso scripts provide business logic,    │ 
 │   agent orchestration, api, application    │ 
 ├─────────────┬───────────────┬──────────────┤ 
 │   stdlib    │   contrib     │   custom     │ 
 │   modules   │   modules     │   modules    │ 
 └─────────────┴───────────────┴──────────────┘ 
go layers:
 ┌─────────────┬───────────────┬──────────────┐ 
 │ script      │ built-ins     │  custom      │ 
 │ interpreter │ runtime       │  go code     │ 
 ├─────────────┴───────────────┴──────────────┤ 
 │     go provides speed, cross-platform,     │ 
 │   concurrency, networking, file io, etc.   │ 
 └────────────────────────────────────────────┘ 
```


### **Tier 1: Out-of-Box (Zero Customization)**

Download a single Duso binary and run scripts instantly. No installation, no dependencies.

```bash
duso script.du
```

The binary includes:
- **stdlib modules** - Core utilities like `http`, maintained by Duso org
- **contrib modules** - Community modules curated by Duso org
- Everything frozen at release time, runs forever

Perfect for: Quick scripting, agents, automation, archival.

---

### **Tier 2: Light Customization (Duso Modules)**

Fork Duso, add your own `.du` modules to `contrib/`, build a custom binary.

```bash
# In your fork: my-org/duso
# Add: contrib/mycompany-helpers/
#      ├── helpers.du
#      └── helpers.md

go generate ./cmd/duso
mkdir -p bin
go build -o bin/duso-myorg ./cmd/duso
```

Your team now uses your binary:

```duso
helpers = require("mycompany-helpers")
result = helpers.process_data(input)
```

Perfect for: Team standardization, code sharing, freezing org-specific utilities.

---

### **Tier 3: Heavy Customization (Go Layer)**

Modify the Duso runtime itself—add new operators, syntax, or built-in functions.

```bash
# In your fork: my-org/duso
# Modify: pkg/script/evaluator.go (add features)
#         contrib/ (add modules)

mkdir -p bin
go build -o bin/duso-custom ./cmd/duso
```

Perfect for: Domain-specific languages, specialized agents, custom operators.

---

### **Tier 4: Full Embedding**

Embed Duso as a scripting layer in your own Go applications.

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

Perfect for: Plugin systems, configuration languages, user-extensible applications.

---

## The Freezing Advantage

All tiers share the same benefit: **Zero bitrot, forever stable.**

```
my-archive/
├── duso-v0.5.2          # Binary from 2025
├── scripts/
│   ├── script1.du
│   └── script2.du
└── data/

# Run in 2035:
./duso-v0.5.2 scripts/script1.du  # Works exactly as it did
```

Every Duso binary is completely self-contained:
- All stdlib modules baked in
- All contrib modules baked in
- Zero external dependencies (Go stdlib only)
- Works indefinitely—no package managers, no version conflicts, no broken links

This is different from npm, pip, or other systems where packages disappear, versions break, and dependencies conflict. Archive your binary with your scripts and it works forever.

---

**[CLI User Guide](docs/cli/)** - For Tier 1 script writers
**[Embedding Guide](docs/embedding/)** - For Tier 4 developers
**[Contributing](CONTRIBUTING.md)** - For Tier 2-3 customization

## Key Features

- **Zero external dependencies** - Runs on Go stdlib only
- **Full lexical scoping** - Closures, `var` keyword, implicit locals
- **Objects as blueprints** - Constructor pattern with field overrides
- **String templates** - Embed expressions with `{{expr}}` syntax
- **Multiline strings** - Clean syntax with `"""..."""`
- **Exception handling** - `try/catch` blocks
- **Functional programming** - `map()`, `filter()`, `reduce()` for data transformation
- **Parallel execution** - `parallel()` for concurrent independent operations
- **Claude integration** (CLI) - Claude API module via `require("claude")`
- **File I/O** (CLI) - `load()`, `save()`, `include()` functions
- **Extensible** - Register custom Go functions from host applications

## Quick Start

### Using the CLI

```bash
./build.sh

# Run a script
./bin/duso examples/core/basic.du

# Write your own script
echo 'print("Hello, {{name}}")' > hello.du
./bin/duso hello.du
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
./build.sh

# Run tests (if available)
go test ./...

# Run examples
./bin/duso examples/core/basic.du
```

## License

MIT License - see [LICENSE](LICENSE) file for details.
