# highlight

Add syntax highlighting to code blocks in HTML.

## Signature

```duso
highlight = require("highlight")
highlight.html(html_string)  // → HTML with highlighted code blocks
```

## html()

Process HTML and add syntax highlighting to `<pre><code>` blocks.

### Parameters

- `html_string` (string) - HTML content with code blocks

### Returns

- HTML string with syntax highlighting applied to code blocks

### Example

```duso
highlight = require("highlight")

html = """
  <p>Here's some code:</p>
  <pre><code>function greet(name)
  return "Hello, " + name
end</code></pre>
"""

highlighted = highlight.html(html)
print(highlighted)

// Output will have <keyword>, <string>, <builtin> tags around highlighted code
```

## Styling

The highlighter uses custom HTML tags that can be styled with CSS:

- `<comment>` - Comments (`//` and `/* */`)
- `<string>` - Strings and regex literals
- `<keyword>` - Control flow keywords
- `<builtin>` - Built-in functions

Include this CSS to style:

```css
comment { color: #6a737d; font-style: italic; }
string { color: #22863a; }
keyword { color: #d73a49; font-weight: bold; }
builtin { color: #6f42c1; }
```

## Supported Languages

Currently supports **Duso** syntax highlighting (keywords, builtins, comments, strings, regex patterns).

## Notes

- Only processes `<pre><code>` blocks
- HTML entities in code are automatically handled
- Language detection not yet implemented (future enhancement)
