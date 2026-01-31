# doc()

Access documentation for modules and builtins. Available in `duso` CLI only.

## Signature

```duso
doc(name)
```

## Parameters

- `name` (string) - Name of module or builtin function

## Returns

Markdown documentation as string, or `nil` if not found

## Examples

Get module documentation and render with markdown module:

```duso
markdown = require("markdown")
docs = doc("http")
print(markdown.parse_ansi(docs))
```

Get builtin documentation:

```duso
markdown = require("markdown")
docs = doc("split")
ansi_output = markdown.parse_ansi(docs)
print(ansi_output)
```

Display documentation:

```duso
markdown = require("markdown")
func_name = "map"
docs = doc(func_name)
if docs then
  print(markdown.parse_ansi(docs))
else
  print("No documentation found for: " + func_name)
end
```

## Search Order

Searches for documentation in this order:
1. Module documentation (.md files matching modules)
2. Builtin reference documentation (docs/reference/)

## See Also

- [markdown module](/stdlib/markdown/markdown.md) - Render markdown to HTML or ANSI
- [require() - Load modules](/docs/reference/require.md)
- [CLI reference documentation](/docs/reference/)
