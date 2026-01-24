# tobool()

Convert a value to a boolean.

## Signature

```duso
tobool(value)
```

## Parameters

- `value` - Any Duso value to convert

## Returns

Boolean: true or false

## Examples

Falsy values become false:

```duso
print(tobool(0))                // false
print(tobool(""))               // false
print(tobool(nil))              // false
print(tobool([]))               // false
print(tobool({}))               // false
```

Truthy values become true:

```duso
print(tobool(1))                // true
print(tobool(-1))               // true
print(tobool("text"))           // true
print(tobool("0"))              // true
print(tobool([1]))              // true
print(tobool({a=1}))            // true
```

## See Also

- [type() - Get value type](./type.md)
- [tonumber() - Convert to number](./tonumber.md)
- [tostring() - Convert to string](./tostring.md)
- [Type coercion](../language-spec.md#type-coercion)
