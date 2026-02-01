---
name: Duso Scripting Language
description: A lightweight scripting language designed for LLM agent orchestration. Learn syntax, built-in functions, and Claude integration patterns.
---

# Duso Scripting Language Reference

Duso is a custom scripting language built for orchestrating LLM workflows. It features first-class Claude API integration, string templates, and clean syntax inspired by Lua.

## Quick Overview

- **File extension:** `.du`
- **Key features:** Templates with `{{}}`, closures, objects with methods, Claude API integration
- **Syntax:** Function/control flow uses `function...end`, `for...do...end`, `if...then...end`
- **No external deps:** Runs on Go stdlib only

## Core Syntax

### Variables and Types

```duso
x = 5                     // number
name = "Alice"            // string
flag = true               // boolean
items = [1, 2, 3]         // array
config = {host = "localhost", port = 8080}  // object
```

### String Templates

Use `{{expr}}` to embed expressions in strings:

```duso
name = "Alice"
age = 30
msg = "Hello {{name}}, you are {{age}} years old"

// With expressions
x = 10
result = "Sum={{x + 5}}"  // "Sum=15"
```

### Functions

```duso
function greet(name)
  return "Hello, " + name
end

result = greet("Alice")   // "Hello, Alice"

// Closures
function makeAdder(n)
  function add(x)
    return x + n
  end
  return add
end

addFive = makeAdder(5)
print(addFive(10))        // 15
```

### Control Flow

```duso
// If/Elseif/Else
if x > 10 then
  print("big")
elseif x > 5 then
  print("medium")
else
  print("small")
end

// While loop
i = 0
while i < 5 do
  print(i)
  i = i + 1
end

// For loop (numeric)
for i = 1, 5 do
  print(i)
end

// For loop (iterator)
for item in [1, 2, 3] do
  print(item)
end
```

### Objects and Methods

```duso
// Object with method
agent = {
  name = "Alice",
  skill = 90,
  greet = function(msg)
    print(msg + ", I am " + name)
  end
}

agent.greet("Hello")      // "Hello, I am Alice"

// Objects as constructors
Config = {timeout = 30, retries = 3}
config = Config(timeout = 60)  // New object with overrides
```

## Built-in Functions

### I/O
- `print(...args)` - Output values
- `input([prompt])` - Read from stdin

### String Functions
- `upper(string)` - Convert to uppercase
- `lower(string)` - Convert to lowercase
- `trim(string)` - Remove whitespace
- `substr(string, start [, length])` - Extract substring
- `split(string, separator)` - Split into array
- `join(array, separator)` - Join array to string
- `contains(string, substring [, exact])` - Check substring
- `replace(string, old, new [, exact])` - Replace all

### Array/Object Functions
- `len(value)` - Length of array, object, or string
- `append(array, value)` - Add element to array
- `keys(object)` - Get array of keys
- `values(object)` - Get array of values
- `sort(array [, compare_fn])` - Sort array

### Type Functions
- `type(value)` - Get type name
- `tonumber(value)` - Convert to number
- `tostring(value)` - Convert to string
- `tobool(value)` - Convert to boolean

### Math Functions
- `floor(n)`, `ceil(n)`, `round(n)` - Rounding
- `abs(n)` - Absolute value
- `min(...numbers)`, `max(...numbers)` - Min/max
- `sqrt(n)`, `pow(base, exp)` - Power functions
- `clamp(value, min, max)` - Clamp to range

### JSON Functions
- `parse_json(string)` - Parse JSON to objects/arrays
- `format_json(value [, indent])` - Convert to JSON string

### Date/Time Functions
- `now()` - Current Unix timestamp
- `format_time(timestamp [, format])` - Format timestamp to string
- `parse_time(string [, format])` - Parse string to timestamp

### Utility Functions
- `range(start, end [, step])` - Create number array

## Claude Integration

Load the Claude module with `require()`:

```duso
claude = require("claude")
```

### Single-Shot Query

```duso
// Simple prompt
response = claude.prompt("What is 2 + 2?")

// With model and token limit
response = claude.prompt(
  "Analyze this data",
  model = "claude-3-5-sonnet-20241022",
  tokens = 1000
)
```

### Stateful Conversation

Create multi-turn conversations that maintain context:

```duso
// Create conversation with system prompt
analyst = claude.conversation(
  system = "You are a data analyst. Be concise.",
  model = "claude-opus-4-5-20251101"
)

// Make requests - context is preserved
insight1 = analyst.prompt("Analyze this data: " + data)
insight2 = analyst.prompt("What about trends?")
insight3 = analyst.prompt("Suggest improvements")
```

### Multi-Agent Orchestration

```duso
claude = require("claude")

researcher = claude.conversation(system = "You are a researcher")
writer = claude.conversation(system = "You are a technical writer")

facts = researcher.prompt("What are 3 facts about quantum computing?")
article = writer.prompt("Write an article based on: {{facts}}")

print(article)
```

## Common Patterns

### Processing LLM JSON Responses

```duso
claude = require("claude")

prompt = "Generate JSON with {name, age, skills} for a programmer"
response = claude.prompt(prompt)

data = parse_json(response)
print(data.name)      // Access parsed fields
print(data.skills[0])
```

### Batch Processing

```duso
claude = require("claude")

tasks = ["task 1", "task 2", "task 3"]
results = []

for task in tasks do
  result = claude.prompt("Process: {{task}}")
  results = append(results, result)
end

json_output = format_json(results)
print(json_output)
```

### Configuration with Overrides

```duso
// Template
DefaultConfig = {
  timeout = 30,
  retries = 3,
  debug = false
}

// Create instance with overrides
config = DefaultConfig(timeout = 60, debug = true)
print(config.timeout)   // 60
print(config.retries)   // 3 (default)
```

### Error Handling

```duso
claude = require("claude")

try
  response = claude.prompt("risky prompt")
  print(response)
catch (error)
  print("Error: " + error)
end
```

## Named Arguments

Functions support both positional and named arguments:

```duso
function configure(timeout, retries, verbose)
  // ...
end

// Positional
configure(30, 3, true)

// Named (any order)
configure(retries = 5, verbose = true, timeout = 60)

// Mixed
configure(30, verbose = true)
```

## Key Language Features

### String Handling
- `{{expr}}` for templates (evaluates expressions inline)
- Multiline strings with `"""..."""` (strips leading/trailing whitespace)
- Concatenation with `+` (auto-coerces types)

### Scope and Variables
- Use `var` to create local variables explicitly
- Without `var`, assignment modifies outer scope if variable exists
- Function parameters and loop variables are implicitly local

### Type Coercion
- `+` coerces both operands to strings
- Comparisons coerce strings to numbers when possible
- Falsy values: `nil`, `false`, `0`, `""`, `[]`, `{}`
- Truthy: everything else

### Control Flow Keywords
- `if...then...end`, `elseif...then...end`, `else...end`
- `while...do...end`
- `for i = start, end [, step] do...end` (numeric)
- `for item in array do...end` (iterator)
- `break`, `continue` (in loops)
- `return` (in functions)

## API Key Configuration

Set your Claude API key as an environment variable:

```bash
export ANTHROPIC_API_KEY=sk-ant-...
duso script.du
```

Or pass directly to the claude module:

```duso
claude = require("claude")
response = claude.prompt("prompt", key = "sk-ant-...")
```

## Full Documentation

For complete details, see:
- **[docs/learning-duso.md](/docs/learning-duso.md)** - Guided tour of the language with examples
- **[docs/internals.md](/docs/internals.md)** - Architecture and runtime design
