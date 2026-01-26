# range()

Create an array of numbers in a sequence.

## Signature

```duso
range(start, end [, step])
```

## Parameters

- `start` (number) - Starting value (inclusive)
- `end` (number) - Ending value (exclusive)
- `step` (optional, number) - Increment between values. Defaults to 1

## Returns

Array of numbers from start to end

## Examples

Basic range:

```duso
nums = range(1, 5)
print(nums)                     // [1 2 3 4]
```

With step:

```duso
evens = range(0, 10, 2)
print(evens)                    // [0 2 4 6 8]
```

Descending:

```duso
countdown = range(5, 0, -1)
print(countdown)                // [5 4 3 2 1]
```

Use in loop:

```duso
for i in range(1, 4) do
  print(i)
end
// Prints: 1, 2, 3
```
