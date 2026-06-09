# acos()

Calculate the inverse cosine (arccosine) of a value in radians.


`acos(x)`

```

## Parameters

- `x` (number) - Value between -1 and 1

## Returns

Angle in radians between 0 and π

## Examples

Basic inverse cosine:

```duso
print(acos(1))
print(acos(0.5))
print(acos(0))
print(acos(-1))

/*
  output:
  0
  ~1.047 (60 degrees)
  ~1.5708 (90 degrees, π/2)
  ~3.14159 (180 degrees, π)
*/
```

Converting to degrees:

```duso
r = acos(0.5)
d = r * 180 / pi()
print("{{r}} radians = {{d}} degrees")

/*
  output:
  1.0471975511966 radians = 60 degrees
*/
```

Triangle calculations:

```duso
// Find angle given adjacent side and hypotenuse
a = 3
h = 5
angle = acos(a / h)
print("Angle: {{angle}} radians")

/*
  output:
  Angle: 0.927295218 radians
*/
```

## Notes

Input must be between -1 and 1. Values outside this range will produce invalid results (NaN).

## See Also

- [cos() - Cosine](/docs/reference/cos.md)
- [asin() - Inverse sine](/docs/reference/asin.md)
- [atan() - Inverse tangent](/docs/reference/atan.md)
- [pi() - Mathematical constant π](/docs/reference/pi.md)
