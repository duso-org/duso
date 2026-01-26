# max()

Find the maximum value among multiple numbers.

## Signature

```duso
max(...numbers)
```

## Parameters

- `...numbers` - One or more numbers to compare

## Returns

The largest number

## Examples

Multiple numbers:

```duso
print(max(5, 2, 8, 1))          // 8
print(max(100, 50, 75))         // 100
```

Two numbers:

```duso
a = 10
b = 3
print(max(a, b))                // 10
```

Finding bounds:

```duso
scores = [85, 92, 78, 95, 88]
highest = max(scores[0], scores[1], scores[2], scores[3], scores[4])
print(highest)                  // 95
```

## See Also

- [min() - Find minimum value](./min.md)
- [clamp() - Clamp value to range](./clamp.md)
