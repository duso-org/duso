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

type Parser struct {
	tokens []Token
	pos    int
}

func NewParser(tokens []Token) *Parser {
	return &Parser{
		tokens: tokens,
		pos:    0,
	}
}

func (p *Parser) current() Token {
	if p.pos >= len(p.tokens) {
		return p.tokens[len(p.tokens)-1] // Return EOF
	}
	return p.tokens[p.pos]
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
		return fmt.Errorf("expected %v, got %v at line %d", typ, p.current().Type, p.current().Line)
	}
	p.advance()
	return nil
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
		p.advance()
		return &BreakStatement{}, nil
	case TOK_CONTINUE:
		p.advance()
		return &ContinueStatement{}, nil
	case TOK_VAR:
		return p.parseVarDeclaration()
	default:
		// Try to parse as assignment or expression statement
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		// Check if it's an assignment
		if p.current().Type == TOK_ASSIGN {
			p.advance()
			value, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			return &AssignStatement{Target: expr, Value: value}, nil
		}

		// Check for compound assignment operators
		if p.isCompoundAssign(p.current().Type) {
			op := p.current().Type
			p.advance()
			value, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			return &CompoundAssignStatement{Target: expr, Operator: op, Value: value}, nil
		}

		// Check for post-increment/decrement
		if p.current().Type == TOK_INCREMENT || p.current().Type == TOK_DECREMENT {
			op := p.current().Type
			p.advance()
			return &PostIncrementStatement{Target: expr, Operator: op}, nil
		}

		// Just an expression statement (like print(...))
		return expr, nil
	}
}

func (p *Parser) parseIfStatement() (*IfStatement, error) {
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

	return &WhileStatement{Condition: condition, Body: body}, nil
}

func (p *Parser) parseForStatement() (*ForStatement, error) {
	p.advance() // skip "for"

	varName := p.current().Value
	if err := p.expect(TOK_IDENT); err != nil {
		return nil, err
	}

	stmt := &ForStatement{Var: varName}

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
		return nil, fmt.Errorf("expected '=' or 'in' in for loop at line %d", p.current().Line)
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

	return &FunctionDef{Name: name, Parameters: params, Body: body}, nil
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
		Block:      block,
		CatchVar:   catchVar,
		CatchBlock: catchBlock,
	}, nil
}

func (p *Parser) parseReturnStatement() (*ReturnStatement, error) {
	p.advance() // skip "return"

	var value Node
	if !p.match(TOK_END, TOK_EOF) {
		var err error
		value, err = p.parseExpression()
		if err != nil {
			return nil, err
		}
	}

	return &ReturnStatement{Value: value}, nil
}

func (p *Parser) parseVarDeclaration() (*AssignStatement, error) {
	p.advance() // skip "var"

	// Expect an identifier
	if p.current().Type != TOK_IDENT {
		return nil, fmt.Errorf("expected identifier after 'var', got %v", p.current())
	}
	name := p.current().Value
	p.advance()

	// Expect assignment
	if p.current().Type != TOK_ASSIGN {
		return nil, fmt.Errorf("expected '=' after variable name, got %v", p.current())
	}
	p.advance()

	// Parse the value expression
	value, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	return &AssignStatement{
		Target:           &Identifier{Name: name},
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
		p.advance()
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Op: op, Left: left, Right: right}
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
		p.advance()
		right, err := p.parseEquality()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Op: op, Left: left, Right: right}
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
		p.advance()
		right, err := p.parseComparison()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Op: op, Left: left, Right: right}
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
		p.advance()
		right, err := p.parseAddition()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Op: op, Left: left, Right: right}
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
		p.advance()
		right, err := p.parseMultiplication()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Op: op, Left: left, Right: right}
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
		p.advance()
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Op: op, Left: left, Right: right}
	}

	return left, nil
}

func (p *Parser) parseUnary() (Node, error) {
	if p.current().Type == TOK_NOT {
		p.advance()
		operand, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &UnaryExpr{Op: TOK_NOT, Operand: operand}, nil
	}

	if p.current().Type == TOK_MINUS {
		p.advance()
		operand, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &UnaryExpr{Op: TOK_MINUS, Operand: operand}, nil
	}

	// Handle pre-increment and pre-decrement
	if p.current().Type == TOK_INCREMENT || p.current().Type == TOK_DECREMENT {
		op := p.current().Type
		p.advance()
		operand, err := p.parsePostfix()
		if err != nil {
			return nil, err
		}
		return &UnaryExpr{Op: op, Operand: operand}, nil
	}

	return p.parsePostfix()
}

func (p *Parser) parsePostfix() (Node, error) {
	expr, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	for {
		switch p.current().Type {
		case TOK_LPAREN:
			// Function call or structure instantiation
			expr, err = p.parseCall(expr)
			if err != nil {
				return nil, err
			}
		case TOK_LBRACKET:
			// Array indexing
			p.advance()
			index, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			if err := p.expect(TOK_RBRACKET); err != nil {
				return nil, err
			}
			expr = &IndexExpr{Object: expr, Index: index}
		case TOK_DOT:
			// Property access
			p.advance()
			propName := p.current().Value
			if err := p.expect(TOK_IDENT); err != nil {
				return nil, err
			}
			expr = &PropertyAccess{Object: expr, Property: propName}
		default:
			return expr, nil
		}
	}
}

func (p *Parser) parseCall(expr Node) (Node, error) {
	p.advance() // skip "("

	var args []Node
	namedArgs := make(map[string]Node)

	for p.current().Type != TOK_RPAREN && p.current().Type != TOK_EOF {
		// Check if this is a named argument
		if p.current().Type == TOK_IDENT && p.peek().Type == TOK_ASSIGN {
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
			return nil, fmt.Errorf("expected ',' or ')' in call at line %d", p.current().Line)
		}
	}

	if err := p.expect(TOK_RPAREN); err != nil {
		return nil, err
	}

	return &CallExpr{Func: expr, Arguments: args, NamedArgs: namedArgs}, nil
}

func (p *Parser) parsePrimary() (Node, error) {
	switch p.current().Type {
	case TOK_FUNCTION:
		return p.parseFunctionExpr()

	case TOK_NUMBER:
		value, _ := strconv.ParseFloat(p.current().Value, 64)
		p.advance()
		return &NumberLiteral{Value: value}, nil

	case TOK_STRING:
		rawValue := p.current().Value
		p.advance()

		// Check if this is a template string (contains {{ }})
		if strings.Contains(rawValue, "{{") {
			return p.parseTemplateString(rawValue)
		}
		// Not a template - unescape and return as regular string
		return &StringLiteral{Value: UnescapeString(rawValue)}, nil

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
		name := p.current().Value
		p.advance()
		return &Identifier{Name: name}, nil

	case TOK_LPAREN:
		p.advance()
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if err := p.expect(TOK_RPAREN); err != nil {
			return nil, err
		}
		return expr, nil

	case TOK_LBRACKET:
		// Array literal
		p.advance()
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
				return nil, fmt.Errorf("expected ',' or ']' in array literal at line %d", p.current().Line)
			}
		}
		if err := p.expect(TOK_RBRACKET); err != nil {
			return nil, err
		}
		return &ArrayLiteral{Elements: elements}, nil

	case TOK_LBRACE:
		// Object literal
		p.advance()
		pairs := make(map[string]Node)
		for p.current().Type != TOK_RBRACE && p.current().Type != TOK_EOF {
			key := p.current().Value
			if err := p.expect(TOK_IDENT); err != nil {
				return nil, err
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
				return nil, fmt.Errorf("expected ',' or '}' in object literal at line %d", p.current().Line)
			}
		}
		if err := p.expect(TOK_RBRACE); err != nil {
			return nil, err
		}
		return &ObjectLiteral{Pairs: pairs}, nil

	default:
		return nil, fmt.Errorf("unexpected token %v at line %d", p.current().Type, p.current().Line)
	}
}

func (p *Parser) parseTemplateString(template string) (Node, error) {
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
			return nil, fmt.Errorf("unclosed {{ in template string")
		}

		// Extract and parse expression (raw, no unescaping for expressions)
		exprStr := template[exprStart : exprStart+end]
		exprLexer := NewLexer(exprStr)
		exprTokens := exprLexer.Tokenize()
		exprParser := NewParser(exprTokens)
		expr, err := exprParser.parseExpression()
		if err != nil {
			return nil, fmt.Errorf("error in template expression: %w", err)
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

	return &TemplateLiteral{Parts: parts}, nil
}
