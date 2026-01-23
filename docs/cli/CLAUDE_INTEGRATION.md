# Claude API Integration

Call Claude directly from Duso scripts using `claude()` and `conversation()` functions.

## Setup

Set your API key as an environment variable:

```bash
export ANTHROPIC_API_KEY=sk-ant-xxxxx
duso script.du
```

Or pass it in the script:

```duso
response = claude("prompt", key = "sk-ant-xxxxx")
```

## claude(prompt [, model] [, tokens] [, key])

Single-shot query to Claude.

```duso
response = claude("What is 2+2?")
print(response)
```

**Parameters:**
- `prompt` (string, required) - Your prompt
- `model` (string, optional) - Model ID (default: configured model)
- `tokens` (number, optional) - Max tokens for response
- `key` (string, optional) - API key (if not in environment)

**Returns:**
- `string` - Claude's response

**Examples:**

```duso
// Simple question
answer = claude("What is the capital of France?")

// With model selection
answer = claude("Write code", model = "claude-opus-4-5-20251101")

// With token limit
summary = claude("Summarize this text", tokens = 500)

// All parameters
response = claude(
    prompt = "Analyze this data",
    model = "claude-haiku-4-5-20251001",
    tokens = 1000,
    key = "sk-ant-xxxxx"
)
```

**Available Models:**
- `claude-opus-4-5-20251101` - Most capable (Opus)
- `claude-sonnet-4-20250514` - Fast and powerful (Sonnet)
- `claude-haiku-4-5-20251001` - Fast and affordable (Haiku)

See [Anthropic models documentation](https://docs.anthropic.com/en/docs/about-claude/models/latest) for latest models.

---

## conversation(system [, model] [, tokens] [, key])

Create a stateful conversation that maintains context across multiple prompts.

```duso
agent = conversation(system = "You are a helpful assistant")
answer1 = agent.prompt("Hello!")
answer2 = agent.prompt("How are you?")
```

**Parameters:**
- `system` (string, required) - System prompt defining behavior
- `model` (string, optional) - Model ID
- `tokens` (number, optional) - Max tokens per response
- `key` (string, optional) - API key

**Returns:**
- `object` - Conversation object with `.prompt()` method

**Conversation Object Methods:**

### .prompt(message)

Send a message and get a response. Context is preserved.

```duso
agent = conversation(system = "You are a code reviewer")

review = agent.prompt("Review this function")
questions = agent.prompt("What about error handling?")
suggestions = agent.prompt("What's the top priority?")
```

---

## Common Patterns

### Text Processing

```duso
// Summarization
text = load("article.txt")
summary = claude("Summarize this article concisely:\n\n" + text)
save("summary.txt", summary)

// Translation
english = "Hello, world!"
spanish = claude("Translate to Spanish: " + english)
print(spanish)

// Sentiment analysis
sentiment = claude("Rate the sentiment of: '" + text + "' (positive/negative/neutral)")
print("Sentiment: " + sentiment)
```

### Data Analysis

```duso
data = load("data.json")

analysis = conversation(
    system = "You are a data analyst. Provide insights in JSON format."
)

findings = analysis.prompt("Analyze this data: " + data)
print(findings)

// Follow-up questions
trends = analysis.prompt("What are the key trends?")
forecast = analysis.prompt("Based on these trends, what should we expect next?")
```

### Code Generation

```duso
// Generate code
function_spec = "Create a function that validates email addresses"
code = claude(function_spec, model = "claude-opus-4-5-20251101")
save("validator.go", code)

// Code review
code_to_review = load("main.go")
review = claude("Review this code and suggest improvements:\n\n" + code_to_review)
save("code_review.md", review)
```

### Multi-Agent Workflow

```duso
// Different agents with different roles
researcher = conversation(system = "You are a researcher. Provide factual information.")
writer = conversation(system = "You are a technical writer. Write clearly and concisely.")
editor = conversation(system = "You are an editor. Improve text for clarity and readability.")

// Research phase
topic = "Quantum Computing"
facts = researcher.prompt("What are the key concepts in " + topic + "?")

// Write phase
article = writer.prompt("Write an introduction based on these concepts:\n\n" + facts)

// Edit phase
polished = editor.prompt("Edit and improve this article:\n\n" + article)

// Save result
save("article.md", polished)
```

### Batch Processing

```duso
// Process items using Claude
items = ["apple", "banana", "cherry"]
results = []

for item in items do
    response = claude("Give a fun fact about " + item)
    results = append(results, {
        item = item,
        fact = response
    })
end

// Save results
output = format_json(results)
save("facts.json", output)
```

### Interactive Agent

```duso
// Create an agent
assistant = conversation(
    system = "You are a helpful assistant that can answer questions and help with tasks."
)

// Multi-turn conversation
print("Agent ready. Ask questions (type 'exit' to quit)")

question = "What is machine learning?"
response = assistant.prompt(question)
print("Agent: " + response)

// Next question
followup = "How is it used in practice?"
response = assistant.prompt(followup)
print("Agent: " + response)
```

### Template Generation

```duso
// Generate templates
email_template = claude(
    "Generate a professional email template for requesting a meeting",
    tokens = 500
)
save("email_template.txt", email_template)

// Generate variations
templates = []
for style in ["formal", "casual", "technical"] do
    template = claude("Generate a " + style + " email template")
    templates = append(templates, {style = style, template = template})
end

save("email_templates.json", format_json(templates))
```

---

## Error Handling

```duso
try
    response = claude("prompt")
    print(response)
catch (err)
    print("Error: " + err)
end
```

Common errors:
- **API key not set** - Set ANTHROPIC_API_KEY
- **Network error** - Check internet connection
- **Rate limit** - Wait before retrying
- **Invalid model** - Use a valid model ID

---

## Performance Tips

1. **Use Haiku for simple tasks** - Faster and cheaper
2. **Use Opus for complex tasks** - More capable
3. **Set token limit** - Prevent excessive responses
4. **Cache responses** - Save to file if repeating
5. **Batch requests** - Process multiple items efficiently

```duso
// Cache API responses
cacheFile = "cache.json"
cache = {}

try
    cachedText = load(cacheFile)
    cache = parse_json(cachedText)
catch (err)
    cache = {}
end

// Check cache
function cached_claude(prompt)
    if cache[prompt] != nil then
        return cache[prompt]
    end

    response = claude(prompt)
    cache[prompt] = response
    save(cacheFile, format_json(cache))
    return response
end
```

---

## Costs

Claude API is pay-as-you-go based on tokens:
- **Input tokens** - Cheaper, counted from your prompt
- **Output tokens** - More expensive, counted in response

See [Anthropic pricing](https://www.anthropic.com/pricing) for current rates.

**Tips to reduce costs:**
- Use smaller models (Haiku) for simple tasks
- Set token limits with `tokens = N`
- Reuse conversation context instead of making new calls
- Cache responses when appropriate

---

## See Also

- [Getting Started](GETTING_STARTED.md) - Quick tutorial
- [File I/O](FILE_IO.md) - Reading/writing files
- [Language Reference](../language-spec.md) - Complete language spec
- [Anthropic API Docs](https://docs.anthropic.com/) - Official Claude API docs
