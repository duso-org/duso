package script

import (
	"testing"
)

// TestLexerInitialization tests lexer creation and initial state
func TestLexerInitialization(t *testing.T) {
	tests := []struct {
		name   string
		source string
		expCh  rune
		expPos int
		expLn  int
		expCol int
	}{
		{"empty", "", 0, 1, 1, 1},
		{"single char", "a", 'a', 1, 1, 1},
		{"newline", "\n", '\n', 1, 1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lex := NewLexer(tt.source)
			if lex.ch != tt.expCh {
				t.Errorf("ch = %q, want %q", lex.ch, tt.expCh)
			}
			if lex.pos != tt.expPos {
				t.Errorf("pos = %d, want %d", lex.pos, tt.expPos)
			}
			if lex.line != tt.expLn {
				t.Errorf("line = %d, want %d", lex.line, tt.expLn)
			}
			if lex.column != tt.expCol {
				t.Errorf("column = %d, want %d", lex.column, tt.expCol)
			}
		})
	}
}

// TestReadCharAdvancement tests readChar() position tracking
func TestReadCharAdvancement(t *testing.T) {
	lex := NewLexer("abc")

	// Initial: pos=1 (after first readChar in NewLexer), ch='a'
	if lex.ch != 'a' || lex.pos != 1 {
		t.Fatalf("initial: ch=%q pos=%d", lex.ch, lex.pos)
	}

	lex.readChar() // pos becomes 2, ch becomes 'b'
	if lex.ch != 'b' || lex.pos != 2 {
		t.Errorf("after 1st: ch=%q pos=%d, want ch='b' pos=2", lex.ch, lex.pos)
	}

	lex.readChar() // pos becomes 3, ch becomes 'c'
	if lex.ch != 'c' || lex.pos != 3 {
		t.Errorf("after 2nd: ch=%q pos=%d, want ch='c' pos=3", lex.ch, lex.pos)
	}

	lex.readChar() // pos becomes 4, ch becomes 0 (EOF)
	if lex.ch != 0 || lex.pos != 4 {
		t.Errorf("after 3rd (EOF): ch=%q pos=%d, want ch=0 pos=4", lex.ch, lex.pos)
	}
}

// TestPeekCharWithoutAdvancing tests peekChar() doesn't modify position
func TestPeekCharWithoutAdvancing(t *testing.T) {
	lex := NewLexer("abc")

	// Peek ahead multiple times
	for i := 0; i < 5; i++ {
		p := lex.peekChar()
		if p != 'b' {
			t.Errorf("peek %d: got %q, want 'b'", i, p)
		}
		if lex.ch != 'a' || lex.pos != 1 {
			t.Errorf("peek %d: position changed to ch=%q pos=%d", i, lex.ch, lex.pos)
		}
	}

	lex.readChar() // Now at 'b'
	p := lex.peekChar()
	if p != 'c' {
		t.Errorf("after readChar: peekChar got %q, want 'c'", p)
	}
}

// TestPeekChar2 tests peekChar2() looks two ahead
func TestPeekChar2(t *testing.T) {
	tests := []struct{
		name    string
		readN   int
		expPk2  rune
		expPk1  rune
		expCurr rune
	}{
		{"at a", 0, 'c', 'b', 'a'},
		{"at b", 1, 'd', 'c', 'b'},
		{"at c", 2, 'e', 'd', 'c'},
		{"at d", 3, 0, 'e', 'd'},
		{"at e", 4, 0, 0, 'e'},
		{"at EOF", 5, 0, 0, 0},
	}

	for _, tt := range tests {
		lex := NewLexer("abcde")
		for i := 0; i < tt.readN; i++ {
			lex.readChar()
		}

		t.Run(tt.name, func(t *testing.T) {
			pk2 := lex.peekChar2()
			pk1 := lex.peekChar()
			if pk2 != tt.expPk2 {
				t.Errorf("peekChar2 = %q, want %q", pk2, tt.expPk2)
			}
			if pk1 != tt.expPk1 {
				t.Errorf("peekChar = %q, want %q", pk1, tt.expPk1)
			}
			if lex.ch != tt.expCurr {
				t.Errorf("current = %q, want %q", lex.ch, tt.expCurr)
			}
		})
	}
}

// TestNumberParsing tests readNumber() with all formats
func TestNumberParsing(t *testing.T) {
	tests := []struct {
		name   string
		source string
		expNum string
		expPos int // position after parsing (one past last char read)
	}{
		{"integer", "123", "123", 4},      // pos goes 1->2->3->4
		{"float", "3.14", "3.14", 5},      // pos goes 1->2->3->4->5
		{"leading decimal", ".5", ".5", 3}, // pos goes 1->2->3
		{"scientific 1e10", "1e10", "1e10", 5},    // pos goes 1->2->3->4->5
		{"scientific 1.5e-3", "1.5e-3", "1.5e-3", 7}, // pos goes 1->2->3->4->5->6->7
		{"scientific 2E+5", "2E+5", "2E+5", 5},   // pos goes 1->2->3->4->5
		{"zero", "0", "0", 2},              // pos goes 1->2
		{"large int", "9999999999", "9999999999", 11}, // pos goes 1->2...->11
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lex := NewLexer(tt.source)
			num := lex.readNumber()
			if num != tt.expNum {
				t.Errorf("got %q, want %q", num, tt.expNum)
			}
			if lex.pos != tt.expPos {
				t.Errorf("pos = %d, want %d", lex.pos, tt.expPos)
			}
		})
	}
}

// TestStringParsing tests readString() with escapes and line tracking
func TestStringParsing(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		quote   rune
		expStr  string
		expLine int
	}{
		{
			name:    "simple double quote",
			source:  `"hello"`,
			quote:   '"',
			expStr:  "hello",
			expLine: 1,
		},
		{
			name:    "empty string",
			source:  `""`,
			quote:   '"',
			expStr:  "",
			expLine: 1,
		},
		{
			name:    "string with escape",
			source:  `"hello\nworld"`,
			quote:   '"',
			expStr:  "hello\\nworld", // Raw, unescaped by lexer
			expLine: 1,
		},
		{
			name:    "string with newline",
			source:  "\"hello\nworld\"",
			quote:   '"',
			expStr:  "hello\nworld",
			expLine: 2, // Line count should increase
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lex := NewLexer(tt.source)
			str := lex.readString(tt.quote)
			if str != tt.expStr {
				t.Errorf("got %q, want %q", str, tt.expStr)
			}
			if lex.line != tt.expLine {
				t.Errorf("line = %d, want %d", lex.line, tt.expLine)
			}
		})
	}
}

// TestWhitespaceSkipping tests skipWhitespace() and line tracking
func TestWhitespaceSkipping(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		expCh   rune // char after whitespace
		expLine int
		expCol  int
	}{
		{"spaces", "   a", 'a', 1, 4},
		{"tabs", "\t\ta", 'a', 1, 3},
		{"newline", "\na", 'a', 2, 1},
		{"mixed", "  \t\na", 'a', 2, 1},
		{"semicolon", ";;a", 'a', 1, 3},
		{"no whitespace", "a", 'a', 1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lex := NewLexer(tt.source)
			lex.skipWhitespace()
			if lex.ch != tt.expCh {
				t.Errorf("ch = %q, want %q", lex.ch, tt.expCh)
			}
			if lex.line != tt.expLine {
				t.Errorf("line = %d, want %d", lex.line, tt.expLine)
			}
			if lex.column != tt.expCol {
				t.Errorf("column = %d, want %d", lex.column, tt.expCol)
			}
		})
	}
}

// TestLineCommentSkipping tests skipComment() for line comments
func TestLineCommentSkipping(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		expCh   rune
		expLine int
	}{
		{"// comment\\na", "// comment\na", 'a', 2},
		{"// at EOF", "// comment", 0, 1},
		{"no comment", "a", 'a', 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lex := NewLexer(tt.source)
			lex.skipComment()
			if lex.ch != tt.expCh {
				t.Errorf("ch = %q, want %q", lex.ch, tt.expCh)
			}
			if lex.line != tt.expLine {
				t.Errorf("line = %d, want %d", lex.line, tt.expLine)
			}
		})
	}
}

// TestNestedCommentSkipping tests skipNestedComment() for /* */ comments
func TestNestedCommentSkipping(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		expCh   rune
		expLine int
	}{
		{"simple nested", "/* comment */a", 'a', 1},
		{"nested with newline", "/* line1\nline2 */a", 'a', 2},
		{"nested nesting", "/* outer /* inner */ outer */a", 'a', 1},
		{"no comment", "a", 'a', 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lex := NewLexer(tt.source)
			lex.skipNestedComment()
			if lex.ch != tt.expCh {
				t.Errorf("ch = %q, want %q", lex.ch, tt.expCh)
			}
			if lex.line != tt.expLine {
				t.Errorf("line = %d, want %d", lex.line, tt.expLine)
			}
		})
	}
}

// TestNextTokenSimple tests NextToken() for basic tokens
func TestNextTokenSimple(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expType  TokenType
		expValue string
	}{
		{"if keyword", "if", TOK_IF, "if"},
		{"number", "42", TOK_NUMBER, "42"},
		{"string", `"hello"`, TOK_STRING, "hello"},
		{"identifier", "myVar", TOK_IDENT, "myVar"},
		{"plus", "+", TOK_PLUS, "+"},
		{"equals", "==", TOK_EQUAL, "=="},
		{"EOF", "", TOK_EOF, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lex := NewLexer(tt.source)
			tok := lex.NextToken()
			if tok.Type != tt.expType {
				t.Errorf("type = %v, want %v", tok.Type, tt.expType)
			}
			if tok.Value != tt.expValue {
				t.Errorf("value = %q, want %q", tok.Value, tt.expValue)
			}
		})
	}
}

// TestNextTokenLineTracking tests line and column tracking in tokens
func TestNextTokenLineTracking(t *testing.T) {
	source := "a\nb"
	lex := NewLexer(source)

	tok1 := lex.NextToken()
	if tok1.Line != 1 {
		t.Errorf("first token line = %d, want 1", tok1.Line)
	}

	tok2 := lex.NextToken()
	if tok2.Line != 2 {
		t.Errorf("second token line = %d, want 2", tok2.Line)
	}
}

// TestAllKeywords tests all keyword tokens are recognized
func TestAllKeywords(t *testing.T) {
	keywords := map[string]TokenType{
		"if":       TOK_IF,
		"else":     TOK_ELSE,
		"for":      TOK_FOR,
		"while":    TOK_WHILE,
		"function": TOK_FUNCTION,
		"return":   TOK_RETURN,
		"true":     TOK_TRUE,
		"false":    TOK_FALSE,
		"nil":      TOK_NIL,
		"and":      TOK_AND,
		"or":       TOK_OR,
		"not":      TOK_NOT,
		"try":      TOK_TRY,
		"catch":    TOK_CATCH,
		"end":      TOK_END,
		"then":     TOK_THEN,
		"var":      TOK_VAR,
		"do":       TOK_DO,
		"in":       TOK_IN,
	}

	for kw, expType := range keywords {
		t.Run(kw, func(t *testing.T) {
			lex := NewLexer(kw)
			tok := lex.NextToken()
			if tok.Type != expType {
				t.Errorf("got %v, want %v", tok.Type, expType)
			}
			if tok.Value != kw {
				t.Errorf("value = %q, want %q", tok.Value, kw)
			}
		})
	}
}

// TestAllOperators tests all operator tokens
func TestAllOperators(t *testing.T) {
	operators := map[string]TokenType{
		"+":  TOK_PLUS,
		"-":  TOK_MINUS,
		"*":  TOK_STAR,
		"/":  TOK_SLASH,
		"%":  TOK_PERCENT,
		"=":  TOK_ASSIGN,
		"==": TOK_EQUAL,
		"!=": TOK_NOTEQUAL,
		"<":  TOK_LT,
		"<=": TOK_LTE,
		">":  TOK_GT,
		">=": TOK_GTE,
		"++": TOK_INCREMENT,
		"--": TOK_DECREMENT,
		"(":  TOK_LPAREN,
		")":  TOK_RPAREN,
		"[":  TOK_LBRACKET,
		"]":  TOK_RBRACKET,
		"{":  TOK_LBRACE,
		"}":  TOK_RBRACE,
		",":  TOK_COMMA,
		".":  TOK_DOT,
		":":  TOK_COLON,
	}

	for op, expType := range operators {
		t.Run(op, func(t *testing.T) {
			lex := NewLexer(op)
			tok := lex.NextToken()
			if tok.Type != expType {
				t.Errorf("got %v, want %v", tok.Type, expType)
			}
		})
	}
}

// TestComplexExpression tests tokenization of a complex expression
func TestComplexExpression(t *testing.T) {
	source := "if x > 5 then print(x) end"
	lex := NewLexer(source)

	expected := []TokenType{
		TOK_IF, TOK_IDENT, TOK_GT, TOK_NUMBER, TOK_THEN,
		TOK_IDENT, TOK_LPAREN, TOK_IDENT, TOK_RPAREN, TOK_END,
		TOK_EOF,
	}

	for i, expType := range expected {
		tok := lex.NextToken()
		if tok.Type != expType {
			t.Errorf("token %d: got %v, want %v", i, tok.Type, expType)
		}
	}
}

// TestEOFHandling tests EOF at various positions
func TestEOFHandling(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"empty", ""},
		{"single char EOF", "a"},
		{"after whitespace", "  "},
		{"after comment", "// comment"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lex := NewLexer(tt.source)
			for {
				tok := lex.NextToken()
				if tok.Type == TOK_EOF {
					break
				}
			}

			// Multiple EOF calls should stay at EOF
			tok := lex.NextToken()
			if tok.Type != TOK_EOF {
				t.Errorf("got %v, want TOK_EOF", tok.Type)
			}
		})
	}
}

// TestIdentifierEdgeCases tests identifier parsing edge cases
func TestIdentifierEdgeCases(t *testing.T) {
	tests := []struct {
		name string
		id   string
	}{
		{"underscore prefix", "_private"},
		{"camelCase", "myVarName"},
		{"CONSTANT", "CONSTANT"},
		{"with digits", "var123"},
		{"ending in underscore", "var_"},
		{"multiple underscores", "__init__"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lex := NewLexer(tt.id)
			tok := lex.NextToken()
			if tok.Type != TOK_IDENT {
				t.Errorf("got %v, want TOK_IDENT", tok.Type)
			}
			if tok.Value != tt.id {
				t.Errorf("got %q, want %q", tok.Value, tt.id)
			}
		})
	}
}

// TestCommentInteraction tests comments in various positions
func TestCommentInteraction(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		expToks []TokenType
	}{
		{
			name:    "comment before token",
			source:  "// comment\na",
			expToks: []TokenType{TOK_IDENT, TOK_EOF},
		},
		{
			name:    "comment after token",
			source:  "a // comment",
			expToks: []TokenType{TOK_IDENT, TOK_EOF},
		},
		{
			name:    "nested comment",
			source:  "a /* nested /* comment */ */ b",
			expToks: []TokenType{TOK_IDENT, TOK_IDENT, TOK_EOF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lex := NewLexer(tt.source)
			for i, expType := range tt.expToks {
				tok := lex.NextToken()
				if tok.Type != expType {
					t.Errorf("token %d: got %v, want %v", i, tok.Type, expType)
				}
			}
		})
	}
}
