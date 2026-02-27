# error

First-class error value containing message and stack trace. Created by `parse()` on failure or caught in try/catch blocks.

## Description

An `error` value wraps a thrown value and its formatted stack trace. Error values are returned by `parse()` on syntax errors and bound to catch variables in `try/catch` blocks.

Unlike older patterns where only the message survived, error values preserve full context including file, line number, and call stack.

## Properties

- `.message` - The thrown value (string, object, or any Duso value)
- `.stack` (string) - Formatted error details with file, line number, and call stack

## Type Checking

```duso
e = parse("if x")  // Invalid syntax
print(type(e) == "error")  // true
```

## From parse()

`parse()` returns an error value on syntax errors:

```duso
bad_code = parse("invalid @@@ syntax")
if type(bad_code) == "error" then
  print("Parse failed: " + bad_code.message)
end
```

## From try/catch

Catch blocks now bind error values:

```duso
try
  throw("something went wrong")
catch (e)
  print(type(e) == "error")    // true
  print(e.message)             // "something went wrong"
  print(len(e.stack) > 0)      // true (has stack trace)
end
```

## Stack Traces

Access full error context:

```duso
function f()
  throw("error in f")
end

try
  f()
catch (e)
  print("Error: " + e.message)
  print("Stack:\n" + e.stack)
end
```

Output includes:
- File and line number where error occurred
- Function call chain
- Full context for debugging

## Throwing Values

`throw()` can throw any value:

```duso
// String
throw("simple error")

// Object with details
throw({code = "AUTH_FAILED", user_id = 123})

// Array
throw([1, 2, 3])
```

The caught error's `.message` is whatever was thrown:

```duso
try
  throw({status = 404, reason = "not found"})
catch (e)
  print(e.message.status)  // 404
  print(e.message.reason)  // "not found"
end
```

## Examples

Graceful error handling:

```duso
function divide(a, b)
  if b == 0 then
    throw({code = "DIV_BY_ZERO", value = a})
  end
  exit(a / b)
end

try
  result = run(parse("divide(10, 0)"))
catch (e)
  print("Error: " + e.message.code)
  print("Details: " + e.stack)
end
```

Parse error handling:

```duso
code = parse(user_input)
if type(code) == "error" then
  print("Invalid code: " + code.message)
  print("Details: " + code.stack)
else
  result = run(code)
end
```

## See Also

- [parse() - Create code or error values](/docs/reference/parse.md)
- [code - Code values](/docs/reference/code.md)
- [throw() - Throw an error](/docs/reference/throw.md)
- [try/catch - Error handling](/docs/reference/try.md)
