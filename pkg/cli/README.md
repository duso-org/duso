# CLI Package (`pkg/cli`)

This package contains **CLI-specific functions** and **script wrappers for runtime features**. Functions are available when:

1. **Using the duso CLI** - Automatically registered by `cmd/duso/main.go`
2. **Embedding with CLI features enabled** - Explicitly call `cli.RegisterFunctions()`

## Architecture

**Three-layer structure:**

```
pkg/script/              Core language (always available)
    ↓ (wraps)
pkg/runtime/             HTTP, datastore, concurrency (embeddable)
    ↓ (wraps)
pkg/cli/                 Script function wrappers + CLI features
```

- **`pkg/runtime/`** provides the implementation
- **`pkg/cli/`** provides the Duso script function wrappers and CLI-only features
- **`pkg/cli/register.go`** registers all functions in the interpreter

## Functions Provided

### Runtime Features (Embeddable, but exposed via CLI wrappers)

#### http_server(config)

Creates an HTTP server. Implementation in `pkg/runtime/http_server.go`.

```duso
server = http_server({port = 8080})
server.route("GET", "/", "handlers/home.du")
server.start()
```

#### http_client(config)

Creates an HTTP client. Implementation in `pkg/runtime/http_client.go`.

```duso
client = http_client({timeout = 30})
response = client.fetch({method = "GET", url = "https://example.com"})
```

#### datastore(namespace, config)

Creates a thread-safe datastore. Implementation in `pkg/runtime/datastore.go`.

```duso
store = datastore("myapp", {persist = "data.json"})
store.set("counter", 0)
store.increment("counter", 1)
```

#### spawn(script, context)

Runs a script asynchronously. Implementation in `pkg/runtime/goroutine_context.go`.

#### run(script, context)

Runs a script synchronously. Implementation in `pkg/runtime/goroutine_context.go`.

#### context()

Access request context. Implementation in `pkg/runtime/goroutine_context.go`.

### CLI-Only Functions

#### File I/O Functions

#### load(filename)

Reads a file's contents as a string. Files are resolved relative to the script's directory.

```duso
content = load("data.txt")
data = parse_json(load("config.json"))
```

#### save(filename, content)

Writes content to a file. Files are created relative to the script's directory. Parent directories are created if needed.

```duso
save("output.txt", "Hello, World!")
save("data.json", format_json(myData))
```

#### include(filename)

Loads and executes another Duso script in the current environment. Variables and functions from the included script are available in the current scope.

```duso
include("helpers.du")
result = helper_function()  // Now available
```

### Claude API Functions

#### conversation(system [, model] [, tokens] [, key])

Creates a multi-turn conversation with Claude that maintains context across multiple prompts.

```duso
agent = conversation(
    system = "You are a helpful assistant",
    model = "claude-opus-4-5-20251101"
)

response1 = agent.prompt("Hello!")
response2 = agent.prompt("How are you?")  // Context is maintained
```

Conversation objects have methods:
- `.prompt(message)` - Send a message and get a response
- `.system(prompt)` - Update the system prompt
- `.model(modelID)` - Change the Claude model
- `.key(apiKey)` - Set the API key
- `.tokens(maxTokens)` - Set max tokens per response
- `.clear()` - Clear conversation history
- `.usage()` - Get token usage statistics

#### claude(prompt [, model] [, tokens] [, key])

Single-shot query to Claude. Does not maintain context. Useful for one-off questions.

```duso
answer = claude("What is 2+2?")
code = claude("Write a hello world program", model = "claude-opus-4-5-20251101")
```

## Package Organization

### Files

**Runtime Feature Wrappers:**
- **`http_server.go`** - Wraps `pkg/runtime.HTTPServerValue`
- **`http.go`** - Wraps `pkg/runtime.HTTPClientValue`
- **`datastore.go`** - Wraps `pkg/runtime.GetDatastore()`
- **`run.go`** - Wraps `pkg/runtime` context management
- **`spawn.go`** - Wraps `pkg/runtime` context management
- **`context.go`** - Wraps `pkg/runtime` context access

**CLI-Only Features:**
- **`functions.go`** - File I/O (load, save, include, require, doc, env)
- **`module_resolver.go`** - Module path resolution (require, include)
- **`circular_detector.go`** - Circular dependency detection
- **`file_io_util.go`** - File I/O utilities
- **`register.go`** - Main registration function

**Integration:**
- **`register.go`** - Registers all functions (both runtime wrappers and CLI features)

### Key Types

- **`FileIOContext`** - Configuration for file I/O
- **`ModuleResolver`** - Path resolution for modules
- **`CircularDetector`** - Detects circular requires/includes
- **`RegisterOptions`** - Configuration for registering functions

## For Embedded Applications

If you're embedding Duso and want to enable CLI features:

```go
import "github.com/duso-org/duso/pkg/cli"

interp := script.NewInterpreter(false)

// Enable file I/O and Claude functions (optional)
cli.RegisterFunctions(interp, cli.RegisterOptions{
    ScriptDir: "/path/to/scripts",
})

// Now scripts can use: load(), save(), include(), claude(), conversation()
result, err := interp.Execute(`
    data = load("config.json")
    response = claude("Hello")
`)
```

Or implement your own versions with different behavior:

```go
// Custom load function with access control
interp.RegisterFunction("load", func(args map[string]any) (any, error) {
    filename := args["0"].(string)

    // Only allow loading from specific directory
    if !strings.HasPrefix(filename, "safe/") {
        return nil, fmt.Errorf("access denied")
    }

    // Your implementation
    return loadFile(filename), nil
})
```

## Design Principles

1. **Layered** - Wraps `pkg/runtime` features and adds CLI-specific functionality
2. **Optional** - Core language (`pkg/script`) doesn't depend on CLI
3. **Embeddable** - Runtime features can be used directly in Go apps
4. **Separated** - CLI-only features are clearly marked and isolated
5. **Composable** - Can be combined with custom functions
6. **Overridable** - Embedders can provide their own implementations of any function

## See Also

- [CLI User Guide](/docs/cli/) - Documentation for script writers
- [Embedding Guide](/docs/embedding/) - Documentation for Go developers
- [CONTRIBUTING.md](/CONTRIBUTING.md) - Where CLI code goes
