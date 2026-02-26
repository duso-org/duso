# Learning Duso

Welcome to Duso! This guide walks you through the language fundamentals and shows you how to write scripts.

## Running Your First Script

The simplest way to get started: download a Duso binary and run a script file.

```bash
duso examples/hello.du
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

### Command-Line Options

Duso supports various command-line flags for different workflows:

#### Learning & Documentation

- `-read` Browse files and docs interactively (start here for guided learning)
- `-doc TOPIC` Display formatted documentation for a module or builtin: `duso -doc datastore`
- `-docserver` Start a local webserver with searchable documentation

#### Running Scripts

- `-c CODE` Execute inline code directly: `duso -c 'print("Hello")'`
- `-repl` Start interactive REPL mode for experimenting
- `-debug` Enable interactive debugger with breakpoints and watch expressions

#### Project Setup

- `-init DIR` Create a starter project structure in a directory

#### Advanced Options

- `-config OPTS` Pass runtime configuration as `key=value,key2=value` pairs
- `-extract SRC DST` Extract files from the embedded virtual filesystem to disk
- `-lib-path PATH` Pre-pend a path to the module search path (for custom modules)
- `-no-files` Sandbox filesystem access: disables real filesystem access, restricting I/O to `/EMBED/` (read-only embedded files) and `/STORE/` (datastore virtual filesystem). Critical for running untrusted code like LLM-generated scripts.
- `-stdin-port PORT` Replace stdin/stdout with HTTP GET/POST (useful for sandboxed/containerized environments)

#### Utility
- `-help` Show the help message
- `-lsp` Start in Language Server Protocol mode (for editor integration)
- `-no-stdin` Disable stdin reading (useful for non-interactive execution)
- `-no-color` Disable ANSI color output in terminal
- `-version` Display the current Duso version

## Comments

Comments let you write notes in your code without affecting execution.

**Single-line comments** start with `//` and go to the end of the line:

```duso
// this is a comment
x = 5
print(x)    // also a comment
```

**Multi-line comments** use `/* ... */` and can span multiple lines:

```duso
/*
  This is a block comment
  that spans multiple lines
*/

print("Hello")
```

Multi-line comments support nesting, so you can comment out code that already contains comments:

```duso
/*
  commenting out a function:

  function calculate(a, b)
    // add two numbers
    return a + b
  end
*/
```

See [comments](/docs/reference/comments.md) for details.

## Variables and Types

Duso is loosely typed—variables can hold any kind of value:

```duso
// string
name = "Alice"

// number
age = 30
score = 95.5

// booleab
active = true

// array
skills = ["Go", "Rust", "Python"]

// object
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
// "number"
print(type(42))

// "string"
print(type("hello"))

// "array"
print(type([1, 2, 3]))
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
for i = 0, 4 do
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
// Prints: 1, 3-7
for i = 1, 10 do
  if i == 2 then continue end
  if i == 8 then break end
  print(i)
end
```

See [`for`](/docs/reference/for.md) and [`while`](/docs/reference/while.md) for loop details.

## Working with Data

### Arrays

Arrays are ordered lists, 0-indexed and mutable:

```duso
nums = [10, 20, 30]

// 10
print(nums[0])

// 3
print(len(nums))

// Mutable operations - modify array in place

// Add to end
push(nums, 40)

// [10 20 30 40]
print(nums)

// Functional operations - return new array

// [20 40 60 80]
doubled = map(nums, function(x) return x * 2 end)
print(doubled)
```

Add/remove elements with [`push()`](/docs/reference/push.md), [`pop()`](/docs/reference/pop.md), [`shift()`](/docs/reference/shift.md), and [`unshift()`](/docs/reference/unshift.md).

Transform arrays with [`map()`](/docs/reference/map.md), [`filter()`](/docs/reference/filter.md), [`sort()`](/docs/reference/sort.md), and [`reduce()`](/docs/reference/reduce.md).

See [Array Type Reference](/docs/reference/array.md).

### Array Utilities

Common array operations:

```duso
nums = [3, 1, 4, 1, 5, 9]

// Get length
len(nums)

// Access element at index
nums[0]

// Sort array (ascending)
nums = sort(nums)

// Sort with custom comparison (descending)
function desc(a, b)
  return a > b
end
nums = sort(nums, desc)

// Add multiple elements to end (returns new length)
new_len = push(nums, 2, 6, 5)

// Remove and return last element
last = pop(nums)

// Add elements to beginning
unshift(nums, 0)

// Remove and return first element
first = shift(nums)

// Create sequence of numbers
range(1, 5)

// With step
range(0, 10, 2)

// Descending range
range(5, 0, -1)

// Get all keys from object
keys({name = "Alice", age = 30})

// Get all values from object
values({name = "Alice", age = 30})
```

### Objects

Objects are key-value maps:

```duso
// define a simple object
person = {
  name = "Alice",
  age = 30,
  city = "Portland"
}

// "Alice"
print(person.name)

// "Portland"
print(person["city"])

// Modify
person.age = 31
```

#### Objects with Methods

Objects can contain functions that act as methods. When you call a method with dot notation (`obj.method()`), the object is automatically bound, and the method can access the object's properties:

```duso
// a more complex object with methods
agent = {
  name = "Alice",
  age = 30,
  greet = function(msg)
    // methods can see parent object properties as variables
    return msg + ", I am " + name + " (age " + age + ")"
  end,
  birthday = function()
    // modifies the object's age property
    age = age + 1
  end
}

// prints: "Hello, I am Alice (age 30)"
print(agent.greet("Hello"))

// prints: "Hello, I am Alice (age 31)"
agent.birthday()
print(agent.greet("Hello"))
```

Methods can also call other methods on the same object:

```duso
worker = {
  tasks_done = 0,
  do_task = function()
    tasks_done = tasks_done + 1
  end,
  do_two_tasks = function()
    // Call another method
    do_task()
    do_task()
  end,
  status = function()
    return "Completed " + tasks_done + " tasks"
  end
}

worker.do_two_tasks()

// "Completed 2 tasks"
print(worker.status())
```

The same methods can work with different objects through the constructor pattern—each instance has its own properties while sharing method definitions.

#### Objects as Constructors (Blueprints)

Duso's unique object model allows objects to act as constructors/factories. Call an object with `()` to create a shallow copy, optionally overriding fields:

##### Simple Data Blueprint:

```duso
config = {timeout = 30, retries = 3}

// Creates new copy with defaults
config1 = config()

// Override specific fields using named arguments
config2 = config(timeout = 60)
config3 = config(timeout = 60, retries = 5)
```

##### Objects with Methods as Blueprints:

This is particularly powerful for creating "instances" with shared behavior:

```duso
// Blueprint: method definitions + default state
Counter = {
  count = 0,
  increment = function()
    count = count + 1
  end,
  reset = function()
    count = 0
  end,
  value = function()
    return count
  end
}

// Create two independent counters from the blueprint
counter1 = Counter()
counter2 = Counter()

counter1.increment()
counter1.increment()
print(counter1.value())  // 2

counter2.increment()
print(counter2.value())  // 1 (independent!)
```

Each instance has its own copy of the state (`count`), but all share the same method definitions. This gives you object-oriented-like behavior without needing a `class` keyword.

##### Why This Pattern:
- No class/prototype overhead—objects are lightweight and transparent
- State and behavior travel together
- Easy to create variants by overriding defaults
- Familiar to functional programmers; doesn't require inheritance concepts

#### Arrays as Constructors

Arrays also work as constructors, creating shallow copies with optional elements appended via positional arguments:

```duso
// Template array
template = [1, 2, 3]

// Use as a template to make more arrays
copy = template()
extended = template(4, 5)

// copy = [1, 2, 3]
// extended = [1, 2, 3, 4, 5]
```

Arrays can only be called with positional arguments (appending elements), unlike objects which support named field overrides.

#### Copying Data: Shallow vs Deep

When you create a copy with the constructor pattern (`obj()` or `arr()`), you get a **shallow copy**—nested structures are shared:

```duso
original = {scores = [10, 20, 30]}
copy = original()

// Modifies both original and copy!
copy.scores[0] = 999

// 999
print(original.scores[0])
```

For **deep copies** where nested structures are independent, use [`deep_copy()`](/docs/reference/deep_copy.md):

```duso
original = {scores = [10, 20, 30]}
copy = deep_copy(original)

// Only affects the copy
copy.scores[0] = 999

// 10 (unchanged)
print(original.scores[0])
```

##### Important - Scope Boundaries & Functions:

Each script invoked via `spawn()` or `run()` runs in its own isolated scope. When passing objects between scopes (as arguments to spawned scripts or return values), Duso **automatically performs a deep copy** to prevent stale closures.

`deep_copy()` removes functions from objects because:
- Closures capture variables from their definition scope
- Functions from the parent scope can't safely work in a child scope
- Attempting to use stale closures would cause errors

##### Example - Why This Matters:

```duso
// Parent scope
multiply = function(x) return x * 2 end
obj = {
  value = 10,
  transform = multiply  // closure tied to parent scope
}

// Spawned child scope
worker = spawn("worker.du", {data = obj})
// data.transform is automatically stripped during deep copy
// because it can't work in the child's isolated scope
```

If your spawned script needs transformation logic, pass it as a separate function or use the datastore for coordination instead.

### Serialization Contracts: What Crosses Process Boundaries

When you pass data to `spawn()`, `run()`, or `datastore()`, Duso automatically performs a deep copy for isolation. Understanding what survives is critical for orchestration:

#### What Survives (Serializable):
- **Primitives**: Strings, numbers, booleans, nil
- **Collections**: Arrays and objects (including nested structures)
- **Regex patterns**: Converted to strings; reconstruct with `~pattern~` if needed in spawned scripts

#### What's Stripped (Not Serializable):
- **Functions**: Removed entirely (converted to nil). Closures can't work across scope boundaries.
- **File handles, connections, locks**: Any runtime resources

#### Example:

```duso
// Parent scope
worker_data = {
  name = "worker",
  tasks = [1, 2, 3],
  pattern = ~\w+~,         // OK - becomes string "\\w+"
  process = function() end // STRIPPED - becomes nil
}

// Spawn child
pid = spawn("worker.du", {data = worker_data})

// In worker.du:
ctx = context()
data = ctx.request().data
print(data.name)         // "worker" ✓
print(data.tasks)        // [1 2 3] ✓
print(data.pattern)      // "\\w+" (now a string, not regex) ✓
print(data.process)      // nil ✗ function was stripped
```

**Best Practice:** If spawned scripts need behavior, pass behavior via separate parameters or module imports, not embedded functions. Use datastore for shared coordination instead of shared state.

Use [`keys()`](/docs/reference/keys.md) and [`values()`](/docs/reference/values.md) to extract object contents:

```duso
config = {host = "localhost", port = 8080}

// [host port]
print(keys(config))

// [localhost 8080]
print(values(config))
```

See [Object Type Reference](/docs/reference/object.md).

## Functions

### Defining Functions

Define functions with the [`function`](/docs/reference/function.md) keyword and return values with [`return`](/docs/reference/return.md):

```duso
function greet(name)
  return "Hello, " + name
end

// "Hello, World"
print(greet("World"))
```

You can assign functions to variables:

```duso
double = function(x)
  return x * 2
end

// 10
print(double(5))
```

### Parameters and Arguments

Call functions with positional or named arguments:

```duso
function configure(timeout, retries, verbose)
  return {timeout = timeout, retries = retries, verbose = verbose}
end

// Positional
configure(30, 3, true)

// Named
configure(timeout = 60, retries = 5)

// Mixed
configure(30, verbose = false)
```

### Default Parameters

Function parameters can have default values, which are used when arguments are not provided:

```duso
function greet(name, greeting = "Hello", punctuation = "!")
  return greeting + " " + name + punctuation
end

// Hello Alice!
print(greet("Alice"))

// Hi Bob!
print(greet("Bob", "Hi"))

// Hey Charlie?
print(greet("Charlie", "Hey", "?"))
```

Default values work with all calling styles (positional, named, and mixed).

See [Function Type Reference](/docs/reference/function.md).

### Closures

Functions capture their surrounding scope at definition time.

```duso
function makeAdder(n)
  function add(x)
    // Captures n from outer scope at definition time
    return x + n
  end
  return add
end

addFive = makeAdder(5)

// 15
print(addFive(10))

// 25
print(addFive(20))
```

Variables captured from the outer scope remain live:

```duso
function makeCounter()
  var count = 0
  return function()
    // Modifies the captured variable
    count = count + 1
    return count
  end
end

counter = makeCounter()

// 1
print(counter())

// 2
print(counter())

// 3
print(counter())
```

Each closure maintains its own captured environment.

## String Templates

One of Duso's strengths is template strings—embed expressions directly in strings with `{{...}}`:

```duso
name = "Alice"
age = 30
message = "{{name}} is {{age}} years old"

// "Alice is 30 years old"
print(message)
```

Templates work with any expression—arithmetic, function calls, conditionals:

```duso
nums = [1, 2, 3, 4, 5]

// "Sum=3"
msg = "Sum={{nums[0] + nums[1]}}"
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

## String Utilities

Common string operations:

```duso
text = "Hello World"

// Convert to uppercase
upper(text)

// Convert to lowercase
lower(text)

// Get length
len(text)

// Extract substring (start, optional length)
substr(text, 0, 5)

// Extract from position to end
substr(text, 6)

// Negative indices from end
substr(text, -5)

// Split by delimiter
split(text, " ")

// Split into individual characters
split(text, "")

// Join array with separator
join(["Hello", "World"], "-")

// Remove leading/trailing whitespace
trim("  spaces  ")

// Remove whitespace including tabs and newlines
trim("\t hello \n")

// Replace all occurrences
replace(text, "World", "Duso")

// Case-insensitive replacement
replace(text, "hello", "hi", ignore_case = true)

// Check if contains pattern
contains(text, "World")

// Find all matches (returns array of {text, pos, len})
matches = find(text, ~\w+~)
```

See [String Type Reference](/docs/reference/string.md) for more details.

## Regular Expressions

Duso supports regular expressions using Go's regex syntax, delimited with `~...~`:

```duso
email = "alice@example.com"
if contains(email, ~\w+@\w+\.\w+~) then
  print("Valid email format")
end
```

### Finding Matches

Use [`find()`](/docs/reference/find.md) to locate all matches in a string:

```duso
text = "The years 2020, 2021, and 2022 were busy"
matches = find(text, ~\d+~)
for match in matches do
  // "2020", "2021", "2022"
  print(match.text)

  // Position in string
  print(match.pos)

  // Length of match
  print(match.len)
end
```

### Replacing with Patterns

Use [`replace()`](/docs/reference/replace.md) to replace all matches:

```duso
text = "Hello 123 World 456"
cleaned = replace(text, ~\d+~, "X")

// "Hello X World X"
print(cleaned)
```

Replace with a function to transform matches:

```duso
text = "apple, banana, cherry"
formatted = replace(text, ~\w+~, function(text, pos, len)
  // Function receives text, position, and length
  return upper(text)
end)

// "APPLE, BANANA, CHERRY"
print(formatted)
```

### Checking for Patterns

Use [`contains()`](/docs/reference/contains.md) to check if a pattern exists:

```duso
phone = "555-1234"
if contains(phone, ~\d{3}-\d{4}~) then
  print("Looks like a phone number")
end
```

Patterns are case-sensitive by default, but you can pass `true` as the third argument for case-insensitive matching:

```duso
if contains("HELLO", ~hello~, true) then
  print("Match found")
end
```

### Common Patterns

Some useful regex patterns for common tasks:

```duso
// Email-like pattern
email_pattern = ~[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}~

// Numbers
number_pattern = ~\d+(\.\d+)?~

// Whitespace
space_pattern = ~\s+~

// Word characters
word_pattern = ~\w+~

// URLs
url_pattern = ~https?://[^\s]+~
```

See [Go's regexp documentation](https://golang.org/pkg/regexp/syntax/) for the full syntax reference.

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

Use [`throw()`](/docs/reference/throw.md) to raise an error from your code. `throw()` accepts any data type—strings, objects, arrays—allowing you to create rich, app-specific error responses:

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

#### Structured Error Objects:

For orchestration and agent workflows, use objects to provide detailed error information:

```duso
function fetch_user(id)
  if id == nil then
    throw({
      code = "INVALID_ID",
      message = "User ID is required",
      status_code = 400
    })
  end
  // ... fetch logic
end

function process_batch()
  for user_id in user_ids do
    try
      user = fetch_user(user_id)
      // process user
    catch (err)
      // err can be a string, object, array, or any type
      if type(err) == "object" then
        print("Code: " + err.code + ", Message: " + err.message)
      else
        print("Error: " + err)
      end
    end
  end
end
```

This approach lets you pass domain-specific error information between functions and across process boundaries (via `run()` and `spawn()`), making error handling richer and more contextual.

### Debugging

Use [`breakpoint()`](/docs/reference/breakpoint.md) to pause execution and inspect state (when running with `-debug` flag):

```duso
x = 42

// Execution pauses here in debug mode
breakpoint()

// Breakpoint with values (works like print())
user = {id = 123, name = "Alice"}
breakpoint("user:", user)

// Conditional breakpoint
for i = 1, 100 do
  if i == 50 then
    // Pause only at i=50
    breakpoint("Found i={{i}}")
  end
end
```

Use [`watch()`](/docs/reference/watch.md) to monitor expressions and break when they change:

```duso
count = 0
for i = 1, 100 do
  count = count + 1
  // Breaks when count changes
  watch("count")
end

// Watch multiple expressions
// Monitors x, the boolean x > 5, and array length
watch("x", "y > 5", "len(items)")
```

See [`breakpoint()`](/docs/reference/breakpoint.md) and [`watch()`](/docs/reference/watch.md) for full details on debugging.

## Variable Scope

By default, assignment looks up the scope chain. Use [`var`](/docs/reference/var.md) to explicitly create a local variable:

```duso
x = 10

function test()
  // Modifies outer x
  x = x + 1
end

test()

// 11
print(x)

function create_local()
  // New local x, shadows outer
  var x = 100
  x = x + 1
end

create_local()

// Still 11 (outer x unchanged)
print(x)
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

// Now available
result = helper_function()
```

Modules are cached—subsequent requires return the same value.

#### Standard Library Modules:
- `ansi` ANSI color codes and terminal styling
- `markdown` Markdown rendering and formatting (with ANSI support for terminal output)

#### Contrib Modules:
- `claude` LLM integration for Anthropic's Claude API (prompts and multi-turn conversations). We're working on adding support for other AI vendors soon.
- `couchdb` CouchDB database integration
- `svgraph` SVG graph and visualization generation
- `zlm` Test utility that simulates LLM output without token costs (perfect for testing concurrency and worker swarms at scale)

### Module Discovery, Versioning & Contributing

#### How Modules Work:
- Modules are distributed as pure Duso code embedded in the Duso binary
- Built-in modules are frozen at release time—versions don't change once bundled
- This ensures reproducible, dependency-free deployments

#### Contributing a Module:
1. Create a repository with naming convention `duso-modulename` (e.g., `duso-postgres`)
2. Write your module in pure Duso (.du files)
3. License under Apache 2.0 with documentation and examples
4. Open an issue on the Duso repository requesting review
5. Once approved, module is added to `contrib/` and baked into the binary

For detailed contribution guidelines, see [CONTRIBUTING.md](/CONTRIBUTING.md) and [contrib/README.md](/contrib/README.md).

#### Load Custom Modules from Disk:
Use `-lib-path` to add custom module search directories:

```bash
duso -lib-path ./my_modules script.du
```

```duso
my_util = require("my_modules/utils")
```

### LLM Provider Support

#### Currently Supported:
- **Claude** (Anthropic) - via `claude` module with full API feature support

#### Planned:
- OpenAI (GPT-4, GPT-4 Turbo)
- Google Gemini
- Open-source models (Llama, Mistral via local servers)
- Custom LLM endpoints

For now, if you need other LLM providers, integrate via HTTP calls using `fetch()`:

```duso
// Custom integration example
result = fetch("https://api.openai.com/v1/chat/completions", {
  method = "POST",
  headers = {["Authorization"] = "Bearer " + env("OPENAI_API_KEY")},
  body = format_json({
    model = "gpt-4",
    messages = [{role = "user", content = "Hello"}]
  })
})
response = parse_json(result.body)
print(response.choices[0].message.content)
```

We're actively expanding the module ecosystem and LLM provider support with community contributions. If you'd like to contribute a module or suggest a provider, reach out!

See [`require()`](/docs/reference/require.md) and [`include()`](/docs/reference/include.md) for details.

### Working with Claude

Duso includes built-in Claude integration via a module. Load it with `require()`:

```duso
claude = require("claude")

// Simple prompt
response = claude.prompt("What is 2 + 2?")
print(response)

// Multi-turn conversation
analyst = claude.session(
  system = "You are a data analyst. Be concise."
)

result = analyst.prompt("Analyze this data")
```

You can also enable Claude to use tools—giving it access to functions it can call to answer questions (agent patterns):

```duso
claude = require("claude")

// Define a tool Claude can use
var web_search = {
  name = "web_search",
  description = "Search the web for information",
  input_schema = {
    type = "object",
    properties = {
      query = {type = "string", description = "Search query"}
    },
    required = ["query"]
  }
}

// Create an agent that can call your tools
agent = claude.session({
  tools = [web_search],
  tool_handlers = {
    web_search = function(input)
      // In a real app, this would call a web search API
      return "Search results for: " + input.query
    end
  }
})

// Claude will automatically call tools when needed
response = agent.prompt("What's the latest on Duso?")
print(response)
```

Claude automatically executes tool calls and integrates results into the conversation. The `claude` module makes it easy to orchestrate multi-step AI workflows with tool use loops. See [Claude Module Documentation](/contrib/claude/claude.md) for full details on tools, handlers, and manual tool control.

#### Future LLM Providers:

As we add support for OpenAI, Gemini, and other providers, they will follow the same module pattern:

```duso
// Coming soon - these will work the same way
openai = require("openai")
response = openai.prompt("Your question", {model = "gpt-4"})

gemini = require("gemini")
response = gemini.prompt("Your question", {model = "gemini-2.0"})
```

Community contributions for additional LLM providers are welcome! See the [LLM Provider Support](#llm-provider-support) section for details.

### Working with Files and Directories

Read and write files:

```duso
// Read entire file
content = load("config.json")

// Write to file (create or overwrite)
save("output.txt", "Hello, World!")

// Append to file
append_file("log.txt", "New log entry\n")

// Copy file
copy_file("source.txt", "destination.txt")

// Move/rename file
move_file("old_name.txt", "new_name.txt")
rename_file("old.txt", "new.txt")

// Delete file
remove_file("temp.txt")

// Check if file/directory exists
if file_exists("data.txt") then
  print("File found")
end

// Get file type
file_type("data.txt")

// List directory contents
files = list_dir("./data")

// Create directory (including parents)
make_dir("./output/nested/path")

// Remove empty directory
remove_dir("./empty_folder")

// Get current working directory
pwd = current_dir()
```

### Environment Variables

Access system environment and configuration:

```duso
// Read environment variable
api_key = env("API_KEY")

// Provide fallback if not set
db_host = env("DB_HOST") or "localhost"

// Read multiple settings
config = {
  host = env("HOST") or "0.0.0.0",
  port = tonumber(env("PORT") or "8080"),
  debug = env("DEBUG") == "true"
}
```

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
// Server setup mode
server = http_server({port = 8080})

// setup routes to other scripts
server.route("GET", "/", "handlers/home.du")
server.route("GET", "/api/users", "handlers/users.du")

print("Server listening on http://localhost:8080")

// Blocks until Ctrl+C
server.start()

// we're done
print("Server stopped")
```

For simple applications, a single script can be both the server setup and its own handler (self-referential pattern):

```duso
// get our script context (process info)
ctx = context()

// if we're the main script instance, start server
if ctx == nil then
  server = http_server({port = 8080})

  // setup routes to this script
  server.route("GET", "/")

  // start server and wait until it finishes
  server.start()

  // we're done, user hit Ctrl+C
  exit()
end

// if we got here, we're our own handler
// run in a separate child process

// get req/res info
req = ctx.request()
res = ctx.response()

// send back a simple HTML response
res.html("Hello World!")

// OR

// some json using duso object/arrays
res.json({
  success = true,
  data = "Hello World!"
})
```

See [`http_server()` reference](/docs/reference/http_server.md) for full details.

### Running Scripts

Run other scripts synchronously with [`run()`](/docs/reference/run.md) or asynchronously with [`spawn()`](/docs/reference/spawn.md):

#### Synchronous execution (blocking):

```duso
result = run("processor.du", {data = [1, 2, 3]})
print("Result: " + format_json(result))
```

#### Asynchronous execution (fire-and-forget):

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

// Atomic operation
store.increment("completed", 1)
```

Datastores are **thread-safe key/value stores** that support:

- **Atomic operations**: `increment()`, `push()` no race conditions
- **Waiting**: `wait(key, value)` efficient blocking until value changes
- **Persistence**: Optional disk storage for recovery across restarts
- **Namespaced**: Each namespace is independent, preventing collision

#### In-Memory vs. Disk Storage:

By default, datastores are in-memory and reset when the process exits. For persistent coordination across restarts, use disk storage:

```duso
// In-memory (default)
store = datastore("job_123")

// With optional disk persistence
store = datastore("job_123", {disk = true})
```

#### Timeouts on Blocking Calls:

All blocking operations support optional timeouts to prevent indefinite hangs:

```duso
store = datastore("jobs")

// Wait with 30-second timeout (returns nil if timeout exceeded)
result = store.wait("completed", 5, timeout = 30)

// Other blocking calls also support timeouts:
value = store.pop(timeout = 10)
value = store.shift(timeout = 10)
```

This pattern scales from 2 workers to 1000+ workers with the same clean code. The datastore handles all concurrency - no locks needed in your scripts.

See [`datastore()`](/docs/reference/datastore.md) for full API and examples.

## Functional Programming

Duso includes functions for transforming data:

### Map

Transform each element:

```duso
nums = [1, 2, 3, 4, 5]
squared = map(nums, function(x) return x * x end)

// [1 4 9 16 25]
print(squared)
```

### Filter

Keep only matching elements:

```duso
evens = filter(nums, function(x) return x % 2 == 0 end)

// [2 4]
print(evens)
```

### Reduce

Combine elements into a single value:

```duso
sum = reduce(nums, function(acc, x) return acc + x end, 0)

// 15
print(sum)
```

Chain operations together for powerful transformations:

```duso
result = map(
  filter([1, 2, 3, 4, 5, 6], function(x) return x % 2 == 0 end),
  function(x) return x * 10 end
)

// [20 40 60]
print(result)
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

// "Alice"
print(data.name)
```

Convert Duso values to JSON:

```duso
person = {name = "Bob", age = 25}
json = format_json(person)

// {"name":"Bob","age":25}
print(json)

// Pretty-printed with 2-space indent
pretty = format_json(person, 2)
```

Perfect for working with LLM responses and APIs.

Use [`parse_json()`](/docs/reference/parse_json.md) to parse JSON strings and [`format_json()`](/docs/reference/format_json.md) to convert values to JSON.

## Date and Time

Work with Unix timestamps:

```duso
// Current timestamp
now_ts = now()
formatted = format_time(now_ts, "YYYY-MM-DD")

// "2026-02-14" (example output)
print(formatted)

// Parse a date string to timestamp
ts = parse_time("2026-01-22")
```

Use [`now()`](/docs/reference/now.md) to get the current timestamp, [`format_time()`](/docs/reference/format_time.md) to format timestamps, and [`parse_time()`](/docs/reference/parse_time.md) to parse date strings.

## Working with Math

Duso includes mathematical functions for basic operations, trigonometry, and advanced calculations.

### Basic Math

Common mathematical operations:

```duso
// 42
print(abs(-42))

// 4
print(sqrt(16))

// 8
print(pow(2, 3))

// 3
print(floor(3.7))

// 4
print(ceil(3.2))

// 4
print(round(3.5))
```

### Trigonometry

All trigonometric functions work with angles in radians. Use `pi()` to work with radians:

```duso
// Convert degrees to radians
degrees = 45
radians = degrees * pi() / 180

// Calculate trigonometric functions
// ~0.707
print(sin(radians))

// ~0.707
print(cos(radians))

// ~1
print(tan(radians))

// Inverse functions (return radians)
// Angle of point (1, 1)
angle = atan2(1, 1)

// 45 degrees
print(angle * 180 / pi())

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
x1 = 0
y1 = 0
x2 = 3
y2 = 4
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
// 2 (base 10)
print(log(100))

// ~1 (natural log)
print(ln(2.71828))

// Inverse relationship
x = 5

// 5
print(ln(exp(x)))

```

## Quick Reference: Common Tasks

### Load a file: [`load()`](/docs/reference/load.md)
```duso
content = load("data.txt")
```

### Save to a file: [`save()`](/docs/reference/save.md)
```duso
save("output.json", data)
```

### List files matching pattern: [`list_files()`](/docs/reference/list_files.md)
```duso
scripts = list_files("*.du")
backups = list_files("/STORE/*.bak")
```

> **Virtual Filesystems:** Duso supports two virtual filesystems:
> - `/EMBED/` Read-only embedded resources (baked into the binary)
> - `/STORE/` Read-write virtual filesystem backed by the datastore (survives across runs if using persistent datastores)
>
> These are essential for secure execution. Use `duso -no-files` to sandbox scripts to ONLY these virtual filesystems, blocking all real filesystem and environment access. Perfect for running untrusted code (like LLM-generated scripts). Learn more in the [Virtual Filesystems Guide](/docs/virtual-filesystem.md).

### Parse JSON: [`parse_json()`](/docs/reference/parse_json.md)
```duso
data = parse_json(response)
```

### Convert to JSON: [`format_json()`](/docs/reference/format_json.md)
```duso
json = format_json(data)
```

### Make an API call: `fetch()` builtin
```duso
response = fetch("https://api.example.com/endpoint")
if response.ok then
  data = response.json()
end
```

### Work with Claude: `claude` module
```duso
claude = require("claude")
response = claude.prompt("Your question here")
```

### Try/catch errors:
```duso
try
  result = risky_operation()
catch (error)
  print("Error: " + error)
end
```

### Loop through data:
```duso
for item in items do
  print(item)
end
```

### Transform data: [`map()`](/docs/reference/map.md)
```duso
doubled = map(numbers, function(x) return x * 2 end)
```

### Create a copy: Constructor pattern (shallow copy)
```duso
// Objects: copy with optional named overrides
config = {timeout = 30, retries = 3}

// Shallow copy
copy = config()

// Copy with override
modified = config(timeout = 60)

// Arrays: copy with optional positional appends
template = [1, 2, 3]

// Shallow copy
copy = template()

// Copy with appended elements
extended = template(4, 5)
```

### Deep copy a value: [`deep_copy()`](/docs/reference/deep_copy.md)
```duso
// Independent nested copies
independent = deep_copy(original)
```

## Next Steps

- **Explore examples**: Check out `examples/core/` for feature demonstrations
- **Reference**: See [Reference Documentation](/docs/reference/index.md) for complete API docs

You can also read reference for keywords and builtin functions using the binary itself:

```bash
duso -doc TERM
```

Or run Duso as a local web server with full documentation:

```bash
duso -docserver
```

Happy scripting!
