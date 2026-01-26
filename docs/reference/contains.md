# contains()

Check if a string contains a substring.

## Signature

```duso
contains(string, substring [, exact])
```

## Parameters

- `string` (string) - The string to search in
- `substring` (string) - The substring to find
- `exact` (optional, boolean) - Case-sensitive matching. Default is false (case-insensitive)

## Returns

Boolean: true if substring is found, false otherwise

## Examples

Case-insensitive search:

```duso
print(contains("hello", "HELLO"))       // true
print(contains("Hello World", "world")) // true
print(contains("Duso", "duso"))         // true
```

Case-sensitive search:

```duso
print(contains("hello", "HELLO", true))       // false
print(contains("Hello World", "world", true)) // false
print(contains("Hello World", "World", true)) // true
```

Conditional checks:

```duso
email = "user@example.com"
if contains(email, "@") then
  print("Valid email format")
end
```

## See Also

- [replace() - Replace substring](./replace.md)
