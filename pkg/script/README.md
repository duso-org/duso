# Duso Language Module

A lightweight, embeddable scripting language for agent orchestration. Duso is designed for LLM integration and multi-agent workflows.

## Features

- **Loosely-typed**: Implicit type coercion where sensible
- **Scoping**: Use `var` for explicit local variables, reach outer scope by default
- **Closures**: Functions capture and modify variables from outer scope
- **String templates**: `"Hello {{name}}"` syntax for LLM/JSON use cases
- **Multiline strings**: Triple quotes (`"""..."""`) for clean multi-line text
- **Functions**: First-class with closures, anonymous function expressions
- **Objects with Methods**: Lightweight OOP with automatic property access
- **Callable Objects**: Objects can act as constructors for blueprints
- **Arrays and Objects**: 0-indexed arrays, key-value objects
- **Control Flow**: if/elseif/else, while, for (numeric and iterator), break, continue
- **Exception Handling**: try/catch blocks
- **Multiline Comments**: `/* ... */` with nesting support
- **Go Bindings**: Register Go functions directly
- **No external dependencies**: Go standard library only

## Usage

### Basic Execution

`pkg/script` is the bare language and ships no builtins. Import `pkg/runtime` to get
`print()` and the rest — it registers them from its own `init()`:

```go
import (
    _ "github.com/duso-org/duso/pkg/runtime" // registers the builtins
    "github.com/duso-org/duso/pkg/script"
)

interp := script.NewInterpreter()
_, err := interp.Execute(`
  x = 5
  y = 10
  print(x + y)                      // prints 15 to stdout
`)
if err != nil {
  log.Fatal(err)
}
```

`Execute` returns `(string, error)`, but the string is always empty — script output goes
to stdout, not to the return value.

### Registering Go Functions

```go
interp.RegisterFunction("add", func(ev *script.Evaluator, args map[string]any) (any, error) {
  a := args["0"].(float64)          // positional args are keyed "0", "1", ...
  b := args["1"].(float64)          // named args are keyed by their name
  return a + b, nil
})

_, err := interp.Execute(`
  result = add(3, 4)
  print(result)
`)
```

Numbers arrive as `float64` — duso has no integer type.

### Registering Objects with Methods

```go
interp.RegisterObject("agents", map[string]script.GoFunction{
  "classify": func(ev *script.Evaluator, args map[string]any) (any, error) {
    return map[string]any{
      "confidence": 0.85,
      "category":   "positive",
    }, nil
  },
})

_, err := interp.Execute(`
  result = agents.classify("test input")
  print(result.confidence)
`)
```

## Quick Examples

### Variables & Types

```duso
x = 5                           // number
name = "Alice"                  // string
flag = true                     // boolean
arr = ["alice", "bob"]          // array (0-indexed)
obj = {timeout = 30, port = 8080} // object
```

### String Templates

```duso
name = "Alice"
age = 30
msg = "Hello {{name}}, age {{age}}"
print(msg)  // Output: Hello Alice, age 30

// Templates perfect for JSON/code (no escaping needed!)
json = "{\"user\": \"{{name}}\", \"age\": {{age}}}"
```

### Multiline Strings

```duso
// Triple quotes for multiline strings (no escaping needed!)
prompt = """
You are a helpful assistant.
Please respond in JSON format.
Be concise and accurate.
"""
print(prompt)

// Single quotes work too
doc = '''
This is a multiline string
using single quotes.
'''

// Templates work in multiline strings
name = "Bob"
message = """
Hello {{name}}!
This is a multiline string.
It supports {{1 + 1}} template expressions.
"""

// Perfect for JSON (no quote escaping!)
schema = """
{
  "type": "object",
  "properties": {
    "name": {"type": "string"},
    "age": {"type": "number"}
  }
}
"""
```

**Note on Indentation:** Multiline strings automatically remove common leading whitespace from all lines. This lets you write naturally indented code in your editor without that indentation appearing in the final string:

```
code = """
  def hello():
    return "world"
  """
```

Results in:
```
def hello():
  return "world"
```

The common indentation (2 spaces) is removed, but relative indentation (the 2-space indent of the function body) is preserved. This works with any whitespace (spaces or tabs) and is essential for generating code and writing clean prompts.

### Functions

```duso
function add(x, y)
  return x + y
end

result = add(5, 3)
print(result)  // Output: 8

// Closures
function makeAdder(n)
  function add(x) return x + n end
  return add
end
addFive = makeAdder(5)
print(addFive(10))  // Output: 15

// Function expressions
callback = function(x)
  return x * 2
end
print(callback(5))  // Output: 10
```

### Objects as Constructors

```duso
// Create an object blueprint
Config = {timeout = 30, retries = 3}

// Call it to create a new instance with defaults
config1 = Config()

// Call with overrides
config2 = Config(timeout = 60)
print(config2.timeout)  // Output: 60
```

### Objects with Methods

```duso
// Objects can have function properties (methods)
agent = {
  name = "Alice",
  skill = 90,
  greet = function(msg)
    print(msg + ", I am " + name + " with skill " + skill)
  end
}

agent.greet("Hello")  // Output: "Hello, I am Alice with skill 90"

// Create instances from blueprint
template = {
  name = "Unknown",
  describe = function()
    print("Name: " + name)
  end
}

instance = template(name = "Bob")
instance.describe()  // Output: "Name: Bob"
```

Methods automatically have access to object properties - no need for `self.` prefix.

### Control Flow

```duso
for i = 1, 10 do print(i) end              // numeric loop
for item in ["a", "b"] do print(item) end  // iterator loop
while x < 10 do x = x + 1 end              // while loop
if x > 5 then print("big") end             // if statement

// Exception handling
try
  risky_operation()
catch (error)
  print("Error: " + error)
end
```

### Comments

```duso
// Single-line comment
x = 5  // Inline comment

// Multiline comments with nesting
/* This is a
    multiline comment */

/*
  Outer comment
  /* Nested comment */
  Still in outer comment
*/
```

### Multiple Print Arguments

```duso
print("Value:", 42)              // Output: Value: 42
print("Name: " + name)           // Output: Name: Alice
```

## Built-in Functions

**Core:** `print()`, `input()`, `len()`, `type()`

**Type Conversion:** `tonumber()`, `tostring()`, `tobool()`

**String:** `upper()`, `lower()`, `substr()`, `trim()`, `split()`, `join()`, `contains()`, `replace()`

**Math:** `floor()`, `ceil()`, `round()`, `abs()`, `min()`, `max()`, `sqrt()`, `pow()`, `clamp()`

**Array/Object:** `keys()`, `values()`, `sort()`

**Date/Time:** `now()`, `format_time()`, `parse_time()`

**Utility:** `range()`

**System:** `exit()`

See the [Duso documentation](/docs/learning-duso.md) for complete reference.

## Type Coercion

- **Strings**: `"x: " + 42` → `"x: 42"`
- **Conditions**: `if 0` is false, `if ""` is false, `if []` is false, `if [1]` is true
- **Concatenation**: Any type + string coerces to string
- **Comparisons**: Strings coerce to numbers when compared with numbers: `"10" > 5` → true

## Examples

Runnable duso scripts live in `examples/` at the repository root:

- `examples/http/` - HTTP servers, routing, path params, form handling
- `examples/debug/` - Breakpoints and the interactive debugger
- `examples/htmx/` - Server-rendered UI patterns
- `examples/init/` - The project scaffolds emitted by `duso init`
- `examples/claude/`, `examples/openai/` - LLM API usage via contrib modules
- `examples/silly/` - Small odds and ends (`run.du`, `busy.du`)

Go embedding examples are in `go-embedding/` at the repository root.

## Architecture

- **Lexer** - Tokenization with string template and multiline string support
- **Parser** - Recursive descent parser producing AST
- **Evaluator** - Tree-walking interpreter
- **Value System** - Runtime value representation
- **Environment** - Scope management with closures
- **Resolver** - Compile-time parameter slot annotation
- **Public API** - `script.Interpreter` for easy integration

Builtins are not part of this package — they live in `pkg/runtime`.

## Full Language Reference

See [docs/learning-duso.md](/docs/learning-duso.md) for a guided tour of the language with examples, or [docs/internals.md](/docs/internals.md) for architecture details.
