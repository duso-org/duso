# Duso `pipe_server()` - Implementation Plan

## Overview

Add a `pipe_server()` builtin that spawns a subprocess and provides bidirectional stdin/stdout communication via JSON-RPC. Mirrors the existing `http_server()` pattern but for subprocess I/O.

**Why:** Enables Duso as MCP middleware, aligns with existing LSP infrastructure, follows proven pattern (LSP, DAP, etc.).

## Architecture

### Go Layer (`pkg/runtime/pipe_server.go`)

```go
type PipeServer struct {
    cmd *exec.Cmd
    stdin io.WriteCloser
    stdout io.ReadCloser
    reader *bufio.Reader
    requestID int
}

func (ps *PipeServer) WriteJSON(v any) error       // Write JSON to stdin
func (ps *PipeServer) ReadJSON() (any, error)      // Read JSON from stdout
func (ps *PipeServer) Write(data []byte) error     // Write raw bytes
func (ps *PipeServer) Read() ([]byte, error)       // Read raw line
func (ps *PipeServer) Close() error                // Close subprocess
```

### Duso Builtin (`pkg/cli/pipe_server.go`)

Wraps Go layer, exposes as Duso object with methods:

```duso
proc = pipe_server(command [, {options}])

proc.write_json(object)  // Write JSON to subprocess stdin
proc.read_json()         // Read JSON from subprocess stdout
proc.write(string)       // Write raw string
proc.read()              // Read raw line
proc.close()             // Close subprocess
```

## API Design

```duso
// Initialize
proc = pipe_server("npx @anthropic-ai/inspector-mcp")
proc.write_json({jsonrpc = "2.0", method = "initialize", id = 1})
response = proc.read_json()

// Per-request handling
for request in requests do
  proc.write_json(request)
  response = proc.read_json()
  store.set("response_" + request.id, response)
end

// Cleanup
proc.close()
```

## Integration with Existing Patterns

**Request queue + delegator + handler loop:**

```duso
// Init subprocess once
proc = pipe_server("npx @anthropic-ai/inspector-mcp")

// Delegator loop (spawned background script)
while true do
  store.wait_for("request_queue", function(q) return len(q) > 0 end)
  requests = store.get("request_queue")

  for request in requests do
    proc.write_json(request)
    response = proc.read_json()
    store.set("response_" + request.id, response)
  end

  store.set("request_queue", [])
end
```

## Implementation Steps

### Phase 1: Core Runtime (Go)

1. `pkg/runtime/pipe_server.go`
   - `NewPipeServer(command)` - spawn subprocess, set up pipes
   - `.WriteJSON(v)` - marshal + write to stdin
   - `.ReadJSON()` - read line + unmarshal from stdout
   - Error handling, EOF detection

2. `pkg/cli/pipe_server.go`
   - Wrapper to expose as Duso builtin
   - Register in `pkg/cli/register.go`

3. Tests
   - Echo subprocess (test helper)
   - JSON round-trip
   - Multiple concurrent instances
   - Subprocess errors, timeouts, EOF

### Phase 2: Integration

1. Example: `examples/core/pipe_server.du`
   - Echo test
   - JSON-RPC mock MCP server

2. Documentation
   - `docs/reference/pipe_server.md`
   - Update `docs/learning-duso.md` subprocess section
   - Architecture notes for LSP parallel

### Phase 3: MCP Middleware (Pure Duso)

1. `contrib/mcp/` module
   - `client.du` - MCP client library
   - `delegator.du` - request handler loop
   - Examples showing MCP server wrapper

## Key Decisions

| Decision | Rationale |
|----------|-----------|
| JSON-RPC focus | Matches LSP, DAP, MCP standards |
| Line-delimited JSON | Simple, robust, works with all protocols |
| `.write_json()` / `.read_json()` | Type safety, automatic serialization |
| Subprocess lifecycle = script lifecycle | Simple, no external process management |
| No builtin connection pooling | Let user scripts manage via datastore |

## Use Cases

1. **MCP Middleware** - Route requests to backend MCP servers
2. **LSP Integration** - Programmatic language server control
3. **DAP Integration** - Debugger communication
4. **Generic subprocess IPC** - Any stdin/stdout protocol

## Open Questions

- Timeout handling? (e.g., `read_json(timeout = 5)`)
- Buffering strategy? (line-delimited vs. length-prefixed)
- Error propagation? (subprocess crashes, pipe breaks)
- Backward compatibility? (pure addition, no breaking changes)
