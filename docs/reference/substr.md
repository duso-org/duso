# substr()

Extract a substring from a string.

## Signature

```duso
substr(string, start [, length])
```

## Parameters

- `string` (string) - The string to extract from
- `start` (number) - Starting position (0-indexed)
- `length` (optional, number) - Number of characters to extract. If omitted, extracts to end of string

## Returns

Extracted substring

## Examples

Extract from start position:

```duso
text = "hello world"
print(substr(text, 0, 5))       // "hello"
print(substr(text, 6))          // "world"
```

Negative indices from end:

```duso
text = "hello"
print(substr(text, -2))         // "lo"
print(substr(text, -3, 2))      // "ll"
```

Extract characters:

```duso
word = "duso"
print(substr(word, 1, 2))       // "us"
print(substr(word, 0, 1))       // "d"
```

## See Also

- [split() - Split string into array](/docs/reference/split.md)
