# catch

Handle errors from a try block.

## Syntax

```duso
try
  // statements
catch (error_var)
  // handle error
end
```

## Description

The `catch` clause catches errors thrown by the `try` block. The error variable receives the error message as a string. The `catch` block is optional—without it, errors propagate to the caller.

## Examples

Basic error handling:

```duso
try
  x = undefined_variable
catch (e)
  print("Error: " + e)
end
```

Different error types:

```duso
try
  arr = [1, 2, 3]
  print(arr[10])  // Out of bounds
catch (e)
  print("Array error: " + e)
end

try
  result = 5 / 0  // Division by zero
catch (e)
  print("Math error: " + e)
end
```

Recovery and fallbacks:

```duso
config = {}
try
  config = load("config.json")
catch (e)
  print("Could not load config: " + e)
  // Fallback to defaults
  config = {timeout = 30, retries = 3}
end
```

Inspecting the error:

```duso
try
  operation()
catch (err)
  if contains(err, "timeout") then
    print("Request timed out, retrying...")
  elseif contains(err, "unauthorized") then
    print("Authentication failed")
  else
    print("Unknown error: " + err)
  end
end
```

## See Also

- [try](./try.md) - Execute code that might error
- [throw](./throw.md) - Raise an error
