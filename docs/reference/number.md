# Number

Numbers in Duso are floating-point values (64-bit floats). Used for arithmetic, counting, and any numeric operations.

## Creating Numbers

```duso
count = 42
price = 19.99
negative = -5
zero = 0
scientific = 1.23e4
```

## Arithmetic Operations

All standard arithmetic operators work with numbers:

```duso
a = 10
b = 3

print(a + b)   // 13 (addition)
print(a - b)   // 7 (subtraction)
print(a * b)   // 30 (multiplication)
print(a / b)   // 3.333... (division)
print(a % b)   // 1 (modulo)
```

## Comparison

Numbers can be compared:

```duso
print(5 < 10)     // true
print(5 == 5)     // true
print(5 != 3)     // true
print(5 >= 5)     // true
```

## Type Conversion

Convert other types to numbers with [`tonumber()`](tonumber.md):

```duso
num = tonumber("42")      // 42
num = tonumber("3.14")    // 3.14
num = tonumber(true)      // 1
```

## Math Functions

Duso provides math functions for common operations:

- [`floor()`](floor.md) - Round down
- [`ceil()`](ceil.md) - Round up
- [`round()`](round.md) - Round to nearest
- [`abs()`](abs.md) - Absolute value
- [`min()`](min.md) - Minimum of values
- [`max()`](max.md) - Maximum of values
- [`sqrt()`](sqrt.md) - Square root
- [`pow()`](pow.md) - Exponentiation
- [`clamp()`](clamp.md) - Constrain between min/max

## Boolean Context

In conditions, numbers are truthy except for `0`:

```duso
if 1 then print("true") end      // prints
if 0 then print("true") end      // doesn't print
```

