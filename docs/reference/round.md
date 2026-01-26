# round()

Round a number to the nearest integer.

## Signature

```duso
round(number)
```

## Parameters

- `number` (number) - The number to round

## Returns

Number rounded to nearest integer

## Examples

Round decimals:

```duso
print(round(3.2))               // 3
print(round(3.5))               // 4
print(round(3.7))               // 4
```

Negative numbers:

```duso
print(round(-2.3))              // -2
print(round(-2.5))              // -2 or -3 (banker's rounding)
print(round(-2.7))              // -3
```

Practical use:

```duso
rating = 4.6
stars = round(rating)
print(stars)                    // 5
```

## See Also

- [floor() - Round down](./floor.md)
- [ceil() - Round up](./ceil.md)
