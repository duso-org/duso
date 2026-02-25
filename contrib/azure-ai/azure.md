# Azure OpenAI API Module for Duso

Access Azure's hosted OpenAI models (GPT-4, Claude, etc.) from Duso scripts.

## Setup

Set your Azure credentials as environment variables:

```bash
export AZURE_OPENAI_RESOURCE_NAME=your-resource-name
export AZURE_OPENAI_DEPLOYMENT_ID=your-deployment-id
export AZURE_OPENAI_API_KEY=your-api-key
duso script.du
```

Or configure at runtime:

```duso
azure = require("azure")
azure.set_resource("your-resource", "your-deployment")
response = azure.prompt("Hello", {key = "your-api-key"})
```

## Quick Start

```duso
azure = require("azure")

// One-shot query
response = azure.prompt("What is Azure?")
print(response)

// Multi-turn conversation
chat = azure.session({
  system = "You are a helpful assistant"
})

response1 = chat.prompt("Tell me about Azure OpenAI")
response2 = chat.prompt("What models are available?")
print(chat.usage)
```

## Configuration

Azure requires three pieces of information:

1. **Resource Name** - Your Azure resource name (from Azure portal)
2. **Deployment ID** - Your deployment name (e.g., "gpt-4")
3. **API Key** - Your Azure OpenAI API key

Set these via:
- Environment variables (recommended)
- `azure.set_resource(resource_name, deployment_id)` + config
- Config object in each call

## Configuration Options

Same as OpenAI module - see [openai.md](/contrib/openai/openai.md) for full reference.

Key differences:
- Environment variables: `AZURE_OPENAI_RESOURCE_NAME`, `AZURE_OPENAI_DEPLOYMENT_ID`, `AZURE_OPENAI_API_KEY`
- Uses deployment IDs instead of model names
- Different authentication header format
- Endpoint: `https://{resource}.openai.azure.com/openai/deployments/{deployment}/chat/completions`

## Environment Variables

- `AZURE_OPENAI_RESOURCE_NAME` - Your Azure resource name
- `AZURE_OPENAI_DEPLOYMENT_ID` - Your deployment ID
- `AZURE_OPENAI_API_KEY` - Your API key (required if not passed in config)

## See Also

- [openai.md](/contrib/openai/openai.md) - Full API documentation (identical interface)
- [Azure OpenAI Documentation](https://learn.microsoft.com/en-us/azure/ai-services/openai/)
- [Azure Portal](https://portal.azure.com) - Create and manage resources
