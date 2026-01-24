# tonumber()

Convert a value to a number.

## Signature

```duso
tonumber(value)
```

## Parameters

- `value` - Any Duso value to convert

## Returns

Number (float64)

## Examples

Convert strings:

```duso
print(tonumber("42"))           // 42
print(tonumber("3.14"))         // 3.14
print(tonumber("-5"))           // -5
```

Convert booleans:

```duso
print(tonumber(true))           // 1
print(tonumber(false))          // 0
```

Processing user input:

```duso
user_age = input("Enter your age: ")
age = tonumber(user_age)
if age >= 18 then
  print("You are an adult")
end
```

## See Also

- [type() - Get value type](./type.md)
- [tostring() - Convert to string](./tostring.md)
- [tobool() - Convert to boolean](./tobool.md)
