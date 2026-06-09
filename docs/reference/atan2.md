# atan2()

Calculate the angle (in radians) of a point (x, y) relative to the origin. Handles all four quadrants correctly.


`atan2(y, x)`

```

## Parameters

- `y` (number) - Y-coordinate (vertical)
- `x` (number) - X-coordinate (horizontal)

## Returns

Angle in radians between -π and π

## Notes

Unlike `atan()`, `atan2()` takes two arguments and handles all quadrants correctly. This is essential for calculating angles in 2D space.

## Examples

Basic usage:

```duso
print(atan2(0, 1))
print(atan2(1, 0))
print(atan2(0, -1))
print(atan2(-1, 0))

/*
  output:
  0
  1.571 (90 degrees, up)
  3.142 (180 degrees, left)
  -1.571 (-90 degrees, down)
*/
```

Converting to degrees:

```duso
ar = atan2(3, 4)
ad = ar * 180 / pi()
print("{{ad}} degrees")

// output: 36.87 degrees
```

Calculating direction angles:

```duso
// Find angle from origin to point
x = 10
y = 5
a = atan2(y, x)
print("Direction: {{a}} radians")

// Convert to compass bearing (0-360 degrees)
b = ((a * 180 / pi()) + 360) % 360
print("Bearing: {{b}} degrees")

/*
  output:
  Direction: 0.464 radians
  Bearing: 26 degrees
*/
```

Circular motion:

```duso
// Calculate angle between two points
x1 = 0
y1 = 0
x2 = 3
y2 = 4
a = atan2(y2 - y1, x2 - x1)
print("Angle: {{a}} radians")

// output: Angle: 0.927 radians
```

## Comparison with atan()

`atan()` only returns angles between -π/2 and π/2. For example:

```duso
print(atan(1))
print(atan2(1, 1))
print(atan2(-1, -1))

/*
  output:
  0.785 (ambiguous, could be 45° or 225°)
  0.785 (clearly 45° in first quadrant)
  -2.356 (clearly 225° in third quadrant)
*/
```

## See Also

- [atan() - One-argument arctangent](/docs/reference/atan.md)
- [sin() - Sine](/docs/reference/sin.md)
- [cos() - Cosine](/docs/reference/cos.md)
- [pi() - Mathematical constant π](/docs/reference/pi.md)
