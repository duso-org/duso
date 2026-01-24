# Evaluation of Duso Language Specification

## Top 5 Pros

1. **Excellent LLM/Template Integration** - The `{{expr}}` template syntax with triple-quoted multiline strings is brilliantly designed for LLM prompts and JSON generation. No escaping needed for quotes inside multiline strings, and expressions evaluate inline. This solves a real pain point in agent scripting where you're constantly building prompts with interpolated values.

2. **Objects-as-Constructors Pattern** - The ability to call any object like a function to create copies with overrides (`Config(timeout = 60)`) is elegant and practical. It provides lightweight OOP without classes, keywords like `new`, or inheritance complexity. Combined with implicit `self` access in methods, it hits a sweet spot for configuration and agent blueprints.

3. **Clean Go Integration Story** - No external dependencies, tree-walking interpreter, and explicit host-provided functions (like `claude()`, `load()`, `save()`) make embedding straightforward. The frozen binary approach with baked-in stdlib/contrib modules means true reproducibility—archive script + binary and it works forever.

4. **Sensible Scoping with `var`** - The opt-in `var` keyword for explicit locals while defaulting to scope chain lookup is pragmatic. It allows quick scripts without ceremony while giving control when needed. Loop variables being implicit locals prevents a common bug class.

5. **Parallel Execution Primitive** - The `parallel()` function with read-only parent scope access is well-designed for agent orchestration. It enables concurrent LLM calls without introducing shared mutable state complexity. The array/object result preservation is thoughtful.

## Top 5 Cons

1. **No First-Class Error Values** - Only try/catch with string error messages. No way to return errors as values, check error types, or create custom error objects. For agent orchestration where partial failures are common (one API call fails, others succeed), this limits graceful degradation patterns.

2. **Limited Data Structure Operations** - No slice syntax for arrays (`arr[1:3]`), no spread operator, no destructuring assignment, no `in` operator for membership testing outside loops. Common operations like "get last 3 items" or "merge two objects" require verbose workarounds.

3. **Weak Type Introspection** - Only `type()` returning strings. No `instanceof` equivalent, no way to check if an object has a key without accessing it and catching errors, no schema validation primitives. For processing varied LLM JSON responses, you often need defensive type checking.

4. **No Async/Await or Promises** - While `parallel()` handles concurrent execution, there's no way to express "do X, then when done do Y with the result" chains cleanly. Sequential code with `parallel()` works, but complex orchestration flows (retries, timeouts, conditional branching on async results) get awkward.

5. **String-Only Error Context** - Catch blocks receive just an error string. No stack traces, no error codes, no structured error objects with metadata. Debugging agent failures ("which API call failed? with what parameters?") requires manual logging discipline.

## Top 5 Questions

1. **How does `parallel()` handle timeouts per-function?** - The spec mentions host-level timeouts, but can individual parallel branches have different timeout limits? For agent orchestration, a web scrape might need 30s while an LLM call needs 120s. Is this configurable per-function or only globally?

2. **What happens when object method recursively references itself?** - With implicit property lookup in methods, can a method call itself? If `obj.process` references `process` inside, does it find the method or require `obj.process()`? This affects recursive agent patterns.

3. **How are circular object references handled in `format_json()`?** - If object A references object B which references A, does `format_json()` detect this and error, or infinite loop? Agent state objects often have back-references.

4. **Can `require()` modules export functions that modify module-level state?** - The spec shows returning objects with functions, but can those functions maintain private module state (like connection pools, caches)? Is the module's closure environment preserved across calls?

5. **What's the memory model for large `parallel()` result sets?** - If 100 parallel functions each return 1MB of data, is all 100MB held in memory simultaneously? For agent orchestration processing many documents, this could matter.