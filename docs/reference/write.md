# write()

Output values to stdout, separated by spaces, without adding a newline at the end.

## Signature

```duso
write(...args)
```

## Parameters

- `...args` - Any number of arguments of any type (numbers, strings, arrays, objects, etc.)

## Returns

`nil`

## Examples

Basic output without newline:

```duso
write("Hello")                  // Output: "Hello" (no newline)
write(" ")
write("World")                  // Output on same line: "Hello World"
```

Multiple arguments separated by spaces:

```duso
write("x: ", 42)                // Output: "x: 42" (no newline)
write(" done")
print("")                        // Add newline to complete the output
```

Mixed types:

```duso
arr = [1, 2, 3]
obj = {name = "Alice", age = 30}
write("array: ", arr)           // Output: "array: [1 2 3]" (no newline)
write(" | ")
write("object: ", obj)          // Output on same line: "array: [1 2 3] | object: {name=Alice age=30}" (no newline)
```

Using with busy():

```duso
write("Operation: ")
busy("processing")
sleep(2)
print("done")                   // Outputs newline and completes the line
```

String templates:

```duso
name = "Alice"
write("Hello {{name}}, ")       // Output: "Hello Alice, " (no newline)
print("welcome!")               // Output on same line: "Hello Alice, welcome!"
```

## Notes

- `write()` outputs text WITHOUT a trailing newline, unlike `print()`
- Use `write()` when you want to build up a line of output incrementally
- All values are automatically converted to strings for output
- Arguments are separated by single spaces in the output
- Useful for progress indicators, status messages, or prompts before `input()`
- Works in all contexts (scripts, CLI, embedded Go applications)

## See Also

- [print()](/docs/reference/print.md) - Output with automatic newline
- [busy()](/docs/reference/busy.md) - Display animated spinner with status message
- [input()](/docs/reference/input.md) - Read line from stdin
