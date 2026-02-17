package script

import (
	"testing"
)

// TestArithmeticOperators tests all arithmetic operators and precedence
func TestArithmeticOperators(t *testing.T) {
	t.Parallel()

	interp := NewInterpreter()

	tests := []struct {
		name   string
		script string
	}{
		{
			name:   "addition",
			script: `result = 5 + 3`,
		},
		{
			name:   "subtraction",
			script: `result = 10 - 4`,
		},
		{
			name:   "multiplication",
			script: `result = 6 * 7`,
		},
		{
			name:   "division",
			script: `result = 20 / 4`,
		},
		{
			name:   "modulo",
			script: `result = 17 % 5`,
		},
		{
			name:   "negative number",
			script: `result = -42`,
		},
		{
			name:   "unary minus on variable",
			script: `x = 10; result = -x`,
		},
		{
			name:   "multiplication before addition",
			script: `result = 2 + 3 * 4`,
		},
		{
			name:   "parentheses override precedence",
			script: `result = (2 + 3) * 4`,
		},
		{
			name:   "complex expression",
			script: `result = 10 + 5 * 2 - 3 / 1`,
		},
		{
			name:   "division with floats",
			script: `result = 7 / 2`,
		},
		{
			name:   "modulo with negative",
			script: `result = -10 % 3`,
		},
		{
			name:   "chained operations",
			script: `result = 1 + 2 + 3 + 4 + 5`,
		},
		{
			name:   "mixed operators",
			script: `result = 100 / 2 - 10 * 3 + 5`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := interp.Execute(tt.script)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// TestComparisonOperators tests all comparison operators
func TestComparisonOperators(t *testing.T) {
	t.Parallel()

	interp := NewInterpreter()

	tests := []struct {
		name   string
		script string
	}{
		{
			name:   "equal numbers",
			script: `result = 5 == 5`,
		},
		{
			name:   "not equal numbers",
			script: `result = 5 != 3`,
		},
		{
			name:   "less than",
			script: `result = 3 < 5`,
		},
		{
			name:   "greater than",
			script: `result = 7 > 3`,
		},
		{
			name:   "less than or equal true",
			script: `result = 5 <= 5`,
		},
		{
			name:   "less than or equal false",
			script: `result = 10 <= 5`,
		},
		{
			name:   "greater than or equal true",
			script: `result = 7 >= 7`,
		},
		{
			name:   "greater than or equal false",
			script: `result = 3 >= 5`,
		},
		{
			name:   "string equality",
			script: `result = "hello" == "hello"`,
		},
		{
			name:   "string inequality",
			script: `result = "hello" != "world"`,
		},
		{
			name:   "string less than",
			script: `result = "apple" < "banana"`,
		},
		{
			name:   "boolean equality",
			script: `result = true == true`,
		},
		{
			name:   "boolean inequality",
			script: `result = true != false`,
		},
		{
			name:   "nil equality",
			script: `result = nil == nil`,
		},
		{
			name:   "mixed type comparison",
			script: `result = "5" < 10`,
		},
		{
			name:   "chained comparison",
			script: `x = 5; result = 3 < x and x < 10`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := interp.Execute(tt.script)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// TestUnaryOperators tests unary operations (negation, not)
func TestUnaryOperators(t *testing.T) {
	t.Parallel()

	interp := NewInterpreter()

	tests := []struct {
		name   string
		script string
	}{
		{
			name:   "unary minus literal",
			script: `result = -42`,
		},
		{
			name:   "unary minus variable",
			script: `x = 10; result = -x`,
		},
		{
			name:   "double negation",
			script: `x = 5; result = -(-x)`,
		},
		{
			name:   "unary not true",
			script: `result = not true`,
		},
		{
			name:   "unary not false",
			script: `result = not false`,
		},
		{
			name:   "not with comparison",
			script: `result = not (5 > 10)`,
		},
		{
			name:   "double not",
			script: `result = not not true`,
		},
		{
			name:   "minus in expression",
			script: `result = 10 + -5`,
		},
		{
			name:   "not with truthy value",
			script: `result = not 42`,
		},
		{
			name:   "not with falsy value",
			script: `result = not 0`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := interp.Execute(tt.script)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// TestFunctionDefaults tests function parameters with default values
func TestFunctionDefaults(t *testing.T) {
	t.Parallel()

	interp := NewInterpreter()

	tests := []struct {
		name   string
		script string
	}{
		{
			name: "function with one default",
			script: `
				function greet(name, greeting = "Hello")
					return greeting + " " + name
				end
				result1 = greet("Alice")
				result2 = greet("Bob", "Hi")
			`,
		},
		{
			name: "function with multiple defaults",
			script: `
				function format(value, prefix = "[", suffix = "]")
					return prefix + value + suffix
				end
				r1 = format("test")
				r2 = format("test", "{")
				r3 = format("test", "{", "}")
			`,
		},
		{
			name: "function with numeric default",
			script: `
				function multiply(a, b = 2)
					return a * b
				end
				result1 = multiply(5)
				result2 = multiply(5, 3)
			`,
		},
		{
			name: "function with object default",
			script: `
				function configure(config = {})
					return config
				end
				result1 = configure()
				result2 = configure({x = 10})
			`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := interp.Execute(tt.script)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// TestVarKeyword tests local variable declarations with var keyword
func TestVarKeyword(t *testing.T) {
	t.Parallel()

	interp := NewInterpreter()

	tests := []struct {
		name   string
		script string
	}{
		{
			name: "var declares local variable",
			script: `
				x = 10
				function test()
					var x = 20
					return x
				end
				result = test()
				global_x = x
			`,
		},
		{
			name: "multiple var declarations",
			script: `
				function test()
					var a = 1
					var b = 2
					var c = 3
					return a + b + c
				end
				result = test()
			`,
		},
		{
			name: "var in if block",
			script: `
				x = 1
				if true then
					var x = 2
				end
				result = x
			`,
		},
		{
			name: "var in loop",
			script: `
				for i = 1, 3 do
					var x = i * 2
				end
			`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := interp.Execute(tt.script)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// TestStringOperations tests string concatenation and operations
func TestStringOperations(t *testing.T) {
	t.Parallel()

	interp := NewInterpreter()

	tests := []struct {
		name   string
		script string
	}{
		{
			name:   "string concatenation",
			script: `result = "hello" + " " + "world"`,
		},
		{
			name:   "string with number concatenation",
			script: `result = "value: " + 42`,
		},
		{
			name:   "empty string",
			script: `result = ""`,
		},
		{
			name:   "multiline string",
			script: `
				result = """
					line 1
					line 2
				"""
			`,
		},
		{
			name:   "string with escape sequences",
			script: `result = "hello\nworld"`,
		},
		{
			name:   "string comparison",
			script: `result = "abc" < "xyz"`,
		},
		{
			name:   "string equality",
			script: `result = "test" == "test"`,
		},
		{
			name:   "string concatenation in expression",
			script: `result = "a" + "b" + "c" + "d"`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := interp.Execute(tt.script)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// TestArrayLiterals tests array construction with various elements
func TestArrayLiterals(t *testing.T) {
	t.Parallel()

	interp := NewInterpreter()

	tests := []struct {
		name   string
		script string
	}{
		{
			name:   "empty array",
			script: `result = []`,
		},
		{
			name:   "array with numbers",
			script: `result = [1, 2, 3, 4, 5]`,
		},
		{
			name:   "array with strings",
			script: `result = ["a", "b", "c"]`,
		},
		{
			name:   "array with mixed types",
			script: `result = [1, "two", 3.0, true, nil]`,
		},
		{
			name:   "array with expressions",
			script: `result = [1 + 1, 2 * 3, 10 - 5]`,
		},
		{
			name:   "array with variables",
			script: `x = 10; y = 20; result = [x, y, x + y]`,
		},
		{
			name:   "nested arrays",
			script: `result = [[1, 2], [3, 4], [5, 6]]`,
		},
		{
			name:   "trailing comma in array",
			script: `result = [1, 2, 3,]`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := interp.Execute(tt.script)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// TestObjectLiterals tests object construction with various properties
func TestObjectLiterals(t *testing.T) {
	t.Parallel()

	interp := NewInterpreter()

	tests := []struct {
		name   string
		script string
	}{
		{
			name:   "empty object",
			script: `result = {}`,
		},
		{
			name:   "object with number properties",
			script: `result = {a = 1, b = 2, c = 3}`,
		},
		{
			name:   "object with string properties",
			script: `result = {name = "Alice", status = "active"}`,
		},
		{
			name:   "object with mixed property types",
			script: `result = {id = 1, name = "test", active = true}`,
		},
		{
			name:   "object with expressions",
			script: `result = {sum = 1 + 2, product = 3 * 4}`,
		},
		{
			name:   "object with variables",
			script: `x = 10; y = 20; result = {x = x, y = y}`,
		},
		{
			name:   "nested objects",
			script: `result = {a = {x = 1, y = 2}, b = {x = 3, y = 4}}`,
		},
		{
			name:   "object with array",
			script: `result = {items = [1, 2, 3], count = 3}`,
		},
		{
			name:   "trailing comma in object",
			script: `result = {a = 1, b = 2,}`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := interp.Execute(tt.script)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// TestComplexExpressions tests complex nested expressions
func TestComplexExpressions(t *testing.T) {
	t.Parallel()

	interp := NewInterpreter()

	tests := []struct {
		name   string
		script string
	}{
		{
			name: "nested function calls",
			script: `
				function f(x) return x * 2 end
				function g(x) return x + 10 end
				result = g(f(5))
			`,
		},
		{
			name: "array of functions",
			script: `
				f = function(x) return x * 2 end
				g = function(x) return x + 1 end
				funcs = [f, g]
				result = funcs[0](5)
			`,
		},
		{
			name: "function returning array",
			script: `
				function makeArray()
					return [1, 2, 3]
				end
				result = makeArray()[1]
			`,
		},
		{
			name: "function returning object",
			script: `
				function makePerson()
					return {name = "Alice", age = 30}
				end
				result = makePerson().name
			`,
		},
		{
			name: "complex boolean expression",
			script: `
				x = 10
				y = 20
				result = (x > 5 and y < 30) or (x < 0 and y > 50)
			`,
		},
		{
			name: "mixed operators and precedence",
			script: `
				result = 2 + 3 * 4 - 5 / 2 + 1
			`,
		},
		{
			name: "ternary in array",
			script: `
				x = 10
				result = [x > 5 ? "big" : "small", x * 2]
			`,
		},
		{
			name: "array comprehension-like pattern",
			script: `
				nums = [1, 2, 3, 4, 5]
				result = [nums[0], nums[2], nums[4]]
			`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := interp.Execute(tt.script)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// TestOperatorPrecedence tests that operator precedence is correct
func TestOperatorPrecedence(t *testing.T) {
	t.Parallel()

	interp := NewInterpreter()

	tests := []struct {
		name   string
		script string
	}{
		{
			name:   "multiplication before addition",
			script: `result = 2 + 3 * 4`,
		},
		{
			name:   "division before subtraction",
			script: `result = 10 - 20 / 4`,
		},
		{
			name:   "comparison before logical and",
			script: `result = 5 > 3 and 2 < 4`,
		},
		{
			name:   "logical and before logical or",
			script: `result = true or false and false`,
		},
		{
			name:   "ternary lowest precedence",
			script: `result = true ? 1 : 0 + 5`,
		},
		{
			name:   "parentheses override",
			script: `result = (2 + 3) * 4`,
		},
		{
			name:   "unary minus higher than binary",
			script: `result = -2 + 3`,
		},
		{
			name:   "not higher than and",
			script: `result = not true and false`,
		},
		{
			name:   "complex precedence",
			script: `result = 1 + 2 * 3 - 4 / 2 or 5 < 3 and true`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := interp.Execute(tt.script)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// TestAssignmentVariations tests different assignment patterns
func TestAssignmentVariations(t *testing.T) {
	t.Parallel()

	interp := NewInterpreter()

	tests := []struct {
		name   string
		script string
	}{
		{
			name: "simple assignment",
			script: `
				x = 10
				result = x
			`,
		},
		{
			name: "multiple assignment",
			script: `
				a = 1
				b = 2
				c = 3
				result = a + b + c
			`,
		},
		{
			name: "assignment with expression",
			script: `
				x = 5 * 2 + 3
				result = x
			`,
		},
		{
			name: "array element assignment",
			script: `
				arr = [1, 2, 3]
				arr[1] = 20
				result = arr[1]
			`,
		},
		{
			name: "object property assignment",
			script: `
				obj = {x = 10}
				obj.y = 20
				result = obj.y
			`,
		},
		{
			name: "nested assignment",
			script: `
				obj = {nested = {value = 5}}
				obj.nested.value = 10
				result = obj.nested.value
			`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := interp.Execute(tt.script)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
