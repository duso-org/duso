# continue

Skip to the next iteration of a loop.

## Syntax

```duso
continue
```

## Description

The `continue` statement skips the rest of the current loop iteration and jumps to the next one. Works in both `for` and `while` loops.

## Examples

Skip even numbers:

```duso
for i = 1, 10 do
  if i % 2 == 0 then
    continue
  end
  print(i)
end
// Output: 1 3 5 7 9
```

Skip empty items:

```duso
items = ["apple", "", "banana", "", "cherry"]
for item in items do
  if item == "" then
    continue
  end
  print(item)
end
// Output: apple banana cherry
```

With while loop:

```duso
count = 0
while count < 10 do
  count = count + 1
  if count == 3 then
    continue
  end
  print(count)
end
// Output: 1 2 4 5 6 7 8 9 10
```

## See Also

- [for](./for.md) - Count-based loop or iterate over collections
- [while](./while.md) - Condition-based loop
- [break](./break.md) - Exit a loop
