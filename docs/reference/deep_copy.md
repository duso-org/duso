# deep_copy()

Create a deep (recursive) copy of a value. Arrays and objects are recursively copied; functions are removed (they don't work out of scope).

## Signature

```duso
deep_copy(value)
```

## Parameters

- `value` (any) - The value to deep copy

## Returns

A deep copy of the value with all nested arrays and objects independently copied.

## Behavior

- **Primitives** (number, string, boolean, nil) - Returned as-is
- **Arrays** - Recursively copied; functions become nil
- **Objects** - Recursively copied; functions are removed
- **Functions** - Become nil (safety feature—functions don't work out of scope)

## Examples

Copy an array:

```duso
original = [1, 2, 3]
copy = deep_copy(original)
copy[0] = 999
print(original[0])  // 1 (unchanged)
print(copy[0])      // 999
```

Copy a nested structure:

```duso
original = {
  name = "Alice",
  scores = [10, 20, 30]
}
copy = deep_copy(original)
copy.scores[0] = 999
print(original.scores[0])  // 10 (unchanged)
print(copy.scores[0])      // 999
```

Functions are removed:

```duso
obj = {
  value = 42,
  get_value = function()
    return value
  end
}
copy = deep_copy(obj)
print(copy.value)        // 42 (data preserved)
print(copy.get_value)    // nil (function removed)
```

Compare with shallow copy (constructor pattern):

```duso
original = {data = [1, 2, 3]}

// Shallow copy - shares nested references
shallow = original()
shallow.data[0] = 999
print(original.data[0])  // 999 (affected!)

// Deep copy - independent
deep = deep_copy(original)
deep.data[0] = 999
print(original.data[0])  // 1 (unaffected!)
```

## Use Cases

- Sending data across scope boundaries (run(), spawn(), exit(), datastore())
- Creating truly independent copies of complex structures
- Ensuring nested arrays/objects can't be accidentally modified in other scopes

## See Also

- [Constructor Pattern](/docs/learning-duso.md#copying-data-shallow-vs-deep) - Shallow copy with overrides
- [Arrays](/docs/reference/array.md) - Array type reference
- [Objects](/docs/reference/object.md) - Object type reference
