# lower()

Convert a string to lowercase.

## Signature

```duso
lower(string)
```

## Parameters

- `string` (string) - The string to convert

## Returns

String converted to lowercase

## Examples

Basic conversion:

```duso
print(lower("HELLO"))           // "hello"
print(lower("Hello World"))     // "hello world"
```

Case-insensitive comparison:

```duso
user_input = "DUSO"
command = lower(user_input)
if command == "duso" then
  print("Found it!")
end
```

Processing mixed case:

```duso
email = "User@Example.COM"
normalized = lower(email)
print(normalized)               // "user@example.com"
```

## See Also

- [upper() - Convert to uppercase](./upper.md)
- [String functions](../language-spec.md#string-functions)
