# Claude Template

AI-powered applications using Claude integration:

- **require("claude")** - Load the Claude module
- **claude.prompt()** - One-shot prompts for quick answers
- **claude.conversation()** - Multi-turn conversations with context
- **System prompts** - Control Claude's behavior and role
- **JSON processing** - Send and receive structured data
- **Template strings** - Build dynamic prompts with data

## Setup

Set your Claude API key:

```bash
export ANTHROPIC_API_KEY="sk-ant-..."
```

## Running

```bash
duso claude.du
```

## Examples in main.du

1. **Simple Prompt** - Single question to Claude
2. **Conversation** - Multi-turn conversation with system prompt
3. **JSON Processing** - Analyze structured data with AI

## Building AI Applications

Use Duso + Claude for:
- **Data analysis** - Extract insights from datasets
- **Code generation** - Generate scripts and functions
- **Content creation** - Write, edit, and refine text
- **Automation** - Intelligent workflows and orchestration
- **Agent swarms** - Coordinate multiple AI workers

## Learn More

- [Claude Module Documentation](/contrib/claude/claude.md)
- [Working with Claude](/docs/learning-duso.md#working-with-claude)
- [Parallel Execution](/docs/reference/parallel.md) - Run multiple Claude calls concurrently
