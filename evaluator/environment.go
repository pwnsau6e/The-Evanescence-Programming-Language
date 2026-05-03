package evaluator

import "fmt"

// Environment maps names to values, supporting lexical scoping via a parent link.
// Each binding remembers whether it was declared as a constant.
type Environment struct {
	parent *Environment
	vars   map[string]binding
}

type binding struct {
	value   Value
	isConst bool
}

// NewEnv creates a top-level environment with no parent.
func NewEnv() *Environment {
	return &Environment{vars: map[string]binding{}}
}

// NewChild creates a nested scope whose parent is e.
func (e *Environment) NewChild() *Environment {
	return &Environment{parent: e, vars: map[string]binding{}}
}

// Declare introduces a new binding in the current scope.
// Re-declaring a name in the same scope is an error.
func (e *Environment) Declare(name string, v Value, isConst bool) error {
	if _, exists := e.vars[name]; exists {
		return fmt.Errorf("variable %q already declared in this scope", name)
	}
	e.vars[name] = binding{value: v, isConst: isConst}
	return nil
}

// Assign updates an existing binding, walking up the scope chain to find it.
// Returns an error if the name isn't bound or the binding is constant.
func (e *Environment) Assign(name string, v Value) error {
	for env := e; env != nil; env = env.parent {
		if b, ok := env.vars[name]; ok {
			if b.isConst {
				return fmt.Errorf("cannot assign to constant %q", name)
			}
			env.vars[name] = binding{value: v, isConst: false}
			return nil
		}
	}
	return fmt.Errorf("undefined variable %q", name)
}

// Get looks a name up through the scope chain.
func (e *Environment) Get(name string) (Value, bool) {
	for env := e; env != nil; env = env.parent {
		if b, ok := env.vars[name]; ok {
			return b.value, true
		}
	}
	return nil, false
}
