# DigitalOcean Gradient AI Module for Duso

Access DigitalOcean's Serverless Inference from Duso scripts. One endpoint fronts
models from Anthropic, OpenAI, Meta, Mistral, DeepSeek, Alibaba, NVIDIA and others,
so switching vendors is a matter of changing the `model` string.

## Setup

Set your API key as an environment variable:

```bash
export DIGITAL_OCEAN_MODEL_ACCESS_KEY=dop_v1_xxxxx
duso script.du
```

Or pass it explicitly:

```duso
do_ai = require("digitalocean-ai")
response = do_ai.prompt("Hello", {key = "dop_v1_xxxxx"})
```

The key can be either a **model access key** (create one under Serverless Inference
in the DigitalOcean control panel) or a **personal access token**. The variable name
matches the one DigitalOcean's own docs and sample code use, so a key you already
exported while following their quickstart works here unchanged.

## Quick Start

```duso
do_ai = require("digitalocean-ai")

// One-shot query
response = do_ai.prompt("What is DigitalOcean Gradient AI?")
print(response)

// Multi-turn conversation
chat = do_ai.session({
  system = "You are a helpful assistant",
  model = "anthropic-claude-5-sonnet"
})

response1 = chat.prompt("Tell me about serverless inference")
response2 = chat.prompt("How does billing work?")
print(chat.usage)

// What can I actually call?
for m in do_ai.models() do
  print(m.id)
end
```

## Available Models

The catalog is large and changes over time. `models()` lists it, but note that it
returns DigitalOcean's **entire public catalog, not the models your key can use**.
A model access key is scoped to a subset of models when it is created, and that
scope is immutable afterward. An out-of-scope model still appears in `models()`
and returns a 403 on use, with an error that never names the model. The control
panel is the only place to see a key's actual scope: Serverless Inference → hover
the Models column.

A sample of the catalog as of August 2026:

| Model | Notes |
|---|---|
| `openai-gpt-4o-mini` | Default. Cheap, general purpose |
| `openai-gpt-5.5` | 1M context |
| `anthropic-claude-5-sonnet` | 1M context |
| `anthropic-claude-opus-5` | 1M context, strongest Anthropic option |
| `anthropic-claude-haiku-4.5` | Fast, 200K context |
| `deepseek-v4-pro` | 1M context |
| `llama-4-maverick` | Meta, 128K context |
| `alibaba-qwen3-32b` | Open weights |
| `glm-5.2` | Z.ai, 262K context |
| `nemotron-3-ultra-550b` | NVIDIA |

See [Supported Models](https://docs.digitalocean.com/products/inference/details/models/)
for the current list with context windows and output limits.

Models do get retired. If a call starts failing with a model error, check the catalog
and update the `model` in your config. A 403 means the opposite problem — the model
exists, but your key is not scoped to it.

## Configuration Options

Same as OpenAI module - see [openai.md](/contrib/openai/openai.md) for full reference.

Key differences:
- API key environment variable: `DIGITAL_OCEAN_MODEL_ACCESS_KEY`
- Default model: `openai-gpt-4o-mini`
- Endpoint: `https://inference.do-ai.run/v1/chat/completions`
- Model names are vendor-prefixed (`anthropic-claude-opus-5`, not `claude-opus-5`)

## Scope

This module covers chat completions, which is what the `prompt`/`session` interface
is built around. DigitalOcean's inference endpoint also exposes image generation,
text-to-speech, video generation and an Anthropic-shaped `/v1/messages` route; those
are not wrapped here. Reach them with `fetch()` if you need them.

DigitalOcean **Agents** — preconfigured server-side deployments with knowledge bases
and guardrails, living at `https://<agent-id>.agents.do-ai.run/api/v1/` — are also
OpenAI-compatible. They are built in the DigitalOcean control panel, not from Duso,
so this module does not wrap them. To talk to one you already created:

```duso
openai = require("openai")

agent = openai.create_client(
  "https://<agent-id>.agents.do-ai.run/api/v1/chat/completions",
  "https://<agent-id>.agents.do-ai.run/api/v1/models",
  {default_model = "n/a", key_env = "DIGITALOCEAN_AGENT_KEY"}
)

print(agent.prompt("What does our runbook say about failover?"))
```

Note that agents use their own endpoint access keys, separate from your inference key.

## Environment Variables

- `DIGITAL_OCEAN_MODEL_ACCESS_KEY` - Your model access key or personal access token (required if not passed in config)

## See Also

- [openai.md](/contrib/openai/openai.md) - Full API documentation (identical interface)
- [DigitalOcean Serverless Inference](https://docs.digitalocean.com/products/inference/) - Product docs
- [Supported Models](https://docs.digitalocean.com/products/inference/details/models/) - Current catalog
- [DigitalOcean Control Panel](https://cloud.digitalocean.com) - Create a model access key
