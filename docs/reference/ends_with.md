# ends_with()

Check if a string ends with a given suffix.

## Signature

```duso
ends_with(string, suffix [, ignore_case])
```

## Parameters

- `string` (string) - The string to check
- `suffix` (string) - The suffix to look for
- `ignore_case` (optional, boolean) - Case-insensitive matching. Default is false (case-sensitive)

## Returns

Boolean: true if string ends with suffix, false otherwise

## Examples

Literal string check (case-sensitive by default):

```duso
print(ends_with("hello.txt", ".txt"))           // true
print(ends_with("hello.txt", ".TXT"))           // false
print(ends_with("file", ".txt"))                // false
```

Literal string check (case-insensitive):

```duso
print(ends_with("hello.TXT", ".txt", ignore_case=true))  // true
print(ends_with("document.PDF", ".pdf", ignore_case=true))  // true
```

Using positional argument:

```duso
print(ends_with("style.CSS", ".css", true))     // true
```

Conditional checks:

```duso
path = "/home/user/document.txt"
if ends_with(path, ".txt") then
  print("Text file")
elseif ends_with(path, ".pdf", ignore_case=true) then
  print("PDF file")
end
```

## See Also

- [starts_with() - Check string prefix](/docs/reference/starts_with.md)
- [contains() - Check if contains pattern](/docs/reference/contains.md)
