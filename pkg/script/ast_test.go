package script

import (
	"testing"
)

// TestASTNodeInterfaces tests that all AST node types implement the Node interface
func TestASTNodeInterfaces(t *testing.T) {
	tests := []struct {
		name string
		node Node
	}{
		{"Program", &Program{Statements: []Node{}}},
		{"IfStatement", &IfStatement{Condition: &NumberLiteral{}}},
		{"ElseifClause", &ElseifClause{Condition: &NumberLiteral{}}},
		{"WhileStatement", &WhileStatement{Condition: &NumberLiteral{}}},
		{"ForStatement", &ForStatement{Var: "i", Start: &NumberLiteral{}, End: &NumberLiteral{}}},
		{"FunctionDef", &FunctionDef{Name: "test", Parameters: []*Parameter{}}},
		{"TryStatement", &TryStatement{Block: []Node{}}},
		{"ReturnStatement", &ReturnStatement{Value: &NumberLiteral{}}},
		{"BreakStatement", &BreakStatement{}},
		{"ContinueStatement", &ContinueStatement{}},
		{"AssignStatement", &AssignStatement{Target: &Identifier{Name: "x"}, Value: &NumberLiteral{}}},
		{"CompoundAssignStatement", &CompoundAssignStatement{Target: &Identifier{Name: "x"}, Value: &NumberLiteral{}}},
		{"PostIncrementStatement", &PostIncrementStatement{Target: &Identifier{Name: "x"}}},
		{"BinaryExpr", &BinaryExpr{Left: &NumberLiteral{}, Right: &NumberLiteral{}}},
		{"TernaryExpr", &TernaryExpr{Condition: &BoolLiteral{}, TrueExpr: &NumberLiteral{}, FalseExpr: &NumberLiteral{}}},
		{"UnaryExpr", &UnaryExpr{Operand: &NumberLiteral{}}},
		{"CallExpr", &CallExpr{Func: &Identifier{Name: "test"}, Arguments: []Node{}}},
		{"IndexExpr", &IndexExpr{Object: &Identifier{Name: "arr"}, Index: &NumberLiteral{}}},
		{"PropertyAccess", &PropertyAccess{Object: &Identifier{Name: "obj"}, Property: "field"}},
		{"Identifier", &Identifier{Name: "x"}},
		{"NumberLiteral", &NumberLiteral{Value: 42.0}},
		{"StringLiteral", &StringLiteral{Value: "hello"}},
		{"BoolLiteral", &BoolLiteral{Value: true}},
		{"NilLiteral", &NilLiteral{}},
		{"ArrayLiteral", &ArrayLiteral{Elements: []Node{}}},
		{"ObjectLiteral", &ObjectLiteral{Pairs: map[string]Node{}}},
		{"TemplateLiteral", &TemplateLiteral{Parts: []Node{}}},
		{"TextPart", &TextPart{Value: "text"}},
		{"FunctionExpr", &FunctionExpr{Parameters: []*Parameter{}, Body: []Node{}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify it implements the Node interface
			var _ Node = tt.node
		})
	}
}

// TestProgramNode tests Program node
func TestProgramNode(t *testing.T) {
	prog := &Program{
		Statements: []Node{
			&Identifier{Name: "x"},
			&NumberLiteral{Value: 42.0},
		},
	}

	if len(prog.Statements) != 2 {
		t.Errorf("Expected 2 statements, got %d", len(prog.Statements))
	}

	var _ Node = prog
}

// TestIdentifierNode tests Identifier node
func TestIdentifierNode(t *testing.T) {
	ident := &Identifier{Name: "myVar", Pos: Position{Line: 1, Column: 1}}
	if ident.Name != "myVar" {
		t.Errorf("Expected name 'myVar', got %q", ident.Name)
	}
	var _ Node = ident
}

// TestNumberLiteralNode tests NumberLiteral node
func TestNumberLiteralNode(t *testing.T) {
	numLit := &NumberLiteral{Value: 42.5}
	if numLit.Value != 42.5 {
		t.Errorf("Expected 42.5, got %v", numLit.Value)
	}
	var _ Node = numLit
}

// TestStringLiteralNode tests StringLiteral node
func TestStringLiteralNode(t *testing.T) {
	strLit := &StringLiteral{Value: "hello world"}
	if strLit.Value != "hello world" {
		t.Errorf("Expected 'hello world', got %q", strLit.Value)
	}
	var _ Node = strLit
}

// TestBoolLiteralNode tests BoolLiteral node
func TestBoolLiteralNode(t *testing.T) {
	tests := []bool{true, false}
	for _, val := range tests {
		t.Run(string(rune(48 + len(tests))), func(t *testing.T) {
			boolLit := &BoolLiteral{Value: val}
			if boolLit.Value != val {
				t.Errorf("Expected %v, got %v", val, boolLit.Value)
			}
			var _ Node = boolLit
		})
	}
}

// TestNilLiteralNode tests NilLiteral node
func TestNilLiteralNode(t *testing.T) {
	nilLit := &NilLiteral{}
	var _ Node = nilLit
}

// TestArrayLiteralNode tests ArrayLiteral node
func TestArrayLiteralNode(t *testing.T) {
	arrayLit := &ArrayLiteral{
		Elements: []Node{
			&NumberLiteral{Value: 1.0},
			&NumberLiteral{Value: 2.0},
			&NumberLiteral{Value: 3.0},
		},
	}

	if len(arrayLit.Elements) != 3 {
		t.Errorf("Expected 3 elements, got %d", len(arrayLit.Elements))
	}

	var _ Node = arrayLit
}

// TestObjectLiteralNode tests ObjectLiteral node
func TestObjectLiteralNode(t *testing.T) {
	objectLit := &ObjectLiteral{
		Pairs: map[string]Node{
			"x": &NumberLiteral{Value: 1.0},
			"y": &NumberLiteral{Value: 2.0},
		},
	}

	if len(objectLit.Pairs) != 2 {
		t.Errorf("Expected 2 pairs, got %d", len(objectLit.Pairs))
	}

	if _, ok := objectLit.Pairs["x"]; !ok {
		t.Errorf("Expected key 'x' in object")
	}

	var _ Node = objectLit
}

// TestBinaryExprNode tests BinaryExpr node
func TestBinaryExprNode(t *testing.T) {
	binExpr := &BinaryExpr{
		Left:  &NumberLiteral{Value: 5.0},
		Op:    TOK_PLUS,
		Right: &NumberLiteral{Value: 3.0},
	}

	if binExpr.Op != TOK_PLUS {
		t.Errorf("Expected TOK_PLUS operator")
	}

	var _ Node = binExpr
}

// TestUnaryExprNode tests UnaryExpr node
func TestUnaryExprNode(t *testing.T) {
	unaryExpr := &UnaryExpr{
		Op:      TOK_MINUS,
		Operand: &NumberLiteral{Value: 5.0},
	}

	if unaryExpr.Op != TOK_MINUS {
		t.Errorf("Expected TOK_MINUS operator")
	}

	var _ Node = unaryExpr
}

// TestTernaryExprNode tests TernaryExpr node
func TestTernaryExprNode(t *testing.T) {
	ternaryExpr := &TernaryExpr{
		Condition: &BoolLiteral{Value: true},
		TrueExpr:  &NumberLiteral{Value: 1.0},
		FalseExpr: &NumberLiteral{Value: 0.0},
	}

	var _ Node = ternaryExpr
}

// TestCallExprNode tests CallExpr node
func TestCallExprNode(t *testing.T) {
	callExpr := &CallExpr{
		Func: &Identifier{Name: "test"},
		Arguments: []Node{
			&NumberLiteral{Value: 1.0},
			&NumberLiteral{Value: 2.0},
		},
		NamedArgs: map[string]Node{},
	}

	if len(callExpr.Arguments) != 2 {
		t.Errorf("Expected 2 arguments, got %d", len(callExpr.Arguments))
	}

	var _ Node = callExpr
}

// TestIndexExprNode tests IndexExpr node
func TestIndexExprNode(t *testing.T) {
	indexExpr := &IndexExpr{
		Object: &Identifier{Name: "arr"},
		Index:  &NumberLiteral{Value: 0.0},
	}

	var _ Node = indexExpr
}

// TestPropertyAccessNode tests PropertyAccess node
func TestPropertyAccessNode(t *testing.T) {
	propAccess := &PropertyAccess{
		Object:   &Identifier{Name: "obj"},
		Property: "field",
	}

	if propAccess.Property != "field" {
		t.Errorf("Expected property 'field', got %q", propAccess.Property)
	}

	var _ Node = propAccess
}

// TestIfStatementNode tests IfStatement node
func TestIfStatementNode(t *testing.T) {
	ifStmt := &IfStatement{
		Condition: &BoolLiteral{Value: true},
		Then: []Node{
			&NumberLiteral{Value: 42.0},
		},
		Elseifs: []*ElseifClause{
			{
				Condition: &BoolLiteral{Value: false},
				Then: []Node{
					&NumberLiteral{Value: 43.0},
				},
			},
		},
		Else: []Node{
			&NumberLiteral{Value: 44.0},
		},
	}

	if len(ifStmt.Then) != 1 {
		t.Errorf("Expected 1 then statement, got %d", len(ifStmt.Then))
	}

	if len(ifStmt.Elseifs) != 1 {
		t.Errorf("Expected 1 elseif clause, got %d", len(ifStmt.Elseifs))
	}

	if len(ifStmt.Else) != 1 {
		t.Errorf("Expected 1 else statement, got %d", len(ifStmt.Else))
	}

	var _ Node = ifStmt
}

// TestWhileStatementNode tests WhileStatement node
func TestWhileStatementNode(t *testing.T) {
	whileStmt := &WhileStatement{
		Condition: &Identifier{Name: "condition"},
		Body: []Node{
			&NumberLiteral{Value: 1.0},
		},
	}

	if len(whileStmt.Body) != 1 {
		t.Errorf("Expected 1 body statement, got %d", len(whileStmt.Body))
	}

	var _ Node = whileStmt
}

// TestForStatementNode tests ForStatement node
func TestForStatementNode(t *testing.T) {
	forStmt := &ForStatement{
		Var:       "i",
		Start:     &NumberLiteral{Value: 1.0},
		End:       &NumberLiteral{Value: 10.0},
		Step:      &NumberLiteral{Value: 1.0},
		Body:      []Node{&NumberLiteral{Value: 0.0}},
		IsNumeric: true,
	}

	if forStmt.Var != "i" {
		t.Errorf("Expected var 'i', got %q", forStmt.Var)
	}

	if !forStmt.IsNumeric {
		t.Errorf("Expected IsNumeric true")
	}

	var _ Node = forStmt
}

// TestFunctionDefNode tests FunctionDef node
func TestFunctionDefNode(t *testing.T) {
	funcDef := &FunctionDef{
		Name: "test_func",
		Parameters: []*Parameter{
			{Name: "x"},
			{Name: "y", Default: &NumberLiteral{Value: 10.0}},
		},
		Body: []Node{
			&ReturnStatement{Value: &Identifier{Name: "x"}},
		},
	}

	if funcDef.Name != "test_func" {
		t.Errorf("Expected name 'test_func', got %q", funcDef.Name)
	}

	if len(funcDef.Parameters) != 2 {
		t.Errorf("Expected 2 parameters, got %d", len(funcDef.Parameters))
	}

	if funcDef.Parameters[0].Name != "x" {
		t.Errorf("Expected first param 'x', got %q", funcDef.Parameters[0].Name)
	}

	var _ Node = funcDef
}

// TestFunctionExprNode tests FunctionExpr node
func TestFunctionExprNode(t *testing.T) {
	funcExpr := &FunctionExpr{
		Parameters: []*Parameter{
			{Name: "x"},
			{Name: "y", Default: &NumberLiteral{Value: 10.0}},
		},
		Body: []Node{
			&ReturnStatement{Value: &Identifier{Name: "x"}},
		},
	}

	if len(funcExpr.Parameters) != 2 {
		t.Errorf("Expected 2 parameters, got %d", len(funcExpr.Parameters))
	}

	if funcExpr.Parameters[0].Name != "x" {
		t.Errorf("Expected first param 'x', got %q", funcExpr.Parameters[0].Name)
	}

	var _ Node = funcExpr
}

// TestReturnStatementNode tests ReturnStatement node
func TestReturnStatementNode(t *testing.T) {
	returnStmt := &ReturnStatement{
		Value: &NumberLiteral{Value: 42.0},
	}

	var _ Node = returnStmt
}

// TestBreakStatementNode tests BreakStatement node
func TestBreakStatementNode(t *testing.T) {
	breakStmt := &BreakStatement{}
	var _ Node = breakStmt
}

// TestContinueStatementNode tests ContinueStatement node
func TestContinueStatementNode(t *testing.T) {
	continueStmt := &ContinueStatement{}
	var _ Node = continueStmt
}

// TestAssignStatementNode tests AssignStatement node
func TestAssignStatementNode(t *testing.T) {
	assignStmt := &AssignStatement{
		Target:           &Identifier{Name: "x"},
		Value:            &NumberLiteral{Value: 42.0},
		IsVarDeclaration: false,
	}

	if assignStmt.IsVarDeclaration {
		t.Errorf("Expected IsVarDeclaration false")
	}

	var _ Node = assignStmt
}

// TestCompoundAssignStatementNode tests CompoundAssignStatement node
func TestCompoundAssignStatementNode(t *testing.T) {
	compAssignStmt := &CompoundAssignStatement{
		Target:   &Identifier{Name: "x"},
		Operator: TOK_PLUSASSIGN,
		Value:    &NumberLiteral{Value: 5.0},
	}

	if compAssignStmt.Operator != TOK_PLUSASSIGN {
		t.Errorf("Expected TOK_PLUSASSIGN operator")
	}

	var _ Node = compAssignStmt
}

// TestPostIncrementStatementNode tests PostIncrementStatement node
func TestPostIncrementStatementNode(t *testing.T) {
	postIncStmt := &PostIncrementStatement{
		Target:   &Identifier{Name: "x"},
		Operator: TOK_INCREMENT,
	}

	if postIncStmt.Operator != TOK_INCREMENT {
		t.Errorf("Expected TOK_INCREMENT operator")
	}

	var _ Node = postIncStmt
}

// TestTryStatementNode tests TryStatement node
func TestTryStatementNode(t *testing.T) {
	tryStmt := &TryStatement{
		Block: []Node{
			&Identifier{Name: "risky"},
		},
		CatchVar: "err",
		CatchBlock: []Node{
			&Identifier{Name: "handle_error"},
		},
	}

	if len(tryStmt.Block) != 1 {
		t.Errorf("Expected 1 block statement, got %d", len(tryStmt.Block))
	}

	if tryStmt.CatchVar != "err" {
		t.Errorf("Expected catch var 'err', got %q", tryStmt.CatchVar)
	}

	if len(tryStmt.CatchBlock) != 1 {
		t.Errorf("Expected 1 catch statement, got %d", len(tryStmt.CatchBlock))
	}

	var _ Node = tryStmt
}

// TestParameterNode tests Parameter struct
func TestParameterNode(t *testing.T) {
	tests := []struct {
		name    string
		param   *Parameter
		hasDef  bool
	}{
		{
			"param without default",
			&Parameter{Name: "x"},
			false,
		},
		{
			"param with default",
			&Parameter{Name: "y", Default: &NumberLiteral{Value: 10.0}},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.param.Name == "" {
				t.Errorf("Parameter name should not be empty")
			}

			hasDefault := tt.param.Default != nil
			if hasDefault != tt.hasDef {
				t.Errorf("Expected hasDefault=%v, got %v", tt.hasDef, hasDefault)
			}
		})
	}
}

// TestTemplateLiteralNode tests TemplateLiteral and TextPart nodes
func TestTemplateLiteralNode(t *testing.T) {
	templateLit := &TemplateLiteral{
		Parts: []Node{
			&TextPart{Value: "Hello "},
			&Identifier{Name: "name"},
			&TextPart{Value: "!"},
		},
	}

	if len(templateLit.Parts) != 3 {
		t.Errorf("Expected 3 parts, got %d", len(templateLit.Parts))
	}

	var _ Node = templateLit
}

// TestTextPartNode tests TextPart node
func TestTextPartNode(t *testing.T) {
	textPart := &TextPart{Value: "Hello World"}
	if textPart.Value != "Hello World" {
		t.Errorf("Expected 'Hello World', got %q", textPart.Value)
	}
	var _ Node = textPart
}

// TestPositionNode tests Position struct and IsValid method
func TestPositionNode(t *testing.T) {
	pos := Position{Line: 42, Column: 17}

	if pos.Line != 42 {
		t.Errorf("Expected line 42, got %d", pos.Line)
	}

	if pos.Column != 17 {
		t.Errorf("Expected column 17, got %d", pos.Column)
	}

	// Check IsValid method
	if !pos.IsValid() {
		t.Errorf("Position(42, 17) should be valid")
	}

	zeroPos := Position{}
	if zeroPos.IsValid() {
		t.Errorf("Position(0, 0) should not be valid")
	}
}
