# Duso Language Evaluation

## Top 5 Pros

1. **Excellent LLM/Template Integration** - The `{{expr}}` template syntax with multiline triple-quoted strings is exceptionally well-designed for LLM prompt engineering. The ability to embed arbitrary expressions (variables, function calls, arithmetic) directly in strings without escaping JSON or code is a major productivity win for agent workflows.

2. **Zero External Dependencies** - Building entirely on Go's standard library with modules baked into the binary is a strong architectural choice. This eliminates version conflicts, ensures reproducibility, and makes deployment trivial. The "freeze at release" philosophy means scripts written today will work identically years from now.

3. **Intuitive Object-as-Constructor Pattern** - The ability to use objects as blueprints (`Config = {timeout = 30}; config = Config(timeout = 60)`) provides lightweight OOP without classes. Combined with methods that implicitly access object properties, this enables clean agent/entity modeling without boilerplate.

4. **Clean Host Integration Model** - The design explicitly separates concerns: Duso handles orchestration logic sequentially while the host Go application manages parallelism, timeouts, and system resources. The `parallel()` function with read-only parent scope access is a particularly elegant solution for concurrent API calls.

5. **Thoughtful Scoping with `var` Keyword** - The explicit `var` for local variable creation versus implicit outer-scope modification gives developers control without complexity. This prevents accidental mutations while keeping simple scripts clean—a good balance for scripting.

---

## Top 5 Cons

1. **Inconsistent Error Handling Semantics** - The `parallel()` function silently converts errors to `nil`, which can mask failures. This differs from normal try/catch behavior and requires manual null-checking. A more explicit error handling strategy (error objects, result tuples) would be safer for production agent workflows.

2. **Limited Data Structure Operations** - No native support for sets, queues, or ordered maps. No `delete` operation for removing object keys or array elements. For complex agent state management, these omissions require verbose workarounds or filtering/rebuilding entire structures.

3. **Weak Type Safety for Critical Operations** - Implicit type coercion in comparisons (`5 < "10"` works) can hide bugs. For agent orchestration where data flows through multiple LLM calls and JSON parsing, silent coercion failures could produce subtle, hard-to-debug issues in production.

4. **No Async/Await or Coroutine Support** - While the host-provided parallelism model works, the language lacks any primitives for streaming responses, cancellation tokens, or cooperative multitasking. Modern LLM APIs increasingly support streaming, which this design can't elegantly express.

5. **Module System Limitations** - The `include()` function pollutes the caller's namespace, and there's no way to selectively import specific functions from a module. No versioning mechanism means stdlib/contrib changes between binary versions could break scripts silently.

---

## Top 5 Questions

1. **How does error context propagate through nested function calls?** - The spec shows basic try/catch with string error messages, but how do you get stack traces, error codes, or structured error data? For agent debugging, knowing *where* in a multi-step workflow something failed is critical.

2. **What happens when `parallel()` functions need to communicate or share state?** - The spec mentions parent scope is read-only, but what about coordination patterns like early termination (one function finds an answer, others should stop) or aggregating partial results as they complete?

3. **How are large/streaming LLM responses handled?** - The `claude()` and `conversation()` functions appear to return complete strings. How would you process a 50KB response incrementally, or implement "stop generating" based on partial content? Is there a callback or iterator pattern?

4. **What's the memory model for long-running agent loops?** - If an agent runs continuously (e.g., monitoring loop), how is garbage collection handled? Can closures cause memory leaks? Are there limits on array/object sizes or recursion depth?

5. **How do you test Duso scripts in isolation from LLM APIs?** - The spec shows tight coupling to `claude()` functions. Is there a mocking/stubbing mechanism for unit testing agent logic without making actual API calls? Can you inject test doubles for host-provided functions?