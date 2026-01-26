# return

Send a value back from a function.

## Syntax

```duso
return
return value
```

## Description

The `return` statement exits the function immediately and optionally returns a value to the caller. If no value is specified, the function returns `nil`.

## Examples

Return a computed value:

```duso
function add(a, b)
  return a + b
end

result = add(5, 3)
print(result)  // 8
```

Return early based on condition:

```duso
function absolute(x)
  if x < 0 then
    return -x
  end
  return x
end

print(absolute(-5))   // 5
print(absolute(10))   // 10
```

No return value (returns nil):

```duso
function greet(name)
  print("Hello, " + name)
  // Implicitly returns nil
end

result = greet("Alice")
print(result)  // nil
```

## See Also

- [function](./function.md) - Define a function
