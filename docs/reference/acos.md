# acos()

Calculate the inverse cosine (arccosine) of a value in radians.

## Signature

```duso
acos(x)
```

## Parameters

- `x` (number) - Value between -1 and 1

## Returns

Angle in radians between 0 and π

## Examples

Basic inverse cosine:

```duso
print(acos(1))              // 0
print(acos(0.5))            // ~1.047 (60 degrees)
print(acos(0))              // ~1.5708 (90 degrees, π/2)
print(acos(-1))             // ~3.14159 (180 degrees, π)
```

Converting to degrees:

```duso
radians = acos(0.5)
degrees = radians * 180 / pi()
print("{{radians}} radians = {{degrees}} degrees")  // 60 degrees
```

Triangle calculations:

```duso
// Find angle given adjacent side and hypotenuse
adjacent = 3
hypotenuse = 5
angle = acos(adjacent / hypotenuse)
print("Angle: {{angle}} radians")
```

## Notes

Input must be between -1 and 1. Values outside this range will produce invalid results (NaN).

## See Also

- [cos() - Cosine](/docs/reference/cos.md)
- [asin() - Inverse sine](/docs/reference/asin.md)
- [atan() - Inverse tangent](/docs/reference/atan.md)
- [pi() - Mathematical constant π](/docs/reference/pi.md)
