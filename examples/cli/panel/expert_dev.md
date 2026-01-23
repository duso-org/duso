# Duso Language Evaluation

## My Top 5 Pros

1. **String Templates are Perfect for LLM Work**
   - The `{{expression}}` syntax is incredibly intuitive for building prompts
   - Triple-quoted multiline strings mean I don't have to escape JSON or code blocks
   - This alone would save me hours of frustration when working with AI APIs

2. **Objects-as-Constructors is Brilliantly Simple**
   - Coming from JavaScript, the pattern `Config = {timeout: 30}` then `config = Config(timeout: 60)` feels natural
   - No need to learn class syntax, `new` keywords, or prototype chains
   - I can create blueprints and instances without any OOP ceremony

3. **The Built-in Claude Integration is Thoughtfully Designed**
   - Having `claude()` for one-shots and `conversation()` for stateful chats covers the main use cases
   - The `conversation.prompt()` pattern makes multi-turn agents readable
   - Combined with `parse_json()`, I can go from LLM response to usable data in one line

4. **Explicit Scoping with `var` Prevents Footgun Bugs**
   - As someone still learning, accidentally mutating outer scope variables is a real problem
   - The rule is simple: use `var` = definitely local, no `var` = might modify outer scope
   - For loop variables being automatically local is a nice safety net

5. **Error Handling is Straightforward**
   - `try/catch` works exactly how I'd expect from JavaScript
   - Errors propagate up unless caught - no weird exception hierarchies to learn
   - The error messages in the spec examples are actually helpful and readable

---

## My Top 5 Cons

1. **No Import/Module System is a Major Gap**
   - For anything beyond a single-file script, I'll hit a wall
   - `include()` exists but it's CLI-provided, not core language
   - Can't share utility functions across projects cleanly without copy-paste

2. **Single-Threaded with No Async/Await Pattern**
   - If I'm making multiple LLM calls, I can't parallelize them from the script
   - Having to rely on "host-provided parallelism" feels like a workaround
   - The `fetch_all_tools()` example in the spec looks magical - how do I actually build that?

3. **The Colon vs Equals Distinction Will Trip Me Up**
   - `{key: value}` in objects but `foo(key = value)` in function calls
   - The spec says it's "intentional" but I know I'll mix these up constantly
   - Especially when objects can be called like functions with `Config(timeout = 60)`

4. **Limited Debugging Capabilities**
   - Only `print()` and `type()` for debugging
   - No stack traces mentioned in the spec
   - No way to inspect the current scope or environment that I can see

5. **Array/Object Methods Feel Incomplete**
   - No `map`, `filter`, `reduce`, or `find` for arrays
   - No way to delete a key from an object
   - `append()` returns a new array instead of modifying in place - is that intentional? Will I forget and create bugs?

---

## My Top 5 Questions

1. **How Do I Handle Nested JSON from LLM Responses?**
   - If Claude returns `{"data": {"nested": {"deep": "value"}}}`, can I access `result.data.nested.deep` directly?
   - What happens if a key doesn't exist - do I get `nil` or an error?
   - Is there a safe navigation operator like `result?.data?.nested`?

2. **What Happens When an LLM Call Fails or Times Out?**
   - Does `claude()` throw an error I can catch, or return `nil`?
   - How do I set timeouts for individual LLM calls vs relying on host timeouts?
   - Can I retry a failed call, and how would I implement exponential backoff?

3. **Can I Build Reusable Agent Components?**
   - If I create a useful `conversation()` pattern, how do I share it across scripts?
   - Can objects with methods be serialized/deserialized somehow?
   - Is there a way to create a "library" of prompts and agent configurations?

4. **How Does Memory/Performance Scale?**
   - The spec says "suitable for scripts up to several seconds" - what's the actual limit?
   - If I'm processing a large JSON response (thousands of items), will I hit issues?
   - Does the tree-walking interpreter have any gotchas for recursive functions?

5. **What's the Story for Testing My Scripts?**
   - Is there an assert function or testing framework?
   - How do I mock `claude()` calls during development without burning API credits?
   - Can I run scripts in a "dry run" mode to check for syntax errors before execution?