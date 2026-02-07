# Duso `pipe_server()` - Implementation Plan (Revised)

## Overview

Add a `pipe_server()` builtin that listens on stdin/stdout for JSON-RPC messages and dispatches them to handler scripts. Enables Duso scripts to act as JSON-RPC servers (perfect for MCP, custom protocols, etc).

**Design**: Mirrors `http_server()` pattern - script-based handlers, fresh evaluator per message, coordination via datastore.

## Architecture

### Go Layer (`pkg/runtime/pipe_server.go`)

Core server logic (reuses LSP transport infrastructure):

```go
type PipeServer struct {
    handlers       map[string]string  // method -> handler script path
    reader         *lsp.MessageReader // Reuse LSP message parsing
    writer         *lsp.MessageWriter // Reuse LSP message writing
    interpreter    *script.Interpreter
    scriptDir      string
    nextRequestID  int
    mu             sync.RWMutex
    stopChan       chan struct{}
}

// Core methods
func NewPipeServer(interp *script.Interpreter) *PipeServer
func (ps *PipeServer) OnMethod(method string, handlerScript string)
func (ps *PipeServer) Start() error  // Blocks, reads from stdin, writes to stdout
func (ps *PipeServer) Stop() error
func (ps *PipeServer) handleMessage(msg *lsp.Message) *lsp.Message
```

**Key design decisions:**
- Reuse `pkg/lsp` message framing (Content-Length headers, JSON marshaling)
- Each incoming JSON-RPC message spawns handler script in isolated evaluator
- Handler scripts use `context()` to access the message and send responses
- No need for new transport layer—LSP's stdio transport is perfect

### Duso Builtin (`pkg/runtime/builtin_pipe_server.go`)

Expose as Duso function:

```duso
server = pipe_server()

server.on(method, handlerScript)  // Register handler
server.start()                     // Start listening (blocks)
```

### Handler Scripts

Handler scripts work exactly like `http_server` handlers:

```duso
// handlers/tools.du
ctx = context()
msg = ctx.request()  // Gets the JSON-RPC message

// Process the message
result = handle_tool_call(msg.params)

// Send response
exit({
  "jsonrpc" = "2.0",
  "id" = msg.id,
  "result" = result
})
```

Handler script has access to:
- `context().request()` - The JSON-RPC message (method, params, id)
- `context().response()` - Response helpers (same as http_server)
- All Duso builtins: `spawn()`, `datastore()`, `fetch()`, etc.
- Fresh evaluator per message (no shared state between handlers)

## Implementation Steps

### Phase 1: Core Runtime (Go)

1. **`pkg/runtime/pipe_server.go`**
   - `PipeServer` struct with handler map
   - `OnMethod(method, scriptPath)` - Register handler
   - `Start()` - Main loop reading from stdin, dispatching to handlers
   - `handleMessage(msg)` - Route to handler script, manage evaluator lifecycle
   - Reuse `pkg/lsp.MessageReader/Writer` for JSON-RPC framing
   - Error handling: malformed JSON, handler crashes, timeouts

2. **`pkg/runtime/builtin_pipe_server.go`**
   - `NewPipeServerFunction()` - Expose `pipe_server()` builtin
   - Return server object with `.on(method, script)` and `.start()` methods
   - Thread-safe handler registration (can call `.on()` after `.start()`)

3. **Tests** - `pkg/runtime/builtin_pipe_server_test.go`
   - Mock stdin/stdout for testing
   - Single handler test
   - Multiple handlers (different methods)
   - Error handling (handler script errors, malformed JSON)
   - Message framing (Content-Length parsing)

### Phase 2: Integration

1. **Example**: `examples/core/pipe_server.du`
   - Simple echo server (test helper)
   - MCP-like server (initialize, tools/call handlers)
   - Example showing worker spawning, datastore coordination

2. **Documentation**: `docs/reference/pipe_server.md`
   - Signature, parameters, return value
   - Methods: `.on(method, script)`, `.start()`
   - Handler context (`context().request()`, `context().response()`)
   - Examples (echo, MCP-like, with datastore coordination)
   - Concurrency section (fresh evaluator per message)
   - Notes (blocks until SIGINT, handler script from disk, timeouts)

3. **Update existing docs**:
   - `docs/learning-duso.md` - Add pipe_server to subprocess section
   - Update README if pipe_server is a major feature

### Phase 3: MCP Server Example (Pure Duso)

Create template in `contrib/mcp/server.du`:

```duso
// Simple MCP server in pure Duso
server = pipe_server()

server.on("initialize", "handlers/initialize.du")
server.on("tools/list", "handlers/tools_list.du")
server.on("tools/call", "handlers/tools_call.du")

print("MCP server listening on stdin/stdout")
server.start()
```

Handlers could:
- Call backend services via `pipe_server("redis-cli")` (client mode)
- Coordinate work via `datastore()`
- Spawn workers for async tasks
- Any normal Duso logic

## Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| Script-based handlers | Isolation per message, natural concurrency, no closure complexity |
| Reuse LSP transport | Already built, tested, handles JSON-RPC framing correctly |
| stdin/stdout only | Simpler than TCP, perfect for subprocess IPC, pipes are faster locally |
| Fresh evaluator per message | No shared state between handlers, clean isolation |
| Blocks on `.start()` | Matches `http_server()` pattern, natural event loop |
| Handler uses `context()` | Matches http_server, familiar pattern for Duso devs |

## Use Cases

1. **MCP Servers** - Duso as MCP server for Claude integration
2. **Custom JSON-RPC Servers** - LSP, DAP, or custom protocols
3. **Subprocess Coordination** - Parent process talks to Duso server via pipe
4. **Integration Layer** - Duso server wraps external services (Redis, SQLite) via pipe_server client calls

## Comparison with http_server

| Feature | http_server | pipe_server |
|---------|-------------|------------|
| Transport | HTTP over TCP/IP | JSON-RPC over stdin/stdout |
| Route matching | Path-based + method | Method-based (JSON-RPC) |
| Handler pattern | Script-based | Script-based |
| Context access | `context().request()` | `context().request()` |
| Response sending | `exit()` or response helpers | `exit()` or response helpers |
| Concurrency | Fresh evaluator per request | Fresh evaluator per message |
| Use case | Web services | Subprocess servers (MCP, LSP, etc) |

## Open Questions / Future Enhancements

- Bidirectional communication? (server → client push messages for MCP progress/sampling)
- Handler timeouts? (similar to `request_handler_timeout` in http_server)
- Named vs positional args for `.on(method, script)` vs `.on("method", "script")`?
- Auto-restart on handler crash? (or fail the request?)

## Files to Create/Modify

**Create:**
- `pkg/runtime/pipe_server.go` (core server, ~500 lines)
- `pkg/runtime/builtin_pipe_server.go` (Duso wrapper, ~200 lines)
- `pkg/runtime/builtin_pipe_server_test.go` (tests, ~400 lines)
- `docs/reference/pipe_server.md` (documentation, ~250 lines)
- `examples/core/pipe_server.du` (examples, ~100 lines)

**Modify:**
- `pkg/runtime/register.go` - Register `pipe_server` builtin
- `docs/learning-duso.md` - Add pipe_server section

## Timeline Estimate

- Phase 1 (Core): 3-4 hours (implement + test)
- Phase 2 (Docs + Examples): 1-2 hours
- Phase 3 (MCP Template): 1 hour (if starting with simple server)

**Total: ~1 day** for Phase 1 + 2
