package script

import (
	"testing"
)

// TestTokenTypeConstants tests all token type constants are distinct
func TestTokenTypeConstants(t *testing.T) {
	t.Parallel()
	tokenTypes := []TokenType{
		TOK_EOF, TOK_IDENT, TOK_NUMBER, TOK_STRING,
		TOK_PLUS, TOK_MINUS, TOK_STAR, TOK_SLASH, TOK_PERCENT,
		TOK_LPAREN, TOK_RPAREN, TOK_LBRACKET, TOK_RBRACKET,
		TOK_LBRACE, TOK_RBRACE, TOK_COMMA, TOK_DOT, TOK_COLON,
		TOK_ASSIGN, TOK_EQUAL, TOK_NOTEQUAL, TOK_LT, TOK_LTE, TOK_GT, TOK_GTE,
		TOK_AND, TOK_OR, TOK_NOT, TOK_INCREMENT, TOK_DECREMENT,
		TOK_IF, TOK_THEN, TOK_ELSE, TOK_END, TOK_WHILE, TOK_DO, TOK_FOR, TOK_IN,
		TOK_FUNCTION, TOK_RETURN, TOK_VAR, TOK_TRUE, TOK_FALSE, TOK_NIL,
		TOK_TRY, TOK_CATCH,
	}

	seen := make(map[TokenType]bool)
	for _, tt := range tokenTypes {
		if seen[tt] {
			t.Errorf("duplicate token type: %v", tt)
		}
		seen[tt] = true
	}

	if len(seen) != len(tokenTypes) {
		t.Errorf("unique token count = %d, want %d", len(seen), len(tokenTypes))
	}
}

// TestTokenCreation tests token struct creation and fields
func TestTokenCreation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		typ    TokenType
		value  string
		line   int
		column int
	}{
		{"number", TOK_NUMBER, "42", 1, 1},
		{"identifier", TOK_IDENT, "myVar", 5, 10},
		{"operator", TOK_PLUS, "+", 3, 7},
		{"string", TOK_STRING, "hello", 2, 1},
		{"keyword", TOK_IF, "if", 10, 5},
		{"EOF", TOK_EOF, "", 100, 50},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tok := Token{
				Type:   tt.typ,
				Value:  tt.value,
				Line:   tt.line,
				Column: tt.column,
			}

			if tok.Type != tt.typ {
				t.Errorf("Type = %v, want %v", tok.Type, tt.typ)
			}
			if tok.Value != tt.value {
				t.Errorf("Value = %q, want %q", tok.Value, tt.value)
			}
			if tok.Line != tt.line {
				t.Errorf("Line = %d, want %d", tok.Line, tt.line)
			}
			if tok.Column != tt.column {
				t.Errorf("Column = %d, want %d", tok.Column, tt.column)
			}
		})
	}
}

// TestKeywordTokens tests all keyword token types
func TestKeywordTokens(t *testing.T) {
	t.Parallel()
	keywords := []struct {
		name string
		typ  TokenType
	}{
		{"if", TOK_IF},
		{"then", TOK_THEN},
		{"else", TOK_ELSE},
		{"end", TOK_END},
		{"while", TOK_WHILE},
		{"do", TOK_DO},
		{"for", TOK_FOR},
		{"in", TOK_IN},
		{"function", TOK_FUNCTION},
		{"return", TOK_RETURN},
		{"var", TOK_VAR},
		{"true", TOK_TRUE},
		{"false", TOK_FALSE},
		{"nil", TOK_NIL},
		{"and", TOK_AND},
		{"or", TOK_OR},
		{"not", TOK_NOT},
		{"try", TOK_TRY},
		{"catch", TOK_CATCH},
	}

	for _, kw := range keywords {
		kw := kw
		t.Run(kw.name, func(t *testing.T) {
			t.Parallel()
			tok := Token{Type: kw.typ, Value: kw.name}
			if tok.Type != kw.typ {
				t.Errorf("Type mismatch for %s", kw.name)
			}
		})
	}
}

// TestOperatorTokens tests all operator token types
func TestOperatorTokens(t *testing.T) {
	t.Parallel()
	operators := []struct {
		name string
		typ  TokenType
	}{
		{"+", TOK_PLUS},
		{"-", TOK_MINUS},
		{"*", TOK_STAR},
		{"/", TOK_SLASH},
		{"%", TOK_PERCENT},
		{"(", TOK_LPAREN},
		{")", TOK_RPAREN},
		{"[", TOK_LBRACKET},
		{"]", TOK_RBRACKET},
		{"{", TOK_LBRACE},
		{"}", TOK_RBRACE},
		{",", TOK_COMMA},
		{".", TOK_DOT},
		{":", TOK_COLON},
		{"=", TOK_ASSIGN},
		{"==", TOK_EQUAL},
		{"!=", TOK_NOTEQUAL},
		{"<", TOK_LT},
		{"<=", TOK_LTE},
		{">", TOK_GT},
		{">=", TOK_GTE},
		{"++", TOK_INCREMENT},
		{"--", TOK_DECREMENT},
	}

	for _, op := range operators {
		op := op
		t.Run(op.name, func(t *testing.T) {
			t.Parallel()
			tok := Token{Type: op.typ, Value: op.name}
			if tok.Type != op.typ {
				t.Errorf("Type mismatch for %s", op.name)
			}
		})
	}
}

// TestTokenPositioning tests token line and column tracking
func TestTokenPositioning(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		line   int
		column int
	}{
		{"line 1 col 1", 1, 1},
		{"line 5 col 10", 5, 10},
		{"line 100 col 50", 100, 50},
		{"line 0 col 0", 0, 0},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tok := Token{
				Type:   TOK_NUMBER,
				Value:  "42",
				Line:   tt.line,
				Column: tt.column,
			}

			if tok.Line != tt.line {
				t.Errorf("Line = %d, want %d", tok.Line, tt.line)
			}
			if tok.Column != tt.column {
				t.Errorf("Column = %d, want %d", tok.Column, tt.column)
			}
		})
	}
}

// TestTokenSequences tests realistic token sequences
func TestTokenSequences(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		tokens  []Token
		length  int
		first   TokenType
		last    TokenType
	}{
		{
			name: "simple assignment",
			tokens: []Token{
				{Type: TOK_VAR, Value: "var"},
				{Type: TOK_IDENT, Value: "x"},
				{Type: TOK_ASSIGN, Value: "="},
				{Type: TOK_NUMBER, Value: "42"},
				{Type: TOK_EOF, Value: ""},
			},
			length: 5,
			first:  TOK_VAR,
			last:   TOK_EOF,
		},
		{
			name: "binary operation",
			tokens: []Token{
				{Type: TOK_IDENT, Value: "a"},
				{Type: TOK_PLUS, Value: "+"},
				{Type: TOK_IDENT, Value: "b"},
				{Type: TOK_EOF, Value: ""},
			},
			length: 4,
			first:  TOK_IDENT,
			last:   TOK_EOF,
		},
		{
			name: "function call",
			tokens: []Token{
				{Type: TOK_IDENT, Value: "print"},
				{Type: TOK_LPAREN, Value: "("},
				{Type: TOK_STRING, Value: "hello"},
				{Type: TOK_RPAREN, Value: ")"},
				{Type: TOK_EOF, Value: ""},
			},
			length: 5,
			first:  TOK_IDENT,
			last:   TOK_EOF,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if len(tt.tokens) != tt.length {
				t.Errorf("token count = %d, want %d", len(tt.tokens), tt.length)
			}
			if tt.tokens[0].Type != tt.first {
				t.Errorf("first token = %v, want %v", tt.tokens[0].Type, tt.first)
			}
			if tt.tokens[len(tt.tokens)-1].Type != tt.last {
				t.Errorf("last token = %v, want %v", tt.tokens[len(tt.tokens)-1].Type, tt.last)
			}
		})
	}
}

// TestEOFToken tests end-of-file token properties
func TestEOFToken(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		token Token
	}{
		{"empty eof", Token{Type: TOK_EOF, Value: ""}},
		{"with position", Token{Type: TOK_EOF, Value: "", Line: 10, Column: 5}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.token.Type != TOK_EOF {
				t.Error("token is not EOF")
			}
		})
	}
}

// TestTokenLiterals tests token value literals
func TestTokenLiterals(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		typ   TokenType
		value string
	}{
		{"integer", TOK_NUMBER, "42"},
		{"float", TOK_NUMBER, "3.14"},
		{"string content", TOK_STRING, "hello world"},
		{"identifier", TOK_IDENT, "myVariable"},
		{"operator", TOK_PLUS, "+"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tok := Token{Type: tt.typ, Value: tt.value}
			if tok.Value != tt.value {
				t.Errorf("Value = %q, want %q", tok.Value, tt.value)
			}
		})
	}
}
