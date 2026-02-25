# Groq API Module for Duso

Access Groq's ultra-fast inference API directly from Duso scripts.

## Setup

Set your API key as an environment variable:

```bash
export GROQ_API_KEY=gsk_xxxxx
duso script.du
```

Or pass it explicitly:

```duso
groq = require("groq")
response = groq.prompt("Hello", {key = "gsk_xxxxx"})
```

## Quick Start

```duso
groq = require("groq")

// One-shot query
response = groq.prompt("What is Groq?")
print(response)

// Multi-turn conversation
chat = groq.session({
  system = "You are a helpful assistant",
  model = "mixtral-8x7b-32768"
})

response1 = chat.prompt("What is Groq known for?")
response2 = chat.prompt("What are its advantages?")
print(chat.usage)
```

## Available Models

- `mixtral-8x7b-32768` (default) - Fast, powerful open model
- `llama2-70b-4096` - Meta's Llama 2
- `gemma-7b-it` - Google's Gemma

See Groq's [models page](https://console.groq.com/docs/models) for latest list.

## Configuration Options

Same as OpenAI module - see [openai.md](/contrib/openai/openai.md) for full reference.

Key differences:
- API key environment variable: `GROQ_API_KEY` (not `OPENAI_API_KEY`)
- Default model: `mixtral-8x7b-32768`
- Endpoint: `https://api.groq.com/openai/v1/chat/completions`

## Environment Variables

- `GROQ_API_KEY` - Your API key (required if not passed in config)

## See Also

- [openai.md](/contrib/openai/openai.md) - Full API documentation (identical interface)
- [Groq Console](https://console.groq.com) - Get your API key
- [Groq Documentation](https://console.groq.com/docs/speech-text)
