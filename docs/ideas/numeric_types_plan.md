# Numeric Types Plan: Adding Integer Support to Duso

## Problem Statement

Duso currently supports only 64-bit floating-point numbers, causing precision loss in critical use cases:

1. **Large API IDs** - Numbers > 2^53 lose precision (e.g., user IDs in large systems)
2. **Financial calculations** - Decimal precision issues (e.g., $19.99 * 3 != $59.97)
3. **Token counting** - Needs exact integers for LLM token counts
4. **JavaScript compatibility mistake** - Modern dynamic languages (Python, Lua 5.3) distinguish int/float

Expert panel feedback: "This is a known pain point in JavaScript that languages designed in 2025 should not repeat."

## Design Goals

1. **Precision where needed** - Exact integer arithmetic for API IDs, counts, indices
2. **Keep loose typing** - Automatic coercion between int and float (no explicit casting)
3. **Simplicity** - Don't add complexity for the 80% case (small numbers, general math)
4. **Backward compatible** - Existing scripts should work unchanged
5. **Performant** - int64 operations should be as fast as current float ops

## Design: Hybrid Int/Float Type

### Value Representation

```go
type Value struct {
    Type ValueType      // VAL_NUMBER (same as now)
    Data any            // int64 or float64 (discriminate by runtime type check)
}
```

Keep the existing `VAL_NUMBER` type. At runtime, `Data` contains either `int64` or `float64`. Type discrimination happens via `reflect.TypeOf()` or an explicit type field.

**Alternative:** Add a `VAL_INTEGER` type for cleaner separation, but costs an extra byte in Value and requires more updates.

**Recommendation:** Use the `any` field approach—minimal API changes, preserve Value size.

### Number Literal Parsing

**Lexer rules:**
- `123` → parsed as integer literal
- `123.0` → parsed as float literal
- `123.45` → parsed as float literal
- `1e5` → parsed as float literal (scientific notation implies float)
- `0x1F` → hex integer (parse as int64)

**Parser behavior:**
- Integer literals create `Value{Type: VAL_NUMBER, Data: int64(value)}`
- Float literals create `Value{Type: VAL_NUMBER, Data: float64(value)}`

### Type Coercion Rules

#### Arithmetic Operations

| Operation | Left | Right | Result |
|-----------|------|-------|--------|
| `+` | int | int | int |
| `+` | int | float | float |
| `+` | float | anything | float |
| `-` | int | int | int |
| `-` | int | float | float |
| `-` | float | anything | float |
| `*` | int | int | int |
| `*` | int | float | float |
| `*` | float | anything | float |
| `/` | int | int | **float** (division always produces decimals) |
| `/` | int | float | float |
| `/` | float | anything | float |
| `%` | int | int | int |
| `%` | int | float | error (modulo requires integer divisor) |
| `^` (power) | int | int | **float** (result may not be integer) |

#### Comparison Operations

- Type mismatch allowed: `5 == 5.0` → true
- String comparison always string: `"5" == 5` → false

#### Overflow Handling

- **int + int overflow** → Wraps like Go (undefined behavior for very large numbers)
- **int * int overflow** → Wraps (don't promote to float automatically)
- Document: "For arbitrary precision, use string parsing and convert explicitly"

### Built-in Function Updates

Functions that need updates:

1. **`type(value)`**
   - `type(5)` → `"number"` (not `"integer"`)
   - `type(5.0)` → `"number"`
   - Rationale: Loose typing, users shouldn't need to distinguish

2. **`tonumber(string)`**
   - `tonumber("123")` → int64
   - `tonumber("123.0")` → float64
   - `tonumber("123.45")` → float64

3. **`tostring(value)` / String coercion**
   - `tostring(5)` → `"5"`
   - `tostring(5.0)` → `"5.0"` (preserve float-ness)

4. **`format_json(value)`**
   - `format_json(5)` → `5` (integer in JSON)
   - `format_json(5.0)` → `5.0` (float in JSON)
   - Critical for APIs expecting exact types

5. **Array/Object operations**
   - Array indices: accept int only (convert float to int if exact: `a[5.0]` → `a[5]`)
   - Object keys: always string (coerce int to string)

6. **Comparison functions** (`min`, `max`, `sort`)
   - Work with both types, mixed types coerce to float for comparison

7. **Math functions** (`floor`, `ceil`, `round`, `sqrt`, `sin`, etc.)
   - Input: int or float
   - Output: `floor`, `ceil`, `round` → int; `sqrt`, `sin` → float

### String Representation

When printing or converting to string:
- `5` prints as `"5"` (integer)
- `5.0` prints as `"5.0"` (float, preserve decimal point)
- `5.5` prints as `"5.5"`

This is important for:
- REPL feedback
- String concatenation: `"count: " + 5` → `"count: 5"`
- JSON serialization

## Implementation Plan

### Phase 1: Core Value Representation (3-4 days)

1. **Lexer** (`pkg/script/lexer.go`)
   - Add `NumberLiteral` token that carries a flag: `isFloat` bool
   - Or: Create two tokens `INT_LITERAL` and `FLOAT_LITERAL`
   - Update number tokenization to parse and distinguish

2. **Parser** (`pkg/script/parser.go`)
   - Update `parseNumber()` to handle both literal types
   - Create appropriate Value at AST build time

3. **Value Type** (`pkg/script/value.go`)
   - Keep `VAL_NUMBER` type
   - Add helper functions:
     - `func (v Value) IsInteger() bool` - checks if Data is int64
     - `func (v Value) IsFloat() bool` - checks if Data is float64
     - `func (v Value) AsInt64() int64` - extract as int
     - `func (v Value) AsFloat64() float64` - extract as float
     - `func (v Value) AsNumber() float64` - for backward compat, convert to float

4. **Evaluator Arithmetic** (`pkg/script/evaluator.go`)
   - Replace simple float operations with coercion logic
   - Add helper: `func coerce(left, right Value) (leftFloat, rightFloat float64, result Value)`
   - Implement per-operator coercion rules
   - Lines ~776-815 (TOK_PLUS, TOK_MINUS, etc.)

### Phase 2: Built-in Functions (2-3 days)

Update these files:

1. **`pkg/script/builtins.go`** - Core functions
   - `tonumber()` - distinguish int/float parse
   - `tostring()` / string coercion
   - `floor()`, `ceil()`, `round()` → return int
   - `min()`, `max()` - handle mixed types
   - `type()` - still returns `"number"` for both

2. **`pkg/cli/format.go`** - Serialization
   - `format_json()` - preserve int/float distinction
   - JSON output must show `5` not `5.0` for integers

3. **`pkg/runtime/datastore.go`** - Datastore operations
   - Ensure type preservation across datastore boundary
   - Values serialized to disk should preserve int/float

### Phase 3: Testing (2-3 days)

1. **Unit tests** (`pkg/script/*_test.go`)
   - Test literal parsing (123 vs 123.0)
   - Coercion rules for all operators
   - Overflow behavior
   - Type mixing in functions

2. **Integration tests** (`examples/core/`)
   - Add `examples/core/numeric-types.du` showing:
     - Basic int arithmetic
     - Coercion behavior
     - Large numbers, API IDs
     - Financial-like calculations
     - JSON serialization

3. **Regression tests**
   - Verify existing scripts still work
   - Check float arithmetic produces same results

### Phase 4: Documentation (1-2 days)

1. **Update `docs/learning-duso.md`**
   - Add section: "Integer and Float Numbers"
   - Explain coercion rules with examples
   - Show when each type is useful
   - Document overflow behavior

2. **Update reference docs** (`docs/reference/`)
   - `tonumber.md` - mention int/float distinction
   - `type.md` - clarify still returns `"number"`
   - Arithmetic operators - document coercion table

3. **Migration guide** (if needed)
   - Explicitly state: backward compatible
   - Scripts using floats work unchanged
   - No breaking changes

## Backward Compatibility

✅ **Fully backward compatible** because:
- `5.0` still works, parsed as float64 (as now)
- `5 == 5.0` is true (coercion rules handle comparison)
- All existing float operations still work
- `type()` still returns `"number"` (doesn't distinguish)

⚠️ **Minor semantic changes:**
- `format_json(5)` now outputs `5` instead of `5.0` (more correct JSON)
- `5 / 2` now returns `2.5` (was already true, float division)
- `tostring(5)` outputs `"5"` not `"5.0"` (more natural)

These are improvements, not breaks.

## Testing Strategy

### Test Categories

1. **Literal parsing**
   - `123` → int64
   - `123.0` → float64
   - `1e5` → float64
   - `0x1F` → int64

2. **Arithmetic coercion**
   - All combinations of int/int, int/float, float/float
   - All operators (+, -, *, /, %, ^)
   - Division always returns float
   - Modulo rejects float operands

3. **Built-in functions**
   - `tonumber()` preserves type
   - `tostring()` shows difference
   - `floor/ceil/round` return int
   - `type()` still says `"number"`

4. **JSON round-trip**
   - `format_json(5)` → parse → int preserved
   - `format_json(5.0)` → parse → float preserved
   - API ID preservation (large numbers)

5. **Comparison & equality**
   - `5 == 5.0` → true
   - `5 < 6.0` → true
   - `sort()` works mixed types

6. **Edge cases**
   - Array indexing: `a[5.0]` vs `a[5]`
   - Object keys (always string)
   - Overflow wrapping (int + int overflow)
   - Very large numbers approaching float limits

## Risk Assessment

### Low Risk
- Lexer changes (isolated, well-tested)
- Arithmetic coercion (well-defined rules)
- New built-in tests

### Medium Risk
- Value representation change (touches evaluator widely)
- JSON serialization (must preserve type)
- Existing code paths using numbers

### Mitigation
- Comprehensive test suite before merge
- Run full test suite on existing examples
- Code review for evaluator changes
- Staged rollout: test locally, then merge

## Open Questions

1. **Array indexing**: Should `a[5.0]` auto-convert to `a[5]`? Or error?
   - **Proposal**: Auto-convert if exact (5.0 → 5), error if not (5.5 → error)

2. **Overflow behavior**: Should `int64.max + 1` wrap or error?
   - **Proposal**: Wrap (matches Go), document as undefined. Users needing arbitrary precision should use strings.

3. **Big numbers**: Should we support big.Int or decimal for financial math?
   - **Proposal**: Not now. Document as future feature. Workaround: use strings + external calculation.

4. **Hex/octal literals**: Support `0x1F`, `0o17`?
   - **Proposal**: Yes, for int literals. `0xFF_FF` for readability OK too.

## Success Criteria

✅ All expert panel concerns addressed:
- ✅ Large API IDs preserve precision
- ✅ Token counts are exact integers
- ✅ No JavaScript precision trap repeated

✅ No breaking changes to existing scripts

✅ Coercion rules clear and documented

✅ Performance unchanged (or better)

✅ 1,290+ tests pass (maintain coverage)

## Timeline Estimate

**Total: ~10-14 days of focused work**
- Phase 1 (Core): 3-4 days
- Phase 2 (Built-ins): 2-3 days
- Phase 3 (Testing): 2-3 days
- Phase 4 (Docs): 1-2 days
- Buffer & review: 1-2 days

## Next Steps

1. **Review & feedback** - Validate design with team
2. **Create GitHub issue** - Track implementation
3. **Create feature branch** - `feature/int-float-types`
4. **Implement Phase 1** - Start with lexer/parser
5. **Test incrementally** - Add tests as you go
6. **Document as you build** - Keep docs in sync
