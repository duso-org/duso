# end

Closes a block statement (function, if, while, for, try).

## Syntax

```duso
if condition then
  // statements
end

while condition do
  // statements
end

for i = 0; i < 10; i = i + 1 do
  // statements
end

function myFunc() do
  // statements
end

try
  // statements
catch(error) do
  // handle error
end
```

## Description

The `end` keyword terminates a code block that was opened by `if`, `elseif`, `while`, `for`, `function`, or `try`. Every block-opening statement requires a matching `end` to define where the block ends.

`end` is required—Duso uses explicit block terminators rather than indentation or braces to delimit scope.

## Examples

Closing an if block:

```duso
if x > 10 then
  print("Greater than 10")
end
```

Closing nested blocks:

```duso
function checkValue(x) do
  if x > 0 then
    print("Positive")
  end
  return x
end
```

Closing a try block:

```duso
try
  result = load("data.json")
catch(err) do
  print("Error: " + err)
end
```

## See Also

- [if](/docs/reference/if.md) - Conditional statements
- [while](/docs/reference/while.md) - While loops
- [for](/docs/reference/for.md) - For loops
- [function](/docs/reference/function.md) - Function declarations
- [try](/docs/reference/try.md) - Error handling
