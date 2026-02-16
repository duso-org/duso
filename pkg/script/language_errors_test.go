package script

import (
	"testing"
)

// TestUndefinedVariables tests error cases with undefined variables
func TestUndefinedVariables(t *testing.T) {
	t.Parallel()

	interp := NewInterpreter(false)

	tests := []struct {
		name        string
		script      string
		expectError bool
	}{
		{
			name:        "undefined variable in expression",
			script:      `result = undefined_var + 1`,
			expectError: true,
		},
		{
			name:        "undefined variable in assignment",
			script:      `x = y`,
			expectError: true,
		},
		{
			name:        "undefined function call",
			script:      `result = nonexistent_function()`,
			expectError: true,
		},
		{
			name:        "undefined in comparison",
			script:      `if undefined > 5 then x = 1 end`,
			expectError: true,
		},
		{
			name:        "undefined in array access",
			script:      `result = arr[0]`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := interp.Execute(tt.script)
			if (err != nil) != tt.expectError {
				t.Fatalf("expected error=%v, got error=%v", tt.expectError, err != nil)
			}
		})
	}
}

// TestTypeErrors tests error cases with type mismatches
func TestTypeErrors(t *testing.T) {
	t.Parallel()

	interp := NewInterpreter(false)

	tests := []struct {
		name        string
		script      string
		expectError bool
	}{
		{
			name:        "subtract string from number",
			script:      `result = 10 - "text"`,
			expectError: true,
		},
		{
			name:        "divide number by string",
			script:      `result = 10 / "text"`,
			expectError: true,
		},
		{
			name:        "multiply array by number",
			script:      `result = [1, 2] * 3`,
			expectError: true,
		},
		{
			name:        "modulo with non-number",
			script:      `result = 10 % "text"`,
			expectError: true,
		},
		{
			name:        "call non-function",
			script:      `x = 5; result = x()`,
			expectError: true,
		},
		{
			name:        "index string with non-number",
			script:      `x = "hello"; result = x["key"]`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := interp.Execute(tt.script)
			if (err != nil) != tt.expectError {
				t.Fatalf("expected error=%v, got error=%v", tt.expectError, err != nil)
			}
		})
	}
}

// TestArrayIndexErrors tests array indexing errors
func TestArrayIndexErrors(t *testing.T) {
	t.Parallel()

	interp := NewInterpreter(false)

	tests := []struct {
		name        string
		script      string
		expectError bool
	}{
		{
			name:        "array index out of bounds positive",
			script:      `arr = [1, 2, 3]; result = arr[10]`,
			expectError: true,
		},
		{
			name:        "array index negative",
			script:      `arr = [1, 2, 3]; result = arr[-1]`,
			expectError: true,
		},
		{
			name:        "index non-array non-object",
			script:      `x = 42; result = x[0]`,
			expectError: true,
		},
		{
			name:        "index with string key on array",
			script:      `arr = [1, 2, 3]; result = arr["key"]`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := interp.Execute(tt.script)
			if (err != nil) != tt.expectError {
				t.Fatalf("expected error=%v, got error=%v", tt.expectError, err != nil)
			}
		})
	}
}

// TestPropertyAccessErrors tests object property access errors
func TestPropertyAccessErrors(t *testing.T) {
	t.Parallel()

	interp := NewInterpreter(false)

	tests := []struct {
		name        string
		script      string
		expectError bool
	}{
		{
			name:        "property access on non-object",
			script:      `x = 42; result = x.property`,
			expectError: true,
		},
		{
			name:        "property access on array",
			script:      `arr = [1, 2, 3]; result = arr.property`,
			expectError: true,
		},
		{
			name:        "property access on string",
			script:      `s = "hello"; result = s.property`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := interp.Execute(tt.script)
			if (err != nil) != tt.expectError {
				t.Fatalf("expected error=%v, got error=%v", tt.expectError, err != nil)
			}
		})
	}
}

// TestFunctionErrors tests function-related errors
func TestFunctionErrors(t *testing.T) {
	t.Parallel()

	interp := NewInterpreter(false)

	tests := []struct {
		name        string
		script      string
		expectError bool
	}{
		{
			name: "function with wrong argument count",
			script: `
				function test(a, b)
					return a + b
				end
				result = test(1)
			`,
			expectError: true,
		},
		{
			name: "function redefinition",
			script: `
				function test()
					return 1
				end
				function test()
					return 2
				end
			`,
			expectError: false, // Allowed - overwrites
		},
		{
			name: "function parameter shadowing",
			script: `
				x = 10
				function test(x)
					return x
				end
				result = test(20)
				y = x
			`,
			expectError: false, // This is valid
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := interp.Execute(tt.script)
			if (err != nil) != tt.expectError {
				t.Fatalf("expected error=%v, got error=%v", tt.expectError, err != nil)
			}
		})
	}
}

// TestParsingErrors tests syntax errors
func TestParsingErrors(t *testing.T) {
	t.Parallel()

	interp := NewInterpreter(false)

	tests := []struct {
		name        string
		script      string
		expectError bool
	}{
		{
			name:        "unclosed parenthesis",
			script:      `x = (1 + 2`,
			expectError: true,
		},
		{
			name:        "unclosed array bracket",
			script:      `x = [1, 2, 3`,
			expectError: true,
		},
		{
			name:        "unclosed object brace",
			script:      `x = {a = 1`,
			expectError: true,
		},
		{
			name:        "unexpected token",
			script:      `x = 1 +* 2`,
			expectError: true,
		},
		{
			name:        "invalid assignment",
			script:      `1 = x`,
			expectError: true,
		},
		{
			name:        "missing then in if",
			script:      `
				if x > 5
					y = 1
				end
			`,
			expectError: true,
		},
		{
			name:        "missing end for function",
			script: `
				function test()
					return 1
			`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := interp.Execute(tt.script)
			if (err != nil) != tt.expectError {
				t.Fatalf("expected error=%v, got error=%v", tt.expectError, err != nil)
			}
		})
	}
}

// TestArithmeticErrors tests arithmetic edge cases and errors
func TestArithmeticErrors(t *testing.T) {
	t.Parallel()

	interp := NewInterpreter(false)

	tests := []struct {
		name        string
		script      string
		expectError bool
	}{
		{
			name:        "division by zero",
			script:      `result = 10 / 0`,
			expectError: true,
		},
		{
			name:        "modulo by zero",
			script:      `result = 10 % 0`,
			expectError: true,
		},
		{
			name:        "unary minus on string",
			script:      `x = "hello"; result = -x`,
			expectError: true,
		},
		{
			name:        "addition of incompatible types",
			script:      `result = [1, 2] + {a = 1}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := interp.Execute(tt.script)
			if (err != nil) != tt.expectError {
				t.Fatalf("expected error=%v, got error=%v", tt.expectError, err != nil)
			}
		})
	}
}

// TestControlFlowErrors tests control flow statement errors
func TestControlFlowErrors(t *testing.T) {
	t.Parallel()

	interp := NewInterpreter(false)

	tests := []struct {
		name        string
		script      string
		expectError bool
	}{
		{
			name:        "break outside loop",
			script:      `break`,
			expectError: true,
		},
		{
			name:        "continue outside loop",
			script:      `continue`,
			expectError: true,
		},
		{
			name:        "return outside function",
			script:      `return 42`,
			expectError: true,
		},
		{
			name: "break in if inside loop",
			script: `
				for i = 1, 5 do
					if i == 3 then
						break
					end
				end
			`,
			expectError: false, // Valid
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := interp.Execute(tt.script)
			if (err != nil) != tt.expectError {
				t.Fatalf("expected error=%v, got error=%v", tt.expectError, err != nil)
			}
		})
	}
}

// TestTryCatchErrors tests try/catch error behavior
func TestTryCatchErrors(t *testing.T) {
	t.Parallel()

	interp := NewInterpreter(false)

	tests := []struct {
		name        string
		script      string
		expectError bool
	}{
		{
			name: "catch undefined error",
			script: `
				try
					x = undefined_var
				catch (e)
					result = e
				end
			`,
			expectError: false, // Error caught
		},
		{
			name: "throw in try",
			script: `
				try
					throw("custom error")
				catch (e)
					result = e
				end
			`,
			expectError: false, // Error caught
		},
		{
			name: "error not caught",
			script: `
				try
					x = 1
				catch (e)
					x = 2
				end
				y = undefined_var
			`,
			expectError: true, // Error outside try/catch
		},
		{
			name: "throw in catch re-throws",
			script: `
				try
					throw("first")
				catch (e)
					throw("second")
				end
			`,
			expectError: true, // Second error propagates
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := interp.Execute(tt.script)
			if (err != nil) != tt.expectError {
				t.Fatalf("expected error=%v, got error=%v", tt.expectError, err != nil)
			}
		})
	}
}

// TestLoopErrors tests loop-related errors
func TestLoopErrors(t *testing.T) {
	t.Parallel()

	interp := NewInterpreter(false)

	tests := []struct {
		name        string
		script      string
		expectError bool
	}{
		{
			name: "for loop with invalid range",
			script: `
				for i = "start", "end" do
					x = 1
				end
			`,
			expectError: true,
		},
		{
			name: "for loop over non-iterable",
			script: `
				for item in 42 do
					x = 1
				end
			`,
			expectError: true,
		},
		{
			name: "while with non-boolean condition",
			script: `
				x = 0
				while "not a bool" do
					x = 1
					break
				end
			`,
			expectError: false, // Truthy value is OK
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := interp.Execute(tt.script)
			if (err != nil) != tt.expectError {
				t.Fatalf("expected error=%v, got error=%v", tt.expectError, err != nil)
			}
		})
	}
}

// TestInvalidOperations tests invalid operations between incompatible types
func TestInvalidOperations(t *testing.T) {
	t.Parallel()

	interp := NewInterpreter(false)

	tests := []struct {
		name        string
		script      string
		expectError bool
	}{
		{
			name:        "compare array to number",
			script:      `result = [1, 2] > 5`,
			expectError: true,
		},
		{
			name:        "compare object to string",
			script:      `result = {a = 1} < "text"`,
			expectError: true,
		},
		{
			name:        "logical and with non-boolean left",
			script:      `result = 42 and true`,
			expectError: false, // Truthy coercion is OK
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := interp.Execute(tt.script)
			if (err != nil) != tt.expectError {
				t.Fatalf("expected error=%v, got error=%v", tt.expectError, err != nil)
			}
		})
	}
}
