# markdown

Render markdown text to HTML or ANSI-formatted output for terminal display.

## Signature

```duso
markdown = require("markdown")
markdown.parse(text)                   // → HTML
markdown.parse_ansi(text [, theme])    // → ANSI
```

## Module Functions

### parse(text)

Render markdown to HTML.

**Parameters:**
- `text` (string) - Markdown text to render

**Returns:**
- HTML string with rendered content

**Example:**
```duso
markdown = require("markdown")
html = markdown.parse("# Hello\n\n**Bold** text")
print(html)  // <h1>Hello</h1><p><strong>Bold</strong> text</p>
```

### parse_ansi(text [, theme])

Render markdown to ANSI-formatted output for terminal display.

**Parameters:**
- `text` (string) - Markdown text to render
- `theme` (object, optional) - Custom theme object with color codes

**Returns:**
- String with ANSI color codes for terminal rendering

**Example:**
```duso
markdown = require("markdown")
docs = doc("split")
ansi_output = markdown.parse_ansi(docs)
print(ansi_output)
```

## Supported Markdown

- **Headers** (# ## ### etc) - Yellow underlined, sized by level
- **Code blocks** (``` ... ```) - Green, indented
- **Inline code** (` ... `) - Bright green
- **Bold** (**text**) - Bright bold
- **Italic** (*text*) - Italic
- **Links** ([text](url)) - Blue bold underlined
- **Lists** (- or * items) - With nested bullet characters (•, ◦, ▪, ◈)
- **Blockquotes** (> or >> nesting) - Gray italic
- **Horizontal rules** (---, ***, ___) - Gray

## Custom Themes

Create a custom theme object for parse_ansi():

```duso
markdown = require("markdown")
ansi = require("ansi")

custom_theme = {
  h1 = ansi.combine(fg="red", bold=true),
  h2 = ansi.combine(fg="red"),
  code_inline = ansi.combine(fg="blue"),
  link = ansi.combine(fg="cyan", underline=true),
  reset = ansi.clear
}

ansi_output = markdown.parse_ansi(text, custom_theme)
```

## See Also

- [ansi module](../ansi/) - Create custom color themes
- [doc() builtin](../reference/doc.md) - Access documentation
