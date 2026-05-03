package evaluator

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
)

// stdin/stdout are package-level so tests can swap them out.
var (
	StdOut io.Writer = os.Stdout
	StdIn  io.Reader = os.Stdin
)

// inputReader wraps StdIn lazily; it's recreated whenever StdIn changes
// (so tests can swap StdIn between runs and still get correct buffering).
var (
	inputReader *bufio.Reader
	lastStdIn   io.Reader
)

// builtins returns the table of built-in functions, fresh per interpreter
// (so different interpreters could in principle install different sets).
func builtins() map[string]*BuiltinValue {
	return map[string]*BuiltinValue{
		"print": {Name: "print", Fn: builtinPrint},
		"input": {Name: "input", Fn: builtinInput},
		"len":   {Name: "len", Fn: builtinLen},
		"str":   {Name: "str", Fn: builtinStr},
		"int":   {Name: "int", Fn: builtinInt},
		"type":  {Name: "type", Fn: builtinType},
	}
}

func builtinPrint(args []Value) (Value, error) {
	for i, a := range args {
		if i > 0 {
			fmt.Fprint(StdOut, " ")
		}
		fmt.Fprint(StdOut, a.String())
	}
	fmt.Fprintln(StdOut)
	return &NullValue{}, nil
}

func builtinInput(args []Value) (Value, error) {
	if len(args) > 0 {
		fmt.Fprint(StdOut, args[0].String())
	}
	if inputReader == nil || lastStdIn != StdIn {
		inputReader = bufio.NewReader(StdIn)
		lastStdIn = StdIn
	}
	line, err := inputReader.ReadString('\n')
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("input(): %v", err)
	}
	// strip trailing \r and \n (Windows line endings included)
	for len(line) > 0 && (line[len(line)-1] == '\n' || line[len(line)-1] == '\r') {
		line = line[:len(line)-1]
	}
	return &StringValue{Value: line}, nil
}

func builtinLen(args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("len() takes exactly 1 argument, got %d", len(args))
	}
	switch v := args[0].(type) {
	case *StringValue:
		return &NumberValue{Value: float64(len(v.Value))}, nil
	case *ListValue:
		return &NumberValue{Value: float64(len(v.Elements))}, nil
	}
	return nil, fmt.Errorf("len() not supported for %s", args[0].TypeName())
}

func builtinStr(args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("str() takes exactly 1 argument, got %d", len(args))
	}
	return &StringValue{Value: args[0].String()}, nil
}

func builtinInt(args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("int() takes exactly 1 argument, got %d", len(args))
	}
	switch v := args[0].(type) {
	case *NumberValue:
		return &NumberValue{Value: float64(int64(v.Value))}, nil
	case *StringValue:
		n, err := strconv.ParseInt(v.Value, 10, 64)
		if err != nil {
			// try float then truncate
			f, ferr := strconv.ParseFloat(v.Value, 64)
			if ferr != nil {
				return nil, fmt.Errorf("int(): cannot convert %q", v.Value)
			}
			return &NumberValue{Value: float64(int64(f))}, nil
		}
		return &NumberValue{Value: float64(n)}, nil
	case *BoolValue:
		if v.Value {
			return &NumberValue{Value: 1}, nil
		}
		return &NumberValue{Value: 0}, nil
	}
	return nil, fmt.Errorf("int(): cannot convert %s", args[0].TypeName())
}

func builtinType(args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("type() takes exactly 1 argument, got %d", len(args))
	}
	return &StringValue{Value: args[0].TypeName()}, nil
}
