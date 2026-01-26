# Function

Functions are reusable blocks of code that can be called with arguments and return values.

## Defining Functions

```duso
function greet(name)
  return "Hello, " + name
end

function add(a, b)
  return a + b
end
```

Functions with a single expression can use compact syntax:

```duso
double = function(x) return x * 2 end
```

## Calling Functions

```duso
result = greet("Alice")   // "Hello, Alice"
result = add(5, 3)        // 8
```

## Parameters and Arguments

Functions can accept positional or named arguments:

```duso
function configure(timeout, retries, verbose)
  // ...
end

configure(30, 3, true)                // Positional
configure(timeout = 60, retries = 5)  // Named
configure(30, verbose = false)        // Mixed
```

## Return Values

Use `return` to send a value back to the caller:

```duso
function compute(x)
  if x < 0 then
    return 0
  end
  return x * 2
end
```

Functions without an explicit return return `nil`.

## Closures

Functions capture their definition environment, allowing them to access variables from outer scopes:

```duso
function makeAdder(n)
  function add(x)
    return x + n  // Captures n from outer scope
  end
  return add
end

addFive = makeAdder(5)
print(addFive(10))  // 15
print(addFive(20))  // 25
```

## Function Expressions

Functions can be assigned to variables or object properties:

```duso
double = function(x) return x * 2 end
print(double(5))    // 10

obj = {
  callback = function(msg)
    print("Message: " + msg)
  end
}

obj.callback("Hello")  // "Message: Hello"
```

## Truthiness

In conditions, functions are always truthy:

```duso
f = function() return 42 end
if f then print("true") end  // prints
```

