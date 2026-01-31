# clamp()

Constrain a number between a minimum and maximum value.

## Signature

```duso
clamp(value, min, max)
```

## Parameters

- `value` (number) - The number to clamp
- `min` (number) - Minimum allowed value
- `max` (number) - Maximum allowed value

## Returns

Value clamped to the range [min, max]

## Examples

Clamp within range:

```duso
print(clamp(15, 10, 20))        // 15
print(clamp(5, 10, 20))         // 10
print(clamp(25, 10, 20))        // 20
```

Volume control:

```duso
volume = 150
clamped = clamp(volume, 0, 100)
print(clamped)                  // 100
```

Brightness adjustment:

```duso
brightness = -10
brightness = clamp(brightness, 0, 255)
print(brightness)               // 0
```

## See Also

- [min() - Find minimum](/docs/reference/min.md)
- [max() - Find maximum](/docs/reference/max.md)
