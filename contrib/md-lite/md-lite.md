# md-lite

Lightweight markdown to HTML formatter for Duso.

## Features

- **Code blocks** - Preserves formatting for triple-backtick blocks
- **Headers** - Converts `#`, `##`, `###` to styled HTML headers
- **Links** - Converts `[text](url)` to HTML `<a>` tags
- **Images** - Converts `![alt](url)` to HTML `<img>` tags

## Usage

```duso
formatter = require("md-lite")

markdown = """
# Hello

[Link](https://example.com)

\`\`\`duso
code block
\`\`\`
"""

html = formatter(markdown)
print(html)
```

## API

`format(markdown_string)` → string

Converts markdown string to HTML. Returns formatted HTML suitable for embedding in a page.

**Parameters:**
- `markdown_string` - Markdown text

**Returns:**
- HTML string with formatted content

## Example

```duso
formatter = require("md-lite")

content = """
  # Documentation

  See the [guide](/docs/guide.md) for more info.

  \`\`\`
  code example
  \`\`\`
"""

html = formatter(content)
```

## Limitations

- Does not parse all markdown syntax
- Intended for simple, lightweight formatting
- No support for tables, footnotes, or advanced features
