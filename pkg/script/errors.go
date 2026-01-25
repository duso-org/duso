package script

import (
	"strconv"
	"strings"
)

// DusoError represents an error with position information and call stack
type DusoError struct {
	Message   string
	FilePath  string
	Position  Position
	CallStack []CallFrame
}

func (e *DusoError) Error() string {
	var buf strings.Builder

	// Format: "file:line:col: message"
	if e.FilePath != "" {
		buf.WriteString(e.FilePath)
		buf.WriteByte(':')
	}

	if e.Position.IsValid() {
		buf.WriteString(strconv.Itoa(e.Position.Line))
		if e.Position.Column > 0 {
			buf.WriteByte(':')
			buf.WriteString(strconv.Itoa(e.Position.Column))
		}
		buf.WriteString(": ")
	}

	buf.WriteString(e.Message)

	// Add call stack if present
	if len(e.CallStack) > 0 {
		buf.WriteString("\n\nCall stack:")
		// Print in reverse order (most recent call last)
		for i := len(e.CallStack) - 1; i >= 0; i-- {
			frame := e.CallStack[i]
			buf.WriteString("\n  at ")
			buf.WriteString(frame.FunctionName)
			buf.WriteString(" (")
			if frame.FilePath != "" {
				buf.WriteString(frame.FilePath)
				buf.WriteByte(':')
			}
			buf.WriteString(strconv.Itoa(frame.Position.Line))
			if frame.Position.Column > 0 {
				buf.WriteByte(':')
				buf.WriteString(strconv.Itoa(frame.Position.Column))
			}
			buf.WriteByte(')')
		}
	}

	return buf.String()
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

// BreakpointError signals debug breakpoint hit and captures call stack for display
type BreakpointError struct {
	FilePath  string
	Position  Position
	CallStack []CallFrame
	Env       *Environment // Current environment at breakpoint for scope access
}

func (e *BreakpointError) Error() string {
	return "breakpoint"
}
