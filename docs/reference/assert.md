# assert()

Check a condition and throw an error if false. Essential for testing and validation.


```
assert(condition [, message])
```

## Parameters

- `condition` (any) - Expression to check. Truthy values pass; falsy values (false, nil, 0, "") throw
- `message` (optional, string) - Custom error message. Defaults to "assertion failed"

## Returns

`nil` if condition is truthy. Throws an error if condition is falsy.

## Examples

Basic assertion:

```duso
assert(1 == 1, "numbers should be equal")
print("✓ assertion passed")

// output: ✓ assertion passed
```

Assertion with expression:

```duso
ds = datastore("users")
ds.set("alice", {age = 30})
u = ds.get("alice")
assert(u.age >= 18, "user must be adult")
print("Assertion passed")

// output: Assertion passed
```

Catching assertion failures:

```duso
try
  assert(false, "this will fail")
catch (e)
  print("Caught error: {{e}}")
end

// output: Caught error: error("this will fail")
```

Testing datastore operations:

```duso
ds = datastore("test")
ds.set("key", "value")
assert(ds.get("key") == "value", "get should return set value")
assert(ds.exists("key") == true, "exists should return true")

d = ds.delete("key")
assert(d == "value", "delete should return deleted value")
assert(ds.get("key") == nil, "key should not exist after delete")
print("All assertions passed")

// output: All assertions passed
```

## Truthiness Rules

- `true` → truthy
- `false`, `nil` → falsy
- `0` → falsy; non-zero numbers → truthy
- `""` (empty string) → falsy; non-empty strings → truthy
- Objects, arrays, functions → truthy (unless nil)

## Error Handling

Assertions throw errors that include call stack information. Use `try/catch` to handle assertion failures:

```duso
x = -5
try
  assert(x > 0, "x must be positive")
  print("x is valid")
catch (e)
  print("Validation failed: {{e}}")
end

// output: Validation failed: error("x must be positive")
```

## See Also

- [throw()](/docs/reference/throw.md) - Throw a custom error
- [try/catch](/docs/reference/try.md) - Error handling
