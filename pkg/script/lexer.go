// lexer.go - Duso language tokenizer
//
// This file implements the lexer (scanner/tokenizer) that converts source code strings
// into a stream of tokens. It is the first stage of compilation, before parsing.
//
// CORE LANGUAGE COMPONENT: This is part of the minimal core language.
// It is required for all script execution, both in embedded applications and the CLI.
//
// The lexer handles:
// - Character-by-character reading from source code
// - Token identification (keywords, operators, literals, identifiers)
// - Line and column tracking for error reporting
// - String and number literal parsing
// - Comment handling
package script

import (
	"strings"
	"unicode"
)

type Lexer struct {
	source string
	pos    int
	line   int
	column int
	ch     rune
}

func NewLexer(source string) *Lexer {
	l := &Lexer{
		source: source,
		pos:    0,
		line:   1,
		column: 0,
	}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.pos >= len(l.source) {
		l.ch = 0
	} else {
		l.ch = rune(l.source[l.pos])
	}
	l.pos++
	l.column++
}

func (l *Lexer) peekChar() rune {
	if l.pos >= len(l.source) {
		return 0
	}
	return rune(l.source[l.pos])
}

func (l *Lexer) peekChar2() rune {
	if l.pos+1 >= len(l.source) {
		return 0
	}
	return rune(l.source[l.pos+1])
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' || l.ch == '\n' {
		if l.ch == '\n' {
			l.line++
			l.column = 0
		}
		l.readChar()
	}
}

func (l *Lexer) skipComment() {
	// Skip "//" comment until end of line
	if l.ch == '/' && l.peekChar() == '/' {
		for l.ch != '\n' && l.ch != 0 {
			l.readChar()
		}
		if l.ch == '\n' {
			l.line++
			l.column = 0
			l.readChar()
		}
	}
}

func (l *Lexer) skipNestedComment() {
	// Skip "/* ... */" comment with nesting support
	if l.ch == '/' && l.peekChar() == '*' {
		depth := 1
		l.readChar() // skip '/'
		l.readChar() // skip '*'

		for depth > 0 && l.ch != 0 {
			if l.ch == '/' && l.peekChar() == '*' {
				// Found another opening
				depth++
				l.readChar()
				l.readChar()
			} else if l.ch == '*' && l.peekChar() == '/' {
				// Found a closing
				depth--
				l.readChar()
				l.readChar()
			} else {
				// Track newlines
				if l.ch == '\n' {
					l.line++
					l.column = 0
				}
				l.readChar()
			}
		}
	}
}

func (l *Lexer) readString(quote rune) string {
	start := l.pos - 1  // Position of the opening quote
	l.readChar() // Skip opening quote, move to first character

	for l.ch != quote && l.ch != 0 {
		if l.ch == '\\' {
			l.readChar() // Skip escape char
			if l.ch != 0 {
				l.readChar()
			}
		} else {
			// Track newlines for accurate line counting
			if l.ch == '\n' {
				l.line++
				l.column = 0
			}
			l.readChar()
		}
	}

	// Extract the raw string content (without quotes)
	// l.pos-1 because l.pos is now pointing past the closing quote
	result := l.source[start+1 : l.pos-1]
	if l.ch == quote {
		l.readChar() // Skip closing quote
	}

	// Return raw string WITHOUT unescaping - let parser handle that
	// This allows us to detect templates before unescaping
	return result
}

func (l *Lexer) readRawString() string {
	// ~string~ - like raw"string", no unescaping, no templates
	start := l.pos - 1  // Position of the opening ~
	l.readChar() // Skip opening ~, move to first character

	for l.ch != 0 {
		if l.ch == '\\' && l.peekChar() == '~' {
			// Escaped tilde - skip both characters, not a delimiter
			l.readChar() // Skip backslash
			l.readChar() // Skip the tilde
		} else if l.ch == '~' {
			// End delimiter found
			break
		} else {
			// Track newlines for accurate line counting
			if l.ch == '\n' {
				l.line++
				l.column = 0
			}
			l.readChar()
		}
	}

	// Extract the raw string content (without tildes)
	// After the last loop iteration, we've called readChar() which read the closing ~
	// and incremented l.pos past it, so we need l.pos-1 to get the position of the ~
	result := l.source[start+1 : l.pos-1]
	if l.ch == '~' {
		l.readChar() // Skip closing ~
	}

	// Return raw content - preserve backslashes
	return result
}

func (l *Lexer) readMultilineString(quote rune) string {
	// Skip first quote (already in l.ch)
	l.readChar() // Now reading 2nd quote, l.pos points to 3rd quote
	l.readChar() // Now reading 3rd quote, l.pos points to first content char
	l.readChar() // Now reading first content character

	contentStart := l.pos - 1

	// Find closing triple quotes
	for {
		if l.ch == 0 {
			break // EOF without closing
		}

		if l.ch == quote && l.peekChar() == quote && l.peekChar2() == quote {
			// Found closing triple quotes
			break
		}

		// Track newlines
		if l.ch == '\n' {
			l.line++
			l.column = 0
		}

		l.readChar()
	}

	// Extract string content (from start to closing 3 quotes)
	// l.pos is pointing to 2nd closing quote (because l.pos is always one ahead of l.ch)
	result := l.source[contentStart : l.pos-1]

	// Skip closing triple quotes
	if l.ch == quote {
		l.readChar()
		l.readChar()
		l.readChar()
	}

	// Strip leading/trailing whitespace including newlines
	result = strings.TrimSpace(result)

	// Remove common leading whitespace from all lines (dedent)
	result = dedentString(result)

	return result
}

// dedentString removes common leading whitespace from all lines while preserving relative indentation
// This allows developers to write code naturally indented in their editor without pollution
// Example:
//   """
//     hello
//       world
//   """
// Becomes: "hello\n  world" (the common 4 spaces are removed, but relative indentation is preserved)
func dedentString(s string) string {
	lines := strings.Split(s, "\n")
	if len(lines) == 0 {
		return s
	}

	// Find the common leading whitespace (must be identical across all non-empty lines)
	var commonPrefix string
	for _, line := range lines {
		if len(strings.TrimSpace(line)) == 0 {
			// Skip empty lines
			continue
		}

		// Get leading whitespace of this line
		leadingWS := ""
		for _, ch := range line {
			if ch == ' ' || ch == '\t' {
				leadingWS += string(ch)
			} else {
				break
			}
		}

		if commonPrefix == "" {
			commonPrefix = leadingWS
		} else {
			// Find common prefix between current commonPrefix and this line's leading whitespace
			commonPrefix = findCommonPrefix(commonPrefix, leadingWS)
			if commonPrefix == "" {
				// No common prefix, stop searching
				break
			}
		}
	}

	// Remove the common prefix from each line
	if commonPrefix != "" {
		var result []string
		for _, line := range lines {
			if strings.HasPrefix(line, commonPrefix) {
				result = append(result, line[len(commonPrefix):])
			} else if len(strings.TrimSpace(line)) == 0 {
				// Keep empty lines as-is
				result = append(result, line)
			} else {
				// This shouldn't happen if commonPrefix was calculated correctly, but keep line as-is
				result = append(result, line)
			}
		}
		return strings.Join(result, "\n")
	}

	return s
}

// findCommonPrefix finds the longest common prefix between two strings
// Only considers whitespace characters (spaces and tabs)
func findCommonPrefix(a, b string) string {
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}

	for i := 0; i < minLen; i++ {
		if a[i] != b[i] {
			return a[:i]
		}
	}

	return a[:minLen]
}

func (l *Lexer) readNumber() string {
	start := l.pos - 1

	for unicode.IsDigit(l.ch) {
		l.readChar()
	}

	// Check for decimal point
	if l.ch == '.' && unicode.IsDigit(l.peekChar()) {
		l.readChar() // Skip '.'
		for unicode.IsDigit(l.ch) {
			l.readChar()
		}
	}

	return l.source[start : l.pos-1]
}

func (l *Lexer) readIdent() string {
	start := l.pos - 1

	for unicode.IsLetter(l.ch) || unicode.IsDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}

	return l.source[start : l.pos-1]
}

func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	// Skip comments and whitespace
	for {
		if l.ch == '/' && l.peekChar() == '*' {
			l.skipNestedComment()
			l.skipWhitespace()
		} else if l.ch == '/' && l.peekChar() == '/' {
			l.skipComment()
			l.skipWhitespace()
		} else {
			break
		}
	}

	line := l.line
	column := l.column

	switch l.ch {
	case 0:
		return Token{Type: TOK_EOF, Line: line, Column: column}
	case '+':
		l.readChar()
		if l.ch == '=' {
			l.readChar()
			return Token{Type: TOK_PLUSASSIGN, Value: "+=", Line: line, Column: column}
		}
		if l.ch == '+' {
			l.readChar()
			return Token{Type: TOK_INCREMENT, Value: "++", Line: line, Column: column}
		}
		return Token{Type: TOK_PLUS, Value: "+", Line: line, Column: column}
	case '-':
		l.readChar()
		if l.ch == '=' {
			l.readChar()
			return Token{Type: TOK_MINUSASSIGN, Value: "-=", Line: line, Column: column}
		}
		if l.ch == '-' {
			l.readChar()
			return Token{Type: TOK_DECREMENT, Value: "--", Line: line, Column: column}
		}
		return Token{Type: TOK_MINUS, Value: "-", Line: line, Column: column}
	case '*':
		l.readChar()
		if l.ch == '=' {
			l.readChar()
			return Token{Type: TOK_STARASSIGN, Value: "*=", Line: line, Column: column}
		}
		return Token{Type: TOK_STAR, Value: "*", Line: line, Column: column}
	case '/':
		l.readChar()
		if l.ch == '=' {
			l.readChar()
			return Token{Type: TOK_SLASHASSIGN, Value: "/=", Line: line, Column: column}
		}
		return Token{Type: TOK_SLASH, Value: "/", Line: line, Column: column}
	case '%':
		l.readChar()
		if l.ch == '=' {
			l.readChar()
			return Token{Type: TOK_MODASSIGN, Value: "%=", Line: line, Column: column}
		}
		return Token{Type: TOK_PERCENT, Value: "%", Line: line, Column: column}
	case '=':
		l.readChar()
		if l.ch == '=' {
			l.readChar()
			return Token{Type: TOK_EQUAL, Value: "==", Line: line, Column: column}
		}
		return Token{Type: TOK_ASSIGN, Value: "=", Line: line, Column: column}
	case '!':
		l.readChar()
		if l.ch == '=' {
			l.readChar()
			return Token{Type: TOK_NOTEQUAL, Value: "!=", Line: line, Column: column}
		}
		return Token{Type: TOK_NOT, Value: "not", Line: line, Column: column}
	case '<':
		l.readChar()
		if l.ch == '=' {
			l.readChar()
			return Token{Type: TOK_LTE, Value: "<=", Line: line, Column: column}
		}
		return Token{Type: TOK_LT, Value: "<", Line: line, Column: column}
	case '>':
		l.readChar()
		if l.ch == '=' {
			l.readChar()
			return Token{Type: TOK_GTE, Value: ">=", Line: line, Column: column}
		}
		return Token{Type: TOK_GT, Value: ">", Line: line, Column: column}
	case '(':
		l.readChar()
		return Token{Type: TOK_LPAREN, Value: "(", Line: line, Column: column}
	case ')':
		l.readChar()
		return Token{Type: TOK_RPAREN, Value: ")", Line: line, Column: column}
	case '[':
		l.readChar()
		return Token{Type: TOK_LBRACKET, Value: "[", Line: line, Column: column}
	case ']':
		l.readChar()
		return Token{Type: TOK_RBRACKET, Value: "]", Line: line, Column: column}
	case '{':
		l.readChar()
		return Token{Type: TOK_LBRACE, Value: "{", Line: line, Column: column}
	case '}':
		l.readChar()
		return Token{Type: TOK_RBRACE, Value: "}", Line: line, Column: column}
	case ',':
		l.readChar()
		return Token{Type: TOK_COMMA, Value: ",", Line: line, Column: column}
	case '.':
		// Check if this is a float literal starting with . (e.g., .5)
		if unicode.IsDigit(l.peekChar()) {
			start := l.pos - 1 // Capture position of '.'
			l.readChar()       // Move to first digit
			// Read remaining digits
			for unicode.IsDigit(l.ch) {
				l.readChar()
			}
			// Prepend "0" to make it a valid float (e.g., ".5" -> "0.5")
			value := "0" + l.source[start : l.pos-1]
			return Token{Type: TOK_NUMBER, Value: value, Line: line, Column: column}
		}
		l.readChar()
		return Token{Type: TOK_DOT, Value: ".", Line: line, Column: column}
	case ':':
		l.readChar()
		return Token{Type: TOK_COLON, Value: ":", Line: line, Column: column}
	case '?':
		l.readChar()
		return Token{Type: TOK_QUESTION, Value: "?", Line: line, Column: column}
	case '"', '\'':
		quote := l.ch
		// Check for triple quotes
		if l.peekChar() == quote && l.peekChar2() == quote {
			value := l.readMultilineString(quote)
			return Token{Type: TOK_STRING, Value: value, Line: line, Column: column}
		}
		value := l.readString(quote)
		return Token{Type: TOK_STRING, Value: value, Line: line, Column: column}
	case '~':
		value := l.readRawString()
		return Token{Type: TOK_TILDE_STRING, Value: value, Line: line, Column: column}
	default:
		if unicode.IsLetter(l.ch) || l.ch == '_' {
			value := l.readIdent()
			typ := LookupKeyword(value)
			return Token{Type: typ, Value: value, Line: line, Column: column}
		}
		if unicode.IsDigit(l.ch) {
			value := l.readNumber()
			return Token{Type: TOK_NUMBER, Value: value, Line: line, Column: column}
		}
		l.readChar()
		return Token{Type: TOK_EOF, Value: "", Line: line, Column: column}
	}
}

func (l *Lexer) Tokenize() []Token {
	var tokens []Token
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == TOK_EOF {
			break
		}
	}
	return tokens
}

// UnescapeString processes escape sequences in a string, preserving UTF-8
// Uses rune-based iteration to handle multi-byte characters correctly
func UnescapeString(s string) string {
	result := ""
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		if runes[i] == '\\' && i+1 < len(runes) {
			i++
			switch runes[i] {
			case 'n':
				result += "\n"
			case 't':
				result += "\t"
			case 'r':
				result += "\r"
			case '"':
				result += "\""
			case '\'':
				result += "'"
			case '\\':
				result += "\\"
			case '{':
				result += "{"
			case '}':
				result += "}"
			case '0', '1', '2', '3', '4', '5', '6', '7':
				// Octal escape sequence: \ddd
				octal := string(runes[i])
				if i+1 < len(runes) && runes[i+1] >= '0' && runes[i+1] <= '7' {
					i++
					octal += string(runes[i])
					if i+1 < len(runes) && runes[i+1] >= '0' && runes[i+1] <= '7' {
						i++
						octal += string(runes[i])
					}
				}
				// Parse octal string to byte
				var val byte
				for _, ch := range octal {
					val = val*8 + byte(ch-'0')
				}
				result += string(val)
			case 'x':
				// Hex escape sequence: \xhh
				if i+2 < len(runes) {
					hex := string(runes[i+1 : i+3])
					var val byte
					for _, ch := range hex {
						val = val * 16
						if ch >= '0' && ch <= '9' {
							val += byte(ch - '0')
						} else if ch >= 'a' && ch <= 'f' {
							val += byte(ch - 'a' + 10)
						} else if ch >= 'A' && ch <= 'F' {
							val += byte(ch - 'A' + 10)
						}
					}
					result += string(val)
					i += 2
				} else {
					result += string(runes[i])
				}
			default:
				result += string(runes[i])
			}
		} else {
			result += string(runes[i])
		}
	}
	return result
}
