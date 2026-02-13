package script

import (
	"io"
	"os"
	"strings"
	"testing"
)

// Helper to execute Duso code and capture stdout
func captureOutput(t *testing.T, code string) (string, error) {
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

	return output.String(), execErr
}

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
			output, err := captureOutput(t, tt.code)
			if err != nil {
				t.Fatalf("execution error: %v", err)
			}
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
	output, err := captureOutput(t, code)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
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
	output, err := captureOutput(t, code)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
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
	output, err := captureOutput(t, code)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if !strings.Contains(output, "caught") {
		t.Errorf("expected caught division by zero, got %q", output)
	}
}

// TestError_ThrowObject tests throwing and catching objects
func TestError_ThrowObject(t *testing.T) {
	code := `try
  throw({code = "ERR_001", message = "custom error", level = 5})
catch (e)
  print("caught error object")
end
`
	output, err := captureOutput(t, code)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if output != "caught error object\n" {
		t.Errorf("expected 'caught error object\\n', got %q", output)
	}
}

// TestError_ThrowObjectWithPropertyAccess tests accessing properties of thrown object
func TestError_ThrowObjectWithPropertyAccess(t *testing.T) {
	code := `try
  throw({code = "ERR_404", status = 404})
catch (e)
  if e.code then
    print(e.code)
  end
end
`
	output, err := captureOutput(t, code)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if !strings.Contains(output, "ERR_404") {
		t.Errorf("expected 'ERR_404' in output, got %q", output)
	}
}

// TestError_ThrowArray tests throwing arrays
func TestError_ThrowArray(t *testing.T) {
	code := `try
  throw([1, 2, 3, "error"])
catch (e)
  print("caught array")
end
`
	output, err := captureOutput(t, code)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if output != "caught array\n" {
		t.Errorf("expected 'caught array\\n', got %q", output)
	}
}

// TestError_ThrowNumber tests throwing numbers
func TestError_ThrowNumber(t *testing.T) {
	code := `try
  throw(42)
catch (e)
  print("caught number")
end
`
	output, err := captureOutput(t, code)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if output != "caught number\n" {
		t.Errorf("expected 'caught number\\n', got %q", output)
	}
}

// TestError_ThrowBoolean tests throwing booleans
func TestError_ThrowBoolean(t *testing.T) {
	code := `try
  throw(true)
catch (e)
  print("caught boolean")
end
`
	output, err := captureOutput(t, code)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if output != "caught boolean\n" {
		t.Errorf("expected 'caught boolean\\n', got %q", output)
	}
}

// TestError_ThrowNil tests throwing nil/null
func TestError_ThrowNil(t *testing.T) {
	code := `try
  throw(nil)
catch (e)
  print("caught nil")
end
`
	output, err := captureOutput(t, code)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if output != "caught nil\n" {
		t.Errorf("expected 'caught nil\\n', got %q", output)
	}
}

// TestError_ThrowComplexObject tests throwing complex nested objects
func TestError_ThrowComplexObject(t *testing.T) {
	code := `try
  throw({
    error = "validation",
    details = {
      field = "email",
      reason = "invalid format"
    },
    codes = [400, 422]
  })
catch (e)
  if e.details.field then
    print(e.details.field)
  end
end
`
	output, err := captureOutput(t, code)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if !strings.Contains(output, "email") {
		t.Errorf("expected 'email' in output, got %q", output)
	}
}

// TestError_ThrownObjectPreservedsInCatch tests that thrown object is accessible in catch block
func TestError_ThrownObjectPreservedsInCatch(t *testing.T) {
	code := `obj = {id = 123, name = "test"}
try
  throw(obj)
catch (e)
  print(e.id)
  print(e.name)
end
`
	output, err := captureOutput(t, code)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	// Should print both id and name
	if !strings.Contains(output, "123") || !strings.Contains(output, "test") {
		t.Errorf("expected '123' and 'test' in output, got %q", output)
	}
}

// TestError_DefaultErrorMessage tests that no argument to throw() defaults to "unknown error"
func TestError_DefaultErrorMessage(t *testing.T) {
	code := `try
  throw()
catch (e)
  print("caught default error")
end
`
	output, err := captureOutput(t, code)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if output != "caught default error\n" {
		t.Errorf("expected 'caught default error\\n', got %q", output)
	}
}
