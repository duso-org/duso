# tan()

Calculate the tangent of an angle in radians.

## Signature

```duso
tan(angle)
```

## Parameters

- `angle` (number) - Angle in radians

## Returns

Tangent of the angle as a number

## Examples

Common angles:

```duso
print(tan(0))               // 0
print(tan(pi() / 4))        // 1 (45 degrees)
print(tan(pi() / 3))        // ~1.732 (60 degrees)
```

Converting degrees to radians:

```duso
degrees = 45
radians = degrees * pi() / 180
print(tan(radians))         // 1
```

Slope calculations:

```duso
// Calculate slope from angle
angle = atan(1.5)
slope = tan(angle)
print("Slope: {{slope}}")   // 1.5
```

## Notes

Tangent has vertical asymptotes (undefined values) at odd multiples of π/2. Avoid those angles.

## See Also

- [sin() - Sine](/docs/reference/sin.md)
- [cos() - Cosine](/docs/reference/cos.md)
- [atan() - Inverse tangent](/docs/reference/atan.md)
- [pi() - Mathematical constant π](/docs/reference/pi.md)
