package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/duso-org/duso/pkg/script"
)

// isExpressionStatement checks if a parsed program is a single expression statement
func isExpressionStatement(program *script.Program) script.Node {
	if program == nil || len(program.Statements) != 1 {
		return nil
	}

	stmt := program.Statements[0]

	// Check if the statement is an expression node type that should be auto-printed
	switch stmt.(type) {
	case *script.BinaryExpr, *script.TernaryExpr, *script.UnaryExpr,
		*script.CallExpr, *script.IndexExpr, *script.PropertyAccess,
		*script.Identifier, *script.NumberLiteral, *script.StringLiteral,
		*script.BoolLiteral, *script.ArrayLiteral, *script.ObjectLiteral,
		*script.TemplateLiteral, *script.FunctionExpr:
		return stmt
	}

	return nil
}

// wrapExpressionWithPrint wraps a bare expression with print() for REPL auto-output
func wrapExpressionWithPrint(code string) string {
	// Try to parse the code
	lexer := script.NewLexer(code)
	tokens := lexer.Tokenize()

	parser := script.NewParser(tokens)
	program, err := parser.Parse()
	if err != nil {
		// If parsing fails, return original code
		return code
	}

	// Check if it's a single expression statement
	if isExpressionStatement(program) != nil {
		return "print(" + code + ")"
	}

	return code
}

// NewConsoleDebugHandler creates a debug event handler for console-based debugging.
// It displays debug information to stderr and opens an interactive REPL for inspection.
// The interpreter is needed to access the debug session mutex to serialize REPL access.
func NewConsoleDebugHandler(interp *script.Interpreter) script.DebugHandler {
	return func(event *script.DebugEvent) {
		handleConsoleDebugEvent(interp, event)
	}
}

// handleConsoleDebugEvent processes a debug event and displays it to the user
func handleConsoleDebugEvent(interp *script.Interpreter, event *script.DebugEvent) {
	// Format location info
	loc := event.FilePath
	if event.Position.Line > 0 {
		loc = fmt.Sprintf("%s:%d", loc, event.Position.Line)
		if event.Position.Column > 0 {
			loc = fmt.Sprintf("%s:%d", loc, event.Position.Column)
		}
	}

	// brightRed := "\033[91m"
	// reset := "\033[0m"

	// Print invocation stack (how we got here via spawn/run calls)
	if event.InvocationStack != nil {
		fmt.Fprintf(os.Stderr, "\nInvocation stack:\n")
		frame := event.InvocationStack
		isTopFrame := true
		for frame != nil {
			fmt.Fprintf(os.Stderr, "  %s", frame.Reason)
			if frame.Filename != "" {
				// For the top frame (current invocation), show the current breakpoint position
				// For parent frames (outer invocations), show their invocation position
				if isTopFrame && event.Position.Line > 0 {
					// Top frame - show current execution position (breakpoint location)
					fmt.Fprintf(os.Stderr, " (%s:%d:%d)", frame.Filename, event.Position.Line, event.Position.Column)
				} else if frame.Line > 0 || frame.Col > 0 {
					// Parent frames - show their invocation position (only if non-zero)
					fmt.Fprintf(os.Stderr, " (%s:%d:%d)", frame.Filename, frame.Line, frame.Col)
				} else {
					// Fallback to filename only if no position available
					fmt.Fprintf(os.Stderr, " (%s)", frame.Filename)
				}
			}
			fmt.Fprintf(os.Stderr, "\n")
			frame = frame.Parent
			isTopFrame = false
		}
	}

	// Create a synthetic breakpoint error for the REPL
	bpErr := &script.BreakpointError{
		FilePath:  event.FilePath,
		Position:  event.Position,
		CallStack: event.CallStack,
		Env:       event.Env,
	}

	// Open debug REPL - user can inspect variables and type 'c' to continue
	openConsoleDebugREPL(interp, bpErr, event.Message)

	// Signal the child script to resume execution
	if event.ResumeChan != nil {
		select {
		case event.ResumeChan <- true:
		default:
			// Channel closed or already received, skip
		}
	}
}

// openConsoleDebugREPL opens an interactive debug session at a breakpoint
// It serializes access to stdin using the interpreter's debug session mutex
func openConsoleDebugREPL(interp *script.Interpreter, bpErr *script.BreakpointError, message string) error {
	// Acquire exclusive access to stdin - only one REPL at a time
	sessionMu := interp.GetDebugSessionMutex()
	sessionMu.Lock()
	defer sessionMu.Unlock()

	brightRed := "\033[91m"
	reset := "\033[0m"

	// Show message if present (from breakpoint() or watch())
	if message != "" {
		fmt.Fprintf(os.Stderr, "\n%s%s%s\n", brightRed, message, reset)
	}

	// Format breakpoint location
	loc := bpErr.FilePath
	if bpErr.Position.Line > 0 {
		loc = fmt.Sprintf("%s:%d", loc, bpErr.Position.Line)
		if bpErr.Position.Column > 0 {
			loc = fmt.Sprintf("%s:%d", loc, bpErr.Position.Column)
		}
	}

	// Show source code context around the breakpoint
	if bpErr.Position.Line > 0 {
		showConsoleSourceContext(bpErr.FilePath, bpErr.Position.Line, bpErr.Position.Column)
	}

	// Print call stack
	if len(bpErr.CallStack) > 0 {
		fmt.Fprintf(os.Stderr, "\nCall stack:\n")
		for i := len(bpErr.CallStack) - 1; i >= 0; i-- {
			frame := bpErr.CallStack[i]
			fmt.Fprintf(os.Stderr, "  at %s", frame.FunctionName)
			if frame.FilePath != "" {
				fmt.Fprintf(os.Stderr, " (%s:%d", frame.FilePath, frame.Position.Line)
				if frame.Position.Column > 0 {
					fmt.Fprintf(os.Stderr, ":%d", frame.Position.Column)
				}
				fmt.Fprintf(os.Stderr, ")")
			}
			fmt.Fprintf(os.Stderr, "\n")
		}
	}

	fmt.Fprintf(os.Stderr, "\nType 'c' to continue, or inspect variables.\n")

	// Run the debug REPL
	// Use the interpreter's InputReader so it works with HTTP stdin/stdout transport
	for {
		// Use InputReader (which may be HTTP-backed via -stdin-port)
		line, err := interp.InputReader("debug> ")
		if err != nil {
			// EOF or error - just return silently
			return nil
		}

		// 'c' command continues execution
		if line == "c" {
			return nil
		}

		// Evaluate code in the breakpoint's environment
		if line != "" {
			codeToExecute := wrapExpressionWithPrint(line)
			_, err := interp.EvalInEnvironment(codeToExecute, bpErr.Env)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}
		}
	}
}

// showConsoleSourceContext displays source code around a breakpoint location
func showConsoleSourceContext(filePath string, line int, col int) {
	// Read the file
	source, err := os.ReadFile(filePath)
	if err != nil {
		return // Silently fail if we can't read the file
	}

	lines := strings.Split(string(source), "\n")
	if line < 1 || line > len(lines) {
		return
	}

	// Show 2 lines before, the line itself, and 3 lines after
	start := line - 2
	if start < 1 {
		start = 1
	}
	end := line + 3
	if end > len(lines) {
		end = len(lines)
	}

	fmt.Fprintf(os.Stderr, "\n")

	// Calculate width needed for line numbers
	lineNumWidth := len(fmt.Sprintf("%d", end))

	yellow := "\033[93m"
	brightRed := "\033[91m"
	reset := "\033[0m"

	for i := start; i <= end; i++ {
		lineContent := ""
		if i <= len(lines) {
			lineContent = lines[i-1]
		}

		// Blank line before the breakpoint line
		if i == line {
			fmt.Fprintf(os.Stderr, "\n")
		}

		// Highlight the line with the breakpoint
		if i == line {
			fmt.Fprintf(os.Stderr, "%s%*d | %s%s\n", yellow, lineNumWidth, i, lineContent, reset)

			// Show column marker if column is specified
			if col > 0 {
				// Account for line number width + " | " separator
				marker := strings.Repeat(" ", lineNumWidth+3+col-1) + "^"
				fmt.Fprintf(os.Stderr, "%s%s%s\n", brightRed, marker, reset) // Bright red for caret
			}
		} else {
			fmt.Fprintf(os.Stderr, "%*d | %s\n", lineNumWidth, i, lineContent)
		}
	}

	fmt.Fprintf(os.Stderr, "\n")
}
