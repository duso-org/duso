# Runtime Package (`pkg/runtime`)

This package contains **embeddable runtime features** for Duso scripts. These features can be used:

1. **In the CLI** - Registered by `cmd/duso/main.go` via `pkg/cli/` wrappers
2. **In embedded Go applications** - Used directly or wrapped as custom functions
3. **In custom distributions** - Bundled into custom Duso binaries

## Features Provided

### HTTP Server

**Type:** `HTTPServerValue`

Create HTTP servers for handling requests. Routes are defined as Duso scripts.

```go
// Direct Go usage
server := &runtime.HTTPServerValue{
    Port: 8080,
    Timeout: 30 * time.Second,
}
server.ListenAndServe()

// Via CLI wrapper (pkg/cli/http_server.go)
// Available in Duso scripts as: http_server({port = 8080})
```

### HTTP Client

**Type:** `HTTPClientValue`

Make HTTP requests with built-in connection pooling.

```go
// Direct Go usage
client, _ := runtime.NewHTTPClient(map[string]any{
    "timeout": 30.0,
})
response, _ := client.Send(map[string]any{
    "method": "GET",
    "url": "https://example.com",
})

// Via CLI wrapper (pkg/cli/http.go)
// Available in Duso scripts as: http_client({timeout = 30})
```

### Datastore

**Type:** `DatastoreValue`

Thread-safe in-memory key-value store with optional persistence. Perfect for coordinating work across concurrent scripts.

```go
// Direct Go usage
store := runtime.GetDatastore("myapp", map[string]any{
    "persist": "data.json",
})
store.Set("counter", 0)
store.Increment("counter", 1)

// Via CLI wrapper (pkg/cli/datastore.go)
// Available in Duso scripts as: datastore("myapp", {persist = "data.json"})
```

**Features:**
- Atomic operations (Set, Get, Increment, Append, Delete)
- Condition variables (Wait, WaitFor)
- Disk persistence (optional)
- Per-namespace isolation
- Global registry for cross-script coordination

### Goroutine Context Management

**Type:** `RequestContext`

Store request-scoped context in goroutine-local storage. Used by `spawn()` and `run()` functions.

```go
gid := runtime.GetGoroutineID()
ctx, exists := runtime.GetRequestContext(gid)
if exists {
    // Access request-specific data
}

runtime.SetRequestContextWithData(gid, &runtime.RequestContext{}, data)
defer runtime.ClearRequestContext(gid)
```

**Use Cases:**
- HTTP request handling (route context)
- Spawned script context (parent-child data passing)
- Run/parallel execution (isolated contexts)

## Package Organization

### Files

- **`datastore.go`** - Thread-safe coordination primitive
- **`http_server.go`** - HTTP server implementation
- **`http_client.go`** - HTTP client implementation
- **`goroutine_context.go`** - Request context storage
- **`metrics.go`** - Runtime metrics/monitoring

### Key Types

- **`DatastoreValue`** - Namespace-scoped key-value store
- **`HTTPServerValue`** - HTTP server with route handling
- **`HTTPClientValue`** - HTTP client with connection pooling
- **`RequestContext`** - Request-scoped data container

## For Embedded Applications

If you're embedding Duso and want runtime features:

### Option 1: Use CLI Wrappers (Recommended)

```go
import "github.com/duso-org/duso/pkg/cli"

interp := script.NewInterpreter(false)

// This registers runtime features as script functions
cli.RegisterFunctions(interp, cli.RegisterOptions{
    ScriptDir: "/path/to/scripts",
})

// Now scripts can use:
// - http_server() and http_client()
// - datastore()
// - spawn() and run()
// - context()
result, err := interp.Execute(`
    store = datastore("myapp")
    store.set("data", 42)
`)
```

### Option 2: Direct Go Usage

```go
import "github.com/duso-org/duso/pkg/runtime"

// Create and use datastore directly
store := runtime.GetDatastore("myapp", map[string]any{
    "persist": "data.json",
})
store.Set("key", "value")

// Create HTTP server directly
server := &runtime.HTTPServerValue{
    Port: 8080,
}
// ... configure routes, start server
```

### Option 3: Wrap for Script Use

```go
// Create custom wrapper function
interp.RegisterFunction("my_store", func(args map[string]any) (any, error) {
    namespace := args["0"].(string)
    store := runtime.GetDatastore(namespace, map[string]any{})

    // Return object with methods
    return map[string]any{
        "get": func(args map[string]any) (any, error) {
            key := args["0"].(string)
            return store.Get(key)
        },
        "set": func(args map[string]any) (any, error) {
            key := args["0"].(string)
            val := args["1"]
            return nil, store.Set(key, val)
        },
    }, nil
})

// Use in script
result, _ := interp.Execute(`
    s = my_store("test")
    s.set("x", 10)
    print(s.get("x"))
`)
```

## Design Principles

1. **Embeddable** - No dependencies on CLI or file I/O
2. **Concurrent** - Safe for use across goroutines
3. **Flexible** - Can be used directly in Go or wrapped for scripts
4. **Observable** - Integrates with Duso's error handling and call stacks
5. **Persistent** - Optional disk persistence for coordination

## Thread Safety

All runtime types are thread-safe:

- **Datastore**: Uses `sync.RWMutex` and `sync.Cond` for safe concurrent access
- **HTTP Client**: Connection pooling is thread-safe
- **HTTP Server**: Request handlers run in separate goroutines
- **Goroutine Context**: Uses `sync.Map` for safe concurrent access

## See Also

- [CLI Package](/pkg/cli/README.md) - Script function wrappers
- [Script Package](/pkg/script/README.md) - Language core
- [Embedding Guide](/docs/embedding/) - Using in Go applications
- [Learning Duso](/docs/learning-duso.md) - Language reference
