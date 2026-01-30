# Comments

Comments are text in your code that are ignored by the interpreter. Use them to document what your code does.

## Single-Line Comments

Single-line comments start with `//` and continue to the end of the line:

```duso
// This is a comment
x = 5

// You can have multiple comment lines
// This is clear documentation
function process(data)
  return data * 2
end

// Inline comments work too
y = x + 10  // Add ten to x
```

## Multi-Line Comments

Multi-line comments use `/* ... */` syntax and span multiple lines:

```duso
/* This is a block comment
   that spans multiple lines
   and is useful for longer explanations */

/*
  You can format them nicely:
  - First point
  - Second point
  - Third point
*/
print("Hello")
```

## Nested Comments

Multi-line comments support nesting, which is useful for commenting out code blocks that already contain comments:

```duso
function helper()
  // This function does something useful
  return 42
end

/*
  Commenting out a block of code:

  result = helper()
  // Process the result
  print(result)
*/
```

Without nesting support, the inner `//` comment would break the outer block comment. Duso handles this gracefully.

## Best Practices

- Use `//` for quick notes and inline comments
- Use `/* ... */` for longer explanations and documentation blocks
- Comment the "why", not just the "what"—explain intent rather than obvious code

## See Also

- [Variables and Types](/docs/learning_duso.md#comments)
