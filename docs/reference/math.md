# Math Functions

Duso provides a comprehensive set of mathematical functions for basic arithmetic, trigonometry, and other calculations.

## Basic Operations

- [abs()](/docs/reference/abs.md) - Absolute value
- [ceil()](/docs/reference/ceil.md) - Round up to nearest integer
- [floor()](/docs/reference/floor.md) - Round down to nearest integer
- [round()](/docs/reference/round.md) - Round to nearest integer
- [sqrt()](/docs/reference/sqrt.md) - Square root
- [pow()](/docs/reference/pow.md) - Raise to power (exponentiation)
- [min()](/docs/reference/min.md) - Find minimum value
- [max()](/docs/reference/max.md) - Find maximum value
- [clamp()](/docs/reference/clamp.md) - Constrain value between min and max
- [random()](/docs/reference/random.md) - Get random float between 0 and 1

## Trigonometric Functions

All trigonometric functions work with angles in radians. Use `pi()` for π.

- [sin()](/docs/reference/sin.md) - Sine of angle in radians
- [cos()](/docs/reference/cos.md) - Cosine of angle in radians
- [tan()](/docs/reference/tan.md) - Tangent of angle in radians
- [asin()](/docs/reference/asin.md) - Inverse sine (arcsine), x between -1 and 1
- [acos()](/docs/reference/acos.md) - Inverse cosine (arccosine), x between -1 and 1
- [atan()](/docs/reference/atan.md) - Inverse tangent (arctangent)
- [atan2()](/docs/reference/atan2.md) - Inverse tangent with quadrant correction

## Exponential & Logarithmic

- [exp()](/docs/reference/exp.md) - e raised to the power x
- [log()](/docs/reference/log.md) - Logarithm base 10
- [ln()](/docs/reference/ln.md) - Natural logarithm (base e)
- [pi()](/docs/reference/pi.md) - Mathematical constant π (3.14159...)

## Quick Examples

### Basic Arithmetic

```duso
print(abs(-5))           // 5
print(sqrt(16))          // 4
print(pow(2, 8))         // 256
print(round(3.7))        // 4
print(clamp(5, 1, 3))    // 3
```

### Trigonometry

```duso
// Convert degrees to radians: degrees * pi() / 180
angle = 45 * pi() / 180
print(sin(angle))        // ~0.707
print(cos(angle))        // ~0.707

// Find angle from coordinates
y = 1
x = 1
angle = atan2(y, x)      // ~0.785 radians (45 degrees)
```

### Min/Max/Random

```duso
print(min(3, 7, 2))      // 2
print(max(3, 7, 2))      // 7
print(random())          // Random float 0-1
```
