package script

import (
	"strings"
	"testing"
)

// TestError_SyntaxErrors tests various syntax errors
func TestError_SyntaxErrors(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		shouldError bool
	}{
		{"missing closing paren", `print(5`, true},
		{"missing closing bracket", `arr = [1, 2, 3`, true},
		{"missing closing brace", `obj = {a = 1`, true},
		{"invalid syntax", `if x y then`, true},
		{"unexpected token", `print 5`, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interp := NewInterpreter(false)
			_, err := interp.Execute(tt.code)
			if (err != nil) != tt.shouldError {
				t.Errorf("expected error: %v, got error: %v", tt.shouldError, err != nil)
			}
		})
	}
}

// TestError_RuntimeErrors tests runtime errors
func TestError_RuntimeErrors(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		shouldError bool
	}{
		{"undefined variable", `print(undefined_var)`, true},
		{"call non-function", `x = 5
x()`, true},
		{"invalid type operation", `x = "hello" - 5`, true},
		{"array index out of bounds", `arr = [1, 2, 3]
print(arr[10])`, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interp := NewInterpreter(false)
			_, err := interp.Execute(tt.code)
			if (err != nil) != tt.shouldError {
				t.Errorf("expected error: %v, got error: %v", tt.shouldError, err != nil)
			}
		})
	}
}

// TestError_TryCatch tests try-catch error handling
func TestError_TryCatch(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"catch thrown error", `try
  throw("custom error")
catch (e)
  print("caught error")
end`, "caught error\n"},
		{"catch undefined", `try
  x = undefined_var
catch (e)
  print("caught error")
end`, "caught error\n"},
		{"no error in try", `try
  print("success")
catch (e)
  print("error")
end`, "success\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interp := NewInterpreter(false)
			_, err := interp.Execute(tt.code)
			if err != nil {
				t.Fatalf("execution error: %v", err)
			}
			output := interp.GetOutput()
			if output != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, output)
			}
		})
	}
}

// TestError_ControlFlowWithErrors tests error handling combined with control flow
func TestError_ControlFlowWithErrors(t *testing.T) {
	code := `for i = 1, 3 do
  try
    if i == 2 then
      throw("skip 2")
    end
    print(i)
  catch (e)
    print("error at " + i)
  end
end
`
	interp := NewInterpreter(false)
	_, err := interp.Execute(code)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	output := interp.GetOutput()
	expected := "1\nerror at 2\n3\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

// TestError_ErrorMessageFormat tests that error messages are formatted correctly
func TestError_ErrorMessageFormat(t *testing.T) {
	code := `undefined_var`
	interp := NewInterpreter(false)
	_, err := interp.Execute(code)
	if err == nil {
		t.Fatal("expected error but execution succeeded")
	}
	errMsg := err.Error()
	if !strings.Contains(errMsg, "undefined") || !strings.Contains(errMsg, "undefined_var") {
		t.Errorf("error message should mention undefined variable: %s", errMsg)
	}
}

// TestError_TypeErrors tests type-related errors
func TestError_TypeErrors(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		shouldError bool
	}{
		{"string minus number", `x = "hello" - 5`, true},
		{"array plus number", `x = [1, 2] + 5`, true},
		{"object minus string", `x = {} - "key"`, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interp := NewInterpreter(false)
			_, err := interp.Execute(tt.code)
			if (err != nil) != tt.shouldError {
				t.Errorf("expected error: %v, got error: %v", tt.shouldError, err != nil)
			}
		})
	}
}

// TestError_FunctionErrors tests function-related errors
func TestError_FunctionErrors(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		shouldError bool
	}{
		{"call non-function", `x = 5
result = x()`, true},
		{"call undefined function", `result = undefined_function()`, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interp := NewInterpreter(false)
			_, err := interp.Execute(tt.code)
			if (err != nil) != tt.shouldError {
				t.Errorf("expected error: %v, got error: %v", tt.shouldError, err != nil)
			}
		})
	}
}

// TestError_NestedTryCatch tests nested try-catch blocks
func TestError_NestedTryCatch(t *testing.T) {
	code := `try
  try
    throw("inner error")
  catch (e)
    print("inner caught")
  end
  throw("outer error")
catch (e)
  print("outer caught")
end
`
	interp := NewInterpreter(false)
	_, err := interp.Execute(code)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	output := interp.GetOutput()
	expected := "inner caught\nouter caught\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

// TestError_UncaughtError tests that uncaught errors propagate
func TestError_UncaughtError(t *testing.T) {
	code := `throw("uncaught error")`
	interp := NewInterpreter(false)
	_, err := interp.Execute(code)
	if err == nil {
		t.Fatal("expected error but execution succeeded")
	}
}

// TestError_BuiltinErrors tests errors from builtin functions
func TestError_BuiltinErrors(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		shouldError bool
	}{
		{"len on number", `x = len(42)`, true},
		{"substr wrong args", `x = substr()`, true},
		{"invalid increment", `x = "hello"
x = increment(x, 1)`, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interp := NewInterpreter(false)
			_, err := interp.Execute(tt.code)
			if (err != nil) != tt.shouldError {
				t.Errorf("expected error: %v, got error: %v", tt.shouldError, err != nil)
			}
		})
	}
}

// TestError_ZeroDivision tests division by zero
func TestError_ZeroDivision(t *testing.T) {
	code := `try
  x = 10 / 0
  print(x)
catch (e)
  print("caught division by zero")
end
`
	interp := NewInterpreter(false)
	_, err := interp.Execute(code)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	output := interp.GetOutput()
	if !strings.Contains(output, "caught") {
		t.Errorf("expected caught division by zero, got %q", output)
	}
}
