# shift()

Remove and return the first element of an array. Mutates the array in place.

## Signature

```duso
shift(array)
```

## Parameters

- `array` (array) - The array to modify

## Returns

The first element of the array, or `nil` if the array is empty

## Examples

Shift a single element:

```duso
arr = [1, 2, 3]
first = shift(arr)
print(first)                    // 1
print(arr)                      // [2 3]
```

Process queue of items:

```duso
queue = ["task1", "task2", "task3"]
while len(queue) > 0 do
  task = shift(queue)
  print("Processing:", task)
end
print(queue)                    // []
```

Handle empty array:

```duso
empty = []
result = shift(empty)
print(result)                   // nil
```

## See Also

- [unshift() - Add to beginning](/docs/reference/unshift.md)
- [pop() - Remove last element](/docs/reference/pop.md)
- [push() - Add to end](/docs/reference/push.md)
