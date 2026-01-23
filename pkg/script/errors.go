package script

import "fmt"

// ScriptException represents an exception that occurs during script execution
type ScriptException struct {
	Message string
	Line    int
}

func (e *ScriptException) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("error at line %d: %s", e.Line, e.Message)
	}
	return fmt.Sprintf("error: %s", e.Message)
}

// RuntimeError represents a runtime error
type RuntimeError struct {
	Message string
	Line    int
}

func (e *RuntimeError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("runtime error at line %d: %s", e.Line, e.Message)
	}
	return fmt.Sprintf("runtime error: %s", e.Message)
}

// ReturnValue is used to signal a return from a function
type ReturnValue struct {
	Value Value
}

func (e *ReturnValue) Error() string {
	return "return"
}

// BreakIteration is used to signal a break from a loop (for future use)
type BreakIteration struct{}

func (e *BreakIteration) Error() string {
	return "break"
}

// ContinueIteration is used to signal a continue in a loop (for future use)
type ContinueIteration struct{}

func (e *ContinueIteration) Error() string {
	return "continue"
}

// ExitExecution is used to signal exit() with optional return values
type ExitExecution struct {
	Values []any
}

func (e *ExitExecution) Error() string {
	return "exit"
}
