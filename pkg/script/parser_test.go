package script

import (
	"testing"
)

// Helper to parse Duso code and return the AST
func parseCode(source string) (*Program, error) {
	lexer := NewLexer(source)
	tokens := lexer.Tokenize()
	parser := NewParser(tokens)
	return parser.Parse()
}

// Helper to check if parsing succeeds
func expectParseSuccess(t *testing.T, code string) *Program {
	prog, err := parseCode(code)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if prog == nil {
		t.Fatal("expected non-nil program")
	}
	return prog
}

// Helper to check if parsing fails
func expectParseError(t *testing.T, code string) error {
	_, err := parseCode(code)
	if err == nil {
		t.Fatal("expected parse error but got none")
	}
	return err
}

// TestParser_Literals tests parsing of literal values
func TestParser_Literals(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"integer", "42"},
		{"float", "3.14"},
		{"string double quote", `"hello"`},
		{"string single quote", `'world'`},
		{"multiline string", `"""hello\nworld"""`},
		{"true", "true"},
		{"false", "false"},
		{"nil", "nil"},
		{"empty array", "[]"},
		{"array with elements", "[1, 2, 3]"},
		{"nested array", "[[1, 2], [3, 4]]"},
		{"empty object", "{}"},
		{"object with fields", `{x = 1, y = 2}`},
		{"object with string keys", `{"name" = "Alice"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectParseSuccess(t, tt.code)
		})
	}
}

// TestParser_BinaryOperations tests parsing binary expressions
func TestParser_BinaryOperations(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"addition", "1 + 2"},
		{"subtraction", "5 - 3"},
		{"multiplication", "4 * 3"},
		{"division", "10 / 2"},
		{"modulo", "10 % 3"},
		{"equality", "x == 5"},
		{"not equal", "x != 5"},
		{"less than", "x < 10"},
		{"greater than", "x > 5"},
		{"less or equal", "x <= 10"},
		{"greater or equal", "x >= 5"},
		{"logical and", "x and y"},
		{"logical or", "x or y"},
		{"chained", "1 + 2 * 3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectParseSuccess(t, tt.code)
		})
	}
}

// TestParser_UnaryOperations tests unary expressions
func TestParser_UnaryOperations(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"negative number", "-5"},
		{"logical not", "!true"},
		{"double negative", "--5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectParseSuccess(t, tt.code)
		})
	}
}

// TestParser_TernaryOperator tests ternary conditional expressions
func TestParser_TernaryOperator(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"simple ternary", "x > 5 ? 10 : 20"},
		{"nested ternary", "x > 10 ? 1 : x > 5 ? 2 : 3"},
		{"with function calls", `len(x) > 0 ? "has items" : "empty"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectParseSuccess(t, tt.code)
		})
	}
}

// TestParser_FunctionCalls tests parsing function calls
func TestParser_FunctionCalls(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"no args", "print()"},
		{"positional args", "print(1, 2, 3)"},
		{"string arg", `print("hello")`},
		{"named args", "func(x = 5, y = 10)"},
		{"mixed args", "func(1, y = 2)"},
		{"nested calls", "print(len([1, 2, 3]))"},
		{"chained calls", "map([1, 2], function(x) return x * 2 end)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectParseSuccess(t, tt.code)
		})
	}
}

// TestParser_PropertyAccess tests dot notation and indexing
func TestParser_PropertyAccess(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"dot notation", "obj.name"},
		{"nested dot", "obj.person.name"},
		{"array index", "arr[0]"},
		{"string index", `data["key"]`},
		{"computed index", "arr[i + 1]"},
		{"chained", "obj.arr[0].name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectParseSuccess(t, tt.code)
		})
	}
}

// TestParser_Assignments tests assignment statements
func TestParser_Assignments(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"simple", "x = 5"},
		{"string", `name = "Alice"`},
		{"expression", "x = 2 + 3"},
		{"function call", "result = add(1, 2)"},
		{"array", "nums = [1, 2, 3]"},
		{"object", "config = {timeout = 30}"},
		{"compound plus", "x += 5"},
		{"compound minus", "x -= 3"},
		{"compound multiply", "x *= 2"},
		{"compound divide", "x /= 2"},
		{"compound modulo", "x %= 3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectParseSuccess(t, tt.code)
		})
	}
}

// TestParser_IfStatements tests if/elseif/else parsing
func TestParser_IfStatements(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"simple if", `if x > 5 then
  print("yes")
end`},
		{"if with else", `if x > 5 then
  print("yes")
else
  print("no")
end`},
		{"if with elseif", `if x > 10 then
  print("big")
elseif x > 5 then
  print("medium")
else
  print("small")
end`},
		{"nested if", `if x then
  if y then
    print("both")
  end
end`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectParseSuccess(t, tt.code)
		})
	}
}

// TestParser_ForLoops tests for loop parsing
func TestParser_ForLoops(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"numeric for", `for i = 1, 5 do
  print(i)
end`},
		{"numeric with step", `for i = 1, 10, 2 do
  print(i)
end`},
		{"iterator for", `for item in items do
  print(item)
end`},
		{"break in loop", `for i = 1, 10 do
  if i == 5 then break end
  print(i)
end`},
		{"continue in loop", `for i = 1, 10 do
  if i == 5 then continue end
  print(i)
end`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectParseSuccess(t, tt.code)
		})
	}
}

// TestParser_WhileLoops tests while loop parsing
func TestParser_WhileLoops(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"basic while", `while x < 10 do
  x = x + 1
end`},
		{"while with break", `while true do
  if x > 5 then break end
  x = x + 1
end`},
		{"while with continue", `while x < 10 do
  x = x + 1
  if x == 5 then continue end
  print(x)
end`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectParseSuccess(t, tt.code)
		})
	}
}

// TestParser_FunctionDefinitions tests function parsing
func TestParser_FunctionDefinitions(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"no params", `function greet()
  return "hello"
end`},
		{"with params", `function add(a, b)
  return a + b
end`},
		{"single expression function", "double = function(x) return x * 2 end"},
		{"with body", `function calculate(x, y)
  result = x + y
  return result * 2
end`},
		{"nested function", `function outer()
  function inner()
    return 5
  end
  return inner()
end`},
		{"closure", `function makeAdder(n)
  return function(x)
    return x + n
  end
end`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectParseSuccess(t, tt.code)
		})
	}
}

// TestParser_TryCatchBlocks tests exception handling
func TestParser_TryCatchBlocks(t *testing.T) {
	tests := []struct {
		name string
		code string
		shouldParse bool
	}{
		{"try with catch", `try
  risky()
catch (error)
  print(error)
end`, true},
		{"nested try", `try
  try
    risky()
  catch (inner_error)
    print(inner_error)
  end
catch (outer_error)
  print(outer_error)
end`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldParse {
				expectParseSuccess(t, tt.code)
			} else {
				expectParseError(t, tt.code)
			}
		})
	}
}

// TestParser_ReturnStatements tests return parsing
func TestParser_ReturnStatements(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"return with value", `function test()
  return 42
end`},
		{"return expression", `function test()
  return x + y
end`},
		{"return nil implicit", `function test()
  x = 5
end`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectParseSuccess(t, tt.code)
		})
	}
}

// TestParser_VarDeclaration tests var keyword for local scope
func TestParser_VarDeclaration(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"var with value", "var x = 5"},
		{"var with expression", "var y = x + 1"},
		{"var shadowing", `function test()
  var x = 10
end`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectParseSuccess(t, tt.code)
		})
	}
}

// TestParser_TemplateStrings tests string template parsing
func TestParser_TemplateStrings(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"simple template", `msg = "value is {{x}}"`},
		{"expression in template", `msg = "result is {{2 + 3}}"`},
		{"nested in template", `msg = "outer {{inner {{x}}}}"`},
		{"multiple templates", `msg = "{{x}} and {{y}}"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectParseSuccess(t, tt.code)
		})
	}
}

// TestParser_MultipleStatements tests parsing multiple statements
func TestParser_MultipleStatements(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"two assignments", `x = 5
y = 10`},
		{"assignment then print", `name = "Alice"
print(name)`},
		{"mixed statements", `x = 5
if x > 3 then
  print("yes")
end
y = x + 1`},
		{"function then call", `function greet(name)
  return "Hello " + name
end
result = greet("World")`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectParseSuccess(t, tt.code)
		})
	}
}

// TestParser_ComplexExpressions tests complex nested expressions
func TestParser_ComplexExpressions(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"arithmetic precedence", "result = 2 + 3 * 4"},
		{"logical precedence", "result = true and false or true"},
		{"function in arithmetic", "result = len(arr) + 5"},
		{"method chaining", "result = filter(map(arr, fn1), fn2)"},
		{"ternary with operations", "result = x > 5 ? y * 2 : z / 2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectParseSuccess(t, tt.code)
		})
	}
}

// TestParser_SyntaxErrors tests that various invalid syntax is rejected
func TestParser_SyntaxErrors(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"unclosed paren", "print("},
		{"unclosed bracket", "[1, 2"},
		{"unclosed brace", "{x = 1"},
		{"missing end keyword", `if x then
  print("yes")`},
		{"missing function body", `function test()`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectParseError(t, tt.code)
		})
	}
}

// TestParser_EdgeCases tests edge cases and corner cases
func TestParser_EdgeCases(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"empty program", ""},
		{"single number", "42"},
		{"single string", `"hello"`},
		{"empty array access", "[][0]"},
		{"object with no value", `{x = , y = 2}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Some of these should succeed, some should fail
			// Just verify they don't crash
			parseCode(tt.code)
		})
	}
}
