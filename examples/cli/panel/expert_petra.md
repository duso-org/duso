# Evaluation of Duso Language Specification

*From the perspective of Petra, a Python expert with data science and dynamic typing background*

---

## Top 5 Pros

1. **Excellent String Templating for LLM Workflows**
   - The `{{expr}}` template syntax is incredibly intuitive and powerful for constructing prompts. Triple-quoted multiline strings with embedded expressions solve a real pain point when working with LLMs—no more awkward f-string escaping or concatenation chains. The fact that JSON braces don't need escaping is a thoughtful design choice that will save countless debugging hours in agent orchestration scenarios.

2. **Pythonic Dynamic Typing with Sensible Coercion**
   - The type system feels familiar to Python developers: loose typing, truthiness rules (empty collections are falsy), and implicit string coercion. The coercion rules are well-documented and predictable—particularly the comparison coercion that allows `5 < "10"` to work naturally. This reduces boilerplate when processing LLM outputs that often return strings representing numbers.

3. **Objects-as-Constructors Pattern is Elegant**
   - The ability to use any object as a blueprint via `Config(timeout=60)` syntax provides lightweight prototypal inheritance without the complexity of class hierarchies. This is perfect for configuration-heavy agent workflows where you want sensible defaults with easy overrides. It's simpler than Python's dataclasses while achieving similar ergonomics.

4. **First-Class Conversation State Management**
   - The `conversation()` function with maintained context across `.prompt()` calls directly addresses a core challenge in agent orchestration. This is a thoughtful abstraction that would require custom wrapper classes in Python. Having it as a language-level primitive significantly reduces boilerplate for multi-turn agent interactions.

5. **Clean Scope Rules with Explicit `var` Keyword**
   - The scoping model strikes a good balance: assignments reach up the scope chain by default (useful for closures and state modification), but `var` provides explicit local variable creation. This is more intentional than JavaScript's historical mess and clearer than Python's `global`/`nonlocal` keywords. The automatic locality of loop variables prevents a common bug class.

---

## Top 5 Cons

1. **No List Comprehensions or Functional Array Methods**
   - Coming from Python/data science, the absence of `map()`, `filter()`, list comprehensions, or lambda expressions is painful. Processing LLM outputs often involves transforming arrays of results. Currently, this requires verbose for-loops:
   ```duso
   // Python: results = [x.name for x in items if x.score > 0.8]
   // Duso requires:
   results = []
   for item in items do
     if item.score > 0.8 then
       results = append(results, item.name)
     end
   end
   ```
   This verbosity will hurt adoption among data-oriented users.

2. **Immutable Array `append()` Returns New Array**
   - The `append(arr, value)` pattern that returns a new array (requiring `arr = append(arr, value)`) is unintuitive and error-prone. It violates expectations from Python where `list.append()` mutates in place. This will cause silent bugs when users forget the reassignment, especially in loops building up results.

3. **No Dictionary/Object Comprehension or Iteration with Values**
   - Iterating over objects only yields keys, requiring manual value lookup:
   ```duso
   for key in config do
     value = config[key]  // Extra step every time
   end
   ```
   There's no `items()` equivalent noted (though `pairs()` is listed as future). For data processing tasks, this adds friction compared to Python's `for k, v in dict.items()`.

4. **Limited Error Handling—No Custom Exceptions or Error Types**
   - The `catch (error)` only provides a string message. There's no way to distinguish error types programmatically, create custom errors, or re-throw with context. For agent orchestration where you might want to retry on network errors but fail on validation errors, this is limiting:
   ```duso
   // Can't do this:
   catch (e)
     if e.type == "network" then retry() end
     if e.type == "validation" then exit(e) end
   end
   ```

5. **Single-Threaded with No Async/Await or Promises**
   - While the spec justifies this as "host handles parallelism," it pushes significant complexity onto Go developers. Agent orchestration often involves embarrassingly parallel operations (multiple LLM calls, tool invocations). Requiring users to write custom Go functions for `fetch_all_tools()` rather than expressing parallelism in Duso limits the language's standalone utility and increases the learning curve for the intended use case.

---

## Top 5 Questions

1. **How does garbage collection work, and what are the memory characteristics for long-running agent processes?**
   - Agent orchestration scripts may run for extended periods, accumulating conversation history and intermediate results. Does Duso rely entirely on Go's GC? Are there any mechanisms to explicitly release large objects (like conversation histories) or monitor memory usage from within scripts?

2. **What happens when `parse_json()` encounters LLM outputs with markdown code fences or partial JSON?**
   - LLMs frequently wrap JSON in ```json blocks or produce incomplete responses. Does `parse_json()` handle common edge cases like stripping markdown, or will users need to preprocess strings? How are parsing errors surfaced—is the error message detailed enough to debug malformed responses?

3. **How do the Claude integration functions handle rate limiting, retries, and streaming responses?**
   - The spec shows `claude()` and `conversation()` but doesn't address:
   - What happens on API rate limits—automatic retry with backoff, or immediate error?
   - Can responses be streamed for long generations, or must the script block until completion?
   - How are API errors (authentication, quota exceeded) distinguished from content errors?

4. **Can objects have computed properties or getters, and how do method calls resolve `this`/`self`?**
   - The spec shows methods accessing properties via implicit scope (`name` instead of `self.name`). What happens if a local variable shadows an object property? Is there any way to explicitly reference the containing object? Can properties be dynamically computed on access, or are they always stored values?

5. **What is the story for testing, debugging, and development tooling?**
   - For serious adoption in agent orchestration, developers need:
   - A REPL for interactive development
   - Debugger or at least stack traces with line numbers
   - Unit testing framework or conventions
   - IDE support (syntax highlighting, basic completion)
   
   Are any of these available or planned? The spec mentions only `type()` for debugging—is there a way to inspect the current scope or call stack?