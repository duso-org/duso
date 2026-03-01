# markdown_html()

Render markdown text to HTML.

## Signature

```duso
markdown_html(text, options)
```

## Parameters

- `text` (string) - Markdown text to render
- `options` (object, optional) - Enable/disable markdown extensions (all true by default):
  - `tables` (boolean) - Support markdown tables
  - `strikethrough` (boolean) - Support strikethrough text (~~text~~)
  - `footnotes` (boolean) - Support footnotes
  - `tasklists` (boolean) - Support task lists

## Returns

HTML string with rendered markdown content

## Examples

Basic markdown to HTML:

```duso
html = markdown_html("# Hello\n\n**Bold** and *italic* text")
print(html)
// <h1>Hello</h1><p><strong>Bold</strong> and <em>italic</em> text</p>
```

With tables:

```duso
markdown = """
| Name  | Age |
|-------|-----|
| Alice | 30  |
| Bob   | 25  |
"""

html = markdown_html(markdown, {tables = true})
print(html)
```

Disable extensions:

```duso
md = "This is ~~strikethrough~~ text"
html = markdown_html(md, {strikethrough = false})
```

## See Also

- [markdown_ansi() - Render to ANSI terminal output](/docs/reference/markdown_ansi.md)
