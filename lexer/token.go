package lexer

// TokenType identifies the kind of a token produced by the lexer.
type TokenType string

const (
	// Special
	TOKEN_EOF     TokenType = "EOF"
	TOKEN_ILLEGAL TokenType = "ILLEGAL"

	// Literals
	TOKEN_IDENT  TokenType = "IDENT"
	TOKEN_NUMBER TokenType = "NUMBER"
	TOKEN_STRING TokenType = "STRING"

	// Keywords
	TOKEN_VAR      TokenType = "VAR"
	TOKEN_CONST    TokenType = "CONST"
	TOKEN_IF       TokenType = "IF"
	TOKEN_ELSE     TokenType = "ELSE"
	TOKEN_WHILE    TokenType = "WHILE"
	TOKEN_FOR      TokenType = "FOR"
	TOKEN_IN       TokenType = "IN"
	TOKEN_RANGE    TokenType = "RANGE"
	TOKEN_FUNC     TokenType = "FUNC"
	TOKEN_RETURN   TokenType = "RETURN"
	TOKEN_BREAK    TokenType = "BREAK"
	TOKEN_CONTINUE TokenType = "CONTINUE"
	TOKEN_TRUE     TokenType = "TRUE"
	TOKEN_FALSE    TokenType = "FALSE"
	TOKEN_NULL     TokenType = "NULL"
	TOKEN_AND      TokenType = "AND"
	TOKEN_OR       TokenType = "OR"

	// Single-char punctuation
	TOKEN_LPAREN    TokenType = "("
	TOKEN_RPAREN    TokenType = ")"
	TOKEN_LBRACE    TokenType = "{"
	TOKEN_RBRACE    TokenType = "}"
	TOKEN_LBRACKET  TokenType = "["
	TOKEN_RBRACKET  TokenType = "]"
	TOKEN_COMMA     TokenType = ","
	TOKEN_SEMICOLON TokenType = ";"

	// Operators
	TOKEN_ASSIGN     TokenType = "="
	TOKEN_PLUS       TokenType = "+"
	TOKEN_MINUS      TokenType = "-"
	TOKEN_STAR       TokenType = "*"
	TOKEN_SLASH      TokenType = "/"
	TOKEN_FLOORDIV   TokenType = "//"
	TOKEN_PERCENT    TokenType = "%"
	TOKEN_POWER      TokenType = "**"
	TOKEN_PLUSEQ     TokenType = "+="
	TOKEN_MINUSEQ    TokenType = "-="
	TOKEN_STAREQ     TokenType = "*="
	TOKEN_SLASHEQ    TokenType = "/="
	TOKEN_EQ         TokenType = "=="
	TOKEN_NEQ        TokenType = "!="
	TOKEN_LT         TokenType = "<"
	TOKEN_GT         TokenType = ">"
	TOKEN_LTE        TokenType = "<="
	TOKEN_GTE        TokenType = ">="
	TOKEN_BANG       TokenType = "!"
)

// Token is a single lexical unit with its source location.
type Token struct {
	Type    TokenType
	Lexeme  string
	Line    int
	Column  int
}

// keywords maps reserved identifiers to their token type.
var keywords = map[string]TokenType{
	"var":      TOKEN_VAR,
	"const":    TOKEN_CONST,
	"if":       TOKEN_IF,
	"else":     TOKEN_ELSE,
	"while":    TOKEN_WHILE,
	"for":      TOKEN_FOR,
	"in":       TOKEN_IN,
	"range":    TOKEN_RANGE,
	"func":     TOKEN_FUNC,
	"return":   TOKEN_RETURN,
	"break":    TOKEN_BREAK,
	"continue": TOKEN_CONTINUE,
	"true":     TOKEN_TRUE,
	"false":    TOKEN_FALSE,
	"null":     TOKEN_NULL,
	"and":      TOKEN_AND,
	"or":       TOKEN_OR,
}

// LookupIdent returns the keyword token type for ident, or TOKEN_IDENT.
func LookupIdent(ident string) TokenType {
	if tt, ok := keywords[ident]; ok {
		return tt
	}
	return TOKEN_IDENT
}
