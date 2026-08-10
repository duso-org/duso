# range()

Create an array of numbers in a sequence.

`range(start, end [, step])`

## Parameters

- `start` (number) - Starting value (inclusive)
- `end` (number) - Ending value (inclusive)
- `step` (optional, number) - Increment between values. Defaults to 1

Both arguments are required; `range(5)` is an error.

## Returns

Array of numbers from start to end

## Examples

Basic range:

```duso
nums = range(1, 5)
print(nums)                     // [1 2 3 4 5]
```

With step:

```duso
evens = range(0, 10, 2)
print(evens)                    // [0 2 4 6 8 10]
```

Descending:

```duso
countdown = range(5, 0, -1)
print(countdown)                // [5 4 3 2 1 0]
```

Use in loop:

```duso
for i in range(1, 4) do
  print(i)
end
// Prints: 1, 2, 3, 4
```

## Notes

The end value is included when the step lands on it exactly. A step that
overshoots simply stops short:

```duso
print(range(1, 10, 2))          // [1 3 5 7 9]  -- 11 would pass 10
print(range(1, 10, 3))          // [1 4 7 10]   -- lands on 10
```

A range of one value returns one element, and a descending range without a
negative step returns empty:

```duso
print(range(1, 1))              // [1]
print(range(5, 1))              // []
```
