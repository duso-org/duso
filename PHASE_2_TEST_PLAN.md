# Phase 2: Test Coverage Expansion Plan
## Goal: Reach 80% coverage in pkg/script/

**Status:** Phase 1 Complete - 128/128 tests passing, 47.3% coverage

---

## Priority 1: Datastore (Critical for Concurrency)

Datastore is the cornerstone of thread-safe inter-process communication between concurrent script processes.

### Tests Needed:
- `TestDatastore_Basic` - set(), get(), increment(), append()
- `TestDatastore_Atomic` - increment() and append() atomicity
- `TestDatastore_Wait` - wait() blocking behavior
- `TestDatastore_WaitFor` - wait_for() condition checking
- `TestDatastore_Namespacing` - independent namespaces don't conflict
- `TestDatastore_Concurrency` - parallel access patterns
- `TestDatastore_Persistence` - optional JSON save/load
- `TestDatastore_KeyNotFound` - proper nil/error handling

**Files to test:**
- `datastore_value.go` (0% coverage)
- `script.go` datastore methods
- `evaluator.go` builtin_datastore

**Estimate:** 12-15 tests = ~3-5% coverage gain

---

## Priority 2: HTTP Server (Critical for Web/Agent APIs)

HTTP server enables Duso to serve as a web application and agent API handler.

### Tests Needed:
- `TestHTTPServer_BasicServer` - start/stop lifecycle
- `TestHTTPServer_MethodRouting` - GET, POST, DELETE routing
- `TestHTTPServer_PrefixMatching` - route prefix patterns
- `TestHTTPServer_RequestContext` - ctx.request() data
- `TestHTTPServer_ResponseHandling` - ctx.response() format
- `TestHTTPServer_SelfReferential` - script as own handler
- `TestHTTPServer_ConcurrentRequests` - parallel requests in goroutines
- `TestHTTPServer_ErrorHandling` - 404, 500 responses
- `TestHTTPServer_HeaderManagement` - custom headers
- `TestHTTPServer_CallStack` - ctx.callstack() traces

**Files to test:**
- `http_server_value.go` (0% coverage)
- `http_server.go` (0% coverage)
- `http_value.go` (0% coverage)
- `evaluator.go` builtin_http_server

**Estimate:** 10-12 tests = ~3-5% coverage gain

---

## Priority 3: Edge Cases & Error Paths

### Partially Covered Builtins (30-70% coverage):
- `replace()` (35%) - more patterns, edge cases
- `toBool()` (40%) - all type conversions
- `upper()`/`lower()` (45%) - Unicode, empty strings
- `toNumber()` (61%) - parse errors, edge cases
- `trim()` (60%) - various whitespace
- `len()` (57%) - all types
- `append()` (71%) - immutability verification

### Tests Needed:
- `TestBuiltins_ErrorCases` - invalid inputs to each function
- `TestBuiltins_BoundaryConditions` - empty arrays, nil values, etc.
- `TestBuiltins_TypeErrors` - calling with wrong types
- `TestBuiltins_LargeInputs` - performance boundaries
- `TestEvaluator_ErrorPaths` - division by zero, undefined vars, type mismatches
- `TestEvaluator_TypeMismatch` - operations between incompatible types

**Estimate:** 20-25 tests = ~3-5% coverage gain

---

## Priority 4: Evaluator Runtime Paths

### Low Coverage Functions:
- `value.go` String() (22.2%)
- `errors.go` Error() (48.1%)
- `evaluator.go` tryCoerceToNumber (25%)

### Tests Needed:
- `TestValue_StringConversion` - all 7 types as strings
- `TestValue_TypeComparisons` - cross-type comparisons
- `TestError_Formatting` - error message format and position info
- `TestCoercion_NumericContext` - implicit conversions in math
- `TestEvaluator_ImplicitCasts` - where coercion happens

**Estimate:** 8-10 tests = ~2-3% coverage gain

---

## Priority 5: System Features

### Medium Effort, Good Coverage:
- `template()` builtin (0%) - dynamic template compilation
- `input()` builtin (0%) - stdin handling (can mock)
- `load()` / `save()` (need edge case tests) - file I/O
- `run()` / `spawn()` - script execution

### Tests Needed:
- `TestBuiltin_Template` - template compilation and execution
- `TestBuiltin_Input` - stdin mocking
- `TestBuiltin_FileIO` - temp files, permissions, not found
- `TestBuiltin_ScriptExecution` - run() return values, spawn() async

**Estimate:** 12-15 tests = ~2-3% coverage gain

---

## Coverage Targets by Phase

| Component | Phase 1 | Phase 2 Target | Notes |
|-----------|---------|----------------|-------|
| Parser | ~85% | ~90% | Already good |
| Lexer | ~90% | ~95% | Already good |
| Evaluator | ~40% | ~75% | Heavy focus |
| Builtins | ~35% | ~70% | Add edge cases |
| Datastore | 0% | ~80% | CRITICAL |
| HTTP Server | 0% | ~75% | CRITICAL |
| Overall | 47.3% | **~75-80%** | **Goal** |

---

## Implementation Order

1. **Week 1: Datastore** (12-15 tests)
   - Foundation for understanding concurrent patterns
   - High business value
   - Enables testing spawn()/run() properly

2. **Week 2: HTTP Server** (10-12 tests)
   - Second critical system
   - Agent/API capabilities
   - Good integration with datastore tests

3. **Week 3: Builtin Edge Cases** (20-25 tests)
   - Lower difficulty
   - Quick wins
   - Fill gaps in existing coverage

4. **Week 4: Evaluator & Runtime** (8-10 tests)
   - Type coercion edge cases
   - Error handling paths
   - System features (template, input, file I/O)

---

## Testing Strategy

### Datastore Testing:
```go
// Use goroutines to test concurrent access
// Verify atomic operations
// Test wait() with channels
// Mock timestamp for wait_for conditions
```

### HTTP Server Testing:
```go
// Start actual server on random port
// Make HTTP requests to it
// Test request/response context passing
// Verify handler isolation (fresh evaluator per request)
// Test error scenarios (bad routes, invalid methods)
```

### Edge Cases:
```go
// Create comprehensive error case tables
// Test with nil, empty, and extreme values
// Verify type coercion consistency
// Test boundary conditions
```

---

## Success Criteria

- ✅ All Phase 2 tests pass
- ✅ Coverage reaches 75-80% in pkg/script/
- ✅ Datastore thoroughly tested (agent orchestration confidence)
- ✅ HTTP server thoroughly tested (production API readiness)
- ✅ All builtin functions have >60% coverage
- ✅ Error paths well covered

---

## Known Gaps (Acceptable)

These don't need 80% coverage for launch:
- System metrics (`sys_metrics.go`) - observability, not core logic
- Debug/breakpoint features (`context.go` Depth, SetDebugMode) - optional feature
- Some CLI-only features (stdin handling edge cases)

**Rationale:** These don't affect core language semantics or multi-script orchestration.

---

## Next Steps

1. Create `datastore_test.go` with concurrent access patterns
2. Create `http_server_test.go` with integration scenarios
3. Expand `builtins_test.go` with edge cases and error tests
4. Run coverage report: `go test -cover ./pkg/script/...`
5. Track progress toward 80% target
