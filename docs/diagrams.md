# Duso Visual Architecture & Diagrams

Visual explanations of Duso's architecture, execution model, and runtime concepts.

## 1. Three-Layer Architecture

Duso is organized into three distinct layers, each with clear responsibilities:

```
┌─────────────────────────────────────────────────────────────────┐
│                    CLI Executable (duso)                         │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  pkg/cli/  ~1500 LOC                                    │   │
│  │  • File I/O (load/save)                                │   │
│  │  • Module resolution (DUSO_LIB paths)                 │   │
│  │  • Claude API integration                              │   │
│  │  • Function wrappers & registration                   │   │
│  └──────────────────────────────────────────────────────────┘   │
│                            ↓                                      │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  pkg/runtime/  ~1500 LOC                               │   │
│  │  • HTTP server/client (server.go, client.go)          │   │
│  │  • Concurrency orchestration (spawn, parallel)        │   │
│  │  • Datastore for inter-process communication         │   │
│  │  • Goroutine context management                       │   │
│  │  • Builtin functions: spawn, run, http_server, etc   │   │
│  └──────────────────────────────────────────────────────────┘   │
│                            ↓                                      │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  pkg/script/  ~3500 LOC                                │   │
│  │  • Lexer (tokenization)                                │   │
│  │  • Parser (AST generation)                             │   │
│  │  • Evaluator (execution engine)                        │   │
│  │  • Type system & built-in methods                      │   │
│  │  • Environment & scope management                      │   │
│  │  Dependencies: Go stdlib only ✓                        │   │
│  └──────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘

Embeddable combinations:
├─ script only: Lightweight script evaluation
├─ script + runtime: Add HTTP, concurrency, datastore
└─ script + runtime + cli: Full featured with file I/O & Claude
```

## 2. Execution Pipeline

How a Duso script flows from text to output:

```
┌──────────────────┐
│  Source Code     │
│  (hello.du)      │
└────────┬─────────┘
         │
         ↓
    ┌─────────────────────────────────┐
    │  Lexer (lexer.go)               │
    │  Input:  "x = 5 + 3"            │
    │  Output: Token Stream           │
    │  ┌─────────────────────────────┐│
    │  │ IDENT  NUM  PLUS  NUM  EOF ││
    │  │  x     5    +     3        ││
    │  └─────────────────────────────┘│
    └────────┬────────────────────────┘
             │
             ↓
    ┌─────────────────────────────────┐
    │  Parser (parser.go)             │
    │  Algorithm: Recursive descent   │
    │  + Operator precedence climbing │
    │  ┌─────────────────────────────┐│
    │  │      AssignStatement        ││
    │  │     /       \               ││
    │  │   "x"     BinaryExpr        ││
    │  │           /     \           ││
    │  │          +       3          ││
    │  │         /                   ││
    │  │        5                    ││
    │  └─────────────────────────────┘│
    │  Output: Abstract Syntax Tree   │
    └────────┬────────────────────────┘
             │
             ↓
    ┌─────────────────────────────────┐
    │  Evaluator (evaluator.go)       │
    │  Dispatches on node type        │
    │  Walks AST recursively          │
    │  Maintains Environment (scope)  │
    │  ┌─────────────────────────────┐│
    │  │ Eval(BinaryExpr(+, 5, 3))  ││
    │  │   → Value(8)                ││
    │  │ SetVariable("x", 8)         ││
    │  └─────────────────────────────┘│
    │  Output: Value objects          │
    │  Manages: Variables, functions, │
    │           control flow signals  │
    └────────┬────────────────────────┘
             │
             ↓
    ┌─────────────────────────────────┐
    │  Runtime Values                 │
    │  ┌─────────────────────────────┐│
    │  │ Value {                    ││
    │  │   Type: VAL_NUMBER        ││
    │  │   Data: 8                 ││
    │  │ }                          ││
    │  └─────────────────────────────┘│
    └────────┬────────────────────────┘
             │
             ↓
    ┌─────────────────────────────────┐
    │  Output / Side Effects          │
    │  • Write to stdout              │
    │  • HTTP responses               │
    │  • File writes                  │
    │  • Datastore operations         │
    └─────────────────────────────────┘
```

## 3. Type System

All values in Duso are unified through a single `Value` struct:

```
┌─────────────────────────────────────────────────────────┐
│  Value struct                                           │
│  ┌──────────────────────────────────────────────────┐  │
│  │ Type: ValueType (enum)    Data: interface{}      │  │
│  │ ├─ VAL_NIL                 ├─ nil               │  │
│  │ ├─ VAL_NUMBER              ├─ float64           │  │
│  │ ├─ VAL_STRING              ├─ string            │  │
│  │ ├─ VAL_BOOL                ├─ bool              │  │
│  │ ├─ VAL_ARRAY               ├─ []Value           │  │
│  │ ├─ VAL_OBJECT              ├─ map[string]Value  │  │
│  │ └─ VAL_FUNCTION            └─ *ScriptFunction   │  │
│  │                                or GoFunction     │  │
│  └──────────────────────────────────────────────────┘  │
│                                                         │
│  Benefits:                                              │
│  • Heterogeneous arrays: [1, "hello", true]           │
│  • Flexible objects: {x: 5, name: "test"}             │
│  • First-class functions: f = function() return 42 end│
│  • Simple type checks: type(x) == "number"            │
└─────────────────────────────────────────────────────────┘

Type coercion rules:
┌─────────────────────────────────────┐
│ Implicit conversions                │
├─────────────────────────────────────┤
│ String concatenation (+)            │
│   "Hello " + name → forces string   │
│ Array indexing (a[i])               │
│   Requires i to be a number         │
│ Truthiness in conditionals          │
│   if x then ... end                 │
│   → IsTruthy(x)                     │
│ Logical operators (and, or)         │
│   Return actual values, not bools   │
└─────────────────────────────────────┘
```

## 4. Execution Flow - Single Script

When you run `duso script.du`:

```
┌────────────────────────────────────────────────────────────┐
│  1. CLI Layer (cmd/duso/main.go)                           │
│     • Parse command-line flags                             │
│     • Load script file                                     │
│     • Set up Interpreter with capabilities                │
│     └─→ FileReader, FileWriter, OutputWriter, etc.        │
└────────┬─────────────────────────────────────────────────┘
         │
         ↓
┌────────────────────────────────────────────────────────────┐
│  2. Create Interpreter                                     │
│     Interpreter {                                          │
│       Globals:        shared global variables              │
│       FilePath:       current script directory             │
│       ScriptDir:      for spawn/run path resolution        │
│       ScriptLoader:   capability for loading scripts       │
│       FileReader:     capability for file I/O              │
│       FileWriter:     capability for file I/O              │
│       OutputWriter:   capability for print()               │
│       Datastore:      shared key-value store               │
│       NoFiles:        restrict file access                 │
│     }                                                      │
└────────┬─────────────────────────────────────────────────┘
         │
         ↓
┌────────────────────────────────────────────────────────────┐
│  3. Register Functions                                     │
│     • Built-in methods (type-specific)                     │
│     • Global functions (print, len, split, etc.)          │
│     • Runtime functions (spawn, http_server, etc.)        │
│     • CLI functions (load, save, claude, etc.)            │
│     All registered on the same Interpreter                │
└────────┬─────────────────────────────────────────────────┘
         │
         ↓
┌────────────────────────────────────────────────────────────┐
│  4. Execute Script                                         │
│     Interpreter.Execute(source)                            │
│     ├─ Lexer: tokenize                                     │
│     ├─ Parser: build AST                                   │
│     ├─ Create Evaluator with Interpreter                  │
│     ├─ Evaluator.EvalProgram(ast)                          │
│     │  └─ Walk AST, eval each statement                    │
│     └─ Return final value or error                         │
└────────┬─────────────────────────────────────────────────┘
         │
         ↓
┌────────────────────────────────────────────────────────────┐
│  5. Output / Exit                                          │
│     • Print results to stdout                              │
│     • Return exit code                                     │
│     • Cleanup (close files, stop servers, etc.)            │
└────────────────────────────────────────────────────────────┘
```

## 5. Concurrency Model

How Duso handles parallel execution and spawned processes:

```
┌──────────────────────────────────────────────────────────────┐
│                    Main Script                               │
│                  (Goroutine 1)                               │
│                                                              │
│  pid1 = spawn("worker.du")                                 │
│  pid2 = spawn("worker.du")                                 │
│  pids = [pid1, pid2]                                       │
│                                                              │
└──────────┬──────────────────────────┬───────────────────────┘
           │                          │
           ↓                          ↓
    ┌────────────────┐        ┌────────────────┐
    │  Goroutine 2   │        │  Goroutine 3   │
    │  (worker.du)   │        │  (worker.du)   │
    │   pid=1        │        │   pid=2        │
    │                │        │                │
    │  Fresh eval    │        │  Fresh eval    │
    │  Own vars      │        │  Own vars      │
    │                │        │                │
    └────────────────┘        └────────────────┘

Shared Resources (Thread-safe):
┌──────────────────────────────────────────┐
│  Datastore (shared across goroutines)    │
│  • Global key-value store                │
│  • Protected by RWMutex                  │
│  • Used for inter-process communication  │
│                                          │
│  cache.set("key", value)  ← any worker   │
│  cache.get("key")         ← any worker   │
└──────────────────────────────────────────┘

Control: Goroutines are independent
• No shared variables except datastore
• Each has own evaluator instance
• Each has own environment/scope
• Communication via datastore + wait patterns

When main script exits:
• Spawned goroutines continue running
• Program doesn't exit until all done
• Or user can stop with Ctrl+C
```

## 6. HTTP Server Model

How `http_server()` handles concurrent requests:

```
┌─────────────────────────────────────────────────────────────┐
│  Script startup (goroutine 1)                               │
│                                                              │
│  server = http_server({port = 5150})                       │
│  server.route("GET", "/api/*")                             │
│  server.start()  ← blocks here                             │
└──────────┬────────────────────────────────────────────────┘
           │
           ↓
      (Go's http.Server listening on port 5150)
           │
           ├─────────────────┬─────────────────┬──────────
           │                 │                 │
           ↓                 ↓                 ↓
    ┌────────────────┐ ┌──────────────┐ ┌──────────────┐
    │ Request 1      │ │ Request 2    │ │ Request 3    │
    │ GET /api/users │ │ GET /api/foo │ │ POST /api/bar│
    │ Goroutine 2    │ │ Goroutine 3  │ │ Goroutine 4  │
    └────────┬───────┘ └──────┬───────┘ └──────┬───────┘
             │                │                │
             ↓                ↓                ↓
      ┌────────────────────────────────────────────────┐
      │ Each request:                                  │
      │ 1. Create fresh Evaluator                     │
      │ 2. Copy functions from parent                 │
      │ 3. Set context() via goroutine-local storage  │
      │ 4. Execute handler script                     │
      │ 5. Call ctx.response().json/html/text/etc     │
      │ 6. Return response to client                  │
      └────────────┬───────────────────────────────────┘
                   │
           ┌───────┴───────┬───────────┐
           │               │           │
           ↓               ↓           ↓
        GET 200       GET 200      POST 200
        JSON data     JSON data    Created

Shared Resources:
┌──────────────────────────────────────────┐
│  Datastore (thread-safe caching)         │
│  • Cache pages/responses                 │
│  • Coordinate between requests           │
│  • Protected by RWMutex per key          │
└──────────────────────────────────────────┘

Each request is isolated:
✓ Own variables
✓ Own execution context
✓ Can't interfere with other requests
✗ Can't share mutable state directly
→ Share via datastore only
```

## 7. Module System (require/include)

How scripts load and use modules:

```
Script A                              Script B (module)
┌──────────────────────────┐        ┌──────────────────────┐
│  colors = require(...)   │        │  // my_colors.du     │
│         ↓                │        │  function blend() ... │
│  ┌────────────────────┐  │        │  return {            │
│  │ Module resolution: │  │        │    blend: blend,     │
│  │                    │  │        │    mix: mix,         │
│  │ 1. User path       │  │        │  }                   │
│  │    ./my_colors.du  │  │        └──────────────────────┘
│  │ 2. DUSO_LIB paths  │  │                  ↑
│  │ 3. Embedded stdlib │  │                  │
│  └────────┬───────────┘  │                  │
│           │              │                  │
│           └──────────────┼──────────────────┘
│                          │
│                  (File loaded via
│                   capability)
│
│  Load script in isolated scope:
│  ┌────────────────────┐
│  │ Fresh evaluator    │
│  │ New environment    │
│  │ No parent vars     │
│  │                    │
│  │ Exports stored:    │
│  │  colors = {        │
│  │    blend: fn       │
│  │    mix: fn         │
│  │  }                 │
│  └────────────────────┘
│
│  colors.blend()  ← call exported function
│  colors.mix()
└──────────────────────────────────────────────────────┘

Difference: require vs include

require("module")
├─ Isolated scope
├─ Returns exports object
├─ Cached (requires only once per script)
├─ Good for libraries & dependencies
└─ x = require("math").sqrt(4)

include("script")
├─ Executes in current scope
├─ No return value
├─ Not cached (runs each time)
├─ Good for loading utilities into current context
└─ include("setup.du")  ← runs, can modify current vars
```

## 8. Environment & Scope Chain

How variables are looked up and stored:

```
Function Call Stack:
┌─────────────────────────────────────────┐
│  Global Environment                     │
│  {                                      │
│    print: function                      │
│    split: function                      │
│    x: 10                                │
│  }                                      │
│                                         │
│  ┌──────────────────────────────────┐  │
│  │  Function A Environment          │  │
│  │  {                               │  │
│  │    a: 5                          │  │
│  │    Parent: Global ↑              │  │
│  │  }                               │  │
│  │                                  │  │
│  │  ┌────────────────────────────┐  │  │
│  │  │ Function B Environment     │  │  │
│  │  │ {                          │  │  │
│  │  │   b: 20                    │  │  │
│  │  │   Parent: Function A ↑     │  │  │
│  │  │ }                          │  │  │
│  │  │ ← currently executing      │  │  │
│  │  └────────────────────────────┘  │  │
│  └──────────────────────────────────┘  │
└─────────────────────────────────────────┘

Variable Lookup (from inside Function B):
print(x)   → Walk chain: B → A → Global → found! (10)
print(a)   → Walk chain: B → A → found! (5)
print(b)   → Walk chain: B → found! (20)
print(z)   → Walk chain: B → A → Global → not found! Error

Assignment Rules:
a = 100  → Updates in Function A's scope
          (not in B, respects function barriers)
x = 100  → Updates in Global scope
          (can reach up, affects all)
new_var = 5  → Creates in current scope (B)
               (doesn't leak to parent)
```

## 9. Error Handling & Control Flow

How return, break, and errors are signaled:

```
Control Flow as Error Values:
┌──────────────────────────────────┐
│  ReturnValue (signal)            │
│  • Caught by function definition  │
│  • Stops execution               │
│  • Returns the value             │
│  ├─ return 42                    │
│  └─ return x + y                 │
│                                  │
│  BreakSignal (signal)            │
│  • Caught by for/while           │
│  • Exits loop early              │
│  ├─ break                        │
│  └─ (from inside loop)           │
│                                  │
│  ContinueSignal (signal)         │
│  • Caught by for/while           │
│  • Jumps to next iteration       │
│  ├─ continue                     │
│  └─ (from inside loop)           │
│                                  │
│  DusoError (error)               │
│  • Halts execution               │
│  • Includes stack trace          │
│  ├─ Syntax errors (parser)       │
│  ├─ Runtime errors (evaluator)   │
│  ├─ Type errors (operations)     │
│  └─ Caught by try/catch          │
└──────────────────────────────────┘

Error Handling:

try
  result = dangerous_operation()
catch(err)
  print("Error: " + err)
  result = fallback_value
end

Stack Trace Example:
│
DusoError: Undefined variable 'foo'
  at line 42, column 15 (myScript.du)
  in function 'calculate'
    called from line 25 (myScript.du)
  in main script
│
```

## 10. Builtin Functions vs Methods

Two ways to extend Duso:

```
Builtin Functions (global scope):
┌──────────────────────────────────┐
│ print(x)                         │
│ len(x)                           │
│ type(x)                          │
│ range(n)                         │
│ exit(code)                       │
│ spawn(script, context)           │
│ parallel(fn1, fn2, ...)          │
│ require(module)                  │
│ datastore(name)                  │
│ http_server(config)              │
└──────────────────────────────────┘

Built-in Methods (type-specific):
┌──────────────────────────────────┐
│ Number methods:                  │
│  • (5).to_string()              │
│  • (3.14).floor()               │
│                                  │
│ String methods:                  │
│  • "hello".len()                │
│  • "hello".upper()              │
│  • "hello".split("")            │
│                                  │
│ Array methods:                   │
│  • [1,2,3].len()                │
│  • [1,2,3].map(fn)              │
│  • [1,2,3].join(",")            │
│                                  │
│ Object methods:                  │
│  • {x: 5}.keys()                │
│  • {x: 5}.values()              │
│  • {x: 5}.has("x")              │
│                                  │
│ Datastore methods:               │
│  • cache.set(key, value)        │
│  • cache.get(key)               │
│  • cache.delete(key)            │
└──────────────────────────────────┘

Implementation (pkg/script/builtins.go):
┌──────────────────────────────────────┐
│ For each type, register methods:     │
│                                      │
│ RegisterNumberMethods()              │
│ RegisterStringMethods()              │
│ RegisterArrayMethods()               │
│ RegisterObjectMethods()              │
│                                      │
│ Each method is a GoFunction:         │
│  func(e *Evaluator, args ...Value)  │
│                                      │
│ Called via:  value.method_name()    │
│              ↓                       │
│         Dispatch to GoFunction       │
│              ↓                       │
│         Return new Value             │
└──────────────────────────────────────┘
```

## 11. Data Flow Example: spawn() Process

Detailed walkthrough of spawning a background process:

```
Main Script:
┌────────────────────────────────────────┐
│  pid = spawn("worker.du", {user_id: 5})│
└──────────┬─────────────────────────────┘
           │
           ↓
Runtime (pkg/runtime/builtin_spawn.go):
┌────────────────────────────────────────────────┐
│ NewSpawnFunction() creates a Go function that: │
│                                                │
│ 1. Get script path from arguments              │
│    → "worker.du"                               │
│                                                │
│ 2. Load script via capability:                 │
│    → interp.ScriptLoader("worker.du")          │
│    ↓                                            │
│    CLI layer: resolve from file or embedded    │
│                                                │
│ 3. Get context from arguments (optional)       │
│    → {user_id: 5}                              │
│                                                │
│ 4. Increment spawn counter (atomic)            │
│    → pid = 1  (first spawned process)          │
│                                                │
│ 5. Start new goroutine:                        │
│    go func() {                                 │
│      Create fresh evaluator                    │
│      Set goroutine context to {user_id: 5}    │
│      Execute the script                        │
│      Store result in goroutine-local storage   │
│    }()                                         │
│                                                │
│ 6. Return pid to caller                        │
│    → 1                                         │
└────────────────────────────────────────────────┘
           │
           ↓ (concurrent execution)
┌────────────────────────────────────────┐
│  Spawned Goroutine:                    │
│                                        │
│  ctx = context()  ← {user_id: 5}      │
│  print("Processing user: " + ctx.user_id)
│  // Do work...                         │
│  datastore("results").set("done", true)│
└────────────────────────────────────────┘

Main script continues immediately:
┌────────────────────────────────────┐
│  print("PID: " + pid)               │
│  ← prints "PID: 1" right away      │
│                                    │
│  wait_for("results", "done", true) │
│  ← blocks until spawned proc sets it
└────────────────────────────────────┘
```

## 12. Request Handling in HTTP Server

How a single HTTP request is handled end-to-end:

```
Client Request:
┌──────────────────┐
│ GET /api/users   │
└────────┬─────────┘
         │
         ↓
Go's http.Server (standard library)
         │
         ↓
duso handler (pkg/runtime/goroutine_context.go)
┌──────────────────────────────────────────────┐
│ 1. Create new goroutine for this request     │
│ 2. Extract HTTP request info:                │
│    • Method: GET                             │
│    • Path: /api/users                        │
│    • Query: ?search=john                     │
│    • Headers: {...}                          │
│    • Body: (if POST/PUT)                     │
└────────┬─────────────────────────────────────┘
         │
         ↓
┌──────────────────────────────────────────────┐
│ 3. Store in goroutine-local context          │
│    RequestContext {                          │
│      request: {...}                          │
│      response: {...}                         │
│    }                                         │
└────────┬─────────────────────────────────────┘
         │
         ↓
┌──────────────────────────────────────────────┐
│ 4. Create fresh Evaluator                    │
│    • Copy parent's functions                 │
│    • Create new environment                  │
│    • Link to parent for globals              │
└────────┬─────────────────────────────────────┘
         │
         ↓
┌──────────────────────────────────────────────┐
│ 5. Execute handler script                    │
│                                              │
│    ctx = context()  ← fetches from goroutine │
│    req = ctx.request()                       │
│    res = ctx.response()                      │
│                                              │
│    path = req.path  ← "/api/users"           │
│    query = req.query ← search=john           │
│                                              │
│    if path == "/api/users" then              │
│      users = fetch_users(query.search)       │
│      res.json(users)                         │
│    end                                       │
└────────┬─────────────────────────────────────┘
         │
         ↓
┌──────────────────────────────────────────────┐
│ 6. Response helper called: res.json(users)   │
│    pkg/runtime/response.go:                  │
│    • Set Content-Type: application/json      │
│    • Convert value to JSON                   │
│    • Write to response writer                │
│    • Return to HTTP client                   │
└────────┬─────────────────────────────────────┘
         │
         ↓
┌──────────────────────┐
│ HTTP 200 OK          │
│ Content-Type: JSON   │
│ Body: [...]          │
└──────────────────────┘
```

---

## Quick Reference

| Concept | Location | File | Purpose |
|---------|----------|------|---------|
| **Lexer** | pkg/script | lexer.go | Convert source → tokens |
| **Parser** | pkg/script | parser.go | Convert tokens → AST |
| **Evaluator** | pkg/script | evaluator.go | Execute AST |
| **Type System** | pkg/script | value.go | Runtime values |
| **Environment** | pkg/script | environment.go | Variable scopes |
| **Builtins** | pkg/script | builtins.go | Type methods |
| **HTTP Server** | pkg/runtime | server.go | Handle requests |
| **Concurrency** | pkg/runtime | goroutine_context.go | Spawn/parallel |
| **File I/O** | pkg/cli | functions.go | load/save |
| **Claude API** | pkg/cli | claude.go | AI integration |

