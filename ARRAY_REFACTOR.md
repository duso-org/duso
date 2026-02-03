# Array Reference Type Refactor

## Problem
Current immutable arrays require copying on every operation (sort, map, filter). This is expensive in loops with large arrays. Arrays are stored by value, so modifications don't affect the original without reassignment. Immutability provides no safety since spawned scripts get their own scope copies anyway.

## Solution
Change arrays from pass-by-value `[]Value` to reference type `*[]Value`. This allows:
- In-place mutations without copying
- Natural semantics for push/pop/shift/unshift (no reassignment needed)
- Same isolation for spawned scripts (via deep copy at boundaries)

## Changes Required

### 1. value.go (~10 lines)
- Change array storage: `[]Value` → `*[]Value`
- Update `NewArray()` to allocate pointer
- Update `AsArray()` getter to dereference
- Update `IsArray()` if needed

### 2. evaluator.go (~15 lines)
- `evalArrayLiteral()`: allocate pointer to slice
- `evalIndexExpr()`: dereference before indexing
- `evalIndexAssign()`: dereference before assignment

### 3. builtins.go (~150 lines)
- Add `deepCopy(Value) Value` function (~30 lines) - recursive walk for nested arrays/objects
- Update array functions to work in-place:
  - `len()`: dereference pointer
  - `sort()`: modify in-place, return nil or length
  - `map()`: modify in-place, return nil or length
  - `filter()`: compact in-place, return nil or length
  - `reduce()`: no change needed
- Implement `push(arr, item)`: append in-place, return length
- Implement `pop(arr)`: remove from end, return item
- Implement `shift(arr)`: remove from start, return item
- Implement `unshift(arr, item)`: prepend in-place, return length

### 4. http_server_value.go (~10 lines)
- In `ExecuteScript()` when calling spawn/run: deep copy context via `deepCopy()`
- Ensures spawned scripts get isolated copies

### 5. Tests
- Update array tests to work with pointer semantics
- Add tests for push/pop/shift/unshift
- Scattered small changes in existing tests

## API Changes
```duso
// Before (reassignment required)
arr = sort(arr)
arr = map(arr, fn)

// After (in-place, no reassignment)
sort(arr)
map(arr, fn)
push(arr, item)
pop(arr)
shift(arr)
unshift(arr, item)
```

## Performance Impact
- **Positive**: No copying in loops, in-place mutations are O(n) or O(1) instead of O(n) copy + operation
- **Neutral**: Spawned scripts still deep-copy context (same as now)
- **Minimal overhead**: One extra pointer dereference per array access

## Risk
Low. Changes are mechanical and isolated to array handling. Spawned script isolation maintained via deep copy. Existing tests catch regressions.
