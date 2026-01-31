# append()

Add an element to an array and return the new array.

## Signature

```duso
append(array, value)
```

## Parameters

- `array` (array) - The array to add to
- `value` - The value to append

## Returns

New array with the element added at the end

## Examples

Append a number:

```duso
arr = [1, 2, 3]
arr = append(arr, 4)
print(arr)                      // [1 2 3 4]
```

Append different types:

```duso
items = ["apple", "banana"]
items = append(items, "orange")
items = append(items, 42)
items = append(items, true)
print(items)                    // [apple banana orange 42 true]
```

Building an array in a loop:

```duso
result = []
for i = 1, 5 do
  result = append(result, i * 2)
end
print(result)                   // [2 4 6 8 10]
```

## See Also

- [len() - Get array length](/docs/reference/len.md)
