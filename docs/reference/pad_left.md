# pad_left()

Pad a string on the left with a character to reach a desired width.

## Signature

```duso
pad_left(str, width [, char])
```

## Parameters

- `str` (string) - The string to pad
- `width` (number) - Desired width in characters
- `char` (optional, string) - Character to pad with, default is space

## Returns

Padded string. If string is already at or longer than width, returns unchanged.

## Examples

Pad with spaces (default):

```duso
print("[" + pad_left("42", 5) + "]")     // "[   42]"
print("[" + pad_left("hello", 10) + "]") // "[     hello]"
```

Pad with custom character:

```duso
print("[" + pad_left("x", 5, "*") + "]") // "[****x]"
print("[" + pad_left("y", 5, "-") + "]") // "[----y]"
```

Already at width:

```duso
print(pad_left("hello", 5))              // "hello"
print(pad_left("hello", 3))              // "hello" (unchanged)
```

Format numbers with leading zeros:

```duso
hour = 5
minute = 8
time = "{{pad_left(tostring(hour), 2, \"0\")}}:{{pad_left(tostring(minute), 2, \"0\")}}"
print(time)                              // "05:08"
```

Works with UTF-8:

```duso
print("[" + pad_left("👍", 5) + "]")      // "[    👍]"
print(pad_left("x👍y", 8, "="))          // "x👍y====="
```

## Named Arguments

```duso
pad_left(str = "test", width = 10, char = "*")
```

## See Also

- [pad_right() - Pad on the right](/docs/reference/pad_right.md)
- [trim() - Remove whitespace](/docs/reference/trim.md)
