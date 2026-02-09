# exp()

Calculate e raised to a power (exponential growth).

## Signature

```duso
exp(x)
```

## Parameters

- `x` (number) - The exponent

## Returns

e^x as a number

## Examples

Basic usage:

```duso
print(exp(0))               // 1
print(exp(1))               // ~2.71828 (e)
print(exp(2))               // ~7.389
print(exp(-1))              // ~0.3679
```

Exponential growth:

```duso
// Calculate compound interest
principal = 1000
rate = 0.05  // 5% annual
time = 3
amount = principal * exp(rate * time)
print("Amount: {{amount}}")
```

Logistics and decay:

```duso
// Population growth model
t = 5
population = 100 * exp(0.1 * t)
print("Population: {{population}}")

// Radioactive decay
half_life = 5730
time_elapsed = 1000
remaining = 1 * exp(-(ln(2) / half_life) * time_elapsed)
print("Remaining: {{remaining}}")
```

## See Also

- [ln() - Natural logarithm](/docs/reference/ln.md)
- [log() - Base 10 logarithm](/docs/reference/log.md)
- [pow() - Power function](/docs/reference/pow.md)
