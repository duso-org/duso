# ln()

Calculate the natural logarithm of a number (logarithm base e).

## Signature

```duso
ln(x)
```

## Parameters

- `x` (number) - A positive number

## Returns

Natural logarithm of x (base e)

## Examples

Basic usage:

```duso
print(ln(1))                // 0
print(ln(2.71828))          // ~1 (approximately e)
print(ln(10))               // ~2.3026
```

Inverse of exp:

```duso
x = 2.5
exponential = exp(x)
back_to_x = ln(exponential)
print("Original: {{x}}, After ln(exp(x)): {{back_to_x}}")
```

Half-life calculations:

```duso
// Find time for half of a radioactive substance to decay
decay_rate = 0.1
half_life = -ln(0.5) / decay_rate
print("Half-life: {{half_life}}")
```

Continuous compound interest:

```duso
principal = 1000
rate = 0.05
years = 3
amount = principal * exp(rate * years)
print("Amount: {{amount}}")

// Solve for time needed to reach target
target = 1500
time = ln(target / principal) / rate
print("Years to reach target: {{time}}")
```

## Notes

Input must be positive. ln(0) returns negative infinity, and ln(negative) returns NaN.

## See Also

- [exp() - Exponential function](/docs/reference/exp.md)
- [log() - Base 10 logarithm](/docs/reference/log.md)
- [pow() - Power function](/docs/reference/pow.md)
