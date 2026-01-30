# Getting Started with Duso CLI

Write and run your first Duso script in 5 minutes.

## Step 1: Build the CLI

```bash
./build.sh
```

This handles Go embed setup, fetches the version from git, and builds the binary to `bin/duso`.

**Optional:** Make it available everywhere by creating a symlink:

```bash
ln -s $(pwd)/bin/duso /usr/local/bin/duso
```

Now you can run `duso` from any directory without the `./bin/` prefix.

## Step 2: Create Your First Script

Create `hello.du`:

```duso
name = "World"
message = "Hello, " + name + "!"
print(message)
```

## Step 3: Run It

```bash
duso hello.du
# Output: Hello, World!
```

## Try the Language Features

Create `features.du`:

```duso
// Variables
x = 5
y = 10

// Arrays
numbers = [1, 2, 3, 4, 5]

// Objects
person = {name = "Alice", age = 30}

// Loops
print("Numbers:")
for n in numbers do
    print(n)
end

// Functions
function greet(name)
    return "Hello, " + name
end

// String templates
msg = "{{greet(person.name)}} (age {{person.age}})"
print(msg)

// Try/catch
try
    result = 1 / 0
catch (err)
    print("Caught error: " + err)
end
```

```bash
duso features.du
```

## Working with Files

Create `data.json`:

```json
{
    "users": [
        {"id": 1, "name": "Alice"},
        {"id": 2, "name": "Bob"}
    ]
}
```

Create `process.du`:

```duso
// Read file
jsonContent = load("data.json")
data = parse_json(jsonContent)

// Process
for user in data.users do
    print(user.name + " (ID: " + user.id + ")")
end

// Save result
output = {
    processedAt = format_time(now()),
    userCount = len(data.users)
}

save("result.json", format_json(output))
print("Saved result.json")
```

```bash
duso process.du
cat result.json
```

## Using Claude API

Set your API key:

```bash
export ANTHROPIC_API_KEY=sk-ant-xxxxx
```

Create `chat.du`:

```duso
// Single query
response = claude("What is the capital of France?")
print(response)

// Multi-turn conversation
agent = conversation(system = "You are a helpful math tutor")

answer1 = agent.prompt("What is 2 + 2?")
print(answer1)

answer2 = agent.prompt("How do I add fractions?")
print(answer2)
```

```bash
duso chat.du
```

## Organizing Scripts

Create `helpers.du`:

```duso
function formatCurrency(amount)
    return "$" + round(amount * 100) / 100
end

function greetUser(name)
    return "Welcome, " + name + "!"
end
```

Create `main.du`:

```duso
// Load helper functions
include("helpers.du")

// Use them
price = 19.99
greeting = "Alice"

print(greetUser(greeting))
print("Total: " + formatCurrency(price))
```

```bash
duso main.du
```

## Common Patterns

### Configuration File

```duso
// config.du
app = {
    name = "MyApp",
    port = 8080,
    debug = true
}
```

```duso
// main.du
include("config.du")

print("Starting " + app.name + " on port " + app.port)
```

### Data Transformation

```duso
input = load("input.csv")
lines = split(input, "\n")

results = []
for line in lines do
    parts = split(line, ",")
    results = append(results, {
        name = parts[0],
        value = tonumber(parts[1])
    })
end

save("output.json", format_json(results))
```

### Multi-Step Workflow

```duso
print("Step 1: Loading data...")
data = load("data.json")

print("Step 2: Processing...")
processed = parse_json(data)

print("Step 3: Analyzing...")
count = len(processed)
print("Found " + count + " items")

print("Step 4: Saving...")
save("report.txt", "Processed " + count + " items")

print("Done!")
```

## Next Steps

- **[File I/O Guide](FILE_IO.md)** - load(), save(), include() details
- **[Claude Integration](CLAUDE_INTEGRATION.md)** - Using Claude API
- **[Language Reference](../language-spec.md)** - Complete language spec
- **[Examples](EXAMPLES.md)** - More example scripts

## Tips

1. **Paths** - Relative to script location: `load("./data/file.txt")`
2. **Debugging** - Use `print()` to debug: `print("x=" + x)`
3. **Error Handling** - Use try/catch: `try ... catch (e) ... end`
4. **Formatting** - Use templates: `"Result: {{value}}"` instead of concatenation
5. **Modular** - Use `include()` to split large scripts into pieces

## Troubleshooting

**"cannot read file"**
- Check filename and path
- Paths are relative to script directory
- Use `./filename` for current directory

**"undefined variable"**
- Variable hasn't been defined
- Check spelling and scope

**"undefined function=claude"**
- Set ANTHROPIC_API_KEY environment variable
- Rebuild duso if needed

Run with `-v` flag for debug output:

```bash
duso -v script.du
```
