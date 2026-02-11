# Learning Duso

Welcome to Duso! This guide walks you through the language fundamentals and shows you how to write scripts.

## Running Your First Script

The simplest way to get started: download a Duso binary and run a script file.

```bash
duso script.du
```

Or write a quick script from scratch:

```bash
echo 'print("Hello, Duso!")' > hello.du
duso hello.du
```

You can also run Duso in interactive mode (REPL):

```bash
duso -repl
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

See [`//` and `/* */`](/docs/reference/comments.md) for details.

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

- **[Number](/docs/reference/number.md)** - Floating point (used for arithmetic and counting)
- **[String](/docs/reference/string.md)** - Text (with templates and multiline support)
- **[Boolean](/docs/reference/boolean.md)** - `true` or `false`
- **[Array](/docs/reference/array.md)** - Ordered lists
- **[Object](/docs/reference/object.md)** - Key-value maps
- **[Function](/docs/reference/function.md)** - Callable blocks of code
- **[Nil](/docs/reference/nil.md)** - No value

Check a value's type with [`type()`](/docs/reference/type.md):

```duso
print(type(42))           // "number"
print(type("hello"))      // "string"
print(type([1, 2, 3]))    // "array"
```

See [`print()`](/docs/reference/print.md) to output values.

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

See [`if`](/docs/reference/if.md) for full details.

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

Skip iterations with [`continue`](/docs/reference/continue.md) or exit early with [`break`](/docs/reference/break.md):

```duso
for i = 1, 10 do
  if i == 2 then continue end
  if i == 8 then break end
  print(i)  // Prints 1, 3-7
end
```

See [`for`](/docs/reference/for.md) and [`while`](/docs/reference/while.md) for loop details.

## Working with Data

### Arrays

Arrays are ordered lists, 0-indexed and mutable:

```duso
nums = [10, 20, 30]
print(nums[0])           // 10
print(len(nums))         // 3

// Mutable operations - modify array in place
push(nums, 40)           // Add to end
print(nums)              // [10 20 30 40]

// Functional operations - return new array
doubled = map(nums, function(x) return x * 2 end)
print(doubled)           // [20 40 60 80]
```

Add/remove elements with [`push()`](/docs/reference/push.md), [`pop()`](/docs/reference/pop.md), [`shift()`](/docs/reference/shift.md), and [`unshift()`](/docs/reference/unshift.md).

Transform arrays with [`map()`](/docs/reference/map.md), [`filter()`](/docs/reference/filter.md), [`sort()`](/docs/reference/sort.md), and [`reduce()`](/docs/reference/reduce.md).

See [Array Type Reference](/docs/reference/array.md).

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

#### Objects with Methods

Objects can contain functions that act as methods. When you call a method with dot notation (`obj.method()`), the object is automatically bound, and the method can access the object's properties:

```duso
agent = {
  name = "Alice",
  age = 30,
  greet = function(msg)
    return msg + ", I am " + name + " (age " + age + ")"
  end,
  birthday = function()
    age = age + 1  // Modifies the object's age property
  end
}

print(agent.greet("Hello"))  // "Hello, I am Alice (age 30)"
agent.birthday()
print(agent.greet("Hello"))  // "Hello, I am Alice (age 31)"
```

Methods can also call other methods on the same object:

```duso
worker = {
  tasks_done = 0,
  do_task = function()
    tasks_done = tasks_done + 1
  end,
  do_two_tasks = function()
    do_task()  // Call another method
    do_task()
  end,
  status = function()
    return "Completed " + tasks_done + " tasks"
  end
}

worker.do_two_tasks()
print(worker.status())  // "Completed 2 tasks"
```

The same methods can work with different objects through the constructor pattern—each instance has its own properties while sharing method definitions.

#### Objects as Constructors (Blueprints)

Objects can also act as constructors (blueprints):

```duso
config = {timeout = 30, retries = 3}
config1 = config()              // Creates new copy with defaults
config2 = config(timeout = 60)  // Override specific field
```

This pattern is useful for creating multiple instances with shared defaults.

#### Arrays as Constructors

Arrays also work as constructors, creating shallow copies with optional elements appended via positional arguments:

```duso
template = [1, 2, 3]
copy = template()           // [1, 2, 3]
extended = template(4, 5)   // [1, 2, 3, 4, 5]
```

Arrays can only be called with positional arguments (variadic append)—unlike objects, which support named overrides.

#### Copying Data: Shallow vs Deep

When you create a copy with the constructor pattern (`obj()` or `arr()`), you get a **shallow copy**—nested structures are shared:

```duso
original = {scores = [10, 20, 30]}
copy = original()
copy.scores[0] = 999  // Modifies both original and copy!
print(original.scores[0])  // 999
```

For **deep copies** where nested structures are independent, use [`deep_copy()`](/docs/reference/deep_copy.md):

```duso
original = {scores = [10, 20, 30]}
copy = deep_copy(original)
copy.scores[0] = 999  // Only affects the copy
print(original.scores[0])  // 10 (unchanged)
```

**Note:** `deep_copy()` removes functions from objects. Functions don't work out of scope and their closures won't be valid, so they're excluded. This is a safety feature to prevent stale closures from leaking across scope boundaries.

Use [`keys()`](/docs/reference/keys.md) and [`values()`](/docs/reference/values.md) to extract object contents:

```duso
config = {host = "localhost", port = 8080}
print(keys(config))    // [host port]
print(values(config))  // [localhost 8080]
```

See [Object Type Reference](/docs/reference/object.md).

## Functions

### Defining Functions

Define functions with the [`function`](/docs/reference/function.md) keyword and return values with [`return`](/docs/reference/return.md):

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

See [Function Type Reference](/docs/reference/function.md).

### Closures

Functions capture their surrounding scope at definition time.

```duso
function makeAdder(n)
  function add(x)
    return x + n  // Captures n from outer scope at definition time
  end
  return add
end

addFive = makeAdder(5)
print(addFive(10))       // 15
print(addFive(20))       // 25
```

Variables captured from the outer scope remain live:

```duso
function makeCounter()
  var count = 0
  return function()
    count = count + 1  // Modifies the captured variable
    return count
  end
end

counter = makeCounter()
print(counter())  // 1
print(counter())  // 2
print(counter())  // 3
```

Each closure maintains its own captured environment.

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

See [String Type Reference](/docs/reference/string.md) for more details.

## Error Handling

### Catching Errors

Use [`try`](/docs/reference/try.md) and [`catch`](/docs/reference/catch.md) to handle errors gracefully:

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

Use [`throw()`](/docs/reference/throw.md) to raise an error from your code:

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

Use [`breakpoint()`](/docs/reference/breakpoint.md) to pause execution and inspect state (when running with `-debug` flag):

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

Use [`watch()`](/docs/reference/watch.md) to monitor expressions and break when they change:

```duso
count = 0
for i = 1, 100 do
  count = count + 1
  watch("count")  // Breaks when count changes
end

// Watch multiple expressions
watch("x", "y > 5", "len(items)")  // Monitors x, the boolean x > 5, and array length
```

See [`breakpoint()`](/docs/reference/breakpoint.md) and [`watch()`](/docs/reference/watch.md) for full details on debugging.

## Variable Scope

By default, assignment looks up the scope chain. Use [`var`](/docs/reference/var.md) to explicitly create a local variable:

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
claude = require("claude")
response = claude.prompt("What is 2 + 2?")
```

Or execute a script in your current scope with `include()`:

```duso
include("helpers.du")
result = helper_function()  // Now available
```

Modules are cached—subsequent requires return the same value.

See [`require()`](/docs/reference/require.md) and [`include()`](/docs/reference/include.md) for details.

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

The `claude` module makes it easy to orchestrate multi-step AI workflows. See [Claude Module Documentation](/contrib/claude/claude.md) for full details.

### Making HTTP Requests

Make HTTP requests with the `fetch()` builtin:

```duso
// Simple GET request
response = fetch("https://api.example.com/users")
if response.ok then
  data = response.json()
  print(data)
end

// POST with data
result = fetch("https://api.example.com/users", {
  method = "POST",
  headers = {["Content-Type"] = "application/json"},
  body = format_json({name = "Alice", age = 30})
})
```

The `fetch()` function provides a JavaScript-style API for making HTTP requests. Connection pooling is handled automatically by the runtime.

See [fetch() reference](/docs/reference/fetch.md) for full details.

### Building HTTP Servers

Create HTTP servers with the [`http_server()`](/docs/reference/http_server.md) builtin:

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

See [`http_server()` reference](/docs/reference/http_server.md) for full details.

### Running Scripts

Run other scripts synchronously with [`run()`](/docs/reference/run.md) or asynchronously with [`spawn()`](/docs/reference/spawn.md):

**Synchronous execution (blocking):**

```duso
result = run("processor.du", {data = [1, 2, 3]})
print("Result: " + format_json(result))
```

**Asynchronous execution (fire-and-forget):**

```duso
pid1 = spawn("worker1.du", {data = things})
pid2 = spawn("worker2.du", {data = things})

// Main script continues immediately
print("Spawned workers with PIDs: " + pid1 + ", " + pid2)
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

For scripts that spawn multiple workers, use [`datastore()`](/docs/reference/datastore.md) for safe coordination without shared memory:

```duso
// Orchestrator: spawn 5 workers
store = datastore("job_123")
store.set("completed", 0)

pids = []
for i = 1, 5 do
  pid = spawn("worker.du", {job_id = "job_123", worker_num = i})
  push(pids, pid)
end

// Wait for all workers to finish
store.wait("completed", 5)
print("All workers done! PIDs were: " + format_json(pids))
```

```duso
// worker.du - each spawned script
ctx = context()
job_id = ctx.request().job_id

store = datastore(job_id)
store.increment("completed", 1)  // Atomic operation
```

Datastores are **thread-safe in-memory key/value stores** that support:

- **Atomic operations**: `increment()`, `push()` - no race conditions
- **Waiting**: `wait(key, value)` - efficient blocking until value changes
- **Persistence**: Optional JSON save/load for recovery
- **Namespaced**: Each namespace is independent, preventing collision

This pattern scales from 2 workers to 1000 workers with the same clean code. The datastore handles all concurrency - no locks needed in your scripts.

See [`datastore()`](/docs/reference/datastore.md) for full API and examples.

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

These functions work great together—see [`map()`](/docs/reference/map.md), [`filter()`](/docs/reference/filter.md), and [`reduce()`](/docs/reference/reduce.md) for more examples.

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

See [`parallel()`](/docs/reference/parallel.md) for more details.

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

Use [`parse_json()`](/docs/reference/parse_json.md) to parse JSON strings and [`format_json()`](/docs/reference/format_json.md) to convert values to JSON.

## Date and Time

Work with Unix timestamps:

```duso
now_ts = now()           // Current timestamp
formatted = format_time(now_ts, "YYYY-MM-DD")
print(formatted)         // "2026-01-22"

// Parse a date string to timestamp
ts = parse_time("2026-01-22")
```

Use [`now()`](/docs/reference/now.md) to get the current timestamp, [`format_time()`](/docs/reference/format_time.md) to format timestamps, and [`parse_time()`](/docs/reference/parse_time.md) to parse date strings.

## Working with Math

Duso includes mathematical functions for basic operations, trigonometry, and advanced calculations.

### Basic Math

Common mathematical operations:

```duso
print(abs(-42))             // 42
print(sqrt(16))             // 4
print(pow(2, 3))            // 8
print(floor(3.7))           // 3
print(ceil(3.2))            // 4
print(round(3.5))           // 4
```

### Trigonometry

All trigonometric functions work with angles in radians. Use `pi()` to work with radians:

```duso
// Convert degrees to radians
degrees = 45
radians = degrees * pi() / 180

// Calculate trigonometric functions
print(sin(radians))         // ~0.707
print(cos(radians))         // ~0.707
print(tan(radians))         // ~1

// Inverse functions (return radians)
angle = atan2(1, 1)         // Angle of point (1, 1)
print(angle * 180 / pi())   // 45 degrees
```

Common uses:

```duso
// Circular motion - calculate position on a circle
radius = 100
for i in range(0, 360, 45) do
  angle = i * pi() / 180
  x = radius * cos(angle)
  y = radius * sin(angle)
  print("{{i}}°: ({{x}}, {{y}})")
end

// Find angle between two points
x1, y1 = 0, 0
x2, y2 = 3, 4
angle = atan2(y2 - y1, x2 - x1)
distance = sqrt(pow(x2 - x1, 2) + pow(y2 - y1, 2))
print("Angle: {{angle}}, Distance: {{distance}}")
```

### Exponential & Logarithmic Functions

Growth, decay, and scale calculations:

```duso
// Exponential growth
population = 1000
growth_rate = 0.05
years = 10
final_population = population * exp(growth_rate * years)
print("Population: {{final_population}}")

// Find logarithms
print(log(100))             // 2 (base 10)
print(ln(2.71828))          // ~1 (natural log)

// Inverse relationship
x = 5
print(ln(exp(x)))           // 5
```

## Quick Reference: Common Tasks

**Load a file:** [`load()`](/docs/reference/load.md)
```duso
content = load("data.txt")
```

**Save to a file:** [`save()`](/docs/reference/save.md)
```duso
save("output.json", data)
```

**List files matching pattern:** [`list_files()`](/docs/reference/list_files.md)
```duso
scripts = list_files("*.du")
backups = list_files("/STORE/*.bak")
```

> **Virtual Filesystems:** Duso also supports `/EMBED/` (for embedded resources) and `/STORE/` (for runtime-generated code). Learn more in the [Virtual Filesystems Guide](/docs/virtual-filesystem.md).

**Parse JSON:** [`parse_json()`](/docs/reference/parse_json.md)
```duso
data = parse_json(response)
```

**Convert to JSON:** [`format_json()`](/docs/reference/format_json.md)
```duso
json = format_json(data)
```

**Make an API call:** `fetch()` builtin
```duso
response = fetch("https://api.example.com/endpoint")
if response.ok then
  data = response.json()
end
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

**Transform data:** [`map()`](/docs/reference/map.md)
```duso
doubled = map(numbers, function(x) return x * 2 end)
```

**Create a copy:** Constructor pattern (shallow copy)
```duso
// Objects: copy with optional named overrides
config = {timeout = 30, retries = 3}
copy = config()           // Shallow copy
modified = config(timeout = 60)  // Copy with override

// Arrays: copy with optional positional appends
template = [1, 2, 3]
copy = template()         // Shallow copy
extended = template(4, 5)  // Copy with appended elements
```

**Deep copy a value:** [`deep_copy()`](/docs/reference/deep_copy.md)
```duso
independent = deep_copy(original)  // Independent nested copies
```

## Next Steps

- **Explore examples**: Check out `examples/core/` for feature demonstrations
- **Reference**: See [Reference Documentation](/docs/reference/index.md) for complete API docs

You can also read reference for keywordds and builtin functions using the binary itself:

```bash
duso -doc TERM
```

Or run Duso as a local web server with full documentation:

```bash
duso -docserver
```

Happy scripting!
