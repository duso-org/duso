# if

Conditional execution: run code only if a condition is true.

## Syntax

```duso
if condition then
  // statements
end

if condition then
  // statements
elseif other_condition then
  // other statements
else
  // fallback statements
end
```

## Description

The `if` statement evaluates a condition and executes the following block if the condition is truthy. Use `elseif` for additional conditions and `else` for a fallback.

All values have truthiness: `nil`, `false`, `0`, `""`, `[]`, and `{}` are falsy; everything else is truthy.

## Examples

Basic condition:

```duso
age = 25
if age >= 18 then
  print("Adult")
end
```

With `elseif` and `else`:

```duso
score = 85
if score >= 90 then
  print("A")
elseif score >= 80 then
  print("B")
else
  print("C or lower")
end
```

Checking truthiness:

```duso
items = []
if items then
  print("Has items")
else
  print("Empty")  // This prints (empty array is falsy)
end
```

## Ternary Operator

For quick inline conditionals, use the ternary operator:

```duso
status = age >= 18 ? "adult" : "minor"
```

The syntax is `condition ? true_value : false_value`.

## See Also

- [Comparison operators](/docs/reference/index.md)
- [Logical operators](/docs/reference/index.md)
