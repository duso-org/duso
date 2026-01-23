# Duso Language Evaluation

*From the perspective of Luna, Lua Expert*

---

## Top 5 Pros

1. **Familiar Lua-inspired syntax with modern conveniences** — The `if/then/end`, `for/do/end`, and `function/end` block structure will feel immediately natural to Lua developers. However, Duso adds quality-of-life improvements like compound assignment (`+=`, `-=`), increment/decrement operators (`++`, `--`), and the ternary operator (`?:`), which Lua lacks. This reduces boilerplate without sacrificing readability.

2. **First-class string templates for LLM workflows** — The `{{expr}}` template syntax embedded directly in strings is elegant and purpose-built for the target domain. Unlike Lua's string concatenation or format functions, templates allow complex expressions including function calls and property access inline. The triple-quote multiline strings (`"""..."""`) with automatic whitespace trimming are particularly well-suited for crafting LLM prompts and JSON payloads without escaping nightmares.

3. **Objects-as-constructors pattern is brilliantly simple** — The ability to call any object as a constructor (`Config(timeout = 60)`) provides lightweight prototypal inheritance without the complexity of metatables or class systems. This is more intuitive than Lua's metatable-based OOP while remaining compositional. The implicit `self` binding for methods (accessing object properties directly without `self.`) reduces noise considerably.

4. **Sensible defaults for a scripting language** — 0-based indexing (unlike Lua's 1-based), built-in JSON parsing/formatting, and the `var` keyword for explicit local scoping address common Lua pain points. The truthiness rules (empty arrays/objects are falsy) are more intuitive than Lua's "only nil and false are falsy" approach. Type coercion is thoughtfully limited rather than pervasive.

5. **Clean Go integration story** — The explicit design for host-provided functions (`claude()`, `conversation()`, file I/O) with the concurrency model delegated to Go is pragmatic. Scripts remain single-threaded and predictable while the host handles parallelism. This separation of concerns is cleaner than embedding async primitives in the language itself.

---

## Top 5 Cons

1. **Scoping rules are a footgun waiting to happen** — The "assignment without `var` walks up the scope chain" behavior is Lua's original sin, and Duso inherits it. While the `var` keyword exists, it's optional, meaning accidental global pollution or unintended outer-scope mutation is the default behavior. Lua 5.4 addressed this with `<const>` and stricter warnings; Duso should consider making `var` mandatory or defaulting to local scope.

2. **No module system is a significant gap** — Listed under "Future Features (Deferred)" but critical for any non-trivial project. The `include()` function executes in the current environment, which is namespace pollution by design. Without proper imports/exports, organizing agent orchestration code across files becomes unwieldy. Even Lua 5.0 had `require()`.

3. **Limited collection operations for a data-processing language** — For LLM workflows that often involve transforming arrays of results, the absence of `map()`, `filter()`, `reduce()`, and `find()` is notable. The `sort()` function exists but requires verbose comparison functions. These are trivial to implement in userland but should be built-in for a language targeting this domain.

4. **Error handling is stringly-typed** — The `catch (error)` block receives only a string message, not a structured error object. There's no way to distinguish error types, attach metadata, or implement error hierarchies. For agent orchestration where you might want to retry on rate limits but fail on authentication errors, this is limiting. Lua's `pcall`/`xpcall` with stack traces is more powerful.

5. **Ambiguous specification for edge cases** — What happens with `for i = 1.5, 10`? (Spec says "error" but doesn't specify behavior.) Can object keys be computed (`{[expr]: value}`)? What's the iteration order for object keys? How does `break` inside nested loops behave? These gaps will lead to implementation-defined behavior that surprises users.

---

## Top 5 Questions

1. **Why not default to local scope and require `global` for outer access?** — The current design where `x = 5` might create a local or modify an outer variable depending on context is the most criticized aspect of Lua's design. JavaScript moved to `let`/`const` defaults for this reason. What was the rationale for preserving Lua's original behavior?

2. **How does garbage collection work, and what are the memory semantics for closures?** — The spec mentions closures capture their environment, but doesn't address memory lifecycle. Can circular references between closures and objects cause leaks? Is there a weak reference mechanism? For long-running agent processes, this matters.

3. **What is the execution model for `conversation()` state across script invocations?** — If a script creates a `conversation()` and exits, is that state serializable? Can it persist across multiple script runs? Agent workflows often require suspending and resuming conversations; the spec is silent on this critical use case.

4. **Are there plans for async/await or coroutines?** — The spec states parallelism is "handled by the host," but many agent patterns involve interleaved I/O (wait for user input, then LLM call, then tool execution). Lua's coroutines elegantly solve this. Will Duso support yield/resume semantics, or is this permanently delegated to Go?

5. **How does the language handle untrusted scripts in a sandboxed environment?** — For agent orchestration, scripts may come from users or be generated by LLMs. Can host functions be selectively exposed? Is there execution time/memory limiting? Can scripts access the file system or network beyond what the host provides? The security model is undefined.