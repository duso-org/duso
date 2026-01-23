// ast.go - Duso Abstract Syntax Tree node definitions
//
// This file defines the data structures that represent a parsed Duso program.
// The parser converts tokens into an AST using these node types.
//
// CORE LANGUAGE COMPONENT: This is part of the minimal core language.
// The AST is the intermediate representation between source code and evaluation.
//
// Node types include:
// - Program: Root node containing all statements
// - Statements: if/elseif/else, while/for loops, function definitions, assignments
// - Expressions: Binary/unary operations, function calls, literals, variables
// - Values: Numbers, strings, booleans, arrays, objects, nil
// - Functions: Function definitions and calls (both user and built-in)
//
// The AST structure enables:
// - Proper error reporting (statements have source locations)
// - Correct evaluation order and precedence
// - Support for all language constructs
package script

// Node is the interface that all AST nodes must implement
type Node interface {
	node()
}

// Program is the root node of an AST
type Program struct {
	Statements []Node
}

func (p *Program) node() {}

// Statements

type IfStatement struct {
	Condition Node
	Then      []Node
	Elseifs   []*ElseifClause
	Else      []Node
}

type ElseifClause struct {
	Condition Node
	Then      []Node
}

func (s *IfStatement) node()      {}
func (s *ElseifClause) node()     {}

type WhileStatement struct {
	Condition Node
	Body      []Node
}

func (s *WhileStatement) node() {}

type ForStatement struct {
	Var       string
	Start     Node
	End       Node
	Step      Node // Can be nil for iterator-based for loops
	Iterator  Node // Non-nil for "for item in array" loops
	Body      []Node
	IsNumeric bool // true for numeric for, false for iterator-based
}

func (s *ForStatement) node() {}

// Parameter represents a function parameter with optional default value
type Parameter struct {
	Name    string // Parameter name
	Default Node   // Default value expression (nil if no default)
}

type FunctionDef struct {
	Name       string
	Parameters []*Parameter
	Body       []Node
}

func (s *FunctionDef) node() {}

type TryStatement struct {
	Block      []Node
	CatchVar   string
	CatchBlock []Node
}

func (s *TryStatement) node() {}

type ReturnStatement struct {
	Value Node // Can be nil
}

func (s *ReturnStatement) node() {}

type BreakStatement struct{}

func (s *BreakStatement) node() {}

type ContinueStatement struct{}

func (s *ContinueStatement) node() {}

type AssignStatement struct {
	Target           Node // Can be Identifier, IndexExpr, or PropertyAccess
	Value            Node
	IsVarDeclaration bool // true if "var x = ..." syntax
}

func (s *AssignStatement) node() {}

type CompoundAssignStatement struct {
	Target   Node       // Can be Identifier, IndexExpr, or PropertyAccess
	Operator TokenType  // TOK_PLUSASSIGN, TOK_MINUSASSIGN, etc.
	Value    Node
}

func (s *CompoundAssignStatement) node() {}

type PostIncrementStatement struct {
	Target   Node      // Can be Identifier, IndexExpr, or PropertyAccess
	Operator TokenType // TOK_INCREMENT or TOK_DECREMENT
}

func (s *PostIncrementStatement) node() {}

// Expressions

type BinaryExpr struct {
	Op    TokenType
	Left  Node
	Right Node
}

func (e *BinaryExpr) node() {}

type TernaryExpr struct {
	Condition Node
	TrueExpr  Node
	FalseExpr Node
}

func (e *TernaryExpr) node() {}

type UnaryExpr struct {
	Op    TokenType
	Operand Node
}

func (e *UnaryExpr) node() {}

type CallExpr struct {
	Func      Node
	Arguments []Node
	NamedArgs map[string]Node // For function(name = value) style calls
}

func (e *CallExpr) node() {}

type IndexExpr struct {
	Object Node
	Index  Node
}

func (e *IndexExpr) node() {}

type PropertyAccess struct {
	Object   Node
	Property string
}

func (e *PropertyAccess) node() {}

type Identifier struct {
	Name string
}

func (e *Identifier) node() {}

// Literals

type NumberLiteral struct {
	Value float64
}

func (l *NumberLiteral) node() {}

type StringLiteral struct {
	Value string
}

func (l *StringLiteral) node() {}

type BoolLiteral struct {
	Value bool
}

func (l *BoolLiteral) node() {}

type NilLiteral struct{}

func (l *NilLiteral) node() {}

type ArrayLiteral struct {
	Elements []Node
}

func (l *ArrayLiteral) node() {}

type ObjectLiteral struct {
	Pairs map[string]Node
}

func (l *ObjectLiteral) node() {}

type TemplateLiteral struct {
	Parts []Node // Alternating TextPart and expression nodes
}

type TextPart struct {
	Value string
}

type FunctionExpr struct {
	Parameters []*Parameter
	Body       []Node
}

func (l *TemplateLiteral) node() {}
func (t *TextPart) node()        {}
func (e *FunctionExpr) node()    {}
