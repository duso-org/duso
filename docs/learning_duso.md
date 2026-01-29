# Learning Duso

Welcome to Duso! This guide walks you through the language fundamentals and shows you how to write scripts.

**Quick Navigation:**
- **[Built-in Functions Reference](reference/index.md)** – Quick lookup for all functions
- **[Internals](internals.md)** – Architecture, design decisions, and runtime details
- **[Language Reference](reference/index.md)** – Detailed specifications

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

## Comments

Comments let you write notes in your code without affecting execution.

**Single-line comments** start with `//` and go to the end of the line:

```duso
// This is a comment
x = 5  // Inline comment
print(x)
```

**Multi-line comments** use `/* ... */` and can span multiple lines:

```duso
/* This is a block comment
   that spans multiple lines */
print("Hello")
```

Multi-line comments support nesting, so you can comment out code that already contains comments:

```duso
/*
  Commenting out a function:

  function calculate(a, b)
    // Add two numbers
    return a + b
  end
*/
```

See [`//` and `/* */`](reference/comments.md) for details.

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

One of Duso's strengths is template strings—embed expressions directly in strings with `{{...}}`:

```duso
name = "Alice"
age = 30
message = "{{name}} is {{age}} years old"
print(message)           // "Alice is 30 years old"
```

Templates work with any expression—arithmetic, function calls, conditionals:

```duso
nums = [1, 2, 3, 4, 5]
msg = "Sum={{nums[0] + nums[1]}}"      // "Sum=3"
status = "Age: {{age >= 18 ? "adult" : "minor"}}"
```

### Multiline Strings

For longer text, use triple quotes `"""..."""` to preserve newlines:

```duso
doc = """
This is a multiline string.
It preserves newlines naturally.
No escaping needed!
"""
```

Indentation is automatically handled—extra indents that match your code are filtered out:

```duso
function format_output(name, data)
  return """
    User: {{name}}
    Status: {{data.status}}
    Score: {{data.score}}
  """
end
```

The leading spaces from the code indentation are removed from the final string.

### Templates with Multiline Strings

Combine templates with multiline strings for structured text like JSON or Markdown:

```duso
name = "Alice"
score = 95

// Generate Markdown
report = """
# Report for {{name}}

Score: {{score}}
Grade: {{score >= 90 ? "A" : score >= 80 ? "B" : "C"}}

Generated at: {{format_time(now(), "iso")}}
"""
print(report)
```

Perfect for generating JSON, SQL, HTML, Markdown, or any structured text without escaping quotes or worrying about newlines.

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

### Building HTTP Servers

Create HTTP servers with the [`http_server()`](reference/http_server.md) builtin:

```duso
ctx = context()

if ctx == nil then
  // Server setup mode
  server = http_server({port = 8080})
  server.route("GET", "/", "handlers/home.du")
  server.route("GET", "/api/users", "handlers/users.du")

  print("Server listening on http://localhost:8080")
  server.start()  // Blocks until Ctrl+C

  print("Server stopped")
end

// Handler code only runs during requests (ctx != nil)
```

For simple applications, a single script can be both the server setup and its own handler (self-referential pattern):

```duso
ctx = context()

if ctx == nil then
  server = http_server({port = 8080})
  server.route("GET", "/")  // Uses current script as handler
  server.start()
end

// This runs for each request
req = ctx.request()
ctx.response({
  "status" = 200,
  "body" = "Hello from " + req.path,
  "headers" = {"Content-Type" = "text/plain"}
})
```

Each request runs in a separate goroutine with a fresh evaluator, providing true concurrent request handling. Routes support prefix matching and flexible method specifications (`"GET"`, `"POST"`, `["GET", "POST"]`, `"*"`, or `nil` for all methods).

See [`http_server()` reference](reference/http_server.md) for full details.

### Running Scripts

Run other scripts synchronously with [`run()`](reference/run.md) or asynchronously with [`spawn()`](reference/spawn.md):

**Synchronous execution (blocking):**

```duso
result = run("processor.du", {data = [1, 2, 3]})
print("Result: " + format_json(result))
```

**Asynchronous execution (fire-and-forget):**

```duso
spawn("worker1.du", {data = things})
spawn("worker2.du", {data = things})

// Main script continues immediately
print("workers are running in background")
```

### Returning Values

Scripts use `exit()` to return values:

```duso
// worker.du
exit({status = "done", value = 42})
```

This works in all contexts:
- HTTP handlers: `exit(response_object)` sends HTTP response
- `run()` scripts: `exit(value)` becomes the return value
- `spawn()` scripts: `exit(value)` completes the script

### Gate Pattern

A single script can work both standalone and as a handler using the gate pattern:

```duso
ctx = context()

if ctx == nil then
  // Standalone: spawn other scripts or start server
  result = run("child.du", {config = {...}})
  print("Child returned: " + format_json(result))
else
  // Handler mode: process the request/spawn context
  stack = ctx.callstack()
  print("Called from: " + stack[0].filename)
  exit({status = "done"})
end
```

Use `context().callstack()` for debugging to see the invocation chain (HTTP request, run, spawn, etc.).

### Coordinating Worker Swarms

For scripts that spawn multiple workers, use [`datastore()`](reference/datastore.md) for safe coordination without shared memory:

```duso
// Orchestrator: spawn 5 workers
store = datastore("job_123")
store.set("completed", 0)

for i = 1, 5 do
  spawn("worker.du", {job_id = "job_123", worker_num = i})
end

// Wait for all workers to finish
store.wait("completed", 5)
print("All workers done!")
```

```duso
// worker.du - each spawned script
ctx = context()
job_id = ctx.request().job_id

store = datastore(job_id)
store.increment("completed", 1)  // Atomic operation
```

Datastores are **thread-safe in-memory key/value stores** that support:

- **Atomic operations**: `increment()`, `append()` - no race conditions
- **Waiting**: `wait(key, value)` - efficient blocking until value changes
- **Persistence**: Optional JSON save/load for recovery
- **Namespaced**: Each namespace is independent, preventing collision

This pattern scales from 2 workers to 1000 workers with the same clean code. The datastore handles all concurrency - no locks needed in your scripts.

See [`datastore()`](reference/datastore.md) for full API and examples.

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
