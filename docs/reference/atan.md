# atan()

Calculate the inverse tangent (arctangent) of a value in radians.

## Signature

```duso
atan(x)
```

## Parameters

- `x` (number) - Any real number

## Returns

Angle in radians between -π/2 and π/2

## Examples

Basic inverse tangent:

```duso
print(atan(0))              // 0
print(atan(1))              // ~0.7854 (45 degrees)
print(atan(-1))             // ~-0.7854 (-45 degrees)
```

Converting to degrees:

```duso
radians = atan(1)
degrees = radians * 180 / pi()
print("{{radians}} radians = {{degrees}} degrees")  // 45 degrees
```

Finding slope angles:

```duso
// Convert slope to angle
slope = 2
angle = atan(slope)
print("Angle: {{angle}} radians")

// Convert back
recovered_slope = tan(angle)
print("Slope: {{recovered_slope}}")
```

## Notes

For 2D coordinates (x, y), use [`atan2()`](/docs/reference/atan2.md) instead to get the correct quadrant.

## See Also

- [tan() - Tangent](/docs/reference/tan.md)
- [atan2() - Two-argument arctangent](/docs/reference/atan2.md)
- [asin() - Inverse sine](/docs/reference/asin.md)
- [acos() - Inverse cosine](/docs/reference/acos.md)
- [pi() - Mathematical constant π](/docs/reference/pi.md)
