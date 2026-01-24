# sqrt()

Calculate the square root of a number.

## Signature

```duso
sqrt(number)
```

## Parameters

- `number` (number) - The number to get the square root of

## Returns

Square root as a number

## Examples

Perfect squares:

```duso
print(sqrt(16))                 // 4
print(sqrt(9))                  // 3
print(sqrt(1))                  // 1
```

Decimals:

```duso
print(sqrt(2))                  // 1.414...
print(sqrt(10))                 // 3.162...
```

Distance formula:

```duso
x = 3
y = 4
distance = sqrt(x * x + y * y)
print(distance)                 // 5
```

## See Also

- [pow() - Raise to power](./pow.md)
- [Math functions](../language-spec.md#math-functions)
