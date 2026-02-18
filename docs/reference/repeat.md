# repeat()

Repeat a string multiple times.

## Signature

```duso
repeat(string, count)
```

## Parameters

- `string` (string) - The string to repeat
- `count` (number) - Number of times to repeat (must be non-negative)

## Returns

String repeated the specified number of times

## Examples

Basic repetition:

```duso
print(repeat("x", 5))               // "xxxxx"
print(repeat("ab", 3))              // "ababab"
```

Creating visual patterns:

```duso
print(repeat("-", 20))              // "--------------------"
print(repeat("*", 10))              // "**********"
```

Building output with repetition:

```duso
indent = repeat(" ", 4)
print(indent .. "item1")
print(indent .. "item2")
```

Zero and one repetitions:

```duso
print(repeat("x", 0))               // ""
print(repeat("hello", 1))           // "hello"
```

## See Also

- [substr() - Extract substring](/docs/reference/substr.md)
- [lower() - Convert to lowercase](/docs/reference/lower.md)
- [upper() - Convert to uppercase](/docs/reference/upper.md)
