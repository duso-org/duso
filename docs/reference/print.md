# print()

Output values to stdout, separated by spaces, with a newline at the end.

## Signature

```duso
print(...args)
```

## Parameters

- `...args` - Any number of arguments of any type (numbers, strings, arrays, objects, etc.)

## Returns

`nil`

## Examples

Basic output:

```duso
print("Hello")                  // Output: "Hello"
```

Multiple arguments separated by spaces:

```duso
print("x:", 42)                 // Output: "x: 42"
print(1, 2, 3)                  // Output: "1 2 3"
```

Mixed types:

```duso
arr = [1, 2, 3]
obj = {name = "Alice", age = 30}
print("array:", arr)            // Output: "array: [1 2 3]"
print("object:", obj)           // Output: "object: {name=Alice age=30}"
print(true, false, nil)         // Output: "true false nil"
```

String templates:

```duso
name = "World"
count = 5
print("Hello {{name}}, count={{count}}")  // Output: "Hello World, count=5"
```

## Notes

- Each call to `print()` outputs a complete line (ends with newline)
- All values are automatically converted to strings for output
- Arguments are separated by single spaces in the output
- Useful for debugging, logging, and displaying results
- Works in all contexts (scripts, CLI, embedded Go applications)

## See Also

- [Type conversion with tostring()](./tostring.md)
