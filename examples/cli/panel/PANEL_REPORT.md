# Duso Expert Panel Review

Generated: 2026-01-22 17:12:33
Experts: 9

---

## Panel Members

- Gigi (Go Expert)
- Petra (Python Expert)
- James (JavaScript Expert)
- Tasha (TypeScript Expert)
- Rust Raj (Rust Expert)
- Luna (Lua Expert)
- Casper (CSharp Expert)
- Ivy (LLM Coding Assistant)
- Dev (Junior Developer)

**Individual analyses saved as:**
- expert_gigi.md
- expert_petra.md
- expert_james.md
- expert_tasha.md
- expert_rust raj.md
- expert_luna.md
- expert_casper.md
- expert_ivy.md
- expert_dev.md

---

## Synthesis

# Duso Language Review Analysis

## 1. Consensus Strengths

### String Templating for LLM Workflows (9/9 experts)
Every reviewer praised the `{{expr}}` template syntax with triple-quoted strings as genuinely well-designed for the target domain. Key points:
- No escaping needed for JSON braces or quotes
- Clean multiline prompt construction
- Eliminates a real pain point in LLM development

### Objects-as-Constructors Pattern (8/9 experts)
The `Config(timeout = 60)` syntax for creating copies with overrides was widely appreciated:
- Lightweight composition without class machinery
- Intuitive for configuration-heavy workflows
- Creates copies, not references (reduces aliasing bugs)

### Pragmatic Host Integration Model (7/9 experts)
The separation between Duso (orchestration) and Go (heavy lifting/parallelism) was seen as architecturally sound:
- Keeps scripting layer simple and predictable
- Clear separation of concerns
- Appropriate for the embedding use case

### First-Class Conversation State Management (6/9 experts)
The `conversation()` API with stateful `.prompt()` calls was recognized as a thoughtful domain-specific abstraction that would require significant boilerplate in other languages.

### Sensible Scoping with Explicit `var` (6/9 experts)
The balance between convenience (scope chain walking) and safety (explicit `var` for locals) was seen as reasonable for scripting, though this was also a point of contention (see concerns).

---

## 2. Consensus Concerns

### No Module/Import System (9/9 experts)
Every single reviewer flagged this as a critical gap:
- `include()` pollutes global namespace
- No namespacing, encapsulation, or dependency management
- Limits language to small, single-file scripts
- Acknowledged as "future feature" but seen as essential for non-trivial projects

### Limited/Missing Collection Operations (8/9 experts)
The absence of `map`, `filter`, `reduce`, `find` was consistently criticized:
- Requires verbose for-loops for common transformations
- Particularly painful for data-oriented LLM response processing
- "Trivial to implement in userland but should be built-in"

### Weak Error Handling Model (8/9 experts)
String-based errors without structure or context troubled most reviewers:
- No way to distinguish error types programmatically
- No stack traces or line numbers mentioned
- No custom error types or error hierarchies
- Critical for retry logic in agent orchestration

### Implicit Method Binding is Underspecified/Fragile (7/9 experts)
The "magic" where methods access object properties without explicit `self`/`this`:
- Resolution order unclear when names collide with outer scope
- Behavior when methods are extracted/passed as callbacks undefined
- Fragile for composition and callback patterns

### No Async/Await or Parallel Primitives (7/9 experts)
While the host-delegation approach was understood, many felt this pushes too much complexity outward:
- Can't express parallel intent in scripts
- Common patterns like "fan out to N experts, aggregate" require custom host functions
- Makes scripts host-dependent and harder to test

### Scope Walking Rules as Footgun (6/9 experts)
Despite praising the `var` keyword, many worried about the default behavior:
- Assignment without `var` walking scope chain can cause subtle bugs
- Typo creates new local instead of modifying intended variable
- "Lua's original sin" that Duso inherits

---

## 3. Key Tensions

### Explicit `var` vs. Mandatory Declarations
| Position | Experts | Rationale |
|----------|---------|-----------|
| Current balance is reasonable | Gigi, Petra, Casper, Dev | Good for scripting terseness; `var` exists when needed |
| Should require explicit declaration | Ivy, Luna, Raj | Default scope walking is error-prone; causes silent bugs |

**Why they disagree:** Scripting-focused reviewers value terseness; safety-focused reviewers prioritize explicit over implicit.

### Host-Delegated Parallelism vs. Language-Level Async
| Position | Experts | Rationale |
|----------|---------|-----------|
| Host delegation is correct | Gigi, Tasha, Raj, Ivy | Keeps language simple; avoids async complexity |
| Need some parallel primitives | James, Petra, Casper, Dev | Common patterns become impossible without host help |

**Why they disagree:** Embedders see clean separation; script authors see limited expressiveness.

### Implicit Type Coercion
| Position | Experts | Rationale |
|----------|---------|-----------|
| Pragmatic for LLM workflows | Petra, Ivy | Reduces boilerplate when processing string outputs |
| JavaScript-tier footgun | Gigi, James, Raj | Silent coercion causes subtle bugs; inconsistent rules |

**Why they disagree:** Convenience vs. predictability tradeoff; depends on whether you trust script authors.

### Immutable `append()` Semantics
| Position | Experts | Rationale |
|----------|---------|-----------|
| Reduces aliasing bugs | Raj | Functional approach prevents mutation surprises |
| Unintuitive, error-prone | Petra, Dev | Violates expectations; will cause silent bugs when reassignment forgotten |

**Why they disagree:** Functional programming background vs. imperative expectations.

---

## 4. Perspective-Specific Insights

### Gigi (Go Expert)
- Host function error signaling is unspecified — how does a Go function return an error to Duso?
- Memory model for scripts accumulating context across many LLM calls needs bounds

### Petra (Python Expert)
- Missing list comprehensions hurt adoption among data-oriented users
- Object iteration only yields keys (no `items()` equivalent) adds friction

### James (JavaScript Expert)
- Conversation token limit handling is critical but unspecified
- Circular reference handling could cause infinite loops/stack overflows

### Tasha (TypeScript Expert)
- No way to express contracts between functions hurts enterprise maintainability
- Schema validation for `parse_json()` output is essential for robust pipelines

### Raj (Rust Expert)
- No `const`/immutability guarantees means configurations can be mutated unexpectedly
- Mutable closure captures without annotation hides side effects

### Luna (Lua Expert)
- Iteration order for object keys unspecified — matters for reproducible prompts
- The 0-based arrays vs. 1-based `for i = 1, 10` ranges is internally inconsistent

### Casper (C# Expert)
- No regex or `startswith`/`endswith` makes robust text extraction difficult
- Modern LLM function calling patterns have no language support

### Ivy (LLM Coding Assistant)
- Conversation copy/reference semantics critical for multi-agent patterns
- Cancellation/deadline propagation for timeouts unspecified

### Dev (Junior Developer)
- Safe navigation for nested JSON access (`result?.data?.nested`) would help
- Testing/mocking story completely absent

---

## 5. Most Important Questions

### Error & Exception Model (7 experts asked variants)
1. How do host functions signal errors to Duso? (Gigi, Casper, Ivy)
2. Are errors structured objects or just strings? Can you distinguish error types? (James, Tasha, Casper)
3. What happens to conversation state when a prompt fails mid-conversation? (Gigi, Petra, Tasha)

### Method/Object Binding Semantics (6 experts asked variants)
4. What is the resolution order when method property names collide with outer scope? (Gigi, James, Petra)
5. What happens when a method is detached and called independently? (Gigi, Raj, Ivy)

### Resource Limits & Memory (5 experts asked variants)
6. What are memory limits, execution bounds, and sandboxing mechanisms? (Gigi, Raj, Luna, Ivy, Dev)
7. How does GC work for long-running agent processes? (Petra, Luna, Casper)

### LLM Integration Specifics (5 experts asked variants)
8. How does streaming work for LLM responses? (Tasha, Casper, Dev)
9. How are token limits and conversation truncation handled? (James, Dev)
10. How do you implement retry logic, rate limiting, exponential backoff? (Petra, Dev)

### State & Serialization (4 experts asked variants)
11. Can conversation state persist across script invocations? (Luna, Ivy, Dev)
12. What are the copy vs. reference semantics for conversations? (Ivy, James)

---

## 6. Priority Recommendations

### P0: Critical (Address Before 1.0)

1. **Specify Error Model Completely**
   - Define how host functions return errors
   - Add structured error objects with type/code properties
   - Ensure line numbers and context are available in catch blocks
   - *Impact: Without this, production agent systems cannot implement proper retry/recovery logic*

2. **Document Method Binding Resolution**
   - Explicitly define precedence: object property vs. outer scope
   - Specify behavior when methods are extracted/passed as callbacks
   - Consider adding explicit `self` as an option
   - *Impact: Current ambiguity will cause debugging nightmares*

3. **Add Basic Collection Operations**
   - Implement `map(arr, fn)`, `filter(arr, fn)`, `find(arr, fn)` at minimum
   - Consider `reduce(arr, fn, initial)` for aggregation
   - *Impact: Every data transformation currently requires 4-6 lines instead of 1*

### P1: High (Address Soon After 1.0)

4. **Design Module System**
   - Even a simple `import "file" as name` would help
   - Need namespace isolation and explicit exports
   - *Impact: Language is limited to toy scripts without this*

5. **Add Parallel Execution Primitive**
   - Something like `parallel([fn1, fn2, fn3])` that host implements
   - Doesn't require async/await complexity
   - *Impact: Common "fan out and collect" pattern currently impossible*

6. **Clarify Conversation State Lifecycle**
   - Document what happens on error (is failed message in history?)
   - Define copy vs. reference semantics
   - Consider serialization for long-running workflows
   - *Impact: Core use case of multi-turn agents needs clear semantics*

### P2: Medium (Important for Adoption)

7. **Improve Debugging Story**
   - Stack traces with line numbers in errors
   - Consider a `debug()` function that prints scope
   - *Impact: Development experience is currently bare-bones*

8. **Add String Pattern Matching**
   - At minimum: `startswith()`, `endswith()`, `indexof()`
   - Consider simple regex or glob patterns
   - *Impact: Parsing non-JSON LLM output is currently painful*

9. **Document Resource Limits**
   - Specify how hosts set memory/time limits
   - Define behavior on limit exceeded
   - Document max string/array sizes
   - *Impact: Production deployments need predictable bounds*

### P3: Consider (Quality of Life)

10. **Reconsider Scope Default**
    - Option A: Keep current, add linter warning for shadowing
    - Option B: Require `var` for all declarations
    - *Impact: Prevents a class of subtle bugs*

11. **Unify Colon/Equals Syntax**
    - Either allow both everywhere or pick one
    - *Impact: Reduces cognitive load, prevents mixing errors*

12. **Add Safe Navigation Operator**
    - `obj?.prop?.nested` returns `nil` instead of erroring
    - *Impact: LLM responses often have optional/missing fields*