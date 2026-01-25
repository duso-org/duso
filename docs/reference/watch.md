# watch()

Monitor expression values during debugging. When a watched expression's value changes, automatically break and enter debug mode. A core language feature that can be enabled with the `DebugMode` setting (set automatically by the `-debug` CLI flag).

## Signature

```duso
watch(expr1)
watch(expr1, expr2, expr3, ...)
```

## Parameters

- `expr1, expr2, ...` (required) - String expressions to monitor. Each must be a valid Duso expression that can be evaluated in the current scope.

## Returns

`nil`

## Usage

The `watch()` function is a core language feature that can be enabled by setting `DebugMode`.

**In the CLI**, use the `-debug` flag:

```bash
duso -debug script.du
```

**In embedded Go applications**, enable debug mode on the interpreter:

```go
interp := script.NewInterpreter(false)
interp.SetDebugMode(true)  // Enable watch() functionality
```

Without `DebugMode` enabled, `watch()` is a no-op and execution continues normally.

Each call to `watch()` evaluates its expressions and caches the values. On subsequent calls, if any watched expression has changed, a breakpoint is triggered and you enter debug mode.

## Examples

Watch a single variable:

```duso
count = 0
for i = 1, 100 do
  count = count + 1
  watch("count")  // Breaks every iteration as count changes
end
```

Watch multiple expressions at once:

```duso
user = {id = 1, name = "Alice", status = "active"}
watch("user.status", "user.id", "len(user)")
// If any of these change, break and show all of them
```

Watch with conditions:

```duso
queue = []
for i = 1, 1000 do
  append(queue, i)
  watch("len(queue) > 500")  // Watch a boolean expression
  if len(queue) > 1000
    process_queue(queue)
    queue = []
  end
end
```

Watch nested structures:

```duso
data = {
  users = [{id = 1, active = true}, {id = 2, active = false}],
  timestamp = now()
}
watch("data.users[0].active", "data.timestamp")
// Monitor specific nested values
```

## Debug Output

When a watched expression changes, you see output like:

```
WATCH: count = 42
WATCH: user.status = inactive

[Debug] Breakpoint hit at script.du:8:5

Call stack:
  at process (script.du:5:10)

Type 'c' to continue, or inspect variables.
debug>
```

Each changed expression prints on its own line with the format: `WATCH: <expression> = <value>`

## How Watch Caching Works

- First time you call `watch("expr")`: evaluates and caches the value
- Subsequent calls: evaluates again, compares to cached value
  - If different: prints `WATCH:` line and triggers breakpoint
  - If same: continues without breaking
- Cache is **global by expression**, not scoped
  - `watch("i")` tracks the same `i` whether called in a loop, function, or main scope

## Interactive Mode

When a watched expression triggers a breakpoint, you can:

- Inspect variables by name
- Execute arbitrary Duso expressions
- Type `c` to continue execution
- Type `exit` to exit the debugger

## Notes

- Only activates when `DebugMode` is enabled (CLI: `-debug` flag, embedded: `SetDebugMode(true)`)
- Without `DebugMode`, `watch()` is a complete no-op (no overhead)
- Expressions are always evaluated every time `watch()` is called (to update cache), but breakpoint only triggers if `DebugMode` is true
- Changes are detected by value comparison (deep equality for arrays/objects)
- Multiple watch expressions in one call are more efficient than separate calls
- Watch cache is global—same expression is tracked everywhere
- Can be left in production code as debugging markers; team members see them when debugging with `DebugMode` enabled
- A core language feature, available in both CLI and embedded applications

## See Also

- [breakpoint() - Immediate debugging pause](./breakpoint.md)
- [print() - Output text](./print.md)
- [CLI reference](../cli/README.md)
