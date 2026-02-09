# atan2()

Calculate the angle (in radians) of a point (x, y) relative to the origin. Handles all four quadrants correctly.

## Signature

```duso
atan2(y, x)
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
print(atan2(0, 1))          // 0 (0 degrees, right)
print(atan2(1, 0))          // ~1.5708 (90 degrees, up)
print(atan2(0, -1))         // ~3.14159 (180 degrees, left)
print(atan2(-1, 0))         // ~-1.5708 (-90 degrees, down)
```

Converting to degrees:

```duso
angle_rad = atan2(3, 4)
angle_deg = angle_rad * 180 / pi()
print("{{angle_deg}} degrees")
```

Calculating direction angles:

```duso
// Find angle from origin to point
x = 10
y = 5
angle = atan2(y, x)
print("Direction: {{angle}} radians")

// Convert to compass bearing (0-360 degrees)
bearing = ((angle * 180 / pi()) + 360) % 360
print("Bearing: {{bearing}} degrees")
```

Circular motion:

```duso
// Calculate angle between two points
x1, y1 = 0, 0
x2, y2 = 3, 4
angle = atan2(y2 - y1, x2 - x1)
print("Angle: {{angle}} radians")
```

## Comparison with atan()

`atan()` only returns angles between -π/2 and π/2. For example:

```duso
print(atan(1))              // ~0.7854 (ambiguous, could be 45° or 225°)
print(atan2(1, 1))          // ~0.7854 (clearly 45° in first quadrant)
print(atan2(-1, -1))        // ~-2.356 (clearly 225° in third quadrant)
```

## See Also

- [atan() - One-argument arctangent](/docs/reference/atan.md)
- [sin() - Sine](/docs/reference/sin.md)
- [cos() - Cosine](/docs/reference/cos.md)
- [pi() - Mathematical constant π](/docs/reference/pi.md)
