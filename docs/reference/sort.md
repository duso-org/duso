# sort()

Sort an array in ascending order, optionally with a custom comparison function.

## Signature

```duso
sort(array [, comparison_function])
```

## Parameters

- `array` (array) - The array to sort
- `comparison_function` (optional, function) - Custom comparison function that returns true if first argument comes before second

## Returns

New sorted array

## Examples

Sort numbers:

```duso
nums = [3, 1, 4, 1, 5, 9, 2, 6]
sorted = sort(nums)
print(sorted)                   // [1 1 2 3 4 5 6 9]
```

Sort strings:

```duso
words = ["banana", "apple", "cherry"]
sorted = sort(words)
print(sorted)                   // [apple banana cherry]
```

Descending order with custom function:

```duso
nums = [3, 1, 4, 1, 5]
function desc(a, b)
  return a > b
end
sorted = sort(nums, desc)
print(sorted)                   // [5 4 3 1 1]
```

## See Also

- [map() - Transform array](./map.md)
- [filter() - Filter array](./filter.md)
