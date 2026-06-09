# Duso Code Style Guide

## Indentation
- Use spaces, 2 per indentation level
- No tabs

## Strings

Strings allow for UTF-8 in Duso. Most other values will coerce cleanly to strings without an explicit to_string(). You may concat strings with `+` **NOT** `.` or `..`.

### Single vs Multi-line
- **Single-line**: `'...'` or `"..."`
- **Multi-line**: `"""..."""` or `'''...'''`

### String Templates
- All string types support template expressions `{{...}}`
- **Preferred over concatenation**: Use `"value is {{x}}"` instead of `"value is " + x`
- All strings support template expressions: wrap in `{{...}}`
- Expressions can include: variables, inline valus, function calls, math, ternary operators
- Expressions **cannot** include: loops, branching (if/else)
- Indentation: multi-line strings intelligently strip left-edge space based on shortest common leading whitespace

### String Template Examples
```duso
// Single-line with template
x = 10
s = "x equals {{x}}"

// Ternary in template
status = "Status: {{alive ? 'alive' : 'dead'}}"

// Function call in template
greeting = "Welcome {{get_name()}}"

// Multi-line with template
msg = """
  Hello {{name}},
  Your score is {{score}}.
"""
```

### Do Not Use template() Builtin
Use inline string templates instead. The template() function has a rare and specific purpose.

## Variables

### Naming
- Use **snake_case** for variables and functions
- Use **UPPER_CASE** for constants (naming convention only—duso doesn't enforce immutability)

### Length and Context
- Prefer short names: 1-2 letters when reasonable (i, j, x, y, z, n, s, h, etc.)
- When comparing: use `a`, `b`
- When more letters are needed:
  - Use abbreviations: `pwd` for password, `ts` for timestamp, `h` for hash
  - Use short words: `user`, `req`, `msg`, `data`, `item`
  - **Base on context**: if processing a user document, use `user` not `doc` or `document`

### Avoid
- "Cheap" names with no semantic value: `my_doc`, `my_input`, `my_var`
- Underscore-prefixed names to simulate privacy: `_private`, `_internal`
- Variables used only once (unless it's an illustrative example or best practice point)

## Comments

### Style
- C-style: `//` for single-line, `/* ... */` for multi-line
- Single-line comments only: one per statement or small block
- Multi-line comments only for multi-line text

### Guidelines
- Avoid comments on the same line as code: **don't** write `x = x + 1  // increment x`
- If you write 3+ consecutive single-line comments, switch to a multi-line comment block instead
- Comment the **why**, not the what—code should be clear enough for the what

### Examples
```duso
// Good
h = hash_password(pwd)
ts = now()

// Bad
x = x + 1  // increment x

// Switch to multi-line for multiple comments
/*
  Fetch user by ID, validate permissions,
  then update their profile in the database.
  Returns nil if validation fails.
*/
```

## Function Arguments

Duso supports positional and named arguments. They may be mixed in a function call, but positional must come first, and switch to named in that case. You can't go from poss to nae back to pos, for example.

### Named Arguments
- Use named arguments to break up long argument lists and improve clarity
- Prefer: `set(x, y, relative = true)` over `set(x, y, nil, nil, true)`
- Named arguments make intent clear and avoid positional ambiguity

### Examples
```duso
// Good: named args skip unused positionals
result = api_call(url, method = "POST", timeout = 30)

// Avoid: positional args with nils for skipped params
result = api_call(url, "POST", nil, nil, 30)
```

## Loops and Control Flow

### For Loops
Use `for...do...end` syntax. Two forms:

```duso
// Numeric range (inclusive)
for i = 1, 10 do
  print(i)
end

// Iterate over array
for item in items do
  print(item)
end
```

### While Loops
```duso
count = 0
while count < 5 do
  print(count)
  count = count + 1
end
```

### If/Else
```duso
if age >= 18 then
  print("Adult")
elseif age >= 13 then
  print("Teenager")
else
  print("Child")
end
```

### Ternary for Simple Conditions
```duso
status = age >= 18 ? "adult" : "minor"
```

## Functions

### Definition
```duso
function greet(name)
  return "Hello, {{name}}"
end

// Or assign to variable
double = function(x)
  return x * 2
end
```

### Higher-Order Functions

When passing function literals inside another function, indent them to make it clear at a glance what's happening. See examples.

```duso
// Map: transform each element
squared = map(nums, function(x)
  return x * x
end)

// Filter: keep matching elements
evens = filter(nums, function(x)
  return x % 2 == 0
end)

// Reduce: combine to single value
sum = reduce(nums, function(acc, x)
  return acc + x
end, 0)
```

## Objects and Arrays

### Objects (Key-Value Maps)
```duso
// Define
person = {
  name = "Alice",
  age = 30,
  city = "Portland"
}

// Preffered Access
print(person.name)

// secondary access (useful for variable keys or illegal keys)
field = "name"
print(person[field])
print(req["Content-Type"])

// Modify
person.age = 31
```

### Arrays (Ordered Lists)
```duso
nums = [10, 20, 30]

// 0-indexed
print(nums[0])

// Length
print(len(nums))

// Add to end
push(nums, 40)

// Iterate
for n in nums do
  print(n)
end
```

### Objects with Methods
```duso
// Object with functions as methods
agent = {
  name = "Alice",
  age = 30,
  greet = function(msg)
    return "{{msg}}, I am {{name}} (age {{age}})"
  end,
  birthday = function()
    age = age + 1
  end
}

// Call methods
print(agent.greet("Hello"))
agent.birthday()
```

### Objects as Blueprints/Constructors
```duso
// Blueprint
config = {timeout = 30, retries = 3}

// Create copy with defaults
c1 = config()

// Override specific fields
c2 = config(timeout = 60)

// Counter "class" pattern
counter = {
  count = 0,
  increment = function()
    count = count + 1
  end,
  value = function()
    return count
  end
}

// Create independent instances
counter1 = counter()
counter2 = counter()

counter1.increment()
print(counter1.value())  // 1
print(counter2.value())  // 0 (independent)
```

## Duso Script File Naming
- Use **snake_case** for all script filenames
- End with `.du` extension
- Example: `fetch_user.du`, `hash_password.du`, `string_utils.du`

## MArkdown File Names
- use lower-case
- separate with `-` NOT `_`
- end with `.md`

## Error Handling

### Try/Catch
```duso
try
  data = load("config.json")
catch (e)
  print("Failed to load: {{e}}")
end
```

### Throw Errors
```duso
function validate(age)
  if age < 0 then
    throw("Age cannot be negative")
  end
  return age
end
```

## Summary of Key Differences from Lua/JavaScript

- **No curly braces for blocks** - use `end` keyword instead
- **Multi-line strings with triple quotes** - no escaping needed
- **Built-in string templates** - use `{{expr}}` in any string
- **Named arguments** - prefer over positional for clarity
- **Objects as constructors** - call with `()` to create copies
- **Array/object operations** - mutable (push, pop) and functional (map, filter) available
- **Scope handling** - use `var` to explicitly create locals, assignments look up scope chain

## General Principles

1. **Prefer short, context-aware names** - `user` not `my_user`, `h` not `my_hash`
2. **Use string templates over concatenation** - `"Hello {{name}}"` not `"Hello " + name`
3. **Keep functions small and focused** - one responsibility
4. **Use named arguments for clarity** - especially with 3+ parameters
5. **Comment the why, not the what** - code should be self-documenting
6. **Indent consistently with 2 spaces** - no tabs, ever
