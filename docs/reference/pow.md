# pow()

Raise a number to a power (exponentiation).

## Signature

```duso
pow(base, exponent)
```

## Parameters

- `base` (number) - The base number
- `exponent` (number) - The power to raise to

## Returns

base raised to the exponent

## Examples

Positive exponents:

```duso
print(pow(2, 3))                // 8
print(pow(5, 2))                // 25
print(pow(10, 3))               // 1000
```

Negative exponents:

```duso
print(pow(2, -1))               // 0.5
print(pow(10, -2))              // 0.01
```

Fractional exponents:

```duso
print(pow(4, 0.5))              // 2 (square root)
print(pow(8, 1/3))              // 2 (cube root)
```

## See Also

- [sqrt() - Square root](./sqrt.md)
- [Math functions](../language-spec.md#math-functions)
