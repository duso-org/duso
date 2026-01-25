package script

// evaluator.go - Core interpreter runtime
//
// This file contains the Evaluator, which is the heart of the Duso interpreter.
// The Evaluator executes the abstract syntax tree (AST) produced by the Parser.
//
// CORE RUNTIME: This is part of the minimal, self-contained core language.
// It uses only Go stdlib with NO external dependencies.
//
// The Evaluator handles:
// - Variable scoping and environment management
// - Function definition and calls
// - Control flow execution (if/while/for)
// - Type coercion and comparisons
// - Built-in function execution
// - Custom function registration from Go
//
// Optional features (like file I/O or Claude API) are NOT part of this core.
// They are registered separately via pkg/cli or by embedded applications.

import (
	"fmt"
	"strconv"
	"strings"
)

type Evaluator struct {
	env                *Environment
	builtins           *Builtins
	goFunctions        map[string]GoFunction
	goObjects          map[string]map[string]GoFunction
	isParallelContext  bool // True when executing in a parallel() block - parent scope writes are blocked
	ctx                *ExecContext // Execution context for error reporting and call stack tracking
	watchCache         map[string]Value // Cache for watch() expressions (expr -> last value)
}

// isInteger checks if a float64 is an integer value
func isInteger(n float64) bool {
	return n == float64(int64(n))
}

// Evaluator helper methods

// tryCoerceToNumber attempts to convert a value to a number
// Returns (number, success)
func (e *Evaluator) tryCoerceToNumber(v Value) (float64, bool) {
	if v.IsNumber() {
		return v.AsNumber(), true
	}
	if v.IsString() {
		s := v.AsString()
		// Try to parse the string as a number
		if num, err := parseFloat(s); err == nil {
			return num, true
		}
		return 0, false
	}
	return 0, false
}

// parseFloat parses a string as a float64
func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

// NewEvaluator creates a new evaluator
func NewEvaluator(output *strings.Builder) *Evaluator {
	env := NewEnvironment()

	evaluator := &Evaluator{
		env:         env,
		builtins:    nil, // Will be set below
		goFunctions: make(map[string]GoFunction),
		goObjects:   make(map[string]map[string]GoFunction),
		ctx:         NewExecContext("<stdin>"),
		watchCache:  make(map[string]Value),
	}

	evaluator.builtins = NewBuiltins(output, evaluator)
	evaluator.builtins.RegisterBuiltins(env)

	return evaluator
}

// RegisterFunction registers a Go function
func (e *Evaluator) RegisterFunction(name string, fn GoFunction) {
	e.goFunctions[name] = fn
	e.env.Define(name, NewGoFunction(fn))
}

// RegisterObject registers an object with methods
func (e *Evaluator) RegisterObject(name string, methods map[string]GoFunction) {
	e.goObjects[name] = methods
}

// newError creates a DusoError with current context (file, position, call stack)
func (e *Evaluator) newError(msg string, pos Position) error {
	// Clone the call stack to avoid mutations
	stack := make([]CallFrame, len(e.ctx.CallStack))
	copy(stack, e.ctx.CallStack)

	return &DusoError{
		Message:   msg,
		FilePath:  e.ctx.FilePath,
		Position:  pos,
		CallStack: stack,
	}
}

// wrapError converts a generic error to DusoError if it isn't already, adding position from node
func (e *Evaluator) wrapError(err error, node Node) error {
	if err == nil {
		return nil
	}

	// If already a DusoError, return as-is
	if _, ok := err.(*DusoError); ok {
		return err
	}

	// Try to extract position from node
	pos := NoPos
	switch n := node.(type) {
	case *BinaryExpr:
		pos = n.Pos
	case *UnaryExpr:
		pos = n.Pos
	case *CallExpr:
		pos = n.Pos
	case *IndexExpr:
		pos = n.Pos
	case *PropertyAccess:
		pos = n.Pos
	case *Identifier:
		pos = n.Pos
	case *TernaryExpr:
		pos = n.Pos
	case *TemplateLiteral:
		pos = n.Pos
	case *IfStatement:
		pos = n.Pos
	case *WhileStatement:
		pos = n.Pos
	case *ForStatement:
		pos = n.Pos
	case *AssignStatement:
		pos = n.Pos
	}

	// Clone call stack
	stack := make([]CallFrame, len(e.ctx.CallStack))
	copy(stack, e.ctx.CallStack)

	return &DusoError{
		Message:   err.Error(),
		FilePath:  e.ctx.FilePath,
		Position:  pos,
		CallStack: stack,
	}
}

// Eval evaluates a node
func (e *Evaluator) Eval(node Node) (Value, error) {
	switch n := node.(type) {
	case *Program:
		return e.evalProgram(n)
	case *IfStatement:
		return e.evalIfStatement(n)
	case *WhileStatement:
		return e.evalWhileStatement(n)
	case *ForStatement:
		return e.evalForStatement(n)
	case *FunctionDef:
		return e.evalFunctionDef(n)
	case *TryStatement:
		return e.evalTryStatement(n)
	case *ReturnStatement:
		return e.evalReturnStatement(n)
	case *BreakStatement:
		return NewNil(), &BreakIteration{}
	case *ContinueStatement:
		return NewNil(), &ContinueIteration{}
	case *AssignStatement:
		return e.evalAssignStatement(n)
	case *CompoundAssignStatement:
		return e.evalCompoundAssignStatement(n)
	case *PostIncrementStatement:
		return e.evalPostIncrementStatement(n)
	case *BinaryExpr:
		return e.evalBinaryExpr(n)
	case *TernaryExpr:
		return e.evalTernaryExpr(n)
	case *UnaryExpr:
		return e.evalUnaryExpr(n)
	case *CallExpr:
		return e.evalCallExpr(n)
	case *IndexExpr:
		return e.evalIndexExpr(n)
	case *PropertyAccess:
		return e.evalPropertyAccess(n)
	case *Identifier:
		val, err := e.env.Get(n.Name)
		return val, e.wrapError(err, n)
	case *NumberLiteral:
		return NewNumber(n.Value), nil
	case *StringLiteral:
		return NewString(n.Value), nil
	case *TemplateLiteral:
		return e.evalTemplateLiteral(n)
	case *BoolLiteral:
		return NewBool(n.Value), nil
	case *NilLiteral:
		return NewNil(), nil
	case *ArrayLiteral:
		return e.evalArrayLiteral(n)
	case *ObjectLiteral:
		return e.evalObjectLiteral(n)
	case *FunctionExpr:
		return e.evalFunctionExpr(n)
	default:
		return NewNil(), fmt.Errorf("unknown node type: %T", n)
	}
}

// EvalModule evaluates a program in an isolated module scope and returns the result.
// This is used by require() to load modules in isolation - the module's variables
// don't leak into the caller's scope. The last expression value becomes the module's export.
func (e *Evaluator) EvalModule(prog *Program) (Value, error) {
	// Create isolated module environment (like a function scope)
	moduleEnv := NewFunctionEnvironment(e.env)

	prevEnv := e.env
	e.env = moduleEnv

	var result Value
	for _, stmt := range prog.Statements {
		val, err := e.Eval(stmt)
		if err != nil {
			e.env = prevEnv
			// Allow explicit return statements in modules
			if retVal, ok := err.(*ReturnValue); ok {
				return retVal.Value, nil
			}
			return NewNil(), err
		}
		result = val
	}

	e.env = prevEnv
	return result, nil
}

func (e *Evaluator) evalProgram(prog *Program) (Value, error) {
	var result Value
	for _, stmt := range prog.Statements {
		val, err := e.Eval(stmt)
		if err != nil {
			if _, ok := err.(*ReturnValue); ok {
				return NewNil(), fmt.Errorf("return outside of function")
			}
			return NewNil(), err
		}
		result = val
	}
	return result, nil
}

func (e *Evaluator) evalIfStatement(stmt *IfStatement) (Value, error) {
	condition, err := e.Eval(stmt.Condition)
	if err != nil {
		return NewNil(), err
	}

	if condition.IsTruthy() {
		return e.evalBlock(stmt.Then, e.env)
	}

	// Check elseif clauses
	for _, elif := range stmt.Elseifs {
		condition, err := e.Eval(elif.Condition)
		if err != nil {
			return NewNil(), err
		}
		if condition.IsTruthy() {
			return e.evalBlock(elif.Then, e.env)
		}
	}

	// Else clause
	if stmt.Else != nil {
		return e.evalBlock(stmt.Else, e.env)
	}

	return NewNil(), nil
}

func (e *Evaluator) evalWhileStatement(stmt *WhileStatement) (Value, error) {
	var result Value
	for {
		condition, err := e.Eval(stmt.Condition)
		if err != nil {
			return NewNil(), err
		}

		if !condition.IsTruthy() {
			break
		}

		val, err := e.evalBlock(stmt.Body, e.env)
		if err != nil {
			// Handle break/continue
			if _, ok := err.(*BreakIteration); ok {
				break
			}
			if _, ok := err.(*ContinueIteration); ok {
				continue
			}
			return NewNil(), err
		}
		result = val
	}
	return result, nil
}

func (e *Evaluator) evalForStatement(stmt *ForStatement) (Value, error) {
	var result Value

	if stmt.IsNumeric {
		// Numeric for loop: for i = start, end [, step] do
		startVal, err := e.Eval(stmt.Start)
		if err != nil {
			return NewNil(), err
		}
		if !startVal.IsNumber() {
			return NewNil(), fmt.Errorf("for loop start must be a number")
		}
		startNum := startVal.AsNumber()
		if !isInteger(startNum) {
			return NewNil(), fmt.Errorf("for loop start must be an integer, got %v", startNum)
		}
		start := int64(startNum)

		endVal, err := e.Eval(stmt.End)
		if err != nil {
			return NewNil(), err
		}
		if !endVal.IsNumber() {
			return NewNil(), fmt.Errorf("for loop end must be a number")
		}
		endNum := endVal.AsNumber()
		if !isInteger(endNum) {
			return NewNil(), fmt.Errorf("for loop end must be an integer, got %v", endNum)
		}
		end := int64(endNum)

		step := int64(1)
		if stmt.Step != nil {
			stepVal, err := e.Eval(stmt.Step)
			if err != nil {
				return NewNil(), err
			}
			if !stepVal.IsNumber() {
				return NewNil(), fmt.Errorf("for loop step must be a number")
			}
			stepNum := stepVal.AsNumber()
			if !isInteger(stepNum) {
				return NewNil(), fmt.Errorf("for loop step must be an integer, got %v", stepNum)
			}
			step = int64(stepNum)
		}

		if step == 0 {
			return NewNil(), fmt.Errorf("for loop step cannot be zero")
		}

		// Create child scope
		loopEnv := NewChildEnvironment(e.env)

		for i := start; (step > 0 && i <= end) || (step < 0 && i >= end); i += step {
			loopEnv.Define(stmt.Var, NewNumber(float64(i)))

			// Temporarily switch environment
			prevEnv := e.env
			e.env = loopEnv

			val, err := e.evalBlock(stmt.Body, loopEnv)
			e.env = prevEnv

			if err != nil {
				// Handle break/continue
				if _, ok := err.(*BreakIteration); ok {
					break
				}
				if _, ok := err.(*ContinueIteration); ok {
					continue
				}
				return NewNil(), err
			}
			result = val
		}
	} else {
		// Iterator for loop: for item in array/object do
		iterVal, err := e.Eval(stmt.Iterator)
		if err != nil {
			return NewNil(), err
		}

		if !iterVal.IsArray() && !iterVal.IsObject() {
			return NewNil(), fmt.Errorf("can only iterate over arrays and objects")
		}

		loopEnv := NewChildEnvironment(e.env)

		if iterVal.IsArray() {
			arr := iterVal.AsArray()
			for _, item := range arr {
				loopEnv.Define(stmt.Var, item)

				prevEnv := e.env
				e.env = loopEnv

				val, err := e.evalBlock(stmt.Body, loopEnv)
				e.env = prevEnv

				if err != nil {
					// Handle break/continue
					if _, ok := err.(*BreakIteration); ok {
						break
					}
					if _, ok := err.(*ContinueIteration); ok {
						continue
					}
					return NewNil(), err
				}
				result = val
			}
		} else if iterVal.IsObject() {
			objMap := iterVal.AsObject()
			for key := range objMap {
				loopEnv.Define(stmt.Var, NewString(key))

				prevEnv := e.env
				e.env = loopEnv

				val, err := e.evalBlock(stmt.Body, loopEnv)
				e.env = prevEnv

				if err != nil {
					// Handle break/continue
					if _, ok := err.(*BreakIteration); ok {
						break
					}
					if _, ok := err.(*ContinueIteration); ok {
						continue
					}
					return NewNil(), err
				}
				result = val
			}
		}
	}

	return result, nil
}

func (e *Evaluator) evalFunctionDef(stmt *FunctionDef) (Value, error) {
	fn := &ScriptFunction{
		Name:       stmt.Name,
		FilePath:   e.ctx.FilePath,
		Parameters: stmt.Parameters,
		Body:       stmt.Body,
		Closure:    e.env,
	}
	e.env.Define(stmt.Name, NewFunction(fn))
	return NewNil(), nil
}

func (e *Evaluator) evalTryStatement(stmt *TryStatement) (Value, error) {
	val, err := e.evalBlock(stmt.Block, e.env)

	if err != nil {
		// Execute catch block
		catchEnv := NewChildEnvironment(e.env)
		catchEnv.Define(stmt.CatchVar, NewString(err.Error()))

		prevEnv := e.env
		e.env = catchEnv

		val, catchErr := e.evalBlock(stmt.CatchBlock, catchEnv)
		e.env = prevEnv

		if catchErr != nil {
			return NewNil(), catchErr
		}
		return val, nil
	}

	return val, nil
}

func (e *Evaluator) evalReturnStatement(stmt *ReturnStatement) (Value, error) {
	if stmt.Value == nil {
		return NewNil(), &ReturnValue{Value: NewNil()}
	}

	val, err := e.Eval(stmt.Value)
	if err != nil {
		return NewNil(), err
	}

	return NewNil(), &ReturnValue{Value: val}
}

func (e *Evaluator) evalAssignStatement(stmt *AssignStatement) (Value, error) {
	value, err := e.Eval(stmt.Value)
	if err != nil {
		return NewNil(), err
	}

	switch target := stmt.Target.(type) {
	case *Identifier:
		if stmt.IsVarDeclaration {
			// var declaration: check if trying to shadow a function parameter
			if e.env.IsParameter(target.Name) {
				return NewNil(), fmt.Errorf("cannot use 'var' to declare function parameter '%s'; use '%s = value' instead", target.Name, target.Name)
			}
			// var declaration: always create local variable
			e.env.Define(target.Name, value)
		} else {
			// regular assignment: reach through scope chain or create locally at boundary
			e.env.Set(target.Name, value)
		}
		return value, nil
	case *IndexExpr:
		return e.evalIndexAssign(target, value)
	case *PropertyAccess:
		return e.evalPropertyAssign(target, value)
	default:
		return NewNil(), fmt.Errorf("invalid assignment target")
	}
}

func (e *Evaluator) evalCompoundAssignStatement(stmt *CompoundAssignStatement) (Value, error) {
	// Get the current value
	var currentVal Value
	var err error

	switch target := stmt.Target.(type) {
	case *Identifier:
		currentVal, err = e.env.Get(target.Name)
		if err != nil {
			return NewNil(), e.wrapError(err, target)
		}
	case *IndexExpr:
		currentVal, err = e.Eval(stmt.Target)
		if err != nil {
			return NewNil(), err
		}
	case *PropertyAccess:
		currentVal, err = e.Eval(stmt.Target)
		if err != nil {
			return NewNil(), err
		}
	default:
		return NewNil(), fmt.Errorf("invalid assignment target")
	}

	// Get the value to apply
	rightVal, err := e.Eval(stmt.Value)
	if err != nil {
		return NewNil(), err
	}

	// Apply the operator
	var result Value
	switch stmt.Operator {
	case TOK_PLUSASSIGN:
		if currentVal.IsNumber() && rightVal.IsNumber() {
			result = NewNumber(currentVal.AsNumber() + rightVal.AsNumber())
		} else if currentVal.IsString() || rightVal.IsString() {
			result = NewString(currentVal.String() + rightVal.String())
		} else {
			return NewNil(), fmt.Errorf("invalid operands for +=")
		}
	case TOK_MINUSASSIGN:
		if !currentVal.IsNumber() || !rightVal.IsNumber() {
			return NewNil(), fmt.Errorf("invalid operands for -=")
		}
		result = NewNumber(currentVal.AsNumber() - rightVal.AsNumber())
	case TOK_STARASSIGN:
		if !currentVal.IsNumber() || !rightVal.IsNumber() {
			return NewNil(), fmt.Errorf("invalid operands for *=")
		}
		result = NewNumber(currentVal.AsNumber() * rightVal.AsNumber())
	case TOK_SLASHASSIGN:
		if !currentVal.IsNumber() || !rightVal.IsNumber() {
			return NewNil(), fmt.Errorf("invalid operands for /=")
		}
		if rightVal.AsNumber() == 0 {
			return NewNil(), fmt.Errorf("division by zero")
		}
		result = NewNumber(currentVal.AsNumber() / rightVal.AsNumber())
	case TOK_MODASSIGN:
		if !currentVal.IsNumber() || !rightVal.IsNumber() {
			return NewNil(), fmt.Errorf("invalid operands for %%=")
		}
		if rightVal.AsNumber() == 0 {
			return NewNil(), fmt.Errorf("modulo by zero")
		}
		result = NewNumber(float64(int64(currentVal.AsNumber()) % int64(rightVal.AsNumber())))
	default:
		return NewNil(), fmt.Errorf("unknown compound assignment operator")
	}

	// Assign the result
	switch target := stmt.Target.(type) {
	case *Identifier:
		e.env.Set(target.Name, result)
		return result, nil
	case *IndexExpr:
		return e.evalIndexAssign(target, result)
	case *PropertyAccess:
		return e.evalPropertyAssign(target, result)
	default:
		return NewNil(), fmt.Errorf("invalid assignment target")
	}
}

func (e *Evaluator) evalPostIncrementStatement(stmt *PostIncrementStatement) (Value, error) {
	// Get the current value
	var currentVal Value
	var err error

	switch target := stmt.Target.(type) {
	case *Identifier:
		currentVal, err = e.env.Get(target.Name)
		if err != nil {
			return NewNil(), e.wrapError(err, target)
		}
	case *IndexExpr:
		currentVal, err = e.Eval(stmt.Target)
		if err != nil {
			return NewNil(), err
		}
	case *PropertyAccess:
		currentVal, err = e.Eval(stmt.Target)
		if err != nil {
			return NewNil(), err
		}
	default:
		return NewNil(), fmt.Errorf("invalid increment/decrement target")
	}

	if !currentVal.IsNumber() {
		return NewNil(), fmt.Errorf("cannot increment/decrement non-number")
	}

	// Calculate new value
	var newVal Value
	if stmt.Operator == TOK_INCREMENT {
		newVal = NewNumber(currentVal.AsNumber() + 1)
	} else {
		newVal = NewNumber(currentVal.AsNumber() - 1)
	}

	// Assign the new value
	switch target := stmt.Target.(type) {
	case *Identifier:
		e.env.Set(target.Name, newVal)
	case *IndexExpr:
		_, err := e.evalIndexAssign(target, newVal)
		if err != nil {
			return NewNil(), err
		}
	case *PropertyAccess:
		_, err := e.evalPropertyAssign(target, newVal)
		if err != nil {
			return NewNil(), err
		}
	default:
		return NewNil(), fmt.Errorf("invalid increment/decrement target")
	}

	// Return nil for statement form (as per plan, increment/decrement are statements, not expressions)
	return NewNil(), nil
}

func (e *Evaluator) evalBinaryExpr(expr *BinaryExpr) (Value, error) {
	left, err := e.Eval(expr.Left)
	if err != nil {
		return NewNil(), err
	}

	// Short-circuit evaluation for logical operators
	if expr.Op == TOK_AND {
		if !left.IsTruthy() {
			return NewBool(false), nil
		}
		right, err := e.Eval(expr.Right)
		if err != nil {
			return NewNil(), err
		}
		return NewBool(right.IsTruthy()), nil
	}

	if expr.Op == TOK_OR {
		if left.IsTruthy() {
			return NewBool(true), nil
		}
		right, err := e.Eval(expr.Right)
		if err != nil {
			return NewNil(), err
		}
		return NewBool(right.IsTruthy()), nil
	}

	right, err := e.Eval(expr.Right)
	if err != nil {
		return NewNil(), err
	}

	switch expr.Op {
	case TOK_PLUS:
		if left.IsNumber() && right.IsNumber() {
			return NewNumber(left.AsNumber() + right.AsNumber()), nil
		}
		if left.IsString() || right.IsString() {
			return NewString(left.String() + right.String()), nil
		}
		return NewNil(), fmt.Errorf("cannot add %v and %v", left.Type, right.Type)

	case TOK_MINUS:
		if left.IsNumber() && right.IsNumber() {
			return NewNumber(left.AsNumber() - right.AsNumber()), nil
		}
		return NewNil(), fmt.Errorf("cannot subtract non-numbers")

	case TOK_STAR:
		if left.IsNumber() && right.IsNumber() {
			return NewNumber(left.AsNumber() * right.AsNumber()), nil
		}
		return NewNil(), fmt.Errorf("cannot multiply non-numbers")

	case TOK_SLASH:
		if left.IsNumber() && right.IsNumber() {
			if right.AsNumber() == 0 {
				return NewNil(), e.newError("division by zero", expr.Pos)
			}
			return NewNumber(left.AsNumber() / right.AsNumber()), nil
		}
		return NewNil(), e.newError("cannot divide non-numbers", expr.Pos)

	case TOK_PERCENT:
		if left.IsNumber() && right.IsNumber() {
			if right.AsNumber() == 0 {
				return NewNil(), fmt.Errorf("modulo by zero")
			}
			// Go's % operator for floats
			return NewNumber(float64(int64(left.AsNumber()) % int64(right.AsNumber()))), nil
		}
		return NewNil(), fmt.Errorf("cannot modulo non-numbers")

	case TOK_EQUAL:
		return NewBool(e.valuesEqual(left, right)), nil

	case TOK_NOTEQUAL:
		return NewBool(!e.valuesEqual(left, right)), nil

	case TOK_LT:
		// Try to coerce strings to numbers for comparison
		leftNum, leftIsNum := e.tryCoerceToNumber(left)
		rightNum, rightIsNum := e.tryCoerceToNumber(right)

		if leftIsNum && rightIsNum {
			return NewBool(leftNum < rightNum), nil
		}
		if left.IsString() && right.IsString() {
			return NewBool(left.AsString() < right.AsString()), nil
		}
		return NewNil(), fmt.Errorf("cannot compare %v and %v", left.Type, right.Type)

	case TOK_GT:
		// Try to coerce strings to numbers for comparison
		leftNum, leftIsNum := e.tryCoerceToNumber(left)
		rightNum, rightIsNum := e.tryCoerceToNumber(right)

		if leftIsNum && rightIsNum {
			return NewBool(leftNum > rightNum), nil
		}
		if left.IsString() && right.IsString() {
			return NewBool(left.AsString() > right.AsString()), nil
		}
		return NewNil(), fmt.Errorf("cannot compare %v and %v", left.Type, right.Type)

	case TOK_LTE:
		// Try to coerce strings to numbers for comparison
		leftNum, leftIsNum := e.tryCoerceToNumber(left)
		rightNum, rightIsNum := e.tryCoerceToNumber(right)

		if leftIsNum && rightIsNum {
			return NewBool(leftNum <= rightNum), nil
		}
		if left.IsString() && right.IsString() {
			return NewBool(left.AsString() <= right.AsString()), nil
		}
		return NewNil(), fmt.Errorf("cannot compare %v and %v", left.Type, right.Type)

	case TOK_GTE:
		// Try to coerce strings to numbers for comparison
		leftNum, leftIsNum := e.tryCoerceToNumber(left)
		rightNum, rightIsNum := e.tryCoerceToNumber(right)

		if leftIsNum && rightIsNum {
			return NewBool(leftNum >= rightNum), nil
		}
		if left.IsString() && right.IsString() {
			return NewBool(left.AsString() >= right.AsString()), nil
		}
		return NewNil(), fmt.Errorf("cannot compare %v and %v", left.Type, right.Type)

	default:
		return NewNil(), fmt.Errorf("unknown binary operator: %v", expr.Op)
	}
}

func (e *Evaluator) evalTernaryExpr(expr *TernaryExpr) (Value, error) {
	condition, err := e.Eval(expr.Condition)
	if err != nil {
		return NewNil(), err
	}

	if condition.IsTruthy() {
		return e.Eval(expr.TrueExpr)
	}
	return e.Eval(expr.FalseExpr)
}

func (e *Evaluator) evalUnaryExpr(expr *UnaryExpr) (Value, error) {
	// Handle pre-increment/decrement specially (they modify and return the new value)
	if expr.Op == TOK_INCREMENT || expr.Op == TOK_DECREMENT {
		var currentVal Value
		var err error

		switch target := expr.Operand.(type) {
		case *Identifier:
			currentVal, err = e.env.Get(target.Name)
			if err != nil {
				return NewNil(), e.wrapError(err, target)
			}
		case *IndexExpr:
			currentVal, err = e.Eval(expr.Operand)
			if err != nil {
				return NewNil(), err
			}
		case *PropertyAccess:
			currentVal, err = e.Eval(expr.Operand)
			if err != nil {
				return NewNil(), err
			}
		default:
			return NewNil(), fmt.Errorf("invalid increment/decrement target")
		}

		if !currentVal.IsNumber() {
			return NewNil(), fmt.Errorf("cannot increment/decrement non-number")
		}

		// Calculate new value
		var newVal Value
		if expr.Op == TOK_INCREMENT {
			newVal = NewNumber(currentVal.AsNumber() + 1)
		} else {
			newVal = NewNumber(currentVal.AsNumber() - 1)
		}

		// Assign the new value
		switch target := expr.Operand.(type) {
		case *Identifier:
			e.env.Set(target.Name, newVal)
		case *IndexExpr:
			_, err := e.evalIndexAssign(target, newVal)
			if err != nil {
				return NewNil(), err
			}
		case *PropertyAccess:
			_, err := e.evalPropertyAssign(target, newVal)
			if err != nil {
				return NewNil(), err
			}
		default:
			return NewNil(), fmt.Errorf("invalid increment/decrement target")
		}

		// Return the new value (pre-increment returns new value)
		return newVal, nil
	}

	operand, err := e.Eval(expr.Operand)
	if err != nil {
		return NewNil(), err
	}

	switch expr.Op {
	case TOK_NOT:
		return NewBool(!operand.IsTruthy()), nil
	case TOK_MINUS:
		if operand.IsNumber() {
			return NewNumber(-operand.AsNumber()), nil
		}
		return NewNil(), fmt.Errorf("cannot negate non-number")
	default:
		return NewNil(), fmt.Errorf("unknown unary operator: %v", expr.Op)
	}
}

func (e *Evaluator) evalCallExpr(expr *CallExpr) (Value, error) {
	var receiver Value
	var isMethodCall bool

	// Check if this is a method call (property access)
	if _, ok := expr.Func.(*PropertyAccess); ok {
		isMethodCall = true
		// Evaluate the object being accessed to use as receiver
		propAccess := expr.Func.(*PropertyAccess)
		var err error
		receiver, err = e.Eval(propAccess.Object)
		if err != nil {
			return NewNil(), err
		}
	}

	fn, err := e.Eval(expr.Func)
	if err != nil {
		return NewNil(), err
	}

	// Handle callable objects
	if fn.IsObject() {
		return e.callObject(fn, expr.Arguments, expr.NamedArgs)
	}

	if !fn.IsFunction() {
		return NewNil(), fmt.Errorf("cannot call non-function")
	}

	// Handle script functions
	if scriptFn, ok := fn.Data.(*ScriptFunction); ok {
		return e.callScriptFunction(scriptFn, expr.Arguments, expr.NamedArgs, receiver, isMethodCall, expr.Pos)
	}

	// Handle Go functions
	if goFn, ok := fn.Data.(GoFunction); ok {
		return e.callGoFunction(goFn, expr.Arguments, expr.NamedArgs, expr.Pos)
	}

	return NewNil(), fmt.Errorf("invalid function type")
}

func (e *Evaluator) callScriptFunction(fn *ScriptFunction, args []Node, namedArgs map[string]Node, receiver Value, isMethodCall bool, callPos Position) (Value, error) {
	// Push call frame for stack trace using the function's defined file path
	e.ctx.PushCall(fn.Name, fn.FilePath, callPos)
	defer e.ctx.PopCall()

	// Switch to function's file path for error reporting
	prevFilePath := e.ctx.FilePath
	e.ctx.FilePath = fn.FilePath
	defer func() { e.ctx.FilePath = prevFilePath }()

	// Create function environment (blocks variable walk-up) with receiver if this is a method call
	var fnEnv *Environment
	if isMethodCall {
		fnEnv = NewFunctionEnvironmentWithSelf(fn.Closure, receiver)
	} else {
		fnEnv = NewFunctionEnvironment(fn.Closure)
	}

	// Define all parameters with their defaults, and mark them as parameters
	for _, param := range fn.Parameters {
		var defaultVal Value = NewNil()
		if param.Default != nil {
			// Evaluate default in the closure environment (not the function env)
			prevEnv := e.env
			e.env = fn.Closure
			val, err := e.Eval(param.Default)
			e.env = prevEnv
			if err != nil {
				return NewNil(), err
			}
			defaultVal = val
		}
		fnEnv.Define(param.Name, defaultVal)
		fnEnv.MarkParameter(param.Name)
	}

	// Evaluate positional arguments
	for i, argNode := range args {
		if i < len(fn.Parameters) {
			val, err := e.Eval(argNode)
			if err != nil {
				return NewNil(), err
			}
			fnEnv.Define(fn.Parameters[i].Name, val)
		}
	}

	// Evaluate named arguments
	for name, argNode := range namedArgs {
		val, err := e.Eval(argNode)
		if err != nil {
			return NewNil(), err
		}
		fnEnv.Define(name, val)
	}

	// Execute function body
	prevEnv := e.env
	e.env = fnEnv

	var result Value
	for _, stmt := range fn.Body {
		val, err := e.Eval(stmt)
		if returnVal, ok := err.(*ReturnValue); ok {
			result = returnVal.Value
			break
		}
		if err != nil {
			e.env = prevEnv
			return NewNil(), err
		}
		result = val
	}

	e.env = prevEnv
	return result, nil
}

func (e *Evaluator) callGoFunction(goFn GoFunction, args []Node, namedArgs map[string]Node, callPos Position) (Value, error) {
	// Build argument map
	argMap := make(map[string]any)

	// Add positional arguments with numeric keys
	for i, argNode := range args {
		val, err := e.Eval(argNode)
		if err != nil {
			return NewNil(), err
		}
		argMap[fmt.Sprintf("%d", i)] = valueToInterface(val)
	}

	// Add named arguments
	for name, argNode := range namedArgs {
		val, err := e.Eval(argNode)
		if err != nil {
			return NewNil(), err
		}
		argMap[name] = valueToInterface(val)
	}

	// Call the function
	result, err := goFn(argMap)
	if err != nil {
		// Add position info to DusoError if not already present
		if dusoErr, ok := err.(*DusoError); ok && dusoErr.Position == (Position{}) {
			dusoErr.Position = callPos
		}
		// Add position info to BreakpointError if not already present
		if bpErr, ok := err.(*BreakpointError); ok && bpErr.Position == (Position{}) {
			bpErr.Position = callPos
		}
		return NewNil(), err
	}

	// Convert result back to script value
	return interfaceToValue(result), nil
}

func (e *Evaluator) callObject(obj Value, args []Node, namedArgs map[string]Node) (Value, error) {
	// Objects are callable as constructors - they create a new object (copy) with optional overrides
	// Objects can only be called with named arguments (named argument syntax)

	if len(args) > 0 {
		return NewNil(), fmt.Errorf("objects can only be called with named arguments")
	}

	// Get the object map
	objMap := obj.AsObject()

	// Create a copy of the object
	newObj := make(map[string]Value)
	for k, v := range objMap {
		newObj[k] = v
	}

	// Create a temporary environment where named arguments can reference each other
	// Start with current environment but add a child so we can track new definitions
	tmpEnv := NewChildEnvironment(e.env)
	prevEnv := e.env
	e.env = tmpEnv

	// Apply named argument overrides
	for name, argNode := range namedArgs {
		val, err := e.Eval(argNode)
		if err != nil {
			e.env = prevEnv
			return NewNil(), err
		}
		newObj[name] = val
		// Add to environment so subsequent arguments can reference it
		tmpEnv.Define(name, val)
	}

	e.env = prevEnv
	return NewObject(newObj), nil
}

func (e *Evaluator) evalIndexExpr(expr *IndexExpr) (Value, error) {
	obj, err := e.Eval(expr.Object)
	if err != nil {
		return NewNil(), err
	}

	index, err := e.Eval(expr.Index)
	if err != nil {
		return NewNil(), err
	}

	if obj.IsArray() {
		if !index.IsNumber() {
			return NewNil(), fmt.Errorf("array index must be a number")
		}
		idx := int(index.AsNumber())
		arr := obj.AsArray()
		if idx < 0 || idx >= len(arr) {
			return NewNil(), fmt.Errorf("array index out of bounds")
		}
		return arr[idx], nil
	}

	if obj.IsObject() {
		key := index.String()
		objMap := obj.AsObject()
		if val, ok := objMap[key]; ok {
			return val, nil
		}
		return NewNil(), nil
	}

	if obj.IsString() {
		if !index.IsNumber() {
			return NewNil(), fmt.Errorf("string index must be a number")
		}
		idx := int(index.AsNumber())
		str := obj.AsString()
		if idx < 0 || idx >= len(str) {
			return NewNil(), fmt.Errorf("string index out of bounds")
		}
		return NewString(string(str[idx])), nil
	}

	return NewNil(), fmt.Errorf("cannot index %v", obj.Type)
}

func (e *Evaluator) evalIndexAssign(expr *IndexExpr, value Value) (Value, error) {
	obj, err := e.Eval(expr.Object)
	if err != nil {
		return NewNil(), err
	}

	index, err := e.Eval(expr.Index)
	if err != nil {
		return NewNil(), err
	}

	if !obj.IsArray() && !obj.IsObject() {
		return NewNil(), fmt.Errorf("cannot assign to index of %v", obj.Type)
	}

	if obj.IsArray() {
		if !index.IsNumber() {
			return NewNil(), fmt.Errorf("array index must be a number")
		}
		idx := int(index.AsNumber())
		arr := obj.AsArray()
		if idx < 0 || idx >= len(arr) {
			return NewNil(), fmt.Errorf("array index out of bounds")
		}
		arr[idx] = value
		return value, nil
	}

	if obj.IsObject() {
		key := index.String()
		objMap := obj.AsObject()
		objMap[key] = value
		return value, nil
	}

	return NewNil(), fmt.Errorf("cannot assign to index of %v", obj.Type)
}

func (e *Evaluator) evalPropertyAccess(expr *PropertyAccess) (Value, error) {
	obj, err := e.Eval(expr.Object)
	if err != nil {
		return NewNil(), err
	}

	if obj.IsObject() {
		objMap := obj.AsObject()
		if val, ok := objMap[expr.Property]; ok {
			return val, nil
		}
		return NewNil(), nil
	}

	return NewNil(), fmt.Errorf("cannot access property of %v", obj.Type)
}

func (e *Evaluator) evalPropertyAssign(expr *PropertyAccess, value Value) (Value, error) {
	obj, err := e.Eval(expr.Object)
	if err != nil {
		return NewNil(), err
	}

	if !obj.IsObject() {
		return NewNil(), fmt.Errorf("cannot assign property of %v", obj.Type)
	}

	objMap := obj.AsObject()
	objMap[expr.Property] = value
	return value, nil
}

func (e *Evaluator) evalArrayLiteral(lit *ArrayLiteral) (Value, error) {
	var elements []Value
	for _, elemNode := range lit.Elements {
		elem, err := e.Eval(elemNode)
		if err != nil {
			return NewNil(), err
		}
		elements = append(elements, elem)
	}
	return NewArray(elements), nil
}

func (e *Evaluator) evalObjectLiteral(lit *ObjectLiteral) (Value, error) {
	obj := make(map[string]Value)
	for key, valueNode := range lit.Pairs {
		val, err := e.Eval(valueNode)
		if err != nil {
			return NewNil(), err
		}
		obj[key] = val
	}
	return NewObject(obj), nil
}

func (e *Evaluator) evalTemplateLiteral(lit *TemplateLiteral) (Value, error) {
	result := ""
	for _, part := range lit.Parts {
		if tp, ok := part.(*TextPart); ok {
			result += tp.Value
		} else {
			// Evaluate the expression
			val, err := e.Eval(part)
			if err != nil {
				// For template literal errors, use the template's position instead of inner expression position
				if dusoErr, ok := err.(*DusoError); ok {
					// Replace position with template literal position for more accurate source reporting
					dusoErr.Position = lit.Pos
					return NewNil(), dusoErr
				}
				// For non-Duso errors, wrap with template position
				return NewNil(), e.wrapError(err, lit)
			}
			// Convert to string using our String() method
			result += val.String()
		}
	}
	return NewString(result), nil
}

func (e *Evaluator) evalBlock(stmts []Node, env *Environment) (Value, error) {
	prevEnv := e.env
	e.env = env
	defer func() { e.env = prevEnv }()

	var result Value
	for _, stmt := range stmts {
		val, err := e.Eval(stmt)
		if err != nil {
			return NewNil(), err
		}
		result = val
	}
	return result, nil
}

func (e *Evaluator) evalFunctionExpr(expr *FunctionExpr) (Value, error) {
	fn := &ScriptFunction{
		Parameters: expr.Parameters,
		Body:       expr.Body,
		Closure:    e.env,
	}
	return NewFunction(fn), nil
}

func (e *Evaluator) valuesEqual(a, b Value) bool {
	if a.Type != b.Type {
		return false
	}

	switch a.Type {
	case VAL_NIL:
		return true
	case VAL_NUMBER:
		return a.AsNumber() == b.AsNumber()
	case VAL_STRING:
		return a.AsString() == b.AsString()
	case VAL_BOOL:
		return a.AsBool() == b.AsBool()
	default:
		return false // Arrays, objects, functions are not equal by value
	}
}
