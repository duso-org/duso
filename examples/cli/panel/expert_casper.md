# Duso Language Evaluation

## Top 5 Pros

1. **Excellent LLM/Template Integration**
   - The `{{expr}}` template syntax with multiline triple-quoted strings is genuinely well-designed for LLM prompt engineering. No escaping JSON quotes, clean embedding of variables, and the ability to include complex expressions makes this far more ergonomic than string concatenation in most languages. The `parse_json`/`format_json` built-ins complete the picture for round-tripping LLM responses.

2. **Pragmatic Object-as-Constructor Pattern**
   - Using objects as blueprints with `Config(timeout = 60)` syntax is clever and lightweight. It provides prototype-based inheritance without the complexity of class hierarchies. The implicit `this` scoping for methods (accessing properties directly without `self.` or `this.`) reduces boilerplate while remaining readable.

3. **Clean Host Integration Model**
   - The explicit separation between core language and host-provided functions (`claude()`, `conversation()`, `load()`, `save()`) is architecturally sound. This allows embedding Duso in different contexts without language changes. The `conversation()` object with `.prompt()` method elegantly handles stateful multi-turn LLM interactions.

4. **Sensible Scoping with Explicit `var`**
   - The scoping model strikes a good balance: assignments walk up the scope chain by default (useful for closures), but `var` creates explicit locals. This is more intuitive than JavaScript's historical `var`/`let` confusion and avoids Python's `global`/`nonlocal` keyword proliferation.

5. **Batteries Included for Agent Work**
   - The built-in function set (`split`, `join`, `parse_json`, `format_json`, `now`, `format_time`, `parse_time`, `range`, `sort` with custom comparators) covers 90% of what agent orchestration scripts actually need without external dependencies. The `contains()` and `replace()` functions with case-insensitivity defaults are practical for fuzzy text matching.

---

## Top 5 Cons

1. **No First-Class Error Values or Result Types**
   - The `try/catch` model with string error messages is limiting for enterprise use. There's no way to programmatically distinguish error types, no stack traces, and no structured error objects. When an LLM call fails, you can't tell if it's a timeout, rate limit, authentication failure, or content policy violation without parsing error strings.

2. **Missing Critical Collection Operations**
   - No `map`, `filter`, `reduce`, or `find` for arrays. Agent orchestration frequently needs to transform collections (e.g., filter valid tool results, map responses to structured objects). The workaround requires verbose `for` loops and manual array building with `append()`. This is a significant productivity gap.

3. **Weak Type Safety for Object Shapes**
   - Objects have no schema validation. When calling `Config(typo_field = 60)`, you silently get an extra field rather than an error. For agent orchestration where you're constructing tool invocations or API payloads, this leads to runtime failures that could be caught earlier. No way to define required fields or field types.

4. **No Async/Await or Parallel Primitives**
   - While the spec says "host handles parallelism," this severely limits script expressiveness. Common agent patterns like "fan out to 3 experts, aggregate responses" require host-specific functions. Scripts can't express parallel intent, making them host-dependent and harder to test in isolation. Even a simple `parallel()` or `await_all()` primitive would help.

5. **Limited String Manipulation**
   - No regex support, no `startswith`/`endswith`, no `indexOf`/`lastIndexOf`, no `padLeft`/`padRight`, no string formatting beyond templates. When parsing LLM responses that don't follow exact JSON structure (common), you're left with `contains()` and `split()` which are insufficient for robust extraction. Agent scripts frequently need pattern matching.

---

## Top 5 Questions

1. **How does error handling work across host function boundaries?**
   - If `claude()` fails mid-conversation, what error information is available? Can the script inspect HTTP status codes, retry-after headers, or token usage? The spec shows string errors in catch blocks, but enterprise LLM orchestration needs structured error metadata to implement proper retry logic with exponential backoff.

2. **What is the memory model for long-running agent loops?**
   - If a script runs a `while true` loop processing events, do closures and abandoned objects get garbage collected? The spec mentions "tree-walking interpreter" but says nothing about memory management. For agent processes that run for hours, memory leaks would be critical failures.

3. **How do you handle partial/streaming LLM responses?**
   - The `claude()` and `conversation.prompt()` functions appear to be blocking and return complete strings. For user-facing agents, streaming responses (token by token) is essential for perceived performance. Is there a callback or generator pattern possible, or does this require a different host function design?

4. **What happens when object methods modify the object's properties?**
   - The spec shows methods accessing properties implicitly, but can they modify them? If I have `agent.learn = function() skill = skill + 1 end`, does calling `agent.learn()` actually increment `agent.skill`? The scoping rules suggest it would modify a captured variable, but the spec doesn't clarify method-to-object property binding.

5. **How would you implement tool/function calling patterns for LLMs?**
   - Modern LLM APIs support structured tool definitions and function calling. The spec shows free-form prompts but no mechanism for defining tool schemas, handling tool invocation requests from the LLM, or validating tool arguments. Is this expected to be entirely host-provided, or are there patterns for expressing tools in Duso itself?