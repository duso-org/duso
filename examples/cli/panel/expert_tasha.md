# Duso Language Evaluation

*Perspective: TypeScript Expert with Enterprise Static Types Background*

---

## Top 5 Pros

1. **Excellent Template Syntax for LLM Integration**
   - The `{{expr}}` template syntax with multiline triple-quoted strings is exceptionally well-designed for LLM prompt engineering. Unlike TypeScript's template literals, you don't need to escape quotes inside JSON or code snippets. The natural handling of `{` and `}` without escaping makes constructing JSON payloads trivial—a genuine pain point in most languages when working with LLM APIs.

2. **Pragmatic Object-as-Constructor Pattern**
   - The ability to call objects as constructors (`Config(timeout = 60)`) provides a lightweight, composition-friendly approach to object instantiation without the ceremony of classes. This is particularly elegant for configuration-heavy orchestration scenarios where you want immutable-by-default copies with selective overrides. TypeScript requires significantly more boilerplate for equivalent patterns.

3. **Clean Conversation State Management**
   - The `conversation()` API with maintained context across `.prompt()` calls elegantly solves a common agent orchestration problem. The stateful conversation object pattern is intuitive and maps well to how developers mentally model multi-turn interactions. This is a well-considered domain-specific abstraction.

4. **Sensible Scoping with Explicit `var` Keyword**
   - The scoping model—where assignment walks up the scope chain but `var` creates explicit locals—strikes a reasonable balance between convenience and predictability. For scripting purposes, this is often more practical than JavaScript's `let`/`const`/`var` trichotomy, especially for less experienced users writing orchestration scripts.

5. **Host-Delegated Concurrency Model**
   - The explicit decision to keep the language single-threaded while delegating parallelism to Go host functions is architecturally sound. This avoids the complexity of async/await, promises, or callback hell while still enabling parallel operations where needed. For the target use case (agent orchestration), this is the right tradeoff.

---

## Top 5 Cons

1. **Complete Absence of Static Type Information**
   - As a TypeScript expert, the lack of any type system—even optional type hints—is concerning for enterprise adoption. There's no way to express contracts between functions, no IDE autocomplete support, no compile-time error detection, and no documentation-as-types. For orchestration scripts that will inevitably grow complex, this creates significant maintainability risks. Even Python's optional typing would be preferable.

2. **Ambiguous Error Handling Semantics**
   - The error system relies entirely on string-based error messages caught in `catch (error)`. There's no typed error hierarchy, no way to distinguish between different error categories programmatically, and no mechanism for custom error types. In enterprise systems, you need to differentiate between retriable errors, fatal errors, and validation errors—string matching is brittle and unscalable.

3. **Implicit Method `this` Binding is Fragile**
   - The "magic" where method functions can access object properties without explicit `self` or `this` is clever but dangerous. When methods are extracted, passed as callbacks, or composed, this implicit binding will break silently. TypeScript's explicit `this` parameter typing exists precisely because implicit binding causes real production bugs. This design prioritizes brevity over safety.

4. **No Module or Namespace System**
   - The `include(filename)` approach pollutes the global namespace with no encapsulation. For any non-trivial project, you need namespacing, explicit exports/imports, and dependency management. The "Future Features" section acknowledges this, but shipping without it limits the language to small, single-file scripts—not the "enterprise" orchestration scenarios implied by the design.

5. **Weak Collection Type Safety**
   - Arrays accept mixed types (`[1, 2, "mixed", true]`) with no way to express homogeneous collections. Objects have no schema enforcement. When parsing LLM responses with `parse_json()`, there's no validation that the returned structure matches expectations. In TypeScript, we'd use discriminated unions, type guards, and schema validation—here you're writing defensive runtime checks everywhere.

---

## Top 5 Questions

1. **How do you envision handling LLM response validation at scale?**
   - Given that LLMs frequently return malformed or unexpected JSON, what patterns does Duso recommend for validating parsed responses? Without static types or a schema validation system, how should enterprise users build robust extraction pipelines that don't fail silently on structural mismatches?

2. **What is the error recovery strategy for conversation state?**
   - If a `conversation.prompt()` call fails mid-way (network error, rate limit, content filter), what happens to the conversation state? Is the failed message added to history? Can you retry? Can you inspect/modify the message history programmatically? This is critical for production agent systems.

3. **How do you prevent scope-related bugs in larger scripts?**
   - The scoping rules where assignment "walks up the scope chain" can cause subtle bugs when a variable name collision occurs across nested functions. Have you considered a linting mode or static analysis tool that warns about potential scope conflicts? What's the recommended practice for teams working on shared orchestration scripts?

4. **What's the strategy for testing Duso scripts?**
   - There's no mention of testing infrastructure—no assertions, no mocking capabilities, no test runner. For enterprise adoption, how would teams write unit tests for their orchestration logic? Can host applications inject mock `claude()` functions? Is there a recommended pattern for dependency injection?

5. **How does the language handle streaming LLM responses?**
   - Modern LLM integrations often require streaming for UX and efficiency. The current `claude()` and `conversation.prompt()` appear to be blocking calls that return complete strings. Is there a plan for streaming support? How would the language model handle incremental token processing given its synchronous execution model?