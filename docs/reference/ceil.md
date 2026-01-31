# ceil()

Round a number up to the nearest integer.

## Signature

```duso
ceil(number)
```

## Parameters

- `number` (number) - The number to round

## Returns

Smallest integer greater than or equal to the number

## Examples

Round up decimals:

```duso
print(ceil(3.1))                // 4
print(ceil(3.7))                // 4
print(ceil(3.0))                // 3
```

Negative numbers:

```duso
print(ceil(-2.3))               // -2
print(ceil(-2.9))               // -2
```

Practical use:

```duso
pages_needed = ceil(items / items_per_page)
```

## See Also

- [floor() - Round down](/docs/reference/floor.md)
- [round() - Round to nearest](/docs/reference/round.md)
