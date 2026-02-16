# Duso Go Testing Primer

A comprehensive guide for writing Go tests that exercise the Duso interpreter using integration testing with Duso scripts.

**Last Updated:** February 15, 2026
**Go Version:** 1.25+
**Testing Approach:** Table-driven integration tests using script strings

---

## Table of Contents

1. [Core Principles](#core-principles)
2. [Understanding the Codebase](#understanding-the-codebase)
3. [Testing Method](#testing-method)
4. [Duso Script Reference](#duso-script-reference)
5. [Example Test Patterns](#example-test-patterns)
6. [Checklist for Each Module](#checklist-for-each-module)

---

## Core Principles

### Why Integration Testing is Efficient

Go's coverage tool is **purely instrumented at the Go source level**—it rewrites your Go code at compile time to insert counters on every basic block/statement. This measures which **Go statements executed**, not what caused them to execute.

**Key insight:** A single Duso script fed through the interpreter exercises:
- **Lexer** - Tokenization
- **Parser** - AST construction
- **Evaluator** - Statement execution
- **Builtins** - Function calls
- **All control paths** - if/while/for/try/catch branches

Example:
```go
func TestArithmetic(t *testing.T) {
    interp := script.NewInterpreter(false)
    runtime.RegisterBuiltins()

    _, err := interp.Execute(`
        x = 0
        for i = 1, 10 do
            x = x + i
        end
        x
    `)
    if err != nil {
        t.Fatalf("failed: %v", err)
    }
    // Every Go statement in lexer, parser, evaluator, and builtins that
    // this script touches counts toward coverage. No manual unit tests needed.
}
```

The script exercises:
- Lexer: tokenizes all symbols
- Parser: constructs assignment, for loop, binary operations
- Evaluator: evaluates all AST node types
- Builtin: range() function (if used elsewhere)

---

## Understanding the Codebase

### Package Structure

```
pkg/
├── script/           # Core language (lexer, parser, evaluator, AST)
├── runtime/          # Builtin functions (math, string, array, etc.)
├── cli/              # CLI-specific builtins (file I/O, modules)
└── core/             # Shared utilities (IsInteger, etc.)
```

### Key Components

#### 1. Script Package (`pkg/script/`)

The core language interpreter. No external dependencies.

**Main Types:**
- `Interpreter` - Public API for executing scripts
- `Evaluator` - AST execution engine
- `Lexer` - Tokenizer
- `Parser` - AST builder
- `Program` (AST) - Statements and expressions

**Key Methods:**
```go
// Create interpreter
interp := script.NewInterpreter(verbose bool)

// Execute a script string
output, err := interp.Execute(sourceCode string) (string, error)

// Execute a file
output, err := interp.ExecuteFile(filepath string) (string, error)

// Register builtins (called once at startup)
script.RegisterBuiltin(name string, fn GoFunction)
```

**How It Works:**
```
Input (script string)
    ↓
Lexer.Tokenize() → []Token
    ↓
Parser.Parse() → *Program (AST)
    ↓
Evaluator.Eval(Program) → (value any, error error)
    ↓
Output
```

#### 2. Runtime Package (`pkg/runtime/`)

Builtin functions (math, strings, arrays, objects, etc.). Registers once at startup.

**Setup:**
```go
// Call once at process startup
runtime.RegisterBuiltins()  // Populates script.globalBuiltins

// Set global interpreter (needed by some builtins like spawn/run)
runtime.SetInterpreter(interp)
```

**Builtin Function Signature:**
```go
type GoFunction func(evaluator *Evaluator, args map[string]any) (any, error)

// Example:
func builtinEnv(evaluator *Evaluator, args map[string]any) (any, error) {
    varname := args["0"].(string)  // Positional argument
    return os.Getenv(varname), nil
}
```

**How Arguments Work:**
- Positional: `args["0"]`, `args["1"]`, `args["2"]`, etc.
- Named: `args["paramName"]`
- Both: Can support either (common pattern)

**Example from Duso:**
```duso
env("PATH")              -- args["0"] = "PATH"
env(varname = "PATH")    -- args["varname"] = "PATH"
upper("hello")           -- args["0"] = "hello"
substr("hello", 1, 2)    -- args["0"]="hello", args["1"]=1.0, args["2"]=2.0
```

#### 3. CLI Package (`pkg/cli/`)

Extended functionality for command-line use (file I/O, module loading, REPL).

**Main Functions:**
- `RegisterFunctions(interp)` - Registers CLI-specific builtins
- File operations: load, save, load_file, etc.
- Module system: require, include
- Console: error, write, debug functions
- Busy and concurrency utilities

**Note:** CLI functions depend on the Interpreter's host-provided capabilities:
```go
interp.FileReader = func(path string) ([]byte, error) { ... }
interp.FileWriter = func(path, content string) error { ... }
interp.ScriptLoader = func(path string) ([]byte, error) { ... }
interp.OutputWriter = func(msg string) error { ... }
interp.InputReader = func(prompt string) (string, error) { ... }
```

### Execution Flow

```go
// 1. Setup
interp := script.NewInterpreter(false)
runtime.RegisterBuiltins()
runtime.SetInterpreter(interp)

// Optional: Set host capabilities for CLI functions
interp.OutputWriter = fmt.Println

// 2. Execute script
result, err := interp.Execute(`
    x = 10
    y = 20
    x + y
`)

// 3. Coverage happens automatically as script executes
// All Go code touched by the script counts toward coverage
```

---

## Testing Method

### Test File Organization

For each module to test, create a test file:

```
pkg/runtime/datastore.go        →  pkg/runtime/datastore_test.go
pkg/script/evaluator.go         →  pkg/script/evaluator_test.go
pkg/cli/builtin_load.go         →  pkg/cli/builtin_load_test.go
```

### Test Structure Template

```go
package runtime

import (
    "testing"
    "github.com/duso-org/duso/pkg/script"
)

func TestModuleFeature(t *testing.T) {
    // 1. Setup
    t.Parallel()
    interp := script.NewInterpreter(false)
    runtime.RegisterBuiltins()
    runtime.SetInterpreter(interp)

    // Optional setup for specific test
    // e.g., mock file I/O, set environment variables, etc.

    // 2. Define test cases with scripts
    tests := []struct {
        name     string
        script   string
        check    func(*testing.T, string, error)  // Assertion function
    }{
        {
            name: "simple case",
            script: `
                x = 10
                y = 20
                x + y
            `,
            check: func(t *testing.T, out string, err error) {
                if err != nil {
                    t.Fatalf("unexpected error: %v", err)
                }
                // Script executed without error = success
            },
        },
        // ... more test cases
    }

    // 3. Run tests
    for _, tt := range tests {
        tt := tt  // Capture for parallel execution
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            _, err := interp.Execute(tt.script)
            tt.check(t, "", err)
        })
    }
}
```

### Coverage-Focused Testing

To maximize coverage, create tests for:

1. **Happy path** - Normal usage
2. **Edge cases** - Empty input, boundary values, nil
3. **Error cases** - Invalid types, missing arguments
4. **Complex scenarios** - Nested operations, multiple calls
5. **All branches** - Every if/else path in the Go code

**Example: Testing `len()` builtin**

The `builtinLen` Go function probably has code like:
```go
func builtinLen(evaluator *Evaluator, args map[string]any) (any, error) {
    val := args["0"]
    if val == nil { return 0, nil }
    switch v := val.(type) {
    case string:
        return float64(len(v)), nil
    case []any:  // Array
        return float64(len(v)), nil
    case map[string]any:  // Object
        return float64(len(v)), nil
    default:
        return nil, fmt.Errorf("len() requires string, array, or object")
    }
}
```

Test it thoroughly:
```go
func TestLen(t *testing.T) {
    t.Parallel()
    interp := script.NewInterpreter(false)
    runtime.RegisterBuiltins()

    tests := []struct {
        name   string
        script string
        check  func(*testing.T, error)
    }{
        {
            name: "string length",
            script: `len("hello")`,
            check: func(t *testing.T, err error) {
                if err != nil { t.Fatalf("error: %v", err) }
            },
        },
        {
            name: "empty string",
            script: `len("")`,
            check: func(t *testing.T, err error) {
                if err != nil { t.Fatalf("error: %v", err) }
            },
        },
        {
            name: "array length",
            script: `len([1, 2, 3])`,
            check: func(t *testing.T, err error) {
                if err != nil { t.Fatalf("error: %v", err) }
            },
        },
        {
            name: "empty array",
            script: `len([])`,
            check: func(t *testing.T, err error) {
                if err != nil { t.Fatalf("error: %v", err) }
            },
        },
        {
            name: "object length",
            script: `len({a = 1, b = 2})`,
            check: func(t *testing.T, err error) {
                if err != nil { t.Fatalf("error: %v", err) }
            },
        },
        {
            name: "nil length",
            script: `len(nil)`,
            check: func(t *testing.T, err error) {
                if err != nil { t.Fatalf("error: %v", err) }
            },
        },
    }

    for _, tt := range tests {
        tt := tt
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            _, err := interp.Execute(tt.script)
            tt.check(t, err)
        })
    }
}
```

Each test case exercises a different code path in `builtinLen`:
- `string` branch
- `[]any` (array) branch
- `map[string]any` (object) branch
- `nil` case
- Error path (if you call with wrong type)

All of these count toward coverage.

### Testing Script Package (lexer, parser, evaluator)

The `pkg/script/` modules handle **language syntax and semantics**. The reference docs are essential for comprehensive testing:

**For Parser (`script/parser.go`):**

Read the syntax from reference docs to test all variations:
- `docs/reference/function.md` - function definitions, expressions, closures
- `docs/reference/array.md` - array literals, access, iteration
- `docs/reference/object.md` - object literals, access, methods
- `docs/reference/if.md` - if/elseif/else, ternary operator
- `docs/reference/for.md` - for loops (range and iteration), break, continue
- `docs/reference/while.md` - while loops, break, continue
- `docs/reference/try.md` - try/catch/throw error handling

Example test for parser:
```go
func TestParser(t *testing.T) {
    t.Parallel()
    interp := script.NewInterpreter(false)
    runtime.RegisterBuiltins()

    tests := []struct {
        name   string
        script string
    }{
        // From docs/reference/function.md examples
        {
            name: "function_definition",
            script: `function add(a, b) return a + b end`,
        },
        {
            name: "function_closure",
            script: `
                function makeAdder(n)
                  function add(x) return x + n end
                  return add
                end
            `,
        },
        // From docs/reference/array.md examples
        {
            name: "array_access",
            script: `arr = [1, 2, 3]; arr[0]`,
        },
        // From docs/reference/if.md examples
        {
            name: "if_elseif_else",
            script: `
                x = 15
                if x > 20 then
                  result = "big"
                elseif x > 10 then
                  result = "medium"
                else
                  result = "small"
                end
            `,
        },
        // From docs/reference/for.md examples
        {
            name: "for_range_loop",
            script: `
                sum = 0
                for i = 1, 10 do
                  sum = sum + i
                end
            `,
        },
        // From docs/reference/try.md examples
        {
            name: "try_catch",
            script: `
                try
                  x = risky()
                catch (err)
                  x = nil
                end
            `,
        },
    }

    for _, tt := range tests {
        tt := tt
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            _, err := interp.Execute(tt.script)
            if err != nil {
                t.Fatalf("parse error: %v", err)
            }
        })
    }
}
```

**For Evaluator (`script/evaluator.go`):**

Test all language semantics from reference docs:
- Variable scoping and closure capture
- Control flow execution (if branches, loops)
- Type coercion and comparisons
- Truthiness (arrays, objects, functions)
- Error handling (try/catch propagation)

Example:
```go
func TestEvaluator(t *testing.T) {
    t.Parallel()
    interp := script.NewInterpreter(false)
    runtime.RegisterBuiltins()

    tests := []struct {
        name   string
        script string
    }{
        // From docs/reference/function.md - closure capture
        {
            name: "closure_capture",
            script: `
                function makeAdder(n)
                  function add(x) return x + n end
                  return add
                end
                addFive = makeAdder(5)
                addFive(10)  // Should be 15
            `,
        },
        // From docs/reference/array.md - truthiness
        {
            name: "array_truthiness_nonempty",
            script: `
                if [1, 2, 3] then
                  result = "truthy"
                else
                  result = "falsy"
                end
            `,
        },
        // From docs/reference/array.md - truthiness edge case
        {
            name: "array_truthiness_empty",
            script: `
                if [] then
                  result = "truthy"
                else
                  result = "falsy"
                end
            `,
        },
        // From docs/reference/object.md - truthiness
        {
            name: "object_truthiness",
            script: `
                obj = {a = 1}
                if obj then result = "truthy" end
            `,
        },
        // From docs/reference/try.md - error propagation
        {
            name: "try_catch_error",
            script: `
                try
                  throw("test error")
                catch (err)
                  result = err
                end
            `,
        },
    }

    for _, tt := range tests {
        tt := tt
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            _, err := interp.Execute(tt.script)
            if err != nil {
                t.Fatalf("execution error: %v", err)
            }
        })
    }
}
```

### Best Practices

1. **Use table-driven tests** with `t.Run()` and subtests
2. **Always use `t.Parallel()`** to run tests concurrently
3. **Capture loop variable** with `tt := tt` before parallel execution
4. **Focus on coverage first** - Verify script runs without error
5. **Add assertions** only for critical outputs (most tests just need "no error")
6. **Keep scripts readable** - Use proper indentation and comments
7. **Test error paths** - Invalid arguments, missing arguments, wrong types
8. **Use `runtime.SetInterpreter(interp)`** for builtins that need it (spawn, run, etc.)
9. **Mock file I/O** - Set `interp.FileReader`, `interp.FileWriter`, etc. for tests that need it
10. **Reset state** - Each test should be independent (freshly instantiate if needed)

---

## Named Function Arguments & Reference Docs

### CRITICAL: Read docs/reference/*.md for EVERYTHING

**Before writing tests, read the relevant reference files.** The `docs/reference/` directory contains documentation for:

- **Builtins** (print, len, map, filter, etc.)
- **Data Types** (array, object, string, number, boolean, nil)
- **Keywords & Control Flow** (if, for, while, function, return, break, continue, try, catch, throw, var)

Each file documents:

1. **All calling signatures** - positional vs named arguments
2. **Parameter variations** - optional, defaults, overloads
3. **Behavior edge cases** - negative indices, empty input, special cases
4. **All code examples** - shows usage patterns you should test

**Examples: Before writing tests, read:**

For builtins:
```bash
docs/reference/replace.md      # Shows: positional/named args, optional ignore_case param
docs/reference/format_json.md  # Shows: optional indent parameter
```

For script package (lexer, parser, evaluator):
```bash
docs/reference/function.md     # Shows: syntax, parameters, closures, return values
docs/reference/array.md        # Shows: access patterns, iteration, edge cases, truthiness
docs/reference/if.md           # Shows: if/elseif/else syntax, ternary operator
docs/reference/for.md          # Shows: range loops, array iteration, break/continue
docs/reference/object.md       # Shows: creation, access, constructor pattern, truthiness
docs/reference/try.md          # Shows: try/catch syntax, error handling patterns
```

Each reference file provides:
- Syntax rules and examples
- Edge cases and special behaviors
- Error conditions
- Variations to test

### Named Arguments in Duso

Duso supports **named arguments** alongside positional arguments:

```duso
-- Positional
replace("hello", "h", "H")

-- Named
replace(string="hello", pattern="h", replacement="H")

-- Mixed
replace("hello", pattern="h", replacement="H")

-- With optional parameter
replace("hello", "h", "H", ignore_case=true)
```

Both calling styles go into the same `args map[string]any` in the Go function:

```go
// Positional call: replace("hello", "h", "H", false)
args["0"] = "hello"
args["1"] = "h"
args["2"] = "H"
args["3"] = false

// Named call: replace(string="hello", pattern="h", replacement="H", ignore_case=false)
args["string"] = "hello"
args["pattern"] = "h"
args["replacement"] = "H"
args["ignore_case"] = false
```

### Testing Named Arguments (Pragmatic Approach)

From a **Go code coverage perspective**, you need:

1. **At least ONE positional test** - exercises `args["0"]`, `args["1"]`, etc.
2. **At least ONE named test** - exercises `args["paramName"]` branches
3. **At least ONE mixed test** - positional + named together
4. **Error cases** - missing required args, wrong types

You **don't** need to test every combination—each parameter is extracted independently:

```go
// In the builtin function:
if v, ok := args["0"]; ok {              // Branch 1: positional
    str = v.(string)
} else if v, ok := args["string"]; ok {  // Branch 2: named
    str = v.(string)
} else {
    return nil, fmt.Errorf("missing")    // Branch 3: error
}
```

Once you've tested positional (branch 1) and named (branch 2), that covers the extraction logic. Parameter 1 being positional doesn't affect how parameter 2 is extracted.

**Example: Efficient test covering all patterns:**

```go
func TestReplaceFunction(t *testing.T) {
    t.Parallel()
    interp := script.NewInterpreter(false)
    runtime.RegisterBuiltins()

    tests := []struct {
        name   string
        script string
    }{
        // Positional: exercises args["0"], args["1"], args["2"], args["3"]
        {
            name:   "positional",
            script: `replace("hello hello", "hello", "hi", false)`,
        },

        // Named: exercises args["string"], args["pattern"], etc.
        {
            name:   "named",
            script: `replace(string="hello hello", pattern="hello", replacement="hi", ignore_case=false)`,
        },

        // Mixed: positional + named together
        {
            name:   "mixed",
            script: `replace("hello hello", "hello", replacement="hi", ignore_case=false)`,
        },

        // Optional omitted: exercises default handling
        {
            name:   "optional_omitted",
            script: `replace("hello hello", "hello", "hi")`,
        },

        // Behavior variants (from reference docs)
        {
            name:   "regex_pattern",
            script: `replace("Price: 10, Qty: 5", ~\d+~, "X")`,
        },
        {
            name:   "function_replacement",
            script: `replace("1 2 3", ~\d+~, function(text, pos, len) return tostring(tonumber(text) * 2) end)`,
        },

        // Error cases
        {
            name:   "missing_required",
            script: `replace("text")`,
        },
    }

    for _, tt := range tests {
        tt := tt
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            _, err := interp.Execute(tt.script)

            if tt.name == "missing_required" {
                if err == nil {
                    t.Fatal("expected error for missing argument")
                }
            } else {
                if err != nil {
                    t.Fatalf("unexpected error: %v", err)
                }
            }
        })
    }
}
```

This test covers:
- ✅ Positional argument extraction
- ✅ Named argument extraction
- ✅ Mixed calling style
- ✅ Optional parameter defaults
- ✅ Function behavior (from reference docs)
- ✅ Error paths

All with minimal test cases that exercise all Go code branches.

---

## Duso Script Reference

The Duso language executed in tests. Full reference at `docs/learning-duso.md`.

### Data Types

```duso
-- Number (float64)
x = 42
y = 3.14

-- String
s = "hello"
multiline = """
  Line 1
  Line 2
"""

-- Boolean
active = true
inactive = false

-- Nil
nothing = nil

-- Array
nums = [1, 2, 3]
items = []

-- Object
person = {name = "Alice", age = 30}
config = {}
```

### Control Flow

```duso
-- If/Else
if x > 10 then
  print("big")
elseif x > 5 then
  print("medium")
else
  print("small")
end

-- Ternary
status = x > 10 ? "big" : "small"

-- For loop (range)
for i = 1, 10 do
  print(i)
end

-- For loop (array)
for item in items do
  print(item)
end

-- While loop
count = 0
while count < 5 do
  print(count)
  count = count + 1
end

-- Break and Continue
for i = 1, 10 do
  if i == 5 then break end
  if i == 2 then continue end
  print(i)
end
```

### Functions

```duso
-- Define function
function add(a, b)
  return a + b
end

-- Call function
result = add(2, 3)

-- Anonymous function
double = function(x) return x * 2 end
nums = map([1, 2, 3], double)

-- Default parameters
function greet(name, greeting = "Hello")
  return greeting + " " + name
end

greet("Alice")            -- "Hello Alice"
greet("Bob", "Hi")        -- "Hi Bob"
```

### String Templates

```duso
name = "Alice"
age = 30
msg = "{{name}} is {{age}} years old"
-- Output: "Alice is 30 years old"
```

### Error Handling

```duso
try
  result = risky_operation()
catch (error)
  print("Error: " + error)
  result = nil
end

-- Throw error
if invalid then
  throw("Invalid input")
end

-- Throw object for structured errors
throw({
  code = "ERR_NOT_FOUND",
  message = "Item not found",
  status = 404
})
```

### Common Builtins

**Output:**
- `print(value)` - Print to stdout
- `input(prompt)` - Read from stdin

**Type Operations:**
- `type(value)` - "number", "string", "array", "object", "function", "nil"
- `len(value)` - Length of string, array, object
- `tonumber(value)` - Convert to number
- `tostring(value)` - Convert to string
- `tobool(value)` - Convert to boolean

**String Operations:**
- `upper(s)` - Uppercase
- `lower(s)` - Lowercase
- `substr(s, start, length)` - Substring
- `trim(s)` - Remove whitespace
- `split(s, delimiter)` - Split string
- `join(array, delimiter)` - Join array to string
- `contains(s, pattern)` - Check if pattern exists
- `find(s, pattern)` - Find matches
- `replace(s, pattern, replacement)` - Replace

**Array Operations:**
- `push(array, value)` - Add to end
- `pop(array)` - Remove from end
- `shift(array)` - Remove from start
- `unshift(array, value)` - Add to start
- `keys(object)` - Get keys
- `values(object)` - Get values
- `range(start, end, step)` - Create sequence

**Math Operations:**
- `abs(n)`, `floor(n)`, `ceil(n)`, `round(n)`
- `sqrt(n)`, `pow(a, b)`
- `sin(n)`, `cos(n)`, `tan(n)`
- `min(a, b)`, `max(a, b)`, `clamp(n, min, max)`
- `random()` - Random 0-1

**Functional:**
- `map(array, function)` - Transform
- `filter(array, function)` - Keep matching
- `reduce(array, function, initial)` - Combine
- `sort(array, compareFunction)` - Sort

**Data Operations:**
- `parse_json(string)` - Parse JSON
- `format_json(value, indent)` - Serialize to JSON
- `deep_copy(value)` - Deep copy

**Other:**
- `env(varname)` - Read environment variable
- `now()` - Current Unix timestamp
- `format_time(timestamp, format)` - Format timestamp
- `uuid()` - Generate UUID
- `throw(error)` - Raise error
- `exit(value)` - Exit script with value

### Advanced Features

**Spawning other scripts:**
```duso
-- Async spawn
pid = spawn("worker.du", {data = config})

-- Sync run (blocking)
result = run("processor.du", {input = data})

-- Context in spawned script
ctx = context()
if ctx then
  request_data = ctx.request()
end
```

**Datastore (for coordination):**
```duso
store = datastore("job_123")
store.set("status", "running")
store.increment("count", 1)
result = store.wait("done", true, timeout = 30)
```

**HTTP:**
```duso
response = fetch("https://api.example.com", {
  method = "POST",
  headers = {["Authorization"] = "Bearer token"},
  body = format_json({key = "value"})
})

if response.ok then
  data = response.json()
end

-- Server
server = http_server({port = 8080})
server.route("GET", "/", "handler.du")
server.start()
```

---

## Example Test Patterns

### Pattern 1: Simple Builtin Test

Test a builtin function with multiple cases:

```go
func TestStringFunctions(t *testing.T) {
    t.Parallel()
    interp := script.NewInterpreter(false)
    runtime.RegisterBuiltins()

    tests := []struct {
        name   string
        script string
    }{
        {
            name:   "upper",
            script: `upper("hello")`,
        },
        {
            name:   "lower",
            script: `lower("WORLD")`,
        },
        {
            name:   "trim",
            script: `trim("  spaces  ")`,
        },
        {
            name:   "substr",
            script: `substr("hello", 1, 3)`,
        },
    }

    for _, tt := range tests {
        tt := tt
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            _, err := interp.Execute(tt.script)
            if err != nil {
                t.Fatalf("unexpected error: %v", err)
            }
        })
    }
}
```

### Pattern 2: Complex Operations

Test interactions between multiple features:

```go
func TestArrayOperations(t *testing.T) {
    t.Parallel()
    interp := script.NewInterpreter(false)
    runtime.RegisterBuiltins()

    script := `
        -- Create array
        nums = [1, 2, 3, 4, 5]

        -- Map
        doubled = map(nums, function(x) return x * 2 end)

        -- Filter
        evens = filter(doubled, function(x) return x % 2 == 0 end)

        -- Reduce
        sum = reduce(evens, function(acc, x) return acc + x end, 0)

        -- Result
        sum
    `

    _, err := interp.Execute(script)
    if err != nil {
        t.Fatalf("script error: %v", err)
    }
}
```

### Pattern 3: Error Path Testing

Test error handling:

```go
func TestErrorCases(t *testing.T) {
    t.Parallel()
    interp := script.NewInterpreter(false)
    runtime.RegisterBuiltins()

    tests := []struct {
        name        string
        script      string
        expectError bool
    }{
        {
            name:        "valid script",
            script:      `x = 10; x + 5`,
            expectError: false,
        },
        {
            name:        "invalid syntax",
            script:      `x = 10; y z`,
            expectError: true,
        },
        {
            name:        "undefined variable",
            script:      `undefined_var + 1`,
            expectError: true,
        },
    }

    for _, tt := range tests {
        tt := tt
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            _, err := interp.Execute(tt.script)
            if (err != nil) != tt.expectError {
                t.Fatalf("expected error=%v, got=%v", tt.expectError, err != nil)
            }
        })
    }
}
```

### Pattern 4: Testing with Control Flow

Test language features like loops and conditionals:

```go
func TestControlFlow(t *testing.T) {
    t.Parallel()
    interp := script.NewInterpreter(false)
    runtime.RegisterBuiltins()

    tests := []struct {
        name   string
        script string
    }{
        {
            name: "for loop",
            script: `
                sum = 0
                for i = 1, 10 do
                    sum = sum + i
                end
                sum
            `,
        },
        {
            name: "while loop",
            script: `
                count = 0
                while count < 5 do
                    count = count + 1
                end
                count
            `,
        },
        {
            name: "if else",
            script: `
                x = 15
                if x > 20 then
                    result = "big"
                elseif x > 10 then
                    result = "medium"
                else
                    result = "small"
                end
                result
            `,
        },
        {
            name: "break and continue",
            script: `
                sum = 0
                for i = 1, 10 do
                    if i == 3 then continue end
                    if i == 8 then break end
                    sum = sum + i
                end
                sum
            `,
        },
    }

    for _, tt := range tests {
        tt := tt
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            _, err := interp.Execute(tt.script)
            if err != nil {
                t.Fatalf("script error: %v", err)
            }
        })
    }
}
```

---

## Checklist for Each Module

When testing a module (e.g., `runtime/datastore.go`), follow this checklist:

### Phase 1: Understand the Module
- [ ] **Read `docs/reference/` for each function being tested** - This is THE source of truth
- [ ] Read the module source code
- [ ] Identify all public functions and their parameters
- [ ] Check for builtin functions in `register.go`
- [ ] Note any dependencies on Interpreter or other packages
- [ ] Check if Duso scripts call these functions (look in examples/)
- [ ] Note all optional parameters and their defaults
- [ ] Identify which functions support named arguments

### Phase 2: Design Tests

For each public function (after reading its reference doc):

**For named argument support:**
- [ ] At least one positional test (exercises `args["0"]`, `args["1"]`, etc.)
- [ ] At least one named test (exercises `args["paramName"]` branches)
- [ ] At least one mixed test (positional + named together)
- [ ] Optional parameters both with and without

**For all functions:**
- [ ] Happy path - normal usage (from reference doc examples)
- [ ] Edge cases - empty/nil/zero inputs (from reference doc)
- [ ] Error paths - invalid types, missing required arguments
- [ ] Complex scenarios - multiple calls, nested operations, callbacks
- [ ] All branches - ensure every if/else in Go code is exercised

### Phase 3: Write Test File

Create `module_test.go` with:
- [ ] Proper package declaration
- [ ] Import statements (script, runtime, testing, etc.)
- [ ] Setup in each test (NewInterpreter, RegisterBuiltins, SetInterpreter)
- [ ] Table-driven tests with multiple cases
- [ ] Use `t.Parallel()` and subtests
- [ ] Clear, readable test names
- [ ] Comments explaining what each test exercises

### Phase 4: Coverage Validation

After writing tests:
- [ ] Run: `go test ./pkg/runtime -coverprofile=cover.out`
- [ ] View: `go tool cover -html=cover.out`
- [ ] Check module coverage
- [ ] Identify uncovered branches
- [ ] Write additional tests for uncovered paths

### Phase 5: Module-Specific Notes

**For builtin functions:**
- [ ] **Read `docs/reference/[function].md` first** - source of truth
- [ ] Test all calling patterns: positional, named, mixed
- [ ] Test optional parameters present and omitted
- [ ] Test type coercion (string to number, etc.)
- [ ] Test with all supported input types (from reference doc)
- [ ] Test all behavior variants (from reference doc examples)
- [ ] Test error cases (invalid types, missing required args)

**For module system (require/include):**
- [ ] Test module caching
- [ ] Test circular dependency detection
- [ ] Test path resolution
- [ ] Test module isolation

**For concurrency (spawn/run/parallel):**
- [ ] Test spawning works
- [ ] Test context passing
- [ ] Test datastore coordination
- [ ] Test error handling in spawned scripts

**For file I/O (load/save):**
- [ ] Test reading files
- [ ] Test writing files
- [ ] Test error handling (file not found, etc.)
- [ ] Test with virtual filesystems (/STORE/, /EMBED/)

**For HTTP (fetch/http_server):**
- [ ] Test various HTTP methods (GET, POST, etc.)
- [ ] Test headers
- [ ] Test request/response bodies
- [ ] Test error handling

---

## Quick Reference: Test Setup

### Minimal Test Template

```go
package runtime

import (
    "testing"
    "github.com/duso-org/duso/pkg/script"
    "github.com/duso-org/duso/pkg/runtime"
)

func TestFeature(t *testing.T) {
    t.Parallel()

    // Setup
    interp := script.NewInterpreter(false)
    runtime.RegisterBuiltins()
    runtime.SetInterpreter(interp)

    // Execute script
    _, err := interp.Execute(`
        x = 42
        x + 1
    `)

    // Verify
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
}
```

### Coverage Command

```bash
# Generate coverage report
go test ./pkg/runtime ./pkg/script ./pkg/cli -coverprofile=cover.out

# View in browser
go tool cover -html=cover.out

# Check coverage percentage per package
go test ./pkg/runtime -cover
go test ./pkg/script -cover
go test ./pkg/cli -cover
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run with race detector
go test -race ./...

# Run specific test file
go test ./pkg/runtime -run TestDatastore

# Run specific test function
go test ./pkg/runtime -run TestDatastore/persistence

# Parallel execution (automatic with t.Parallel())
go test -parallel 16 ./...

# Verbose output
go test -v ./pkg/runtime

# Show coverage per function
go test -covermode=count ./pkg/runtime
```

---

## Troubleshooting

### Script Won't Execute
- Check Duso syntax in the script string
- Verify builtins are registered: `runtime.RegisterBuiltins()`
- Check if builtin needs interpreter: `runtime.SetInterpreter(interp)`

### Builtin Not Found
- Ensure `RegisterBuiltins()` was called
- Check if builtin is registered in `register.go`
- For CLI functions, also call `cli.RegisterFunctions(interp)`

### Coverage Not Showing
- Use `interp.Execute()` to run scripts (coverage is automatic)
- Don't call evaluator/parser directly—use Interpreter API
- Remember: coverage is measured on Go code, scripts exercise it

### Concurrency Issues
- Use `t.Parallel()` to enable parallel execution
- Capture loop variable: `tt := tt`
- Don't share state between subtests
- Use `runtime.SetInterpreter(interp)` only once per test

---

## References

- **Duso Language:** `docs/learning-duso.md`
- **Builtin Reference:** `docs/reference/` (one file per builtin)
- **Go Testing:** https://golang.org/pkg/testing/
- **Go Coverage:** https://golang.org/doc/coverage
- **Duso Examples:** `examples/` directory

---

**This primer is your guide. Use it for every module you test.**
