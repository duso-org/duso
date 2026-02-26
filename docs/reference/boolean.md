# Boolean

Booleans represent truth values: `true` or `false`. Used in conditions and logical operations.

## Creating Booleans

```duso
flag = true
active = false
```

## Logical Operators

Combine booleans with logical operators:

```duso
a = true
b = false

print(a and b)    // false
print(a or b)     // true
print(not a)      // false
print(not b)      // true
```

- `and` - Both conditions must be true (short-circuit evaluation)
- `or` - At least one condition must be true (short-circuit evaluation)
- `not` - Negate a condition

## Comparison

Conditions evaluate to boolean:

```duso
result = 5 < 10   // true
result = 5 > 10   // false
result = 5 == 5   // true
result = 5 != 3   // true
```

## Truthiness

In conditions, values are truthy or falsy:

```duso
if true then print("true") end      // prints
if false then print("true") end     // doesn't print
if 1 then print("true") end         // prints (1 is truthy)
if 0 then print("true") end         // doesn't print (0 is falsy)
if "" then print("true") end        // doesn't print (empty string is falsy)
```

**Falsy values:** `false`, `nil`, `0`, `""` (empty string), `[]` (empty array), `{}` (empty object)

**Truthy values:** Everything else, including `true`, non-zero numbers, non-empty strings, non-empty arrays, non-empty objects

## Type Conversion

Convert to boolean with [`tobool()`](/docs/reference/tobool.md):

```duso
b = tobool(1)       // true
b = tobool(0)       // false
b = tobool("text")  // true
b = tobool("")      // false
```

