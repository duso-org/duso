# split()

Split a string into an array of substrings based on a separator. The separator can be a literal string or a regex pattern.

`split(string, separator [, ignore_case])`

## Parameters

- `string` (string) - The string to split
- `separator` (string or regex) - The delimiter to split on. A string is matched literally; a `~...~` regex splits on every match
- `ignore_case` (optional, boolean) - Case-insensitive matching, for both string and regex separators. Default is false

## Returns

Array of strings

## Examples

Basic splitting:

```duso
parts = split("a,b,c", ",")
print(parts[0])                 // Output: a
print(parts[1])                 // Output: b
print(len(parts))               // Output: 3
```

Splitting with whitespace:

```duso
words = split("hello world from duso", " ")
print(words)                    // Output: ["hello", "world", "from", "duso"]
```

Splitting CSV-like data:

```duso
csv_line = "Alice,30,Engineer"
fields = split(csv_line, ",")
print(fields[0])                // Output: Alice
print(fields[1])                // Output: 30
print(fields[2])                // Output: Engineer
```

Working with multiline strings:

```duso
text = "line1
line2
line3"
lines = split(text, "\n")
print(len(lines))               // Output: 3
```

Splitting on a regex pattern:

```duso
words = split("one   two \t three", ~\s+~)
print(len(words))               // Output: 3

fields = split("a1b22c333d", ~\d+~)
print(fields)                   // Output: ["a", "b", "c", "d"]

parts = split("2026-09-04T12:30:00", ~[-T:]~)
print(parts[0])                 // Output: 2026
```

Trimming around delimiters in one pass:

```duso
tags = split("go , rust,  duso ", ~\s*,\s*~)
print(tags[1])                  // Output: rust
```

Case-insensitive separator, string or regex:

```duso
print(split("oneXtwoxthree", "x"))                        // Output: ["oneXtwo", "three"]
print(split("oneXtwoxthree", "x", ignore_case = true))    // Output: ["one", "two", "three"]
print(split("a-B-c", "b", ignore_case = true))            // Output: ["a-", "-c"]
print(split("oneXtwoxthree", "x", true))                  // positional also works

print(split("oneXtwoxthree", ~x~))                        // Output: ["oneXtwo", "three"]
print(split("oneXtwoxthree", ~x~, ignore_case = true))    // Output: ["one", "two", "three"]
```

## Edge Cases

Empty separator:

```duso
result = split("abc", "")       // Splits into individual characters
print(result)                   // Output: ["a", "b", "c"]
```

A string separator is always literal, never a pattern:

```duso
print(split("a.b.c", "."))      // Output: ["a", "b", "c"] (literal dot)
print(split("a.b.c", ~\.~))     // Output: ["a", "b", "c"] (escaped regex dot)
print(split("abc", ~.~))        // Output: ["", "", "", ""] (unescaped . matches every char)
```

No matches:

```duso
result = split("hello", ",")
print(result)                   // Output: ["hello"] (returns array with original string)
```

## See Also

- [join() - Join array elements into string](/docs/reference/join.md)
- [find() - Find all regex matches in a string](/docs/reference/find.md)
- [replace() - Replace matches of a pattern](/docs/reference/replace.md)
