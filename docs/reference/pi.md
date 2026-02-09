# pi()

Get the mathematical constant π (pi), approximately 3.14159...

## Signature

```duso
pi()
```

## Parameters

None

## Returns

The mathematical constant π (pi) as a number

## Examples

Basic usage:

```duso
print(pi())                 // 3.141592653589793
```

Circle calculations:

```duso
// Circle circumference
radius = 5
circumference = 2 * pi() * radius
print("Circumference: {{circumference}}")

// Circle area
area = pi() * radius * radius
print("Area: {{area}}")
```

Converting between radians and degrees:

```duso
// Degrees to radians
degrees = 45
radians = degrees * pi() / 180
print("45 degrees = {{radians}} radians")

// Radians to degrees
radians = pi() / 4
degrees = radians * 180 / pi()
print("π/4 radians = {{degrees}} degrees")
```

Trigonometric calculations:

```duso
// Full circle is 2π radians
angle = pi() / 3
print("sin(π/3) = {{sin(angle)}}")
print("cos(π/3) = {{cos(angle)}}")
```

Sphere calculations:

```duso
// Volume of a sphere
radius = 3
volume = (4/3) * pi() * pow(radius, 3)
print("Volume: {{volume}}")

// Surface area of a sphere
surface_area = 4 * pi() * radius * radius
print("Surface area: {{surface_area}}")
```

## See Also

- [sin() - Sine](/docs/reference/sin.md)
- [cos() - Cosine](/docs/reference/cos.md)
- [tan() - Tangent](/docs/reference/tan.md)
- [exp() - Exponential function](/docs/reference/exp.md)
- [pow() - Power function](/docs/reference/pow.md)
