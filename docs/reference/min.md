# min()

Find the minimum value among multiple numbers.

## Signature

```duso
min(...numbers)
```

## Parameters

- `...numbers` - One or more numbers to compare

## Returns

The smallest number

## Examples

Multiple numbers:

```duso
print(min(5, 2, 8, 1))          // 1
print(min(100, 50, 75))         // 50
```

Two numbers:

```duso
a = 10
b = 3
print(min(a, b))                // 3
```

Finding bounds:

```duso
temps = [72, 68, 75, 70, 69]
lowest = min(temps[0], temps[1], temps[2], temps[3], temps[4])
print(lowest)                   // 68
```

## See Also

- [max() - Find maximum value](/docs/reference/max.md)
- [clamp() - Clamp value to range](/docs/reference/clamp.md)
