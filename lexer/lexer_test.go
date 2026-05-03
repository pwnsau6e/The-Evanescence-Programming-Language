package lexer

import "testing"

func TestSimpleTokens(t *testing.T) {
	src := `var x = 10;`
	want := []TokenType{
		TOKEN_VAR, TOKEN_IDENT, TOKEN_ASSIGN, TOKEN_NUMBER, TOKEN_SEMICOLON, TOKEN_EOF,
	}
	got, err := New(src).Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("got %d tokens, want %d: %#v", len(got), len(want), got)
	}
	for i, tt := range want {
		if got[i].Type != tt {
			t.Errorf("token %d: got %s, want %s", i, got[i].Type, tt)
		}
	}
}

func TestKeywordsAndIdentifiers(t *testing.T) {
	src := `if else while for in range func return break continue true false null and or var const myVar`
	want := []TokenType{
		TOKEN_IF, TOKEN_ELSE, TOKEN_WHILE, TOKEN_FOR, TOKEN_IN, TOKEN_RANGE,
		TOKEN_FUNC, TOKEN_RETURN, TOKEN_BREAK, TOKEN_CONTINUE,
		TOKEN_TRUE, TOKEN_FALSE, TOKEN_NULL, TOKEN_AND, TOKEN_OR,
		TOKEN_VAR, TOKEN_CONST, TOKEN_IDENT, TOKEN_EOF,
	}
	got, err := New(src).Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i, tt := range want {
		if got[i].Type != tt {
			t.Errorf("token %d (%q): got %s, want %s", i, got[i].Lexeme, got[i].Type, tt)
		}
	}
}

func TestOperators(t *testing.T) {
	src := `+ - * / // % ** += -= *= /= == != < <= > >= = ! and or`
	want := []TokenType{
		TOKEN_PLUS, TOKEN_MINUS, TOKEN_STAR, TOKEN_SLASH, TOKEN_FLOORDIV, TOKEN_PERCENT, TOKEN_POWER,
		TOKEN_PLUSEQ, TOKEN_MINUSEQ, TOKEN_STAREQ, TOKEN_SLASHEQ,
		TOKEN_EQ, TOKEN_NEQ, TOKEN_LT, TOKEN_LTE, TOKEN_GT, TOKEN_GTE,
		TOKEN_ASSIGN, TOKEN_BANG, TOKEN_AND, TOKEN_OR, TOKEN_EOF,
	}
	got, err := New(src).Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("token count: got %d want %d", len(got), len(want))
	}
	for i, tt := range want {
		if got[i].Type != tt {
			t.Errorf("token %d (%q): got %s, want %s", i, got[i].Lexeme, got[i].Type, tt)
		}
	}
}

func TestNumbersAndStrings(t *testing.T) {
	src := `42 3.14 "hello" "with \"quotes\" and \n newline"`
	got, err := New(src).Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got[0].Type != TOKEN_NUMBER || got[0].Lexeme != "42" {
		t.Errorf("expected NUMBER 42, got %v", got[0])
	}
	if got[1].Type != TOKEN_NUMBER || got[1].Lexeme != "3.14" {
		t.Errorf("expected NUMBER 3.14, got %v", got[1])
	}
	if got[2].Type != TOKEN_STRING || got[2].Lexeme != "hello" {
		t.Errorf("expected STRING hello, got %v", got[2])
	}
	if got[3].Type != TOKEN_STRING || got[3].Lexeme != "with \"quotes\" and \n newline" {
		t.Errorf("expected STRING with escapes, got %q", got[3].Lexeme)
	}
}

func TestCommentsAndWhitespaceIgnored(t *testing.T) {
	src := `
        # line comment
        var x = 1;     # inline
        /* block
           comment */
        var y = 2;
    `
	got, err := New(src).Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []TokenType{
		TOKEN_VAR, TOKEN_IDENT, TOKEN_ASSIGN, TOKEN_NUMBER, TOKEN_SEMICOLON,
		TOKEN_VAR, TOKEN_IDENT, TOKEN_ASSIGN, TOKEN_NUMBER, TOKEN_SEMICOLON,
		TOKEN_EOF,
	}
	if len(got) != len(want) {
		t.Fatalf("got %d tokens, want %d: %#v", len(got), len(want), got)
	}
	for i, tt := range want {
		if got[i].Type != tt {
			t.Errorf("token %d: got %s, want %s", i, got[i].Type, tt)
		}
	}
}

func TestUnterminatedString(t *testing.T) {
	_, err := New(`"oops`).Tokenize()
	if err == nil {
		t.Fatal("expected error for unterminated string, got none")
	}
}

func TestLineTracking(t *testing.T) {
	src := "var x = 1;\nvar y = 2;"
	got, _ := New(src).Tokenize()
	// the second `var` should be on line 2
	if got[5].Type != TOKEN_VAR || got[5].Line != 2 {
		t.Errorf("expected second VAR at line 2, got line %d", got[5].Line)
	}
}

func TestFunctionDeclaration(t *testing.T) {
	src := `func add(a, b) { return a + b; }`
	want := []TokenType{
		TOKEN_FUNC, TOKEN_IDENT, TOKEN_LPAREN, TOKEN_IDENT, TOKEN_COMMA, TOKEN_IDENT, TOKEN_RPAREN,
		TOKEN_LBRACE, TOKEN_RETURN, TOKEN_IDENT, TOKEN_PLUS, TOKEN_IDENT, TOKEN_SEMICOLON, TOKEN_RBRACE,
		TOKEN_EOF,
	}
	got, err := New(src).Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("token count: got %d want %d", len(got), len(want))
	}
	for i, tt := range want {
		if got[i].Type != tt {
			t.Errorf("token %d (%q): got %s, want %s", i, got[i].Lexeme, got[i].Type, tt)
		}
	}
}
