# breakpoint()

Pause execution and enter interactive debug mode. A core language feature that can be enabled with the `DebugMode` setting (set automatically by the `-debug` CLI flag).

## Signature

```duso
breakpoint()
breakpoint(value1, value2, ...)
```

## Parameters

- `value1, value2, ...` (optional) - Values to print before hitting the breakpoint. Works like `print()`; useful for debugging diagnostics without extra statements.

## Returns

`nil`

## Usage

The `breakpoint()` function is a core language feature that can be enabled by setting `DebugMode`.

**In the CLI**, use the `-debug` flag:

```bash
duso -debug script.du
```

**In embedded Go applications**, enable debug mode on the interpreter:

```go
interp := script.NewInterpreter(false)
interp.SetDebugMode(true)  // Enable breakpoint() functionality
```

Without `DebugMode` enabled, `breakpoint()` is a no-op and execution continues normally.

## Examples

Pause at a specific point in script:

```duso
x = 42
breakpoint()  // Execution pauses here in debug mode
print(x)
```

Print diagnostic information before breaking:

```duso
user = {id = 123, name = "Alice", score = 95}
breakpoint("user data:", user, "score is:", user.score)
// Output: BREAKPOINT: user data: map[id:123 name:Alice score:95] score is: 95
// Then drops to debug> prompt
```

Conditional debugging with context:

```duso
for i = 1, 100 do
  if i == 50 then
    breakpoint("Loop iteration {{i}} reached")  // Message with template
  end
end
```

Team debugging annotations:

```duso
result = expensive_operation(data)
if result.confidence < 0.7
    breakpoint("LOW CONFIDENCE: {{result.confidence}}, input: {{input}}")
end
// Team leaves these markers in code as documentation of known problem areas
```

## Debug Output

When a breakpoint is hit, you see:

- Any arguments printed (if provided)
- File path where the breakpoint occurred
- Full call stack showing all function calls leading to the breakpoint
- Position (line and column) for each call

Example output without arguments:

```
[Debug] Breakpoint hit at script.du:5:1

Call stack:
  at inner (script.du:2:9)
  at outer (script.du:6:9)

Type 'c' to continue, or inspect variables.
debug>
```

Example output with arguments:

```
BREAKPOINT: user data: map[id:123 name:Alice score:95] score is: 95

[Debug] Breakpoint hit at script.du:12:5

Call stack:
  at process (script.du:8:10)

Type 'c' to continue, or inspect variables.
debug>
```

## Interactive Mode

When a breakpoint is hit in debug mode, you can:

- Inspect variables by name
- Execute arbitrary Duso expressions
- Type `c` to continue execution
- Type `exit` to exit the debugger

## Notes

- Only activates when `DebugMode` is enabled (CLI: `-debug` flag, embedded: `SetDebugMode(true)`)
- Without `DebugMode`, `breakpoint()` is a complete no-op (no overhead)
- Useful for inspecting program state at critical points
- Can be left in production code as debugging annotations; team members will see them when debugging with `DebugMode` enabled
- Call stack helps identify the execution path leading to the breakpoint
- Arguments are printed using the same logic as `print()`, so all values are space-separated
- Argument evaluation happens before the breakpoint, so you can use template strings and expressions
- A core language feature, available in both CLI and embedded applications

## See Also

- [throw() - Throw an error](/docs/reference/throw.md)
- [print() - Output text](/docs/reference/print.md)
- [CLI reference](/docs/cli/README.md)
