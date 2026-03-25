# Duso Internals

This document describes Duso's architecture, design decisions, and how the runtime works under the hood. It's intended for contributors, embedders, and anyone curious about how the language actually executes.

## Overview

Duso is an AST-based interpreter written in pure Go with no external dependencies. It's designed to be:

- **Simple to embed**: Use in Go applications with minimal setup
- **LLM-friendly**: Syntax and design that's intuitive even without training data
- **Concurrent**: Built-in concurrency primitives for orchestration tasks
- **Observable**: Debug mode, call stacks, and error context built in
- **Self-contained**: All stdlib and contrib modules embedded in the binary

The runtime is split into three layers:

1. **Core Language** (`pkg/script/`): Lexer, parser, AST, evaluator, type system, builtins (~3500 LOC)
   - Embeddable: Yes (core language only)
   - Dependencies: Go stdlib only

2. **Runtime Orchestration** (`pkg/runtime/`): HTTP server/client, datastore, concurrency context, goroutine management (~1500 LOC)
   - Embeddable: Yes (can use directly in Go apps)
   - Dependencies: `pkg/script` only
   - Can be used with or without the CLI

3. **CLI Extensions** (`pkg/cli/`): File I/O, Claude integration, module resolution, function wrappers (~1500 LOC)
   - Embeddable: Optional (for script writers who want file access)
   - Dependencies: `pkg/script`, `pkg/runtime`, `pkg/anthropic`
   - CLI-specific features like load/save and claude API integration

**Usage patterns:**
- **Embedded in Go**: Use `pkg/script` directly, optionally add `pkg/runtime` features
- **CLI usage**: Uses all three: `script` → `runtime` → `cli`
- **Custom distributions**: Can use `script` + `runtime` with custom CLI features

## Architecture Overview

```
Source Code
    ↓
Lexer (lexer.go) → Token Stream
    ↓
Parser (parser.go) → AST (ast.go)
    ↓
Evaluator (evaluator.go) ↔ Environment (environment.go)
    ↓
Value (value.go)
    ↓
Output / Side Effects
```

Every layer is independent and testable. The evaluator doesn't know about the file system; the CLI layer adds that.

## Core Components

### Lexer

File: `lexer.go`

The lexer converts source code into a token stream. It handles:

- **Keywords**: `if`, `for`, `function`, `return`, etc.
- **Operators**: `+`, `-`, `*`, `/`, `%`, `==`, `!=`, `<`, `>`, etc.
- **Literals**: Numbers, strings (single/double/triple quotes), booleans, nil
- **Template strings**: `"{{expr}}"` syntax parsed as special tokens
- **Comments**: `//` (single-line) and `/* */` (nested multi-line)
- **Identifiers**: Variable names, function names

The lexer tracks position (line, column) for error reporting. It doesn't do any semantic analysis—that's the parser's job.

### Parser

File: `parser.go`

Uses recursive descent parsing to convert tokens into an AST. Key techniques:

- **Operator precedence climbing** for binary expressions
- **Bracket/paren tracking** for better error messages ("expected closing brace at...")
- **Special handling for templates** to preserve interpolated expressions
- **Control flow statements** (if/else/while/for) parsed as dedicated AST nodes

The parser produces an untyped AST; type checking and execution happen during evaluation.

### AST Structure

File: `ast.go`

The AST is composed of nodes that implement the `Node` interface:

```
Node (interface)
├── Statement nodes (if, while, for, function def, assignment)
├── Expression nodes (binary ops, function calls, literals)
└── Literal nodes (numbers, strings, arrays, objects)
```

Key distinction: **Statements** (execute for side effects) vs **Expressions** (produce values). An `AssignStatement` is distinct from a `BinaryExpr` because assignment has evaluation order semantics.

### Evaluator

File: `evaluator.go`

The core runtime. Implements the `Eval(node Node) (Value, error)` function that:

1. Dispatches on node type
2. Recursively evaluates children
3. Applies operators and control flow
4. Returns `Value` (runtime value) or `error` (DusoError, control signals, etc.)

Key patterns:

- **Control flow as errors**: `return` is signaled via `ReturnValue` error, caught by function definitions
- **Variable lookup**: Walks environment chain upward (unless in parallel/function scope)
- **Lazy evaluation**: Some operations only evaluate needed branches (e.g., ternary operator)

The evaluator is single-threaded per goroutine. For concurrent execution, we create isolated child evaluators (see [Concurrency Model](#concurrency-model)).

### Type System

File: `value.go`

Duso has 10 runtime types, all wrapped in a `Value` struct:

```go
type Value struct {
  Type ValueType    // VAL_NUMBER, VAL_STRING, etc.
  Data interface{}  // Actual value (float64, string, []Value, etc.)
}

type ValueType int
const (
  VAL_NIL ValueType = iota
  VAL_NUMBER        // float64
  VAL_STRING        // string
  VAL_BOOL          // bool
  VAL_ARRAY         // []Value
  VAL_OBJECT        // map[string]Value
  VAL_FUNCTION      // ScriptFunction or GoFunction
  VAL_CODE          // Pre-parsed code (AST + metadata)
  VAL_ERROR         // First-class error value (message + stack)
  VAL_BINARY        // Immutable binary data (files, images)
)
```

**Why this design?**

- Single unified type for Duso values, enabling heterogeneous arrays and objects
- `Data` as `interface{}` avoids type assertions in most code (but allows casts when needed)
- Simplicity: no complex type hierarchy or tagging schemes

**Type conversions** are handled by builtins (`tonumber()`, `tostring()`, etc.) and implicit coercion in specific places (e.g., array indexing requires numbers).

### Environment & Scope

File: `environment.go`

Variable scoping is a linked-list of environments:

```
Current Env
  ↑ parent
Parent Env
  ↑ parent
... (up to root)
```

Each `Environment` has:
- `variables`: map of variable names to Values
- `parent`: pointer to parent (or nil for root)
- `isFunctionScope`: true if this env is a function boundary

**Lookup**: Walk up the parent chain until found (or error if not found)

**Set**:
- If already exists locally, update
- If in function scope and doesn't exist locally, create locally (don't leak to parent)
- Otherwise, walk up and update in the scope where it exists

**Why this design?**

This is how Lua does it, and it's elegant: functions capture their closure (parent at definition time), and mutations within a function scope don't leak outward unless the variable was already accessible.

The `var` keyword explicitly creates a local variable, shadowing any outer binding.

## Module System

**Note:** Module resolution is CLI-specific and found in `pkg/cli/`. Embedded applications can implement their own module loading using `require()` and `include()` by registering custom functions.

### Module Resolution

When `require("foo")` or `include("foo.du")` is called (CLI usage):

1. **Current directory**: Files in the current working directory (supports absolute and relative paths)
2. **Search paths**: Directories in `DUSO_LIB` environment variable
3. **Embedded modules**: `/EMBED/stdlib/`, `/EMBED/contrib/` (for stdlib like `http`, `claude`)

File: `pkg/cli/module_resolver.go` (CLI-specific path resolution)
File: `pkg/script/circular_detector.go` (Circular dependency detection)

### Module Caching

Two caches:

1. **Parse cache** (`parseCache`): Maps file path → AST (with mtime validation)
   - Embedded files (`/EMBED/`) use cached AST forever
   - Real files check `mtime` on every access; invalidate if newer
   - Reduces parsing overhead for frequently used modules

2. **Module cache** (`moduleCache`): Maps path → exported value (only for `require()`)
   - Caches the **result** of executing a module, not the AST
   - Used to ensure `require()` returns the same value across multiple calls

**Thread-safe** via `sync.RWMutex`

### require() vs include()

- **`include(file)`**: Executes file in current scope. Variables leak into caller's scope. No caching of results (AST cached, but always re-execute).

- **`require(module)`**: Executes in isolated scope. Only exported value (last expression or explicit return) is visible. Result cached and reused.

### Circular Dependency Detection

File: `pkg/cli/circular_detector.go`

Uses a stack-based tracker: as modules load, they're pushed onto a stack. If we encounter a module already on the stack, it's a cycle. Error is thrown with the cycle path.

## Runtime Values & Functions

### ScriptFunction

A function defined in Duso:

```go
type ScriptFunction struct {
  Name        string
  FilePath    string
  Parameters  []*Parameter
  Body        []Node
  Closure     *Environment  // Parent env at definition time (closure)
}
```

When called:
1. Create child environment with Closure as parent
2. Bind parameters to child environment
3. Execute Body statements
4. Catch `ReturnValue` error → return its value
5. If no explicit return, return last expression value or nil

### GoFunction

A function implemented in Go:

```go
type GoFunction func(args map[string]any) (any, error)
```

Arguments are passed as a map containing:
- `"0"`, `"1"`, `"2"`, ... for positional arguments
- Named argument keys for named arguments

The return `any` is automatically converted to a `Value`. Errors are propagated as `DusoError`.

### Built-in Functions

File: `pkg/script/builtins.go` and `pkg/runtime/` (84+ functions)

Core functions include:
- **String**: `len()`, `substr()`, `upper()`, `lower()`, `contains()`, `replace()`, `split()`, `join()`, `repeat()`, `trim()`, `starts_with()`, `ends_with()`, `find()`
- **Array**: `map()`, `filter()`, `reduce()`, `sort()`, `push()`, `pop()`, `shift()`, `unshift()`, `range()`, `keys()`, `values()`
- **Math**: `abs()`, `floor()`, `ceil()`, `round()`, `sqrt()`, `pow()`, `min()`, `max()`, `sin()`, `cos()`, `tan()`, `exp()`, `log()`, `ln()`, `pi()`, `random()`, `clamp()`, trigonometric and logarithmic functions
- **Type**: `type()`, `tonumber()`, `tostring()`, `tobool()`, `deep_copy()`
- **JSON**: `format_json()`, `parse_json()`
- **Time**: `now()`, `format_time()`, `parse_time()`, `sleep()`
- **Crypto**: `hash()`, `hash_password()`, `verify_password()`, `sign_rsa()`, `verify_rsa()`, `encode_base64()`, `decode_base64()`
- **Markdown**: `markdown_html()`, `markdown_ansi()`, `markdown_text()`
- **HTTP**: `fetch()`, `http_server()`
- **Concurrency**: `parallel()`, `spawn()`, `run()`, `kill()`, `context()`
- **Data**: `datastore()`, `template()`
- **Control**: `exit()`, `throw()`, `parse()`
- **Debug**: `breakpoint()`, `watch()`
- **System**: `env()`, `uuid()`, `input()`

These are registered during interpreter creation and available in all scripts.

## Concurrency Model

Duso has three concurrency primitives, each with different semantics:

### 1. parallel(functions)

Executes functions concurrently and waits for all to complete.

**Implementation**:
1. Iterate over functions (array, object, varargs)
2. For each function, create a child `Evaluator` with `isParallelContext = true`
3. When `isParallelContext` is true, `Environment.Set()` is blocked from walking to parent (isolated scope)
4. Launch all in `sync.WaitGroup`
5. Collect results (or `nil` if error)
6. Return results in same structure as input

**Semantics**:
- True parallelism (goroutines run concurrently)
- Read-only access to parent scope
- Each goroutine gets its own Evaluator
- If one errors, that result is `nil`

### 2. spawn(script, context)

Executes script asynchronously in a background goroutine.

**Implementation**:
1. Return immediately (non-blocking)
2. Launch goroutine that:
   - Creates `RequestContext` for goroutine-local storage
   - Executes script in fresh Evaluator
   - Stores result in `RequestContext.ExitChan`
3. Inherits all registered functions from parent interpreter

**Semantics**:
- Fire-and-forget
- Script has access to its own context via `context()`
- Can call `exit(value)` to signal completion
- Useful for background tasks, workers

### 3. run(script, context, timeout?)

Executes script synchronously and waits for result.

**Implementation**:
1. Create result channel
2. If timeout specified, use `context.WithTimeout`
3. Spawn goroutine (same as `spawn()`)
4. Block on result channel or timeout
5. Return value from `exit()` or error

**Semantics**:
- Blocking (waits for script to finish)
- Script runs in a separate goroutine (benefits from Go scheduling)
- Returns value passed to `exit()`
- Timeout support for long-running scripts

### Goroutine-Local Context Storage

File: `pkg/runtime/goroutine_context.go`

Each spawned goroutine needs its own "request context" (call stack, exit channel, etc.). Go doesn't have true goroutine-local storage, so we use:

```go
var requestContexts sync.Map

// Key is goroutine ID (from runtime/cgo.GetGoroutineID)
// Value is *RequestContext

type RequestContext struct {
  Frame        *InvocationFrame  // For call stack
  ExitChan     chan any          // For receiving exit() value
  ContextData  any               // User data from spawn/run
}
```

This avoids global state issues and allows multiple concurrent scripts without interference.

### HTTP Server Configuration

File: `pkg/runtime/http_server.go`

The `http_server()` function supports extensive configuration options:

- **Network**: `address`, `port`, `tls_enabled`, `cert_file`, `key_file`, `websocket_enabled`
- **Performance**: `timeout`, `request_handler_timeout`, `idle_timeout`, `max_body_size`, `max_header_size`, `max_headers`, `max_form_fields`
- **Caching**: `cache_control`, `static_cache_control`
- **Security**: `jwt_config` (HS256/RS256), `cors` (origins, methods, headers, credentials)
- **Serving**: `show_directory_listing`, `default_files`, `access_log`
- **Routes**: Regex-based route matching with parameter extraction

## Error Handling

### DusoError

File: `errors.go`

Errors include:

```go
type DusoError struct {
  Message   string
  FilePath  string
  Position  Position  // Line and column
  CallStack []CallFrame
}
```

When formatted, includes source context:
```
file.du:42:10: undefined variable 'foo'

Call stack:
  at main (file.du:42:10)
  at helper (file.du:35:5)
  at global (file.du:1:0)
```

### Control Flow Errors

Certain operations are signaled via error returns (not thrown):

- `ReturnValue { Value }`: Caught by function definitions
- `BreakIteration`: Caught by for/while loops
- `ContinueIteration`: Caught by for/while loops
- `ExitExecution { Values }`: Propagates to interpreter, causes exit
- `BreakpointError { Env }`: Caught by debug REPL

This is efficient and allows precise control flow without special syntax.

### Error Queueing in Debug Mode

File: `cmd/duso/main.go` (debug REPL)

When running with `-debug`:

1. Parse statements (not whole program)
2. Execute statement-by-statement
3. If error occurs:
   - Print error with source context
   - Queue error for later review
   - Enter debug REPL in current environment
   - User can inspect variables, step through, etc.
   - Continue on `c` command
4. Errors are queued so user isn't flooded (process one at a time)

This prevents the common debugging nightmare of "here are 500 errors, which one matters?"

## Debugging

### Breakpoints

`breakpoint()` function (in debug mode):

```duso
x = 42
breakpoint()  // Pause here
y = x + 1
```

With `-debug` flag:
1. Execution pauses at `breakpoint()`
2. Debug REPL enters with current environment
3. User can inspect variables, step, continue

### Watches

`watch(expr, ...)` function (in debug mode):

```duso
watch("x")              // Break if x changes
watch("x > 5", "y")     // Break if either expression changes
```

Useful for conditional breakpoints without writing if statements.

## Public Go Embedding API

File: `pkg/script/script.go`

### Basic Usage

```go
interp := script.NewInterpreter(verbose bool)
output, err := interp.Execute("print(1 + 2)")
```

### Common Methods

```go
// Execution
output, err := interp.Execute(source string) (string, error)

// Custom Go functions
err := interp.RegisterFunction(name string, fn GoFunction) error

// Module execution (returns last value, not output)
value, err := interp.ExecuteModule(source string) (Value, error)

// Configuration
interp.SetDebugMode(enabled bool)
interp.SetScriptDir(dir string)
interp.SetFilePath(path string)

// Inspection
output := interp.GetOutput()
stack := interp.GetCallStack() []CallFrame
cache, exists := interp.GetModuleCache(path string)
```

### Integration with CLI Extensions

```go
interp := script.NewInterpreter(false)
err := cli.RegisterFunctions(interp, cli.RegisterOptions{
  ScriptDir: ".",
  HTTPPort:  8080,
  // ... other options
})
// Now has: load, save, include, require, spawn, run, http_server, datastore, etc.
```

## Performance Notes

### AST-Based Interpreter Performance

Duso is an AST-based interpreter (not bytecode), which is simpler but slower than bytecode or JIT. Benchmarks show:

- Simple arithmetic: ~1M ops/sec (expected for AST interpreter)
- String operations: Good (string builtins are Go functions)
- API calls: Bottleneck is I/O, not Duso evaluation
- Array operations: Reasonable for typical sizes

**Optimization strategies in the runtime**:

1. **Parse caching**: AST cached with mtime validation, no re-parsing on module reuse
2. **Go builtins**: Heavy lifting (string ops, JSON, HTTP) done in Go, not Duso
3. **Goroutine per request**: HTTP server requests are handled in separate goroutines, enabling true concurrency
4. **Minimal allocations**: Environment chain reuses parent pointers; values are stack-allocated when possible

For LLM orchestration (the primary use case), performance is adequate—the bottleneck is API latency, not Duso evaluation.

## Design Philosophy

### LLM-Friendly

The language was designed with the assumption that LLMs (like Claude) would be reading and understanding Duso code without training data. This influences:

- **Readable syntax**: No special characters or cryptic operators
- **Clear semantics**: Behavior is predictable even without documentation
- **Helpful errors**: Call stacks and position info included automatically
- **Consistent structure**: Similar operations have similar syntax

### Simplicity Over Cleverness

- No complex type system (just 7 types)
- No operator overloading or implicit conversions
- No advanced metaprogramming features
- Control flow via explicit statements, not hidden magic

### Self-Contained

- No external Go dependencies
- All stdlib/contrib modules embedded in binary
- No runtime configuration complexity
- Executable is self-sufficient

## Embedding in Go Applications

To embed Duso in a Go app:

1. Import `github.com/duso-org/duso/pkg/script`
2. Create interpreter: `interp := script.NewInterpreter(false)`
3. Optionally register custom Go functions
4. Execute: `output, err := interp.Execute(source)`

For scripts that need file I/O or HTTP:

```go
import "github.com/duso-org/duso/pkg/cli"

interp := script.NewInterpreter(false)
cli.RegisterFunctions(interp, cli.RegisterOptions{})
output, err := interp.Execute(source)
```

---

For questions or contributions, see the main README and CONTRIBUTING guide.
