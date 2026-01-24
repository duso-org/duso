# tostring()

Convert a value to a string.

## Signature

```duso
tostring(value)
```

## Parameters

- `value` - Any Duso value to convert

## Returns

String representation of the value

## Examples

Convert numbers:

```duso
print(tostring(42))             // "42"
print(tostring(3.14))           // "3.14"
```

Convert booleans:

```duso
print(tostring(true))           // "true"
print(tostring(false))          // "false"
```

Convert arrays and objects:

```duso
arr = [1, 2, 3]
print(tostring(arr))            // "[1 2 3]"

obj = {x = 10, y = 20}
print(tostring(obj))            // "{x=10 y=20}"
```

## See Also

- [type() - Get value type](./type.md)
- [tonumber() - Convert to number](./tonumber.md)
- [tobool() - Convert to boolean](./tobool.md)
