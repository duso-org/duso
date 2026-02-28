# starts_with()

Check if a string starts with a given prefix.

## Signature

```duso
starts_with(string, prefix [, ignore_case])
```

## Parameters

- `string` (string) - The string to check
- `prefix` (string) - The prefix to look for
- `ignore_case` (optional, boolean) - Case-insensitive matching. Default is false (case-sensitive)

## Returns

Boolean: true if string starts with prefix, false otherwise

## Examples

Literal string check (case-sensitive by default):

```duso
print(starts_with("Hello World", "Hello"))      // true
print(starts_with("Hello World", "hello"))      // false
print(starts_with("hello", "HELLO"))            // false
```

Literal string check (case-insensitive):

```duso
print(starts_with("Hello World", "hello", ignore_case=true))  // true
print(starts_with("Duso", "duso", ignore_case=true))          // true
```

Using positional argument:

```duso
print(starts_with("file.txt", "file", true))    // true
```

Conditional checks:

```duso
filename = "report_2024.pdf"
if starts_with(filename, "report_") then
  print("Found report file")
end
```

## See Also

- [ends_with() - Check string suffix](/docs/reference/ends_with.md)
- [contains() - Check if contains pattern](/docs/reference/contains.md)
