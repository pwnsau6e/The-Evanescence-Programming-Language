package evaluator

import (
	"fmt"
	"strconv"
	"strings"

	"evanescence/parser"
)

// Value is any runtime value in Evanescence.
// We use a small set of concrete types over an interface to keep the
// evaluator easy to read and to avoid tagging-by-string.
type Value interface {
	TypeName() string
	String() string
}

// ----- Concrete value types -----

type NumberValue struct{ Value float64 }
type StringValue struct{ Value string }
type BoolValue struct{ Value bool }
type NullValue struct{}
type ListValue struct{ Elements []Value }

// FunctionValue is a user-defined function with its captured environment.
type FunctionValue struct {
	Name    string
	Params  []string
	Body    *parser.BlockStmt
	Closure *Environment
}

// BuiltinValue is a Go function exposed to Evanescence (e.g. print, len).
type BuiltinValue struct {
	Name string
	Fn   func(args []Value) (Value, error)
}

func (n *NumberValue) TypeName() string   { return "number" }
func (s *StringValue) TypeName() string   { return "string" }
func (b *BoolValue) TypeName() string     { return "bool" }
func (n *NullValue) TypeName() string     { return "null" }
func (l *ListValue) TypeName() string     { return "list" }
func (f *FunctionValue) TypeName() string { return "function" }
func (b *BuiltinValue) TypeName() string  { return "builtin" }

func (n *NumberValue) String() string {
	if n.Value == float64(int64(n.Value)) {
		return strconv.FormatInt(int64(n.Value), 10)
	}
	return strconv.FormatFloat(n.Value, 'g', -1, 64)
}
func (s *StringValue) String() string { return s.Value }
func (b *BoolValue) String() string {
	if b.Value {
		return "true"
	}
	return "false"
}
func (n *NullValue) String() string { return "null" }
func (l *ListValue) String() string {
	parts := make([]string, len(l.Elements))
	for i, e := range l.Elements {
		if s, ok := e.(*StringValue); ok {
			parts[i] = "\"" + s.Value + "\""
		} else {
			parts[i] = e.String()
		}
	}
	return "[" + strings.Join(parts, ", ") + "]"
}
func (f *FunctionValue) String() string { return fmt.Sprintf("<func %s>", f.Name) }
func (b *BuiltinValue) String() string  { return fmt.Sprintf("<builtin %s>", b.Name) }

// Truthy returns the boolean view of a value, mirroring Python rules:
// 0, empty string, empty list, false, and null are falsy; everything else is true.
func Truthy(v Value) bool {
	switch x := v.(type) {
	case *BoolValue:
		return x.Value
	case *NullValue:
		return false
	case *NumberValue:
		return x.Value != 0
	case *StringValue:
		return x.Value != ""
	case *ListValue:
		return len(x.Elements) > 0
	}
	return true
}

// Equal compares two values by their underlying representation.
// Different types compare as not-equal, except numbers vs numbers.
func Equal(a, b Value) bool {
	switch x := a.(type) {
	case *NumberValue:
		y, ok := b.(*NumberValue)
		return ok && x.Value == y.Value
	case *StringValue:
		y, ok := b.(*StringValue)
		return ok && x.Value == y.Value
	case *BoolValue:
		y, ok := b.(*BoolValue)
		return ok && x.Value == y.Value
	case *NullValue:
		_, ok := b.(*NullValue)
		return ok
	case *ListValue:
		y, ok := b.(*ListValue)
		if !ok || len(x.Elements) != len(y.Elements) {
			return false
		}
		for i := range x.Elements {
			if !Equal(x.Elements[i], y.Elements[i]) {
				return false
			}
		}
		return true
	}
	return false
}

// ----- Internal control-flow signals -----
//
// We model `return`, `break`, and `continue` as sentinel error types.
// This is a small, well-known trick that avoids threading status flags
// through every visitor method.

type returnSignal struct{ Value Value }
type breakSignal struct{}
type continueSignal struct{}

func (returnSignal) Error() string   { return "return" }
func (breakSignal) Error() string    { return "break" }
func (continueSignal) Error() string { return "continue" }
