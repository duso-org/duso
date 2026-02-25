# Ollama API Module for Duso

Access local LLMs through Ollama's OpenAI-compatible API.

## Setup

1. Install [Ollama](https://ollama.ai)
2. Start Ollama: `ollama serve`
3. Pull a model: `ollama pull mistral` (or any model you want)
4. Use in Duso:

```duso
ollama = require("ollama")
response = ollama.prompt("What is Ollama?")
print(response)
```

## Quick Start

```duso
ollama = require("ollama")

// One-shot query
response = ollama.prompt("Explain machine learning simply")

// Multi-turn conversation
chat = ollama.session({
  system = "You are a helpful coding assistant",
  model = "mistral"
})

response1 = chat.prompt("How do I write a loop?")
response2 = chat.prompt("Can you show an example?")
print(chat.usage)
```

## Endpoint

Default: `http://localhost:11434`

To use a different host/port, modify the endpoint in ollama.du.

## Available Models

Run `ollama list` to see what you have installed. Popular choices:

- `mistral` (default) - Fast and capable
- `llama2` - Meta's Llama 2
- `neural-chat` - Intel's model
- `orca-mini` - Small and quick
- `dolphin-mixtral` - Uncensored Mixtral

Pull new models: `ollama pull <model-name>`

## No API Key Required

Ollama runs locally - no authentication needed.

## Configuration Options

Same as OpenAI module - see [openai.md](/contrib/openai/openai.md) for full reference.

Key differences:
- No API key needed
- Default model: `mistral`
- Endpoint: `http://localhost:11434/v1/chat/completions`
- Runs completely offline

## See Also

- [openai.md](/contrib/openai/openai.md) - Full API documentation (identical interface)
- [Ollama](https://ollama.ai) - Download and documentation
- [Ollama Model Library](https://ollama.ai/library) - Browse available models
