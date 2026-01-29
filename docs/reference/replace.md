# replace()

Replace all matches of a pattern in a string. Supports literal strings, regular expressions, and custom replacement functions.

## Signature

```duso
replace(string, pattern, replacement [, ignore_case])
```

## Parameters

- `string` (string) - The string to search in
- `pattern` (string or regex) - The pattern to find (regex or literal string)
- `replacement` (string or function) - Replacement value
  - If string: Replace each match with this literal string
  - If function: Called for each match with `(text, pos, len)` parameters
- `ignore_case` (optional, boolean) - Case-insensitive matching. Default is false (case-sensitive)

## Returns

New string with all matches replaced

## Examples

Simple string replacement:

```duso
text = "Hello hello hello"
result = replace(text, "hello", "hi")
print(result)  // "Hi Hi Hi" (case-insensitive by default)
```

Case-sensitive replacement:

```duso
text = "Hello hello HELLO"
result = replace(text, "hello", "hi", ignore_case=false)
print(result)  // "Hello hi HELLO" (only matches lowercase)
```

Regex replacement with literal string:

```duso
text = "Price: 10 dollars, quantity: 5 items"
result = replace(text, ~\d+~, "X")
print(result)  // "Price: X dollars, quantity: X items"
```

Dynamic replacement with function:

```duso
text = "I have 2 apples and 3 oranges"
result = replace(text, ~\d+~, function(text, pos, len)
  return tostring(tonumber(text) * 2)
end)
print(result)  // "I have 4 apples and 6 oranges"
```

Custom formatting with function:

```duso
text = "ID: 123, Amount: 456"
result = replace(text, ~\d+~, function(text, pos, len)
  return "[" + text + "]"
end)
print(result)  // "ID: [123], Amount: [456]"
```

## See Also

- [contains() - Check if pattern exists](./contains.md)
- [find() - Find all matches](./find.md)
