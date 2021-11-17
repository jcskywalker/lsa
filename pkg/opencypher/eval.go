package opencypher

import (
	"errors"
	"math"
	"time"

	"github.com/neo4j/neo4j-go-driver/neo4j"
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
	ErrNotABooleanExpression          = errors.New("Not a boolean expression")
	ErrMapKeyNotString                = errors.New("Map key is not a string")
	ErrInvalidMapKey                  = errors.New("Invalid map key")
	ErrIntValueRequired               = errors.New("Int value required")
)

type Value struct {
	v interface{}

	variable string
	box      *Value

	Class uint
}

func (v *Value) Get(ctx *EvalContext) (interface{}, error) {
	if !v.IsLvalue() {
		return v.v, nil
	}
	if v.box == nil {
		var err error
		v.box, err = ctx.GetVar(v.variable)
		if err != nil {
			return nil, err
		}
	}
	return v.box.v, nil
}

func (v *Value) Set(value interface{}) {
	if v.box != nil {
		v.box.Set(value)
		return
	}
	v.v = value
}

func (v *Value) Lvalue(lvalue bool) {
	if lvalue {
		v.Class |= c_lvalue
	} else {
		v.Class &= ^c_lvalue
	}
}

func (v Value) Evaluate(ctx *EvalContext) (Value, error) { return v, nil }

func (v Value) IsConst() bool {
	return (v.Class & c_const) != 0
}

func (v Value) IsLvalue() bool {
	return (v.Class & c_lvalue) != 0
}

func (literal IntLiteral) Evaluate(ctx *EvalContext) (Value, error) {
	return Value{
		v:     int(literal),
		Class: c_const,
	}, nil
}

func (literal BooleanLiteral) Evaluate(ctx *EvalContext) (Value, error) {
	return Value{
		v:     bool(literal),
		Class: c_const,
	}, nil
}

func (literal DoubleLiteral) Evaluate(ctx *EvalContext) (Value, error) {
	return Value{
		v:     float64(literal),
		Class: c_const,
	}, nil
}

func (literal StringLiteral) Evaluate(ctx *EvalContext) (Value, error) {
	return Value{
		v:     string(literal),
		Class: c_const,
	}, nil
}

func (literal NullLiteral) Evaluate(ctx *EvalContext) (Value, error) {
	return Value{
		Class: c_const,
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
		val, err := value.Get(ctx)
		if err != nil {
			return Value{}, err
		}
		if intValue, ok := val.(int); ok {
			value.Set(-intValue)
		} else if floatValue, ok := val.(float64); ok {
			value.Set(-floatValue)
		} else {
			return value, ErrInvalidUnaryOperation
		}
		value.Lvalue(false)
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
	for i := range expr.Parts {
		val, err := expr.Parts[i].Evaluate(ctx)
		if err != nil {
			return val, err
		}
		if i == 0 {
			ret = val
		} else {
			var valValue float64
			if intValue, ok := val.Value.(int); ok {
				valValue = float64(intValue)
			} else if floatValue, ok := val.Value.(float64); ok {
				valValue = floatValue
			} else {
				return Value{}, ErrInvalidPowerOperation
			}
			if i, ok := ret.Value.(int); ok {
				ret.Value = math.Pow(float64(i), valValue)
			} else if f, ok := ret.Value.(float64); ok {
				ret.Value = math.Pow(f, valValue)
			} else {
				return Value{}, ErrInvalidPowerOperation
			}
			ret.Class &= val.Class
		}
	}
	if ret.IsConst() {
		expr.constValue = &ret
	}
	return ret, nil
}

func mulintint(a, b int, op rune) (int, error) {
	switch op {
	case '*':
		return a * b, nil
	case '/':
		if b == 0 {
			return 0, ErrDivideByZero
		}
		return a / b, nil
	}
	if b == 0 {
		return 0, ErrDivideByZero
	}
	return a % b, nil
}

func mulintfloat(a int, b float64, op rune) (float64, error) {
	switch op {
	case '*':
		return float64(a) * b, nil
	case '/':
		if b == 0 {
			return 0, ErrDivideByZero
		}
		return float64(a) / b, nil
	}
	if b == 0 {
		return 0, ErrDivideByZero
	}
	return math.Mod(float64(a), b), nil
}

func mulfloatint(a float64, b int, op rune) (float64, error) {
	switch op {
	case '*':
		return a * float64(b), nil
	case '/':
		if b == 0 {
			return 0, ErrDivideByZero
		}
		return a / float64(b), nil
	}
	if b == 0 {
		return 0, ErrDivideByZero
	}
	return math.Mod(a, float64(b)), nil
}

func mulfloatfloat(a, b float64, op rune) (float64, error) {
	switch op {
	case '*':
		return a * b, nil
	case '/':
		if b == 0 {
			return 0, ErrDivideByZero
		}
		return a / b, nil
	}
	if b == 0 {
		return 0, ErrDivideByZero
	}
	return math.Mod(a, b), nil
}

func muldurint(a neo4j.Duration, b int64, op rune) (neo4j.Duration, error) {
	switch op {
	case '*':
		return neo4j.DurationOf(a.Months()*b, a.Days()*b, a.Seconds()*b, a.Nanos()*int(b)), nil
	case '/':
		if b == 0 {
			return neo4j.Duration{}, ErrDivideByZero
		}
		return neo4j.DurationOf(a.Months()/b, a.Days()/b, a.Seconds()/b, a.Nanos()/int(b)), nil
	}
	return neo4j.Duration{}, ErrInvalidDurationOperation
}

func mulintdur(a int64, b neo4j.Duration, op rune) (neo4j.Duration, error) {
	switch op {
	case '*':
		return neo4j.DurationOf(b.Months()*a, b.Days()*a, b.Seconds()*a, b.Nanos()*int(a)), nil
	default:
		return neo4j.Duration{}, ErrInvalidDurationOperation
	}
}

func muldurfloat(a neo4j.Duration, b float64, op rune) (neo4j.Duration, error) {
	val := int64(b)
	switch op {
	case '*':
		return neo4j.DurationOf(int64(a.Months()*val), int64(a.Days()*val), int64(a.Seconds()*val), a.Nanos()*int(val)), nil
	case '/':
		if b == 0 {
			return neo4j.Duration{}, ErrDivideByZero
		}
		return neo4j.DurationOf(int64(a.Months()/val), int64(a.Days()/val), int64(a.Seconds()/val), a.Nanos()/int(val)), nil
	}
	return neo4j.Duration{}, ErrInvalidDurationOperation
}

func mulfloatdur(a float64, b neo4j.Duration, op rune) (neo4j.Duration, error) {
	val := int64(a)
	switch op {
	case '*':
		return neo4j.DurationOf(b.Months()*val, b.Days()*val, b.Seconds()*val, b.Nanos()*int(val)), nil
	default:
		return neo4j.Duration{}, ErrInvalidDurationOperation
	}
}

func (expr *MultiplyDivideModuloExpression) Evaluate(ctx *EvalContext) (Value, error) {
	if expr.constValue != nil {
		return *expr.constValue, nil
	}

	var ret Value
	var err error
	for i := range expr.Parts {
		var val Value
		val, err = expr.Parts[i].Expr.Evaluate(ctx)
		if err != nil {
			return val, err
		}
		if i == 0 {
			ret = val
		} else {
			if ret.Value == nil {
				return Value{}, ErrOperationWithNull
			}
			ret.Class &= val.Class
			switch result := ret.Value.(type) {
			case int:
				switch operand := val.Value.(type) {
				case int:
					ret.Value, err = mulintint(result, operand, expr.Parts[i].Op)
				case float64:
					ret.Value, err = mulintfloat(result, operand, expr.Parts[i].Op)
				case neo4j.Duration:
					ret.Value, err = mulintdur(int64(result), operand, expr.Parts[i].Op)
				default:
					err = ErrInvalidMultiplicativeOperation
				}
			case float64:
				switch operand := val.Value.(type) {
				case int:
					ret.Value, err = mulfloatint(result, operand, expr.Parts[i].Op)
				case float64:
					ret.Value, err = mulfloatfloat(result, operand, expr.Parts[i].Op)
				case neo4j.Duration:
					ret.Value, err = mulfloatdur(result, operand, expr.Parts[i].Op)
				default:
					err = ErrInvalidMultiplicativeOperation
				}
			case neo4j.Duration:
				switch operand := val.Value.(type) {
				case int:
					ret.Value, err = muldurint(result, int64(operand), expr.Parts[i].Op)
				case float64:
					ret.Value, err = muldurfloat(result, operand, expr.Parts[i].Op)
				default:
					err = ErrInvalidMultiplicativeOperation
				}
			default:
				err = ErrInvalidMultiplicativeOperation
			}
		}
	}
	if err != nil {
		return Value{}, err
	}
	if ret.IsConst() {
		expr.constValue = &ret
	}
	return ret, nil
}

func addintint(a int, b int, sub bool) int {
	if sub {
		return a - b
	}
	return a + b
}

func addintfloat(a int, b float64, sub bool) float64 {
	if sub {
		return float64(a) - b
	}
	return float64(a) + b
}

func addfloatint(a float64, b int, sub bool) float64 {
	if sub {
		return a - float64(b)
	}
	return a + float64(b)
}

func addfloatfloat(a float64, b float64, sub bool) float64 {
	if sub {
		return a - b
	}
	return a + b
}

func addstringstring(a string, b string, sub bool) (string, error) {
	if sub {
		return "", ErrInvalidStringOperation
	}
	return a + b, nil
}

func adddatedur(a neo4j.Date, b neo4j.Duration, sub bool) neo4j.Date {
	t := a.Time()
	if sub {
		return neo4j.DateOf(time.Date(t.Year(), t.Month()-time.Month(b.Months()), t.Day()-int(b.Days()), 0, 0, 0, 0, t.Location()))
	}
	return neo4j.DateOf(time.Date(t.Year(), t.Month()+time.Month(b.Months()), t.Day()+int(b.Days()), 0, 0, 0, 0, t.Location()))
}

func addtimedur(a neo4j.LocalTime, b neo4j.Duration, sub bool) neo4j.LocalTime {
	t := a.Time()
	if sub {
		return neo4j.LocalTimeOf(time.Date(1970, 1, 1, t.Hour(), t.Minute(), t.Second()-int(b.Seconds()), t.Nanosecond()-b.Nanos(), t.Location()))
	}
	return neo4j.LocalTimeOf(time.Date(1970, 1, 1, t.Hour(), t.Minute(), t.Second()+int(b.Seconds()), t.Nanosecond()+b.Nanos(), t.Location()))
}

func adddatetimedur(a neo4j.LocalDateTime, b neo4j.Duration, sub bool) neo4j.LocalDateTime {
	t := a.Time()
	if sub {
		return neo4j.LocalDateTimeOf(time.Date(t.Year(), t.Month()-time.Month(b.Months()), t.Day()-int(b.Days()), t.Hour(), t.Minute(), t.Second()-int(b.Seconds()), t.Nanosecond()-b.Nanos(), t.Location()))
	}
	return neo4j.LocalDateTimeOf(time.Date(t.Year(), t.Month()+time.Month(b.Months()), t.Day()+int(b.Days()), t.Hour(), t.Minute(), t.Second()+int(b.Seconds()), t.Nanosecond()+b.Nanos(), t.Location()))
}

func adddurdate(a neo4j.Duration, b neo4j.Date, sub bool) (neo4j.Date, error) {
	if sub {
		return neo4j.Date{}, ErrInvalidDateOperation
	}
	return adddatedur(b, a, false), nil
}

func adddurtime(a neo4j.Duration, b neo4j.LocalTime, sub bool) (neo4j.LocalTime, error) {
	if sub {
		return neo4j.LocalTime{}, ErrInvalidDateOperation
	}
	return addtimedur(b, a, false), nil
}

func adddurdatetime(a neo4j.Duration, b neo4j.LocalDateTime, sub bool) (neo4j.LocalDateTime, error) {
	if sub {
		return neo4j.LocalDateTime{}, ErrInvalidDateOperation
	}
	return adddatetimedur(b, a, false), nil
}

func adddurdur(a neo4j.Duration, b neo4j.Duration, sub bool) (neo4j.Duration, error) {
	if sub {
		return neo4j.DurationOf(a.Months()-b.Months(), a.Days()-b.Days(), a.Seconds()-b.Seconds(), a.Nanos()-b.Nanos()), nil
	}
	return neo4j.DurationOf(a.Months()+b.Months(), a.Days()+b.Days(), a.Seconds()+b.Seconds(), a.Nanos()+b.Nanos()), nil
}

func (expr *AddOrSubtractExpression) Evaluate(ctx *EvalContext) (Value, error) {
	if expr.constValue != nil {
		return *expr.constValue, nil
	}

	var ret Value
	first := true

	accumulate := func(operand Value, sub bool) error {
		if first {
			first = false
			ret = operand
			return nil
		}
		ret.Class &= operand.Class
		var err error
		switch retValue := ret.Value.(type) {
		case int:
			switch operandValue := operand.Value.(type) {
			case int:
				ret.Value = addintint(retValue, operandValue, sub)
			case float64:
				ret.Value = addintfloat(retValue, operandValue, sub)
			default:
				err = ErrInvalidAdditiveOperation
			}
		case float64:
			switch operandValue := operand.Value.(type) {
			case int:
				ret.Value = addfloatint(retValue, operandValue, sub)
			case float64:
				ret.Value = addfloatfloat(retValue, operandValue, sub)
			default:
				err = ErrInvalidAdditiveOperation
			}
		case string:
			switch operandValue := operand.Value.(type) {
			case string:
				ret.Value, err = addstringstring(retValue, operandValue, sub)
			default:
				err = ErrInvalidAdditiveOperation
			}
		case neo4j.Duration:
			switch operandValue := operand.Value.(type) {
			case neo4j.Duration:
				ret.Value, err = adddurdur(retValue, operandValue, sub)
			case neo4j.Date:
				ret.Value, err = adddurdate(retValue, operandValue, sub)
			case neo4j.LocalTime:
				ret.Value, err = adddurtime(retValue, operandValue, sub)
			case neo4j.LocalDateTime:
				ret.Value, err = adddurdatetime(retValue, operandValue, sub)
			default:
				err = ErrInvalidAdditiveOperation
			}
		case neo4j.Date:
			switch operandValue := operand.Value.(type) {
			case neo4j.Duration:
				ret.Value = adddatedur(retValue, operandValue, sub)
			default:
				err = ErrInvalidAdditiveOperation
			}
		case neo4j.LocalTime:
			switch operandValue := operand.Value.(type) {
			case neo4j.Duration:
				ret.Value = addtimedur(retValue, operandValue, sub)
			default:
				err = ErrInvalidAdditiveOperation
			}
		case neo4j.LocalDateTime:
			switch operandValue := operand.Value.(type) {
			case neo4j.Duration:
				ret.Value = adddatetimedur(retValue, operandValue, sub)
			default:
				err = ErrInvalidAdditiveOperation
			}
		}
		return err
	}

	for i := range expr.Add {
		val, err := expr.Add[i].Evaluate(ctx)
		if err != nil {
			return Value{}, err
		}
		if err = accumulate(val, false); err != nil {
			return Value{}, err
		}
	}
	for i := range expr.Sub {
		val, err := expr.Add[i].Evaluate(ctx)
		if err != nil {
			return Value{}, err
		}
		if err = accumulate(val, true); err != nil {
			return Value{}, err
		}
	}
	if ret.IsConst() {
		expr.constValue = &ret
	}
	return ret, nil
}

func compareValues(v1, v2 interface{}) (int, error) {
	if v1 == nil || v2 == nil {
		return 0, ErrOperationWithNull
	}
	switch value1 := v1.(type) {
	case bool:
		switch value2 := v2.(type) {
		case bool:
			if value1 == value2 {
				return 0, nil
			}
			if value1 {
				return 1, nil
			}
			return -1, nil
		}
	case int:
		switch value2 := v2.(type) {
		case int:
			return value1 - value2, nil
		case float64:
			if float64(value1) == value2 {
				return 0, nil
			}
			if float64(value1) < value2 {
				return -1, nil
			}
			return 1, nil
		}
	case float64:
		switch value2 := v2.(type) {
		case int:
			if value1 == float64(value2) {
				return 0, nil
			}
			if value1 < float64(value2) {
				return -1, nil
			}
			return 1, nil
		case float64:
			if value1 == value2 {
				return 0, nil
			}
			if value1 < value2 {
				return -1, nil
			}
			return 1, nil
		}
	case string:
		if str, ok := v2.(string); ok {
			if value1 == str {
				return 0, nil
			}
			if value1 < str {
				return -1, nil
			}
			return 1, nil
		}
	case neo4j.Duration:
		if dur, ok := v2.(neo4j.Duration); ok {
			if value1.Days() == dur.Days() && value1.Months() == dur.Months() && value1.Seconds() == dur.Seconds() && value1.Nanos() == dur.Nanos() {
				return 0, nil
			}
			if value1.Days() < dur.Days() {
				return -1, nil
			}
			if value1.Months() < dur.Months() {
				return -1, nil
			}
			if value1.Seconds() < dur.Seconds() {
				return -1, nil
			}
			if value1.Nanos() < dur.Nanos() {
				return -1, nil
			}
			return 1, nil
		}
	case neo4j.Date:
		if date, ok := v2.(neo4j.Date); ok {
			t1 := value1.Time()
			t2 := date.Time()
			if t1.Equal(t2) {
				return 0, nil
			}
			if t1.Before(t2) {
				return -1, nil
			}
			return 0, nil
		}
	case neo4j.LocalTime:
		if date, ok := v2.(neo4j.LocalTime); ok {
			t1 := value1.Time()
			t2 := date.Time()
			if t1.Equal(t2) {
				return 0, nil
			}
			if t1.Before(t2) {
				return -1, nil
			}
			return 0, nil
		}
	case neo4j.LocalDateTime:
		if date, ok := v2.(neo4j.LocalDateTime); ok {
			t1 := value1.Time()
			t2 := date.Time()
			if t1.Equal(t2) {
				return 0, nil
			}
			if t1.Before(t2) {
				return -1, nil
			}
			return 0, nil
		}
	}
	return 0, ErrInvalidComparison

}

func (expr ComparisonExpression) Evaluate(ctx *EvalContext) (Value, error) {
	val, err := expr.First.Evaluate(ctx)
	if err != nil {
		return Value{}, err
	}

	for i := range expr.Second {
		second, err := expr.Second[i].Expr.Evaluate(ctx)
		if err != nil {
			return Value{}, err
		}
		result, err := compareValues(val.Value, second.Value)
		if err != nil {
			return Value{}, err
		}
		switch expr.Second[i].Op {
		case "=":
			val.Value = result == 0
		case "<>":
			val.Value = result != 0
		case "<":
			val.Value = result < 0
		case "<=":
			val.Value = result <= 0
		case ">":
			val.Value = result > 0
		case ">=":
			val.Value = result >= 0
		}
		val.Class &= second.Class
	}
	return val, nil
}

func (expr NotExpression) Evaluate(ctx *EvalContext) (Value, error) {
	val, err := expr.Part.Evaluate(ctx)
	if err != nil {
		return Value{}, err
	}
	value, ok := val.Value.(bool)
	if !ok {
		return Value{}, ErrNotABooleanExpression
	}
	val.Value = !value
	return val, nil
}

func (expr AndExpression) Evaluate(ctx *EvalContext) (Value, error) {
	var ret Value
	for i := range expr.Parts {
		val, err := expr.Parts[i].Evaluate(ctx)
		if err != nil {
			return Value{}, err
		}
		if i == 0 {
			ret = val
		} else {
			bval, ok := ret.Value.(bool)
			if !ok {
				return Value{}, ErrNotABooleanExpression
			}
			vval, ok := val.Value.(bool)
			if !ok {
				return Value{}, ErrNotABooleanExpression
			}
			ret.Class &= val.Class
			ret.Value = bval && vval
			if !bval || !vval {
				break
			}
		}
	}
	return ret, nil
}

func (expr XorExpression) Evaluate(ctx *EvalContext) (Value, error) {
	var ret Value
	for i := range expr.Parts {
		val, err := expr.Parts[i].Evaluate(ctx)
		if err != nil {
			return Value{}, err
		}
		if i == 0 {
			ret = val
		} else {
			bval, ok := ret.Value.(bool)
			if !ok {
				return Value{}, ErrNotABooleanExpression
			}
			vval, ok := val.Value.(bool)
			if !ok {
				return Value{}, ErrNotABooleanExpression
			}
			ret.Class &= val.Class
			ret.Value = bval != vval
		}
	}
	return ret, nil
}

func (expr OrExpression) Evaluate(ctx *EvalContext) (Value, error) {
	var ret Value
	for i := range expr.Parts {
		val, err := expr.Parts[i].Evaluate(ctx)
		if err != nil {
			return Value{}, err
		}
		if i == 0 {
			ret = val
		} else {
			bval, ok := ret.Value.(bool)
			if !ok {
				return Value{}, ErrNotABooleanExpression
			}
			vval, ok := val.Value.(bool)
			if !ok {
				return Value{}, ErrNotABooleanExpression
			}
			ret.Class &= val.Class
			ret.Value = bval || vval
			if bval || vval {
				break
			}
		}
	}
	return ret, nil
}

func (lst *ListLiteral) Evaluate(ctx *EvalContext) (Value, error) {
	if lst.constValue != nil {
		return *lst.constValue, nil
	}
	ret := make([]Value, 0, len(lst.Values))
	var val Value
	for i := range lst.Values {
		v, err := lst.Values[i].Evaluate(ctx)
		if err != nil {
			return Value{}, err
		}
		if i == 0 {
			val.Class = v.Class
		} else {
			val.Class &= v.Class
		}
		ret = append(ret, v)
	}
	val.Value = ret
	if val.IsConst() {
		lst.constValue = &val
	}
	return val, nil
}

func (mp *MapLiteral) Evaluate(ctx *EvalContext) (Value, error) {
	if mp.constValue != nil {
		return *mp.constValue, nil
	}
	var val Value
	ret := make(map[string]Value)
	for i := range mp.KeyValues {
		if mp.KeyValues[i].Key.SymbolicName == nil {
			return Value{}, ErrInvalidMapKey
		}
		keyStr := string(*mp.KeyValues[i].Key.SymbolicName)
		value, err := mp.KeyValues[i].Value.Evaluate(ctx)
		if err != nil {
			return Value{}, err
		}
		ret[keyStr] = value
		if i == 0 {
			val.Class = value.Class
		} else {
			val.Class &= value.Class
		}
	}
	val.Value = ret
	if val.IsConst() {
		mp.constValue = &val
	}
	return val, nil
}

func (expr StringListNullOperatorExpression) Evaluate(ctx *EvalContext) (Value, error) {
	val, err := expr.PropertyOrLabels.Evaluate(ctx)
	if err != nil {
		return Value{}, err
	}
	for range expr.Parts {
		panic("unimplemented")
	}
	return val, nil
}

func (pl PropertyOrLabelsExpression) Evaluate(ctx *EvalContext) (Value, error) {
	val, err := pl.Atom.Evaluate(ctx)
	if err != nil {
		return Value{}, err
	}
	for range pl.PropertyLookup {
		panic("Unimplemented")
	}
	if pl.NodeLabels != nil {
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
				isConst = v.IsConst()
			} else if !v.IsConst() {
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

func (query RegularQuery) Evaluate(ctx *EvalContext) (Value, error)       { panic("Unimplemented") }
func (query SinglePartQuery) Evaluate(ctx *EvalContext) (Value, error)    { panic("Unimplemented") }
func (match Match) Evaluate(ctx *EvalContext) (Value, error)              { panic("Unimplemented") }
func (unwind Unwind) Evaluate(ctx *EvalContext) (Value, error)            { panic("Unimplemented") }
func (pattern Pattern) Evaluate(ctx *EvalContext) (Value, error)          { panic("Unimplemented") }
func (ls ListComprehension) Evaluate(ctx *EvalContext) (Value, error)     { panic("Unimplemented") }
func (p PatternComprehension) Evaluate(ctx *EvalContext) (Value, error)   { panic("Unimplemented") }
func (flt FilterAtom) Evaluate(ctx *EvalContext) (Value, error)           { panic("Unimplemented") }
func (rel RelationshipsPattern) Evaluate(ctx *EvalContext) (Value, error) { panic("Unimplemented") }
func (v Variable) Evaluate(ctx *EvalContext) (Value, error)               { panic("Unimplemented") }
func (cnt CountAtom) Evaluate(ctx *EvalContext) (Value, error)            { panic("Unimplemented") }
