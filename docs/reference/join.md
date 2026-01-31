# join()

Join array elements into a single string with a separator.

## Signature

```duso
join(array, separator)
```

## Parameters

- `array` (array) - The array to join
- `separator` (string) - String to insert between elements

## Returns

Single string with elements joined by separator

## Examples

Join with comma:

```duso
items = ["apple", "banana", "orange"]
result = join(items, ", ")
print(result)                   // "apple, banana, orange"
```

Join with different separators:

```duso
words = ["hello", "world"]
print(join(words, " "))         // "hello world"
print(join(words, "-"))         // "hello-world"
print(join(words, " | "))       // "hello | world"
```

Join numbers:

```duso
numbers = [1, 2, 3, 4, 5]
result = join(numbers, ",")
print(result)                   // "1,2,3,4,5"
```

## See Also

- [split() - Split string into array](/docs/reference/split.md)
