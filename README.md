![GitHub Release](https://img.shields.io/github/v/release/:user/:repo) ![Apache 2.0 License](https://img.shields.io/badge/Apache%202.0-License-blue) ![Go 1.21+](https://img.shields.io/badge/go-v1.21%2B-cyan?logo=go) ![Tests](https://img.shields.io/badge/tests-394-brightgreen) ![Coverage](https://img.shields.io/badge/coverage-59.1%25-yellow)

![](docs/duso-logo.png)

# Duso

**Write intelligent automation, agent orchestration, and business logic with a practical scripting language designed for human and AI collaboration. Run swarms of running tasks with a simple concurrency model and interactively debug them one at a time. Build and deploy your app as a single binary file with zero install headache.**

## Why

Why make a new language? I wanted the power of go but the simplicity of some of my favorite scripting languages.

LLMs like Claude often struggle with human-friendly languages like Python and JavaScript. Both are wonderfully terse and expressive, but filled with small footguns that can be hard to debug.

So I gravitated toward go. It's copiled, rigid, and powerful. It worked. Bug counts went down, but development time went up. Go is powerful but with that comes a level of complexity that can confuse LLMs.

## What it's for

Duso runs hundreds or thousands of coordinated scripts concurrently and efficiently. It is purpose-built for AI-assisted workflows. String templates are perfect for AI prompts.

Build a single binary with zero dependencies that works the same way in 2026 and 2036.

Duso puts a simple scripting language into a powerful architecture built in Go. with everything needed to develop, build, and deploy. No npm hell. No version conflicts. No missing packages.

## Quick Start

### 1. Build the binary

```bash
./build.sh
```

This handles Go embed setup, fetches the version from git, and builds the binary to `bin/duso`.

**Optional:** Make it available everywhere by creating a symlink:

```bash
ln -s $(pwd)/bin/duso /usr/local/bin/duso
```

Now you can run `duso` from any directory.

### 2. Try the REPL (30 seconds)

```bash
duso -repl
```

Then type:

```duso
print("Hello, Duso!")
claude = require("claude")
response = claude.prompt("What is 2 + 2?")
print(response)
```

Exit with `exit()`.

### 3. Run a script

```bash
duso examples/core/basic.du
```

### 4. Write your own

```bash
echo 'print("Hello, World!")' > hello.du
duso hello.du
```

Or inline:

```bash
duso -c 'print("1 + 2 =", 1 + 2)'
```

## What You Can Build

Duso's language and runtime are well featured. But everything aligns around orchestrating AI agents, applying business logic, and processing information. It's like having a server with its own development environment and tools built-in.

### Multi-agent analysis (run in parallel):

```duso
claude = require("claude")

results = parallel(
  function()
    return claude.prompt("Analyze from perspective A")
  end,
  function()
    return claude.prompt("Analyze from perspective B")
  end,
  function()
    return claude.prompt("Analyze from perspective C")
  end
)

// Synthesize results
synthesis = claude.prompt("Synthesize these three analyses: " + format_json(results))
print(synthesis)
```

### Orchestrate agent swarms

Beyond parallel execution, Duso enables complex orchestration patterns for spawning and coordinating multiple workers:
- `run()`: Execute script synchronously, blocking
- `spawn()`: Execute script in background, non-blocking
- `context()`: Access request data and metadata
- `datastore()`: Thread-safe key-value store with optional disk persistence. Atomic operations, condition variables (`wait()`, `wait_for()`) for synchronization, and coordination across concurrent processes—essential for distributed agent workflows.

Learn more: `duso -doc datastore` for swarm coordination, or `duso -doc` for the full reference.

### Intelligent automation

```duso
claude = require("claude")
reviewer = claude.conversation(system = "You are a code reviewer")

for file in files do
  code = load(file)
  review = reviewer.prompt("Review: " + code)
  save("reviews/{{file}}.md", review)
end
```

### Business logic with AI

```duso
claude = require("claude")

// String templates make crafting prompts natural
context = load("customer-data.json")
prompt = """
  Analyze this customer data and identify opportunities:

  {{context}}

  Format response as JSON with keys: opportunities, risk_score, recommendation
"""

result = claude.prompt(prompt)
save("analysis.json", result)
```

## Key Features

- **Parallel execution**: Run multiple agents concurrently with `parallel()`
- **Advanced concurrency support**: Backed by go, well-known for its solid concurreny support, with simple tooling at the script level
- **Swarm-friendly**: coordinate spawned agents and other processes with a fast, thread-safe, in-memory key-value datastore
- **Claude integration**: `require("claude")` and start building with AI immediately
- **String templates**: Embed expressions with `{{expr}}` for dynamic prompts
- **Closures & lexical scoping**: Full closure support with `var` keyword
- **Functional programming**: `map()`, `filter()`, `reduce()` for data transformation
- **Objects as blueprints**: Simple constructor pattern, no class complexity
- **Exception handling**: `try/catch` blocks
- **Console debugger**: `-debug` brings `breakpoint()`, `watch()`, and execptions alive with code context, stack trace and interactive inspection and resume
- **Concurrent-friendly debugging**: All debugs are queued, separate script processes are held, you can go through issues one by one
- **File I/O**: `load()`, `save()`, `include()`, plus smart `require("module")`
- **Extensible**: Register custom Go functions or add Duso modules
- **Forever stable**: Single binary with all libs and docs embedded, zero external dependencies, runs forever
- **Custom builds**: embed your own scripts, sandbox bits you don't want, ship to production

## Deployment: Choose Your Level

1. **Out-of-Box**: Download a binary. Run scripts. Done. No setup, no dependencies.

2. **Custom Modules**: Fork Duso, add your own `.du` modules to `contrib/`, build a custom binary for your team.

3. **Custom Runtime**: Modify the interpreter itself. Add operators, syntax, or built-in functions. Build a domain-specific language.

4. **Embedding**: Embed Duso as a scripting layer inside your Go applications. Users write scripts, you control the sandbox.

**All tiers share the same superpower:** Deploy once, run forever. Your binary from 2025 runs identically in 2035—zero external dependencies, no version conflicts, no bitrot.

## Built-In Documentation

### Browser-Based (Web Server)

Launch an interactive documentation server in your browser:

```bash
duso -docserver
# Opens http://localhost:5150
# Searchable docs, all in-process
# Built entirely from Duso scripts
```

This demonstrates Duso as a server: the docserver is a Duso HTTP application serving documentation. Build your own:

```duso
server = http_server({port = 8080})
server.route("POST", "/analyze", "handlers/analyze.du")
server.route("GET", "/status", "handlers/status.du")
server.start()

// Request handlers are Duso scripts with access to request context
// Perfect for AI agent APIs
```

### Terminal (CLI Reference)

Comprehensive built-in documentation without leaving your terminal:

```bash
duso -doc spawn       # Agent spawning patterns
duso -doc datastore  # Coordination primitives
duso -doc claude     # Claude API integration
duso -doc            # Full reference
```

No website. No hunting. Just `duso -doc TOPIC` and get instant answers.

### Web Requests (Handy Curl Replacement)

Need to fetch a URL quickly? Use Duso's `fetch()` builtin as a lightweight curl alternative:

```bash
duso -c 'print(fetch("https://example.com").body)'
```

No caching, automatic redirects, and response body directly to stdout. Perfect for testing APIs, webhooks, and local servers during development.

## Quality & Testing

- **394 tests** - Comprehensive coverage across script language and runtime
- **59.1% coverage** - Runtime package thoroughly verified
- [Detailed test coverage →](./TEST_COVERAGE.md)

## Contributing

**We need you.** Duso thrives on community contributions.

- **Everyone**: Report bugs, broken docs, share cool examples. Even forking the repo will increase our exposure and help us get syntax highlighting for `.du` files in GitHub!
- **Module authors**: Write a stdlib or contrib module (database clients, API wrappers, etc.). These are what make the runtime actually useful to real people.
- **Go developers** m: Performance optimizations, new built-ins, ideas for the core runtime. Help us make Duso faster and more powerful.

See [CONTRIBUTING.md](/CONTRIBUTING.md) for guidelines on how to get involved.

## Development

```bash
# Build
./build.sh

# Run tests
go test ./...

# Run an example
duso examples/core/basic.du
```

## Contributors

- Dave Balmer: design, development, documentation, dedication
- Maybe you...?

## Sponsors

- **[Shannan.dev](https://shannan.dev)**: Provides business intelligence solutions
- **[Ludonode](https://ludonode.com)**: Provides agentic development and consulting
- Also maybe you...?

## Learn More

- **[Learning Duso](/docs/learning-duso.md)**: Guided tour of the language with examples
- **[CLI User Guide](/docs/cli/README.md)**: Building and running Duso scripts
- **[Embedding Guide](/docs/embedding/README.md)**: Using Duso in Go applications
- **[Internals](/docs/internals.md)**: Architecture and runtime design

## License

Apache License 2.0 (see [LICENSE](/LICENSE) file for details) (C) 2026 Ludonode LLC

