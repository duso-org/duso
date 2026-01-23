# Duso Language Evaluation

## Top 5 Pros

1. **Purpose-Built String Templating for LLM Workflows**
   - The `{{expression}}` template syntax with triple-quoted multiline strings is exceptionally well-designed for prompt engineering. Being able to embed JSON directly without escaping quotes (`"""{"type": "{{type}}"}"""`) dramatically reduces friction when constructing LLM prompts and parsing responses. This is a genuine quality-of-life improvement over most general-purpose languages.

2. **Objects-as-Constructors Pattern is Elegant**
   - The ability to call objects like functions to create copies with overrides (`Config(timeout = 60)`) provides lightweight, composition-friendly OOP without classes. Combined with implicit `self` access in methods, this creates a clean pattern for agent blueprints and configuration objects that feels natural for orchestration use cases.

3. **Pragmatic Host Integration Model**
   - The design philosophy of keeping Duso single-threaded while delegating parallelism to the Go host is architecturally sound. This separation of concerns keeps the scripting layer simple and predictable while allowing sophisticated concurrent orchestration at the application level. The `conversation()` stateful wrapper is a good example of this principle in action.

4. **Sensible Defaults for LLM Integration**
   - Built-in `parse_json()`/`format_json()`, case-insensitive `contains()`/`replace()` by default, and automatic string coercion with `+` are all pragmatic choices for text-heavy LLM workflows. The type coercion rules (especially numeric string comparisons) reduce boilerplate when processing LLM outputs.

5. **Low Cognitive Load Syntax**
   - The Lua-inspired `then`/`do`/`end` block delimiters, combined with familiar operators and control flow, make the language immediately readable. The distinction between colons in object literals vs. equals in function calls is well-reasoned and aids comprehension.

---

## Top 5 Cons

1. **Scope Resolution Rules Are Error-Prone**
   - The rule that assignment without `var` walks up the scope chain to find existing variables (or creates a local if none found) is a footgun. A typo in a variable name silently creates a new local instead of modifying the intended outer variable. This is the opposite of Python's explicit `global`/`nonlocal` and worse than JavaScript's strict mode behavior. Consider requiring explicit declaration for all variables.

2. **No Native Async/Await or Parallel Primitives**
   - While delegating parallelism to the host is principled, it means common patterns like "fan out to 5 LLM calls, collect results" require custom host functions for every parallel scenario. A simple `parallel([fn1, fn2, fn3])` built-in that the host could implement would enable more expressive orchestration without complicating the execution model.

3. **Limited Error Context and Stack Traces**
   - The specification shows basic error messages like `"array index out of bounds"` but doesn't mention line numbers, stack traces, or structured error objects. For debugging multi-step agent workflows, knowing *where* in the script an error occurred is critical. The `catch (error)` binding as a plain string loses valuable context.

4. **No Module/Import System**
   - The `include()` function executes scripts in the current environment, which is basically `eval()` for files. There's no namespacing, no way to selectively import, and no protection against naming collisions. For any non-trivial agent library, this becomes unmanageable quickly. This is acknowledged as a "future feature" but is a significant gap.

5. **Inconsistent Collection Semantics**
   - Arrays are 0-indexed but `for i = 1, 10` is 1-based inclusive. `append()` returns a new array (immutable style) but object property assignment mutates in place. `for key in object` iterates keys (not entries), requiring separate `obj[key]` access. These inconsistencies add friction when working with data structures.

---

## Top 5 Questions

1. **How are host-provided functions registered, and what's the error contract?**
   - The spec shows `claude()` and `conversation()` as CLI-provided, but doesn't explain how custom Go functions are bound, what happens when they return errors, or whether they can return multiple values. Can host functions yield/suspend execution, or must they block synchronously?

2. **What happens when a conversation object is copied or passed between functions?**
   - Given that `conversation()` maintains state, what are the semantics of `conv2 = conv1`? Is it a reference (shared state) or a copy? Can conversations be serialized/restored for long-running workflows? This is critical for multi-agent patterns.

3. **How does the interpreter handle resource limits and infinite loops?**
   - The spec mentions "host application sets execution timeouts" but provides no mechanism for scripts to cooperate with cancellation. Is there a context/deadline propagated? Can a script check if cancellation was requested? What happens to partial state on timeout?

4. **What are the guarantees around object key ordering?**
   - The spec shows `keys(obj)` returning `[a b c]` but doesn't specify if order is insertion-order, alphabetical, or undefined. For reproducible prompts and deterministic JSON output, this matters significantly. Go maps are unordered by default.

5. **How do closures interact with object method `this` binding?**
   - The spec says methods have "implicit variable lookup" for object properties, but what happens when a method is extracted and called standalone (`fn = obj.method; fn()`)? Is `this` bound at definition or call time? What about methods that return closures referencing the object's properties?