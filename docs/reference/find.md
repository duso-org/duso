# find()

Find all matches of a pattern in a string.

## Signature

```duso
find(string, pattern [, ignore_case])
```

## Parameters

- `string` (string) - The string to search in
- `pattern` (string or regex) - The pattern to find (regex or literal string)
- `ignore_case` (optional, boolean) - Case-insensitive matching. Default is false (case-sensitive)

## Returns

Array of match objects. Each object has:
- `text` (string) - The matched text
- `pos` (number) - Starting position of the match (0-indexed)
- `len` (number) - Length of the matched text

Returns empty array `[]` if no matches found.

## Examples

Find all digits:

```duso
text = "Price: 10 dollars, quantity: 5 items"
matches = find(text, ~\d+~)

for match in matches do
  print("Found", match.text, "at position", match.pos)
end
// Output:
// Found 10 at position 8
// Found 5 at position 35
```

Find all words:

```duso
text = "The quick brown fox"
matches = find(text, ~\w+~)
print("Found", len(matches), "words")  // Found 4 words

for match in matches do
  print(match.text)
end
```

Find email addresses:

```duso
text = "Contact alice@example.com or bob@test.org"
matches = find(text, ~\w+@\w+\.\w+~)

for match in matches do
  print("Email:", match.text)
end
```

## See Also

- [contains() - Check if pattern exists](./contains.md)
- [replace() - Replace matches](./replace.md)
