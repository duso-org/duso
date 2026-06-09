# pad_right()

Pad a string on the right with a character to reach a desired width.


`pad_right(str, width [, char])`

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
print("[" + pad_right("42", 5) + "]")     // "[42   ]"
print("[" + pad_right("hello", 10) + "]") // "[hello     ]"
```

Pad with custom character:

```duso
print("[" + pad_right("x", 5, "-") + "]") // "[x----]"
print("[" + pad_right("y", 5, "=") + "]") // "[y====]"
```

Already at width:

```duso
print(pad_right("hello", 5))              // "hello"
print(pad_right("hello", 3))              // "hello" (unchanged)
```

Format table columns:

```duso
headers = ["Name", "Age", "City"]
formatted = [
  pad_right(headers[0], 10),
  pad_right(headers[1], 5),
  pad_right(headers[2], 8)
]
print("{{formatted[0]}} | {{formatted[1]}} | {{formatted[2]}}")
// Output: "Name       | Age   | City    "
```

Works with UTF-8:

```duso
print("[" + pad_right("👍", 5) + "]")      // "[👍    ]"
print(pad_right("x👍y", 8, "="))          // "x👍y====="
```

## Named Arguments

```duso
pad_right(str = "test", width = 10, char = "*")
```

## See Also

- [pad_left() - Pad on the left](/docs/reference/pad_left.md)
- [trim() - Remove whitespace](/docs/reference/trim.md)
