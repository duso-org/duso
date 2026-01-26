# break

Exit a loop immediately.

## Syntax

```duso
break
```

## Description

The `break` statement exits the enclosing `for` or `while` loop immediately, skipping any remaining iterations.

## Examples

Exit when a condition is met:

```duso
for i = 1, 100 do
  if i == 5 then
    break
  end
  print(i)
end
// Output: 1 2 3 4
```

Finding a value:

```duso
items = ["apple", "banana", "cherry", "date"]
found = false
for item in items do
  if item == "cherry" then
    found = true
    break
  end
end
print(found)  // true
```

## See Also

- [for](./for.md) - Count-based loop or iterate over collections
- [while](./while.md) - Condition-based loop
- [continue](./continue.md) - Skip to next iteration
