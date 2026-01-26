# var

Explicitly declare a local variable that shadows outer scope variables.

## Syntax

```duso
var name = value
```

## Description

The `var` keyword creates a new local variable, shadowing any variable with the same name in an outer scope. Without `var`, assignment walks up the scope chain: if a variable exists in any parent scope, it modifies that variable; otherwise, it creates a new local.

## Examples

Shadowing outer variables:

```duso
x = 10
function test()
  var x = 0  // New local x, shadows outer
  x = x + 1
  print(x)   // 1
end
test()
print(x)     // Still 10 (outer unchanged)
```

Without `var`, modifies outer:

```duso
x = 10
function modify()
  x = x + 5  // No var = modifies outer x
end
modify()
print(x)     // 15
```

In closures:

```duso
function makeCounter(start)
  var count = start  // Local to this function
  function increment()
    count = count + 1  // Modifies captured count
    return count
  end
  return increment
end

counter = makeCounter(0)
print(counter())  // 1
print(counter())  // 2
```

## Best Practice

Use `var` to be explicit about scope and prevent accidental mutations of outer variables:

```duso
function process(items)
  var result = []  // Clearly a new local
  for item in items do
    result = append(result, item * 2)
  end
  return result
end
```

## See Also

- [Function scope](../learning_duso.md#variable-scope)
