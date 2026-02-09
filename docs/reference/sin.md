# sin()

Calculate the sine of an angle in radians.

## Signature

```duso
sin(angle)
```

## Parameters

- `angle` (number) - Angle in radians

## Returns

Sine of the angle as a number between -1 and 1

## Examples

Common angles:

```duso
print(sin(0))               // 0
print(sin(pi() / 2))        // 1 (90 degrees)
print(sin(pi()))            // 0 (180 degrees)
```

Converting degrees to radians:

```duso
degrees = 30
radians = degrees * pi() / 180
print(sin(radians))         // 0.5
```

Oscillating values:

```duso
for i in range(0, 4) do
  angle = i * pi() / 2
  print("sin({{angle}}) = {{sin(angle)}}")
end
```

## See Also

- [cos() - Cosine](/docs/reference/cos.md)
- [tan() - Tangent](/docs/reference/tan.md)
- [asin() - Inverse sine](/docs/reference/asin.md)
- [pi() - Mathematical constant π](/docs/reference/pi.md)
