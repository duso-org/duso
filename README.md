[![Apache 2.0 License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0) ![Go 1.25](https://img.shields.io/badge/Go-1.25-darkcyan?logo=go) ![Tests 1290](https://img.shields.io/badge/Tests-1290-green) ![Coverage ~45%](https://img.shields.io/badge/Coverage-~45%25-yellowgreen) ![Current Release](https://img.shields.io/github/v/release/duso-org/duso)

![Duso logo which is a stylized ASL hand sign for the letter "D"](docs/duso-logo.png)

# Duso

Write intelligent automation, agent orchestration, and business logic with a practical scripting language designed for human and AI collaboration. Run swarms of running tasks with a simple concurrency model and interactively debug them one at a time. Build and deploy your app as a single binary file with zero install headache.

Duso puts a simple scripting language into a powerful architecture built in Go. A single binary with everything you need to develop, build, and deploy. No npm hell. No version conflicts. No missing packages.

For the adventurous, build the go binary with your own scripts inside and lanch it as a zero-install app. For the super adventurous, add your own custom go modules, or embed the language into your own go-based apps.

## Quick Start

### 0. Install the binary

Homebrew (Mac, Linux):

```bash
# First time
brew tap duso-org/homebrew-duso
brew install duso

# Later: update
brew upgrade duso

# Run it!
duso
```

Direct from Github (Mac, Windows, Linux):

> TODO: link here from first release!


### 1. Build the binary

You'll need go installed on your system. Then just use our handy build script in the project directory:

```bash
./build.sh
```

This handles Go embed setup, fetches the version from git, and builds the binary to `bin/duso`.

**Optional:** Make it available everywhere by creating a symlink:

```bash
ln -s $(pwd)/bin/duso /usr/local/bin/duso
```

Now you can run `duso` from any directory.

### 3. Run a script

```bash
duso examples/agents/self-aware-claude.du
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

## Editor integration

### VSCode extension

- search for "duso" in the VSCode marketplace and install it
- syntax highlighting
- keyword hints
- autocomplete
- problem hints

### Other editors that support LSP

The `duso` binary includes LSP support built-in. Just point your editor to:

```bash
duso -lsp
```

### Help wanted

I would love to have syntax highlighting and SLP for all. The basic TM syntax highlighting is buried in the [VSCode extension GitHub repo](https://github.com/duso-org/duso-vscode).

## What You Can Build

Duso's language and runtime are well featured. But everything aligns around orchestrating AI agents, applying business logic, and processing information. It's like having a server with its own development environment and tools built-in.

### Multi-agent analysis (run in parallel):

> TODO: this is illutrative but doesn't work, make a better short example here

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

> TODO: an example here!

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

- **Parallel execution**: Run multiple functions simply and concurrently with `parallel()`
- **Swarm-friendly**: coordinate spawned agents and other processes with a fast, thread-safe, in-memory key-value datastore
- **Advanced concurrency support**: Backed by go, well-known for its solid concurreny support, with simple tooling at the script level
- **Claude integration**: `require("claude")` and start building with AI immediately
- **String templates**: Embed expressions with `{{expr}}` for dynamic prompts
- **Functional programming bits**: `map()`, `filter()`, `reduce()` for data transformation
- **Closures & lexical scoping**: Full closure support, even in objects without needing `self`
- **No globals**: the highest scope is in a script process (goroutine)
- **Automatic deep-copy**: to keep cuncurrent processes thread-safe when moving data between them
- **Objects as blueprints**: Simple constructor pattern, no class complexity
- **Exception handling**: `try/catch()` blocks, `throw()`
- **Console debugger**: `-debug` brings `breakpoint()`, `watch()`, and execptions alive with code context, stack trace and interactive inspection and resume
- **Concurrent-friendly debugging**: All debugs are queued, separate script processes are held, you can go through issues one by one
- **File I/O**: all the usual suspects with basic globs plus smart `require("module")`
- **Extensible**: Register custom Go functions or add Duso modules (please!)
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
```

The docserver is a smallish Duso script serving documentation.

### Terminal (CLI Reference)

Comprehensive built-in documentation without leaving your terminal for all keywords and built-in funcitons with examples:

```bash
duso -doc

# or give it a keyword or function name
duso -doc datastore
duso -doc claude
```

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
- **Go developers**: Performance optimizations, new built-ins, ideas for the core runtime. Help us make Duso faster and more powerful.

- [CONTRIBUTING.md](/CONTRIBUTING.md) for crontributing guidelines
- [COMMUNITY.md](/COMMUNITY.md) for community guielines

## Development

```bash
# Build
./build.sh

# Run an example
./bin/duso examples/core/basic.du
```

## Contributors

- Dave Balmer: design, development, documentation, dedication

## Sponsors

- **[Shannan.dev](https://shannan.dev)**: Provides AI-driven business intelligence solutions
- **[Ludonode](https://ludonode.com)**: Provides agentic development and consulting

## Learn More

- **[Learning Duso](/docs/learning-duso.md)**: Guided tour of the language with examples
- **[Function Reference](/docs/reference/index.md)**: All 100+ built-in functions with examples
- **[Embedding Guide](/docs/embedding/README.md)**: Using Duso in Go applications
- **[Internals](/docs/internals.md)**: Architecture and runtime design
- Lots of examples in the, well, `examples` directory

## Learning as an LLM

Duso is designed to be natural for AI. If you're an LLM (or using one), the binary includes everything you need to learn and explore:

```bash
duso -read                    # Start with the README overview
duso -read docs/              # Browse available docs
duso -read docs/learning-duso.md  # Learn the language
duso -read docs/reference/    # Browse all functions
duso -read docs/reference/map.md  # Look up specific functions
```

All documentation is embedded in the binary. No cloning, no network calls—just pure text output you can parse and learn from. Perfect for agentic workflows.

## License

Apache License 2.0 (see [LICENSE](/LICENSE) file for details) © 2026 Ludonode LLC

