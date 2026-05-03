package lexer

import (
	"fmt"
	"unicode"
)

// Lexer turns Evanescence source text into a stream of tokens.

type Lexer struct {
	source []rune
	pos    int
	line   int
	col    int
}

func New(source string) *Lexer {
	return &Lexer{source: []rune(source), pos: 0, line: 1, col: 1}
}

func (l *Lexer) Tokenize() ([]Token, error) {
	var tokens []Token
	for {
		tok, err := l.next()
		if err != nil {
			return tokens, err
		}
		tokens = append(tokens, tok)
		if tok.Type == TOKEN_EOF {
			return tokens, nil
		}
	}
}

func (l *Lexer) next() (Token, error) {
	l.skipWhitespaceAndComments()
	if l.atEnd() {
		return l.makeToken(TOKEN_EOF, ""), nil
	}

	startLine, startCol := l.line, l.col
	ch := l.peek()

	switch {
	case isLetter(ch):
		return l.identifier(startLine, startCol), nil
	case isDigit(ch):
		return l.number(startLine, startCol)
	case ch == '"':
		return l.stringLiteral(startLine, startCol)
	}

	// Operators and punctuation
	l.advance()
	switch ch {
	case '(':
		return Token{TOKEN_LPAREN, "(", startLine, startCol}, nil
	case ')':
		return Token{TOKEN_RPAREN, ")", startLine, startCol}, nil
	case '{':
		return Token{TOKEN_LBRACE, "{", startLine, startCol}, nil
	case '}':
		return Token{TOKEN_RBRACE, "}", startLine, startCol}, nil
	case '[':
		return Token{TOKEN_LBRACKET, "[", startLine, startCol}, nil
	case ']':
		return Token{TOKEN_RBRACKET, "]", startLine, startCol}, nil
	case ',':
		return Token{TOKEN_COMMA, ",", startLine, startCol}, nil
	case ';':
		return Token{TOKEN_SEMICOLON, ";", startLine, startCol}, nil
	case '%':
		return Token{TOKEN_PERCENT, "%", startLine, startCol}, nil
	case '+':
		if l.match('=') {
			return Token{TOKEN_PLUSEQ, "+=", startLine, startCol}, nil
		}
		return Token{TOKEN_PLUS, "+", startLine, startCol}, nil
	case '-':
		if l.match('=') {
			return Token{TOKEN_MINUSEQ, "-=", startLine, startCol}, nil
		}
		return Token{TOKEN_MINUS, "-", startLine, startCol}, nil
	case '*':
		if l.match('*') {
			return Token{TOKEN_POWER, "**", startLine, startCol}, nil
		}
		if l.match('=') {
			return Token{TOKEN_STAREQ, "*=", startLine, startCol}, nil
		}
		return Token{TOKEN_STAR, "*", startLine, startCol}, nil
	case '/':
		if l.match('/') {
			return Token{TOKEN_FLOORDIV, "//", startLine, startCol}, nil
		}
		if l.match('=') {
			return Token{TOKEN_SLASHEQ, "/=", startLine, startCol}, nil
		}
		return Token{TOKEN_SLASH, "/", startLine, startCol}, nil
	case '=':
		if l.match('=') {
			return Token{TOKEN_EQ, "==", startLine, startCol}, nil
		}
		return Token{TOKEN_ASSIGN, "=", startLine, startCol}, nil
	case '!':
		if l.match('=') {
			return Token{TOKEN_NEQ, "!=", startLine, startCol}, nil
		}
		return Token{TOKEN_BANG, "!", startLine, startCol}, nil
	case '<':
		if l.match('=') {
			return Token{TOKEN_LTE, "<=", startLine, startCol}, nil
		}
		return Token{TOKEN_LT, "<", startLine, startCol}, nil
	case '>':
		if l.match('=') {
			return Token{TOKEN_GTE, ">=", startLine, startCol}, nil
		}
		return Token{TOKEN_GT, ">", startLine, startCol}, nil
	}

	return Token{TOKEN_ILLEGAL, string(ch), startLine, startCol},
		fmt.Errorf("line %d:%d: unexpected character %q", startLine, startCol, ch)
}

func (l *Lexer) identifier(line, col int) Token {
	start := l.pos
	for !l.atEnd() && (isLetter(l.peek()) || isDigit(l.peek())) {
		l.advance()
	}
	lex := string(l.source[start:l.pos])
	return Token{LookupIdent(lex), lex, line, col}
}

func (l *Lexer) number(line, col int) (Token, error) {
	start := l.pos
	for !l.atEnd() && isDigit(l.peek()) {
		l.advance()
	}
	if !l.atEnd() && l.peek() == '.' && l.pos+1 < len(l.source) && isDigit(l.source[l.pos+1]) {
		l.advance() // '.'
		for !l.atEnd() && isDigit(l.peek()) {
			l.advance()
		}
	}
	return Token{TOKEN_NUMBER, string(l.source[start:l.pos]), line, col}, nil
}

func (l *Lexer) stringLiteral(line, col int) (Token, error) {
	l.advance()
	var buf []rune
	for !l.atEnd() && l.peek() != '"' {
		ch := l.peek()
		if ch == '\\' {
			l.advance()
			if l.atEnd() {
				return Token{TOKEN_ILLEGAL, "", line, col},
					fmt.Errorf("line %d:%d: unterminated escape", line, col)
			}
			esc := l.peek()
			l.advance()
			switch esc {
			case 'n':
				buf = append(buf, '\n')
			case 't':
				buf = append(buf, '\t')
			case 'r':
				buf = append(buf, '\r')
			case '\\':
				buf = append(buf, '\\')
			case '"':
				buf = append(buf, '"')
			default:
				return Token{TOKEN_ILLEGAL, "", line, col},
					fmt.Errorf("line %d:%d: invalid escape \\%c", line, col, esc)
			}
			continue
		}
		buf = append(buf, ch)
		l.advance()
	}
	if l.atEnd() {
		return Token{TOKEN_ILLEGAL, "", line, col},
			fmt.Errorf("line %d:%d: unterminated string literal", line, col)
	}
	l.advance()
	return Token{TOKEN_STRING, string(buf), line, col}, nil
}

func (l *Lexer) skipWhitespaceAndComments() {
	for !l.atEnd() {
		ch := l.peek()
		switch {
		case ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n':
			l.advance()
		case ch == '#':
			for !l.atEnd() && l.peek() != '\n' {
				l.advance()
			}
		case ch == '/' && l.pos+1 < len(l.source) && l.source[l.pos+1] == '*':
			l.advance()
			l.advance()
			for !l.atEnd() && !(l.peek() == '*' && l.pos+1 < len(l.source) && l.source[l.pos+1] == '/') {
				l.advance()
			}
			if !l.atEnd() {
				l.advance()
				l.advance()
			}
		default:
			return
		}
	}
}

func (l *Lexer) atEnd() bool { return l.pos >= len(l.source) }

func (l *Lexer) peek() rune { return l.source[l.pos] }

func (l *Lexer) advance() {
	if l.source[l.pos] == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	l.pos++
}

func (l *Lexer) match(expected rune) bool {
	if l.atEnd() || l.peek() != expected {
		return false
	}
	l.advance()
	return true
}

func (l *Lexer) makeToken(t TokenType, lex string) Token {
	return Token{Type: t, Lexeme: lex, Line: l.line, Column: l.col}
}

func isLetter(r rune) bool { return unicode.IsLetter(r) || r == '_' }
func isDigit(r rune) bool  { return r >= '0' && r <= '9' }
