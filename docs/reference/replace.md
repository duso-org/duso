# replace()

Replace all occurrences of a substring with another string.

## Signature

```duso
replace(string, old, new [, exact])
```

## Parameters

- `string` (string) - The string to search in
- `old` (string) - The substring to find and replace
- `new` (string) - The replacement string
- `exact` (optional, boolean) - Case-sensitive matching. Default is false (case-insensitive)

## Returns

New string with all occurrences replaced

## Examples

Case-insensitive replacement:

```duso
text = "Hello hello HELLO"
result = replace(text, "hello", "hi")
print(result)                   // "hi hi hi"
```

Case-sensitive replacement:

```duso
text = "Hello hello HELLO"
result = replace(text, "hello", "hi", true)
print(result)                   // "Hello hi HELLO"
```

Multi-character replacement:

```duso
text = "The quick brown fox"
result = replace(text, "brown", "lazy")
print(result)                   // "The quick lazy fox"
```

## See Also

- [contains() - Check for substring](./contains.md)
