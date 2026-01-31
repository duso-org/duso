# floor()

Round a number down to the nearest integer.

## Signature

```duso
floor(number)
```

## Parameters

- `number` (number) - The number to round

## Returns

Largest integer less than or equal to the number

## Examples

Round down decimals:

```duso
print(floor(3.7))               // 3
print(floor(3.2))               // 3
print(floor(3.0))               // 3
```

Negative numbers:

```duso
print(floor(-2.3))              // -3
print(floor(-2.9))              // -3
```

Practical use:

```duso
hours = 3.75
whole_hours = floor(hours)
print(whole_hours)              // 3
```

## See Also

- [ceil() - Round up](/docs/reference/ceil.md)
- [round() - Round to nearest](/docs/reference/round.md)
