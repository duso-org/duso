# asin()

Calculate the inverse sine (arcsine) of a value in radians.

## Signature

```duso
asin(x)
```

## Parameters

- `x` (number) - Value between -1 and 1

## Returns

Angle in radians between -π/2 and π/2

## Examples

Basic inverse sine:

```duso
print(asin(0))              // 0
print(asin(0.5))            // ~0.5236 (30 degrees)
print(asin(1))              // ~1.5708 (90 degrees, π/2)
```

Converting to degrees:

```duso
radians = asin(0.5)
degrees = radians * 180 / pi()
print("{{radians}} radians = {{degrees}} degrees")  // 30 degrees
```

Triangle calculations:

```duso
// Find angle given opposite side and hypotenuse
opposite = 3
hypotenuse = 5
angle = asin(opposite / hypotenuse)
print("Angle: {{angle}} radians")
```

## Notes

Input must be between -1 and 1. Values outside this range will produce invalid results (NaN).

## See Also

- [sin() - Sine](/docs/reference/sin.md)
- [acos() - Inverse cosine](/docs/reference/acos.md)
- [atan() - Inverse tangent](/docs/reference/atan.md)
- [pi() - Mathematical constant π](/docs/reference/pi.md)
