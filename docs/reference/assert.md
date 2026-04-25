# assert()

Check a condition and throw an error if false. Essential for testing and validation.

## Signature

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
```

Assertion with expression:

```duso
ds = datastore("users")
ds.set("alice", {age = 30})
user = ds.get("alice")
assert(user.age >= 18, "user must be adult")
```

Catching assertion failures:

```duso
try
  assert(false, "this will fail")
catch (err)
  print("Caught error: " + err)
end
```

Testing datastore operations:

```duso
ds = datastore("test")
ds.set("key", "value")
assert(ds.get("key") == "value", "get should return set value")
assert(ds.exists("key") == true, "exists should return true")

deleted = ds.delete("key")
assert(deleted == "value", "delete should return deleted value")
assert(ds.get("key") == nil, "key should not exist after delete")
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
try
  assert(x > 0, "x must be positive")
  print("x is valid")
catch (err)
  print("Validation failed: " + err)
  exit(1)
end
```

## See Also

- [throw()](/docs/reference/throw.md) - Throw a custom error
- [try/catch](/docs/reference/try.md) - Error handling
