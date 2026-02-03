# unshift()

Add one or more elements to the beginning of an array. Mutates the array in place and returns the new length.

## Signature

```duso
unshift(array, value1 [, value2, ...])
```

## Parameters

- `array` (array) - The array to modify
- `value1, value2, ...` - One or more values to add to the beginning

## Returns

The new length of the array (as a number)

## Examples

Unshift a single element:

```duso
arr = [2, 3]
len = unshift(arr, 1)
print(arr)                      // [1 2 3]
print(len)                      // 3
```

Unshift multiple elements:

```duso
items = ["c", "d"]
unshift(items, "a", "b")
print(items)                    // [a b c d]
```

Build array from back to front:

```duso
result = []
for i = 5, 1, -1 do
  unshift(result, i)
end
print(result)                   // [1 2 3 4 5]
```

## See Also

- [shift() - Remove first element](/docs/reference/shift.md)
- [push() - Add to end](/docs/reference/push.md)
- [pop() - Remove last element](/docs/reference/pop.md)
