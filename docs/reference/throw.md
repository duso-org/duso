# throw()

Throw an error with a message. Includes call stack and position information for debugging.

## Signature

```duso
throw(message)
```

## Parameters

- `message` (string) - Error message describing what went wrong

## Returns

Never returns (throws an error)

## Examples

Throw on invalid condition:

```duso
function divide(a, b)
  if b == 0
    throw("Cannot divide by zero")
  end
  return a / b
end

result = divide(10, 0)
```

Throw with descriptive context:

```duso
function validate_email(email)
  if not contains(email, "@") then
    throw("Invalid email: missing @")
  end
  parts = split(email, "@")
  if len(parts[0]) == 0 then
    throw("Invalid email: empty local part")
  end
  return email
end
```

## Error Output

When `throw()` is called, the error displays with:

- File path, line number, and column of the throw call
- Full call stack showing all function calls leading to the error
- Error message

Example output:

```
Error: script.du:5: Invalid input value

Call stack:
  at validate (/script.du:5:5)
  at process_data (/script.du:10:8)
```

## Notes

- Throws interrupt execution immediately
- The call stack helps identify where an error originated from
- Errors can be caught with `try/catch` blocks

## See Also

- [print() - Output text](/docs/reference/print.md)
