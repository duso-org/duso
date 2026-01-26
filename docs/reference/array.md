# Array

Arrays are ordered lists of values, 0-indexed. Elements can be any type.

## Creating Arrays

```duso
numbers = [1, 2, 3, 4, 5]
mixed = [1, "hello", true, nil]
empty = []
```

## Accessing Elements

Arrays are 0-indexed (first element is at position 0):

```duso
arr = ["a", "b", "c"]
print(arr[0])   // "a"
print(arr[1])   // "b"
print(arr[2])   // "c"
```

Out-of-bounds access causes an error.

## Length

Get the number of elements with [`len()`](len.md):

```duso
arr = [1, 2, 3]
print(len(arr))  // 3
```

## Adding Elements

Use [`append()`](append.md) to add an element (returns new array):

```duso
arr = [1, 2, 3]
arr = append(arr, 4)
print(arr)      // [1 2 3 4]
```

## Iteration

Loop through arrays with `for...in`:

```duso
items = ["apple", "banana", "cherry"]
for item in items do
  print(item)
end
```

## Functional Operations

Transform arrays with built-in functions:

- [`map()`](map.md) - Transform each element
- [`filter()`](filter.md) - Keep matching elements
- [`reduce()`](reduce.md) - Combine into single value
- [`sort()`](sort.md) - Sort elements

```duso
doubled = map([1, 2, 3], function(x) return x * 2 end)
evens = filter([1, 2, 3, 4], function(x) return x % 2 == 0 end)
sum = reduce([1, 2, 3], function(a, x) return a + x end, 0)
```

## Truthiness

In conditions, non-empty arrays are truthy:

```duso
if [1, 2, 3] then print("true") end  // prints
if [] then print("true") end         // doesn't print
```

## See Also

- [Functional Programming](../learning_duso.md#functional-programming)
