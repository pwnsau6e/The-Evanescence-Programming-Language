package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"evanescence/evaluator"
	"evanescence/lexer"
	"evanescence/parser"
)

func main() {
	args := os.Args[1:]
	switch len(args) {
	case 0:
		runREPL()
	case 1:
		runFile(args[0])
	default:
		fmt.Fprintln(os.Stderr, "usage: eve [file.eve]")
		os.Exit(2)
	}
}

func runFile(path string) {
	if filepath.Ext(path) != ".eve" {
		fmt.Fprintf(os.Stderr, "warning: expected a .eve file, got %q\n", path)
	}
	src, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot read %s: %v\n", path, err)
		os.Exit(1)
	}
	if err := exec(string(src)); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runREPL() {
	fmt.Println("Evanescence REPL — type 'exit' to quit. Statements must end with ';'.")
	interp := evaluator.New()
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("eve> ")
		if !scanner.Scan() {
			return
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "exit" || line == "quit" {
			return
		}
		if line == "" {
			continue
		}
		// allow REPL users to omit the trailing ';' for convenience
		if !strings.HasSuffix(line, ";") && !strings.HasSuffix(line, "}") {
			line += ";"
		}
		if err := execWith(interp, line); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
		}
	}
}

func exec(src string) error {
	tokens, err := lexer.New(src).Tokenize()
	if err != nil {
		return err
	}
	prog, err := parser.New(tokens).Parse()
	if err != nil {
		return err
	}
	return evaluator.New().Run(prog)
}

func execWith(interp *evaluator.Interpreter, src string) error {
	tokens, err := lexer.New(src).Tokenize()
	if err != nil {
		return err
	}
	prog, err := parser.New(tokens).Parse()
	if err != nil {
		return err
	}
	return interp.Run(prog)
}
