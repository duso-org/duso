# type()

Get the type name of a value for debugging and type checking.

## Signature

```duso
type(value)
```

## Parameters

- `value` - Any Duso value

## Returns

String containing the type name: `"nil"`, `"number"`, `"string"`, `"boolean"`, `"array"`, `"object"`, or `"function"`

## Examples

Basic type checking:

```duso
print(type(42))                 // "number"
print(type("hello"))            // "string"
print(type(true))               // "boolean"
print(type(nil))                // "nil"
```

Complex types:

```duso
print(type([1, 2, 3]))          // "array"
print(type({a = 1}))            // "object"
print(type(function() end))     // "function"
```

Conditional logic based on type:

```duso
value = get_something()
if type(value) == "number" then
  print("It's a number: " + value)
elseif type(value) == "string" then
  print("It's a string: " + value)
end
```

## See Also

- [tonumber() - Convert to number](/docs/reference/tonumber.md)
- [tostring() - Convert to string](/docs/reference/tostring.md)
