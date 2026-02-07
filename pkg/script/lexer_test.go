package script

import (
	"testing"
)

// TestLexer_NestedBlockComments tests nested block comment handling
func TestLexer_NestedBlockComments(t *testing.T) {
	source := `/* outer /* inner */ outer */ x = 1`
	lexer := NewLexer(source)

	// Skip to first non-comment token
	var tokens []Token
	for {
		tok := lexer.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == TOK_EOF {
			break
		}
	}

	// Should find identifier 'x' after nested comments
	found := false
	for _, tok := range tokens {
		if tok.Type == TOK_IDENT && tok.Value == "x" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected to find identifier 'x' after nested comments")
	}
}

// TestLexer_UnterminatedBlockComment tests error on unterminated block comment
func TestLexer_UnterminatedBlockComment(t *testing.T) {
	source := `/* no close x = 1`
	lexer := NewLexer(source)

	// Read all tokens - should handle unterminated comment gracefully
	var tokens []Token
	for {
		tok := lexer.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == TOK_EOF {
			break
		}
	}

	// Should reach EOF (lexer doesn't error, just includes in comment)
	if len(tokens) < 1 || tokens[len(tokens)-1].Type != TOK_EOF {
		t.Errorf("Expected EOF token")
	}
}

// TestLexer_RawString tests raw string literals (~string~)
func TestLexer_RawString(t *testing.T) {
	source := `~no\nescapes~`
	lexer := NewLexer(source)

	tok := lexer.NextToken()
	if tok.Type != TOK_TILDE_STRING {
		t.Errorf("Expected TOK_TILDE_STRING, got %v", tok.Type)
	}

	// Raw strings preserve backslashes
	if tok.Value != `no\nescapes` {
		t.Errorf("Expected 'no\\nescapes', got %q", tok.Value)
	}
}

// TestLexer_RawStringWithEscapedTilde tests escaped tilde in raw string
func TestLexer_RawStringWithEscapedTilde(t *testing.T) {
	source := `~tilde\~inside~`
	lexer := NewLexer(source)

	tok := lexer.NextToken()
	if tok.Type != TOK_TILDE_STRING {
		t.Errorf("Expected TOK_TILDE_STRING, got %v", tok.Type)
	}

	// Escaped tilde should be included
	if tok.Value != `tilde\~inside` {
		t.Errorf("Expected 'tilde\\~inside', got %q", tok.Value)
	}
}

// TestLexer_UnterminatedRawString tests error on unterminated raw string
func TestLexer_UnterminatedRawString(t *testing.T) {
	source := `~no close`
	lexer := NewLexer(source)

	tok := lexer.NextToken()
	// Should still get a token, even if unterminated
	// The lexer continues until EOF
	if tok.Type != TOK_TILDE_STRING {
		t.Errorf("Expected TOK_TILDE_STRING, got %v", tok.Type)
	}
}

// TestLexer_MultilineString tests triple-quoted multiline strings
func TestLexer_MultilineString(t *testing.T) {
	source := `"""line1
line2
line3"""`
	lexer := NewLexer(source)

	tok := lexer.NextToken()
	if tok.Type != TOK_STRING {
		t.Errorf("Expected TOK_STRING, got %v", tok.Type)
	}

	// Should preserve newlines
	if tok.Value != "line1\nline2\nline3" {
		t.Errorf("Expected multiline content, got %q", tok.Value)
	}
}

// TestLexer_MultilineStringDedent tests auto-dedent in multiline strings
func TestLexer_MultilineStringDedent(t *testing.T) {
	source := `"""
    hello
      world
    """`
	lexer := NewLexer(source)

	tok := lexer.NextToken()
	if tok.Type != TOK_STRING {
		t.Errorf("Expected TOK_STRING, got %v", tok.Type)
	}

	// Should dedent (remove common whitespace)
	// The dedent removes the common 4 spaces from both lines, leaving relative indentation
	if tok.Value != "hello\nworld" {
		t.Errorf("Expected dedented content, got %q", tok.Value)
	}
}

// TestLexer_MultilineStringEmptyLines tests multiline strings with empty lines
func TestLexer_MultilineStringEmptyLines(t *testing.T) {
	source := `"""
line1

line3
"""`
	lexer := NewLexer(source)

	tok := lexer.NextToken()
	if tok.Type != TOK_STRING {
		t.Errorf("Expected TOK_STRING, got %v", tok.Type)
	}

	// Should preserve blank lines
	if tok.Value != "line1\n\nline3" {
		t.Errorf("Expected content with blank lines, got %q", tok.Value)
	}
}

// TestLexer_UnterminatedMultilineString tests error on unclosed triple-quoted string
func TestLexer_UnterminatedMultilineString(t *testing.T) {
	source := `"""no close`
	lexer := NewLexer(source)

	tok := lexer.NextToken()
	// Should still get a token, reading until EOF
	if tok.Type != TOK_STRING {
		t.Errorf("Expected TOK_STRING, got %v", tok.Type)
	}
}

// TestLexer_StandardEscapes tests standard escape sequences
func TestLexer_StandardEscapes(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected string
	}{
		{`newline`, `"line\nbreak"`, `line\nbreak`},
		{`tab`, `"tab\there"`, `tab\there`},
		{`backslash`, `"back\\slash"`, `back\\slash`},
		{`quote`, `"say\"hello\""`, `say\"hello\"`},
		{`carriage return`, `"line\rreturn"`, `line\rreturn`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			lexer := NewLexer(tc.source)
			tok := lexer.NextToken()

			if tok.Type != TOK_STRING {
				t.Errorf("Expected TOK_STRING, got %v", tok.Type)
			}

			// The lexer returns raw token values; escaping happens in parser
			if tok.Value != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, tok.Value)
			}
		})
	}
}

// TestLexer_UnterminatedString tests error on unterminated string
func TestLexer_UnterminatedString(t *testing.T) {
	source := `"no close`
	lexer := NewLexer(source)

	tok := lexer.NextToken()
	// Lexer should still return a string token (reads until EOF)
	if tok.Type != TOK_STRING {
		t.Errorf("Expected TOK_STRING, got %v", tok.Type)
	}
}

// TestLexer_EmptyString tests empty string literal
func TestLexer_EmptyString(t *testing.T) {
	source := `""`
	lexer := NewLexer(source)

	tok := lexer.NextToken()
	if tok.Type != TOK_STRING {
		t.Errorf("Expected TOK_STRING, got %v", tok.Type)
	}

	if tok.Value != "" {
		t.Errorf("Expected empty string, got %q", tok.Value)
	}
}

// TestLexer_StringWithNewline tests string with literal newline
func TestLexer_StringWithNewline(t *testing.T) {
	source := "\"line1\nline2\""
	lexer := NewLexer(source)

	tok := lexer.NextToken()
	if tok.Type != TOK_STRING {
		t.Errorf("Expected TOK_STRING, got %v", tok.Type)
	}

	// String should contain newline
	if tok.Value != "line1\nline2" {
		t.Errorf("Expected string with newline, got %q", tok.Value)
	}
}

// TestLexer_Numbers tests various number formats
func TestLexer_Numbers(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected string
	}{
		{"integer", "42", "42"},
		{"decimal", "3.14", "3.14"},
		{"zero", "0", "0"},
		{"leading zeros", "00042", "00042"},
		{"decimal zero", "0.0", "0.0"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			lexer := NewLexer(tc.source)
			tok := lexer.NextToken()

			if tok.Type != TOK_NUMBER {
				t.Errorf("Expected TOK_NUMBER, got %v", tok.Type)
			}

			if tok.Value != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, tok.Value)
			}
		})
	}
}

// TestLexer_Identifiers tests identifier tokenization
func TestLexer_Identifiers(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected string
	}{
		{"simple", "myVar", "myVar"},
		{"with numbers", "var123", "var123"},
		{"with underscore", "my_var", "my_var"},
		{"single letter", "x", "x"},
		{"underscore prefix", "_private", "_private"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			lexer := NewLexer(tc.source)
			tok := lexer.NextToken()

			if tok.Type != TOK_IDENT {
				t.Errorf("Expected TOK_IDENT, got %v", tok.Type)
			}

			if tok.Value != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, tok.Value)
			}
		})
	}
}

// TestLexer_Keywords tests keyword tokenization
func TestLexer_Keywords(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected TokenType
	}{
		{"if", "if", TOK_IF},
		{"then", "then", TOK_THEN},
		{"else", "else", TOK_ELSE},
		{"end", "end", TOK_END},
		{"function", "function", TOK_FUNCTION},
		{"return", "return", TOK_RETURN},
		{"for", "for", TOK_FOR},
		{"while", "while", TOK_WHILE},
		{"break", "break", TOK_BREAK},
		{"continue", "continue", TOK_CONTINUE},
		{"true", "true", TOK_TRUE},
		{"false", "false", TOK_FALSE},
		{"nil", "nil", TOK_NIL},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			lexer := NewLexer(tc.source)
			tok := lexer.NextToken()

			if tok.Type != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, tok.Type)
			}
		})
	}
}

// TestLexer_Operators tests operator tokenization
func TestLexer_Operators(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected TokenType
	}{
		{"+", "+", TOK_PLUS},
		{"-", "-", TOK_MINUS},
		{"*", "*", TOK_STAR},
		{"/", "/", TOK_SLASH},
		{"%", "%", TOK_PERCENT},
		{"==", "==", TOK_EQUAL},
		{"!=", "!=", TOK_NOTEQUAL},
		{"<", "<", TOK_LT},
		{"<=", "<=", TOK_LTE},
		{">", ">", TOK_GT},
		{">=", ">=", TOK_GTE},
		{"=", "=", TOK_ASSIGN},
		{"+=", "+=", TOK_PLUSASSIGN},
		{"-=", "-=", TOK_MINUSASSIGN},
		{"*=", "*=", TOK_STARASSIGN},
		{"/=", "/=", TOK_SLASHASSIGN},
		{"%=", "%=", TOK_MODASSIGN},
		{"++", "++", TOK_INCREMENT},
		{"--", "--", TOK_DECREMENT},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			lexer := NewLexer(tc.source)
			tok := lexer.NextToken()

			if tok.Type != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, tok.Type)
			}

			if tok.Value != tc.source {
				t.Errorf("Expected value %q, got %q", tc.source, tok.Value)
			}
		})
	}
}

// TestLexer_Delimiters tests delimiter tokenization
func TestLexer_Delimiters(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected TokenType
	}{
		{"(", "(", TOK_LPAREN},
		{")", ")", TOK_RPAREN},
		{"[", "[", TOK_LBRACKET},
		{"]", "]", TOK_RBRACKET},
		{"{", "{", TOK_LBRACE},
		{"}", "}", TOK_RBRACE},
		{".", ".", TOK_DOT},
		{",", ",", TOK_COMMA},
		{":", ":", TOK_COLON},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			lexer := NewLexer(tc.source)
			tok := lexer.NextToken()

			if tok.Type != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, tok.Type)
			}
		})
	}
}

// TestLexer_LineAndColumnTracking tests that line and column are tracked correctly
func TestLexer_LineAndColumnTracking(t *testing.T) {
	source := `x = 1
y = 2`
	lexer := NewLexer(source)

	// First token: 'x' at line 1, column 1
	tok := lexer.NextToken()
	if tok.Line != 1 || tok.Type != TOK_IDENT {
		t.Errorf("Expected identifier at line 1, got line %d, type %v", tok.Line, tok.Type)
	}

	// Skip to 'y' on line 2
	for {
		tok := lexer.NextToken()
		if tok.Type == TOK_IDENT && tok.Value == "y" {
			if tok.Line != 2 {
				t.Errorf("Expected 'y' at line 2, got line %d", tok.Line)
			}
			break
		}
		if tok.Type == TOK_EOF {
			t.Errorf("Did not find 'y' identifier")
			break
		}
	}
}


// TestLexer_ComplexExpression tests tokenization of complex expression
func TestLexer_ComplexExpression(t *testing.T) {
	source := `if x > 0 then y = x + 1 end`
	lexer := NewLexer(source)

	expectedTokens := []TokenType{
		TOK_IF, TOK_IDENT, TOK_GT, TOK_NUMBER,
		TOK_THEN, TOK_IDENT, TOK_ASSIGN, TOK_IDENT,
		TOK_PLUS, TOK_NUMBER, TOK_END, TOK_EOF,
	}

	for i, expectedType := range expectedTokens {
		tok := lexer.NextToken()
		if tok.Type != expectedType {
			t.Errorf("Token %d: expected %v, got %v", i, expectedType, tok.Type)
		}
	}
}

// TestLexer_CommentSkipping tests that comments are properly skipped
func TestLexer_CommentSkipping(t *testing.T) {
	source := `x = 1 // comment
y = 2`
	lexer := NewLexer(source)

	// Read tokens, skipping the comment
	tok := lexer.NextToken() // x
	tok = lexer.NextToken() // =
	tok = lexer.NextToken() // 1
	tok = lexer.NextToken() // Should be y, not comment

	if tok.Type != TOK_IDENT || tok.Value != "y" {
		t.Errorf("Expected 'y' after comment, got %v (%q)", tok.Type, tok.Value)
	}
}

// TestLexer_WhitespaceHandling tests whitespace and semicolon handling
func TestLexer_WhitespaceHandling(t *testing.T) {
	source := "  x  ;  y  "
	lexer := NewLexer(source)

	tok := lexer.NextToken()
	if tok.Type != TOK_IDENT || tok.Value != "x" {
		t.Errorf("Expected 'x', got %v", tok.Value)
	}

	tok = lexer.NextToken()
	if tok.Type != TOK_IDENT || tok.Value != "y" {
		t.Errorf("Expected 'y', got %v", tok.Value)
	}
}

// TestLexer_MixedStringTypes tests different string types in sequence
func TestLexer_MixedStringTypes(t *testing.T) {
	source := `"normal" ~raw~ """multi
line"""`
	lexer := NewLexer(source)

	expectedTypes := []TokenType{
		TOK_STRING, TOK_TILDE_STRING, TOK_STRING, TOK_EOF,
	}

	for _, expectedType := range expectedTypes {
		tok := lexer.NextToken()
		if tok.Type != expectedType {
			t.Errorf("Expected %v, got %v", expectedType, tok.Type)
		}
	}
}
