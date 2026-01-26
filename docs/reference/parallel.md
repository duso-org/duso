# parallel()

Execute functions concurrently and collect results in parallel.

## Signature

```duso
parallel(function1, function2, ...)       // Varargs (primary)
parallel(array_of_functions)              // Array form
parallel(object_of_functions)             // Object form
```

## Parameters

- `function1, function2, ...` (functions) - One or more functions to execute in parallel
- `array_of_functions` (array) - Array of functions to execute in parallel
- `object_of_functions` (object) - Object with function values to execute in parallel

## Returns

- Varargs or array input: Array with results in same order
- Object input: Object with same keys and result values

## Key Characteristics

- **True parallelism** - Functions run concurrently in isolated evaluators
- **Read-only parent scope** - Functions can read parent variables but cannot modify them
- **Error handling** - If a function errors, that result becomes `nil`
- **Order preservation** - Results maintain input structure

## Examples

Simplest form - pass functions directly (varargs):

```duso
topic = "Duso"
language = "Go"

results = parallel(
  function()
    return "What is {{topic}}? It's a scripting language for agents."
  end,

  function()
    return "{{topic}} is built on {{language}} stdlib."
  end,

  function()
    return "You can build web scrapers, AI agents, and data pipelines."
  end
)

for i in results do
  print(i)
end
```

Array form - for computed lists:

```duso
functions = [
  function() return "Task 1" end,
  function() return "Task 2" end,
  function() return "Task 3" end
]

results = parallel(functions)
print(results)
```

Object form - for named results:

```duso
user_id = 42

results = parallel({
  profile = function()
    return "User {{user_id}} profile data"
  end,

  activity = function()
    return "User {{user_id}} activity logs"
  end,

  recommendations = function()
    return "User {{user_id}} recommendations"
  end
})

print("Profile: " + results.profile)
print("Activity: " + results.activity)
print("Recommendations: " + results.recommendations)
```

Parallel API calls with Claude:

```duso
claude = require("claude")
topic = "machine learning"

results = parallel(
  function()
    return claude.prompt("Explain {{topic}} to a beginner in 2 sentences.")
  end,

  function()
    return claude.prompt("List 3 advanced concepts in {{topic}}.")
  end,

  function()
    return claude.prompt("Name 5 applications of {{topic}}.")
  end
)

print("Beginner: " + results[0])
print("Advanced: " + results[1])
print("Applications: " + results[2])
```

Handle errors in parallel execution:

```duso
results = parallel(
  function()
    return "success"
  end,

  function()
    error = 1 / 0  // This will error
  end,

  function()
    return "also success"
  end
)

// Check for nil results (errors become nil)
if results[0] != nil then
  print("Task 1: " + results[0])
end

if results[1] == nil then
  print("Task 2 failed")
end

if results[2] != nil then
  print("Task 3: " + results[2])
end
```

## Important Notes

- Parent scope is **read-only** - functions can read variables but assignments stay local
- Use `parallel()` for **independent operations** (API calls, data fetches, computations)
- Not suitable for operations that need to share mutable state
- Errors in one function don't stop others - check for `nil` results
- Results order matches input order - predictable and safe

## See Also

- [map() - Transform array](./map.md)
- [filter() - Filter array](./filter.md)
