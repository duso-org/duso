# CLI Examples

Example Duso scripts demonstrating CLI features.

## Language Examples (`examples/core/`)

These work with the CLI and demonstrate language features:

**Start here:**
- `examples/core/basic.du` - Variables, operations
- `examples/core/functions.du` - Function definition and calling
- `examples/core/structures.du` - Objects as blueprints

**Data structures:**
- `examples/core/arrays.du` - Array operations
- `examples/core/methods.du` - Object methods

**Advanced:**
- `examples/core/templates.du` - String templates
- `examples/core/test_json.du` - JSON parsing/formatting
- `examples/core/ternary-test.du` - Conditional expressions

**Run any:**
```bash
duso examples/core/basic.du
duso examples/core/functions.du
```

---

## CLI Examples (`examples/cli/`)

These use CLI-specific features (file I/O, Claude):

### File I/O

**`examples/cli/file-io.du`** - Reading and writing files
```bash
duso examples/cli/file-io.du
```
Demonstrates:
- `load()` - Read files
- `save()` - Write files
- Data processing

**`examples/cli/multi-file.du`** - Using multiple files
```bash
duso examples/cli/multi-file.du
```
Demonstrates:
- `include()` - Load other scripts
- Code organization
- Shared functions

**`examples/cli/with-include.du`** - Include pattern
```bash
duso examples/cli/with-include.du
```
Demonstrates:
- Including helper files
- Accessing shared variables

### Claude Integration

**`examples/cli/fun.du`** - Interactive conversation
```bash
ANTHROPIC_API_KEY=sk-ant-xxxxx duso examples/cli/fun.du
```
Demonstrates:
- `conversation()` - Multi-turn conversation
- Maintaining context across prompts
- Interactive workflows

**`examples/cli/comedy_writer.du`** - Single-shot Claude calls
```bash
ANTHROPIC_API_KEY=sk-ant-xxxxx duso examples/cli/comedy_writer.du
```
Demonstrates:
- `claude()` - Single query
- Text processing
- Saving results

**`examples/cli/spy_workflow.du`** - Multi-step agent workflow
```bash
ANTHROPIC_API_KEY=sk-ant-xxxxx duso examples/cli/spy_workflow.du
```
Demonstrates:
- Multiple agent conversations
- Complex workflows
- Context management

**`examples/cli/spy_vs_guard.du`** - Multiple perspectives
```bash
ANTHROPIC_API_KEY=sk-ant-xxxxx duso examples/cli/spy_vs_guard.du
```
Demonstrates:
- Multiple independent agents
- Comparison workflows

**`examples/cli/panel/panel.du`** - Expert panel analysis
```bash
ANTHROPIC_API_KEY=sk-ant-xxxxx duso examples/cli/panel/panel.du
```
Demonstrates:
- Multiple expert agents
- Parallel analysis
- Result synthesis
- File I/O integration

---

## Running Examples

### Without Claude (local only)

```bash
# Language features
duso examples/core/basic.du
duso examples/core/arrays.du
duso examples/core/functions.du

# File I/O
duso examples/cli/file-io.du
duso examples/cli/multi-file.du
```

### With Claude (requires API key)

```bash
# Set your API key
export ANTHROPIC_API_KEY=sk-ant-xxxxx

# Run examples
duso examples/cli/fun.du
duso examples/cli/comedy_writer.du
duso examples/cli/spy_workflow.du
duso examples/cli/panel/panel.du
```

---

## Example Snippets

### Reading a File

```duso
content = load("input.txt")
print("File has " + len(content) + " characters")
```

### Processing JSON

```duso
jsonText = load("data.json")
data = parse_json(jsonText)

for item in data do
    print(item.name)
end
```

### Saving Results

```duso
results = [
    {id = 1, status = "done"},
    {id = 2, status = "pending"}
]

save("output.json", format_json(results))
```

### Including Another Script

```duso
include("helpers.du")

result = helper_function(data)
```

### Simple Claude Query

```duso
question = "What is the capital of France?"
answer = claude(question)
print(answer)
```

### Multi-Turn Conversation

```duso
agent = conversation(system = "You are a helpful assistant")

response1 = agent.prompt("Hello!")
response2 = agent.prompt("How do I learn Duso?")
response3 = agent.prompt("What's a good first project?")
```

---

## Creating Your Own Scripts

### Basic Script

```duso
// my-script.du
print("Hello from Duso!")

x = 5
y = 10
print("Result: " + (x + y))
```

```bash
duso my-script.du
```

### Using Files

```duso
// process.du
input = load("data.txt")
output = upper(input)
save("output.txt", output)
print("Processed!")
```

### With Functions

```duso
// script.du
function process(data)
    return upper(data)
end

content = load("input.txt")
result = process(content)
save("output.txt", result)
```

### With Claude

```duso
// analyze.du
data = load("input.txt")

analysis = claude(
    "Analyze this text and provide 3 key points:\n\n" + data
)

save("analysis.txt", analysis)
print("Analysis complete!")
```

---

## Best Practices

1. **Start with basics** - Learn language with `examples/core/`
2. **Test locally** - File I/O examples before Claude
3. **Handle errors** - Use try/catch for external operations
4. **Organize code** - Use `include()` for larger scripts
5. **Save results** - Use `save()` to persist output
6. **Be specific** - Better Claude prompts get better results

---

## Next Steps

- **[Getting Started](GETTING_STARTED.md)** - Write your first script
- **[File I/O Guide](FILE_IO.md)** - load(), save(), include()
- **[Claude Integration](CLAUDE_INTEGRATION.md)** - API details
- **[Language Reference](/docs/language-spec.md)** - Complete spec

## Troubleshooting Examples

**"Examples not found"**
- Make sure you're in the repo root directory
- Or use full path: `duso /path/to/duso/examples/core/basic.du`

**"Error: file not found"**
- Check working directory: examples assume you run from repo root
- Or use full paths in load(): `load("/path/to/file.txt")`

**"Claude not working"**
- Set API key: `export ANTHROPIC_API_KEY=sk-ant-xxxxx`
- Verify key is valid
- Check internet connection

**"Script runs but no output"**
- Add `print()` statements to see progress
- Use `-v` flag for verbose output: `duso -v script.du`
