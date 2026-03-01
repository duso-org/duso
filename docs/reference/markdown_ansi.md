# markdown_ansi()

Render markdown text to ANSI-formatted output for terminal display with colors and styling.

## Signature

```duso
markdown_ansi(text, theme)
```

## Parameters

- `text` (string) - Markdown text to render
- `theme` (object, optional) - Custom theme with ANSI color codes:
  - `h1`, `h2`, `h3`, `h4`, `h5`, `h6` (string) - Heading styles
  - `code_start` (string) - Code block prefix
  - `code_inline` (string) - Inline code style
  - `blockquote` (string) - Blockquote style
  - `list_item` (string) - List item prefix
  - `bold` (string) - Bold text style
  - `italic` (string) - Italic text style
  - `link` (string) - Link style
  - `hr` (string) - Horizontal rule
  - `reset` (string) - Reset styling

## Returns

String with ANSI color codes for terminal rendering

## Examples

Render with default theme:

```duso
md = "# Title\n\nSome **bold** text"
output = markdown_ansi(md)
print(output)
```

Custom theme with colors:

```duso
ansi = require("ansi")

custom_theme = {
  h1 = ansi.combine(fg="red", bold=true),
  h2 = ansi.combine(fg="blue"),
  bold = ansi.combine(bold=true),
  code_inline = ansi.combine(fg="yellow"),
  reset = ansi.clear
}

output = markdown_ansi("# Red Title\n\n**Bold** and `code`", custom_theme)
print(output)
```

## See Also

- [markdown_html() - Render to HTML](/docs/reference/markdown_html.md)
- [ansi module](/stdlib/ansi/ansi.md) - Create custom ANSI themes
