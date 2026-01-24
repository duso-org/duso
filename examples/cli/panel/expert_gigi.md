# Duso Language Evaluation

## Top 5 Pros

1. **Clean template syntax for LLM workflows** - The `{{expr}}` template syntax inside strings (including multiline `"""` blocks) is genuinely useful for prompt engineering. No escaping needed for JSON braces, and expressions evaluate in-place. This is the right primitive for the domain.

2. **Objects-as-constructors is elegant** - Using `Config = {timeout = 30}` then `Config(timeout = 60)` for instantiation is simple and avoids class/prototype complexity. It's immediately understandable and covers 90% of OOP needs without the machinery.

3. **Frozen binary distribution model** - Baking stdlib/contrib modules into the binary with no external dependencies is operationally sound. Scripts + binary archived together work forever. This sidesteps the npm/pip dependency hell that plagues most scripting languages.

4. **`parallel()` with read-only parent scope** - The design choice to let parallel blocks read parent scope but not write to it is correct. It prevents data races while remaining useful. Returns `nil` on error per-slot rather than failing everything—practical for resilient agent workflows.

5. **Consistent `=` syntax throughout** - Object literals, named arguments, and constructor calls all use `=`. No context-switching between `:` and `=` like Lua. This reduces cognitive load and makes the language feel cohesive.

---

## Top 5 Cons

1. **Scope rules are confusing** - Assignment without `var` walks up the scope chain to find and modify outer variables, but creates a local if nothing is found. This is the worst of both worlds: easy to accidentally mutate parent scope, yet `var` is "optional." Go solved this cleanly with `:=` vs `=`. Duso should require `var` for new locals or use a different operator.

2. **No method receiver/self** - Methods access object properties through implicit scope lookup (`name` instead of `self.name`). This works for simple cases but becomes confusing with nested objects, callbacks, or when a method is extracted and called elsewhere. The "magic" of how `name` resolves is opaque.

3. **Arrays and objects are falsy when empty** - `[]` and `{}` being falsy is a footgun. Code like `if results then` will fail silently when results is an empty array. Most languages (including Go, JavaScript, Python) treat empty containers as truthy. This will cause bugs.

4. **`try/catch` with string errors only** - Error handling is stringly-typed with no error codes, types, or structured data. For agent orchestration where you need to distinguish rate limits from auth failures from timeouts, `catch (error)` where `error` is just a string is insufficient.

5. **Numeric for-loop requires integers but numbers are float64** - The spec says loop bounds "must be integers" but the only number type is float64. This means runtime errors for `for i = 1.5, 10` rather than a type-system guarantee. It's a leaky abstraction that will surprise users.

---

## Top 5 Questions

1. **How does method binding actually work when methods are passed around?** - If I do `callback = agent.greet` then `callback("Hi")`, does `name` still resolve to `agent.name`? Or does it break because there's no implicit binding like JavaScript's `this`? The spec doesn't address extracted methods.

2. **What happens when `parallel()` functions capture and read a variable that's being modified?** - The spec says functions can "READ parent variables" but doesn't specify when the read happens. If I modify a variable between defining the parallel functions and calling `parallel()`, which value do they see? Is it captured at definition or at `parallel()` call time?

3. **How does the module cache interact with `parallel()`?** - If two parallel blocks both `require("http")` for the first time, do they race to populate the cache? Is there a lock? Or does each parallel block get its own module cache (violating the "cached" semantics)?

4. **What's the memory model for large LLM responses?** - Agent workflows often handle large text payloads (10KB+ responses, concatenated conversation history). Are strings immutable with copy-on-write? Does concatenation with `+` create quadratic memory usage in loops? This matters for production agent systems.

5. **How do Go host functions signal different error types to D