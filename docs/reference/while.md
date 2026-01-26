# while

Loop while a condition is true.

## Syntax

```duso
while condition do
  // statements
end
```

## Description

The `while` loop executes repeatedly as long as the condition is truthy. The condition is checked at the start of each iteration.

## Examples

Basic loop:

```duso
count = 0
while count < 5 do
  print(count)
  count = count + 1
end
// Output: 0 1 2 3 4
```

With complex condition:

```duso
name = ""
while len(name) == 0 do
  name = load("name.txt")
  if name == "" then
    print("Please provide a name")
  end
end
```

## Break and Continue

Exit early with [`break`](./break.md) or skip to the next iteration with [`continue`](./continue.md):

```duso
count = 0
while true do
  count = count + 1
  if count == 3 then continue end
  if count == 7 then break end
  print(count)  // 1 2 4 5 6
end
```

## See Also

- [for](./for.md) - Count-based loop or iterate over collections
- [break](./break.md) - Exit a loop
- [continue](./continue.md) - Skip to next iteration
