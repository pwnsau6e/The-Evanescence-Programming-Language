package parser

import (
	"testing"

	"evanescence/lexer"
)

func parse(t *testing.T, src string) *Program {
	t.Helper()
	tokens, err := lexer.New(src).Tokenize()
	if err != nil {
		t.Fatalf("lex error: %v", err)
	}
	p, err := New(tokens).Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	return p
}

func TestVarDecl(t *testing.T) {
	p := parse(t, `var x = 42;`)
	if len(p.Statements) != 1 {
		t.Fatalf("got %d statements", len(p.Statements))
	}
	v, ok := p.Statements[0].(*VarDecl)
	if !ok {
		t.Fatalf("expected *VarDecl, got %T", p.Statements[0])
	}
	if v.Name != "x" || v.IsConst {
		t.Errorf("bad var decl: %+v", v)
	}
	if n, ok := v.Value.(*NumberLiteral); !ok || n.Value != 42 {
		t.Errorf("bad value: %+v", v.Value)
	}
}

func TestConstDecl(t *testing.T) {
	p := parse(t, `const PI = 3.14;`)
	v := p.Statements[0].(*VarDecl)
	if !v.IsConst || v.Name != "PI" {
		t.Errorf("bad const decl: %+v", v)
	}
}

func TestArithmeticPrecedence(t *testing.T) {
	// 1 + 2 * 3  should parse as 1 + (2 * 3)
	p := parse(t, `var x = 1 + 2 * 3;`)
	v := p.Statements[0].(*VarDecl)
	bin, ok := v.Value.(*BinaryExpr)
	if !ok || bin.Op != "+" {
		t.Fatalf("expected '+' at root, got %+v", v.Value)
	}
	right, ok := bin.Right.(*BinaryExpr)
	if !ok || right.Op != "*" {
		t.Fatalf("expected '*' on right, got %+v", bin.Right)
	}
}

func TestPowerRightAssociative(t *testing.T) {
	// 2 ** 3 ** 2  should parse as 2 ** (3 ** 2)
	p := parse(t, `var x = 2 ** 3 ** 2;`)
	v := p.Statements[0].(*VarDecl)
	bin := v.Value.(*BinaryExpr)
	if bin.Op != "**" {
		t.Fatalf("expected '**', got %s", bin.Op)
	}
	right, ok := bin.Right.(*BinaryExpr)
	if !ok || right.Op != "**" {
		t.Fatalf("right side should also be '**', got %+v", bin.Right)
	}
}

func TestIfElseChain(t *testing.T) {
	src := `
        if (x > 0) { var a = 1; }
        else if (x < 0) { var a = 2; }
        else { var a = 3; }
    `
	p := parse(t, src)
	stmt := p.Statements[0].(*IfStmt)
	elseIf, ok := stmt.Else.(*IfStmt)
	if !ok {
		t.Fatalf("expected else-if, got %T", stmt.Else)
	}
	if _, ok := elseIf.Else.(*BlockStmt); !ok {
		t.Fatalf("expected final else block, got %T", elseIf.Else)
	}
}

func TestWhile(t *testing.T) {
	p := parse(t, `while (i < 10) { i = i + 1; }`)
	w := p.Statements[0].(*WhileStmt)
	if w.Cond == nil || w.Body == nil {
		t.Fatal("nil pieces in while statement")
	}
}

func TestForRange(t *testing.T) {
	p := parse(t, `for i in range(0, 10, 2) { print(i); }`)
	f := p.Statements[0].(*ForStmt)
	if f.VarName != "i" {
		t.Errorf("bad var name: %q", f.VarName)
	}
	if len(f.RangeArgs) != 3 {
		t.Errorf("expected 3 range args, got %d", len(f.RangeArgs))
	}
	if f.Iterable != nil {
		t.Errorf("expected nil iterable")
	}
}

func TestForList(t *testing.T) {
	p := parse(t, `for x in [1, 2, 3] { print(x); }`)
	f := p.Statements[0].(*ForStmt)
	if f.RangeArgs != nil {
		t.Errorf("expected nil range args")
	}
	if _, ok := f.Iterable.(*ListLiteral); !ok {
		t.Errorf("expected list literal as iterable, got %T", f.Iterable)
	}
}

func TestFuncDecl(t *testing.T) {
	p := parse(t, `func add(a, b) { return a + b; }`)
	f := p.Statements[0].(*FuncDecl)
	if f.Name != "add" {
		t.Errorf("bad name: %q", f.Name)
	}
	if len(f.Params) != 2 || f.Params[0] != "a" || f.Params[1] != "b" {
		t.Errorf("bad params: %v", f.Params)
	}
	if len(f.Body.Statements) != 1 {
		t.Errorf("expected 1 body statement, got %d", len(f.Body.Statements))
	}
}

func TestFuncCallChain(t *testing.T) {
	p := parse(t, `var x = foo(1)(2);`)
	v := p.Statements[0].(*VarDecl)
	outer, ok := v.Value.(*CallExpr)
	if !ok {
		t.Fatalf("expected outer CallExpr, got %T", v.Value)
	}
	inner, ok := outer.Callee.(*CallExpr)
	if !ok {
		t.Fatalf("expected inner CallExpr as callee, got %T", outer.Callee)
	}
	if id, ok := inner.Callee.(*Identifier); !ok || id.Name != "foo" {
		t.Errorf("expected identifier 'foo', got %+v", inner.Callee)
	}
}

func TestIndexExpression(t *testing.T) {
	p := parse(t, `var x = arr[0];`)
	v := p.Statements[0].(*VarDecl)
	if _, ok := v.Value.(*IndexExpr); !ok {
		t.Fatalf("expected IndexExpr, got %T", v.Value)
	}
}

func TestAssignmentOps(t *testing.T) {
	p := parse(t, `x += 5; x -= 1; x *= 2; x /= 3; x = 10;`)
	want := []string{"+=", "-=", "*=", "/=", "="}
	if len(p.Statements) != 5 {
		t.Fatalf("expected 5 statements, got %d", len(p.Statements))
	}
	for i, s := range p.Statements {
		a := s.(*AssignStmt)
		if a.Op != want[i] {
			t.Errorf("stmt %d: got op %q, want %q", i, a.Op, want[i])
		}
	}
}

func TestParseError(t *testing.T) {
	tokens, _ := lexer.New(`var = 5;`).Tokenize()
	if _, err := New(tokens).Parse(); err == nil {
		t.Fatal("expected parse error for missing identifier")
	}
}

func TestBooleanAndNull(t *testing.T) {
	p := parse(t, `var a = true; var b = false; var c = null;`)
	if _, ok := p.Statements[0].(*VarDecl).Value.(*BoolLiteral); !ok {
		t.Error("expected BoolLiteral for true")
	}
	if _, ok := p.Statements[1].(*VarDecl).Value.(*BoolLiteral); !ok {
		t.Error("expected BoolLiteral for false")
	}
	if _, ok := p.Statements[2].(*VarDecl).Value.(*NullLiteral); !ok {
		t.Error("expected NullLiteral")
	}
}

func TestLogicalAndComparison(t *testing.T) {
	// a < b and b < c  -> ( (a<b) and (b<c) )
	p := parse(t, `var x = a < b and b < c;`)
	v := p.Statements[0].(*VarDecl)
	root := v.Value.(*BinaryExpr)
	if root.Op != "and" {
		t.Fatalf("expected 'and' at root, got %q", root.Op)
	}
}
