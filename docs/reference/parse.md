# parse()

Parse source code string into a code value or error value. Never throws—always returns a value.

## Signature

```duso
parse(source [, metadata])
```

## Parameters

- `source` (string) - Duso source code to parse
- `metadata` (optional, object) - User-defined metadata to attach to the code value

## Returns

- **code value** - If parsing succeeds
- **error value** - If parsing fails

Parse errors return an error value rather than throwing, making it safe for dynamic code generation.

## Examples

Parse valid code:

```duso
c = parse("exit(42)")
print(type(c))        // "code"
result = run(c)
print(result)         // 42
```

Parse with metadata:

```duso
code = parse("exit(100 * 2)", {
  origin = "ai-agent",
  model = "claude-3"
})
print(code.metadata.origin)  // "ai-agent"
```

Handle parse errors:

```duso
code = parse("if x")  // Invalid syntax
if type(code) == "error" then
  print("Parse error: " + code.message)
end
```

## See Also

- [code - Code values](/docs/reference/code.md)
- [error - Error values](/docs/reference/error.md)
- [run() - Execute code synchronously](/docs/reference/run.md)
- [spawn() - Execute code asynchronously](/docs/reference/spawn.md)
