package script

import (
	"testing"
)

// Helper for evaluator tests
func execTest(t *testing.T, code string, expected string) {
	interp := NewInterpreter(false)
	output, err := interp.Execute(code)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

// TestEvaluator_ClosureVariableMutation tests that closures can mutate captured variables
func TestEvaluator_ClosureVariableMutation(t *testing.T) {
	code := `function makeCounter()
  var count = 0
  return function()
    count = count + 1
    return count
  end
end
c1 = makeCounter()
print(c1())
print(c1())
c2 = makeCounter()
print(c2())
print(c1())
`
	execTest(t, code, "1\n2\n1\n3\n")
}

// TestEvaluator_ClosureIndependence tests that different closures have independent captured vars
func TestEvaluator_ClosureIndependence(t *testing.T) {
	code := `function makeAdder(n)
  return function(x)
    return x + n
  end
end
add5 = makeAdder(5)
add10 = makeAdder(10)
print(add5(3))
print(add10(3))
`
	execTest(t, code, "8\n13\n")
}

// TestEvaluator_NestedClosures tests nested function closures
func TestEvaluator_NestedClosures(t *testing.T) {
	code := `function outer(a)
  return function(b)
    return function(c)
      return a + b + c
    end
  end
end
f = outer(1)
g = f(2)
print(g(3))
`
	execTest(t, code, "6\n")
}

// TestEvaluator_VarShadowing tests that var creates local shadow
func TestEvaluator_VarShadowing(t *testing.T) {
	code := `x = 10
function test()
  var x = 20
  print(x)
end
test()
print(x)
`
	execTest(t, code, "20\n10\n")
}

// TestEvaluator_GlobalModification tests modifying global from function
func TestEvaluator_GlobalModification(t *testing.T) {
	code := `x = 10
function increment()
  x = x + 1
end
increment()
increment()
print(x)
`
	execTest(t, code, "12\n")
}

// TestEvaluator_ArrayMutation tests that arrays are mutable
func TestEvaluator_ArrayMutation(t *testing.T) {
	code := `arr = [1, 2, 3]
arr[0] = 10
print(arr[0])
print(arr)
`
	execTest(t, code, "10\n[10 2 3]\n")
}

// TestEvaluator_ObjectMutation tests that objects are mutable
func TestEvaluator_ObjectMutation(t *testing.T) {
	code := `obj = {x = 1, y = 2}
obj.x = 10
print(obj.x)
obj["z"] = 3
print(obj["z"])
`
	execTest(t, code, "10\n3\n")
}

// TestEvaluator_MethodBinding tests that methods receive implicit this
func TestEvaluator_MethodBinding(t *testing.T) {
	code := `person = {
  name = "Alice",
  greet = function(msg)
    return msg + ", I am " + name
  end
}
print(person.greet("Hello"))
`
	execTest(t, code, "Hello, I am Alice\n")
}

// TestEvaluator_MethodMutation tests that methods can mutate object
func TestEvaluator_MethodMutation(t *testing.T) {
	code := `counter = {
  count = 0,
  increment = function()
    count = count + 1
  end
}
counter.increment()
counter.increment()
print(counter.count)
`
	execTest(t, code, "2\n")
}

// TestEvaluator_MethodCall tests that calling methods work correctly
func TestEvaluator_MethodCall(t *testing.T) {
	code := `obj = {
  value = 42,
  getValue = function()
    return value
  end
}
print(obj.getValue())
`
	execTest(t, code, "42\n")
}

// TestEvaluator_TypeCoercion tests implicit type conversions
func TestEvaluator_TypeCoercion(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"string concat with number", `print("value: " + 42)`, "value: 42\n"},
		{"add string to number coerces", `x = 5 + "hello"
print(x)`, "5hello\n"},
		{"logical short circuit and", `x = false and (1/0)
print("ok")`, "ok\n"},
		{"logical short circuit or", `x = true or (1/0)
print("ok")`, "ok\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execTest(t, tt.code, tt.expected)
		})
	}
}

// TestEvaluator_LoopVariableScope tests that loop variables are scoped to loop
func TestEvaluator_LoopVariableScope(t *testing.T) {
	code := `for i = 1, 2 do
  print(i)
end
print("done")
`
	execTest(t, code, "1\n2\ndone\n")
}

// TestEvaluator_MultipleReturnPaths tests multiple return statements
func TestEvaluator_MultipleReturnPaths(t *testing.T) {
	code := `function classify(x)
  if x < 0 then
    return "negative"
  elseif x == 0 then
    return "zero"
  else
    return "positive"
  end
end
print(classify(-5))
print(classify(0))
print(classify(5))
`
	execTest(t, code, "negative\nzero\npositive\n")
}

// TestEvaluator_EarlyReturn tests that return exits immediately
func TestEvaluator_EarlyReturn(t *testing.T) {
	code := `function test()
  print("a")
  return 5
  print("b")
end
print(test())
`
	execTest(t, code, "a\n5\n")
}

// TestEvaluator_BreakExitsImmediately tests break exits loop
func TestEvaluator_BreakExitsImmediately(t *testing.T) {
	code := `for i = 1, 10 do
  if i == 3 then break end
  print(i)
end
print("done")
`
	execTest(t, code, "1\n2\ndone\n")
}

// TestEvaluator_ContinueSkipsStatement tests continue skips to next iteration
func TestEvaluator_ContinueSkipsStatement(t *testing.T) {
	code := `for i = 1, 5 do
  if i == 2 then continue end
  if i == 4 then continue end
  print(i)
end
`
	execTest(t, code, "1\n3\n5\n")
}

// TestEvaluator_TernaryWithSideEffects tests ternary doesn't evaluate unused branch
func TestEvaluator_TernaryWithSideEffects(t *testing.T) {
	code := `x = true ? 1 : (1/0)
print("ok")
`
	execTest(t, code, "ok\n")
}

// TestEvaluator_ObjectAsConstructor tests object constructor pattern
func TestEvaluator_ObjectAsConstructor(t *testing.T) {
	code := `Config = {timeout = 30, retries = 3}
c1 = Config()
c1.timeout = 60
c2 = Config(timeout = 90)
print(c1.timeout)
print(c2.timeout)
print(Config.timeout)
`
	execTest(t, code, "60\n90\n30\n")
}

// TestEvaluator_RecursiveFunctions tests recursive function calls
func TestEvaluator_RecursiveFunctions(t *testing.T) {
	code := `function countdown(n)
  if n <= 0 then
    return
  end
  print(n)
  countdown(n - 1)
end
countdown(3)
`
	execTest(t, code, "3\n2\n1\n")
}

// TestEvaluator_MutualRecursion tests mutually recursive functions
func TestEvaluator_MutualRecursion(t *testing.T) {
	code := `function isEven(n)
  if n == 0 then return true end
  return isOdd(n - 1)
end
function isOdd(n)
  if n == 0 then return false end
  return isEven(n - 1)
end
print(isEven(4))
print(isOdd(4))
`
	execTest(t, code, "true\nfalse\n")
}

// TestEvaluator_FunctionParameterBinding tests parameter binding
func TestEvaluator_FunctionParameterBinding(t *testing.T) {
	code := `function test(a, b, c)
  print(a)
  print(b)
  print(c)
end
test(1, 2, 3)
`
	execTest(t, code, "1\n2\n3\n")
}

// TestEvaluator_NamedParameterBinding tests named parameter binding
func TestEvaluator_NamedParameterBinding(t *testing.T) {
	code := `function test(a, b, c)
  print(a)
  print(b)
  print(c)
end
test(c = 3, a = 1, b = 2)
`
	execTest(t, code, "1\n2\n3\n")
}

// TestEvaluator_MixedParameterBinding tests mixed positional and named
func TestEvaluator_MixedParameterBinding(t *testing.T) {
	code := `function test(a, b, c)
  print(a)
  print(b)
  print(c)
end
test(1, c = 3, b = 2)
`
	execTest(t, code, "1\n2\n3\n")
}

// TestEvaluator_DefaultParameters tests that missing args become nil
func TestEvaluator_DefaultParameters(t *testing.T) {
	code := `function test(a, b)
  print(a)
  print(b)
end
test(1)
`
	execTest(t, code, "1\n<nil>\n")
}

// TestEvaluator_ExtraArguments tests extra arguments are ignored
func TestEvaluator_ExtraArguments(t *testing.T) {
	code := `function test(a)
  print(a)
end
test(1, 2, 3)
print("ok")
`
	execTest(t, code, "1\nok\n")
}

// TestEvaluator_OperatorPrecedence tests operator precedence is correct
func TestEvaluator_OperatorPrecedence(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"mult before add", "print(2 + 3 * 4)", "14\n"},
		{"parens override", "print((2 + 3) * 4)", "20\n"},
		{"unary before binary", "print(-2 + 3)", "1\n"},
		{"comparison before logical", `print(2 < 3 and 4 < 5)`, "true\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execTest(t, tt.code, tt.expected)
		})
	}
}

// TestEvaluator_StringInterpolation tests template string evaluation
func TestEvaluator_StringInterpolation(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"simple var", `x = 42
print("value: {{x}}")`, "value: 42\n"},
		{"expression", `x = 5
print("doubled: {{x * 2}}")`, "doubled: 10\n"},
		{"function call", `print("length: {{len([1,2,3])}}")`, "length: 3\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execTest(t, tt.code, tt.expected)
		})
	}
}

// TestEvaluator_ControlFlowIntegration tests complex control flow scenarios
func TestEvaluator_ControlFlowIntegration(t *testing.T) {
	code := `result = ""
for i = 1, 5 do
  if i % 2 == 0 then
    result = result + "even "
  else
    result = result + "odd "
  end
end
print(result)
`
	execTest(t, code, "odd even odd even odd \n")
}

// TestEvaluator_NestedLoops tests nested loop behavior
func TestEvaluator_NestedLoops(t *testing.T) {
	code := `for i = 1, 2 do
  for j = 1, 2 do
    print(i + j)
  end
end
`
	execTest(t, code, "2\n3\n3\n4\n")
}

// TestEvaluator_NestedLoopsWithBreak tests break in nested loop
func TestEvaluator_NestedLoopsWithBreak(t *testing.T) {
	code := `for i = 1, 3 do
  for j = 1, 3 do
    if j == 2 then break end
    print(i + j)
  end
end
`
	execTest(t, code, "2\n3\n4\n")
}

// TestEvaluator_LoopWithFunction tests calling functions in loops
func TestEvaluator_LoopWithFunction(t *testing.T) {
	code := `function double(x)
  return x * 2
end
for i = 1, 3 do
  print(double(i))
end
`
	execTest(t, code, "2\n4\n6\n")
}
