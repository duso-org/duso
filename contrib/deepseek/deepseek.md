# DeepSeek API Module for Duso

Access DeepSeek's LLMs directly from Duso scripts.

## Setup

Set your API key as an environment variable:

```bash
export DEEPSEEK_API_KEY=sk_xxxxx
duso script.du
```

Or pass it explicitly:

```duso
deepseek = require("deepseek")
response = deepseek.prompt("Hello", {key = "sk_xxxxx"})
```

## Quick Start

```duso
deepseek = require("deepseek")

// One-shot query
response = deepseek.prompt("What is DeepSeek?")
print(response)

// Multi-turn conversation
chat = deepseek.session({
  system = "You are a helpful assistant",
  model = "deepseek-chat"
})

response1 = chat.prompt("Tell me about DeepSeek")
response2 = chat.prompt("What makes it special?")
print(chat.usage)
```

## Available Models

- `deepseek-chat` (default) - General purpose
- `deepseek-coder` - Code generation

See [DeepSeek documentation](https://platform.deepseek.com/docs) for latest models.

## Configuration Options

Same as OpenAI module - see [openai.md](/contrib/openai/openai.md) for full reference.

Key differences:
- API key environment variable: `DEEPSEEK_API_KEY`
- Default model: `deepseek-chat`
- Endpoint: `https://api.deepseek.com/chat/completions`

## Environment Variables

- `DEEPSEEK_API_KEY` - Your API key (required if not passed in config)

## See Also

- [openai.md](/contrib/openai/openai.md) - Full API documentation (identical interface)
- [DeepSeek Platform](https://platform.deepseek.com) - Get your API key
- [DeepSeek Documentation](https://platform.deepseek.com/docs)
