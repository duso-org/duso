# len()

Get the length of arrays, objects, or strings.

## Signature

```duso
len(value)
```

## Parameters

- `value` - An array, object, string, or binary

## Returns

Number: array length, object key count, string character count, or binary size in bytes

## Examples

Array length:

```duso
arr = [1, 2, 3, 4, 5]
print(len(arr))                 // 5
```

Object key count:

```duso
config = {timeout = 30, retries = 3, debug = false}
print(len(config))              // 3
```

String length:

```duso
text = "hello"
print(len(text))                // 5
```

Binary size:

```duso
data = load_binary("image.png")
print(len(data))                // byte count
```

Loop with length:

```duso
items = ["a", "b", "c"]
for i = 0, len(items) - 1 do
  print(items[i])
end
```

## See Also

- [binary - Binary data](/docs/reference/binary.md)
- [push() - Add element to array](/docs/reference/push.md)
