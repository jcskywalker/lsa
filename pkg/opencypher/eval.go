package opencypher

import (
	"errors"
	"strings"
)

const (
	c_const  uint = 0x001
	c_lvalue uint = 0x002
)

var (
	ErrDivideByZero                   = errors.New("Divide by zero")
	ErrInvalidUnaryOperation          = errors.New("Invalid unary operation")
	ErrInvalidPowerOperation          = errors.New("Invalid power operation")
	ErrInvalidMultiplicativeOperation = errors.New("Invalid multiplicative operation")
	ErrInvalidDurationOperation       = errors.New("Invalid duration operation")
	ErrOperationWithNull              = errors.New("Operation with null")
	ErrInvalidStringOperation         = errors.New("Invalid string operation")
	ErrInvalidDateOperation           = errors.New("Invalid date operation")
	ErrInvalidAdditiveOperation       = errors.New("Invalid additive operation")
	ErrInvalidComparison              = errors.New("Invalid comparison")
	ErrInvalidListIndex               = errors.New("Invalid list index")
	ErrNotAList                       = errors.New("Not a list")
	ErrNotABooleanExpression          = errors.New("Not a boolean expression")
	ErrMapKeyNotString                = errors.New("Map key is not a string")
	ErrInvalidMapKey                  = errors.New("Invalid map key")
	ErrNotAGraphObject                = errors.New("Not a graph object")
	ErrIntValueRequired               = errors.New("Int value required")
	ErrExpectingResultSet             = errors.New("Expecting a result set")
)

func (match Match) Evaluate(ctx *EvalContext) (Value, error) {

}

func (expr StringListNullOperatorExpression) Evaluate(ctx *EvalContext) (Value, error) {
	val, err := expr.PropertyOrLabels.Evaluate(ctx)
	if err != nil {
		return Value{}, err
	}
	for _, part := range expr.Parts {
		val, err = part.evaluate(ctx, val)
		if err != nil {
			return Value{}, err
		}
	}
	return val, nil
}

func (expr StringListNullOperatorExpressionPart) evaluate(ctx *EvalContext, inputValue Value) (Value, error) {
	switch {
	case expr.IsNull != nil:
		if *expr.IsNull {
			return Value{Value: inputValue.Value == nil}, nil
		}
		return Value{Value: inputValue.Value != nil}, nil

	case expr.ListIndex != nil:
		listValue, ok := inputValue.Value.([]Value)
		if !ok {
			if inputValue.Value != nil {
				return Value{}, ErrNotAList
			}
		}
		indexValue, err := expr.ListIndex.Evaluate(ctx)
		if err != nil {
			return Value{}, err
		}
		if indexValue.Value == nil {
			return Value{}, nil
		}
		intValue, ok := indexValue.Value.(int)
		if !ok {
			return Value{}, ErrInvalidListIndex
		}
		if listValue == nil {
			return Value{}, nil
		}
		if intValue >= 0 {
			if intValue >= len(listValue) {
				return Value{}, nil
			}
			return listValue[intValue], nil
		}
		index := len(listValue) + intValue
		if index < 0 {
			return Value{}, nil
		}
		return listValue[index], nil

	case expr.ListIn != nil:
		listValue, err := expr.ListIn.Evaluate(ctx)
		if err != nil {
			return Value{}, err
		}
		list, ok := listValue.Value.([]Value)
		if ok {
			if listValue.Value != nil {
				return Value{}, ErrNotAList
			}
		}
		if inputValue.Value == nil {
			return Value{}, nil
		}
		hasNull := false
		for _, elem := range list {
			if elem.Value == nil {
				hasNull = true
			} else {
				v, err := compareValues(inputValue.Value, elem.Value)
				if err != nil {
					return Value{}, err
				}
				if v == 0 {
					return Value{Value: true}, nil
				}
			}
		}
		if hasNull {
			return Value{}, nil
		}
		return Value{Value: false}, nil

	case expr.ListRange != nil:
		constant := inputValue.Constant
		listValue, ok := inputValue.Value.([]Value)
		if !ok {
			if inputValue.Value != nil {
				return Value{}, ErrNotAList
			}
		}
		from, err := expr.ListRange.First.Evaluate(ctx)
		if err != nil {
			return Value{}, err
		}
		if from.Value == nil {
			return Value{}, nil
		}
		if !from.Constant {
			constant = false
		}
		fromi, ok := from.Value.(int)
		if !ok {
			return Value{}, ErrInvalidListIndex
		}
		to, err := expr.ListRange.Second.Evaluate(ctx)
		if err != nil {
			return Value{}, err
		}
		if to.Value == nil {
			return Value{}, nil
		}
		if !to.Constant {
			constant = false
		}
		toi, ok := to.Value.(int)
		if !ok {
			return Value{}, ErrInvalidListIndex
		}
		if fromi < 0 || toi < 0 {
			return Value{}, ErrInvalidListIndex
		}
		if fromi >= len(listValue) {
			fromi = len(listValue) - 1
		}
		if toi >= len(listValue) {
			toi = len(listValue) - 1
		}
		if fromi > toi {
			fromi = toi
		}
		arr := make([]Value, 0, toi-fromi)
		for i := fromi; i < toi; i++ {
			if !listValue[i].Constant {
				constant = false
			}
			arr = append(arr, listValue[i])
		}
		return Value{Value: arr, Constant: constant}, nil
	}
	return expr.String.evaluate(ctx, inputValue)
}

func (expr StringOperatorExpression) evaluate(ctx *EvalContext, inputValue Value) (Value, error) {
	inputStrValue, ok := inputValue.Value.(string)
	if !ok {
		return Value{}, ErrInvalidStringOperation
	}
	exprValue, err := expr.Expr.Evaluate(ctx)
	if err != nil {
		return Value{}, err
	}
	strValue, ok := exprValue.Value.(string)
	if !ok {
		return Value{}, ErrInvalidStringOperation
	}
	if expr.Operator == "STARTS" {
		return Value{Value: strings.HasPrefix(inputStrValue, strValue)}, nil
	}
	if expr.Operator == "ENDS" {
		return Value{Value: strings.HasSuffix(inputStrValue, strValue)}, nil
	}
	return Value{Value: strings.Contains(inputStrValue, strValue)}, nil
}

func (pl PropertyOrLabelsExpression) Evaluate(ctx *EvalContext) (Value, error) {
	val, err := pl.Atom.Evaluate(ctx)
	if err != nil {
		return Value{}, err
	}
	if pl.NodeLabels != nil {
		gobj, ok := val.Value.(GraphObject)
		if !ok {
			return Value{}, ErrNotAGraphObject
		}
		for _, label := range *pl.NodeLabels {
			str := label.String()
			if !gobj.HasLabel(str) {
				gobj.Labels = append(gobj.Labels, str)
			}
		}
		val.Value = gobj
	}
	for range pl.PropertyLookup {
		panic("Unimplemented")
	}
	return val, nil
}

func (f *FunctionInvocation) Evaluate(ctx *EvalContext) (Value, error) {
	if f.function == nil {
		fn, err := ctx.GetFunction(f.Name)
		if err != nil {
			return Value{}, err
		}
		f.function = fn
	}
	args := f.args
	if args == nil {
		args = make([]Evaluatable, 0, len(f.Args))
		isConst := false

		for a := range f.Args {
			v, err := f.Args[a].Evaluate(ctx)
			if err != nil {
				return Value{}, err
			}
			if a == 0 {
				isConst = v.Constant
			} else if !v.Constant {
				isConst = false
			}
			args = append(args, v)
		}
		if isConst {
			f.args = args
		}
	}
	return f.function(ctx, args)
}

func (cs Case) Evaluate(ctx *EvalContext) (Value, error) {
	var testValue Value
	if cs.Test != nil {
		v, err := cs.Test.Evaluate(ctx)
		if err != nil {
			return Value{}, err
		}
		testValue = v
	}
	for _, alternative := range cs.Alternatives {
		when, err := alternative.When.Evaluate(ctx)
		if err != nil {
			return Value{}, err
		}
		if cs.Test != nil {
			result, err := compareValues(testValue, when)
			if err != nil {
				return Value{}, err
			}
			if result == 0 {
				return alternative.Then.Evaluate(ctx)
			}
		} else {
			boolValue, ok := when.Value.(bool)
			if !ok {
				return Value{}, ErrNotABooleanExpression
			}
			if boolValue {
				return alternative.Then.Evaluate(ctx)
			}
		}
	}
	if cs.Default != nil {
		return cs.Default.Evaluate(ctx)
	}
	return Value{}, nil
}

func (v Variable) Evaluate(ctx *EvalContext) (Value, error) {
	return ctx.GetVar(string(v))
}

// Evaluate a regular query, which is a single query with an optional
// union list
func (query RegularQuery) Evaluate(ctx *EvalContext) (Value, error) {
	result, err := query.SingleQuery.Evaluate(ctx)
	if err != nil {
		return Value{}, err
	}
	resultSet, ok := result.Value.(ResultSet)
	if !ok {
		return Value{}, ErrExpectingResultSet
	}
	for _, u := range query.Unions {
		newResult, err := u.SingleQuery.Evaluate(ctx)
		if err != nil {
			return Value{}, err
		}
		newResultSet, ok := newResult.Value.(ResultSet)
		if !ok {
			return Value{}, ErrExpectingResultSet
		}
		if err := resultSet.Union(newResultSet, u.All); err != nil {
			return Value{}, err
		}
	}
	return Value{Value: resultSet}, nil
}

func (query SinglePartQuery) Evaluate(ctx *EvalContext) (Value, error) {
	panic("Unimplemented")
}

func (unwind Unwind) Evaluate(ctx *EvalContext) (Value, error)            { panic("Unimplemented") }
func (pattern Pattern) Evaluate(ctx *EvalContext) (Value, error)          { panic("Unimplemented") }
func (ls ListComprehension) Evaluate(ctx *EvalContext) (Value, error)     { panic("Unimplemented") }
func (p PatternComprehension) Evaluate(ctx *EvalContext) (Value, error)   { panic("Unimplemented") }
func (flt FilterAtom) Evaluate(ctx *EvalContext) (Value, error)           { panic("Unimplemented") }
func (rel RelationshipsPattern) Evaluate(ctx *EvalContext) (Value, error) { panic("Unimplemented") }
func (cnt CountAtom) Evaluate(ctx *EvalContext) (Value, error)            { panic("Unimplemented") }
func (mq MultiPartQuery) Evaluate(ctx *EvalContext) (Value, error)        { panic("Unimplemented") }
