# split()

Split a string into an array of substrings based on a separator.

## Signature

```duso
split(string, separator)
```

## Parameters

- `string` (string) - The string to split
- `separator` (string) - The delimiter to split on

## Returns

Array of strings

## Examples

Basic splitting:

```duso
parts = split("a,b,c", ",")
print(parts[0])                 // Output: "a"
print(parts[1])                 // Output: "b"
print(len(parts))               // Output: 3
```

Splitting with whitespace:

```duso
words = split("hello world from duso", " ")
print(words)                    // Output: [hello world from duso]
```

Splitting CSV-like data:

```duso
csv_line = "Alice,30,Engineer"
fields = split(csv_line, ",")
print(fields[0])                // Output: "Alice"
print(fields[1])                // Output: "30"
print(fields[2])                // Output: "Engineer"
```

Working with multiline strings:

```duso
text = "line1
line2
line3"
lines = split(text, "\n")
print(len(lines))               // Output: 3
```

## Edge Cases

Empty separator:

```duso
result = split("abc", "")       // Splits into individual characters
print(result)                   // Output: [a b c]
```

No matches:

```duso
result = split("hello", ",")
print(result)                   // Output: [hello] (returns array with original string)
```

## See Also

- [join() - Join array elements into string](./join.md)
- [String functions in language-spec](../language-spec.md#string-functions)
