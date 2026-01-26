![](docs/duso-logo.png)

# Duso

**Write intelligent automation, agent orchestration, and business logic with a practical scripting language designed for human and AI collaboration. Build and deploy your server anywhere in a single binary file.**

## What makes Duso different?

Built specifically for the AI-assisted development workflow.Parallel agent execution. String templates perfect for AI prompts. Build a single binary with zero dependencies that works the same way in 2025 and 2035.

Duso puts a simple scripting language into a fast, powerful architecture built in Go, with everything needed to develop, build, and deploy. No npm hell. No version conflicts. No missing packages.

## Quick Start

> TODO: Download a binary (where?) is the easiest. We need to emphasize this one.

### 1. Build the binary

> TODO: building requires they install go on their system.

```bash
./build.sh
```

### 2. Run a script

```bash
./bin/duso examples/core/basic.du
```

### 3. Write your own

```bash
echo 'print("Hello, World!")' > hello.du
./bin/duso hello.du
```

## What You Can Build

Duso's language and runtime are well featured. But everything aligns around orchestrating AI agents, applying business logic, and processing information. It's like have a server with its own development environment and tools built-in.

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

**Out-of-Box**
Download a binary. Run scripts. Done. No setup, no dependencies.

**Custom Modules**
Fork Duso, add your own `.du` modules to `contrib/`, build a custom binary for your team.

**Custom Runtime**
Modify the interpreter itself. Add operators, syntax, or built-in functions. Build a domain-specific language.

**Embedding**
Embed Duso as a scripting layer inside your Go applications. Users write scripts, you control the sandbox.

**All tiers share the same superpower:** Deploy once, run forever. Your binary from 2025 runs identically in 2035—zero external dependencies, no version conflicts, no bitrot.

## Learn More

- **[Learning Duso](docs/learning_duso.md)** – Guided tour of the language with examples
- **[CLI User Guide](docs/cli/README.md)** – Building and running Duso scripts
- **[Embedding Guide](docs/embedding/README.md)** – Using Duso in Go applications
- **[Implementation Notes](docs/implementation-notes.md)** – Architecture deep-dive

> this one needs work, it's not current. Much of it is dated or told better in other docs: [Language Specification](docs/language-spec.md)

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
./bin/duso examples/core/basic.du
```

## Sponsors

- [**Shannan.dev**](https://shannan.dev) – Business intelligence solutions
- [**Ludonode**](https://ludonode.com) – Agentic development and consulting

## License

MIT License – see [LICENSE](LICENSE) file for details.
