# contains()

Check if a string contains a pattern or substring. Supports both literal strings and regular expressions.

## Signature

```duso
contains(string, pattern [, ignore_case])
```

## Parameters

- `string` (string) - The string to search in
- `pattern` (string or regex) - The pattern to find (regex or literal string)
- `ignore_case` (optional, boolean) - Case-insensitive matching. Default is false (case-sensitive)

## Returns

Boolean: true if pattern is found, false otherwise

## Examples

Literal string search (case-sensitive by default):

```duso
print(contains("Hello World", "World"))       // true
print(contains("Hello World", "world"))       // false
print(contains("hello", "HELLO"))             // false
```

Literal string search (case-insensitive):

```duso
print(contains("Hello World", "world", ignore_case=true))  // true
print(contains("Duso", "duso", ignore_case=true))          // true
```

Regex patterns with tilde syntax:

```duso
text = "Contact: user@example.com"
print(contains(text, ~\w+@\w+\.\w+~))  // true - email pattern
print(contains(text, ~\d+~))            // false - no digits

phone = "555-1234"
print(contains(phone, ~\d+~))           // true - has digits
```

Conditional checks:

```duso
email = "user@example.com"
if contains(email, "@") then
  print("Valid email format")
end
```

## See Also

- [find() - Find all matches](./find.md)
- [replace() - Replace matches](./replace.md)
