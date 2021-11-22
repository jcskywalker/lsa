package opencypher

import (
	"strings"
)

type Function func(*EvalContext, []Evaluatable) (Value, error)

type EvalContext struct {
	parent    *EvalContext
	funcMap   map[string]Function
	variables map[string]Value
}

func NewEvalContext() *EvalContext {
	return &EvalContext{funcMap: globalFuncs}
}

type ErrUnknownFunction struct {
	Name string
}

func (e ErrUnknownFunction) Error() string { return "Unknown function: " + e.Name }

type ErrUnknownVariable struct {
	Name string
}

func (e ErrUnknownVariable) Error() string { return "Unknown variable:" + e.Name }

func (ctx *EvalContext) getFunction(name string) (Function, error) {
	f := ctx.funcMap[name]
	if f == nil {
		return nil, ErrUnknownFunction{name}
	}
	return f, nil
}

func (ctx *EvalContext) GetFunction(name []SymbolicName) (Function, error) {
	bld := strings.Builder{}
	for i, x := range name {
		if i > 0 {
			bld.WriteRune('.')
		}
		bld.WriteString(string(x))
	}
	return ctx.getFunction(bld.String())
}

func (ctx *EvalContext) GetVar(name string) (Value, error) {
	val, ok := ctx.variables[name]
	if !ok {
		return Value{}, ErrUnknownVariable{Name: name}
	}
	val.Constant = false
	return val, nil
}
