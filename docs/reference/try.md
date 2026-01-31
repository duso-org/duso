# try

Execute code that might error, with optional error handling.

## Syntax

```duso
try
  // statements that might error
catch (error_var)
  // handle error
end
```

## Description

The `try` block executes statements. If an error occurs, execution jumps to the `catch` block (if present) with the error message as a string. Without `catch`, errors propagate up to the caller.

## Examples

Catching and handling errors:

```duso
try
  data = load("config.json")
  print("Loaded successfully")
catch (error)
  print("Failed to load: " + error)
  data = {}
end
```

Graceful degradation:

```duso
try
  response = make_api_call()
  result = parse_json(response)
catch (e)
  print("API error: " + e)
  result = {status = "offline"}
end

print(result)
```

Nested try blocks:

```duso
try
  file = load("data.txt")
  try
    data = parse_json(file)
  catch (parse_error)
    print("JSON parse error: " + parse_error)
  end
catch (file_error)
  print("File load error: " + file_error)
end
```

## Error Variable

The variable name in `catch (error_var)` receives the error message as a string:

```duso
try
  result = 1 / 0
catch (err)
  print(type(err))  // "string"
  print(err)        // "division by zero"
end
```

## See Also

- [catch](/docs/reference/catch.md) - Handle errors
- [throw](/docs/reference/throw.md) - Raise an error
