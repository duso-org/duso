# map()

Transform each element in an array by applying a function.

## Signature

```duso
map(array, function)
```

## Parameters

- `array` (array) - The array to transform
- `function` (function) - Function that takes one element and returns transformed value

## Returns

New array with transformed elements

## Examples

Double each number:

```duso
numbers = [1, 2, 3, 4, 5]
doubled = map(numbers, function(x) return x * 2 end)
print(doubled)                  // [2 4 6 8 10]
```

Convert to uppercase:

```duso
words = ["hello", "world", "duso"]
upper_words = map(words, function(w) return upper(w) end)
print(upper_words)              // [HELLO WORLD DUSO]
```

Use named function:

```duso
function square(n)
  return n * n
end
numbers = [1, 2, 3, 4, 5]
squares = map(numbers, square)
print(squares)                  // [1 4 9 16 25]
```

## See Also

- [filter() - Filter array](/docs/reference/filter.md)
- [reduce() - Reduce array](/docs/reference/reduce.md)
