# markdown_text()

Render markdown text to plain text output without ANSI color codes or HTML tags.

## Signature

```duso
markdown_text(text)
```

## Parameters

- `text` (string) - Markdown text to render

## Returns

String with plain text only (no ANSI codes or HTML formatting)

## Examples

Basic markdown to plain text:

```duso
md = "# Title\n\nSome **bold** text and `code` examples"
output = markdown_text(md)
print(output)

// prints:
// Title
//
// Some bold text and code examples
```

Render documentation as plain text:

```duso
docs = doc("split")
plain = markdown_text(docs)
print(plain)
```

## See Also

- [markdown_html() - Render to HTML](/docs/reference/markdown_html.md)
- [markdown_ansi() - Render with colors](/docs/reference/markdown_ansi.md)
- [markdown module - Deprecated](/stdlib/markdown/markdown.md)
