# Test Coverage Report

**Last Updated:** February 12, 2026
**Total Tests:** 1,290
**Overall Coverage:** ~45%

## Summary

Duso has comprehensive test coverage across its core interpreter and runtime. With nearly 1,300 tests, the test suite validates language semantics, type coercion, built-in functions, and file I/O operations.

## Coverage by Package

| Package | Tests | Coverage | Status |
|---------|-------|----------|--------|
| `pkg/script` | ~700 | 56.2% | ✅ Interpreter/runtime - highest priority |
| `pkg/runtime` | ~300 | 39.9% | ✅ Runtime behavior and task orchestration |
| `pkg/cli` | ~290 | 38.3% | ✅ File I/O operations and module loading |
| **Total** | **~1,290** | **~45%** | ✅ All tests passing |

## What's Tested

### pkg/script (56.2% coverage)
The interpreter and language evaluation engine - the most critical component.

**Well-covered areas:**
- Lexer and tokenization
- Parser and AST construction
- Type system and type coercion
- Expression evaluation
- Built-in functions (100+ functions)
- Control flow (if/else, loops, etc.)
- Function definitions and calls
- Error handling and propagation
- Variable scoping and environments

**Gap areas:**
- Some edge cases in complex nested expressions
- Certain error recovery scenarios

### pkg/runtime (39.9% coverage)
Task execution, concurrency, and agent orchestration.

**Well-covered areas:**
- Task creation and execution
- Concurrent task scheduling
- Result aggregation
- Datastore operations
- Module resolution and loading

**Gap areas:**
- Complex task interaction scenarios
- Rare concurrency edge cases

### pkg/cli (38.3% coverage)
File I/O, module loading, and command-line operations.

**Well-covered areas:**
- File reading/writing (local, /EMBED/, /STORE/)
- Directory listing and navigation
- Glob pattern matching
- File metadata operations
- Module resolution

**Gap areas:**
- Complex nested directory operations
- Edge cases with unusual file permissions
- Some error handling scenarios

## Running Tests

### Run all tests with coverage:

```bash
go test ./pkg/... -cover
```

### Run specific package tests:

```bash
go test ./pkg/script -v
go test ./pkg/cli -v
go test ./pkg/runtime -v
```

### Generate coverage HTML report:

```bash
go test ./pkg/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### View coverage summary:

```bash
go test ./pkg/... -cover 2>&1 | grep "coverage:"
```

## Coverage Goals

For a 22-day-old language:

- **Current:** ~45% coverage with 1,290 tests
- **Short-term (next 30 days):** 55%+ coverage
- **Long-term:** 70%+ coverage for production readiness

Priority areas for improvement:
1. Error handling edge cases across all packages
2. Complex type coercion scenarios
3. Unusual file system operations
4. Rare concurrency patterns

## Test Infrastructure

### Key Test Features

- ✅ Automated test discovery and execution
- ✅ Parallel test execution
- ✅ Coverage measurement
- ✅ Type checking validation
- ✅ Integration testing via script execution
- ✅ File I/O testing with temporary directories

### Test Commands

All tests in the project:
```bash
go test ./pkg/...
```

With verbose output:
```bash
go test ./pkg/... -v
```

With race condition detection:
```bash
go test ./pkg/... -race
```

## Continuous Improvement

The test suite is actively maintained alongside language development. As new features are added, tests are written to validate:

- Correctness of new functionality
- Backward compatibility
- Edge cases and error conditions
- Performance characteristics

New contributors should follow the test-driven development approach: write tests first, then implement features to pass them.
