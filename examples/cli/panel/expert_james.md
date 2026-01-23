# Duso Language Evaluation

## Top 5 Pros

1. **Excellent String Templating for LLM Workflows**
   - The `{{expr}}` template syntax with multiline triple-quoted strings is perfectly suited for constructing LLM prompts. Unlike JavaScript template literals, the double-brace syntax avoids conflicts with JSON's curly braces, eliminating the escaping headaches that plague prompt engineering in other languages.

2. **Objects-as-Constructors Pattern is Elegant**
   - The ability to call any object like a function to create copies with overrides (`Config(timeout = 60)`) provides lightweight prototypal inheritance without the complexity of classes. This is cleaner than JavaScript's prototype chain manipulation or constructor functions for simple configuration and blueprint patterns.

3. **First-Class LLM Integration Design**
   - The `conversation()` function maintaining stateful context across calls addresses a real pain point in agent orchestration. The ability to chain `.prompt()` calls with preserved history is exactly what multi-step agent workflows need, something that requires significant boilerplate in JavaScript.

4. **Sensible Scoping with Explicit `var`**
   - The scoping model where assignment without `var` walks up the scope chain (like Python) but `var` creates an explicit local is a reasonable middle ground. It's more predictable than JavaScript's historical `var` hoisting issues while avoiding the verbosity of requiring declarations everywhere.

5. **Clean Try/Catch Without Ceremony**
   - The exception handling syntax is minimal and readable. Combined with the single-threaded, sequential execution model, error handling becomes straightforward to reason about—no async error propagation complexities or unhandled promise rejections to worry about.

---

## Top 5 Cons

1. **No Async/Await or Promises**
   - For a language designed for agent orchestration, the complete absence of async primitives is concerning. While the spec mentions "host-provided parallelism," this pushes significant complexity to Go integrators. JavaScript's async/await has proven essential for I/O-heavy workflows. Expecting every parallel operation to be a synchronous-looking host function creates a leaky abstraction—what happens when a script needs to orchestrate multiple independent LLM calls with different timing?

2. **Implicit Method `this` Binding is Confusing**
   - The claim that methods "automatically have access to the object's properties as if they were in scope" without explicit `self` or `this` is underspecified and potentially dangerous. What happens with nested objects? Name collisions between local variables and object properties? JavaScript's explicit `this` has problems, but implicit binding creates even more ambiguity.

3. **Limited Collection Operations**
   - No `map`, `filter`, `reduce`, or `find` functions for arrays. These are fundamental for data transformation in agent workflows where you're processing LLM responses, filtering results, and transforming data structures. The only iteration options are `for` loops, which will lead to verbose, imperative code where functional approaches would be cleaner.

4. **Weak Type Coercion Edge Cases**
   - String-to-number coercion in comparisons (`5 < "10"` → true) is a footgun. JavaScript's similar behavior has caused countless bugs. The spec even acknowledges `"hello" < 5` throws an error, meaning developers must defensively check types anyway. The "sensible" coercion promise is undermined by these inconsistencies.

5. **No Module System**
   - Listed as a "future feature," but the lack of any import/export mechanism severely limits code organization. The `include()` function just executes code in the current environment, providing no namespace isolation. For any non-trivial agent system with multiple specialists, tools, or prompt libraries, this becomes a maintenance nightmare.

---

## Top 5 Questions

1. **How does method binding actually resolve property lookups?**
   - The spec shows `agent.greet()` accessing `name` and `skill` implicitly. If I have a local variable `name` in the calling scope and an object property `name`, which wins? What's the lookup order? Does the method create a new scope that shadows outer variables with object properties, or vice versa?

2. **What happens when a `conversation()` call exceeds token limits mid-conversation?**
   - The spec shows `.prompt()` preserving context across calls, but LLMs have context windows. Does the conversation object handle truncation? Throw an error? Is there a way to inspect or manage the accumulated context? This is critical for long-running agent workflows.

3. **How are circular references handled in objects?**
   - Can I do `obj = {}; obj.self = obj`? If so, what happens with `format_json(obj)`? What about object-as-constructor calls on objects with circular references? The spec doesn't address this, but it's a common source of infinite loops and stack overflows.

4. **What's the execution model for nested try/catch with host function errors?**
   - If a Go-provided function like `claude()` times out or fails at the network level, does that surface as a catchable Duso error? Can host functions throw structured errors that scripts can inspect beyond just the string message? The `catch (error)` only binds a string—is there no error object with type/code properties?

5. **How does the interpreter handle script resource limits?**
   - The spec mentions "host application sets execution timeouts," but what about memory limits? Infinite loops? If I write `while true do x = append(x, 1) end`, does the host have to implement guards, or is there any built-in protection? For a language designed to run potentially untrusted agent code, sandboxing details matter significantly.