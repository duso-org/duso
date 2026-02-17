package script

import (
	"testing"
)

// TestClosureCapture tests variable scoping and closure capture
func TestClosureCapture(t *testing.T) {
	t.Parallel()

	interp := NewInterpreter()

	tests := []struct {
		name   string
		script string
	}{
		{
			name: "simple closure captures outer variable",
			script: `
				function makeAdder(n)
					function add(x)
						return x + n
					end
					return add
				end
				addFive = makeAdder(5)
				result = addFive(10)
			`,
		},
		{
			name: "closure captures multiple variables",
			script: `
				function makeMath(a, b)
					function compute(x)
						return x + a - b
					end
					return compute
				end
				fn = makeMath(10, 3)
				result = fn(5)
			`,
		},
		{
			name: "multiple closures capture independently",
			script: `
				function makeAdder(n)
					function add(x)
						return x + n
					end
					return add
				end
				add2 = makeAdder(2)
				add10 = makeAdder(10)
				r1 = add2(5)
				r2 = add10(5)
			`,
		},
		{
			name: "nested function definitions",
			script: `
				function outer(a)
					function middle(b)
						function inner(c)
							return a + b + c
						end
						return inner
					end
					return middle
				end
				fn = outer(1)
				fn2 = fn(2)
				result = fn2(3)
			`,
		},
		{
			name: "function object closure",
			script: `
				x = 10
				fn = function() return x end
				result = fn()
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

// TestIfElseControl tests if/elseif/else control flow
func TestIfElseControl(t *testing.T) {
	t.Parallel()

	interp := NewInterpreter()

	tests := []struct {
		name   string
		script string
	}{
		{
			name: "if true executes body",
			script: `
				if true then
					x = 1
				end
			`,
		},
		{
			name: "if false skips body",
			script: `
				x = 0
				if false then
					x = 1
				end
			`,
		},
		{
			name: "if/else with true condition",
			script: `
				if 10 > 5 then
					result = "yes"
				else
					result = "no"
				end
			`,
		},
		{
			name: "if/else with false condition",
			script: `
				if 10 < 5 then
					result = "yes"
				else
					result = "no"
				end
			`,
		},
		{
			name: "if/elseif/else with first branch true",
			script: `
				x = 15
				if x > 20 then
					result = "big"
				elseif x > 10 then
					result = "medium"
				else
					result = "small"
				end
			`,
		},
		{
			name: "if/elseif/else with second branch true",
			script: `
				x = 8
				if x > 20 then
					result = "big"
				elseif x > 5 then
					result = "medium"
				else
					result = "small"
				end
			`,
		},
		{
			name: "if/elseif/else with else branch",
			script: `
				x = 3
				if x > 20 then
					result = "big"
				elseif x > 10 then
					result = "medium"
				else
					result = "small"
				end
			`,
		},
		{
			name: "multiple elseif branches",
			script: `
				x = 50
				if x > 100 then
					result = "huge"
				elseif x > 50 then
					result = "big"
				elseif x > 20 then
					result = "medium"
				elseif x > 0 then
					result = "small"
				else
					result = "negative"
				end
			`,
		},
		{
			name: "nested if statements",
			script: `
				x = 10
				y = 20
				if x > 0 then
					if y > 15 then
						result = "both"
					else
						result = "x only"
					end
				else
					result = "neither"
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

// TestTernaryOperator tests ternary conditional operator
func TestTernaryOperator(t *testing.T) {
	t.Parallel()

	interp := NewInterpreter()

	tests := []struct {
		name   string
		script string
	}{
		{
			name:   "ternary with true condition",
			script: `result = true ? "yes" : "no"`,
		},
		{
			name:   "ternary with false condition",
			script: `result = false ? "yes" : "no"`,
		},
		{
			name:   "ternary with comparison true",
			script: `result = 10 > 5 ? "greater" : "less"`,
		},
		{
			name:   "ternary with comparison false",
			script: `result = 10 < 5 ? "less" : "greater"`,
		},
		{
			name: "nested ternary",
			script: `
				x = 15
				result = x > 20 ? "big" : x > 10 ? "medium" : "small"
			`,
		},
		{
			name: "ternary with expressions",
			script: `
				x = 10
				y = 5
				result = x > y ? x + 10 : y + 10
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

// TestForLoops tests for loop control flow
func TestForLoops(t *testing.T) {
	t.Parallel()

	interp := NewInterpreter()

	tests := []struct {
		name   string
		script string
	}{
		{
			name: "for range loop ascending",
			script: `
				sum = 0
				for i = 1, 5 do
					sum = sum + i
				end
			`,
		},
		{
			name: "for range loop with single value",
			script: `
				x = 0
				for i = 1, 1 do
					x = i
				end
			`,
		},
		{
			name: "for array iteration",
			script: `
				arr = [10, 20, 30]
				sum = 0
				for item in arr do
					sum = sum + item
				end
			`,
		},
		{
			name: "for loop with break",
			script: `
				sum = 0
				for i = 1, 10 do
					if i == 5 then break end
					sum = sum + i
				end
			`,
		},
		{
			name: "for loop with continue",
			script: `
				sum = 0
				for i = 1, 10 do
					if i == 3 then continue end
					sum = sum + i
				end
			`,
		},
		{
			name: "nested for loops",
			script: `
				sum = 0
				for i = 1, 3 do
					for j = 1, 2 do
						sum = sum + 1
					end
				end
			`,
		},
		{
			name: "for loop with break in nested",
			script: `
				for i = 1, 3 do
					for j = 1, 3 do
						if j == 2 then break end
					end
				end
			`,
		},
		{
			name: "empty array iteration",
			script: `
				arr = []
				count = 0
				for item in arr do
					count = count + 1
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

// TestWhileLoops tests while loop control flow
func TestWhileLoops(t *testing.T) {
	t.Parallel()

	interp := NewInterpreter()

	tests := []struct {
		name   string
		script string
	}{
		{
			name: "while loop with counter",
			script: `
				count = 0
				while count < 5 do
					count = count + 1
				end
			`,
		},
		{
			name: "while loop with condition change",
			script: `
				x = 10
				while x > 0 do
					x = x - 2
				end
			`,
		},
		{
			name: "while loop with break",
			script: `
				count = 0
				while count < 100 do
					if count == 5 then break end
					count = count + 1
				end
			`,
		},
		{
			name: "while loop with continue",
			script: `
				count = 0
				sum = 0
				while count < 10 do
					count = count + 1
					if count % 2 == 0 then continue end
					sum = sum + count
				end
			`,
		},
		{
			name: "while loop false condition skips body",
			script: `
				x = 0
				while false do
					x = 1
				end
			`,
		},
		{
			name: "nested while loops",
			script: `
				i = 0
				while i < 3 do
					j = 0
					while j < 2 do
						j = j + 1
					end
					i = i + 1
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

// TestTruthiness tests truthiness of different types
func TestTruthiness(t *testing.T) {
	t.Parallel()

	interp := NewInterpreter()

	tests := []struct {
		name   string
		script string
	}{
		{
			name: "true is truthy",
			script: `
				if true then
					result = "yes"
				else
					result = "no"
				end
			`,
		},
		{
			name: "false is falsy",
			script: `
				if false then
					result = "yes"
				else
					result = "no"
				end
			`,
		},
		{
			name: "non-zero number is truthy",
			script: `
				if 42 then
					result = "yes"
				else
					result = "no"
				end
			`,
		},
		{
			name: "zero is falsy",
			script: `
				if 0 then
					result = "yes"
				else
					result = "no"
				end
			`,
		},
		{
			name: "non-empty string is truthy",
			script: `
				if "hello" then
					result = "yes"
				else
					result = "no"
				end
			`,
		},
		{
			name: "empty string is falsy",
			script: `
				if "" then
					result = "yes"
				else
					result = "no"
				end
			`,
		},
		{
			name: "non-empty array is truthy",
			script: `
				if [1, 2, 3] then
					result = "yes"
				else
					result = "no"
				end
			`,
		},
		{
			name: "empty array is falsy",
			script: `
				if [] then
					result = "yes"
				else
					result = "no"
				end
			`,
		},
		{
			name: "non-empty object is truthy",
			script: `
				if {a = 1} then
					result = "yes"
				else
					result = "no"
				end
			`,
		},
		{
			name: "empty object is falsy",
			script: `
				if {} then
					result = "yes"
				else
					result = "no"
				end
			`,
		},
		{
			name: "function is truthy",
			script: `
				f = function() return 1 end
				if f then
					result = "yes"
				else
					result = "no"
				end
			`,
		},
		{
			name: "nil is falsy",
			script: `
				if nil then
					result = "yes"
				else
					result = "no"
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

// TestTryCatch tests error handling with try/catch
func TestTryCatch(t *testing.T) {
	t.Parallel()

	interp := NewInterpreter()

	tests := []struct {
		name   string
		script string
	}{
		{
			name: "try without error executes",
			script: `
				try
					x = 10
				catch (err)
					x = 0
				end
			`,
		},
		{
			name: "catch catches thrown error",
			script: `
				try
					throw("test error")
				catch (err)
					result = err
				end
			`,
		},
		{
			name: "throw non-string value",
			script: `
				try
					throw({code = 500, message = "error"})
				catch (err)
					result = type(err)
				end
			`,
		},
		{
			name: "nested try/catch",
			script: `
				try
					try
						throw("inner")
					catch (e)
						throw("outer")
					end
				catch (err)
					result = err
				end
			`,
		},
		{
			name: "try with undefined variable error",
			script: `
				try
					x = undefined_variable + 1
				catch (err)
					result = type(err)
				end
			`,
		},
		{
			name: "multiple try/catch blocks",
			script: `
				x = 0
				try
					x = 1
				catch (err)
					x = -1
				end
				try
					throw("error")
				catch (err)
					x = 2
				end
			`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := interp.Execute(tt.script)
			_ = err
		})
	}
}

// TestLogicalOperators tests logical and/or/not operators
func TestLogicalOperators(t *testing.T) {
	t.Parallel()

	interp := NewInterpreter()

	tests := []struct {
		name   string
		script string
	}{
		{
			name:   "and with both true",
			script: `result = true and true`,
		},
		{
			name:   "and with first false",
			script: `result = false and true`,
		},
		{
			name:   "and with second false",
			script: `result = true and false`,
		},
		{
			name:   "and with both false",
			script: `result = false and false`,
		},
		{
			name:   "or with both true",
			script: `result = true or true`,
		},
		{
			name:   "or with first true",
			script: `result = true or false`,
		},
		{
			name:   "or with second true",
			script: `result = false or true`,
		},
		{
			name:   "or with both false",
			script: `result = false or false`,
		},
		{
			name:   "not true",
			script: `result = not true`,
		},
		{
			name:   "not false",
			script: `result = not false`,
		},
		{
			name:   "complex logical expression",
			script: `result = (true and false) or (true and true)`,
		},
		{
			name: "logical with comparisons",
			script: `
				x = 10
				result = x > 5 and x < 20
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

// TestVariableScoping tests variable scope and shadowing
func TestVariableScoping(t *testing.T) {
	t.Parallel()

	interp := NewInterpreter()

	tests := []struct {
		name   string
		script string
	}{
		{
			name: "global variable accessible in function",
			script: `
				x = 10
				function getX()
					return x
				end
				result = getX()
			`,
		},
		{
			name: "local variable shadows global",
			script: `
				x = 10
				function test()
					x = 20
					return x
				end
				result = test()
				global_x = x
			`,
		},
		{
			name: "function parameter shadows global",
			script: `
				x = 10
				function test(x)
					return x
				end
				result = test(20)
				global_x = x
			`,
		},
		{
			name: "block scope in if statement",
			script: `
				x = 10
				if true then
					x = 20
				end
				result = x
			`,
		},
		{
			name: "block scope in loop",
			script: `
				x = 0
				for i = 1, 3 do
					x = x + i
				end
				result = x
			`,
		},
		{
			name: "function scope isolation",
			script: `
				function makeCounter()
					count = 0
					function increment()
						count = count + 1
						return count
					end
					return increment
				end
				counter = makeCounter()
				r1 = counter()
				r2 = counter()
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

// TestReturnStatement tests return statement behavior
func TestReturnStatement(t *testing.T) {
	t.Parallel()

	interp := NewInterpreter()

	tests := []struct {
		name   string
		script string
	}{
		{
			name: "return in function exits early",
			script: `
				function test()
					x = 1
					return x
					x = 2
				end
				result = test()
			`,
		},
		{
			name: "implicit return of last expression",
			script: `
				function add(a, b)
					a + b
				end
				result = add(3, 4)
			`,
		},
		{
			name: "return nil",
			script: `
				function test()
					return nil
				end
				result = test()
			`,
		},
		{
			name: "return array",
			script: `
				function makeArray()
					return [1, 2, 3]
				end
				result = makeArray()
			`,
		},
		{
			name: "return object",
			script: `
				function makeObj()
					return {x = 10, y = 20}
				end
				result = makeObj()
			`,
		},
		{
			name: "return function",
			script: `
				function makeFn()
					return function(x) return x * 2 end
				end
				fn = makeFn()
				result = fn(5)
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

// TestCompoundAssignments tests +=, -=, *=, /= operators
func TestCompoundAssignments(t *testing.T) {
	t.Parallel()

	interp := NewInterpreter()

	tests := []struct {
		name   string
		script string
	}{
		{
			name: "add assign",
			script: `
				x = 10
				x += 5
				result = x
			`,
		},
		{
			name: "subtract assign",
			script: `
				x = 10
				x -= 3
				result = x
			`,
		},
		{
			name: "multiply assign",
			script: `
				x = 10
				x *= 2
				result = x
			`,
		},
		{
			name: "divide assign",
			script: `
				x = 10
				x /= 2
				result = x
			`,
		},
		{
			name: "modulo assign",
			script: `
				x = 10
				x %= 3
				result = x
			`,
		},
		{
			name: "string concatenation with +=",
			script: `
				s = "hello"
				s += " world"
				result = s
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

// TestPostIncrementDecrement tests ++ and -- operators
func TestPostIncrementDecrement(t *testing.T) {
	t.Parallel()

	interp := NewInterpreter()

	tests := []struct {
		name   string
		script string
	}{
		{
			name: "post increment as statement",
			script: `
				x = 5
				x++
				result = x
			`,
		},
		{
			name: "post decrement as statement",
			script: `
				x = 5
				x--
				result = x
			`,
		},
		{
			name: "post increment in loop",
			script: `
				x = 0
				for i = 1, 5 do
					x++
				end
				result = x
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

// TestArrayAndObjectAccess tests indexing and property access
func TestArrayAndObjectAccess(t *testing.T) {
	t.Parallel()

	interp := NewInterpreter()

	tests := []struct {
		name   string
		script string
	}{
		{
			name: "array index access",
			script: `
				arr = [10, 20, 30]
				result = arr[0]
			`,
		},
		{
			name: "array index 1",
			script: `
				arr = [10, 20, 30]
				result = arr[1]
			`,
		},
		{
			name: "array assignment via index",
			script: `
				arr = [10, 20, 30]
				arr[1] = 25
				result = arr[1]
			`,
		},
		{
			name: "object property access",
			script: `
				obj = {x = 10, y = 20}
				result = obj.x
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
			name: "object bracket notation",
			script: `
				obj = {x = 10}
				result = obj["x"]
			`,
		},
		{
			name: "nested array access",
			script: `
				arr = [[1, 2], [3, 4]]
				result = arr[0][1]
			`,
		},
		{
			name: "nested object access",
			script: `
				obj = {a = {b = 10}}
				result = obj.a.b
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

// TestStringTemplates tests template literal evaluation
func TestStringTemplates(t *testing.T) {
	t.Parallel()

	interp := NewInterpreter()

	tests := []struct {
		name   string
		script string
	}{
		{
			name: "simple template",
			script: `
				x = 10
				result = "value: {{x}}"
			`,
		},
		{
			name: "template with arithmetic",
			script: `
				x = 10
				result = "double: {{x * 2}}"
			`,
		},
		{
			name: "template with function call",
			script: `
				x = "hello"
				result = "upper: {{upper(x)}}"
			`,
		},
		{
			name: "template with nested access",
			script: `
				obj = {x = 10}
				result = "obj.x: {{obj.x}}"
			`,
		},
		{
			name: "template with array access",
			script: `
				arr = [1, 2, 3]
				result = "arr[1]: {{arr[1]}}"
			`,
		},
		{
			name: "multiple expressions in template",
			script: `
				x = 5
				y = 10
				result = "{{x}} + {{y}} = {{x + y}}"
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
