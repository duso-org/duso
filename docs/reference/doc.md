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

Get module documentation:

```duso
docs = doc("http")
print(markdown(docs))
```

Get builtin documentation:

```duso
docs = doc("split")
print(markdown(docs))
```

Display documentation:

```duso
func_name = "map"
docs = doc(func_name)
if docs then
  print(markdown(docs))
else
  print("No documentation found for: " + func_name)
end
```

## Search Order

Searches for documentation in this order:
1. Module documentation (.md files matching modules)
2. Builtin reference documentation (docs/reference/)

## See Also

- [markdown() - Format documentation](./markdown.md)
- [require() - Load modules](./require.md)
- [CLI reference documentation](./README.md)
