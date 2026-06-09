# atan()

Calculate the inverse tangent (arctangent) of a value in radians.


`atan(x)`

```

## Parameters

- `x` (number) - Any real number

## Returns

Angle in radians between -π/2 and π/2

## Examples

Basic inverse tangent:

```duso
print(atan(0))
print(atan(1))
print(atan(-1))

/*
  output:
  0
  0.785 (45 degrees)
  -0.785 (-45 degrees)
*/
```

Converting to degrees:

```duso
r = atan(1)
d = r * 180 / pi()
print("{{r}} radians = {{d}} degrees")

// output: 0.785 radians = 45 degrees
```

Finding slope angles:

```duso
// Convert slope to angle
s = 2
a = atan(s)
print("Angle: {{a}} radians")

// Convert back
rs = tan(a)
print("Slope: {{rs}}")

/*
  output:
  Angle: 1.107 radians
  Slope: 2
*/
```

## Notes

For 2D coordinates (x, y), use [`atan2()`](/docs/reference/atan2.md) instead to get the correct quadrant.

## See Also

- [tan() - Tangent](/docs/reference/tan.md)
- [atan2() - Two-argument arctangent](/docs/reference/atan2.md)
- [asin() - Inverse sine](/docs/reference/asin.md)
- [acos() - Inverse cosine](/docs/reference/acos.md)
- [pi() - Mathematical constant π](/docs/reference/pi.md)
