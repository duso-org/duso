# Implementation Plan: Runtime Limits and Caps

## Context

Duso currently has no limits on resource usage, creating potential for:
- Stack overflow from infinite recursion
- Memory exhaustion from unlimited goroutine spawning
- Runaway HTTP server connections
- Unbounded datastore growth
- Excessive string/array allocations

Adding configurable limits provides operational safety for production deployments while maintaining flexibility for power users.

## Limits to Implement

1. **Max recursion depth** - Prevent stack overflow
2. **Max spawned goroutines** - Prevent resource exhaustion
4. **Max HTTP connections** - Server stability
5. **Max datastore size** - Memory bounds per namespace
8. **Max string/array size** - Prevent huge allocations

## Implementation Strategy

### 1. Max Recursion Depth

**Files to modify:**
- `/Users/dbalmer/Projects/duso-org/duso/pkg/script/evaluator.go`
- `/Users/dbalmer/Projects/duso-org/duso/pkg/script/script.go`

**Approach:**
- Add `MaxRecursionDepth int` field to `Evaluator` struct (default: 10,000)
- In `callScriptFunction()` after `e.ctx.PushCall()` (line ~1067):
  ```go
  if e.ctx.Depth() > e.MaxRecursionDepth {
      return NewNil(), e.newError(
          fmt.Sprintf("maximum recursion depth exceeded (%d)", e.MaxRecursionDepth),
          callPos,
      )
  }
  ```
- Add `SetMaxRecursionDepth(depth int)` to public API

**Reuses:** Existing `ctx.Depth()` method from `ExecContext`

### 2. Max Spawned Goroutines

**Files to modify:**
- `/Users/dbalmer/Projects/duso-org/duso/pkg/runtime/builtin_spawn.go`
- `/Users/dbalmer/Projects/duso-org/duso/pkg/runtime/metrics.go`
- `/Users/dbalmer/Projects/duso-org/duso/pkg/script/builtins.go` (parallel())

**Approach:**
- Add `activeGoroutines atomic.Int32` to `SystemMetrics`
- Add `maxGoroutines int32` to `SystemMetrics` (default: 10,000)
- Increment counter before spawning:
  ```go
  if systemMetrics.activeGoroutines.Load() >= systemMetrics.maxGoroutines {
      return error("maximum concurrent goroutines exceeded")
  }
  systemMetrics.activeGoroutines.Add(1)
  defer systemMetrics.activeGoroutines.Add(-1)
  ```
- Apply in: `spawn()`, `run()`, `parallel()`, HTTP handler

**Reuses:** Existing `SystemMetrics` infrastructure

### 4. Max HTTP Connections

**Files to modify:**
- `/Users/dbalmer/Projects/duso-org/duso/pkg/runtime/http_server.go`

**Approach:**
- Add to `HTTPServerValue`:
  ```go
  MaxConnections    int32
  activeConnections atomic.Int32
  ```
- In `handleRequest()` before spawning handler (line ~404):
  ```go
  if s.activeConnections.Load() >= s.MaxConnections {
      http.Error(w, "Server capacity exceeded", http.StatusServiceUnavailable)
      return
  }
  s.activeConnections.Add(1)
  defer s.activeConnections.Add(-1)
  ```
- Default: 10,000 connections
- Configurable via `http_server({max_connections = N})`

**Reuses:** Existing HTTP request handling flow

### 5. Max Datastore Size

**Files to modify:**
- `/Users/dbalmer/Projects/duso-org/duso/pkg/script/value.go`
- `/Users/dbalmer/Projects/duso-org/duso/pkg/script/datastore_value.go`

**Approach:**

**Step A: Add Size() method to Value**
```go
// In value.go
func (v Value) Size() int64 {
    switch v.Type {
    case VAL_STRING:
        return int64(len(v.Data.(string)))
    case VAL_NUMBER:
        return 8 // float64
    case VAL_BOOL:
        return 1
    case VAL_ARRAY:
        arr := v.Data.(*[]Value)
        size := int64(len(*arr) * 24) // slice header + pointer overhead
        for _, elem := range *arr {
            size += elem.Size()
        }
        return size
    case VAL_OBJECT:
        obj := v.Data.(map[string]Value)
        size := int64(len(obj) * 48) // map overhead estimate
        for k, v := range obj {
            size += int64(len(k)) + v.Size()
        }
        return size
    default:
        return 0
    }
}
```

**Step B: Track size in DatastoreValue**
```go
// Add to DatastoreValue struct
totalSize   int64
maxSize     int64  // Optional limit (default: 1GB)
sizeMutex   sync.Mutex
```

**Step C: Update operations**
- `Set()`: Calculate new size, check limit before storing
- `Delete()`: Subtract deleted value size
- `Push()`: Add item size
- `Pop()/Shift()`: Subtract item size

### 8. Max String/Array Size

**Files to modify:**
- `/Users/dbalmer/Projects/duso-org/duso/pkg/script/value.go`
- `/Users/dbalmer/Projects/duso-org/duso/pkg/script/evaluator.go`

**Approach:**
- Add `MaxValueSize int64` to `Evaluator` (default: 1GB)
- Check size in `NewString()`, `NewArray()` constructors:
  ```go
  if size > evaluator.MaxValueSize {
      return error("value exceeds maximum size")
  }
  ```
- Also check during array push operations in builtins

## Configuration

All limits configurable via:

**Go API:**
```go
interp := script.NewInterpreter(false)
interp.SetMaxRecursionDepth(5000)
interp.SetMaxGoroutines(1000)
interp.SetMaxValueSize(100 * 1024 * 1024) // 100MB
```

**Environment variables (CLI):**
```bash
export DUSO_MAX_RECURSION_DEPTH=5000
export DUSO_MAX_GOROUTINES=1000
export DUSO_MAX_VALUE_SIZE=104857600
export DUSO_MAX_HTTP_CONNECTIONS=5000
export DUSO_MAX_DATASTORE_SIZE=1073741824
```

## Default Limits

- Max recursion depth: **10,000** calls
- Max goroutines: **10,000** concurrent
- Max HTTP connections: **10,000** per server
- Max datastore size: **1 GB** per namespace
- Max value size: **1 GB** per string/array

## Error Messages

Clear, actionable error messages:
- "maximum recursion depth exceeded (10000 calls)"
- "maximum concurrent goroutines exceeded (10000)"
- "HTTP server capacity exceeded (10000 connections)"
- "datastore size limit exceeded (1GB)"
- "value size exceeds maximum (1GB)"

## Testing Strategy

**Unit tests to add:**
1. Test recursion limit triggers correctly
2. Test goroutine spawning blocks at limit
3. Test HTTP connection rejection at capacity
4. Test datastore size enforcement
5. Test string/array size limits
6. Test that limits are configurable

**Integration tests:**
1. Spawn 100 goroutines, verify cleanup
2. Recursive function hits depth limit
3. HTTP server handles connection limit gracefully
4. Datastore approaches limit and rejects new data

## Critical Files

- `/Users/dbalmer/Projects/duso-org/duso/pkg/script/evaluator.go` - Recursion limit
- `/Users/dbalmer/Projects/duso-org/duso/pkg/runtime/builtin_spawn.go` - Goroutine limits
- `/Users/dbalmer/Projects/duso-org/duso/pkg/runtime/metrics.go` - Metrics tracking
- `/Users/dbalmer/Projects/duso-org/duso/pkg/runtime/http_server.go` - HTTP connection limits
- `/Users/dbalmer/Projects/duso-org/duso/pkg/script/datastore_value.go` - Datastore size tracking
- `/Users/dbalmer/Projects/duso-org/duso/pkg/script/value.go` - Value Size() method

## Verification

After implementation:

1. **Test recursion limit:**
   ```duso
   function recurse(n)
     return recurse(n + 1)
   end
   recurse(0)  // Should fail at ~10,000 depth
   ```

2. **Test goroutine limit:**
   ```duso
   for i = 1, 20000 do
     spawn("test.du", {})
   end  // Should fail after 10,000
   ```

3. **Test HTTP limit:**
   ```bash
   # Send 15,000 concurrent requests
   # Should reject after 10,000
   ```

4. **Test datastore size:**
   ```duso
   store = datastore("test")
   // Fill with large values until limit hit
   ```

5. **Test value size:**
   ```duso
   huge = range(1, 100000000)  // Should fail if exceeds limit
   ```

## Implementation Order

1. **Max recursion depth** (simplest, highest safety value)
2. **Max goroutines** (prevents resource exhaustion)
3. **Value Size() method** (foundation for datastore/value limits)
4. **Max HTTP connections** (server stability)
5. **Max datastore size** (uses Value.Size())
6. **Max string/array size** (uses Value.Size())

## Notes

- All limits should be opt-out (reasonable defaults)
- Limits can be set to 0 or -1 to disable
- Error messages include current limit value
- Metrics exposed for monitoring
- Limits are per-interpreter instance (embedded apps can customize)
