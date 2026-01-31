# markdown

Render markdown text to HTML or ANSI-formatted output for terminal display.

> This is not a complete markdown implementation. But it is good for most purposes and it's easy to extend and bend to your wishes. If your application requires extensive markdown features, I suggest adding goldmark (https://github.com/yuin/goldmark) to the go layer and building a custom binary.

## Signature

```duso
markdown = require("markdown")
markdown.parse(text)                   // → HTML
markdown.parse_ansi(text [, theme])    // → ANSI
```

## parse()

Render markdown to HTML.

### Parameters:

- `text` (string) - Markdown text to render

### Returns:

- HTML string with rendered content

### Example:

```duso
markdown = require("markdown")
html = markdown.parse("# Hello\n\n**Bold** text")
print(html)

// prints <h1>Hello</h1><p><strong>Bold</strong>text</p>
```

## parse_ansi()

Render markdown to ANSI-formatted output for terminal display.

### Parameters:

- `text` (string) - Markdown text to render
- `theme` (object, optional) - Custom theme object with color codes

### Returns:

- String with ANSI color codes for terminal rendering

### Example:

```duso
markdown = require("markdown")
docs = doc("split")
ansi_output = markdown.parse_ansi(docs)
print(ansi_output)
```

# Custom Themes

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

# See Also

- [ansi module](/stdlib/ansi/ansi.md/) - Create custom color themes
- [doc() builtin](/doc/reference/doc.md) - Access documentation
