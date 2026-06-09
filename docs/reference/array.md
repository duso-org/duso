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
a = ["a", "b", "c"]
print(a[0], a)

/*
  output:
  a, ["a" "b" "c"]
*/
```

Out-of-bounds access causes an error.

## Length

Get the number of elements with [`len()`](/docs/reference/len.md):

```duso
a = [1, 2, 3]
print(len(a))

// output: 3
```

## Adding Elements

Use [`push()`](/docs/reference/push.md) and [`unshift()`](/docs/reference/unshift.md) to modify array in place:

```duso
a = [2, 3]
push(a, 4)
unshift(a, 1)
print(a)

// output: [1, 2, 3, 4]
```

## Removing Elements

**Mutable approach** - Use [`pop()`](/docs/reference/pop.md) and [`shift()`](/docs/reference/shift.md) to modify array in place:

```duso
a = [1, 2, 3, 4]
l = pop(a)
f = shift(a)
print(l, f)

// output: 4 1
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

- [`map()`](/docs/reference/map.md) - Transform each element
- [`filter()`](/docs/reference/filter.md) - Keep matching elements
- [`reduce()`](/docs/reference/reduce.md) - Combine into single value
- [`sort()`](/docs/reference/sort.md) - Sort elements

```duso
d = map([1, 2, 3], function(x) return x * 2 end)
e = filter([1, 2, 3, 4], function(x) return x % 2 == 0 end)
s = reduce([1, 2, 3], function(a, x) return a + x end, 0)
print(d)
print(e)
print(s)

/*
  output:
  [2, 4, 6]
  [2, 4]
  6
*/
```

## Truthiness

In conditions, non-empty arrays are truthy:

```duso
if [1, 2, 3] then print("true") end
if [] then print("true") end

// output: true
```

## Mutability & Thread Safety

**Arrays are mutable:** Functions like [`push()`](/docs/reference/push.md), [`pop()`](/docs/reference/pop.md), [`shift()`](/docs/reference/shift.md), and [`unshift()`](/docs/reference/unshift.md) modify arrays in place. Some functions like [`sort()`](/docs/reference/sort.md), [`map()`](/docs/reference/map.md), [`filter()`](/docs/reference/filter.md), and [`reduce()`](/docs/reference/reduce.md) return new arrays by design, but this is not because arrays are immutable—it's simply how those operations work.

**Deep Copying for Thread Safety:** When arrays pass between script scopes (entering/exiting function calls, spawned scripts, or datastore operations), they are automatically deep-copied to prevent race conditions. This ensures thread safety across concurrent operations while maintaining performance within a single scope.

Example:

```duso
a = [1, 2, 3]
ds = datastore("test")
ds.set("data", a)

retrieved = ds.get("data")
push(retrieved, 4)
print(ds.get("data"))

// output: [1, 2, 3]
```

## See Also

- [Functional Programming](/docs/learning-duso.md#functional-programming)
