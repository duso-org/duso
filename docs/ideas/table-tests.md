# Modern Go Testing Patterns (2025)

Accelerating test coverage using Go's built-in testing features and idiomatic patterns.

## Problem Statement

Testing 16,385 lines of Go code across ~670 functions/types traditionally requires 200-300+ separate test functions to achieve 80% coverage. Modern Go testing patterns can reduce this dramatically.

## Solution: Table-Driven Tests with Subtests

The most powerful pattern combines:
1. **Table-driven tests** - Define test cases as data structures
2. **Subtests** (`t.Run()`) - Run multiple cases within one test function
3. **Parallel execution** (`t.Parallel()`) - Run independent tests concurrently
4. **Fuzzing** - Property-based testing for edge cases

## Pattern 1: Table-Driven Tests with Subtests

Instead of writing 50 test functions, write ONE with a table of cases:

```go
func TestEvaluatorBinaryOps(t *testing.T) {
    tests := []struct {
        name     string
        left     any
        op       string
        right    any
        expected any
    }{
        {"add", 2.0, "+", 3.0, 5.0},
        {"subtract", 5.0, "-", 3.0, 2.0},
        {"multiply", 3.0, "*", 4.0, 12.0},
        {"divide", 10.0, "/", 2.0, 5.0},
        // ... 20+ more cases
    }

    for _, tt := range tests {
        tt := tt // Important: capture for parallel scope
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            result := evaluate(tt.left, tt.op, tt.right)
            if result != tt.expected {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}
```

### Benefits
- **1 test function = 20-50 test cases** (not 20-50 separate functions)
- Clear test case organization
- Easy to add new scenarios
- Reduced code duplication

## Pattern 2: Parallel Execution

Mark independent tests to run concurrently - dramatically speeds up test suite:

```go
func TestLexer(t *testing.T) {
    t.Parallel() // Run this test in parallel with others
    // ... test code
}
```

**Critical for parallel table-driven tests:** The `tt := tt` assignment creates a new variable scoped to each loop iteration, ensuring each goroutine has its own copy of test case data and preventing race conditions.

## Pattern 3: Fuzzing (Go 1.18+)

Property-based testing using Go's built-in fuzzing for edge case discovery:

```go
func FuzzParser(f *testing.F) {
    // Seed corpus - provides starting point
    f.Add("1 + 2")
    f.Add("if x then y end")
    f.Add("function foo() end")

    f.Fuzz(func(t *testing.T, input string) {
        ast, err := Parse(input)
        // Go generates 1000s of input variations automatically
        // Just verify the parser doesn't panic on any input
        _ = ast
        _ = err
    })
}
```

Go 1.25 automatically:
- Generates thousands of test variations
- Stores failing inputs as corpus entries
- Improves corpus over time
- Perfect for lexer/parser edge cases

## Pattern 4: Test Fixtures

Organize test data in `testdata/` directories:

```
pkg/script/
├── evaluator.go
├── evaluator_test.go
└── testdata/
    ├── fixtures.go
    ├── sample_programs/
    │   ├── loop.du
    │   ├── function.du
    │   └── complex.du
    └── fuzz/
        └── FuzzParser/
            ├── corpus/file1
            └── corpus/file2
```

## Estimated Coverage Impact for Duso

### Without Table-Driven Tests
- 200-300 separate test functions needed
- 80% coverage achievable

### With Table-Driven + Subtests + Parallel
- **60-80 test functions needed** (68% reduction)
- 80% coverage achievable
- Tests run 2-3x faster due to `t.Parallel()`

### Breakdown by Package

| Package | Approach | Functions | Cases/Function | Total Tests |
|---------|----------|-----------|-----------------|-------------|
| script/evaluator | Tables + Subtests | 5 | 8-12 | 40-60 cases |
| script/parser | Tables + Subtests | 4 | 10-15 | 40-60 cases |
| script/lexer | Tables + Subtests | 3 | 8-10 | 24-30 cases |
| runtime/builtins | Tables + Subtests | 20 | 3-5 | 60-100 cases |
| runtime/datastore | Tables + Subtests | 2 | 15-20 | 30-40 cases |
| cli/ | Tables + Subtests | 4 | 5-8 | 20-32 cases |
| lsp/ | Tables + Subtests | 3 | 4-6 | 12-18 cases |
| **Total** | | **41** | **~6 avg** | **226-340 cases across 41 functions** |

## Implementation Strategy

### Phase 1: Core Interpreter (Highest Impact)
Focus on these for maximum coverage gain:
1. `script/evaluator.go` - 40-60 test cases in 5 functions
2. `script/parser.go` - 40-60 test cases in 4 functions
3. `script/lexer.go` - 24-30 test cases in 3 functions

### Phase 2: Runtime Builtins
1. Create generic table-driven test for each builtin family
2. Use `t.Parallel()` for independent builtin tests
3. Consider fuzzing for string/array operations

### Phase 3: Advanced Features
1. Fuzzing for parser/lexer robustness
2. Integration tests using fixtures in testdata/
3. Concurrency tests for goroutine-related code

## Key Go 1.25 Advantages

- Native fuzzing with `testing.F` and `FuzzXxx` functions
- Built-in corpus management
- Automatic failure case recording
- `t.Parallel()` is zero-cost for independent tests
- Subtests with `t.Run()` organize complex test logic

## References

- [Table-Driven Tests - Go Wiki](https://go.dev/wiki/TableDrivenTests)
- [Prefer Table Driven Tests - Dave Cheney](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)
- [Parallel Table-Driven Tests in Go (2025)](https://www.glukhov.org/post/2025/12/parallel-table-driven-tests-in-go/)
- [Go Fuzzing - Official Documentation](https://go.dev/doc/security/fuzz/)
- [Tutorial: Getting started with fuzzing](https://go.dev/doc/tutorial/fuzz)
- [Go Unit Testing Best Practices (2025)](https://www.glukhov.org/post/2025/11/unit-tests-in-go/)
- [Go Testing Excellence: Table-Driven Tests](https://dasroot.net/posts/2026/01/go-testing-excellence-table-driven-tests-mocking/)
- [Testing Package Documentation](https://pkg.go.dev/testing)

## Example: Converting Traditional Tests to Tables

### Before (3 Functions, 3 Cases)
```go
func TestAddition(t *testing.T) {
    if evaluate(2.0, "+", 3.0) != 5.0 {
        t.Fail()
    }
}

func TestSubtraction(t *testing.T) {
    if evaluate(5.0, "-", 3.0) != 2.0 {
        t.Fail()
    }
}

func TestMultiplication(t *testing.T) {
    if evaluate(3.0, "*", 4.0) != 12.0 {
        t.Fail()
    }
}
```

### After (1 Function, 3+ Cases, Parallelizable)
```go
func TestArithmetic(t *testing.T) {
    t.Parallel()
    tests := []struct {
        name     string
        left, right float64
        op       string
        expected float64
    }{
        {"add", 2.0, 3.0, "+", 5.0},
        {"subtract", 5.0, 3.0, "-", 2.0},
        {"multiply", 3.0, 4.0, "*", 12.0},
        {"divide", 10.0, 2.0, "/", 5.0},
        {"modulo", 7.0, 3.0, "%", 1.0},
    }

    for _, tt := range tests {
        tt := tt
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            if got := evaluate(tt.left, tt.right, tt.op); got != tt.expected {
                t.Errorf("evaluate(%v, %v, %q) = %v, want %v",
                    tt.left, tt.right, tt.op, got, tt.expected)
            }
        })
    }
}
```

**Result:** 3 functions → 1 function, 3 cases → 5+ cases, parallelizable
