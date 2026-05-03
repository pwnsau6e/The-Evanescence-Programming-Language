# The Evanescence Programming Language

Evanescence (`.eve`) is a small, **interpreted, dynamically-typed** language with
Python-like semantics, but **without significant whitespace** — blocks are
delimited by braces `{ ... }` and statements end with `;`.

The interpreter is written in Go and is split into three independent stages:

| Stage      | Package           | Responsibility                       |
| ---------- | ----------------- | ------------------------------------ |
| Tokenizer  | `lexer/`          | Source text → stream of tokens       |
| Parser     | `parser/`         | Tokens → Abstract Syntax Tree        |
| Evaluator  | `evaluator/`      | Tree-walking interpreter             |
| CLI        | `main.go`         | `eve file.eve` runner + REPL         |

---

## Build & run

```bash
# build the interpreter
go build -o eve .

# run a program
./eve examples/fizzbuzz.eve

# start the REPL (no arguments)
./eve

# run all tests
go test ./...
```

On Windows, use `eve.exe` and the same commands work in `bash` or PowerShell.

---

## Language tour

### Hello, world

```evanescence
print("Hello, Evanescence!");
```

### Variables and constants

```evanescence
var   x  = 10;
const PI = 3.14;

x = x + 1;     # OK
PI = 3;        # error: cannot assign to constant "PI"
```

### Numbers, strings, booleans, lists, null

```evanescence
var n  = 42;
var f  = 3.14;
var s  = "hello";
var b  = true;
var z  = null;
var xs = [1, 2, 3, "mixed", false];

print(xs[3]);       # mixed
print(len(xs));     # 5
```

### Arithmetic and comparison

```evanescence
print(2 + 3 * 4);       # 14
print(2 ** 10);         # 1024  (power is right-associative)
print(7 // 2);          # 3     (floor division)
print(7 %  2);          # 1     (modulo)
print(1 < 2 and 3 > 2); # true
```

### Control flow

```evanescence
var x = 5;

if (x > 10) {
    print("big");
} else if (x > 3) {
    print("medium");
} else {
    print("small");
}

var i = 0;
while (i < 3) {
    print(i);
    i += 1;
}
```

### Loops

```evanescence
# range(stop) | range(start, stop) | range(start, stop, step)
for i in range(0, 10, 2) {
    print(i);            # 0 2 4 6 8
}

# iterate any list (or string, character by character)
for fruit in ["apple", "banana", "cherry"] {
    print(fruit);
}

# break and continue work as you'd expect
for i in range(0, 10) {
    if (i == 3) { continue; }
    if (i == 6) { break;    }
    print(i);                # 0 1 2 4 5
}
```

### Functions and closures

```evanescence
func add(a, b) {
    return a + b;
}
print(add(2, 3));   # 5

# A function with no return implicitly returns null — that's a "procedure".
func greet(name) {
    print("hi " + name);
}
greet("world");

# Inner functions capture the enclosing scope (true closures).
func makeCounter() {
    var n = 0;
    func tick() {
        n += 1;
        return n;
    }
    return tick;
}
var c = makeCounter();
print(c()); print(c()); print(c());   # 1 2 3
```

### Built-in functions

| Built-in       | Description                                         |
| -------------- | --------------------------------------------------- |
| `print(x...)`  | Print values separated by spaces, then newline      |
| `input(p?)`    | Read a line from stdin (with optional prompt `p`)   |
| `len(x)`       | Length of a string or list                          |
| `str(x)`       | Convert a value to its string representation        |
| `int(x)`       | Convert a number/string/bool to an integer          |
| `type(x)`      | Type name: `"number"`, `"string"`, `"bool"`, ...    |

### Comments

```evanescence
# line comment (Python-style)

/*
  block comment, can span
  multiple lines
*/
```

---

## Grammar (EBNF, abbreviated)

```
program        = statement* EOF
statement      = var_decl | const_decl | assign_stmt | if_stmt
               | while_stmt | for_stmt | func_decl | return_stmt
               | break_stmt | continue_stmt | expr_stmt | block

var_decl       = "var"   IDENT "=" expression ";"
const_decl     = "const" IDENT "=" expression ";"
assign_stmt    = IDENT ("=" | "+=" | "-=" | "*=" | "/=") expression ";"

if_stmt        = "if" "(" expression ")" block
                 { "else" "if" "(" expression ")" block }
                 [ "else" block ]
while_stmt     = "while" "(" expression ")" block
for_stmt       = "for" IDENT "in" ( range_expr | expression ) block
range_expr     = "range" "(" expression
                          [ "," expression [ "," expression ] ] ")"
func_decl      = "func" IDENT "(" [ IDENT { "," IDENT } ] ")" block
block          = "{" statement* "}"

expression     = logic_or
logic_or       = logic_and { "or"  logic_and }
logic_and      = equality  { "and" equality  }
equality       = comparison { ("=="|"!=") comparison }
comparison     = term       { ("<"|"<="|">"|">=") term }
term           = factor     { ("+"|"-") factor }
factor         = unary      { ("*"|"/"|"//"|"%") unary }
unary          = ("!"|"-") unary | power
power          = call [ "**" unary ]            # right-associative
call           = primary { "(" [ args ] ")" | "[" expression "]" }
primary        = NUMBER | STRING | "true" | "false" | "null"
               | IDENT | "(" expression ")" | list_literal
list_literal   = "[" [ expression { "," expression } ] "]"
```

The full grammar lives in [grammar.txt](grammar.txt).

---

## Project layout

```
.
├── main.go              # CLI entry point ('eve file.eve' + REPL)
├── go.mod
├── grammar.txt          # Full language grammar
├── lexer/               # Tokenizer
│   ├── token.go
│   ├── lexer.go
│   └── lexer_test.go
├── parser/              # Recursive-descent parser & AST
│   ├── ast.go
│   ├── parser.go
│   └── parser_test.go
├── evaluator/           # Tree-walking interpreter
│   ├── object.go
│   ├── environment.go
│   ├── builtins.go
│   ├── evaluator.go
│   └── evaluator_test.go
└── examples/            # Sample .eve programs
    ├── hello.eve
    ├── factorial.eve
    └── guess.eve
```

Every phase has its own test file. To verify any single stage in isolation:

```bash
go test ./lexer/      -v
go test ./parser/     -v
go test ./evaluator/  -v
```


# RESOURCES 

- https://craftinginterpreters.com/contents.html
- https://github.com/jablonskidev/writing-an-interpreter-in-go
- https://edu.anarcho-copy.org/Programming%20Languages/Go/writing%20an%20INTERPRETER%20in%20go.pdf