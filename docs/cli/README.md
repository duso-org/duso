# Duso CLI User Guide

Write and run Duso scripts from the command line with built-in file I/O and Claude API integration.

## What is the CLI?

The Duso CLI (`duso` command) lets you:

- **Write scripts** - In Duso language with full syntax support
- **Read and write files** - Access the filesystem with `load()` and `save()`
- **Include other scripts** - Compose scripts with `include()`
- **Integrate with Claude** - Call Claude API with `claude()` and `conversation()`
- **Run workflows** - Orchestrate multi-step processes and agents

## Getting Started

→ **[Quick Start Guide](GETTING_STARTED.md)** - Write your first Duso script in 5 minutes

## Features

### Full Language Support

Every feature of the Duso language works in the CLI:

- Variables, objects, arrays
- Functions with closures
- Control flow (if/while/for)
- Exception handling (try/catch)
- String templates with `{{expr}}`
- All built-in functions (math, string, array, JSON, date/time)

See [language-spec.md](../language-spec.md) for complete reference.

### File I/O

Read and write files with `load()` and `save()`:

```duso
// Read a file
config = load("config.json")

// Process it
data = parse_json(config)

// Save result
save("output.json", format_json(data))
```

→ **[File I/O Guide](FILE_IO.md)** - Complete file operations documentation

### Include Other Scripts

Use `include()` to load and execute other scripts in your current environment:

```duso
// Load helpers
include("helpers.du")

// Now use functions defined in helpers.du
result = helper_function()
```

→ **[File I/O Guide](FILE_IO.md)** for `include()` details

### Claude Integration

Call Claude API directly from your scripts:

```duso
// Single-shot query
response = claude("What is 2+2?")

// Multi-turn conversation
agent = conversation(system = "You are a helpful assistant")
answer1 = agent.prompt("Hello!")
answer2 = agent.prompt("How are you?")
```

→ **[Claude Integration Guide](CLAUDE_INTEGRATION.md)** - Full Claude API documentation

## Installation

Build the CLI from source:

```bash
# Clone repository
git clone https://github.com/duso-org/duso
cd duso

# Build
go build -o duso cmd/duso/main.go

# Run
./duso examples/core/basic.du
```

## Usage

```bash
duso [options] <script>
```

**Options:**
- `-v` - Verbose mode (debug output)

**Examples:**

```bash
# Run a script
duso script.du

# Run with verbose output
duso -v script.du

# Run script from examples
duso examples/core/basic.du
duso examples/cli/file-io.du
```

## Script Examples

All examples in `/examples/core/` work with the CLI:

**Language Fundamentals:**
- `examples/core/basic.du` - Variables and operators
- `examples/core/functions.du` - Function definition and calling
- `examples/core/arrays.du` - Array operations

**CLI-Specific Features:**
- `examples/cli/file-io.du` - Reading and writing files
- `examples/cli/multi-file.du` - Using include() for modular scripts
- `examples/cli/comedy_writer.du` - Claude integration example
- `examples/cli/spy_workflow.du` - Multi-step agent workflow
- `examples/cli/panel/` - Expert panel analysis

→ **[CLI Examples](EXAMPLES.md)** - Links to all example scripts

## Common Workflows

### Reading Configuration

```duso
// config.du (created by user or another program)
settings = {
    apiKey = "sk-xxx",
    timeout = 30,
    debug = false
}

// script.du
config = load("config.du")
print(config)
```

```bash
duso script.du
```

### Processing Data

```duso
// Load and process data
input = load("data.json")
data = parse_json(input)

// Transform
processed = []
for item in data do
    processed = append(processed, {
        id = item.id,
        name = upper(item.name)
    })
end

// Save result
save("output.json", format_json(processed))
```

### Calling Claude

```duso
// Single query
response = claude("Summarize this text: " + text)

// Multi-turn conversation
analysis = conversation(
    system = "You are a code reviewer",
    model = "claude-opus-4-5-20251101"
)

review = analysis.prompt("Review this code: " + code)
improvements = analysis.prompt("What's most critical to fix?")
```

### Script Composition

```duso
// helpers.du
function parseConfig(filename)
    content = load(filename)
    return parse_json(content)
end

// main.du
include("helpers.du")

config = parseConfig("config.json")
print(config)
```

## Environment Variables

**Claude API Key:**

```bash
export ANTHROPIC_API_KEY=sk-ant-xxxxx
duso my-script.du
```

Or pass directly in script:

```duso
response = claude("prompt", key = "sk-ant-xxxxx")
```

## Next Steps

- **[Getting Started](GETTING_STARTED.md)** - Write your first script
- **[File I/O Guide](FILE_IO.md)** - Work with files and include()
- **[Claude Integration](CLAUDE_INTEGRATION.md)** - Use Claude API
- **[Language Spec](../language-spec.md)** - Full language reference
- **[Examples](EXAMPLES.md)** - Script examples

## Troubleshooting

**"duso command not found"**
- Build it: `go build -o duso cmd/duso/main.go`
- Or add to PATH: `export PATH=$PATH:/path/to/duso`

**"Error: undefined variable=xxx"**
- Variable `xxx` hasn't been defined yet
- Check spelling and ensure it's in scope

**"Error: cannot read file: no such file or directory"**
- Check the filename and path
- Paths are relative to the script's directory

**"Error: undefined function=claude"**
- Make sure you've set `ANTHROPIC_API_KEY` environment variable
- Check that your API key is valid

## See Also

- [Language Specification](../language-spec.md) - Complete syntax and built-in functions
- [File I/O Guide](FILE_IO.md) - load(), save(), include()
- [Claude Integration](CLAUDE_INTEGRATION.md) - claude(), conversation()
- [Examples](EXAMPLES.md) - Example scripts
