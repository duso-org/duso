# markdown_html()

Render markdown text to HTML.


`markdown_html(text, options)`

```

## Parameters

- `text` (string) - Markdown text to render
- `options` (object, optional) - Enable/disable markdown extensions:
  - `tables` (boolean) - Support markdown tables (default: true)
  - `strikethrough` (boolean) - Support strikethrough text (~~text~~) (default: true)
  - `highlight` (boolean) - Support highlighted text (==text==) (default: false)
  - `tasklists` (boolean) - Support task lists (default: true)
  - `smartquotes` (boolean) - Convert straight quotes to smart quotes (default: true)
  - `code_language` (boolean) - Add language classes to code blocks (default: true)

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

Highlight text:

```duso
md = "This is ==highlighted== text"

// Enable highlighting:
html = markdown_html(md, {highlight = true})
// <p>This is <mark>highlighted</mark> text</p>

// Without enabling, == is treated as literal:
html = markdown_html(md)
// <p>This is ==highlighted== text</p>
```

Disable extensions:

```duso
md = "This is ~~strikethrough~~ text"
html = markdown_html(md, {strikethrough = false})
```

Code blocks with language highlighting:

```duso
md = "```python\nprint('hello')\n```"
html = markdown_html(md)
// <pre><code class="language-python">print('hello')</code></pre>

// Disable language classes if not needed:
html = markdown_html(md, {code_language = false})
// <pre><code>print('hello')</code></pre>
```

## See Also

- [markdown_ansi() - Render to ANSI terminal output](/docs/reference/markdown_ansi.md)
