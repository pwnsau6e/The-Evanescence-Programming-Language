package evaluator

import (
	"errors"
	"fmt"
	"math"

	"evanescence/parser"
)

// Interpreter is a tree-walking evaluator over the AST produced by parser.
type Interpreter struct {
	globals *Environment
}

// New creates an Interpreter with built-ins pre-installed in the global env.
func New() *Interpreter {
	env := NewEnv()
	for name, fn := range builtins() {
		_ = env.Declare(name, fn, true)
	}
	return &Interpreter{globals: env}
}

// Run executes a parsed program.
func (i *Interpreter) Run(prog *parser.Program) error {
	for _, stmt := range prog.Statements {
		if err := i.execStmt(stmt, i.globals); err != nil {
			// Top-level break/continue/return are program errors.
			var rs returnSignal
			if errors.As(err, &rs) {
				return fmt.Errorf("'return' outside of function")
			}
			var bs breakSignal
			if errors.As(err, &bs) {
				return fmt.Errorf("'break' outside of loop")
			}
			var cs continueSignal
			if errors.As(err, &cs) {
				return fmt.Errorf("'continue' outside of loop")
			}
			return err
		}
	}
	return nil
}

// ----- Statements -----

func (i *Interpreter) execStmt(s parser.Statement, env *Environment) error {
	switch n := s.(type) {
	case *parser.VarDecl:
		val, err := i.evalExpr(n.Value, env)
		if err != nil {
			return err
		}
		return env.Declare(n.Name, val, n.IsConst)
	case *parser.AssignStmt:
		return i.execAssign(n, env)
	case *parser.IfStmt:
		return i.execIf(n, env)
	case *parser.WhileStmt:
		return i.execWhile(n, env)
	case *parser.ForStmt:
		return i.execFor(n, env)
	case *parser.FuncDecl:
		fn := &FunctionValue{Name: n.Name, Params: n.Params, Body: n.Body, Closure: env}
		return env.Declare(n.Name, fn, false)
	case *parser.ReturnStmt:
		var v Value = &NullValue{}
		if n.Value != nil {
			r, err := i.evalExpr(n.Value, env)
			if err != nil {
				return err
			}
			v = r
		}
		return returnSignal{Value: v}
	case *parser.BreakStmt:
		return breakSignal{}
	case *parser.ContinueStmt:
		return continueSignal{}
	case *parser.ExprStmt:
		_, err := i.evalExpr(n.Expr, env)
		return err
	case *parser.BlockStmt:
		return i.execBlock(n, env.NewChild())
	}
	return fmt.Errorf("unknown statement type %T", s)
}

func (i *Interpreter) execBlock(b *parser.BlockStmt, env *Environment) error {
	for _, s := range b.Statements {
		if err := i.execStmt(s, env); err != nil {
			return err
		}
	}
	return nil
}

func (i *Interpreter) execAssign(n *parser.AssignStmt, env *Environment) error {
	rhs, err := i.evalExpr(n.Value, env)
	if err != nil {
		return err
	}
	if n.Op == "=" {
		return env.Assign(n.Name, rhs)
	}
	// Compound: load current, apply op, store back.
	cur, ok := env.Get(n.Name)
	if !ok {
		return fmt.Errorf("undefined variable %q", n.Name)
	}
	op := string(n.Op[0])
	result, err := applyBinary(op, cur, rhs)
	if err != nil {
		return err
	}
	return env.Assign(n.Name, result)
}

func (i *Interpreter) execIf(n *parser.IfStmt, env *Environment) error {
	cond, err := i.evalExpr(n.Cond, env)
	if err != nil {
		return err
	}
	if Truthy(cond) {
		return i.execBlock(n.Then, env.NewChild())
	}
	if n.Else != nil {
		return i.execStmt(n.Else, env)
	}
	return nil
}

func (i *Interpreter) execWhile(n *parser.WhileStmt, env *Environment) error {
	for {
		cond, err := i.evalExpr(n.Cond, env)
		if err != nil {
			return err
		}
		if !Truthy(cond) {
			return nil
		}
		err = i.execBlock(n.Body, env.NewChild())
		if err != nil {
			var bs breakSignal
			if errors.As(err, &bs) {
				return nil
			}
			var cs continueSignal
			if errors.As(err, &cs) {
				continue
			}
			return err
		}
	}
}

func (i *Interpreter) execFor(n *parser.ForStmt, env *Environment) error {
	var items []Value
	if n.RangeArgs != nil {
		seq, err := i.buildRange(n.RangeArgs, env)
		if err != nil {
			return err
		}
		items = seq
	} else {
		val, err := i.evalExpr(n.Iterable, env)
		if err != nil {
			return err
		}
		switch v := val.(type) {
		case *ListValue:
			items = v.Elements
		case *StringValue:
			items = make([]Value, 0, len(v.Value))
			for _, ch := range v.Value {
				items = append(items, &StringValue{Value: string(ch)})
			}
		default:
			return fmt.Errorf("cannot iterate over %s", val.TypeName())
		}
	}

	for _, it := range items {
		loopEnv := env.NewChild()
		_ = loopEnv.Declare(n.VarName, it, false)
		if err := i.execBlock(n.Body, loopEnv); err != nil {
			var bs breakSignal
			if errors.As(err, &bs) {
				return nil
			}
			var cs continueSignal
			if errors.As(err, &cs) {
				continue
			}
			return err
		}
	}
	return nil
}

func (i *Interpreter) buildRange(args []parser.Expression, env *Environment) ([]Value, error) {
	nums := make([]float64, len(args))
	for j, a := range args {
		v, err := i.evalExpr(a, env)
		if err != nil {
			return nil, err
		}
		n, ok := v.(*NumberValue)
		if !ok {
			return nil, fmt.Errorf("range() requires numbers, got %s", v.TypeName())
		}
		nums[j] = n.Value
	}
	var start, stop, step float64
	switch len(nums) {
	case 1:
		start, stop, step = 0, nums[0], 1
	case 2:
		start, stop, step = nums[0], nums[1], 1
	case 3:
		start, stop, step = nums[0], nums[1], nums[2]
	}
	if step == 0 {
		return nil, fmt.Errorf("range() step must not be zero")
	}
	var out []Value
	if step > 0 {
		for v := start; v < stop; v += step {
			out = append(out, &NumberValue{Value: v})
		}
	} else {
		for v := start; v > stop; v += step {
			out = append(out, &NumberValue{Value: v})
		}
	}
	return out, nil
}

// ----- Expressions -----

func (i *Interpreter) evalExpr(e parser.Expression, env *Environment) (Value, error) {
	switch n := e.(type) {
	case *parser.NumberLiteral:
		return &NumberValue{Value: n.Value}, nil
	case *parser.StringLiteral:
		return &StringValue{Value: n.Value}, nil
	case *parser.BoolLiteral:
		return &BoolValue{Value: n.Value}, nil
	case *parser.NullLiteral:
		return &NullValue{}, nil
	case *parser.Identifier:
		v, ok := env.Get(n.Name)
		if !ok {
			return nil, fmt.Errorf("undefined variable %q", n.Name)
		}
		return v, nil
	case *parser.ListLiteral:
		out := &ListValue{}
		for _, el := range n.Elements {
			v, err := i.evalExpr(el, env)
			if err != nil {
				return nil, err
			}
			out.Elements = append(out.Elements, v)
		}
		return out, nil
	case *parser.UnaryExpr:
		return i.evalUnary(n, env)
	case *parser.BinaryExpr:
		return i.evalBinary(n, env)
	case *parser.CallExpr:
		return i.evalCall(n, env)
	case *parser.IndexExpr:
		return i.evalIndex(n, env)
	}
	return nil, fmt.Errorf("unknown expression type %T", e)
}

func (i *Interpreter) evalUnary(n *parser.UnaryExpr, env *Environment) (Value, error) {
	v, err := i.evalExpr(n.Operand, env)
	if err != nil {
		return nil, err
	}
	switch n.Op {
	case "-":
		num, ok := v.(*NumberValue)
		if !ok {
			return nil, fmt.Errorf("unary '-' requires number, got %s", v.TypeName())
		}
		return &NumberValue{Value: -num.Value}, nil
	case "!":
		return &BoolValue{Value: !Truthy(v)}, nil
	}
	return nil, fmt.Errorf("unknown unary op %q", n.Op)
}

func (i *Interpreter) evalBinary(n *parser.BinaryExpr, env *Environment) (Value, error) {
	// Short-circuit logical operators.
	if n.Op == "and" {
		l, err := i.evalExpr(n.Left, env)
		if err != nil {
			return nil, err
		}
		if !Truthy(l) {
			return l, nil
		}
		return i.evalExpr(n.Right, env)
	}
	if n.Op == "or" {
		l, err := i.evalExpr(n.Left, env)
		if err != nil {
			return nil, err
		}
		if Truthy(l) {
			return l, nil
		}
		return i.evalExpr(n.Right, env)
	}
	l, err := i.evalExpr(n.Left, env)
	if err != nil {
		return nil, err
	}
	r, err := i.evalExpr(n.Right, env)
	if err != nil {
		return nil, err
	}
	return applyBinary(n.Op, l, r)
}

func applyBinary(op string, l, r Value) (Value, error) {
	switch op {
	case "==":
		return &BoolValue{Value: Equal(l, r)}, nil
	case "!=":
		return &BoolValue{Value: !Equal(l, r)}, nil
	}
	// String concatenation
	if op == "+" {
		if ls, lok := l.(*StringValue); lok {
			if rs, rok := r.(*StringValue); rok {
				return &StringValue{Value: ls.Value + rs.Value}, nil
			}
		}
	}
	// All remaining ops are numeric.
	ln, lok := l.(*NumberValue)
	rn, rok := r.(*NumberValue)
	if !lok || !rok {
		return nil, fmt.Errorf("operator %q not supported between %s and %s", op, l.TypeName(), r.TypeName())
	}
	a, b := ln.Value, rn.Value
	switch op {
	case "+":
		return &NumberValue{Value: a + b}, nil
	case "-":
		return &NumberValue{Value: a - b}, nil
	case "*":
		return &NumberValue{Value: a * b}, nil
	case "/":
		if b == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return &NumberValue{Value: a / b}, nil
	case "//":
		if b == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return &NumberValue{Value: math.Floor(a / b)}, nil
	case "%":
		if b == 0 {
			return nil, fmt.Errorf("modulo by zero")
		}
		return &NumberValue{Value: math.Mod(a, b)}, nil
	case "**":
		return &NumberValue{Value: math.Pow(a, b)}, nil
	case "<":
		return &BoolValue{Value: a < b}, nil
	case "<=":
		return &BoolValue{Value: a <= b}, nil
	case ">":
		return &BoolValue{Value: a > b}, nil
	case ">=":
		return &BoolValue{Value: a >= b}, nil
	}
	return nil, fmt.Errorf("unknown binary op %q", op)
}

func (i *Interpreter) evalCall(n *parser.CallExpr, env *Environment) (Value, error) {
	callee, err := i.evalExpr(n.Callee, env)
	if err != nil {
		return nil, err
	}
	args := make([]Value, len(n.Args))
	for j, a := range n.Args {
		v, err := i.evalExpr(a, env)
		if err != nil {
			return nil, err
		}
		args[j] = v
	}
	switch fn := callee.(type) {
	case *BuiltinValue:
		return fn.Fn(args)
	case *FunctionValue:
		if len(args) != len(fn.Params) {
			return nil, fmt.Errorf("function %s expected %d args, got %d", fn.Name, len(fn.Params), len(args))
		}
		callEnv := fn.Closure.NewChild()
		for j, p := range fn.Params {
			_ = callEnv.Declare(p, args[j], false)
		}
		err := i.execBlock(fn.Body, callEnv)
		if err == nil {
			return &NullValue{}, nil
		}
		var rs returnSignal
		if errors.As(err, &rs) {
			return rs.Value, nil
		}
		return nil, err
	}
	return nil, fmt.Errorf("cannot call value of type %s", callee.TypeName())
}

func (i *Interpreter) evalIndex(n *parser.IndexExpr, env *Environment) (Value, error) {
	col, err := i.evalExpr(n.Collection, env)
	if err != nil {
		return nil, err
	}
	idx, err := i.evalExpr(n.Index, env)
	if err != nil {
		return nil, err
	}
	idxNum, ok := idx.(*NumberValue)
	if !ok {
		return nil, fmt.Errorf("index must be a number, got %s", idx.TypeName())
	}
	i64 := int(idxNum.Value)
	switch v := col.(type) {
	case *ListValue:
		if i64 < 0 || i64 >= len(v.Elements) {
			return nil, fmt.Errorf("list index %d out of range (len=%d)", i64, len(v.Elements))
		}
		return v.Elements[i64], nil
	case *StringValue:
		runes := []rune(v.Value)
		if i64 < 0 || i64 >= len(runes) {
			return nil, fmt.Errorf("string index %d out of range (len=%d)", i64, len(runes))
		}
		return &StringValue{Value: string(runes[i64])}, nil
	}
	return nil, fmt.Errorf("cannot index %s", col.TypeName())
}
