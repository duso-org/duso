# Duso Language Evaluation

## Top 5 Pros

1. **Excellent LLM Integration Design** - The triple-quoted multiline strings with `{{template}}` syntax are perfectly suited for prompt engineering. No escaping quotes in JSON, clean embedding of expressions, and automatic whitespace handling make this ideal for the stated use case of agent orchestration.

2. **Objects-as-Constructors Pattern is Elegant** - The ability to call any object as a constructor with named argument overrides (`Config(timeout = 60)`) provides lightweight prototypal inheritance without the complexity of class hierarchies. This aligns well with configuration-heavy agent workflows.

3. **Pragmatic Scoping with `var` Keyword** - The explicit choice between modifying outer scope (no `var`) and creating locals (`var x = 0`) gives developers control while maintaining simplicity. The closure support enables powerful patterns like `makeCounter` without complex syntax.

4. **Host Integration Philosophy is Sound** - Keeping the language single-threaded while delegating parallelism to the Go host via `parallel()` and custom functions is architecturally clean. This allows Go's concurrency primitives to handle the hard problems while scripts remain predictable.

5. **Zero External Dependencies** - Building entirely on Go stdlib with baked-in modules eliminates version conflicts and dependency hell. The "freeze at release" approach for stdlib/contrib ensures script reproducibility indefinitely—critical for production agent systems.

---

## Top 5 Cons

1. **Implicit Scope Modification is Error-Prone** - The default behavior of walking up the scope chain for assignment (without `var`) inverts the common pattern from most languages. This will cause subtle bugs when developers accidentally modify outer variables, especially in nested functions or loops.

2. **Weak Type Safety for Enterprise Use** - While loose typing aids scripting convenience, enterprise C# developers expect stronger guarantees. No way to declare expected types, no interface contracts, and silent coercion (e.g., `"10" > 5` being valid) can hide bugs until runtime in production.

3. **Limited Error Context** - The specification mentions "clean error messages" but shows only basic string errors in catch blocks. No stack traces, no line numbers in the spec, no error codes or structured error types. Debugging complex agent workflows will be painful.

4. **No Async/Await for Sequential Async Operations** - While `parallel()` handles concurrent operations, there's no mechanism for sequential async operations (e.g., `await http.fetch()`). Long-running LLM calls would block the entire interpreter unless the host implements workarounds.

5. **Object Iteration Returns Keys Only** - When iterating objects with `for key in obj`, you only get keys and must separately access values. This differs from most modern languages and adds boilerplate. A `pairs()` or `entries()` function is notably listed as "deferred."

---

## Top 5 Questions

1. **How does error handling interact with `parallel()`?** - The spec states failed parallel operations return `nil`, but how do you distinguish between a function that legitimately returned `nil` versus one that errored? Is there a way to capture the actual error message from failed parallel branches?

2. **What happens when object method references escape their object context?** - If I do `callback = agent.greet` and later call `callback("Hello")`, does it still resolve `name` and `skill` from `agent`? Or does the implicit binding break when the method is detached from its object?

3. **How are circular references handled in `format_json()`?** - If an object contains a reference to itself (directly or indirectly), what happens when serializing to JSON? Does it error, infinite loop, or have cycle detection?

4. **What's the memory model for large scripts with many closures?** - Since closures capture their entire definition environment, do long-running agent orchestrations risk memory leaks? Is there any garbage collection, and how aggressive is it?

5. **How do `require()` cached modules interact with mutable state?** - If module A returns an object, and scripts B and C both `require("A")`, they share the same cached object. If B mutates it, does C see those changes? Is this intentional shared state, and how should developers manage it?