package script

import (
	"strings"
	"testing"
)

// TestDusoError_Error tests the DusoError.Error() method
func TestDusoError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *DusoError
		contains []string
		notContains []string
	}{
		{
			name: "basic error with file and position",
			err: &DusoError{
				Message:  "undefined variable",
				FilePath: "test.duso",
				Position: Position{Line: 5, Column: 10},
			},
			contains: []string{"test.duso", "5", "10", "undefined variable"},
		},
		{
			name: "error with call stack",
			err: &DusoError{
				Message:  "runtime error",
				FilePath: "main.duso",
				Position: Position{Line: 20, Column: 5},
				CallStack: []CallFrame{
					{
						FunctionName: "processData",
						FilePath:     "main.duso",
						Position:     Position{Line: 15, Column: 2},
					},
					{
						FunctionName: "main",
						FilePath:     "main.duso",
						Position:     Position{Line: 1, Column: 1},
					},
				},
			},
			contains: []string{"main.duso", "20", "5", "runtime error", "Call stack", "processData", "15", "2", "main", "1", "1"},
		},
		{
			name: "error without file path",
			err: &DusoError{
				Message:  "error message",
				Position: Position{Line: 10, Column: 0},
			},
			contains: []string{"10", "error message"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			errStr := tc.err.Error()

			// Check for required strings
			for _, want := range tc.contains {
				if !strings.Contains(errStr, want) {
					t.Errorf("Error() output missing %q.\nGot: %s", want, errStr)
				}
			}

			// Check that unwanted strings are not present
			for _, notWant := range tc.notContains {
				if strings.Contains(errStr, notWant) {
					t.Errorf("Error() output should not contain %q.\nGot: %s", notWant, errStr)
				}
			}
		})
	}
}

// TestReturnValue_Error tests the ReturnValue.Error() method
func TestReturnValue_Error(t *testing.T) {
	retVal := &ReturnValue{
		Value: NewNumber(42.0),
	}

	if retVal.Error() != "return" {
		t.Errorf("Expected 'return', got %q", retVal.Error())
	}

	if retVal.Value.AsNumber() != 42.0 {
		t.Errorf("Expected value 42.0, got %v", retVal.Value.AsNumber())
	}
}

// TestBreakIteration_Error tests the BreakIteration.Error() method
func TestBreakIteration_Error(t *testing.T) {
	breakIter := &BreakIteration{}

	if breakIter.Error() != "break" {
		t.Errorf("Expected 'break', got %q", breakIter.Error())
	}
}

// TestContinueIteration_Error tests the ContinueIteration.Error() method
func TestContinueIteration_Error(t *testing.T) {
	contIter := &ContinueIteration{}

	if contIter.Error() != "continue" {
		t.Errorf("Expected 'continue', got %q", contIter.Error())
	}
}

// TestExitExecution_Error tests the ExitExecution.Error() method
func TestExitExecution_Error(t *testing.T) {
	exitExec := &ExitExecution{
		Values: []any{1.0, "success"},
	}

	if exitExec.Error() != "exit" {
		t.Errorf("Expected 'exit', got %q", exitExec.Error())
	}

	if len(exitExec.Values) != 2 {
		t.Errorf("Expected 2 values, got %d", len(exitExec.Values))
	}

	if exitExec.Values[0] != 1.0 {
		t.Errorf("Expected first value 1.0, got %v", exitExec.Values[0])
	}

	if exitExec.Values[1] != "success" {
		t.Errorf("Expected second value 'success', got %v", exitExec.Values[1])
	}
}

// TestBreakpointError_Error tests the BreakpointError.Error() method
func TestBreakpointError_Error(t *testing.T) {
	bpErr := &BreakpointError{
		FilePath: "debug.duso",
		Position: Position{Line: 25, Column: 8},
		Message:  "breakpoint hit",
		CallStack: []CallFrame{
			{
				FunctionName: "debugFunc",
				FilePath:     "debug.duso",
				Position:     Position{Line: 20, Column: 4},
			},
		},
	}

	if bpErr.Error() != "breakpoint" {
		t.Errorf("Expected 'breakpoint', got %q", bpErr.Error())
	}

	if bpErr.FilePath != "debug.duso" {
		t.Errorf("Expected FilePath 'debug.duso', got %q", bpErr.FilePath)
	}

	if bpErr.Message != "breakpoint hit" {
		t.Errorf("Expected Message 'breakpoint hit', got %q", bpErr.Message)
	}

	if len(bpErr.CallStack) != 1 {
		t.Errorf("Expected 1 call frame, got %d", len(bpErr.CallStack))
	}
}

// TestCallFrame_Structure tests the CallFrame structure
func TestCallFrame_Structure(t *testing.T) {
	frame := CallFrame{
		FunctionName: "testFunc",
		FilePath:     "test.duso",
		Position:     Position{Line: 10, Column: 5},
	}

	if frame.FunctionName != "testFunc" {
		t.Errorf("Expected FunctionName 'testFunc', got %q", frame.FunctionName)
	}

	if frame.FilePath != "test.duso" {
		t.Errorf("Expected FilePath 'test.duso', got %q", frame.FilePath)
	}

	if frame.Position.Line != 10 {
		t.Errorf("Expected Line 10, got %d", frame.Position.Line)
	}

	if frame.Position.Column != 5 {
		t.Errorf("Expected Column 5, got %d", frame.Position.Column)
	}
}

// TestPosition_IsValid tests the Position.IsValid() method
func TestPosition_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		pos     Position
		isValid bool
	}{
		{
			name:    "valid position with line and column",
			pos:     Position{Line: 5, Column: 10},
			isValid: true,
		},
		{
			name:    "valid position with line only",
			pos:     Position{Line: 5, Column: 0},
			isValid: true,
		},
		{
			name:    "invalid position with zero line",
			pos:     Position{Line: 0, Column: 0},
			isValid: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.pos.IsValid() != tc.isValid {
				t.Errorf("Position.IsValid() for %v: expected %v, got %v",
					tc.pos, tc.isValid, tc.pos.IsValid())
			}
		})
	}
}
