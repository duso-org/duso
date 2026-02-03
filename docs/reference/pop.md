# pop()

Remove and return the last element of an array. Mutates the array in place.

## Signature

```duso
pop(array)
```

## Parameters

- `array` (array) - The array to modify

## Returns

The last element of the array, or `nil` if the array is empty

## Examples

Pop a single element:

```duso
arr = [1, 2, 3]
last = pop(arr)
print(last)                     // 3
print(arr)                      // [1 2]
```

Pop from array until empty:

```duso
items = ["a", "b", "c"]
while len(items) > 0 do
  item = pop(items)
  print("Popped:", item)
end
print(items)                    // []
```

Handle empty array:

```duso
empty = []
result = pop(empty)
print(result)                   // nil
```

## See Also

- [push() - Add to end](/docs/reference/push.md)
- [shift() - Remove first element](/docs/reference/shift.md)
- [unshift() - Add to beginning](/docs/reference/unshift.md)
