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

**Functional approach** - Use [`append()`](append.md) to create new array with element added:

```duso
arr = [1, 2, 3]
arr = append(arr, 4)
print(arr)      // [1 2 3 4]
```

**Mutable approach** - Use [`push()`](push.md) and [`unshift()`](unshift.md) to modify array in place:

```duso
arr = [2, 3]
push(arr, 4)              // Add to end, returns new length
unshift(arr, 1)           // Add to beginning, returns new length
print(arr)                // [1 2 3 4]
```

## Removing Elements

**Mutable approach** - Use [`pop()`](pop.md) and [`shift()`](shift.md) to modify array in place:

```duso
arr = [1, 2, 3, 4]
last = pop(arr)           // Remove and return last, arr = [1 2 3]
first = shift(arr)        // Remove and return first, arr = [2 3]
print(last, first)        // 4, 1
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

## Mutability & Thread Safety

**Arrays are mutable:** Functions like [`push()`](push.md), [`pop()`](pop.md), [`shift()`](shift.md), and [`unshift()`](unshift.md) modify arrays in place. Some functions like [`sort()`](sort.md), [`map()`](map.md), [`filter()`](filter.md), and [`reduce()`](reduce.md) return new arrays by design, but this is not because arrays are immutable—it's simply how those operations work.

**Deep Copying for Thread Safety:** When arrays pass between script scopes (entering/exiting function calls, spawned scripts, or datastore operations), they are automatically deep-copied to prevent race conditions. This ensures thread safety across concurrent operations while maintaining performance within a single scope.

Example:

```duso
arr = [1, 2, 3]
ds = datastore("test")

// Array is deep-copied going INTO datastore
ds.set("data", arr)

// Array is deep-copied coming OUT of datastore
retrieved = ds.get("data")

// Modifications to retrieved don't affect datastore copy
push(retrieved, 4)
print(ds.get("data"))     // [1 2 3] - unchanged
```

## See Also

- [Functional Programming](/docs/learning-duso.md#functional-programming)
