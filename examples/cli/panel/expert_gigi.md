# Duso Language Evaluation

## Top 5 Pros

1. **Pragmatic LLM-First Design** — The `{{expr}}` template syntax and triple-quoted strings are genuinely useful for the stated purpose. Building JSON payloads and prompts without escaping nightmares is a real win. The `parse_json`/`format_json` builtins show someone actually thought about the workflow.

2. **Objects-as-Constructors is Elegant** — The pattern of calling objects to create copies with overrides (`Config(timeout = 60)`) is simple and avoids the typical class/prototype machinery. It's lightweight composition that fits the "scripts, not applications" use case well.

3. **Sensible Scope Defaults** — The `var` keyword for explicit locals while defaulting to lexical lookup is a reasonable middle ground. Most scripting languages get this wrong in confusing ways. The closure semantics are clean.

4. **Honest About What It's Not** — "Tree-walking interpreter, no optimization, suitable for scripts up to several seconds" — this is refreshing. The explicit deferral of parallelism to the host Go application is the right call. Don't build what you don't need.

5. **Host Integration Philosophy** — Keeping `load`/`save`/`claude` as CLI-provided rather than core language features is correct. Different embedders have different needs. The contract is clear: Duso orchestrates, Go does the heavy lifting.

## Top 5 Cons

1. **Implicit Method Binding is Magic** — When `agent.greet()` magically accesses `name` without `self.` or explicit binding, you've hidden important information. What happens if `name` exists in an outer scope? The spec doesn't clarify precedence. This will cause debugging nightmares.

2. **Mixed Metaphors in Named Arguments** — `Config(timeout = 60)` for object construction uses `=`, but `{timeout: 60}` uses `:`. The spec *explains* this but the explanation ("are you creating data, or assigning to parameters?") doesn't hold — you're doing both when calling an object-as-constructor. Pick one syntax.

3. **Truthiness Rules Are a Minefield** — Empty array `[]` is falsy but `[0]` containing a falsy value is truthy. Empty object `{}` is falsy. `"0"` is truthy but `0` is falsy. This is JavaScript-tier confusion. Go's explicit `len(x) == 0` is clearer.

4. **No Clear Error Model** — Errors are strings caught in `catch(error)`. No error types, no stack traces mentioned, no way to rethrow or wrap. For "agent orchestration" where you're calling external services, you need better error context than string matching.

5. **Iteration Inconsistency** — `for item in array` gives values, `for key in object` gives keys. Why not values? If I want key-value pairs from objects, I need `keys()` then bracket access. The asymmetry is arbitrary.

## Top 5 Questions

1. **What's the method resolution order?** — When calling `agent.greet()` where `greet` references `name`, and `name` exists both as an object property AND in an enclosing scope, which wins? The spec says "as if they were in its scope" but doesn't define precedence. This is critical.

2. **How do host functions signal errors?** — The spec mentions Go function integration but not how a Go function returns an error to Duso. Does it panic? Return a special value? Trigger the try/catch mechanism? This is the primary integration point.

3. **What happens to conversation state on error?** — If `analyst.prompt()` fails mid-conversation (network timeout, rate limit), is the conversation object still usable? Can you retry? Is the failed message in the history? For agent workflows, this matters enormously.

4. **Why no `self` or `this`?** — The implicit property access in methods means you can't pass a method as a callback and have it work. `callback = agent.greet; callback("Hi")` — does `name` resolve? Most languages solved this with explicit receivers for good reason.

5. **What's the memory model for large scripts?** — "Scripts up to several seconds" — but agent orchestration often means accumulating context across many LLM calls. Are there limits on string size, array length, object depth? What happens when you hit them? For production use, these bounds matter.