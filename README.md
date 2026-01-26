![](docs/duso-logo.png)

# Duso - Scripted Intelligence

A batteries-included app ecosystem designed for building and deploying intelligence applications. Purpose-built to work smoothly with AI coding assistants and human developers alike.

Includes the duso scripting language, runtime, and customizable embedded application scripts and modules. It's like a "best of" some of your favorite scripting languages and environments. Simple, practical, powerful.

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

## Develop

- zero-install experience: download a single binary and start developing
- all docs, libraries, and code examples included in the binary, just `-run`, `-doc`, or `-extract` what you want to play with

## Deploy

- simple one-file deployment
- always work with the same library and environment

## Customize

- extend core features with custom duso modules or go
- build your own modded binary with your mix of features
- embed the script interpreter and runtime in your own go apps

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

# Language


- **Full lexical scoping** Closures, `var` keyword, implicit locals
- **Objects as blueprints** Constructor pattern with field overrides
- **String templates** Embed expressions with `{{expr}}` syntax
- **Multiline strings** Clean syntax with `"""..."""`
- **Exception handling** `try/catch` blocks
- **Functional programming** `map()`, `filter()`, `reduce()` for data transformation
- **Parallel execution** `parallel()` for concurrent independent operations
- **Claude integration** Claude API module via `require("claude")`
- **File I/O** (CLI) `load()`, `save()`, `include()` functions
- **Extensible** Add duso modules or register custom Go functions from host applications

## Example code

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

// Multi-line strings are great for markdown
// (and our markdown() builtin formats it nicely for consoles)
print(markdown("""
  # Results

  The following data is a synthesis of analyses from
  {{len(agents)}} agents given slightly different priorities
  and access to the test data along with any web searches
  each found useful.

  {{agents.map(function(response)
    print(response + "\n---\n")
  end)}}
"""))

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

## It looks a bit like lua but a few differences:

  - array indexes start at 0 not 1
  - no lua tables, duso uses arrays and objects
  - objects are simple key/value structures
  - no prototypes, classes, etc. but composition-leaning
  - simple parallel process scheme
  - great string templates and smart multilie strings
  - interactive console debugging
  - breakpoint() and watch() supported as first-class builtins, not a debugger afterthought
  - lots more, but still with a familiar feel

## Learning the Language

- **[Learning Duso](docs/learning_duso.md)** Full introduction nd brief tour of Duso with short examples and links to more detailed info.

The examples directory is also loaded with additional examples.

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

# Four Tiers of Deployments

Choose your level of customization—from zero setup to full Go integration.

## **Tier 1: Out-of-Box (Zero Customization)**

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

## **Tier 2: Light Customization (Duso Modules)**

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

## **Tier 3: Heavy Customization (Go Layer)**

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

## **Tier 4: Full Embedding**

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

# Project Structure

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

# Contributing

Please see [CONTRIBUTING.md] for more details.

## Go developers

You're wizards. We need you. Please reach out with any suggestions for optimization, new built-ins, middleware, or just making our code better. You will walk among us with our reverence.

## Duso developers

We need more modules! Let us know your ideas or what you're working on. We want to vet and include as many useful modules as possible. You'd be helping our community and we'd all love and admire you for it!

We have `stdlib/` modules. These are core-level, vendor-neutral things like http services. These we take great care with because other modules oftn depend on them.

We also have `contrib/` modules. These are often vendor specific (db vendors, specific apis, etc). They are hugely important for frowing our community. These modules are what helps bring in devs who just need to get the job done and don't have time to craft a solid lib.

# Sponsors

- [**Shannan.dev**](https://shannan.dev) business intelligence solutions
- [**Ludonode**](https://ludonode.com) agentic development and consulting

# License

MIT License - see [LICENSE](LICENSE) file for details.
