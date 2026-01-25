package script

// CallFrame represents a function call in the execution stack
type CallFrame struct {
	FunctionName string
	FilePath     string
	Position     Position
}

// ExecContext tracks execution state including file path and call stack
type ExecContext struct {
	FilePath  string
	CallStack []CallFrame
}

// NewExecContext creates a new execution context with the given file path
func NewExecContext(filePath string) *ExecContext {
	return &ExecContext{
		FilePath:  filePath,
		CallStack: make([]CallFrame, 0, 16), // Pre-allocate for typical recursion depth
	}
}

// PushCall adds a function call to the stack
func (ctx *ExecContext) PushCall(name, file string, pos Position) {
	ctx.CallStack = append(ctx.CallStack, CallFrame{
		FunctionName: name,
		FilePath:     file,
		Position:     pos,
	})
}

// PopCall removes the last function call from the stack
func (ctx *ExecContext) PopCall() {
	if len(ctx.CallStack) > 0 {
		ctx.CallStack = ctx.CallStack[:len(ctx.CallStack)-1]
	}
}

// Depth returns the current call stack depth
func (ctx *ExecContext) Depth() int {
	return len(ctx.CallStack)
}
