# Nil

Nil represents the absence of a value. It's the default return value for functions that don't explicitly return anything.

## Creating Nil

Use the `nil` keyword:

```duso
x = nil
empty = nil
```

## When Nil Appears

Nil is returned:
- From functions with no `return` statement
- When accessing undefined variables in error context
- As a default parameter value when not provided

```duso
function no_return()
  print("Does nothing")
  // Implicitly returns nil
end

result = no_return()
print(result)  // nil
```

## Checking for Nil

Compare against `nil`:

```duso
x = nil
if x == nil then
  print("x is nil")
end

if x != nil then
  print("x has a value")
end
```

## Truthiness

In conditions, nil is falsy:

```duso
if nil then print("true") end        // doesn't print
if nil == false then print("true") end // doesn't print (they're different)
```

## Type Conversion

Convert to string with [`tostring()`](tostring.md):

```duso
s = tostring(nil)  // "nil"
```

Convert to boolean with [`tobool()`](tobool.md):

```duso
b = tobool(nil)  // false
```

