# abs()

Get the absolute value of a number.

## Signature

```duso
abs(number)
```

## Parameters

- `number` (number) - The number to convert

## Returns

Absolute value (always non-negative)

## Examples

Positive and negative numbers:

```duso
print(abs(42))                  // 42
print(abs(-42))                 // 42
print(abs(0))                   // 0
```

Decimals:

```duso
print(abs(3.14))                // 3.14
print(abs(-3.14))               // 3.14
```

Distance calculation:

```duso
a = 5
b = -12
distance = abs(a - b)
print(distance)                 // 17
```

## See Also

- [floor() - Round down](/docs/reference/floor.md)
