# Duso Language Usage Guide

How to run Duso scripts with the `duso` command and write scripts.

**For complete language specification:** See [language-spec.md](language-spec.md)

## Quick Start

The `duso` command allows you to run Duso scripts from the command line:

```bash
duso examples/basic.du
duso examples/structures.du
duso examples/functions.du
```

## Usage

```bash
duso [options] <script-path>
```

### Options

- `-v` - Enable verbose output for debugging

### Examples

**Run a script:**
```bash
duso examples/basic.du
```

**Run a script with verbose output:**
```bash
duso -v examples/variables.du
```

**Run a script in a subdirectory:**
```bash
duso ../scripts/my-script.du
```

**Get help:**
```bash
duso
```

## How It Works

1. The `duso` command runs your Duso script
2. Output is displayed in real-time
3. Script errors are displayed with context

## Creating Your Own Scripts

Create a new Duso script file (e.g., `my-script.du`):

```duso
// Define an object as a template/blueprint
Person = {name = "Unknown", age = 0}

// Create an instance with overrides
person = Person(name = "Alice", age = 30)
print(person.name)
print(person.age)

// Work with arrays
experts = ["alice", "bob", "charlie"]
for name in experts do
  print(name)
end

// Define functions
function greet(name)
  return "Hello, " + name
end

print(greet("World"))
```

Then run it:

```bash
duso my-script.du
```

## Language Features

See [language-spec.md](language-spec.md) for complete reference. Quick reference:

### Variables & Types
```duso
x = 5                    // number
name = "Alice"           // string
flag = true              // boolean
arr = ["a", "b", "c"]    // array (0-indexed)
obj = {key = "value"}     // object
```

### String Templates
```duso
name = "Alice"
msg = "Hello {{name}}"   // Output: Hello Alice

// Perfect for JSON/code (no escaping!)
json = "{\"key\": \"{{value}}\"}"
```

### Functions & Object Templates
```duso
function add(x, y) return x + y end

// Objects can be used as templates
Config = {timeout = 30, retries = 3}

// Call the template to create instances
config = Config(timeout = 60)
```

### Control Flow
```duso
if x > 5 then print("big") end
while x < 10 do x = x + 1 end
for i = 1, 10 do print(i) end
for item in arr do print(item) end

try
  operation()
catch error
  print("Error: " + error)
end
```

### Built-in Functions
```duso
print("value", 42)       // Multiple arguments
input("Enter name: ")    // Read from stdin
len(array)               // Array/object/string length
append(array, value)     // Add to array
type(value)              // Get type name
tonumber("42")           // Convert to number
```

### AI Integration (CLI only)
```duso
// Single-shot Claude call
response = claude("What is 2+2?")

// Multi-turn conversation
agent = conversation(system = "You are a helpful assistant.", tokens = 1024)
agent.system("Update system prompt")
agent.model("claude-haiku-4-5-20251001")  // Change model
response = agent.prompt("Hello!")
print(agent.usage())  // Get token usage
```

### Type Coercion
```duso
print("Value: " + 42)    // Automatic string conversion
if [] then print("array is truthy") end
```

### Function Scoping
Variables defined inside a function are local to that function and don't affect outer scopes:

```duso
x = 10
function f(param)       // param is implicit local
  x = 99                // Creates local x, doesn't modify global
  y = 5                 // Also local to function
  return x + y
end

f(1)
print(x)                // Still 10 (global unchanged)
print(y)                // Error: y not defined (was local to function)

for i = 1, 3 do         // i is implicit local to for loop
  print(i)
end
print(i)                // Error: i not defined
```

## See Also

- `script/examples/` - Example scripts
- `pkg/script/README.md` - Language reference
- `script/implementation-notes.md` - Implementation details
