![](docs/duso-logo.png)

# Duso

**Write intelligent automation, agent orchestration, and business logic with a practical scripting language designed for human and AI collaboration. Build and deploy your server anywhere in a single binary file.**

## What makes Duso different?

Built specifically for the AI-assisted development workflow. Parallel agent execution. String templates perfect for AI prompts. Build a single binary with zero dependencies that works the same way in 2025 and 2035.

Duso puts a simple scripting language into a fast, powerful architecture built in Go, with everything needed to develop, build, and deploy. No npm hell. No version conflicts. No missing packages.

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

**Multi-agent analysis (run in parallel):**
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

**Orchestrate agent swarms:**

Beyond parallel execution, Duso enables complex orchestration patterns for spawning and coordinating multiple workers:
- `run()` – Execute script synchronously, blocking
- `spawn()` – Execute script in background, non-blocking
- `context()` – Access request data and metadata
- `datastore()` – Thread-safe key-value coordination

Learn more: `duso -doc datastore` for swarm coordination, or `duso -doc` for the full reference.

**Intelligent automation:**
```duso
claude = require("claude")
reviewer = claude.conversation(system = "You are a code reviewer")

for file in files do
  code = load(file)
  review = reviewer.prompt("Review: " + code)
  save("reviews/{{file}}.md", review)
end
```

**Business logic with AI:**
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

- **Claude integration** – `require("claude")` and start building with AI
- **Parallel execution** – Run multiple agents concurrently with `parallel()`
- **String templates** – Embed expressions with `{{expr}}` for dynamic prompts
- **Closures & lexical scoping** – Full closure support with `var` keyword
- **Functional programming** – `map()`, `filter()`, `reduce()` for data transformation
- **Objects as blueprints** – Simple constructor pattern, no class complexity
- **Exception handling** – `try/catch` blocks
- **File I/O** – `load()`, `save()`, `include()` (CLI mode)
- **Extensible** – Register custom Go functions or add Duso modules
- **Forever stable** – Single binary, zero external dependencies, runs forever

## Deployment: Choose Your Level

1. **Out-of-Box**
Download a binary. Run scripts. Done. No setup, no dependencies.

2. **Custom Modules**
Fork Duso, add your own `.du` modules to `contrib/`, build a custom binary for your team.

3. **Custom Runtime**
Modify the interpreter itself. Add operators, syntax, or built-in functions. Build a domain-specific language.

4. **Embedding**
Embed Duso as a scripting layer inside your Go applications. Users write scripts, you control the sandbox.

**All tiers share the same superpower:** Deploy once, run forever. Your binary from 2025 runs identically in 2035—zero external dependencies, no version conflicts, no bitrot.

## Instant Documentation

Every binary includes comprehensive built-in docs. Look up functions, modules, and features directly:

```bash
duso -doc string      # String functions (len, substr, split, replace, etc.)
duso -doc spawn       # Spawning background workers
duso -doc claude      # Claude API integration
```

No website. No hunting. Just `duso -doc TOPIC` and get instant answers in your terminal.

## Learn More

- **[Learning Duso](docs/learning_duso.md)** – Guided tour of the language with examples
- **[CLI User Guide](docs/cli/README.md)** – Building and running Duso scripts
- **[Embedding Guide](docs/embedding/README.md)** – Using Duso in Go applications
- **[Internals](docs/internals.md)** – Architecture and runtime design

## Contributing

**We need you.** Duso thrives on community contributions.

- **Module authors** – Write a stdlib or contrib module (http, database clients, API wrappers, etc.). These are what make the runtime actually useful to real people.
- **Go developers** – Performance optimizations, new built-ins, ideas for the core runtime. Help us make Duso faster and more powerful.

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on how to get involved.

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

- Dave Balmer design, development, documentation

## Sponsors

- [**Shannan.dev**](https://shannan.dev) Provides business intelligence solutions
- [**Ludonode**](https://ludonode.com) Provides agentic development and consulting

## License

MIT License – see [LICENSE](LICENSE) file for details.
