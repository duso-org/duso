package script

import "fmt"

type TokenType int

const (
	// Special
	TOK_EOF TokenType = iota
	TOK_COMMENT

	// Literals
	TOK_NUMBER
	TOK_STRING
	TOK_TILDE_STRING
	TOK_TRUE
	TOK_FALSE
	TOK_NIL
	TOK_IDENT

	// Keywords
	TOK_IF
	TOK_THEN
	TOK_ELSE
	TOK_ELSEIF
	TOK_END
	TOK_WHILE
	TOK_DO
	TOK_FOR
	TOK_IN
	TOK_FUNCTION
	TOK_RETURN
	TOK_BREAK
	TOK_CONTINUE
	TOK_TRY
	TOK_CATCH
	TOK_AND
	TOK_OR
	TOK_NOT
	TOK_VAR
	TOK_RAW

	// Operators
	TOK_PLUS
	TOK_MINUS
	TOK_STAR
	TOK_SLASH
	TOK_PERCENT
	TOK_EQUAL
	TOK_NOTEQUAL
	TOK_LT
	TOK_GT
	TOK_LTE
	TOK_GTE
	TOK_ASSIGN
	TOK_PLUSASSIGN
	TOK_MINUSASSIGN
	TOK_STARASSIGN
	TOK_SLASHASSIGN
	TOK_MODASSIGN
	TOK_INCREMENT
	TOK_DECREMENT

	// Delimiters
	TOK_LPAREN
	TOK_RPAREN
	TOK_LBRACKET
	TOK_RBRACKET
	TOK_LBRACE
	TOK_RBRACE
	TOK_COMMA
	TOK_DOT
	TOK_COLON
	TOK_QUESTION
)

var tokenNames = map[TokenType]string{
	TOK_EOF:          "EOF",
	TOK_COMMENT:      "COMMENT",
	TOK_NUMBER:       "NUMBER",
	TOK_STRING:       "STRING",
	TOK_TILDE_STRING: "TILDE_STRING",
	TOK_TRUE:         "TRUE",
	TOK_FALSE:     "FALSE",
	TOK_NIL:       "NIL",
	TOK_IDENT:     "IDENT",
	TOK_IF:        "IF",
	TOK_THEN:      "THEN",
	TOK_ELSE:      "ELSE",
	TOK_ELSEIF:    "ELSEIF",
	TOK_END:       "END",
	TOK_WHILE:     "WHILE",
	TOK_DO:        "DO",
	TOK_FOR:       "FOR",
	TOK_IN:        "IN",
	TOK_FUNCTION:  "FUNCTION",
	TOK_RETURN:    "RETURN",
	TOK_BREAK:     "BREAK",
	TOK_CONTINUE:  "CONTINUE",
	TOK_TRY:       "TRY",
	TOK_CATCH:     "CATCH",
	TOK_AND:       "AND",
	TOK_OR:        "OR",
	TOK_NOT:       "NOT",
	TOK_VAR:       "VAR",
	TOK_RAW:       "RAW",
	TOK_PLUS:      "+",
	TOK_MINUS:     "-",
	TOK_STAR:      "*",
	TOK_SLASH:     "/",
	TOK_PERCENT:   "%",
	TOK_EQUAL:     "==",
	TOK_NOTEQUAL:  "!=",
	TOK_LT:        "<",
	TOK_GT:        ">",
	TOK_LTE:       "<=",
	TOK_GTE:       ">=",
	TOK_ASSIGN:       "=",
	TOK_PLUSASSIGN:   "+=",
	TOK_MINUSASSIGN:  "-=",
	TOK_STARASSIGN:   "*=",
	TOK_SLASHASSIGN:  "/=",
	TOK_MODASSIGN:    "%=",
	TOK_INCREMENT:    "++",
	TOK_DECREMENT:    "--",
	TOK_LPAREN:       "(",
	TOK_RPAREN:    ")",
	TOK_LBRACKET:  "[",
	TOK_RBRACKET:  "]",
	TOK_LBRACE:    "{",
	TOK_RBRACE:    "}",
	TOK_COMMA:     ",",
	TOK_DOT:       ".",
	TOK_COLON:     ":",
	TOK_QUESTION:  "?",
}

// String returns a human-readable name for the TokenType
func (t TokenType) String() string {
	name, ok := tokenNames[t]
	if !ok {
		return fmt.Sprintf("UNKNOWN(%d)", t)
	}
	return name
}

type Token struct {
	Type   TokenType
	Value  string
	Line   int
	Column int
}

func (t Token) String() string {
	name, ok := tokenNames[t.Type]
	if !ok {
		name = fmt.Sprintf("UNKNOWN(%d)", t.Type)
	}
	if t.Value != "" && t.Type != TOK_EOF {
		return fmt.Sprintf("%s(%q) at %d:%d", name, t.Value, t.Line, t.Column)
	}
	return fmt.Sprintf("%s at %d:%d", name, t.Line, t.Column)
}

var keywords = map[string]TokenType{
	"if":        TOK_IF,
	"then":      TOK_THEN,
	"else":      TOK_ELSE,
	"elseif":    TOK_ELSEIF,
	"end":       TOK_END,
	"while":     TOK_WHILE,
	"do":        TOK_DO,
	"for":       TOK_FOR,
	"in":        TOK_IN,
	"function":  TOK_FUNCTION,
	"return":    TOK_RETURN,
	"break":     TOK_BREAK,
	"continue":  TOK_CONTINUE,
	"try":       TOK_TRY,
	"catch":     TOK_CATCH,
	"and":       TOK_AND,
	"or":        TOK_OR,
	"not":       TOK_NOT,
	"var":       TOK_VAR,
	"raw":       TOK_RAW,
	"true":      TOK_TRUE,
	"false":     TOK_FALSE,
	"nil":       TOK_NIL,
}

func LookupKeyword(ident string) TokenType {
	if typ, ok := keywords[ident]; ok {
		return typ
	}
	return TOK_IDENT
}
