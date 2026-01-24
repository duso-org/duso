# Duso Language Evaluation

*From the perspective of Dev, a Junior Developer*

---

## Top 5 Pros

1. **Incredibly Readable Syntax for LLM Work**
   - The triple-quoted multiline strings with `{{template}}` syntax are a game-changer for prompt engineering. No more escaping quotes in JSON or wrestling with string concatenation. I can just write my prompts naturally:
   ```duso
   prompt = """
   Hello {{name}}!
   Return JSON: {"result": "{{expected}}"}
   """
   ```
   This alone would save me hours compared to doing this in Python or JavaScript.

2. **Objects-as-Constructors Pattern is Elegant**
   - Coming from classes in other languages, the idea that any object can be called like a constructor with overrides is surprisingly intuitive:
   ```duso
   Config = {timeout = 30, retries = 3}
   myConfig = Config(timeout = 60)
   ```
   No boilerplate class definitions, no `__init__` methods—just simple, readable code. It feels like the "just enough OOP" I actually need.

3. **Parallel Execution with Parent Scope Access**
   - The `parallel()` function is exactly what I'd want for making concurrent API calls. Being able to reference parent variables (read-only) inside parallel blocks while keeping each execution isolated is a smart design choice that prevents race condition bugs while remaining useful.

4. **Built-in JSON and Claude Integration**
   - Having `parse_json()` and `format_json()` as first-class citizens, plus the `claude()` and `conversation()` functions, means I can go from idea to working agent script incredibly fast. No importing libraries, no setup—it just works.

5. **Simple Scoping with Explicit `var` Keyword**
   - I really appreciate that the scoping rules are straightforward: assignments without `var` reach up the scope chain, and `var` creates a new local. Coming from JavaScript's confusing `var`/`let`/`const` situation, this feels refreshingly simple while still giving me control.

---

## Top 5 Cons

1. **No Varargs Support Yet**
   - The spec mentions varargs as a "future feature," but for someone building agents that might need flexible function signatures, this feels like a gap. I can work around it with arrays, but `function log(...messages)` would be much cleaner.

2. **Tree-Walking Interpreter Performance Concerns**
   - The spec explicitly says it's "suitable for scripts up to several seconds of execution." As a junior dev, I'm worried about hitting performance walls if I need to process larger datasets or run longer workflows. The suggestion to "use Go functions for performance-critical paths" assumes I know Go well enough to extend the runtime.

3. **Error Handling Could Be Richer**
   - Currently, `catch (error)` gives me a string message. I'd love to have structured error objects with properties like `error.type`, `error.line`, or `error.stack` for debugging. When something goes wrong in a complex agent workflow, a simple string might not give me enough context.

4. **No Built-in HTTP/Network Functions in Core**
   - For agent orchestration, I'd expect basic HTTP capabilities in the core language, but `http` is relegated to `stdlib` modules. This means I'm dependent on what the CLI provides, and the spec mentions different hosts "may provide different implementations or none at all." That uncertainty makes me nervous about portability.

5. **Limited Collection Methods**
   - While `map`, `filter`, and `reduce` exist, I miss conveniences like `find`, `some`, `every`, `slice`, or `reverse` for arrays. For object iteration, I have to use `for key in obj` and manually access values—a `pairs()` function (mentioned as deferred) would help a lot.

---

## Top 5 Questions

1. **How does the module caching work with mutable state?**
   - If I `require("mymodule")` and the module returns an object, and then I modify that object, will subsequent `require()` calls in other parts of my code see those modifications? This could lead to subtle bugs if shared state isn't clearly documented.

2. **What happens when `parallel()` functions need to return errors?**
   - The spec says "If a function errors, that result becomes `nil`." But how do I distinguish between a function that legitimately returned `