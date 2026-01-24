# markdown()

Render markdown text to ANSI color codes for terminal display. Available in `duso` CLI only.

## Signature

```duso
markdown(text)
```

## Parameters

- `text` (string) - Markdown text to render

## Returns

String with ANSI color codes applied for terminal rendering.

## Examples

Render documentation:

```duso
docs = doc("split")
print(markdown(docs))
```

Render Claude API responses:

```duso
response = claude("Explain closures in Duso")
print(markdown(response))
```

Render markdown files:

```duso
readme = load("README.md")
print(markdown(readme))
```

## Supported Markdown

- Headers (# ## ### etc) - Underlined, with colors based on level
- Code blocks (``` ... ```) - Bright cyan
- Inline code (` ... `) - Bright cyan
- Bold (**text**) - Bold white
- Italic (*text*) - Bold white
- Links ([text](url)) - Bright green with underline
- Regular text - White
