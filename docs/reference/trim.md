# trim()

Remove leading and trailing whitespace from a string.

## Signature

```duso
trim(string)
```

## Parameters

- `string` (string) - The string to trim

## Returns

String with leading and trailing whitespace removed

## Examples

Remove spaces:

```duso
print(trim("  hello  "))        // "hello"
print(trim("  world"))          // "world"
print(trim("test  "))           // "test"
```

Tabs and newlines:

```duso
print(trim("\t hello \n"))      // "hello"
print(trim("\n\ntext\n\n"))     // "text"
```

Processing user input:

```duso
user_input = "  alice  "
username = trim(user_input)
if username == "alice" then
  print("Welcome, alice")
end
```
