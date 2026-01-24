# Duso Language Specification

Complete specification for the Duso scripting language.

## Overview

A lightweight, loosely-typed scripting language designed for agent orchestration and LLM integration. Built on Go standard library with no external dependencies.

**Design Principles:**
- Simplicity over features
- Implicit type coercion where sensible
- Templates for LLM/JSON use cases
- Easy Go function integration
- Clean error messages

## Syntax Summary

```duso
// Variables and assignment
x = 5
name = "Alice"
flag = true

// Objects (can be used as constructor blueprints)
Config = {timeout = 30, retries = 3}
config = Config(timeout = 60)  // Creates new object with overrides

// Control flow
if condition then
  statements
elseif other_condition then
  statements
else
  statements
end

while condition do
  statements
end

for i = 1, 10 do
  statements
end

for item in array do
  statements
end

// Functions
function process(input, config)
  result = compute(input)
  return result
end

// Exception handling
try
  risky_operation()
catch (error)
  print("Error=" + error)
end

// Arrays and objects
experts = ["alice", "bob", "charlie"]
config = {threshold = 0.8, max_retries = 3}

// String templates
message = "Hello {{name}}, you are {{age}} years old"
result = "Sum={{x + y}}"

// Multiline strings (with templates)
prompt = """
Hello {{name}}!
This is a multiline string.
"""
```

## Types

All values are loosely typed. At runtime, values have one of these types:

- **nil** - No value
- **number** - Float64 (0-based indexing, arithmetic)
- **string** - Text (concatenation with `+`, templates with `{{}}`)
- **boolean** - `true` or `false`
- **array** - Ordered list, 0-indexed, supports iteration
- **object** - Key-value map, property/bracket access
- **function** - Callable with parameters, supports closures

## Variables

Variables are assigned with `=` and are dynamically typed:

```duso
x = 5                // number
name = "Alice"       // string
flag = true          // boolean
arr = [1, 2, 3]      // array
obj = {key = "val"}   // object
```

Variables are scoped to their environment. Use `var` to explicitly create a local variable:

```duso
x = 10
function test()
  var x = 0        // New local x, shadows outer x
  x = x + 1
end
test()
print(x)           // Still 10 (outer x unchanged)

function modify()
  x = x + 5        // No var = modifies outer x
end
modify()
print(x)           // Now 15
```

**Scoping Rules:**
- Assignment without `var` walks up the scope chain: if a variable exists in any parent scope, it modifies that variable; otherwise, it creates a local
- Assignment with `var` always creates a new local variable, shadowing any outer variables with the same name
- Function and block scopes allow assignment to reach parent scopes
- For loop variables (like `i` in `for i = 1, 10`) are local to the loop

## Objects as Constructors

Objects can be called like functions to create new copies with optional field overrides. This pattern is useful for creating "blueprints" of objects with default values.

**Basic pattern:**

```duso
// Define an object that will serve as a constructor
Config = {timeout = 30, retries = 3, threshold = 0.8}

// Call it to create a new copy with all defaults
config1 = Config()

// Call it with overrides to customize
config2 = Config(timeout = 60)

// Override multiple fields
config3 = Config(timeout = 60, retries = 5)

// Access fields
print(config2.timeout)  // Output: 60
```

**How it works:**
- Any object can be called like a function: `obj()`
- Calling an object creates a new object (a copy)
- You can pass named arguments to override specific fields: `obj(field = value)`
- Only named arguments are supported when calling objects (no positional arguments)
- Unspecified fields keep their original values from the source object

## Objects with Methods

Objects can have function properties that act as methods. When a method is called, it automatically has access to the object's properties through implicit variable lookup:

```duso
// Create an object with a method
agent = {
  name = "Alice",
  skill = 90,
  greet = function(msg)
    print(msg + ", I am " + name + " with skill " + skill)
  end
}

agent.greet("Hello")  // Output: "Hello, I am Alice with skill 90"
```

When `agent.greet()` is called, the function can reference `name` and `skill` directly without needing `self.` or `agent.`. The method automatically has access to the object's properties as if they were in its scope.

**Creating Instances with Overrides:**

```duso
// Template/blueprint
blueprint = {
  name = "Unknown",
  skill = 0,
  describe = function()
    print(name + " has skill " + skill)
  end
}

// Create instances with different values
agent1 = blueprint(name = "Bob", skill = 85)
agent2 = blueprint(name = "Charlie", skill = 95)

agent1.describe()  // Output: "Bob has skill 85"
agent2.describe()  // Output: "Charlie has skill 95"
```

This pattern combines blueprints (objects as constructors) with methods for lightweight, composition-friendly object-oriented programming.

## Arrays

Arrays are 0-indexed ordered lists:

```duso
// Array literal
arr = [1, 2, 3, "mixed", true]

// Access (0-indexed)
print(arr[0])    // Output:1
print(arr[3])    // Output: "mixed"

// Length
print(len(arr))  // Output:5

// Iteration
for item in arr do
  print(item)
end

// Append
arr = append(arr, 4)
```

Key points:
- 0-based indexing (first element is `[0]`)
- Mixed types allowed
- Iteration with `for item in array do`
- `len()` returns array length
- `append(array, value)` adds element

## Objects

Objects are key-value maps:

```duso
// Object literal (keys are identifiers)
config = {timeout = 30, retries = 3, name = "prod"}

// Property access
print(config.timeout)      // Output:30

// Bracket access
print(config["timeout"])   // Output:30

// Length
print(len(config))         // Output:3

// Assignment
config.timeout = 60
config["retries"] = 5
```

Key points:
- Keys must be identifiers in literal syntax
- Access via dot notation or brackets
- `len()` returns number of keys
- Any type can be a value

## Strings

Strings are text values with several features:

### String Literals

```duso
s1 = "double quotes"
s2 = 'single quotes'
```

### Escape Sequences

```duso
s = "Line 1\nLine 2"    // \n newline
s = "Tab\there"         // \t tab
s = "Quote=\"hi\""     // \" quote
s = "Slash=\\"         // \\ backslash
s = "Brace=\{"         // \{ literal brace
```

### Multiline Strings

Two approaches for multiline text:

**Using escape sequences:**
```duso
s = "Line 1\nLine 2\nLine 3"
```

**Using triple quotes (recommended for clean multi-line text):**
```duso
// Double quotes
prompt = """
You are a helpful assistant.
Please respond in JSON format.
Be concise and accurate.
"""

// Single quotes (same behavior)
doc = '''
This is a multiline string
using single quotes.
'''
```

**Features:**
- Triple quotes (`"""..."""` or `'''...'''`)
- Actual newlines are preserved
- Leading and trailing whitespace is automatically stripped
- Templates work inside multiline strings
- Perfect for JSON, SQL, code, and other structured text without escaping quotes

**With templates:**
```duso
name = "Alice"
prompt = """
Hello {{name}}!
This is a multiline string.
It supports {{1 + 1}} template expressions.
"""

// Great for JSON (no quote escaping!)
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

### String Templates

Templates embed expressions using `{{expr}}` syntax:

```duso
// Variables
name = "Alice"
age = 30
msg = "Hello {{name}}, you are {{age}} years old"
print(msg)  // Output:Hello Alice, you are 30 years old

// Expressions
x = 10
y = 20
msg = "Sum={{x + y}}"
print(msg)  // Output:Sum=30

// Property access
config = {timeout = 30}
msg = "Timeout={{config.timeout}}"

// Array indexing
arr = [1, 2, 3]
msg = "First={{arr[0]}}"

// Function calls
function double(n) return n * 2 end
msg = "Result={{double(5)}}"

// Complex expressions
msg = "Calc={{(x + y) * 2}}"
```

**Key points:**
- Templates use `{{expr}}` (double braces)
- Expressions are evaluated in current scope
- Result is converted to string automatically
- Single braces `{` and `}` work without escaping in normal text
- Use `\{` and `\}` only if you need literal single braces

**Perfect for LLM/JSON:**
```duso
// JSON works naturally (no escaping needed)
json = "{\"user\": \"{{username}}\", \"age\": {{user_age}}}"

// Code works naturally
code = "function greet() { return \"Hello {{name}}\"; }"
```

## String Concatenation

Strings concatenate with the `+` operator:

```duso
msg = "Hello " + "World"           // Output: "Hello World"
msg = "Count=" + 42               // Output: "Count=42"
msg = "Array=" + [1, 2, 3]        // Output: "Array=[1 2 3]"
```

The `+` operator coerces operands to strings (see Type Coercion).

## Operators

### Arithmetic

- `+` - Addition (numbers) or concatenation (strings)
- `-` - Subtraction
- `*` - Multiplication
- `/` - Division
- `%` - Modulo

```duso
print(10 + 5)   // Output:15
print(10 - 5)   // Output:5
print(10 * 5)   // Output:50
print(10 / 5)   // Output:2
print(10 % 3)   // Output:1
```

### Comparison

- `==` - Equal
- `!=` - Not equal
- `<` - Less than
- `>` - Greater than
- `<=` - Less than or equal
- `>=` - Greater than or equal

```duso
print(5 == 5)      // Output:true
print(5 != 3)      // Output:true
print(5 < 10)      // Output:true
print("a" < "b")   // Output:true
```

### Logical

- `and` - Logical AND (short-circuit evaluation)
- `or` - Logical OR (short-circuit evaluation)
- `not` - Logical NOT

```duso
if x > 5 and y < 10 then print("both") end
if x < 0 or x > 100 then print("out of range") end
if not flag then print("flag is false") end
```

### Assignment

- `=` - Assign value to variable
- `+=` - Add and assign (shorthand for `x = x + value`)
- `-=` - Subtract and assign
- `*=` - Multiply and assign
- `/=` - Divide and assign
- `%=` - Modulo and assign
- `++` - Increment (post-increment as statement)
- `--` - Decrement (post-decrement as statement)

```duso
x = 5
config.timeout = 30
arr[0] = 100

// Compound assignments
x += 3        // x is now 8
x -= 2        // x is now 6
x *= 2        // x is now 12
x /= 3        // x is now 4
x %= 3        // x is now 1

// Increment/Decrement (as statements)
counter = 10
counter++     // counter is now 11
counter--     // counter is now 10
```

## Control Flow

### If/Elseif/Else

```duso
if x > 10 then
  print("big")
elseif x > 5 then
  print("medium")
else
  print("small")
end
```

### Conditional Expression (Ternary)

```duso
// condition ? true_value : false_value
age = 25
status = age >= 18 ? "adult" : "minor"
print(status)  // Output: adult

// Nested ternary
score = 85
grade = score >= 90 ? "A" : score >= 80 ? "B" : score >= 70 ? "C" : "F"
print(grade)  // Output: B

// In expressions
result = x > 5 ? 10 : 20
price = is_member ? total * 0.9 : total
```

The ternary operator evaluates the condition (must be truthy/falsy), and returns either the true_value or false_value. It has lower precedence than comparison/logical operators.

### While Loop

```duso
i = 0
while i < 5 do
  print(i)
  i = i + 1
end
```

### For Loop (Numeric)

```duso
// Count from 1 to 10
for i = 1, 10 do
  print(i)
end

// Count from 1 to 10 by 2s
for i = 1, 10, 2 do
  print(i)
end

// Count backward
for i = 10, 1, -1 do
  print(i)
end
```

**Note:** Loop bounds and step must be integers. Using non-integer floats will result in an error. For example, `for i = 1.5, 10 do` is invalid. The loop variable `i` itself is always a number type.

### For Loop (Iterator)

```duso
items = ["a", "b", "c"]
for item in items do
  print(item)
end

// Iterate over object keys
config = {host = "localhost", port = 8080}
for key in config do
  print(key)
end
```

**Note:** When iterating over arrays, the loop variable receives each element. When iterating over objects, the loop variable receives each key as a string.

### Break and Continue

Use `break` to exit a loop early, and `continue` to skip to the next iteration:

```duso
// Break example
for i = 1, 10 do
  if i == 5 then break end
  print(i)
end
// Output: 1 2 3 4

// Continue example
for i = 1, 5 do
  if i == 2 then continue end
  print(i)
end
// Output: 1 3 4 5

// Works in while loops too
i = 0
while i < 10 do
  i = i + 1
  if i == 3 then continue end
  if i == 7 then break end
  print(i)
end
// Output: 1 2 4 5 6
```

## Functions

### Definition

```duso
function greet(name)
  return "Hello, " + name
end

function add(x, y)
  return x + y
end

function no_return_value()
  print("This function returns nil")
end
```

### Calling

```duso
result = greet("Alice")   // Output: "Hello, Alice"
result = add(5, 3)        // Output:8
no_return_value()         // Output: "This function returns nil"
```

### Closures

Functions capture their definition environment:

```duso
function makeAdder(n)
  function add(x)
    return x + n  // Captures 'n'
  end
  return add
end

addFive = makeAdder(5)
print(addFive(10))  // Output:15 (10 + 5)
```

### Named Arguments

Functions can be called with named arguments:

```duso
function configure(timeout, retries, verbose)
  // ...
end

// Positional (in order)
configure(30, 3, true)

// Named (any order)
configure(retries = 5, timeout = 60, verbose = false)

// Mixed (positional first, then named)
configure(30, verbose = true)  // timeout=30, retries uses default, verbose=true
```

**Argument Application Order:**
1. Positional arguments are assigned to parameters in order
2. Named arguments are then applied by parameter name, potentially overriding positional values
3. Any parameters without values become `nil`

### Syntax Note: Equals for Consistency

Object literals, named function arguments, and object constructors all use the **`=` operator** for consistency:

- **Object literals** use `=`: `{timeout = 30, retries = 3}`
- **Named function arguments** use `=`: `configure(timeout = 60, verbose = true)`
- **Object constructor calls** use `=`: `Config(timeout = 60)`

This unified syntax makes it clear that you're binding values to named fields/parameters.

**Note on Ternary Operators:**
The ternary operator (`condition ? true_value : false_value`) uses `:` as a separator, which is a different context from object literals.

**Objects as Constructors:**
Objects can only be called with named arguments:

```duso
Config = {timeout = 30, retries = 3}
config = Config(timeout = 60)     // OK: named arguments
config = Config(60)               // ERROR: objects don't accept positional arguments
```

### Function Expressions

Functions can be assigned to variables and object properties:

```duso
// Assign to variable
greet = function(name)
  return "Hello, " + name
end

print(greet("Alice"))  // Output: "Hello, Alice"

// Assign to object property
obj = {
  callback = function(msg)
    print("Message: " + msg)
  end
}

obj.callback("Hi")  // Output: "Message: Hi"
```

Function expressions work just like named functions - they support closures and can be returned from other functions.

## Exception Handling

### Try/Catch

```duso
try
  risky_operation()
catch (error)
  print("Error occurred=" + error)
end
```

The `error` variable contains the error message as a string.

### Error Propagation

Errors propagate up through function calls unless caught:

```duso
function operation1()
  operation2()  // If this errors, it propagates
end

function operation2()
  try
    risky_operation()
  catch (e)
    print("Handled=" + e)
  end
end
```

## Modules (CLI Feature)

Duso provides a module system for organizing code into reusable components. There are two ways to load code:

### `require()` - Isolated Module Loading

The `require()` function loads a module in an isolated scope and returns its exports:

```duso
math = require("math")
result = math.add(2, 3)
```

**Characteristics:**
- Variables defined in the module are **private** to the module
- Only the returned value is accessible to the caller
- Module is **cached** - subsequent requires return the cached value
- Suitable for libraries and APIs

**Module Definition:**

A module exports its public API by returning a value (the last expression):

```duso
-- mylib.du
function add(a, b)
  return a + b
end

function multiply(a, b)
  return a * b
end

-- Export public API as an object
return {
  add = add,
  multiply = multiply
}
```

**Using the Module:**

```duso
lib = require("mylib")
print(lib.add(2, 3))          -- 5
print(lib.multiply(4, 5))     -- 20
```

**Module Return Types:**

Modules can return any value:

```duso
-- Return an object (typical)
return {add = add, multiply = multiply}

-- Return a function
return function(x) return x * 2 end

-- Return a simple value
return "module version 1.0"
```

**Caching:**

Once loaded, a module is cached. Subsequent requires return the same value:

```duso
lib1 = require("mylib")
lib2 = require("mylib")  -- Returns cached value
-- lib1 and lib2 reference the same object
```

### `include()` - Current Scope Loading

The `include()` function loads and executes a file in the current scope:

```duso
include("helpers.du")
result = helper_function()  -- Now available
```

**Characteristics:**
- Variables and functions are **visible** in the caller's scope
- Results are **not cached** - file is re-executed each time
- Suitable for configuration files and shared utilities

**Example:**

`config.du`:
```duso
apiUrl = "https://api.example.com"
timeout = 30
```

`main.du`:
```duso
include("config.du")
print(apiUrl)    -- Available in current scope
print(timeout)   -- Available in current scope
```

### Path Resolution

Both `require()` and `include()` support the same path resolution algorithm:

1. **User-provided paths** (absolute or `~/...`)
   - `require("/usr/local/duso/lib")` - Absolute path
   - `require("~/duso/modules/math")` - Home directory

2. **Relative to script directory**
   - `require("modules/math")` - Subdirectory of script

3. **DUSO_PATH environment variable**
   - Paths separated by colons (`:`) on Unix, semicolons (`;`) on Windows
   - `export DUSO_PATH=/usr/local/duso/lib:~/.duso/modules`

4. **Extension fallback**
   - If file not found, tries adding `.du` extension
   - `require("math")` finds `math.du`

### Circular Dependency Detection

Duso detects and reports circular dependencies:

```duso
-- a.du
require("b")

-- b.du
require("a")

-- Error: circular dependency detected
--   a.du
--   → b.du
--   → a.du (circular)
```

### Built-in Modules: stdlib and contrib

Duso binaries come with two categories of pre-built modules, all baked into the binary:

**stdlib** - Official standard library modules maintained by the Duso team:
```duso
http = require("http")
response = http.fetch("https://example.com")
```

**contrib** - Community-contributed modules curated by the Duso team:
```duso
claude = require("claude")
response = claude.prompt("What is Duso?")
```

**No External Dependencies:**
- All stdlib and contrib modules are baked into the Duso binary at build time
- Modules are written in pure Duso (no external package managers or dependencies)
- The binary is completely self-contained and works indefinitely

**Freezing:**
Each Duso binary is frozen at release time with a specific set of stdlib and contrib modules. This means:
- Scripts written for `duso-v0.5.2` will continue to work with that binary forever
- No version conflicts or dependency resolution issues
- Archive your scripts and binary together for permanent reproducibility

**Listing Available Modules:**
```duso
-- Check what modules are available in a specific binary
-- See contrib/ and stdlib/ in the Duso distribution
```

## Type Coercion

The language uses implicit type coercion in specific contexts:

### String Concatenation

Any type coerces to string with `+`:

```duso
print("Value=" + 42)              // "Value=42"
print("Array=" + [1, 2, 3])       // "Array=[1 2 3]"
print("Bool=" + true)             // "Bool=true"
print("Nil=" + nil)               // "Nil=nil"
```

### Boolean Context

Values have truthiness in conditions:

```duso
if 0 then print("zero") end        // false
if 1 then print("one") end         // true
if "" then print("empty") end      // false
if "hello" then print("text") end  // true
if [] then print("array") end      // false (empty array is falsy)
if [1] then print("array") end     // true (non-empty array is truthy)
if {} then print("object") end     // false (empty object is falsy)
if {a = 1} then print("object") end // true (non-empty object is truthy)
if nil then print("nil") end       // false
if false then print("false") end   // false
```

**Falsy values:** `nil`, `false`, `0`, `""` (empty string), `[]` (empty array), `{}` (empty object)

**Truthy values:** All other values, including non-empty arrays, non-empty objects, and zero in string form (`"0"` is truthy)

### Comparison Coercion

When comparing a string with a number, the string coerces to a number if possible:

```duso
print(5 < "10")       // true (string "10" coerces to number 10)
print("3.14" > 2)     // true (string "3.14" coerces to 3.14)
print("hello" < 5)    // error (non-numeric string cannot coerce)
```

This allows convenient string-to-number comparisons when working with user input or parsed data.

### Comparison

Numbers and strings can be compared:

```duso
print(5 < 10)          // true
print("apple" < "banana")  // true
```

Comparing different types results in error.

## Built-in Functions

### print(...args)

Output values separated by spaces:

```duso
print("Hello")                  // Output: "Hello"
print("x:", 42)                 // Output: "x: 42"
print(1, 2, 3)                  // Output: "1 2 3"
print([1, 2, 3])                // Output: "[1 2 3]"
```

### len(value)

Get length of array, object, or string:

```duso
print(len([1, 2, 3]))           // Output:3
print(len({a = 1, b = 2}))        // Output:2
print(len("hello"))             // Output:5
```

### append(array, value)

Add element to array (returns new array):

```duso
arr = [1, 2, 3]
arr = append(arr, 4)
print(arr)                       // Output:[1 2 3 4]
```

### type(value)

Get type name for debugging:

```duso
print(type(42))                 // Output: "number"
print(type("hello"))            // Output: "string"
print(type([1, 2]))             // Output: "array"
print(type({a = 1}))             // Output: "object"
```

### tonumber(value)

Convert value to number:

```duso
print(tonumber("42"))           // Output:42
print(tonumber("3.14"))         // Output:3.14
print(tonumber(true))           // Output:1
```

### Type Conversion Functions

**tostring(value)** - Convert value to string:
```duso
print(tostring(42))             // Output: "42"
print(tostring(true))           // Output: "true"
```

**tobool(value)** - Convert value to boolean (0 and "" are false, everything else truthy):
```duso
print(tobool(1))                // Output: true
print(tobool(0))                // Output: false
print(tobool("text"))           // Output: true
print(tobool(""))               // Output: false
```

### String Functions

**upper(string)** - Convert to uppercase:
```duso
print(upper("hello"))           // Output: "HELLO"
```

**lower(string)** - Convert to lowercase:
```duso
print(lower("HELLO"))           // Output: "hello"
```

**substr(string, start [, length])** - Extract substring:
```duso
print(substr("hello", 1, 3))    // Output: "ell"
print(substr("hello", -2))      // Output: "lo" (negative index from end)
```

**trim(string)** - Remove leading/trailing whitespace:
```duso
print(trim("  hello  "))        // Output: "hello"
```

**split(string, separator)** - Split string into array:
```duso
parts = split("a,b,c", ",")
print(parts[0])                 // Output: "a"
```

**join(array, separator)** - Join array into string:
```duso
print(join(["a", "b", "c"], "-"))  // Output: "a-b-c"
```

**contains(string, substring, exact)** - Check if string contains substring:
```duso
print(contains("hello", "HELLO"))           // Output: true (case-insensitive by default)
print(contains("hello", "HELLO", true))     // Output: false (case-sensitive)
```

The `exact` parameter defaults to `false` for case-insensitive matching. Set `exact=true` for exact/case-sensitive matching.

**replace(string, old, new, exact)** - Replace all instances of old with new:
```duso
print(replace("Hello World", "hello", "hi"))          // Output: "hi World"
print(replace("Hello hello HELLO", "hello", "HI"))    // Output: "HI HI HI"
print(replace("Hello", "hello", "hi", true))          // Output: "Hello" (no match, case-sensitive)
```

The `exact` parameter defaults to `false` for case-insensitive matching. Replaces all instances of the search string. When case-insensitive, finds matches ignoring case but preserves the original text capitalization in the replacement.

### Math Functions

**floor(number)** - Round down to nearest integer:
```duso
print(floor(3.7))               // Output: 3
```

**ceil(number)** - Round up to nearest integer:
```duso
print(ceil(3.2))                // Output: 4
```

**round(number)** - Round to nearest integer:
```duso
print(round(3.5))               // Output: 4
```

**abs(number)** - Absolute value:
```duso
print(abs(-42))                 // Output: 42
```

**min(...numbers)** - Minimum of arguments:
```duso
print(min(5, 2, 8, 1))          // Output: 1
```

**max(...numbers)** - Maximum of arguments:
```duso
print(max(5, 2, 8, 1))          // Output: 8
```

**sqrt(number)** - Square root:
```duso
print(sqrt(16))                 // Output: 4
```

**pow(base, exponent)** - Raise to power:
```duso
print(pow(2, 3))                // Output: 8
```

**clamp(value, min, max)** - Clamp value between min and max:
```duso
print(clamp(15, 10, 20))        // Output: 15
print(clamp(5, 10, 20))         // Output: 10
print(clamp(25, 10, 20))        // Output: 20
```

### Array and Object Functions

**keys(object)** - Get array of object keys:
```duso
obj = {a = 1, b = 2, c = 3}
print(keys(obj))                // Output: [a b c]
```

**values(object)** - Get array of object values:
```duso
obj = {a = 1, b = 2, c = 3}
print(values(obj))              // Output: [1 2 3]
```

**sort(array [, comparison_function])** - Sort array in ascending order, with optional custom comparison:
```duso
// Default sort (ascending)
print(sort([3, 1, 4, 1, 5]))    // Output: [1 1 3 4 5]

// Custom comparison function (descending)
function reverse_compare(a, b)
  return a > b
end

print(sort([3, 1, 4, 1, 5], reverse_compare))  // Output: [5 4 3 1 1]
```

The comparison function takes two arguments and should return `true` if the first argument comes before the second in the desired order. For ascending order, return `a < b`. For descending, return `a > b`.

### Functional Programming Functions

**map(array, function)** - Transform each element by applying a function:
```duso
numbers = [1, 2, 3, 4, 5]
doubled = map(numbers, function(x) return x * 2 end)
print(doubled)              // Output: [2 4 6 8 10]

// With a named function
function square(x)
  return x * x
end
squares = map(numbers, square)
print(squares)              // Output: [1 4 9 16 25]
```

**filter(array, function)** - Keep only elements that match a predicate:
```duso
numbers = [1, 2, 3, 4, 5, 6]
evens = filter(numbers, function(x) return x % 2 == 0 end)
print(evens)                // Output: [2 4 6]

// Keep strings longer than 3 characters
words = ["hi", "hello", "ok", "world"]
long_words = filter(words, function(w) return len(w) > 3 end)
print(long_words)           // Output: [hello world]
```

**reduce(array, function, initial_value)** - Combine all elements into a single value:
```duso
numbers = [1, 2, 3, 4, 5]
sum = reduce(numbers, function(acc, x) return acc + x end, 0)
print(sum)                  // Output: 15

// Calculate product
product = reduce(numbers, function(acc, x) return acc * x end, 1)
print(product)              // Output: 120

// Build an object
words = ["hello", "world", "duso"]
word_count = reduce(words, function(acc, word)
  acc[word] = 1
  return acc
end, {})
print(word_count)           // Output: {hello=1 world=1 duso=1}
```

**Chaining Operations:**

Functions can be chained together for powerful transformations:
```duso
data = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]

// Filter evens, then double them
result = map(
  filter(data, function(x) return x % 2 == 0 end),
  function(x) return x * 2 end
)
print(result)               // Output: [4 8 12 16 20]

// Sum of squares
sum_of_squares = reduce(
  map(data, function(x) return x * x end),
  function(acc, x) return acc + x end,
  0
)
print(sum_of_squares)       // Output: 385
```

### Parallel Execution

**parallel(functions)** - Execute functions concurrently and collect results:

Executes multiple functions in parallel, each in its own isolated evaluator context. This is ideal for independent operations like multiple API calls, data fetches, or long-running computations.

**Key characteristics:**
- True parallelism: Each function runs concurrently in its own goroutine
- Parent scope access: Functions can READ parent variables (read-only)
- Parent scope protection: Functions cannot WRITE to parent scope (assignments stay local)
- Error handling: If a function errors, that result becomes `nil`; other functions continue
- Result preservation: Results maintain order/structure as input (arrays stay arrays, objects stay objects)

**Array form** - Execute array of functions:
```duso
topic = "Duso"
language = "Go"

results = parallel([
  function()
    return """
    What is {{topic}}?
    A scripting language for agent orchestration.
    """
  end,

  function()
    return """
    What powers {{topic}}?
    Built entirely on {{language}} stdlib.
    """
  end,

  function()
    return """
    What can I build?
    Web scrapers, data pipelines, AI agents, and more.
    """
  end
])

print("""
Results:

{{results[0]}}

{{results[1]}}

{{results[2]}}
""")
```

**Object form** - Execute named functions:
```duso
user_id = 42

results = parallel({
  profile = function()
    return "User {{user_id}} profile data"
  end,

  activity = function()
    return "Activity logs for user {{user_id}}"
  end,

  recommendations = function()
    return "Personalized recommendations for user {{user_id}}"
  end
})

print("""
User {{user_id}} Summary:

Profile:
  {{results.profile}}

Activity:
  {{results.activity}}

Recommendations:
  {{results.recommendations}}
""")
```

**Practical example with Claude API**:
```duso
claude = require("claude")
topic = "machine learning"

// Make 3 concurrent API calls
results = parallel([
  function()
    return claude.prompt("Explain {{topic}} to a beginner in 2-3 sentences.")
  end,

  function()
    return claude.prompt("List 3 advanced concepts in {{topic}}.")
  end,

  function()
    return claude.prompt("Name 5 real-world applications of {{topic}}.")
  end
])

print("""
### {{topic}} Overview

**Beginner Explanation:**
{{results[0]}}

**Advanced Concepts:**
{{results[1]}}

**Real-World Applications:**
{{results[2]}}
""")
```

**Important notes:**
- Blocks execute truly in parallel - use this when operations are independent
- If an operation errors, that slot becomes `nil`; check with `if results[i] != nil then ... end`
- Parent scope is read-only; assignments in blocks create local variables
- Ideal for: independent API calls, concurrent data fetches, parallel computations
- Not suitable for: operations that need to share mutable state

### Utility Functions

**range(start, end [, step])** - Create array of numbers:
```duso
print(range(1, 5))              // Output: [1 2 3 4 5]
print(range(10, 1, -2))         // Output: [10 8 6 4 2]
```

### JSON Functions

**parse_json(string)** - Parse JSON string into Duso objects/arrays:
```duso
json_str = """{"name": "Alice", "age": 30, "skills": ["Go", "Lua"]}"""
data = parse_json(json_str)

print(data.name)                // Output: "Alice"
print(data.age)                 // Output: 30
print(data.skills[0])           // Output: "Go"
```

Perfect for processing LLM responses:
```duso
response = claude("Generate JSON with {name, description}")
result = parse_json(response)
print(result.name)
print(result.description)
```

**format_json(value [, indent])** - Convert Duso value to JSON string:
```duso
// Compact JSON (default)
data = {name = "Alice", age = 30, active = true}
json = format_json(data)
print(json)                     // Output: {"name":"Alice","age":30,"active":true}

// Pretty-printed with indentation
json_pretty = format_json(data, 2)
print(json_pretty)
// Output:
// {
//   "name": "Alice",
//   "age": 30,
//   "active": true
// }
```

Works with arrays:
```duso
arr = [{id = 1, name = "Alice"}, {id = 2, name = "Bob"}]
json = format_json(arr)
print(json)                     // Output: [{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}]
```

**Type Mapping:**
- Duso `true`/`false` → JSON `true`/`false`
- Duso `nil` → JSON `null`
- Duso numbers → JSON numbers
- Duso strings → JSON strings
- Duso arrays → JSON arrays
- Duso objects → JSON objects

Parsing JSON returns Duso objects/arrays that work seamlessly with the language.

### Date/Time Functions

All date/time functions work with Unix timestamps (seconds since epoch) represented as numbers.

**now()** - Get current Unix timestamp:
```duso
current = now()
print(current)                  // Output: 1705862400 (or current timestamp)
```

**format_time(timestamp [, format])** - Format Unix timestamp to string. Accepts number or numeric string (useful for JSON data):
```duso
ts = now()

// Default format (YYYY-MM-DD HH:MM:SS)
print(format_time(ts))          // Output: "2026-01-22 10:30:45"

// Preset formats
print(format_time(ts, "iso"))   // Output: "2026-01-22T10:30:45Z"
print(format_time(ts, "date"))  // Output: "2026-01-22"
print(format_time(ts, "time"))  // Output: "10:30:45"

// Long date formats
print(format_time(ts, "long_date"))     // Output: "January 22, 2026"
print(format_time(ts, "long_date_dow")) // Output: "Thu January 22, 2026"

// Short date formats
print(format_time(ts, "short_date"))    // Output: "Jan 22, 2026"
print(format_time(ts, "short_date_dow"))// Output: "Thu Jan 22, 2026"

// Custom format using standard pattern
print(format_time(ts, "YYYY/MM/DD HH:mm:ss"))  // Output: "2026/01/22 10:30:45"

// Works with numeric strings (from JSON data)
ts_from_json = "1769052551"
print(format_time(ts_from_json, "date"))  // Output: "2026-01-22"
```

Format patterns use standard notation:
- `YYYY` - 4-digit year
- `YY` - 2-digit year
- `MM` - 2-digit month
- `DD` - 2-digit day
- `HH` - 2-digit hour (24-hour)
- `mm` - 2-digit minute
- `ss` - 2-digit second

**parse_time(date_string [, format])** - Parse date string to Unix timestamp:
```duso
// Smart parsing (tries common formats automatically)
ts1 = parse_time("2026-01-22")           // Parses ISO date
ts2 = parse_time("January 22, 2026")     // Parses long date
ts3 = parse_time("Jan 22, 2026")         // Parses short date
ts4 = parse_time("2026-01-22T10:30:45Z") // Parses ISO datetime

// Explicit format
ts5 = parse_time("22/01/2026", "DD/MM/YYYY")
ts6 = parse_time("2026.01.22", "YYYY.MM.DD")
```

When no format is provided, `parse_time()` automatically tries common formats (ISO, long date, short date, date-only). If parsing fails, an error is raised.

### System Functions

**exit([...values])** - Stop execution and optionally return values to host:
```duso
if error_condition then
  exit("fatal error")           // Stops script, returns "fatal error"
end
```

### File I/O Functions (CLI-provided)

The `duso` CLI tool provides three file I/O functions as examples of how apps can customize functionality. These are not part of the core language—different hosts may provide different implementations or none at all.

**load(filename)** - Load file contents as string:
```duso
data = load("config.txt")
print(data)
```

**save(filename, content)** - Write string to file:
```duso
config = "{\"name\": \"test\"}"
save("output.json", config)
```

**include(filename)** - Execute another script file in current environment:
```duso
include("helpers.du")
// Functions and variables from helpers.du are now available
doSomething()
```

Files are resolved relative to the directory of the script being executed. For example:
```
project/
  main.du          (includes "helpers.du")
  helpers.du       (loaded from same directory)
  data/
    config.txt      (loaded with load("data/config.txt"))
```

### Claude Integration Functions (CLI-provided)

The `duso` CLI tool provides functions for integrating with Claude AI. These are host-provided extensions designed for agent orchestration and LLM workflows—not part of the core language.

**claude(prompt [, model] [, tokens])** - One-shot prompt to Claude:

```duso
// Simple prompt
response = claude("What is 2 + 2?")
print(response)

// With model selection
response = claude("Analyze this", model = "claude-opus-4-5-20251101")

// With token limit
response = claude("Write a story", tokens = 500)
```

Returns Claude's response as a string.

**conversation(system [, model] [, tokens])** - Stateful conversation maintaining context:

```duso
// Create a conversation with a specific role
analyst = conversation(
  system = "You are a data analyst. Be concise.",
  model = "claude-haiku-4-5-20251001",
  tokens = 2000
)

// Make requests - context is maintained across calls
result1 = analyst.prompt("Analyze this data: " + data)
result2 = analyst.prompt("What about trends?")
```

**conversation() parameters (named):**
- `system` - System prompt defining the assistant's behavior
- `model` - Claude model to use (default: configured model)
- `tokens` - Maximum tokens per response (default: 1024)

**conversation() methods:**
- `.prompt(message)` - Send a message and get response with context preserved

Perfect for multi-step agent workflows:

```duso
planner = conversation(system = "You are a task planner")

plan = planner.prompt("Create a plan for " + goal)
refined = planner.prompt("Make it more concrete")
final = planner.prompt("Format as JSON steps")

steps = parse_json(final)  // Parse the JSON response
```

## Scope and Environment

### Global Scope

Variables defined at top-level are global:

```duso
x = 5                 // Global
function f()
  print(x)            // Accesses global x
end
```

### Function Scope

Functions can access and modify variables from outer scopes. Use `var` to explicitly create local variables:

```duso
x = 10
function f()
  x = x + 1          // No var = modifies outer x
  print(x)           // Prints 11
end

f()
print(x)             // Prints 11 (outer x was modified)

function g()
  var y = 20         // var = create new local y
  y = y + 5
  print(y)           // Prints 25
end

g()
print(y)             // Error: y is not defined (was local to g)
```

**Best Practice:** Use `var` to be explicit about variable scope, making code clearer and preventing accidental mutations of outer scope variables.

Function parameters are implicit locals:

```duso
x = 5
function f(x)         // x is local parameter, shadows global x
  x = x + 1           // Modifies local x, not global
  return x
end

print(f(10))          // Prints 11
print(x)              // Still prints 5 (global unchanged)
```

For loop variables are also implicit locals:

```duso
for i = 1, 5 do       // i is local to the for loop
  print(i)
end
print(i)              // Error: i is not defined (was local to loop)
```

### Closure

Functions capture their definition environment and can read and modify captured variables:

```duso
function makeCounter(start)
  var count = start
  function increment()
    count = count + 1  // Modifies captured count
    return count
  end
  return increment
end

counter = makeCounter(0)
print(counter())      // Output: 1
print(counter())      // Output: 2
print(counter())      // Output: 3
```

Without `var`, assignments reach up the scope chain to modify existing variables. With `var`, a new local variable is created, shadowing any outer variable:

```duso
x = 10
function test()
  var x = 0  // Creates new local x (shadowing outer)
  x = x + 1
  print(x)   // Prints 1
end

test()
print(x)     // Prints 10 (outer x unchanged)

function modify()
  x = 99     // No var = modifies outer x
end

modify()
print(x)     // Prints 99
```

Block scopes (loops, if statements) also participate in this scoping - they can modify outer variables but you can use `var` to create local ones.

## Concurrency Model

Duso is **single-threaded** by design. It runs sequentially, making scripts predictable and straightforward to reason about.

For **agent orchestration use cases** that require parallel operations (e.g., multiple LLM calls, concurrent tool invocations), parallelism is handled by the **host Go application**, not the Duso script itself.

### Pattern: Host-Provided Parallelism

```duso
// Script runs sequentially...
plan = claude("Create a plan...")

// But host functions can do parallel work internally
results = fetch_all_tools([tool1, tool2, tool3])
// ^ Host application runs these in parallel, returns when all complete

// Script continues with collected results
final = claude("Synthesize: " + results)
```

### Timeouts and Cancellation

For long-running operations (like waiting for an LLM response):
- The **host application** sets execution timeouts
- Scripts can gracefully handle errors with try/catch
- If an operation exceeds the host's timeout, execution is terminated

### Example: Multi-Step Agent Workflow

```duso
// Script orchestrates the workflow sequentially
plan = claude(user_query)           // Wait for plan

// Host function does parallel tool execution
tool_results = run_tools(plan)      // Host handles parallelism

// Collect and synthesize
response = claude("Results: " + tool_results)

return response
```

This design keeps Duso simple while leveraging Go's powerful concurrency primitives at the host level.

## Comments

### Single-line Comments

Comments start with `//` and continue to end of line:

```duso
// This is a comment
x = 5  // Inline comment

// Multiple single-line comments
// Line 1
// Line 2
```

### Multiline Comments

Use `/* ... */` for block comments with nesting support:

```duso
/* This is a block comment
    that spans multiple lines */

/*
  Outer comment
  /* Nested comment */
  Back to outer comment
*/
```

Nested comments are useful for commenting out code that already contains comments:

```duso
x = 10

/*
  Commenting out a function:

  function calculate(a, b)
    // This adds two numbers
    return a + b
  end

  result = calculate(5, 3)
  print("Result: " + result)
*/

print("Code was commented out")
```

## Reserved Keywords

```
if elseif else then end
while do
for in
function return break continue
try catch
and or not
true false nil
```

## Edge Cases and Error Handling

### Division by Zero
```duso
try
  result = 1 / 0
catch (e)
  print(e)  // Output: "division by zero"
end
```

### Index Out of Bounds
```duso
arr = [1, 2, 3]
try
  print(arr[10])
catch (e)
  print(e)  // Output: "array index out of bounds"
end
```

### Calling Non-Function
```duso
x = 5
try
  x()  // Error=x is not a function
catch (e)
  print(e)
end
```

### Undefined Variable
```duso
try
  print(undefined_var)
catch (e)
  print(e)  // Output: "undefined variable=undefined_var"
end
```

## Performance Considerations

- Tree-walking interpreter (not compiled)
- No optimization passes
- Suitable for scripts up to several seconds of execution
- Consider Go functions for performance-critical paths

## Future Features (Deferred)

- Varargs functions
- More string/math built-ins
- Module system (require/import)
- Tail call optimization
- pairs() function for object iteration

## Examples

See `script/examples/` in the repository:
- `basic.du` - Variables and operators
- `arrays.du` - Array operations and iteration
- `functions.du` - Functions and control flow
- `structures.du` - Object templates (using objects as blueprints)
- `coercion.du` - Type coercion and truthiness
- `agents.du` - Agent orchestration patterns
- `templates.du` - String template examples
- `multiline.du` - Multiline string examples
- `print-variants.du` - Multiple print styles
- `benchmark.du` - Prime number counting benchmark
- `colors.du` - ANSI terminal color codes
- `find_replace.du` - String search and replace with contains() and replace()
- `test_var.du` - Variable scoping with var keyword and closures
- `dates.du` - Date and time functions (now, format_time, parse_time)
- `sort_custom.du` - Custom comparison functions for sort()
