package evaluator

import (
	"bytes"
	"strings"
	"testing"

	"evanescence/lexer"
	"evanescence/parser"
)

// run lexes, parses, and evaluates src; returns (stdout, error).
// stdin can be supplied via the optional input string.
func run(t *testing.T, src string, input ...string) (string, error) {
	t.Helper()
	tokens, err := lexer.New(src).Tokenize()
	if err != nil {
		t.Fatalf("lex error: %v", err)
	}
	prog, err := parser.New(tokens).Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	out := &bytes.Buffer{}
	StdOut = out
	defer func() { StdOut = nil }()
	if len(input) > 0 {
		StdIn = strings.NewReader(input[0])
		defer func() { StdIn = nil }()
	}
	interp := New()
	return out.String(), runAndCapture(interp, prog, out)
}

func runAndCapture(interp *Interpreter, prog *parser.Program, out *bytes.Buffer) error {
	err := interp.Run(prog)
	// out already captured; return any runtime error
	return err
}

// runOK runs src expecting no error and returns stdout.
func runOK(t *testing.T, src string, input ...string) string {
	t.Helper()
	out, err := run(t, src, input...)
	// stdout buffer was reset before Run started; re-fetch via the closure.
	if err != nil {
		t.Fatalf("runtime error: %v", err)
	}
	return out
}

// We use a local helper that runs and grabs stdout cleanly.
func eval(t *testing.T, src string) string {
	t.Helper()
	tokens, err := lexer.New(src).Tokenize()
	if err != nil {
		t.Fatalf("lex error: %v", err)
	}
	prog, err := parser.New(tokens).Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	out := &bytes.Buffer{}
	StdOut = out
	defer func() { StdOut = nil }()
	if err := New().Run(prog); err != nil {
		t.Fatalf("runtime error: %v", err)
	}
	return out.String()
}

func evalErr(t *testing.T, src string) error {
	t.Helper()
	tokens, _ := lexer.New(src).Tokenize()
	prog, perr := parser.New(tokens).Parse()
	if perr != nil {
		return perr
	}
	out := &bytes.Buffer{}
	StdOut = out
	defer func() { StdOut = nil }()
	return New().Run(prog)
}

func TestPrintAndArithmetic(t *testing.T) {
	got := eval(t, `print(1 + 2 * 3);`)
	if got != "7\n" {
		t.Errorf("got %q, want %q", got, "7\n")
	}
}

func TestVariables(t *testing.T) {
	got := eval(t, `var x = 10; var y = 20; print(x + y);`)
	if got != "30\n" {
		t.Errorf("got %q", got)
	}
}

func TestConstAssignError(t *testing.T) {
	err := evalErr(t, `const PI = 3; PI = 4;`)
	if err == nil || !strings.Contains(err.Error(), "constant") {
		t.Errorf("expected constant assignment error, got %v", err)
	}
}

func TestStringConcat(t *testing.T) {
	got := eval(t, `print("hello, " + "world");`)
	if got != "hello, world\n" {
		t.Errorf("got %q", got)
	}
}

func TestIfElse(t *testing.T) {
	got := eval(t, `
        var x = 5;
        if (x > 10) { print("big"); }
        else if (x > 3) { print("medium"); }
        else { print("small"); }
    `)
	if got != "medium\n" {
		t.Errorf("got %q", got)
	}
}

func TestWhileLoop(t *testing.T) {
	got := eval(t, `
        var i = 0;
        while (i < 3) {
            print(i);
            i += 1;
        }
    `)
	if got != "0\n1\n2\n" {
		t.Errorf("got %q", got)
	}
}

func TestForRange(t *testing.T) {
	got := eval(t, `for i in range(1, 4) { print(i); }`)
	if got != "1\n2\n3\n" {
		t.Errorf("got %q", got)
	}
}

func TestForRangeStep(t *testing.T) {
	got := eval(t, `for i in range(0, 10, 2) { print(i); }`)
	if got != "0\n2\n4\n6\n8\n" {
		t.Errorf("got %q", got)
	}
}

func TestForList(t *testing.T) {
	got := eval(t, `for x in ["a", "b", "c"] { print(x); }`)
	if got != "a\nb\nc\n" {
		t.Errorf("got %q", got)
	}
}

func TestFunctionCall(t *testing.T) {
	got := eval(t, `
        func add(a, b) { return a + b; }
        print(add(2, 3));
    `)
	if got != "5\n" {
		t.Errorf("got %q", got)
	}
}

func TestRecursion(t *testing.T) {
	got := eval(t, `
        func fact(n) {
            if (n <= 1) { return 1; }
            return n * fact(n - 1);
        }
        print(fact(5));
    `)
	if got != "120\n" {
		t.Errorf("got %q, want 120", got)
	}
}

func TestClosure(t *testing.T) {
	got := eval(t, `
        func makeCounter() {
            var n = 0;
            func inc() {
                n += 1;
                return n;
            }
            return inc;
        }
        var c = makeCounter();
        print(c());
        print(c());
        print(c());
    `)
	if got != "1\n2\n3\n" {
		t.Errorf("got %q, want 1\\n2\\n3\\n", got)
	}
}

func TestBreakContinue(t *testing.T) {
	got := eval(t, `
        for i in range(0, 10) {
            if (i == 3) { continue; }
            if (i == 6) { break; }
            print(i);
        }
    `)
	if got != "0\n1\n2\n4\n5\n" {
		t.Errorf("got %q", got)
	}
}

func TestPower(t *testing.T) {
	got := eval(t, `print(2 ** 10);`)
	if got != "1024\n" {
		t.Errorf("got %q", got)
	}
}

func TestFloorDivAndModulo(t *testing.T) {
	got := eval(t, `print(7 // 2); print(7 % 2);`)
	if got != "3\n1\n" {
		t.Errorf("got %q", got)
	}
}

func TestLogicalShortCircuit(t *testing.T) {
	// If short-circuit works, nope() never runs and no error is raised.
	got := eval(t, `
        func nope() { return 1 / 0; }
        if (false and nope()) { print("X"); }
        if (true  or  nope()) { print("Y"); }
    `)
	if got != "Y\n" {
		t.Errorf("got %q", got)
	}
}

func TestComparisonAndBool(t *testing.T) {
	got := eval(t, `print(1 < 2); print(2 == 2); print(!false);`)
	if got != "true\ntrue\ntrue\n" {
		t.Errorf("got %q", got)
	}
}

func TestListIndex(t *testing.T) {
	got := eval(t, `var a = [10, 20, 30]; print(a[1]);`)
	if got != "20\n" {
		t.Errorf("got %q", got)
	}
}

func TestLenAndStrAndType(t *testing.T) {
	got := eval(t, `
        print(len("hello"));
        print(len([1,2,3,4]));
        print(str(42));
        print(type(3.14));
        print(type("x"));
        print(type([]));
    `)
	want := "5\n4\n42\nnumber\nstring\nlist\n"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestDivisionByZero(t *testing.T) {
	if err := evalErr(t, `print(1 / 0);`); err == nil {
		t.Fatal("expected division-by-zero error")
	}
}

func TestUndefinedVariable(t *testing.T) {
	if err := evalErr(t, `print(z);`); err == nil {
		t.Fatal("expected undefined-variable error")
	}
}

func TestInputBuiltin(t *testing.T) {
	tokens, _ := lexer.New(`var name = input(); print("hi " + name);`).Tokenize()
	prog, _ := parser.New(tokens).Parse()
	out := &bytes.Buffer{}
	StdOut = out
	StdIn = strings.NewReader("alice\n")
	defer func() { StdOut = nil; StdIn = nil }()
	if err := New().Run(prog); err != nil {
		t.Fatalf("runtime error: %v", err)
	}
	if out.String() != "hi alice\n" {
		t.Errorf("got %q", out.String())
	}
}
