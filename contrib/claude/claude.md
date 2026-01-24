# Claude API Module for Duso

Access Anthropic's Claude API directly from Duso scripts.

## Setup

Set your API key as an environment variable:

```bash
export ANTHROPIC_API_KEY=sk-ant-xxxxx
duso script.du
```

Or pass it explicitly in your script:

```duso
claude = require("claude")
response = claude.prompt("Hello", key = "sk-ant-xxxxx")
```

## Usage

### One-shot query

```duso
claude = require("claude")
response = claude.prompt("What is Duso?")
print(response)
```

### Multi-turn conversation

```duso
claude = require("claude")

chat = claude.session(system = "You are a helpful assistant")

response1 = chat.prompt("What is a closure?")
response2 = chat.prompt("Can you give me an example?")

// Context is maintained across prompts
print(chat.messages)     // Array of all messages
print(chat.usage)        // Token usage stats

chat.close()
```

## API Reference

## claude.prompt()

### Signature

`claude.prompt(message, system, model, max_tokens, key)`

Send a one-shot query to Claude.

**Parameters:**
- `message` (string, required) - Your prompt
- `system` (string, optional) - System prompt defining behavior
- `model` (string, optional) - Model ID (default: `claude-haiku-4-5-20251001`)
- `max_tokens` (number, optional) - Max tokens in response (default: 2048)
- `key` (string, optional) - API key (if not in `ANTHROPIC_API_KEY`)

**Returns:**
- `string` - Claude's response

**Example:**
```duso
claude = require("claude")
response = claude.prompt(
    message = "Write a haiku about Duso",
    system = "You are a poet",
    model = "claude-haiku-4-5-20251001",
    max_tokens = 100
)
```

## claude.models()

List all models available for account specified in API key.

**Parameters:**
- `key` (string, optional) - API key (if not in `ANTHROPIC_API_KEY`)

**Returns:**
- `array` - Array of objects with model info

**Example:**

```duso
// List all available models
models = claude.models()
print(models)
```

Get Haiku models:

```duso
models = claude.models()
haiku_models = filter(models, function(m)
  return contains(m.id, "haiku")
end)
print(haiku_models)
```

## Available Models

As of this doc, Anthropic has these as their latest.

- `claude-opus-4-5-20251101` Most capable, best for complex tasks
- `claude-sonnet-4-20250514` Fast and powerful
- `claude-haiku-4-5-20251001` Fast and affordable

See [Anthropic's models page](https://platform.claude.com/docs/about/models) for the latest.

## Environment Variables

- `ANTHROPIC_API_KEY` - Your API key (required if not passed as parameter)

## Examples

- `examples/simple.du` - One-shot query
- `examples/conversation.du` - Multi-turn conversation

## Error Handling

```duso
try
  claude = require("claude")
  response = claude.prompt("Hello")
  print(response)
catch (error)
  print("Error: " + error)
end
```

Common errors:
- Missing API key - Set `ANTHROPIC_API_KEY` or pass `key=` parameter
- Network error - Check internet connection
- Invalid model - Use a valid model ID
- Rate limit - Wait before retrying

## Pricing

Claude API uses pay-as-you-go pricing based on tokens:
- Input tokens - Cheaper
- Output tokens - More expensive

See [Anthropic pricing](https://www.anthropic.com/pricing) for current rates.

**Tips to reduce costs:**
- Use Haiku for simple tasks
- Set `max_tokens` to limit response length
- Reuse conversations instead of making separate calls
