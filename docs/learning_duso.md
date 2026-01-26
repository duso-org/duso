# Learning Duso

Welcome to Duso! This guide walks you through the language fundamentals and shows you how to write scripts. For detailed specifications, see [Language Reference](reference/index.md).

## Running Your First Script

The simplest way to get started: download a Duso binary and run a script file.

```bash
./duso script.du
```

Or write a quick script from scratch:

```bash
echo 'print("Hello, Duso!")' > hello.du
./duso hello.du
```

You can also run Duso in interactive mode (REPL):

```bash
./duso -repl
```

Type commands and see results immediately. Use `exit()` to quit.

## Variables and Types

Duso is loosely typed—variables can hold any kind of value:

```duso
name = "Alice"
age = 30
score = 95.5
active = true
skills = ["Go", "Rust", "Python"]
config = {timeout = 30, retries = 3}
```

Duso supports these types:

- **[Number](reference/number.md)** - Floating point (used for arithmetic and counting)
- **[String](reference/string.md)** - Text (with templates and multiline support)
- **[Boolean](reference/boolean.md)** - `true` or `false`
- **[Array](reference/array.md)** - Ordered lists
- **[Object](reference/object.md)** - Key-value maps
- **[Function](reference/function.md)** - Callable blocks of code
- **[Nil](reference/nil.md)** - No value

Check a value's type with [`type()`](reference/type.md):

```duso
print(type(42))           // "number"
print(type("hello"))      // "string"
print(type([1, 2, 3]))    // "array"
```

See [`print()`](reference/print.md) to output values.

## Control Flow

### If/Else

Make decisions with `if` statements:

```duso
age = 25
if age >= 18 then
  print("Adult")
elseif age >= 13 then
  print("Teenager")
else
  print("Child")
end
```

For quick conditional expressions, use the ternary operator (`?` and `:`):

```duso
status = age >= 18 ? "adult" : "minor"
```

See [`if`](reference/if.md) for full details.

### Loops

Loop through a range of numbers with `for`:

```duso
for i = 1, 5 do
  print(i)
end
```

Iterate over arrays:

```duso
items = ["apple", "banana", "cherry"]
for item in items do
  print(item)
end
```

Use `while` for condition-based loops:

```duso
count = 0
while count < 5 do
  print(count)
  count = count + 1
end
```

Skip iterations with [`continue`](reference/continue.md) or exit early with [`break`](reference/break.md):

```duso
for i = 1, 10 do
  if i == 2 then continue end
  if i == 8 then break end
  print(i)  // Prints 1, 3-7
end
```

See [`for`](reference/for.md) and [`while`](reference/while.md) for loop details.

## Working with Data

### Arrays

Arrays are ordered lists, 0-indexed:

```duso
nums = [10, 20, 30]
print(nums[0])           // 10
print(len(nums))         // 3

nums = append(nums, 40)  // Add element
print(nums)              // [10 20 30 40]
```

Use [`len()`](reference/len.md) to get the number of elements and [`append()`](reference/append.md) to add elements.

Transform arrays with [`map()`](reference/map.md), [`filter()`](reference/filter.md), and [`sort()`](reference/sort.md):

```duso
doubled = map(nums, function(x) return x * 2 end)
print(doubled)           // [20 40 60 80]
```

See [Array Type Reference](reference/array.md).

### Objects

Objects are key-value maps:

```duso
person = {
  name = "Alice",
  age = 30,
  city = "Portland"
}

print(person.name)          // "Alice"
print(person["city"])       // "Portland"
person.age = 31             // Modify
```

Objects can also act as constructors (blueprints):

```duso
Config = {timeout = 30, retries = 3}
config1 = Config()          // Creates new copy with defaults
config2 = Config(timeout = 60)  // Override specific field
```

This pattern is useful for creating multiple instances with shared defaults.

Use [`keys()`](reference/keys.md) and [`values()`](reference/values.md) to extract object contents:

```duso
config = {host = "localhost", port = 8080}
print(keys(config))    // [host port]
print(values(config))  // [localhost 8080]
```

See [Object Type Reference](reference/object.md).

## Functions

### Defining Functions

Define functions with the [`function`](reference/function.md) keyword and return values with [`return`](reference/return.md):

```duso
function greet(name)
  return "Hello, " + name
end

print(greet("World"))    // "Hello, World"
```

You can assign functions to variables:

```duso
double = function(x)
  return x * 2
end

print(double(5))         // 10
```

### Parameters and Arguments

Call functions with positional or named arguments:

```duso
function configure(timeout, retries, verbose)
  return {timeout = timeout, retries = retries, verbose = verbose}
end

configure(30, 3, true)                    // Positional
configure(timeout = 60, retries = 5)      // Named
configure(30, verbose = false)            // Mixed
```

See [Function Type Reference](reference/function.md).

### Closures

Functions capture their surrounding scope:

```duso
function makeAdder(n)
  function add(x)
    return x + n  // Captures n from outer scope
  end
  return add
end

addFive = makeAdder(5)
print(addFive(10))       // 15
print(addFive(20))       // 25
```

This is powerful for creating specialized functions.

## String Templates

One of Duso's strengths is template strings—embed expressions directly in strings:

```duso
name = "Alice"
age = 30
message = "{{name}} is {{age}} years old"
print(message)           // "Alice is 30 years old"
```

Templates work with expressions:

```duso
nums = [1, 2, 3, 4, 5]
msg = "Sum={{nums[0] + nums[1]}}"  // "Sum=3"

result = "Doubled={{2 * 5}}"       // "Doubled=10"
```

For structured text (JSON, SQL, code blocks), use multiline strings:

```duso
json = """
{
  "name": "{{name}}",
  "age": {{age}},
  "active": true
}
"""
```

Multiline strings with `"""..."""` preserve newlines and make templates clean without escaping quotes.

See [String Type Reference](reference/string.md) for more details.

## Error Handling

### Catching Errors

Use [`try`](reference/try.md) and [`catch`](reference/catch.md) to handle errors gracefully:

```duso
try
  data = load("config.json")
catch (error)
  print("Failed to load: " + error)
  data = {}
end
```

The error message is captured as a string and you can handle it however you need.

### Throwing Errors

Use [`throw()`](reference/throw.md) to raise an error from your code:

```duso
function validate(age)
  if age < 0 then
    throw("Age cannot be negative")
  end
  return age
end

try
  result = validate(-5)
catch (e)
  print("Error: " + e)
end
```

### Debugging

Use [`breakpoint()`](reference/breakpoint.md) to pause execution and inspect state (when running with `-debug` flag):

```duso
x = 42
breakpoint()  // Execution pauses here in debug mode

// Breakpoint with values (works like print())
user = {id = 123, name = "Alice"}
breakpoint("user:", user)

// Conditional breakpoint
for i = 1, 100 do
  if i == 50 then
    breakpoint("Found i={{i}}")  // Pause only at i=50
  end
end
```

Use [`watch()`](reference/watch.md) to monitor expressions and break when they change:

```duso
count = 0
for i = 1, 100 do
  count = count + 1
  watch("count")  // Breaks when count changes
end

// Watch multiple expressions
watch("x", "y > 5", "len(items)")  // Monitors x, the boolean x > 5, and array length
```

See [`breakpoint()`](reference/breakpoint.md) and [`watch()`](reference/watch.md) for full details on debugging.

## Variable Scope

By default, assignment looks up the scope chain. Use [`var`](reference/var.md) to explicitly create a local variable:

```duso
x = 10

function test()
  x = x + 1          // Modifies outer x
end

test()
print(x)             // 11

function create_local()
  var x = 100        // New local x, shadows outer
  x = x + 1
end

create_local()
print(x)             // Still 11 (outer x unchanged)
```

Function parameters and loop variables are automatically local.

## Modules and Organization

### Using Modules

Load reusable code with `require()`:

```duso
http = require("http")
response = http.fetch("https://example.com")
```

Or execute a script in your current scope with `include()`:

```duso
include("helpers.du")
result = helper_function()  // Now available
```

Modules are cached—subsequent requires return the same value.

See [`require()`](reference/require.md) and [`include()`](reference/include.md) for details.

### Working with Claude

Duso includes built-in Claude integration via a module. Load it with `require()`:

```duso
claude = require("claude")

// Simple prompt
response = claude.prompt("What is 2 + 2?")
print(response)

// Multi-turn conversation
analyst = claude.conversation(
  system = "You are a data analyst. Be concise.",
  model = "claude-opus-4-5-20251101"
)

result = analyst.prompt("Analyze this data")
```

The `claude` module makes it easy to orchestrate multi-step AI workflows. See [Claude Module Documentation](../contrib/claude/claude.md) for full details.

### Making HTTP Requests

Make HTTP requests with the `http` module:

```duso
http = require("http")

// Simple GET request
response = http.fetch("https://api.example.com/users")
print(response)

// POST with data
result = http.fetch(
  "https://api.example.com/users",
  method = "POST",
  body = format_json({name = "Alice", age = 30})
)
```

The [`http` module](../contrib/http/http.md) provides convenient functions for making HTTP requests. For lower-level control, use the [`http_client()`](reference/http_client.md) builtin to create a client with custom options.

See [HTTP Module Documentation](../contrib/http/http.md) and [`http_client()` reference](reference/http_client.md) for full details.

## Functional Programming

Duso includes functions for transforming data:

### Map

Transform each element:

```duso
nums = [1, 2, 3, 4, 5]
squared = map(nums, function(x) return x * x end)
print(squared)           // [1 4 9 16 25]
```

### Filter

Keep only matching elements:

```duso
evens = filter(nums, function(x) return x % 2 == 0 end)
print(evens)             // [2 4]
```

### Reduce

Combine elements into a single value:

```duso
sum = reduce(nums, function(acc, x) return acc + x end, 0)
print(sum)               // 15
```

Chain operations together for powerful transformations:

```duso
result = map(
  filter([1, 2, 3, 4, 5, 6], function(x) return x % 2 == 0 end),
  function(x) return x * 10 end
)
print(result)            // [20 40 60]
```

These functions work great together—see [`map()`](reference/map.md), [`filter()`](reference/filter.md), and [`reduce()`](reference/reduce.md) for more examples.

## Parallel Execution

For independent operations (like multiple API calls), use `parallel()`:

```duso
claude = require("claude")

results = parallel([
  function()
    return claude.prompt("Explain machine learning")
  end,
  function()
    return claude.prompt("Explain deep learning")
  end,
  function()
    return claude.prompt("Explain neural networks")
  end
])

print(results[0])
print(results[1])
print(results[2])
```

Each function runs concurrently. If one errors, that result becomes `nil`.

See [`parallel()`](reference/parallel.md) for more details.

## Working with JSON

Parse JSON responses from APIs:

```duso
json_str = """{"name": "Alice", "age": 30}"""
data = parse_json(json_str)
print(data.name)         // "Alice"
```

Convert Duso values to JSON:

```duso
person = {name = "Bob", age = 25}
json = format_json(person)
print(json)              // {"name":"Bob","age":25}

pretty = format_json(person, 2)  // Pretty-printed with 2-space indent
```

Perfect for working with LLM responses and APIs.

Use [`parse_json()`](reference/parse_json.md) to parse JSON strings and [`format_json()`](reference/format_json.md) to convert values to JSON.

## Date and Time

Work with Unix timestamps:

```duso
now_ts = now()           // Current timestamp
formatted = format_time(now_ts, "YYYY-MM-DD")
print(formatted)         // "2026-01-22"

// Parse a date string to timestamp
ts = parse_time("2026-01-22")
```

Use [`now()`](reference/now.md) to get the current timestamp, [`format_time()`](reference/format_time.md) to format timestamps, and [`parse_time()`](reference/parse_time.md) to parse date strings.

## Next Steps

- **Explore examples**: Check out `examples/core/` for feature demonstrations
- **CLI guide**: Read [CLI User Guide](../docs/cli/) for file I/O and advanced features
- **Reference**: See [Reference Documentation](reference/index.md) for complete API docs
- **Embedding**: If building a Go app, read [Embedding Guide](../docs/embedding/)

## Quick Reference: Common Tasks

**Load a file:** [`load()`](reference/load.md)
```duso
content = load("data.txt")
```

**Save to a file:** [`save()`](reference/save.md)
```duso
save("output.json", data)
```

**Parse JSON:** [`parse_json()`](reference/parse_json.md)
```duso
data = parse_json(response)
```

**Convert to JSON:** [`format_json()`](reference/format_json.md)
```duso
json = format_json(data)
```

**Make an API call:** `http` module
```duso
http = require("http")
response = http.fetch("https://api.example.com/endpoint")
```

**Work with Claude:** `claude` module
```duso
claude = require("claude")
response = claude.prompt("Your question here")
```

**Try/catch errors:**
```duso
try
  result = risky_operation()
catch (error)
  print("Error: " + error)
end
```

**Loop through data:**
```duso
for item in items do
  print(item)
end
```

**Transform data:** [`map()`](reference/map.md)
```duso
doubled = map(numbers, function(x) return x * 2 end)
```

Happy scripting!
