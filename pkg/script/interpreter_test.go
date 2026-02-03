package script

import (
	"io"
	"os"
	"strings"
	"testing"
)

// Helper to execute code and capture stdout
func executeCode(t *testing.T, code string) string {
	// Capture stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	oldStdout := os.Stdout
	os.Stdout = w

	interp := NewInterpreter(false)
	_, execErr := interp.Execute(code)

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var output strings.Builder
	_, err = io.Copy(&output, r)
	r.Close()
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	if execErr != nil {
		t.Fatalf("execution error: %v", execErr)
	}

	return output.String()
}

// Helper to execute and expect success
func expectSuccess(t *testing.T, code string, expectedOutput string) {
	output := executeCode(t, code)
	if output != expectedOutput {
		t.Errorf("expected %q, got %q", expectedOutput, output)
	}
}

// Helper to execute and expect error
func expectError(t *testing.T, code string) error {
	interp := NewInterpreter(false)
	_, err := interp.Execute(code)
	if err == nil {
		t.Fatal("expected error but got none")
	}
	return err
}

// TestInterpreter_SimplePrint tests basic print() functionality
func TestInterpreter_SimplePrint(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"print string", `print("hello")`, "hello\n"},
		{"print number", "print(42)", "42\n"},
		{"print multiple args", `print("a", "b", "c")`, "a b c\n"},
		{"print boolean", "print(true)", "true\n"},
		{"print nil", "print(nil)", "nil\n"},
		{"print empty array", "print([])", "[]\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectSuccess(t, tt.code, tt.expected)
		})
	}
}

// TestInterpreter_Arithmetic tests arithmetic operations
func TestInterpreter_Arithmetic(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"addition", "print(2 + 3)", "5\n"},
		{"subtraction", "print(5 - 2)", "3\n"},
		{"multiplication", "print(4 * 3)", "12\n"},
		{"division", "print(10 / 2)", "5\n"},
		{"modulo", "print(10 % 3)", "1\n"},
		{"negative", "print(-5)", "-5\n"},
		{"precedence", "print(2 + 3 * 4)", "14\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectSuccess(t, tt.code, tt.expected)
		})
	}
}

// TestInterpreter_Variables tests variable assignment and retrieval
func TestInterpreter_Variables(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"simple assignment", `x = 5
print(x)`, "5\n"},
		{"string assignment", `name = "Alice"
print(name)`, "Alice\n"},
		{"reassignment", `x = 5
x = 10
print(x)`, "10\n"},
		{"arithmetic assignment", `x = 5
y = x + 3
print(y)`, "8\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectSuccess(t, tt.code, tt.expected)
		})
	}
}

// TestInterpreter_ControlFlow tests if/elseif/else
func TestInterpreter_ControlFlow(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"if true", `if true then
  print("yes")
end`, "yes\n"},
		{"if false", `if false then
  print("yes")
else
  print("no")
end`, "no\n"},
		{"elseif", `x = 5
if x > 10 then
  print("big")
elseif x > 3 then
  print("medium")
else
  print("small")
end`, "medium\n"},
		{"nested if", `x = 5
if x > 3 then
  if x < 10 then
    print("yes")
  end
end`, "yes\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectSuccess(t, tt.code, tt.expected)
		})
	}
}

// TestInterpreter_ForLoops tests for loop execution
func TestInterpreter_ForLoops(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"numeric for", `for i = 1, 3 do
  print(i)
end`, "1\n2\n3\n"},
		{"for with step", `for i = 1, 5, 2 do
  print(i)
end`, "1\n3\n5\n"},
		{"reverse for", `for i = 3, 1, -1 do
  print(i)
end`, "3\n2\n1\n"},
		{"iterator for", `items = ["a", "b", "c"]
for item in items do
  print(item)
end`, "a\nb\nc\n"},
		{"for with break", `for i = 1, 5 do
  if i == 3 then break end
  print(i)
end`, "1\n2\n"},
		{"for with continue", `for i = 1, 3 do
  if i == 2 then continue end
  print(i)
end`, "1\n3\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectSuccess(t, tt.code, tt.expected)
		})
	}
}

// TestInterpreter_WhileLoops tests while loop execution
func TestInterpreter_WhileLoops(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"basic while", `x = 1
while x <= 3 do
  print(x)
  x = x + 1
end`, "1\n2\n3\n"},
		{"while with break", `x = 0
while true do
  x = x + 1
  if x == 3 then break end
  print(x)
end`, "1\n2\n"},
		{"while with continue", `x = 0
while x < 3 do
  x = x + 1
  if x == 2 then continue end
  print(x)
end`, "1\n3\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectSuccess(t, tt.code, tt.expected)
		})
	}
}

// TestInterpreter_Functions tests function definition and calls
func TestInterpreter_Functions(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"simple function", `function greet(name)
  return "Hello " + name
end
print(greet("World"))`, "Hello World\n"},
		{"function with multiple params", `function add(a, b)
  return a + b
end
print(add(2, 3))`, "5\n"},
		{"function expression", `double = function(x) return x * 2 end
print(double(5))`, "10\n"},
		{"function with no explicit return", `function test()
  return 5
end
result = test()
print(result)`, "5\n"},
		{"nested function", `function outer()
  function inner()
    return 42
  end
  return inner()
end
print(outer())`, "42\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectSuccess(t, tt.code, tt.expected)
		})
	}
}

// TestInterpreter_Closures tests closure capture
func TestInterpreter_Closures(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"simple closure", `function makeAdder(n)
  return function(x) return x + n end
end
add5 = makeAdder(5)
print(add5(10))`, "15\n"},
		{"closure with multiple vars", `function makeCalc(a, b)
  return function(x)
    return x + a + b
  end
end
calc = makeCalc(2, 3)
print(calc(5))`, "10\n"},
		{"closure mutation", `function makeCounter()
  var count = 0
  return function()
    count = count + 1
    return count
  end
end
counter = makeCounter()
print(counter())
print(counter())
print(counter())`, "1\n2\n3\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectSuccess(t, tt.code, tt.expected)
		})
	}
}

// TestInterpreter_Arrays tests array operations
func TestInterpreter_Arrays(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"array literal", "print([1, 2, 3])", "[1, 2, 3]\n"},
		{"array indexing", "arr = [10, 20, 30]\nprint(arr[0])", "10\n"},
		{"array len", "print(len([1, 2, 3]))", "3\n"},
		{"array push", "arr = []\npush(arr, 1)\npush(arr, 2)\nprint(arr)", "[1, 2]\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectSuccess(t, tt.code, tt.expected)
		})
	}
}

// TestInterpreter_Objects tests object operations
func TestInterpreter_Objects(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"object literal", `obj = {x = 1, y = 2}
print(obj.x)`, "1\n"},
		{"object string key", `obj = {"name" = "Alice"}
print(obj["name"])`, "Alice\n"},
		{"object mutation", `obj = {x = 1}
obj.x = 5
print(obj.x)`, "5\n"},
		{"nested object", `obj = {inner = {value = 42}}
print(obj.inner.value)`, "42\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectSuccess(t, tt.code, tt.expected)
		})
	}
}

// TestInterpreter_StringOperations tests string concatenation and operations
func TestInterpreter_StringOperations(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"string concat", `print("Hello" + " " + "World")`, "Hello World\n"},
		{"string with numbers", `print("Value: " + 42)`, "Value: 42\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectSuccess(t, tt.code, tt.expected)
		})
	}
}

// TestInterpreter_TernaryOperator tests ternary conditional
func TestInterpreter_TernaryOperator(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"true condition", `x = 5
result = x > 3 ? "yes" : "no"
print(result)`, "yes\n"},
		{"false condition", `x = 2
result = x > 3 ? "yes" : "no"
print(result)`, "no\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectSuccess(t, tt.code, tt.expected)
		})
	}
}

// TestInterpreter_ErrorHandling tests try/catch blocks
func TestInterpreter_ErrorHandling(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"try without error", `try
  print("ok")
catch (e)
  print("error")
end`, "ok\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectSuccess(t, tt.code, tt.expected)
		})
	}
}

// TestInterpreter_Scoping tests variable scoping with var keyword
func TestInterpreter_Scoping(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"global scope", `x = 10
function test()
  x = x + 1
end
test()
print(x)`, "11\n"},
		{"local scope with var", `x = 10
function test()
  var x = 20
  print(x)
end
test()
print(x)`, "20\n10\n"},
		{"nested scopes", `x = 1
function outer()
  x = x + 1
  function inner()
    x = x + 1
  end
  inner()
end
outer()
print(x)`, "3\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectSuccess(t, tt.code, tt.expected)
		})
	}
}

// TestInterpreter_Comparison tests comparison operators
func TestInterpreter_Comparison(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"equal true", "print(5 == 5)", "true\n"},
		{"equal false", "print(5 == 6)", "false\n"},
		{"not equal", "print(5 != 6)", "true\n"},
		{"less than", "print(5 < 10)", "true\n"},
		{"greater than", "print(10 > 5)", "true\n"},
		{"less or equal", "print(5 <= 5)", "true\n"},
		{"greater or equal", "print(5 >= 5)", "true\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectSuccess(t, tt.code, tt.expected)
		})
	}
}

// TestInterpreter_LogicalOperators tests logical operations
func TestInterpreter_LogicalOperators(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"not true", "print(!true)", "false\n"},
		{"not false", "print(!false)", "true\n"},
		{"and in if", `if true and true then print("yes") else print("no") end`, "yes\n"},
		{"or in if", `if true or false then print("yes") else print("no") end`, "yes\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectSuccess(t, tt.code, tt.expected)
		})
	}
}

// TestInterpreter_Truthiness tests truthy/falsy values
func TestInterpreter_Truthiness(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"empty string is falsy", `if "" then
  print("yes")
else
  print("no")
end`, "no\n"},
		{"non-empty string is truthy", `if "hello" then
  print("yes")
else
  print("no")
end`, "yes\n"},
		{"zero is falsy", `if 0 then
  print("yes")
else
  print("no")
end`, "no\n"},
		{"non-zero is truthy", `if 5 then
  print("yes")
else
  print("no")
end`, "yes\n"},
		{"nil is falsy", `if nil then
  print("yes")
else
  print("no")
end`, "no\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectSuccess(t, tt.code, tt.expected)
		})
	}
}

// TestInterpreter_ComplexPrograms tests larger scripts
func TestInterpreter_ComplexPrograms(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"factorial", `function factorial(n)
  if n <= 1 then
    return 1
  else
    return n * factorial(n - 1)
  end
end
print(factorial(5))`, "120\n"},
		{"fibonacci", `function fib(n)
  if n <= 1 then
    return n
  else
    return fib(n - 1) + fib(n - 2)
  end
end
print(fib(6))`, "8\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectSuccess(t, tt.code, tt.expected)
		})
	}
}

// TestInterpreter_Errors tests that invalid operations fail appropriately
func TestInterpreter_Errors(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"undefined variable", "print(undefined_var)"},
		{"division by zero", "print(1 / 0)"},
		{"invalid syntax", "x = "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectError(t, tt.code)
		})
	}
}
