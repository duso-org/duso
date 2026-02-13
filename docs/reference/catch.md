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

The `catch` clause catches errors thrown by the `try` block. The error variable receives the thrown value, which can be any type: a string (standard), object, array, number, etc. The `catch` block is optional—without it, errors propagate to the caller.

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

Accessing properties of an error object:

```duso
try
  validate_input(data)
catch (err)
  if err.code then
    print("Error code: " + err.code)
    print("Field: " + err.field)
    print("Message: " + err.message)
  else
    print("Error: " + err)
  end
end
```

## See Also

- [try](/docs/reference/try.md) - Execute code that might error
- [throw](/docs/reference/throw.md) - Raise an error
