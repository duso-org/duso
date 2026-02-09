# cos()

Calculate the cosine of an angle in radians.

## Signature

```duso
cos(angle)
```

## Parameters

- `angle` (number) - Angle in radians

## Returns

Cosine of the angle as a number between -1 and 1

## Examples

Common angles:

```duso
print(cos(0))               // 1
print(cos(pi() / 2))        // 0 (90 degrees)
print(cos(pi()))            // -1 (180 degrees)
```

Converting degrees to radians:

```duso
degrees = 60
radians = degrees * pi() / 180
print(cos(radians))         // 0.5
```

Circular motion:

```duso
// Calculate x-coordinate on unit circle
for i in range(0, 8) do
  angle = i * pi() / 4
  x = cos(angle)
  y = sin(angle)
  print("Angle: {{angle}}, x: {{x}}, y: {{y}}")
end
```

## See Also

- [sin() - Sine](/docs/reference/sin.md)
- [tan() - Tangent](/docs/reference/tan.md)
- [acos() - Inverse cosine](/docs/reference/acos.md)
- [pi() - Mathematical constant π](/docs/reference/pi.md)
