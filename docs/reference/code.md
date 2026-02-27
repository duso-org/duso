# code

Pre-parsed source code with metadata. Created by `parse()`.

## Description

A `code` value represents Duso source code that has been parsed into an abstract syntax tree (AST). It's immutable and can be executed with `run()` or `spawn()`, persisted in a datastore, or passed between scripts.

Code values never contain closures or external scope—they're self-contained and serializable.

## Properties

- `.source` (string) - Original source code string
- `.metadata` (object) - User-provided metadata (empty object if not specified)

## Type Checking

```duso
c = parse("exit(1)")
print(type(c) == "code")  // true
```

## Creating Code Values

Use `parse()`:

```duso
code = parse("exit(42)")
code_with_meta = parse("exit(42)", {purpose = "answer"})
```

## Executing Code

Run synchronously:

```duso
c = parse("exit(10 + 20)")
result = run(c)
print(result)  // 30
```

Spawn asynchronously:

```duso
c = parse("print('background task')")
spawn(c, {})
```

## Storing Code

Code values are serializable (store as source string):

```duso
store = datastore("code_library")
code = parse("exit(42)", {name = "answer"})
store.set("answer", code)

// Later:
retrieved = store.get("answer")
result = run(retrieved)
```

## Code with Metadata

Attach context to code:

```duso
code = parse("exit(x + 1)", {
  agent = "claude",
  task = "increment",
  version = 1
})

print(code.metadata.agent)  // "claude"
print(code.metadata.task)   // "increment"
```

## Examples

Agent-generated code:

```duso
// Simulate agent generating code
generated = parse("exit(42)", {
  origin = "ai-agent",
  timestamp = now()
})

if type(generated) == "code" then
  result = run(generated)
  print("Agent result: " + result)
end
```

Code in collections:

```duso
codes = [
  parse("exit(1)", {name = "first"}),
  parse("exit(2)", {name = "second"}),
  parse("exit(3)", {name = "third"})
]

for code in codes do
  print(code.metadata.name + ": " + run(code))
end
```

## See Also

- [parse() - Create code values](/docs/reference/parse.md)
- [error - Error values](/docs/reference/error.md)
- [run() - Execute code](/docs/reference/run.md)
- [spawn() - Execute code asynchronously](/docs/reference/spawn.md)
