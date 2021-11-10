package opencypher

import (
	"math"
)

const (
	c_const  = 0x001
	c_rvalue = 0x002
)

type Value struct {
	Value interface{}
	Class uint
}

func (v Value) IsConst() bool {
	return (v.Class & c_const) != 0
}

type EvalContext struct{}

func (literal IntLiteral) Evaluate(ctx *EvalContext) (Value, error) {
	return Value{
		Value: int(literal),
		Class: c_const | c_rvalue,
	}, nil
}

func (literal BooleanLiteral) Evaluate(ctx *EvalContext) (Value, error) {
	return Value{
		Value: bool(literal),
		Class: c_const | c_rvalue,
	}, nil
}

func (literal DoubleLiteral) Evaluate(ctx *EvalContext) (Value, error) {
	return Value{
		Value: float64(literal),
		Class: c_const | c_rvalue,
	}, nil
}

func (literal StringLiteral) Evaluate(ctx *EvalContext) (Value, error) {
	return Value{
		Value: string(literal),
		Class: c_const | c_rvalue,
	}, nil
}

func (literal NullLiteral) Evaluate(ctx *EvalContext) (Value, error) {
	return Value{
		Class: c_const | c_rvalue,
	}, nil
}

func (expr *UnaryAddOrSubtractExpression) Evaluate(ctx *EvalContext) (Value, error) {
	if expr.constValue != nil {
		return *expr.constValue, nil
	}

	value, err := expr.Expr.Evaluate(ctx)
	if err != nil {
		return value, err
	}
	if expr.Neg {
		if intValue, ok := value.Value.(int); ok {
			value.value = -intValue
		} else if floatValue, ok := value.Value.(float64); ok {
			value.value = -floatValue
		} else {
			return value, ErrInvalidUnaryOperation
		}
	}
	if value.IsConst() {
		expr.constValue = &value
	}
	return value, nil
}

func (expr *PowerOfExpression) Evaluate(ctx *EvalContext) (Value, error) {
	if expr.constValue != nil {
		return *expr.constValue, nil
	}
	var ret Value
	var result float64
	for i := range expr.Parts {
		val, err := expr.Parts[i].Evaluate(ctx)
		if err != nil {
			return val, err
		}
		if i == 0 {
			ret = val
			if intValue, ok := ret.Value.(int); ok {
				result = float64(intValue)
			} else if floatValue, ok := ret.Value.(float64); ok {
				result = floatValue
			} else {
				return Value{}, ErrInvalidPowerOperation
			}
		} else {
			if intValue, ok := val.Value.(int); ok {
				result = math.Pow(result, float64(intValue))
			} else if floatValue, ok := val.Value.(float64); ok {
				result = math.Pow(result, floatValue)
			} else {
				return Value{}, ErrInvalidPowerOperation
			}
			ret.Class &= val.Class
		}
	}
	ret.Value = result
	if ret.IsConst() {
		expr.constValue = &ret
	}
	return ret, nil
}

func (expr *MultiplyDivideModuloExpression) Evaluate(ctx *EvalContext) (Value, error) {
	if expr.constValue != nil {
		return *expr.constValue, nil
	}

	var resultIsInt, resultIsFloat bool
	var intResult int
	var floatResult float64
	var class uint
	for i := range expr.Parts {
		val, err := expr.Parts[i].Expr.Evaluate(ctx)
		if err != nil {
			return val, err
		}
		if i == 0 {
			class := val.Class
			intResult, resultIsInt := val.Value.(int)
			floatResult, resultIsFloat := val.Value.(float64)
			if !resultIsInt && !resultIsFloat {
				return Value{}, ErrInvalidMultiplicativeOperation
			}
		} else {
			intValue, newValueIsInt := val.Value.(int)
			floatValue, newValueIsFloat := val.Value.(float64)
			if !newValueIsInt && !newValueIsFloat {
				return Value{}, ErrInvalidMultiplicativeOperation
			}
			class &= val.Class

			if resultIsInt && newValueIsFloat {
				floatResult = float64(intResult)
				resultIsFloat = true
				resultIsInt = false
			}
			if resultIsFloat && newValueIsInt {
				floatValue = float64(intValue)
				newValueIsFloat = true
				newValueIsInt = false
			}

			switch expr.Parts[i].Op {
			case '*':
				if resultIsInt {
					intResult *= intValue
				} else {
					floatResult *= floatValue
				}
			case '/':
				if resultIsInt {
					if intValue == 0 {
						return Value{}, ErrDivideByZero
					}
					intResult /= intValue
				} else {
					if floatValue == 0 {
						return Value{}, ErrDivideByZero
					}
					floatResult /= floatValue
				}
			case '%':
				if resultIsInt {
					if intValue == 0 {
						return Value{}, ErrDivideByZero
					}
					intResult %= intValue
				} else {
					if floatValue == 0 {
						return Value{}, ErrDivideByZero
					}
					floatResult = math.Mod(floatResult, floatValue)
				}
			}
		}
	}
	ret := Value{Class: class}
	if resultIsInt {
		ret.Value = intResult
	} else {
		ret.Value = floatResult
	}
	if ret.IsConst() {
		expr.constValue = &ret
	}
	return ret, nil
}
