# for

Loop over a range of numbers or iterate over array/object elements.

## Syntax

```duso
// Numeric loop
for i = start, end do
  // statements
end

for i = start, end, step do
  // statements
end

// Iterator loop
for item in collection do
  // statements
end
```

## Numeric Loop

Counts from `start` to `end` (inclusive). Optional `step` defaults to 1:

```duso
for i = 1, 5 do
  print(i)  // 1 2 3 4 5
end

for i = 1, 10, 2 do
  print(i)  // 1 3 5 7 9
end

for i = 10, 1, -1 do
  print(i)  // 10 9 8 7 6 5 4 3 2 1
end
```

The loop variable `i` is local to the loop.

## Iterator Loop

Iterates over array elements or object keys:

```duso
items = ["apple", "banana", "cherry"]
for item in items do
  print(item)  // apple, banana, cherry
end

config = {host = "localhost", port = 8080}
for key in config do
  print(key)  // host, port
end
```

## Break and Continue

Exit a loop early with [`break`](/docs/reference/break.md) or skip to the next iteration with [`continue`](/docs/reference/continue.md):

```duso
for i = 1, 10 do
  if i == 2 then continue end
  if i == 8 then break end
  print(i)  // 1 3 4 5 6 7
end
```

## See Also

- [while](/docs/reference/while.md) - Condition-based loop
- [break](/docs/reference/break.md) - Exit a loop
- [continue](/docs/reference/continue.md) - Skip to next iteration
