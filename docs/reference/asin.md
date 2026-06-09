# asin()

Calculate the inverse sine (arcsine) of a value in radians.


`asin(x)`

```

## Parameters

- `x` (number) - Value between -1 and 1

## Returns

Angle in radians between -π/2 and π/2

## Examples

Basic inverse sine:

```duso
print(asin(0))
print(asin(0.5))
print(asin(1))

/*
  output:
  0
  0.5235987755982989 (30 degrees)
  1.5707963267948966 (90 degrees, π/2)
*/
```

Converting to degrees:

```duso
r = asin(0.5)
d = r * 180 / pi()
print("{{r}} radians = {{d}} degrees")

// output: 0.524 radians = 30 degrees
```

Triangle calculations:

```duso
// Find angle given opposite side and hypotenuse
o = 3
h = 5
angle = asin(o / h)
print("Angle: {{angle}} radians")

// output: Angle: 0.644 radians
```

## Notes

Input must be between -1 and 1. Values outside this range will produce invalid results (NaN).

## See Also

- [sin() - Sine](/docs/reference/sin.md)
- [acos() - Inverse cosine](/docs/reference/acos.md)
- [atan() - Inverse tangent](/docs/reference/atan.md)
- [pi() - Mathematical constant π](/docs/reference/pi.md)
