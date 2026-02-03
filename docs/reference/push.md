# push()

Add one or more elements to the end of an array. Mutates the array in place and returns the new length.

## Signature

```duso
push(array, value1 [, value2, ...])
```

## Parameters

- `array` (array) - The array to modify
- `value1, value2, ...` - One or more values to add to the end

## Returns

The new length of the array (as a number)

## Examples

Push a single element:

```duso
arr = [1, 2, 3]
len = push(arr, 4)
print(arr)                      // [1 2 3 4]
print(len)                      // 4
```

Push multiple elements:

```duso
items = ["apple", "banana"]
push(items, "orange", "grape", "mango")
print(items)                    // [apple banana orange grape mango]
```

Use in a loop:

```duso
result = []
for i = 1, 5 do
  push(result, i * i)
end
print(result)                   // [1 4 9 16 25]
```

## See Also

- [pop() - Remove and return last element](/docs/reference/pop.md)
- [unshift() - Add to beginning](/docs/reference/unshift.md)
- [append() - Create new array with element](/docs/reference/append.md)
