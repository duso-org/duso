# Duso Language Evaluation

*Evaluated by Petra, Python Expert with Data Science Focus*

---

## Top 5 Pros

1. **Excellent String Templating for LLM Workflows**
   - The `{{expr}}` template syntax is intuitive and powerful for constructing prompts
   - Triple-quoted multiline strings (`"""..."""`) eliminate escaping headaches when embedding JSON schemas or code examples
   - This is a significant ergonomic win over Python's f-strings or `.format()` when building complex prompts with nested structures

2. **Objects-as-Constructors Pattern is Elegant**
   - The ability to call objects like functions (`Config(timeout = 60)`) provides prototype-based inheritance without class keyword complexity
   - This feels natural for configuration-heavy agent workflows where you're constantly creating variations of base configurations
   - Cleaner than Python's `dataclasses` or `dict.copy()` + update patterns for simple cases

3. **Sensible Dynamic Typing with Predictable Coercion**
   - The truthiness rules are well-defined (empty arrays/objects are falsy, unlike JavaScript)
   - String concatenation auto-coercion (`"Value=" + 42`) reduces boilerplate
   - The comparison coercion (string "10" < 15 works) is pragmatic for handling parsed data—something Python explicitly rejects but causes friction in data pipelines

4. **Built-in Parallel Execution with Sane Semantics**
   - `parallel()` with read-only parent scope access is a smart design choice
   - Prevents the race condition footguns that plague concurrent code while enabling the primary use case (multiple independent API calls)
   - Array and object forms give flexibility without complexity—this is better thought-out than Python's `asyncio.gather()` for simple cases

5. **Self-Contained Binary Philosophy**
   - Freezing stdlib/contrib into the binary eliminates dependency hell
   - "Archive your scripts and binary together for permanent reproducibility" is a refreshing stance
   - For agent orchestration (where scripts may run for years), this is more valuable than having the latest packages

---

## Top 5 Cons

1. **No First-Class Error Types or Error Context**
   - `catch (error)` only provides a string message—no stack traces, error types, or structured error data
   - In agent workflows where you need to distinguish "rate limit" from "invalid API key" from "network timeout," string parsing is fragile
   - Python's exception hierarchy with `except APIError as e:` and `e.status_code` is significantly more practical for production code

2. **Limited Data Transformation Primitives**
   - No list comprehensions, generator expressions, or slice syntax (`arr[1:3]`)
   - `map`/`filter`/`reduce` exist but require verbose anonymous functions
   - For data science workflows, `[x*2 for x in data if x > 0]` is dramatically more readable than nested `map(filter(...), ...)`

3. **Implicit Scope Mutation is Dangerous**
   - Assignment without `var` walks up the scope chain and modifies parent variables
   - This is the opposite of Python's explicit `global`/`nonlocal` declarations
   - In larger scripts, accidentally mutating a parent scope variable because you forgot `var` will cause subtle bugs that are hard to trace

4. **No Native Async/Await or Streaming Support**
   - LLM responses often stream tokens; there's no obvious way to handle streaming data
   - `parallel()` only handles "fire all, wait for all" patterns
   - Agent orchestration increasingly needs real-time token streaming, cancellation, and progress callbacks—none addressed here

5. **Weak Type Introspection and Validation**
   - `type()` returns strings, but there's no schema validation, type guards, or assertions
   - When parsing LLM JSON responses, you need to manually check every field exists and has the right type
   - Python's `pydantic` or TypeScript's Zod patterns for validating LLM output are missing, and this is a critical gap for production agents

---

## Top 5 Questions

1. **How does error handling work across `parallel()` boundaries?**
   - The spec says "If a function errors, that result becomes `nil`"
   - But how do I access *what* error occurred? Is the error message lost entirely?
   - For debugging parallel