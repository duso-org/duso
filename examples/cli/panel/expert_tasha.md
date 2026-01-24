# Duso Language Evaluation

*Evaluated by Tasha, TypeScript Expert*

---

## Top 5 Pros

1. **Excellent String Templating for LLM Workflows**
   - The `{{expr}}` template syntax with multiline string support (`"""..."""`) is perfectly tailored for prompt engineering. No escaping needed for JSON or code blocks, which eliminates a massive pain point when building LLM integrations. This is genuinely better than most mainstream languages for this specific use case.

2. **Pragmatic Object-as-Constructor Pattern**
   - The ability to call objects like `Config(timeout = 60)` to create copies with overrides is elegant and removes boilerplate. Combined with implicit `this` binding for methods (accessing `name` directly instead of `self.name`), it enables a clean, composition-friendly OOP style without class ceremony.

3. **Zero External Dependencies Philosophy**
   - Building entirely on Go's stdlib with frozen, baked-in modules is a bold architectural choice that solves real operational pain. No package manager, no version conflicts, no supply chain attacks. Archive script + binary = reproducible forever. This is refreshing in an ecosystem drowning in dependency hell.

4. **Thoughtful Parallel Execution Primitive**
   - The `parallel()` function with read-only parent scope access is a smart design. It provides real concurrency benefits for independent operations (multiple API calls) while preventing the shared mutable state footguns that plague concurrent programming. The array/object form flexibility is well-considered.

5. **Clean Go Integration Story**
   - The explicit design for "Easy Go function integration" with host-provided functions (`claude()`, `load()`, etc.) makes this genuinely useful as an embedded scripting layer. The separation between core language and host-provided capabilities is architecturally sound.

---

## Top 5 Cons

1. **No Static Type Information Whatsoever**
   - As a TypeScript expert, this is painful. No type hints, no interfaces, no way to express contracts. For agent orchestration where you're parsing LLM responses into expected shapes, you have zero compile-time safety. A `parse_json()` result is just... a value. You discover schema mismatches at runtime, which is exactly when you don't want to discover them in production agent systems.

2. **Confusing Scope Semantics with `var`**
   - The default behavior of "assignment without `var` walks up the scope chain" is a footgun. JavaScript learned this lesson painfully. Having to remember to use `var` for local variables inverts the common case—most variables should be local. This will cause subtle bugs in larger scripts, especially with nested functions and closures.

3. **Limited Error Handling Granularity**
   - `catch (error)` only captures the error message as a string. No error types, no stack traces in the language, no ability to catch specific error categories. For agent orchestration where you need to distinguish between "rate limited, retry" vs "invalid input, abort" vs "network timeout, fallback," string parsing is inadequate.

4. **No Async/Await or Promise Model**
   - While `parallel()` handles the embarrassingly parallel case, there's no way to express more complex async patterns—sequential-with-early-exit, race conditions, timeout wrappers, or dynamic parallelism. The "host handles it" approach pushes complexity to Go, fragmenting logic between two languages.

5. **Weak Collection Operations**
   - `map`, `filter`, `reduce` exist but there's no `find`, `findIndex`, `some`, `every`, `flatMap`, `groupBy`, or object spread/merge. For data transformation pipelines (common in agent workflows processing tool results), you'll write verbose loops for operations that should be one-liners.

---

## Top 5 Questions

1. **How do you validate LLM response shapes before using them?**
   - Given `result = parse_json(llm_response)`, what's the recommended pattern for verifying `result` has expected fields with expected types? Is there a schema validation function planned, or do users write manual `if result.field != nil and type(result.field) == "string"` checks everywhere?

2. **What happens when `parallel()` functions need to communicate or share results?**
   - If I have 3 parallel API calls where the second depends on knowing if the first succeeded (not its result, just success/failure), how do I model that? The read-only parent