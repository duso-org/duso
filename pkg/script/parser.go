package script

// parser.go - Duso language parser
//
// This file implements a recursive descent parser that converts a stream of tokens
// (from lexer.go) into an Abstract Syntax Tree (AST) suitable for evaluation.
//
// CORE LANGUAGE COMPONENT: This is part of the minimal core language.
// The parser understands all Duso language syntax including:
// - Variables, objects, arrays
// - Functions, closures, control flow
// - Operators, templates, exception handling
//
// The parser produces AST nodes that the Evaluator executes.

import (
	"fmt"
	"strconv"
	"strings"
)

// BracketInfo tracks opening brackets for better error messages
type BracketInfo struct {
	typ  TokenType
	line int
	col  int
}

type Parser struct {
	tokens        []Token
	pos           int
	bracketStack  []BracketInfo // Track opening brackets for error reporting
	filePath      string        // File path for error reporting
}

func NewParser(tokens []Token) *Parser {
	return &Parser{
		tokens:   tokens,
		pos:      0,
		filePath: "<parser>",
	}
}

// NewParserWithFile creates a parser with an explicit file path for error reporting
func NewParserWithFile(tokens []Token, filePath string) *Parser {
	return &Parser{
		tokens:   tokens,
		pos:      0,
		filePath: filePath,
	}
}

func (p *Parser) current() Token {
	if p.pos >= len(p.tokens) {
		return p.tokens[len(p.tokens)-1] // Return EOF
	}
	return p.tokens[p.pos]
}

// parseError wraps an error with position information
func (p *Parser) parseError(msg string, pos Position) error {
	return &DusoError{
		Message:   msg,
		FilePath:  p.filePath,
		Position:  pos,
		CallStack: make([]CallFrame, 0),
	}
}

func (p *Parser) peek() Token {
	if p.pos+1 >= len(p.tokens) {
		return p.tokens[len(p.tokens)-1]
	}
	return p.tokens[p.pos+1]
}

func (p *Parser) advance() {
	if p.pos < len(p.tokens)-1 {
		p.pos++
	}
}

func (p *Parser) expect(typ TokenType) error {
	if p.current().Type != typ {
		pos := Position{Line: p.current().Line, Column: p.current().Column}
		return p.parseError(fmt.Sprintf("expected %v, got %v", typ, p.current().Type), pos)
	}
	p.advance()
	return nil
}

// pushBracket tracks an opening bracket for better error reporting
func (p *Parser) pushBracket(typ TokenType, line, col int) {
	p.bracketStack = append(p.bracketStack, BracketInfo{typ: typ, line: line, col: col})
}

// expectClosing expects a closing bracket and provides better error messages if it fails
func (p *Parser) expectClosing(openType, closeType TokenType) error {
	if p.current().Type != closeType {
		msg := fmt.Sprintf("expected %v, got %v", closeType, p.current().Type)

		// Add context about where the opening bracket was if we have it
		if len(p.bracketStack) > 0 {
			lastBracket := p.bracketStack[len(p.bracketStack)-1]
			if lastBracket.typ == openType {
				openName := bracketName(openType)
				closeName := bracketName(closeType)
				msg = fmt.Sprintf("expected %s to close %s opened at line %d, col %d, but got %v",
					closeName, openName, lastBracket.line, lastBracket.col, p.current().Type)
			}
		}

		pos := Position{Line: p.current().Line, Column: p.current().Column}
		return p.parseError(msg, pos)
	}

	// Pop the bracket from stack
	if len(p.bracketStack) > 0 {
		p.bracketStack = p.bracketStack[:len(p.bracketStack)-1]
	}

	p.advance()
	return nil
}

// bracketName returns human-readable name for a bracket token
func bracketName(typ TokenType) string {
	switch typ {
	case TOK_LPAREN:
		return "("
	case TOK_RPAREN:
		return ")"
	case TOK_LBRACKET:
		return "["
	case TOK_RBRACKET:
		return "]"
	case TOK_LBRACE:
		return "{"
	case TOK_RBRACE:
		return "}"
	default:
		return fmt.Sprintf("%v", typ)
	}
}

func (p *Parser) match(types ...TokenType) bool {
	for _, typ := range types {
		if p.current().Type == typ {
			return true
		}
	}
	return false
}

func (p *Parser) isCompoundAssign(typ TokenType) bool {
	switch typ {
	case TOK_PLUSASSIGN, TOK_MINUSASSIGN, TOK_STARASSIGN, TOK_SLASHASSIGN, TOK_MODASSIGN:
		return true
	default:
		return false
	}
}

func (p *Parser) Parse() (*Program, error) {
	var statements []Node
	for p.current().Type != TOK_EOF {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			statements = append(statements, stmt)
		}
	}
	return &Program{Statements: statements}, nil
}

func (p *Parser) parseStatement() (Node, error) {
	switch p.current().Type {
	case TOK_IF:
		return p.parseIfStatement()
	case TOK_WHILE:
		return p.parseWhileStatement()
	case TOK_FOR:
		return p.parseForStatement()
	case TOK_FUNCTION:
		return p.parseFunctionDef()
	case TOK_TRY:
		return p.parseTryStatement()
	case TOK_RETURN:
		return p.parseReturnStatement()
	case TOK_BREAK:
		pos := Position{Line: p.current().Line, Column: p.current().Column}
		p.advance()
		return &BreakStatement{Pos: pos}, nil
	case TOK_CONTINUE:
		pos := Position{Line: p.current().Line, Column: p.current().Column}
		p.advance()
		return &ContinueStatement{Pos: pos}, nil
	case TOK_VAR:
		return p.parseVarDeclaration()
	default:
		// Try to parse as assignment or expression statement
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		// Capture position of the expression for assignment statements
		exprPos := Position{}
		if id, ok := expr.(*Identifier); ok {
			exprPos = id.Pos
		}

		// Check if it's an assignment
		if p.current().Type == TOK_ASSIGN {
			p.advance()
			value, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			return &AssignStatement{Pos: exprPos, Target: expr, Value: value}, nil
		}

		// Check for compound assignment operators
		if p.isCompoundAssign(p.current().Type) {
			op := p.current().Type
			p.advance()
			value, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			return &CompoundAssignStatement{Pos: exprPos, Target: expr, Operator: op, Value: value}, nil
		}

		// Check for post-increment/decrement
		if p.current().Type == TOK_INCREMENT || p.current().Type == TOK_DECREMENT {
			op := p.current().Type
			p.advance()
			return &PostIncrementStatement{Pos: exprPos, Target: expr, Operator: op}, nil
		}

		// Just an expression statement (like print(...))
		return expr, nil
	}
}

func (p *Parser) parseIfStatement() (*IfStatement, error) {
	startPos := Position{Line: p.current().Line, Column: p.current().Column}
	p.advance() // skip "if"

	condition, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if err := p.expect(TOK_THEN); err != nil {
		return nil, err
	}

	thenBlock, err := p.parseBlock([]TokenType{TOK_ELSEIF, TOK_ELSE, TOK_END})
	if err != nil {
		return nil, err
	}

	stmt := &IfStatement{
		Pos:       startPos,
		Condition: condition,
		Then:      thenBlock,
	}

	// Parse elseif clauses
	for p.current().Type == TOK_ELSEIF {
		p.advance()
		elifCondition, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		if err := p.expect(TOK_THEN); err != nil {
			return nil, err
		}

		elifBlock, err := p.parseBlock([]TokenType{TOK_ELSEIF, TOK_ELSE, TOK_END})
		if err != nil {
			return nil, err
		}

		stmt.Elseifs = append(stmt.Elseifs, &ElseifClause{
			Condition: elifCondition,
			Then:      elifBlock,
		})
	}

	// Parse else clause
	if p.current().Type == TOK_ELSE {
		p.advance()
		elseBlock, err := p.parseBlock([]TokenType{TOK_END})
		if err != nil {
			return nil, err
		}
		stmt.Else = elseBlock
	}

	if err := p.expect(TOK_END); err != nil {
		return nil, err
	}

	return stmt, nil
}

func (p *Parser) parseWhileStatement() (*WhileStatement, error) {
	startPos := Position{Line: p.current().Line, Column: p.current().Column}
	p.advance() // skip "while"

	condition, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if err := p.expect(TOK_DO); err != nil {
		return nil, err
	}

	body, err := p.parseBlock([]TokenType{TOK_END})
	if err != nil {
		return nil, err
	}

	if err := p.expect(TOK_END); err != nil {
		return nil, err
	}

	return &WhileStatement{Pos: startPos, Condition: condition, Body: body}, nil
}

func (p *Parser) parseForStatement() (*ForStatement, error) {
	startPos := Position{Line: p.current().Line, Column: p.current().Column}
	p.advance() // skip "for"

	// Allow optional "var" keyword before loop variable
	if p.current().Type == TOK_VAR {
		p.advance()
	}

	varName := p.current().Value
	if err := p.expect(TOK_IDENT); err != nil {
		return nil, err
	}

	stmt := &ForStatement{Pos: startPos, Var: varName}

	// Check if it's numeric for or iterator for
	if p.current().Type == TOK_ASSIGN {
		// Numeric: for i = 1, 10 do ... end
		p.advance()
		start, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		if err := p.expect(TOK_COMMA); err != nil {
			return nil, err
		}

		end, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		stmt.Start = start
		stmt.End = end
		stmt.IsNumeric = true

		// Optional step
		if p.current().Type == TOK_COMMA {
			p.advance()
			step, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			stmt.Step = step
		}
	} else if p.current().Type == TOK_IN {
		// Iterator: for item in array do ... end
		p.advance()
		iterator, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		stmt.Iterator = iterator
		stmt.IsNumeric = false
	} else {
		pos := Position{Line: p.current().Line, Column: p.current().Column}
		return nil, p.parseError("expected '=' or 'in' in for loop", pos)
	}

	if err := p.expect(TOK_DO); err != nil {
		return nil, err
	}

	body, err := p.parseBlock([]TokenType{TOK_END})
	if err != nil {
		return nil, err
	}

	if err := p.expect(TOK_END); err != nil {
		return nil, err
	}

	stmt.Body = body
	return stmt, nil
}

func (p *Parser) parseFunctionDef() (*FunctionDef, error) {
	p.advance() // skip "function"

	name := p.current().Value
	if err := p.expect(TOK_IDENT); err != nil {
		return nil, err
	}

	if err := p.expect(TOK_LPAREN); err != nil {
		return nil, err
	}

	var params []*Parameter
	for p.current().Type == TOK_IDENT {
		paramName := p.current().Value
		p.advance()

		var defaultExpr Node = nil
		if p.current().Type == TOK_ASSIGN {
			p.advance()
			expr, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			defaultExpr = expr
		}

		params = append(params, &Parameter{Name: paramName, Default: defaultExpr})

		if p.current().Type == TOK_COMMA {
			p.advance()
		}
	}

	if err := p.expect(TOK_RPAREN); err != nil {
		return nil, err
	}

	body, err := p.parseBlock([]TokenType{TOK_END})
	if err != nil {
		return nil, err
	}

	if err := p.expect(TOK_END); err != nil {
		return nil, err
	}

	// Create function def - we'll set position at function keyword
	return &FunctionDef{Pos: Position{Line: 1, Column: 1}, Name: name, Parameters: params, Body: body}, nil
}

func (p *Parser) parseFunctionExpr() (*FunctionExpr, error) {
	p.advance() // skip "function"

	if err := p.expect(TOK_LPAREN); err != nil {
		return nil, err
	}

	var params []*Parameter
	for p.current().Type == TOK_IDENT {
		paramName := p.current().Value
		p.advance()

		var defaultExpr Node = nil
		if p.current().Type == TOK_ASSIGN {
			p.advance()
			expr, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			defaultExpr = expr
		}

		params = append(params, &Parameter{Name: paramName, Default: defaultExpr})

		if p.current().Type == TOK_COMMA {
			p.advance()
		}
	}

	if err := p.expect(TOK_RPAREN); err != nil {
		return nil, err
	}

	body, err := p.parseBlock([]TokenType{TOK_END})
	if err != nil {
		return nil, err
	}

	if err := p.expect(TOK_END); err != nil {
		return nil, err
	}

	return &FunctionExpr{Parameters: params, Body: body}, nil
}

func (p *Parser) parseTryStatement() (*TryStatement, error) {
	startPos := Position{Line: p.current().Line, Column: p.current().Column}
	p.advance() // skip "try"

	block, err := p.parseBlock([]TokenType{TOK_CATCH})
	if err != nil {
		return nil, err
	}

	if err := p.expect(TOK_CATCH); err != nil {
		return nil, err
	}

	if err := p.expect(TOK_LPAREN); err != nil {
		return nil, err
	}

	catchVar := p.current().Value
	if err := p.expect(TOK_IDENT); err != nil {
		return nil, err
	}

	if err := p.expect(TOK_RPAREN); err != nil {
		return nil, err
	}

	catchBlock, err := p.parseBlock([]TokenType{TOK_END})
	if err != nil {
		return nil, err
	}

	if err := p.expect(TOK_END); err != nil {
		return nil, err
	}

	return &TryStatement{
		Pos:        startPos,
		Block:      block,
		CatchVar:   catchVar,
		CatchBlock: catchBlock,
	}, nil
}

func (p *Parser) parseReturnStatement() (*ReturnStatement, error) {
	startPos := Position{Line: p.current().Line, Column: p.current().Column}
	p.advance() // skip "return"

	var value Node
	if !p.match(TOK_END, TOK_EOF) {
		var err error
		value, err = p.parseExpression()
		if err != nil {
			return nil, err
		}
	}

	return &ReturnStatement{Pos: startPos, Value: value}, nil
}

func (p *Parser) parseVarDeclaration() (*AssignStatement, error) {
	startPos := Position{Line: p.current().Line, Column: p.current().Column}
	p.advance() // skip "var"

	// Expect an identifier
	if p.current().Type != TOK_IDENT {
		pos := Position{Line: p.current().Line, Column: p.current().Column}
		return nil, p.parseError("expected identifier after 'var'", pos)
	}
	name := p.current().Value
	identPos := Position{Line: p.current().Line, Column: p.current().Column}
	p.advance()

	// Expect assignment
	if p.current().Type != TOK_ASSIGN {
		pos := Position{Line: p.current().Line, Column: p.current().Column}
		return nil, p.parseError("expected '=' after variable name", pos)
	}
	p.advance()

	// Parse the value expression
	value, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	return &AssignStatement{
		Pos:              startPos,
		Target:           &Identifier{Pos: identPos, Name: name},
		Value:            value,
		IsVarDeclaration: true,
	}, nil
}

func (p *Parser) parseBlock(terminators []TokenType) ([]Node, error) {
	var statements []Node
	for {
		if p.match(terminators...) || p.current().Type == TOK_EOF {
			break
		}
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			statements = append(statements, stmt)
		}
	}
	return statements, nil
}

func (p *Parser) parseExpression() (Node, error) {
	return p.parseTernary()
}

func (p *Parser) parseTernary() (Node, error) {
	expr, err := p.parseOr()
	if err != nil {
		return nil, err
	}

	// Check for ternary operator
	if p.current().Type == TOK_QUESTION {
		pos := Position{Line: p.current().Line, Column: p.current().Column}
		p.advance() // skip '?'
		trueExpr, err := p.parseOr()
		if err != nil {
			return nil, err
		}

		if err := p.expect(TOK_COLON); err != nil {
			return nil, err
		}

		falseExpr, err := p.parseTernary()
		if err != nil {
			return nil, err
		}

		return &TernaryExpr{
			Pos:       pos,
			Condition: expr,
			TrueExpr:  trueExpr,
			FalseExpr: falseExpr,
		}, nil
	}

	return expr, nil
}

func (p *Parser) parseOr() (Node, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}

	for p.current().Type == TOK_OR {
		op := p.current().Type
		opPos := Position{Line: p.current().Line, Column: p.current().Column}
		p.advance()

		var right Node
		right, err = p.parseAnd()
		if err != nil {
			return nil, p.parseError(err.Error(), opPos)
		}
		left = &BinaryExpr{Pos: opPos, Op: op, Left: left, Right: right}
	}

	return left, nil
}

func (p *Parser) parseAnd() (Node, error) {
	left, err := p.parseEquality()
	if err != nil {
		return nil, err
	}

	for p.current().Type == TOK_AND {
		op := p.current().Type
		opPos := Position{Line: p.current().Line, Column: p.current().Column}
		p.advance()

		var right Node
		right, err = p.parseEquality()
		if err != nil {
			return nil, p.parseError(err.Error(), opPos)
		}
		left = &BinaryExpr{Pos: opPos, Op: op, Left: left, Right: right}
	}

	return left, nil
}

func (p *Parser) parseEquality() (Node, error) {
	left, err := p.parseComparison()
	if err != nil {
		return nil, err
	}

	for p.match(TOK_EQUAL, TOK_NOTEQUAL) {
		op := p.current().Type
		opPos := Position{Line: p.current().Line, Column: p.current().Column}
		p.advance()

		var right Node
		right, err = p.parseComparison()
		if err != nil {
			return nil, p.parseError(err.Error(), opPos)
		}
		left = &BinaryExpr{Pos: opPos, Op: op, Left: left, Right: right}
	}

	return left, nil
}

func (p *Parser) parseComparison() (Node, error) {
	left, err := p.parseAddition()
	if err != nil {
		return nil, err
	}

	for p.match(TOK_LT, TOK_GT, TOK_LTE, TOK_GTE) {
		op := p.current().Type
		opPos := Position{Line: p.current().Line, Column: p.current().Column}
		p.advance()

		var right Node
		right, err = p.parseAddition()
		if err != nil {
			return nil, p.parseError(err.Error(), opPos)
		}
		left = &BinaryExpr{Pos: opPos, Op: op, Left: left, Right: right}
	}

	return left, nil
}

func (p *Parser) parseAddition() (Node, error) {
	left, err := p.parseMultiplication()
	if err != nil {
		return nil, err
	}

	for p.match(TOK_PLUS, TOK_MINUS) {
		op := p.current().Type
		opPos := Position{Line: p.current().Line, Column: p.current().Column}
		p.advance()

		var right Node
		right, err = p.parseMultiplication()
		if err != nil {
			return nil, p.parseError(err.Error(), opPos)
		}
		left = &BinaryExpr{Pos: opPos, Op: op, Left: left, Right: right}
	}

	return left, nil
}

func (p *Parser) parseMultiplication() (Node, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}

	for p.match(TOK_STAR, TOK_SLASH, TOK_PERCENT) {
		op := p.current().Type
		opPos := Position{Line: p.current().Line, Column: p.current().Column}
		p.advance()

		var right Node
		right, err = p.parseUnary()
		if err != nil {
			return nil, p.parseError(err.Error(), opPos)
		}
		left = &BinaryExpr{Pos: opPos, Op: op, Left: left, Right: right}
	}

	return left, nil
}

func (p *Parser) parseUnary() (Node, error) {
	if p.current().Type == TOK_NOT {
		pos := Position{Line: p.current().Line, Column: p.current().Column}
		p.advance()
		operand, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &UnaryExpr{Pos: pos, Op: TOK_NOT, Operand: operand}, nil
	}

	if p.current().Type == TOK_MINUS {
		pos := Position{Line: p.current().Line, Column: p.current().Column}
		p.advance()
		operand, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &UnaryExpr{Pos: pos, Op: TOK_MINUS, Operand: operand}, nil
	}

	// Handle pre-increment and pre-decrement
	if p.current().Type == TOK_INCREMENT || p.current().Type == TOK_DECREMENT {
		op := p.current().Type
		pos := Position{Line: p.current().Line, Column: p.current().Column}
		p.advance()
		operand, err := p.parsePostfix()
		if err != nil {
			return nil, err
		}
		return &UnaryExpr{Pos: pos, Op: op, Operand: operand}, nil
	}

	return p.parsePostfix()
}

func (p *Parser) parsePostfix() (Node, error) {
	expr, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	for {
		var err error
		var pos Position

		switch p.current().Type {
		case TOK_LPAREN:
			// Function call or structure instantiation
			pos = Position{Line: p.current().Line, Column: p.current().Column}
			expr, err = p.parseCall(expr)

		case TOK_LBRACKET:
			// Array indexing
			pos = Position{Line: p.current().Line, Column: p.current().Column}
			p.advance()
			var index Node
			index, err = p.parseExpression()
			if err == nil {
				err = p.expect(TOK_RBRACKET)
				if err == nil {
					expr = &IndexExpr{Pos: pos, Object: expr, Index: index}
				}
			}

		case TOK_DOT:
			// Property access
			pos = Position{Line: p.current().Line, Column: p.current().Column}
			p.advance()
			propName := p.current().Value
			err = p.expect(TOK_IDENT)
			if err == nil {
				expr = &PropertyAccess{Pos: pos, Object: expr, Property: propName}
			}

		default:
			return expr, nil
		}

		if err != nil {
			return nil, p.parseError(err.Error(), pos)
		}
	}
}

func (p *Parser) parseCall(expr Node) (Node, error) {
	pos := Position{Line: p.current().Line, Column: p.current().Column}
	p.advance() // skip "("

	var args []Node
	namedArgs := make(map[string]Node)

	for p.current().Type != TOK_RPAREN && p.current().Type != TOK_EOF {
		// Check if this is a named argument (identifier or string key)
		if (p.current().Type == TOK_IDENT || p.current().Type == TOK_STRING) && p.peek().Type == TOK_ASSIGN {
			name := p.current().Value
			p.advance()
			p.advance() // skip "="
			value, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			namedArgs[name] = value
		} else {
			arg, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
		}

		if p.current().Type == TOK_COMMA {
			p.advance()
		} else if p.current().Type != TOK_RPAREN {
			errPos := Position{Line: p.current().Line, Column: p.current().Column}
			return nil, p.parseError("expected ',' or ')' in function call", errPos)
		}
	}

	if err := p.expect(TOK_RPAREN); err != nil {
		return nil, err
	}

	return &CallExpr{Pos: pos, Func: expr, Arguments: args, NamedArgs: namedArgs}, nil
}

func (p *Parser) parsePrimary() (Node, error) {
	switch p.current().Type {
	case TOK_FUNCTION:
		return p.parseFunctionExpr()

	case TOK_NUMBER:
		value, _ := strconv.ParseFloat(p.current().Value, 64)
		p.advance()
		return &NumberLiteral{Value: value}, nil

	case TOK_RAW:
		// raw "string" - return the string literal without template evaluation
		p.advance()
		if p.current().Type != TOK_STRING {
			errPos := Position{Line: p.current().Line, Column: p.current().Column}
			return nil, p.parseError("expected string after 'raw' keyword", errPos)
		}
		rawValue := p.current().Value
		p.advance()
		// Just return the string unescaped, no template parsing
		return &StringLiteral{Value: UnescapeString(rawValue)}, nil

	case TOK_STRING:
		rawValue := p.current().Value
		pos := Position{Line: p.current().Line, Column: p.current().Column}
		p.advance()

		// Check if this is a template string (contains {{ }})
		if strings.Contains(rawValue, "{{") {
			return p.ParseTemplateString(rawValue, pos)
		}
		// Not a template - unescape and return as regular string
		return &StringLiteral{Value: UnescapeString(rawValue)}, nil

	case TOK_TILDE_STRING:
		// Tilde strings are raw - no template evaluation, no unescaping (except \~)
		rawValue := p.current().Value
		p.advance()
		// Return as-is, no unescaping
		return &StringLiteral{Value: rawValue}, nil

	case TOK_TRUE:
		p.advance()
		return &BoolLiteral{Value: true}, nil

	case TOK_FALSE:
		p.advance()
		return &BoolLiteral{Value: false}, nil

	case TOK_NIL:
		p.advance()
		return &NilLiteral{}, nil

	case TOK_IDENT:
		pos := Position{Line: p.current().Line, Column: p.current().Column}
		name := p.current().Value
		p.advance()
		return &Identifier{Pos: pos, Name: name}, nil

	case TOK_LPAREN:
		openLine := p.current().Line
		openCol := p.current().Column
		p.advance()
		p.pushBracket(TOK_LPAREN, openLine, openCol)
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if err := p.expectClosing(TOK_LPAREN, TOK_RPAREN); err != nil {
			return nil, err
		}
		return expr, nil

	case TOK_LBRACKET:
		// Array literal
		openLine := p.current().Line
		openCol := p.current().Column
		p.advance()
		p.pushBracket(TOK_LBRACKET, openLine, openCol)
		var elements []Node
		for p.current().Type != TOK_RBRACKET && p.current().Type != TOK_EOF {
			elem, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			elements = append(elements, elem)

			if p.current().Type == TOK_COMMA {
				p.advance()
			} else if p.current().Type != TOK_RBRACKET {
				errPos := Position{Line: p.current().Line, Column: p.current().Column}
			return nil, p.parseError("expected ',' or ']' in array literal", errPos)
			}
		}
		if err := p.expectClosing(TOK_LBRACKET, TOK_RBRACKET); err != nil {
			return nil, err
		}
		return &ArrayLiteral{Elements: elements}, nil

	case TOK_LBRACE:
		// Object literal - supports both identifier and string keys
		openLine := p.current().Line
		openCol := p.current().Column
		p.advance()
		p.pushBracket(TOK_LBRACE, openLine, openCol)
		pairs := make(map[string]Node)
		for p.current().Type != TOK_RBRACE && p.current().Type != TOK_EOF {
			// Accept either identifier or string as key
			var key string
			if p.current().Type == TOK_IDENT {
				key = p.current().Value
				p.advance()
			} else if p.current().Type == TOK_STRING {
				key = p.current().Value
				p.advance()
			} else {
				errPos := Position{Line: p.current().Line, Column: p.current().Column}
				return nil, p.parseError("expected identifier or string as object key", errPos)
			}

			// Use '=' for consistency with named arguments and assignments
			if err := p.expect(TOK_ASSIGN); err != nil {
				return nil, err
			}

			value, err := p.parseExpression()
			if err != nil {
				return nil, err
			}

			pairs[key] = value

			if p.current().Type == TOK_COMMA {
				p.advance()
			} else if p.current().Type != TOK_RBRACE {
				errPos := Position{Line: p.current().Line, Column: p.current().Column}
				return nil, p.parseError("expected ',' or '}' in object literal", errPos)
			}
		}
		if err := p.expectClosing(TOK_LBRACE, TOK_RBRACE); err != nil {
			return nil, err
		}
		return &ObjectLiteral{Pairs: pairs}, nil

	default:
		pos := Position{Line: p.current().Line, Column: p.current().Column}
		return nil, p.parseError(fmt.Sprintf("unexpected token %s", p.current().String()), pos)
	}
}

// ParseTemplateString parses a template string containing {{ }} expressions
func (p *Parser) ParseTemplateString(template string, pos Position) (Node, error) {
	var parts []Node

	// Split template by {{ and }}
	i := 0
	for i < len(template) {
		// Find next template expression
		start := strings.Index(template[i:], "{{")
		if start == -1 {
			// No more templates - add remaining text (unescaped)
			if i < len(template) {
				parts = append(parts, &TextPart{Value: UnescapeString(template[i:])})
			}
			break
		}

		// Add text before template (unescaped)
		if start > 0 {
			parts = append(parts, &TextPart{Value: UnescapeString(template[i : i+start])})
		}

		// Find closing }}
		exprStart := i + start + 2
		end := strings.Index(template[exprStart:], "}}")
		if end == -1 {
			pos := Position{Line: 1, Column: exprStart}
			return nil, p.parseError("unclosed {{ in template string", pos)
		}

		// Extract and parse expression (raw, no unescaping for expressions)
		exprStr := template[exprStart : exprStart+end]
		exprLexer := NewLexer(exprStr)
		exprTokens := exprLexer.Tokenize()
		exprParser := NewParser(exprTokens)
		expr, err := exprParser.parseExpression()
		if err != nil {
			pos := Position{Line: 1, Column: exprStart}
			return nil, p.parseError(fmt.Sprintf("error in template expression: %v", err), pos)
		}

		parts = append(parts, expr)

		// Move past }}
		i = exprStart + end + 2
	}

	if len(parts) == 0 {
		return &StringLiteral{Value: ""}, nil
	}

	// If only one part and it's text, return as string literal
	if len(parts) == 1 {
		if tp, ok := parts[0].(*TextPart); ok {
			return &StringLiteral{Value: tp.Value}, nil
		}
	}

	return &TemplateLiteral{Pos: pos, Parts: parts}, nil
}
