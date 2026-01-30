# CLI Package (`pkg/cli`)

This package contains **CLI-specific functions** for Duso scripts. These functions are NOT part of the core language and are only available when:

1. **Using the duso CLI** - Automatically registered by `cmd/duso/main.go`
2. **Embedding with CLI features enabled** - Explicitly call `cli.RegisterFunctions()`

## Functions Provided

### File I/O Functions

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

- **`functions.go`** - File I/O functions (load, save, include)
- **`conversation.go`** - Claude API functions (claude, conversation)
- **`register.go`** - Main registration function for embedders

### Key Types

- **`FileIOContext`** - Holds file I/O configuration (script directory)
- **`ConversationManager`** - Manages active conversations
- **`RegisterOptions`** - Configuration for registering CLI functions

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

1. **Optional** - Core language (pkg/script) doesn't depend on these functions
2. **Separated** - CLI functions are clearly separate from language core
3. **Composable** - Can be combined with custom functions
4. **Overridable** - Embedders can provide their own implementations

## See Also

- [CLI User Guide](/docs/cli/) - Documentation for script writers
- [Embedding Guide](/docs/embedding/) - Documentation for Go developers
- [CONTRIBUTING.md](/CONTRIBUTING.md) - Where CLI code goes
