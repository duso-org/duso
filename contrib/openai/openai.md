# OpenAI API Module for Duso

Access OpenAI's API directly from Duso scripts with an options-based, idiomatic interface.

## Setup

Set your API key as an environment variable:

```bash
export OPENAI_API_KEY=sk-proj-xxxxx
duso script.du
```

Or pass it explicitly in your script:

```duso
openai = require("openai")
response = openai.prompt("Hello", {key = "sk-proj-xxxxx"})
```

## Quick Start

### One-shot query

```duso
openai = require("openai")
response = openai.prompt("What is Duso?")
print(response)
```

### Multi-turn conversation

```duso
openai = require("openai")

chat = openai.session({
  system = "You are a helpful assistant"
})

response1 = chat.prompt("What is a closure?")
response2 = chat.prompt("Can you give me an example?")

print(chat.usage)  // Check token usage
```

### With temperature control

```duso
openai = require("openai")

// Lower temperature = more deterministic
response = openai.prompt("Solve this math problem: 2 + 2", {
  temperature = 0.5
})

// Higher temperature = more creative
response = openai.prompt("Write a poem about code", {
  temperature = 1.0
})
```

### With tools (Agent patterns)

```duso
openai = require("openai")

// Define a tool using openai.tool() helper - clean and simple!
var calculator = openai.tool({
  name = "calculator",
  description = "Performs basic math operations",
  parameters = {
    operation = {type = "string"},
    a = {type = "number"},
    b = {type = "number"}
  },
  required = ["operation", "a", "b"]
}, function(input)
  if input.operation == "add" then return input.a + input.b end
  if input.operation == "multiply" then return input.a * input.b end
end)

// Create agent - handlers are automatically extracted from tools!
agent = openai.session({
  tools = [calculator]
})

// Ask the agent - it will automatically call tools
response = agent.prompt("What is 15 * 27?")
print(response)  // "405"
```

## API Reference

### openai.prompt(message, config)

Send a one-shot query to OpenAI.

**Parameters:**
- `message` (string, required) - Your prompt
- `config` (object, optional) - Configuration options:
  - `system` - System prompt defining behavior
  - `model` - Model ID (default: `gpt-4o-mini`)
  - `max_tokens` - Max tokens in response (default: 2048)
  - `temperature` - Sampling temperature 0-2 (default: 1.0)
  - `top_p` - Nucleus sampling parameter
  - `key` - API key (if not in `OPENAI_API_KEY`)

**Returns:**
- `string` - Assistant's response

**Examples:**

```duso
// Basic
response = openai.prompt("What is the capital of France?")

// With system prompt
response = openai.prompt("Translate 'hello' to Spanish", {
  system = "You are a translator"
})

// With model override
response = openai.prompt("Solve this complex problem", {
  model = "gpt-4o",
  max_tokens = 4096
})

// With temperature
response = openai.prompt("Write a story", {
  temperature = 1.5
})
```

### openai.session(config)

Create a multi-turn conversation session.

**Parameters:**
- `config` (object, optional) - Configuration options (same as `prompt()` plus):
  - `tools` - Array of tool definitions (OpenAI format)
  - `tool_handlers` - Object mapping tool names to handler functions
  - `auto_execute_tools` - Auto-execute tools in response loop (default: true)
  - `tool_choice` - Tool selection strategy: `"auto"`, `"any"`, `"none"` (default: `"auto"`)

**Returns:**
- `session` object with methods:
  - `prompt(message)` - Send a message, returns text response
  - `add_tool_result(tool_call_id, result)` - Manually add tool result (for manual tool handling)
  - `continue_conversation()` - Continue conversation after manual tool result
  - `clear()` - Reset conversation and usage stats
  - `messages` - Array of all messages in conversation
  - `usage` - Token usage: `{input_tokens = N, output_tokens = M}`

**Examples:**

```duso
// Basic conversation
chat = openai.session({
  system = "You are a helpful assistant",
  temperature = 0.8
})

response1 = chat.prompt("Tell me about Duso")
response2 = chat.prompt("What are its main features?")

print(chat.usage)  // {input_tokens = 234, output_tokens = 567}

// With tools
agent = openai.session({
  tools = [my_tool],
  tool_handlers = {my_tool_name = my_handler_function},
  auto_execute_tools = true
})

response = agent.prompt("Use the tool to answer this")

// Manual tool handling
chat = openai.session({
  tools = [my_tool],
  tool_handlers = {},
  auto_execute_tools = false
})

response = chat.prompt("Use the tool")
// Process response.content manually
chat.add_tool_result(tool_call_id, result)
chat.continue_conversation()
```

### openai.tool(definition, handler)

Create a tool definition with optional handler function.

**Parameters:**
- `definition` (object, required) - Tool configuration:
  - `name` - Function name (required)
  - `description` - Function description
  - `parameters` - Object with parameter definitions (keys → type objects)
  - `required` - Array of required parameter names
- `handler` (function, optional) - Handler function that executes the tool
  - Receives `input` object with parameters
  - Returns result (automatically converted to string for API)

**Returns:**
- Tool object ready for `session()`'s `tools` array
- Handlers are automatically extracted and registered

**Examples:**

```duso
// Simple tool
var greet = openai.tool({
  name = "greet",
  description = "Greet someone",
  parameters = {name = {type = "string"}},
  required = ["name"]
}, function(input)
  return "Hello, " + input.name + "!"
end)

// Tool without handler (manual handling)
var info = openai.tool({
  name = "get_info",
  description = "Get information",
  parameters = {topic = {type = "string"}},
  required = ["topic"]
})

// Use in session
agent = openai.session({
  tools = [greet, info]
})
```

### openai.models(key)

List all models available for your account.

**Parameters:**
- `key` (string, optional) - API key (if not in `OPENAI_API_KEY`)

**Returns:**
- `array` - Array of model objects with `id`, `owned_by`, `created`, etc.

**Example:**

```duso
models = openai.models()
for i = 0; i < len(models); i = i + 1
  print(models[i].id)
end
```

## Configuration Options

All config options that can be passed to `prompt()` or `session()`:

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `model` | string | `gpt-4o-mini` | Model ID to use |
| `max_tokens` | number | 2048 | Maximum tokens in response |
| `temperature` | number | 1.0 | Sampling temperature (0-2) |
| `top_p` | number | nil | Nucleus sampling parameter (0-1) |
| `system` | string | nil | System prompt |
| `tools` | array | nil | Array of tool definitions (OpenAI format) |
| `tool_handlers` | object | {} | Map of tool names to handler functions |
| `auto_execute_tools` | bool | true | Auto-execute tools in response loop |
| `tool_choice` | string | `auto` | Tool selection: `auto`, `any`, `none` |
| `key` | string | nil | API key (uses `OPENAI_API_KEY` if not provided) |

## Available Models

As of 2025, OpenAI's latest models are:

- `gpt-4o` - Most capable
- `gpt-4-turbo` - Previous generation
- `gpt-4o-mini` - Affordable, fast
- `gpt-3.5-turbo` - Budget option

See [OpenAI's models page](https://platform.openai.com/docs/models) for the latest.

## Tool Use

Tools enable the assistant to take actions. Create them with `openai.tool()`:

```duso
var my_tool = openai.tool({
  name = "web_search",
  description = "Search the web",
  parameters = {
    query = {type = "string", description = "Search query"}
  },
  required = ["query"]
}, function(input)
  // Implement search
  return results
end)

// Handlers are automatically extracted!
chat = openai.session({
  tools = [my_tool]
})

response = chat.prompt("Search for Duso")
```

You can also mix wrapped tools with plain tool definitions:

```duso
var wrapped = openai.tool({...}, function(input) ... end)
var plain = {type = "function", "function" = {...}}

session = openai.session({
  tools = [wrapped, plain]
})
```

When `auto_execute_tools = true` (default), the assistant's tool calls are automatically executed and results integrated into the conversation. When `false`, you can process tool calls manually:

```duso
chat = openai.session({
  tools = [my_tool],
  auto_execute_tools = false
})

response = chat.prompt("Use the tool")

// Process manually
chat.add_tool_result(tool_call_id, result)
response = chat.continue_conversation()
```

## Temperature and Sampling

Control response creativity with temperature and sampling parameters:

- **`temperature`** (0-2): Controls randomness
  - 0 = Deterministic (best for analysis, math)
  - 0.5 = Balanced (good default)
  - 1.0 = Default/Creative (best for writing, brainstorming)
  - 2.0 = Maximum randomness

- **`top_p`** (0-1): Nucleus sampling - keeps top probability mass
  - Usually 0.9-1.0

Typical configurations:
```duso
// Analytical
{temperature = 0.5, top_p = 0.9}

// Balanced
{temperature = 1.0}

// Creative
{temperature = 1.5, top_p = 0.95}
```

## Environment Variables

- `OPENAI_API_KEY` - Your API key (required if not passed in config)

## Error Handling

```duso
try
  openai = require("openai")
  response = openai.prompt("Hello")
  print(response)
catch (error)
  print("Error: " + error)
end
```

Common errors:
- **Missing API key** - Set `OPENAI_API_KEY` or pass `key =` in config
- **Network error** - Check internet connection
- **Invalid model** - Use a valid model ID
- **Rate limit** - Wait before retrying
- **Insufficient funds** - Check your OpenAI account balance
- **Tool error** - Check tool handler implementation

## Pricing

OpenAI uses pay-as-you-go pricing based on tokens:
- Input tokens - Charged at lower rate
- Output tokens - Charged at higher rate

See [OpenAI pricing](https://openai.com/pricing) for current rates.

**Tips to reduce costs:**
- Use gpt-4o-mini for simple tasks
- Set `max_tokens` to limit response length
- Reuse conversations instead of making separate calls
- Use lower temperature for tasks that don't need creativity

## Examples

- `examples/simple.du` - One-shot queries with temperature
- `examples/conversation.du` - Multi-turn conversation
- `examples/tools.du` - Tool use and agent patterns

## Duso Idioms

This module uses Duso's best practices:

1. **Constructor pattern** - Options object for clean config merging
2. **Options objects** - Single config object instead of many parameters
3. **Idiomatic style** - Uses closures, functional patterns
4. **Proper error handling** - Try/catch for API and tool errors
5. **Resource efficiency** - Reusable sessions for multi-turn conversations

## OpenAI API Compatibility

This module uses OpenAI's native API format, making it compatible with:
- OpenAI's official API
- OpenAI-compatible providers (Groq, Mistral, Together, Anyscale, etc.)

To use a compatible provider, you typically only need to change the API endpoint. Future Duso provider modules will follow this same interface.

## Advanced: Manual Tool Handling

For complex agent patterns, disable auto execution and handle tools manually:

```duso
chat = openai.session({
  tools = [my_tool],
  auto_execute_tools = false
})

// In a loop:
response = chat.prompt(user_input)
tool_calls = extract_tool_calls(response)
if len(tool_calls) > 0 then
  for tool_call in tool_calls
    result = execute_tool(tool_call)
    chat.add_tool_result(tool_call.id, result)
  end
  response = chat.continue_conversation()
end
```

## Notes

- Each session maintains its own message history
- Tokens are tracked cumulatively across turns
- Tool definitions must follow OpenAI's JSON Schema format
- Tool handlers receive input object and should return result as string or number
- System prompt is sent with each request (affects token count)
- Tool calls are requested when `finish_reason == "tool_calls"`
