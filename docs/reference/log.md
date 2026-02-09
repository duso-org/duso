# log()

Calculate the logarithm base 10 of a number.

## Signature

```duso
log(x)
```

## Parameters

- `x` (number) - A positive number

## Returns

Base 10 logarithm of x

## Examples

Basic usage:

```duso
print(log(1))               // 0
print(log(10))              // 1
print(log(100))             // 2
print(log(1000))            // 3
```

Finding magnitude:

```duso
// How many digits in a number?
value = 50000
magnitude = floor(log(value)) + 1
print("Digits: {{magnitude}}")
```

Audio decibels:

```duso
// Calculate decibels (logarithmic scale)
intensity_ratio = 100
decibels = 10 * log(intensity_ratio)
print("Decibels: {{decibels}}")
```

## Notes

Input must be positive. log(0) returns negative infinity, and log(negative) returns NaN.

For natural logarithm, use [`ln()`](/docs/reference/ln.md).

## See Also

- [ln() - Natural logarithm](/docs/reference/ln.md)
- [exp() - Exponential function](/docs/reference/exp.md)
- [pow() - Power function](/docs/reference/pow.md)
