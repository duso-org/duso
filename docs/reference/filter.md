# filter()

Keep only array elements that match a predicate function.

## Signature

```duso
filter(array, function)
```

## Parameters

- `array` (array) - The array to filter
- `function` (function) - Predicate function that returns true to keep element, false to discard

## Returns

New array with only matching elements

## Examples

Filter even numbers:

```duso
numbers = [1, 2, 3, 4, 5, 6]
evens = filter(numbers, function(x) return x % 2 == 0 end)
print(evens)                    // [2 4 6]
```

Filter strings by length:

```duso
words = ["hi", "hello", "ok", "world", "duso"]
long_words = filter(words, function(w) return len(w) > 3 end)
print(long_words)               // [hello world duso]
```

Filter objects in array:

```duso
users = [
  {name = "Alice", age = 25},
  {name = "Bob", age = 17},
  {name = "Charlie", age = 30}
]
adults = filter(users, function(u) return u.age >= 18 end)
```

## See Also

- [map() - Transform array](./map.md)
- [reduce() - Reduce array](./reduce.md)
