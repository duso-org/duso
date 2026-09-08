# DeepInfra API Module for Duso

Access DeepInfra's hosted open models directly from Duso scripts.

## Setup

Set your API key as an environment variable:

```bash
export DEEPINFRA_API_KEY=xxxxx
duso script.du
```

Or pass it explicitly:

```duso
deepinfra = require("deepinfra")
response = deepinfra.prompt("Hello", {key = "xxxxx"})
```

## Quick Start

```duso
deepinfra = require("deepinfra")

// One-shot query
response = deepinfra.prompt("What is DeepInfra?")
print(response)

// Multi-turn conversation
chat = deepinfra.session({
  system = "You are a helpful assistant",
  model = "deepseek-ai/DeepSeek-V4-Flash-0731"
})

response1 = chat.prompt("Tell me about open weight models")
response2 = chat.prompt("Which ones run cheapest?")
print(chat.usage)
```

## Available Models

DeepInfra hosts a large catalog, so pass the full model id including the vendor prefix:

- `deepseek-ai/DeepSeek-V4-Flash-0731` (default) - fast, 1M context
- `deepseek-ai/DeepSeek-V3.2` - general purpose
- `meta-llama/Meta-Llama-3.1-8B-Instruct-Turbo` - small and cheap
- `Qwen/Qwen3-235B-A22B-Instruct-2507` - large general purpose
- `google/gemma-4-31B-it` - mid-size instruct

List the live catalog with `models()`:

```duso
deepinfra = require("deepinfra")
for m in deepinfra.models() do print(m.id) end
```

Note that the catalog mixes chat, embedding, image, and audio models. Only ids
tagged `chat` work with `prompt()` and `session()`.

## Configuration Options

Same as OpenAI module - see [openai.md](/contrib/openai/openai.md) for full reference.

Key differences:
- API key environment variable: `DEEPINFRA_API_KEY`
- Default model: `deepseek-ai/DeepSeek-V4-Flash-0731`
- Endpoint: `https://api.deepinfra.com/v1/openai/chat/completions`
- Model ids include a vendor prefix (`vendor/model-name`)

## Environment Variables

- `DEEPINFRA_API_KEY` - Your API key (required if not passed in config)

## See Also

- [openai.md](/contrib/openai/openai.md) - Full API documentation (identical interface)
- [DeepInfra Dashboard](https://deepinfra.com/dash/api_keys) - Get your API key
- [DeepInfra Documentation](https://deepinfra.com/docs)
