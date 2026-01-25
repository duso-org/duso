# breakpoint()

Pause execution and enter interactive debug mode. Available in `duso` CLI only with `-debug` flag.

## Signature

```duso
breakpoint()
```

## Parameters

None

## Returns

`nil`

## Usage

The `breakpoint()` function only has an effect when running a script with the `-debug` flag:

```bash
duso -debug script.du
```

Without `-debug`, `breakpoint()` does nothing and execution continues normally.

## Examples

Pause at a specific point in script:

```duso
x = 42
breakpoint()  // Execution pauses here in debug mode
print(x)
```

Conditional debugging:

```duso
for i = 1, 100 do
  if i == 50 then
    breakpoint()  // Only pause when i reaches 50
  end
end
```

## Debug Output

When a breakpoint is hit, you see:

- File path where the breakpoint occurred
- Full call stack showing all function calls leading to the breakpoint
- Position (line and column) for each call

Example output:

```
[Debug] Breakpoint hit at script.du

Call stack:
  at inner (script.du:2:9)
  at outer (script.du:6:9)
  at main (script.du:9:3)

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

- Only works when script is run with `-debug` flag
- Useful for inspecting program state at critical points
- Can be left in production code; it's a no-op without `-debug`
- Call stack helps identify the execution path leading to the breakpoint

## See Also

- [throw() - Throw an error](./throw.md)
- [print() - Output text](./print.md)
- [CLI reference](../cli/README.md)
