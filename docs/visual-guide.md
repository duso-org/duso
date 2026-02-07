# Duso Visual Guide: Concepts & Patterns

Practical visual explanations of Duso programming patterns and concepts.

## Data Structures

### Arrays: Ordered Collections

```
Variable:     arr = [1, "hello", true, 3.14]
                    ↓
Memory Model:
┌──────────┬──────────┬──────────┬──────────┐
│ Index 0  │ Index 1  │ Index 2  │ Index 3  │
├──────────┼──────────┼──────────┼──────────┤
│    1     │  "hello" │   true   │   3.14   │
│ (number) │ (string) │  (bool)  │ (number) │
└──────────┴──────────┴──────────┴──────────┘

Access:
arr[0]         → 1
arr[1]         → "hello"
arr[len(arr)-1] → 3.14

Operations:
arr.push(5)    → arr = [1, "hello", true, 3.14, 5]
arr.pop()      → returns 5, arr = [1, "hello", true, 3.14]
arr.map(fn)    → transform each element
arr.filter(fn) → keep matching elements
arr.join(",")  → "1,hello,true,3.14"
arr.reverse()  → [3.14, true, "hello", 1]
```

### Objects: Key-Value Maps

```
Variable:    user = {name: "Alice", age: 30, active: true}
                        ↓
Memory Model:
┌──────────────┬──────────────┬──────────────┐
│    "name"    │    "age"     │   "active"   │
├──────────────┼──────────────┼──────────────┤
│   "Alice"    │      30      │     true     │
└──────────────┴──────────────┴──────────────┘

Access:
user.name      → "Alice"
user["age"]    → 30 (bracket notation)
user.active    → true

Modification:
user.email = "alice@example.com"  → adds new key
user.age = 31                      → updates existing
delete(user, "active")             → removes key

Nested structures:
user = {
  name: "Alice",
  address: {
    street: "Main St",
    city: "Boston"
  }
}

user.address.city  → "Boston"
user["address"]["street"]  → "Main St"
```

## Control Flow

### if/elseif/else Logic

```
Flowchart:
┌──────────────┐
│ Check: x > 0 │
└──────┬───────┘
       │
    Yes│    No
       ├───────────────┐
       ↓               ↓
   ┌────────┐    ┌──────────────┐
   │ "pos"  │    │ Check: x < 0 │
   └────────┘    └──────┬───────┘
                   Yes  │  No
                        ├────────┐
                        ↓        ↓
                    ┌────────┐ ┌──────┐
                    │ "neg"  │ │"zero"│
                    └────────┘ └──────┘

Code:
if x > 0 then
  result = "positive"
elseif x < 0 then
  result = "negative"
else
  result = "zero"
end

Truthiness:
if x then
  // executes if x is "truthy"
  // truthy: non-zero numbers, non-empty strings, true, non-empty arrays
  // falsy: 0, "", false, nil, empty arrays
end
```

### Loops: for and while

```
while loop:
┌────────────┐
│ Check: i<5 │
└──────┬─────┘
    Yes│ No ─→ Exit
       │
       ↓
   ┌────────┐
   │ Do work│
   └────┬───┘
        ↓
    ┌────────┐
    │ i = i+1│
    └────┬───┘
         │
         └─→ Loop back to Check

Code:
i = 0
while i < 5 do
  print(i)
  i = i + 1
end
// Prints: 0 1 2 3 4


for loop:
Item 0: 1
Item 1: 2
Item 2: 3
Item 3: 4

Code:
for i = 0, len(arr)-1 do
  print(arr[i])
end

Or with range:
for i in range(10) do
  print(i)
end
// Prints: 0 1 2 3 4 5 6 7 8 9
```

### Pattern: Conditional Defaults

```
Pattern:  value = provided_value or default_value

Example:
name = request_name or "Guest"
timeout = config_timeout or 30
retry = false or true  // → true

How it works:
┌─────────────────────────────┐
│ Evaluate: x or y            │
├─────────────────────────────┤
│ if x is truthy → return x   │
│ if x is falsy  → return y   │
└─────────────────────────────┘

Note: Returns actual VALUES, not booleans!
  "hello" or 5     → "hello"
  nil or 10        → 10
  0 or 100         → 100 (0 is falsy)
  false or true    → true
```

## Functions

### Function Definition & Scope

```
┌──────────────────────────────────────┐
│ function calculate(a, b)             │
│   // Parameters: a, b (local)        │
│   // Can see: x, print (from parent) │
│                                      │
│   result = a + b                     │
│   return result                      │
│                                      │
│ end  // scope ends here              │
└──────────────────────────────────────┘

When called: calculate(5, 3)
┌──────────────────────────┐
│ Create new scope:        │
│ {                        │
│   a: 5,                  │
│   b: 3,                  │
│   result: undefined,     │
│   Parent: Global scope   │
│ }                        │
│                          │
│ → a + b = 8              │
│ → return 8               │
│ → scope cleaned up       │
└──────────────────────────┘

Return behavior:
┌────────────────┐
│  return 42     │
│                │
│  ↓ Signals     │
│  Stop execution│
│  Exit function │
│  Yield value   │
└────────────────┘

Implicit return (last value):
function greet(name)
  "Hello, " + name
end
// Returns the last expression without explicit return
```

### Callbacks & Higher-Order Functions

```
Pattern: Pass functions to other functions

Code:
numbers = [1, 2, 3, 4, 5]

squared = numbers.map(function(x)
  return x * x
end)
// squared = [1, 4, 9, 16, 25]

Flow:
numbers = [1, 2, 3, 4, 5]
    ↓
map(function(x) return x*x end)
    │
    ├─→ Call with x=1 → 1
    ├─→ Call with x=2 → 4
    ├─→ Call with x=3 → 9
    ├─→ Call with x=4 → 16
    └─→ Call with x=5 → 25
    ↓
[1, 4, 9, 16, 25]
```

## String Operations

### String Concatenation & Templates

```
Simple concatenation:
"Hello" + " " + "World"  → "Hello World"

With variables:
name = "Alice"
greeting = "Hello, " + name + "!"  → "Hello, Alice!"

Template strings (preferred for complex):
name = "Alice"
age = 30
message = """
  Name: {{name}}
  Age: {{age}}

  This is a multi-line template.
"""

Result:
  Name: Alice
  Age: 30

  This is a multi-line template.

Template expression evaluation:
x = 5
y = 10
result = "{{x}} + {{y}} = {{x + y}}"
// → "5 + 10 = 15"

Any expression works:
items = [1, 2, 3]
"Length: {{len(items)}}"  → "Length: 3"

user = {name: "Bob"}
"Name: {{user.name}}"  → "Name: Bob"
```

### String Methods

```
String value methods:
┌─────────────────────────────┐
│ "hello world"               │
├─────────────────────────────┤
│ .len()                      │
│   → 11                      │
│                             │
│ .upper()                    │
│   → "HELLO WORLD"           │
│                             │
│ .lower()                    │
│   → "hello world"           │
│                             │
│ .split(" ")                 │
│   → ["hello", "world"]      │
│                             │
│ .split("")                  │
│   → ["h","e","l","l","o"...]│
│                             │
│ .trim()                     │
│   → removes whitespace      │
│                             │
│ .contains("world")          │
│   → true                    │
│                             │
│ .replace("world", "Duso")   │
│   → "hello Duso"            │
│                             │
│ .substr(0, 5)               │
│   → "hello"                 │
└─────────────────────────────┘
```

## Error Handling

### try/catch Pattern

```
Normal flow:
┌────────────────┐
│ Operation      │
└────────┬───────┘
         │ Success
         ↓
   ┌─────────┐
   │ Result  │
   └─────────┘

With error:
┌────────────────┐
│ Operation      │
└────────┬───────┘
         │ Error!
         ↓
   ┌─────────────┐
   │ Error value │
   └─────────────┘

Using try/catch:
try
  result = risky_operation()
catch(err)
  // err contains error message
  result = default_value
end

Flow diagram:
┌──────────────────────┐
│ try block            │
├──────────────────────┤
│ risky_operation()    │
│        ↓             │
│  Error thrown?       │
│   Yes ↙      ↘ No   │
│   ↓          ↓      │
│ catch      Success   │
│ block      → result  │
│  ↓                   │
│Error value           │
│→ result=default      │
│        ↓             │
│ Continue execution   │
└──────────────────────┘

Example:
try
  data = load("config.json")
  config = parse_json(data)
catch(err)
  print("Failed to load config: " + err)
  config = {timeout: 30}  // fallback
end
```

## Pattern: Request/Response Pattern

```
Used in: http_server, spawn with context, datastore

Request:
┌─────────────────────────────────┐
│ Main process sends request      │
│                                 │
│ data = {                        │
│   user_id: 123,                 │
│   action: "fetch_user"          │
│ }                               │
│                                 │
│ spawn("worker.du", data)        │
└──────────────┬──────────────────┘
               │
               ↓
Spawned process receives:
┌──────────────────────────────────┐
│ ctx = context()                  │
│ // ctx.user_id = 123             │
│ // ctx.action = "fetch_user"     │
│                                  │
│ Do work based on context data... │
│                                  │
│ Store result in shared datastore:│
│ cache = datastore("results")     │
│ cache.set("result_123", {...})   │
└──────────┬───────────────────────┘
           │
           ↓
Main process retrieves:
┌──────────────────────────────────┐
│ cache = datastore("results")     │
│ result = cache.get("result_123") │
│ // result = {...}                │
│                                  │
│ Use the result...                │
└──────────────────────────────────┘
```

## Common Patterns

### Pattern: Accumulate Results

```
Processing multiple items and collecting results:

items = [1, 2, 3, 4, 5]
results = []

for i = 0, len(items)-1 do
  processed = items[i] * 2
  results.push(processed)
end

// results = [2, 4, 6, 8, 10]

Or with map (cleaner):
results = items.map(function(x)
  return x * 2
end)
```

### Pattern: Filter and Transform

```
Select items matching condition, transform them:

users = [
  {name: "Alice", age: 30},
  {name: "Bob", age: 25},
  {name: "Carol", age: 35}
]

// Filter: age > 28
adults = users.filter(function(u)
  return u.age > 28
end)
// [{name: "Alice"...}, {name: "Carol"...}]

// Transform: get names
names = adults.map(function(u)
  return u.name
end)
// ["Alice", "Carol"]

// Chain them:
result = users
  .filter(function(u) return u.age > 28 end)
  .map(function(u) return u.name end)
// ["Alice", "Carol"]
```

### Pattern: Parallel Processing

```
Run multiple expensive operations concurrently:

Code:
results = parallel(
  function()
    return fetch("https://api1.example.com/data")
  end,
  function()
    return fetch("https://api2.example.com/data")
  end,
  function()
    return fetch("https://api3.example.com/data")
  end
)

Execution timeline:
Time →
Start ├─── API 1 (2s) ─┐
      ├─── API 2 (3s) ─┤
      └─── API 3 (1s) ─┘
                       End
                       3s total (not 6s)

Results:
results[0]  → API 1 response
results[1]  → API 2 response
results[2]  → API 3 response
```

### Pattern: Cache with Fallback

```
Try cache first, compute if miss:

cache = datastore("cache")

function get_user(user_id)
  cached = cache.get("user_" + user_id)
  if cached != nil then
    print("From cache: " + user_id)
    return cached
  end

  // Cache miss - compute
  print("Computing: " + user_id)
  user = fetch_user_from_db(user_id)

  // Store for next time
  cache.set("user_" + user_id, user)
  return user
end

Flow:
first call:  cache miss → compute → store → return
second call: cache hit → return immediately
```

### Pattern: Scatter-Gather with Spawning

```
Spawn workers to process items in parallel:

work_items = [1, 2, 3, 4, 5]
results = datastore("scatter_gather")
pids = []

// Scatter: spawn workers
for i = 0, len(work_items)-1 do
  pid = spawn("worker.du", {
    item_id: i,
    value: work_items[i]
  })
  pids.push(pid)
end

// Workers process in parallel...
// Each stores result via: cache.set("result_N", value)

// Gather: collect results
gathered = []
for i = 0, len(work_items)-1 do
  result = results.get("result_" + i)
  gathered.push(result)
end

print(gathered)  // All results collected
```

## Comparison Matrix

```
Array vs Object:

┌────────┬───────────────┬──────────────────┐
│        │  Array        │  Object          │
├────────┼───────────────┼──────────────────┤
│ Order  │ Ordered (0,1,2│ Unordered (keys) │
│        │   indexed)    │ (semantic order) │
├────────┼───────────────┼──────────────────┤
│ Access │ arr[0]        │ obj.key          │
│        │ arr[i]        │ obj["key"]       │
├────────┼───────────────┼──────────────────┤
│ Use    │ Lists, sequences│ Records,configs │
│        │ Collections   │ Dictionaries     │
├────────┼───────────────┼──────────────────┤
│Mixed   │ [1,"hi",true] │ {x:5,y:"test"}   │
│types   │ Heterogeneous │ Heterogeneous    │
├────────┼───────────────┼──────────────────┤
│ Create │ [1,2,3]       │ {a:1, b:2}       │
│        │ []            │ {}               │
└────────┴───────────────┴──────────────────┘
```

## Debug Tips

### Printing Values

```
// Simple print
print(x)

// Multiple values
print(x, y, z)  → x y z (space-separated)

// String conversion
print("Value: " + x)

// JSON output
print(format_json({user: "alice", age: 30}))
// Pretty-printed JSON

// Type check
print(type(x))  → "number", "string", "array", etc.

// Conditional debug
if debug then
  print("DEBUG: x = " + format_json(x))
end
```

### Stack Traces

When an error occurs, Duso shows:
```
DusoError: Undefined variable 'foo'
  at line 42, column 15 (myScript.du)
  in function 'calculate'
    called from line 25 (myScript.du)
  in main script

This tells you:
├─ What went wrong: Undefined variable 'foo'
├─ Where: line 42, column 15
├─ Context: Inside function 'calculate'
└─ Called from: line 25
```

---

## Quick Pattern Reference

| Pattern | Purpose | Example |
|---------|---------|---------|
| **Conditional defaults** | Provide fallback values | `x = input or 42` |
| **Accumulate** | Collect results | `results.push(item)` |
| **Filter/map** | Transform collections | `arr.filter(...).map(...)` |
| **Parallel** | Concurrent operations | `parallel(fn1, fn2, fn3)` |
| **Cache fallback** | Speed with freshness | `cache.get() or compute()` |
| **Scatter-gather** | Distributed work | `spawn(worker, data)` |
| **Request/response** | Inter-process comm | `context()` + `datastore()` |
| **Try/catch** | Error recovery | `try op catch(e)` |

