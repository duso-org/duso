# Duso Internals

This document describes Duso's architecture, design decisions, and how the runtime works under the hood. It's intended for contributors, embedders, and anyone curious about how the language actually executes.

## Overview

Duso is an AST-based interpreter written in Go. It's designed to be:

- **Simple to embed**: Use in Go applications with minimal setup
- **LLM-friendly**: Syntax and design that's intuitive without training data
- **Concurrent**: Built-in concurrency primitives for orchestration tasks
- **Observable**: Debug mode, call stacks, and error context built in
- **Self-contained**: All stdlib and contrib modules embedded in the binary

### Dependencies

The core language has no third-party dependencies — `pkg/script` is Go stdlib only. The
full binary is not dependency-free: `go.mod` requires `github.com/go-sql-driver/mysql`
and `github.com/lib/pq` (for `sql()`), plus `golang.org/x/crypto` and `golang.org/x/net`.
Anything embedding only `pkg/script` pulls none of them.

### Packages

Three layers carry the language itself:

1. **Core Language** (`pkg/script/`, ~8.3k LOC): lexer, parser, AST, resolver, evaluator,
   environment, type system
   - Dependencies: Go stdlib only
   - Embeddable on its own — this is the language with no I/O and no builtins

2. **Runtime** (`pkg/runtime/`, ~24k LOC): every builtin, plus HTTP server/client,
   datastore and replication, websockets, SQL, mail, images, concurrency context
   - Dependencies: `pkg/script`, plus the third-party modules above
   - By far the largest package; 46 `builtin_*.go` files register 151 builtins

3. **CLI Extensions** (`pkg/cli/`, ~3.5k LOC): file I/O, module resolution, embedded-FS
   access, function wrappers

Three smaller packages sit alongside them:

4. **`pkg/core/`** (~200 LOC): small shared helpers used across the other packages
5. **`pkg/lsp/`** (~2.5k LOC): language server — completion, diagnostics, hover
6. **`pkg/version/`** (~140 LOC): version constants

**Usage patterns:**
- **Embedded in Go**: `pkg/script` alone for a pure sandboxed language; add `pkg/runtime`
  for builtins
- **CLI usage**: all three main layers, `script` → `runtime` → `cli`

## Architecture Overview

```
Source Code
    ↓
Lexer (pkg/script/lexer.go) → Token Stream
    ↓
Parser (pkg/script/parser.go) → AST (pkg/script/ast.go)
    ↓
Resolver (pkg/script/resolver.go) → AST annotated with parameter slots
    ↓
Evaluator (pkg/script/evaluator.go) ↔ Environment (pkg/script/environment.go)
    ↓
Value (pkg/script/value.go)
    ↓
Output / Side Effects
```

All core files live in `pkg/script/`; the bare filenames used below are relative to it.

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

### Resolver

File: `resolver.go`

After parsing, `resolveProgram()` (called from `parser.go`) walks the AST and annotates
identifiers that provably refer to a parameter of the enclosing function with a slot
index. Parameters occupy the first inline storage slots of the function environment in
declaration order, so an annotated identifier reads that slot directly instead of walking
the scope chain doing string comparisons.

The pass is deliberately conservative — it annotates only when the dynamic lookup would
provably give the same answer. A parameter shadowed anywhere in the body (by `var`, a loop
variable, a catch variable, or a nested function name) is not slotted at all, and
identifiers inside named-argument expressions are never slotted. Anything ambiguous falls
back to the name-based path.

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

Duso has 11 runtime types, all wrapped in a `Value` struct:

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
  VAL_REGEX         // Compiled regular expression pattern
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
- `names` / `vals`: fixed-size inline arrays holding the first `smallScopeSize` bindings
- `overflow`: `map[string]Value`, nil until a scope outgrows the inline slots
- `parent`: pointer to parent (or nil for root)
- `fnScope`: nearest enclosing function-scope env, so slot reads index its `vals` directly
- `self`: receiver value for method calls
- `isFunctionScope`: true if this env is a function boundary
- `paramFlags` / `parameters`: parameter tracking — a bitmask for common single-letter
  names, a lazily allocated map for the rest

The inline arrays are the reason most scopes cost no map allocation at all; the overflow
map only appears for unusually wide scopes.

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
3. **Embedded modules**: `/EMBED/stdlib/`, `/EMBED/contrib/` — these are duso source, not
   Go: `stdlib/` and `contrib/` ship as `.du` modules loaded through the same resolver as
   any user file

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

File: `pkg/script/circular_detector.go`

Uses a stack-based tracker: as modules load, they're pushed onto a stack. If we encounter a module already on the stack, it's a cycle. Error is thrown with the cycle path.

## Runtime Values & Functions

### ScriptFunction

A function defined in Duso:

```go
type ScriptFunction struct {
  Name        string
  FilePath    string        // Where defined, for error reporting
  Parameters  []*Parameter
  Body        []Node
  Closure     *Environment  // Parent env at definition time (closure)

  // Unexported performance state:
  paramFlags  uint64          // Precomputed parameter marking, copied onto each call env
  paramMap    map[string]bool // Uncommon param names, shared read-only across calls
  poolable    bool            // Body creates no closures, so call envs may be reused
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
type GoFunction func(evaluator *Evaluator, args map[string]any) (any, error)
```

A second, allocation-free form exists for hot builtins, taking a positional slice instead
of a map:

```go
type GoFunctionFast func(evaluator *Evaluator, args []Value) (Value, error)
```

Arguments are passed as a map containing:
- `"0"`, `"1"`, `"2"`, ... for positional arguments
- Named argument keys for named arguments

The return `any` is automatically converted to a `Value`. Errors are propagated as `DusoError`.

### Built-in Functions

Files: `pkg/runtime/builtin_*.go` (46 files), registered in `pkg/runtime/register.go`
(151 builtins). Note these live in `pkg/runtime`, not `pkg/script` — `pkg/script` is the
bare language and ships no builtins of its own.

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
var requestContexts sync.Map   // key: goroutine ID (uint64) → value: *RequestContext
```

The ID comes from `script.GetGoroutineID()` (`pkg/script/execution.go`), which parses the
first line of `runtime.Stack()` — Go exposes no supported goroutine-ID API, so this is the
usual workaround.

`RequestContext` is shared by every context-carrying entry point (HTTP handlers, WebSocket
connections, `spawn()`, `run()`), so it is wide. The load-bearing fields:

```go
type RequestContext struct {
  Request      *http.Request           // nil outside an HTTP handler
  Writer       http.ResponseWriter     // nil outside an HTTP handler
  Data         any                     // generic context data (spawn/run)
  PathParams   map[string]any          // extracted from the route pattern
  Frame        *script.InvocationFrame // root invocation frame, for the call stack
  ExitChan     chan any                // receives the exit() value
  ResponseData map[string]any          // set by the response helpers
  // plus body caching, JWT keys, cache-control and limit fields for HTTP
}
```

This avoids global state issues and allows multiple concurrent scripts without interference.

### HTTP Server Configuration

File: `pkg/runtime/http_server.go`

The `http_server()` function supports extensive configuration options:

- **Network**: `address`, `port`, `https`, `cert_file`, `key_file`, `cert_reload_interval`
- **Limits**: `timeout`, `request_handler_timeout`, `idle_timeout`, `max_body_size`,
  `max_header_size`, `max_headers`, `max_form_fields`, `max_websocket_connections`
- **Caching**: `cache_control`, `static_cache_control`
- **Security**: `jwt` (HS256/RS256), `cors` (origins, methods, headers, credentials)
- **Serving**: `directory`, `default`, `access_log`, `uploads`
- **WebSocket**: `websocket` (queue sizes, timeouts, message size, rate limit)
- **Routes**: pattern matching with path-parameter extraction

See [http_server()](/docs/reference/http_server.md) for the authoritative list and defaults.

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

> **Status: not yet a designed API.** The Go interface is currently a by-product of
> building the CLI rather than a product in its own right. It works, but the surface was
> shaped by what the CLI happened to need, not by what an embedder would want, and it
> shows:
>
> - **`Execute()` returns a string that is always empty.** Script output goes to stdout;
>   the return value is vestigial. Capture output by other means.
> - **`ExecuteFile()` is a stub.** It reads nothing and executes nothing, returning
>   `("", nil)` — file loading lives in `pkg/cli`.
> - **`RegisterOptions` has one field.** Debug and no-files mode are read out of the
>   `duso_sys` datastore instead of being passed in, so configuring an embedded interpreter
>   means writing to a global datastore.
> - **No execution limits.** No timeout, no instruction budget, no memory cap.
> - Several exported methods are really internals the CLI needed to reach.
>
> **None of this is covered by a stability promise, and it will change.** If you are
> embedding duso today, pin a version and read `pkg/script/script.go` as the source of
> truth. Turning this into a first-class API is open work, not a finished story.

### Basic Usage

```go
import (
    _ "github.com/duso-org/duso/pkg/runtime" // registers the builtins
    "github.com/duso-org/duso/pkg/script"
)

interp := script.NewInterpreter()          // no arguments
_, err := interp.Execute(`print(1 + 2)`)   // prints to stdout; the string return is always ""
```

`pkg/script` on its own is the bare language: no builtins, no file access, no network — a
genuine sandbox. `print()` and everything else in the reference docs live in `pkg/runtime`,
which registers them from its own `init()`, so importing that package is all it takes.

### Common Methods

```go
// Execution
_, err := interp.Execute(source string) (string, error)      // string is always ""
value, err := interp.ExecuteModule(source string) (Value, error)   // returns last value
err = interp.ExecuteNode(node Node) error                    // one node; used by the debugger
// interp.ExecuteFile(path) exists but is an unimplemented stub — do not use

// Custom Go functions
err = interp.RegisterFunction(name string, fn GoFunction) error
err = interp.RegisterObject(name string, methods map[string]GoFunction) error

// Configuration
interp.SetScriptDir(dir string)
interp.SetFilePath(path string)

// Inspection
stack := interp.GetCallStack() []CallFrame
ev    := interp.GetEvaluator() *Evaluator

// Module cache
value, mtime, ok := interp.GetModuleCache(path string)
interp.SetModuleCache(path string, value Value, mtime int64)

// Lifecycle
interp.Reset()
```

### Registering a Go function

`GoFunction` receives the calling evaluator and a map of arguments. Positional arguments
arrive under `"0"`, `"1"`, `"2"`, …; named arguments under their own names:

```go
interp.RegisterFunction("add", func(ev *script.Evaluator, args map[string]any) (any, error) {
    a := args["a"].(float64)   // add(a = 1, b = 2)
    b := args["b"].(float64)
    return a + b, nil
})
```

Numbers arrive as `float64` — duso has no integer type. The returned `any` is converted to
a `Value` automatically, and a returned error propagates as a `DusoError`.

### Registering an object with methods

`RegisterObject` defines an object whose properties are Go functions, callable from a
script as `name.method(...)`:

```go
interp.RegisterObject("agents", map[string]script.GoFunction{
    "classify": func(ev *script.Evaluator, args map[string]any) (any, error) {
        return map[string]any{"confidence": 0.85}, nil
    },
})
// script: print(agents.classify("input").confidence)
```

### Adding CLI extensions

```go
import "github.com/duso-org/duso/pkg/cli"

interp := script.NewInterpreter()
err := cli.RegisterFunctions(interp, cli.RegisterOptions{ScriptDir: "."}, nil)
```

`RegisterFunctions` takes three arguments; the third is a `*cli.StdinHTTPServer` and may be
nil. `RegisterOptions` currently carries a single field, `ScriptDir` — debug and no-files
mode are read from the sys datastore at registration time rather than passed in. This is a
good example of the API being shaped by the CLI rather than designed for embedders.

### Working examples

`/go-embedding/` holds four programs that build and run against the current tree:

- `hello-world/` — minimal interpreter and `Execute`
- `custom-functions/` — registering Go functions, named arguments, returning objects
- `config-dsl/` — duso as a configuration language
- `task-scripting/` — scripting an application's own operations

They are compiled by `go build ./...` from that directory, so they are the most reliable
reference for the current signatures.

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

For the target workload — a single-node web server whose time goes to I/O — evaluation
speed is rarely the constraint. See [Performance Report](/docs/performance-report.md) for
measured numbers.

## Design Philosophy

### LLM-Friendly

The language was designed on the assumption that models would read and write duso without
having been trained on it. This influences:

- **Readable syntax**: No special characters or cryptic operators
- **Clear semantics**: Behavior is predictable even without documentation
- **Helpful errors**: Call stacks and position info included automatically
- **Consistent structure**: Similar operations have similar syntax

### Simplicity Over Cleverness

- No complex type system (11 value types, no user-defined types)
- No operator overloading or implicit conversions
- No advanced metaprogramming features
- Control flow via explicit statements, not hidden magic

### Self-Contained

- Core language (`pkg/script`) has no third-party dependencies
- All docs, stdlib and contrib modules embedded in the binary, locked to that revision
- No runtime configuration complexity
- Executable is self-sufficient

---

For questions or contributions, see the main README and CONTRIBUTING guide.
