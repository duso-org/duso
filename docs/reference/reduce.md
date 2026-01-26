# reduce()

Combine all elements of an array into a single value using an accumulator function.

## Signature

```duso
reduce(array, function, initial_value)
```

## Parameters

- `array` (array) - The array to reduce
- `function` (function) - Function taking (accumulator, element) and returning new accumulator
- `initial_value` - Starting value for the accumulator

## Returns

Final accumulated value

## Examples

Sum numbers:

```duso
numbers = [1, 2, 3, 4, 5]
sum = reduce(numbers, function(acc, x) return acc + x end, 0)
print(sum)                      // 15
```

Calculate product:

```duso
numbers = [1, 2, 3, 4, 5]
product = reduce(numbers, function(acc, x) return acc * x end, 1)
print(product)                  // 120
```

Build an object:

```duso
words = ["hello", "world", "duso"]
word_count = reduce(words, function(acc, word)
  acc[word] = 1
  return acc
end, {})
print(word_count)               // {hello=1 world=1 duso=1}
```

## See Also

- [map() - Transform array](./map.md)
- [filter() - Filter array](./filter.md)
