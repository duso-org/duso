# Evaluation of Duso Language Specification

## Top 5 Pros

1. **Zero External Dependencies Philosophy** - The design principle of building entirely on Go's standard library with modules "baked into the binary" is excellent for reliability and deployment. This eliminates dependency hell, version conflicts, and supply chain vulnerabilities. The "frozen binary" concept means scripts remain reproducible indefinitely—a rare and valuable property for production systems.

2. **Pragmatic String Templating for LLM Workflows** - The `{{expr}}` template syntax with triple-quoted multiline strings is exceptionally well-suited for LLM integration. The decision to not require escaping single braces means JSON and code snippets work naturally without contortion. This addresses a real pain point in prompt engineering where escaping often becomes a nightmare.

3. **Objects-as-Constructors Pattern** - The ability to call objects as functions to create copies with overrides (`Config(timeout = 60)`) is elegant and avoids the complexity of class-based OOP while providing practical object instantiation. This hits a sweet spot between simplicity and utility for configuration-heavy orchestration scripts.

4. **Host-Controlled Parallelism Model** - Keeping the language single-threaded while delegating parallelism to the host via `parallel()` is a smart architectural decision. It keeps the language semantics simple and predictable while still enabling concurrent operations where they matter (API calls, tool invocations). The read-only parent scope access in parallel blocks prevents data races by design.

5. **Clean Module System with Clear Semantics** - The distinction between `require()` (isolated, cached) and `include()` (current scope, uncached) provides flexibility without ambiguity. Circular dependency detection and clear path resolution rules prevent common module system foot-guns.

## Top 5 Cons

1. **No Memory Safety Guarantees or Resource Management** - From a Rust perspective, the lack of any ownership model, lifetime tracking, or RAII-style resource cleanup is concerning. There's no `defer`, no destructors, and no way to ensure resources (file handles, network connections) are properly released. For long-running agent orchestration, resource leaks could accumulate. The `try/catch` mechanism doesn't guarantee cleanup in error paths.

2. **Implicit Type Coercion Can Hide Bugs** - While convenient, automatic string-to-number coercion in comparisons (`5 < "10"`) and the complex truthiness rules (empty arrays/objects are falsy but `"0"` is truthy) create subtle bug opportunities. The asymmetry between `0` (falsy) and `"0"` (truthy) will catch developers off guard. Rust's explicit type system exists precisely to prevent this class of errors.

3. **Mutable State and Scope Ambiguity** - The scoping rules where assignment "walks up the scope chain" unless `var` is used is error-prone. A typo in a variable name silently creates a new local instead of modifying the intended outer variable (or vice versa). The implicit `self` in object methods (accessing `name` directly instead of `self.name`) makes it unclear what's being referenced and could conflict with closure captures.

4. **Limited Error Context and No Stack Traces** - The specification shows simple error strings like `"division by zero"` without line numbers, stack traces, or structured error types. For agent orchestration where operations may involve multiple LLM calls and tool invocations, debugging failures without proper context will be painful. There's no `Result` type or error chaining mechanism.

5. **No Bounds Checking or Validation Primitives** - Arrays can be accessed out of bounds (requiring try/catch), but there's no way to express contracts, assertions, or validate data structures. For orchestrating agents that receive unpredictable LLM outputs, the lack of schema validation, pattern matching, or guard clauses means defensive code will be verbose and repetitive.

## Top 5 Questions

1. **How does the `parallel()` function handle resource contention when multiple blocks call the same host-provided function (e.g., rate-limited APIs)?** Is there any mechanism for backpressure, queuing, or retry logic, or must this be implemented in every host function? What happens if the underlying Go goroutines panic?

2. **What are the memory characteristics of the tree-walking interpreter for large data structures?** If an LLM returns a large JSON response that gets parsed into nested objects, how does garbage collection work? Can memory be explicitly released, or must scripts rely on