package designs

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	exprl "github.com/expr-lang/expr"
	"github.com/pkg/errors"
	"github.com/ungerik/go3d/float64/quaternion"
)

type Expr string

const ExprPrefix = "~ "

type ExprEnv struct {
	Inputs map[VarName]ExprEnvInput
}
type ExprEnvInput struct {
	Value any
	Type  string
	Units string
	Min   any
	Max   any
}

// Expression evaluation environments

func MakeExprEnv(
	inputValues map[VarName]any, inputSpecs map[VarName]InputVarSpec,
) (result ExprEnv, err error) {
	result.Inputs = make(map[VarName]ExprEnvInput)
	for name, spec := range inputSpecs {
		input := ExprEnvInput{
			Type:  string(spec.Type),
			Units: spec.Units,
			Min:   spec.Min,
			Max:   spec.Max,
		}
		var ok bool
		if input.Value, ok = inputValues[name]; !ok {
			input.Value = varTypeZeroValues[spec.Type]
		}
		if input.Value, err = convertTo(input.Value, spec.Type); err != nil {
			return result, errors.Wrapf(
				err, "couldn't ensure that %+v is of type %s", input.Value, spec.Type,
			)
		}
		result.Inputs[name] = input
	}

	return result, nil
}

func convertTo(value any, kind VarType) (result any, err error) {
	switch kind {
	case VarTypeBool:
		switch r := value.(type) {
		default:
			return result, errors.Errorf("can't convert %+v from %T to bool", r, r)
		case bool:
			result = r
		case string:
			return strconv.ParseBool(r)
		}
	case VarTypeInt:
		switch r := value.(type) {
		default:
			return result, errors.Errorf("can't convert %+v from %T to int", r, r)
		case int:
			result = r
		case string:
			const base = 10
			return strconv.ParseInt(r, base, strconv.IntSize)
		}
	case VarTypeFloat64:
		switch r := value.(type) {
		default:
			return result, errors.Errorf("can't convert %+v from %T to float64", r, r)
		case float64:
			result = r
		case int:
			result = float64(r)
		case string:
			const bitsize = 64
			return strconv.ParseFloat(r, bitsize)
		}
	case VarTypeString:
		if r, ok := value.(string); ok {
			result = r
			break
		}
		result = fmt.Sprintf("%s", value)
	case VarTypeQuaternion:
		return convertToQuaternion(value)
	}
	return result, nil
}

func convertToQuaternion(value any) (result quaternion.T, err error) {
	const quatLength = 4
	switch r := value.(type) {
	default:
		return result, errors.Errorf("can't convert %+v from %T to quaternion", r, r)
	case [quatLength]float64:
		return quaternion.T(r), nil
	case []float64:
		if l := len(r); l != quatLength {
			return result, errors.Errorf("can't convert %d-element slice into 4-element quaternion", l)
		}
		return quaternion.T(r), nil
	case [quatLength]any:
		for i, elem := range r {
			raw, err := convertTo(elem, VarTypeFloat64)
			if err != nil {
				return result, errors.Wrapf(
					err, "couldn't convert elem %d of %+v from %T to float64 for quaternion", i, r, elem,
				)
			}
			var ok bool
			if result[i], ok = raw.(float64); !ok {
				return result, errors.Wrapf(
					err, "couldn't convert elem %d of %+v from %T to float64 for quaternion", i, r, elem,
				)
			}
		}
		return result, nil
	case []any:
		if l := len(r); l != quatLength {
			return result, errors.Errorf("can't convert %d-element slice into 4-element quaternion", l)
		}
		for i, elem := range r {
			raw, err := convertTo(elem, VarTypeFloat64)
			if err != nil {
				return result, errors.Wrapf(
					err, "couldn't convert elem %d of %+v from %T to float64 for quaternion", i, r, elem,
				)
			}
			var ok bool
			if result[i], ok = raw.(float64); !ok {
				return result, errors.Wrapf(
					err, "couldn't convert elem %d of %+v from %T to float64 for quaternion", i, r, elem,
				)
			}
		}
		return result, nil
	case string:
		return quaternion.Parse(r)
	}
}

func (e ExprEnv) ToMap() map[string]any {
	inputs := make(map[string]map[string]any)
	for varName, input := range e.Inputs {
		inputs[string(varName)] = input.ToMap()
	}
	return map[string]any{
		"inputs": inputs,
	}
}

func (e ExprEnvInput) ToMap() map[string]any {
	return map[string]any{
		"value": e.Value,
		"type":  e.Type,
		"units": e.Units,
		"min":   e.Min,
		"max":   e.Max,
	}
}

// Expr

func (e Expr) evalAsAny(env any, options ...exprl.Option) (result any, err error) {
	options = append([]exprl.Option{}, options...)
	options = append(options, exprl.AsAny())

	raw, err := e.eval(env, options...)
	if err != nil {
		return raw, errors.Wrapf(err, "couldn't evaluate expression for generic result")
	}
	return raw, nil
}

// evalAsString returns the expression as a string literal if it doesn't contain the expected
// expression prefix `~ `; otherwise, it evaluates the expression.
func (e Expr) evalAsString[T ~string](env any, options ...exprl.Option) (result T, err error) {
	if s := string(e); !strings.HasPrefix(s, ExprPrefix) {
		return T(strings.TrimPrefix(s, ExprPrefix)), nil
	}

	options = append([]exprl.Option{}, options...)
	options = append(options, exprl.AsKind(reflect.String))

	raw, err := e.eval(env, options...)
	if err != nil {
		return result, errors.Wrapf(err, "couldn't evaluate expression for %T result", result)
	}
	asString, ok := raw.(string)
	if !ok {
		return result, errors.Errorf(
			"expression result %+v is a %T, but a string is needed for conversion to %T",
			raw, raw, result,
		)
	}
	return T(asString), nil
}

func (e Expr) evalAs[T any](env any, options ...exprl.Option) (result T, err error) {
	options = append([]exprl.Option{}, options...)
	options = append(options, exprl.AsKind(reflect.ValueOf(result).Kind()))

	raw, err := e.eval(env, options...)
	if err != nil {
		return result, errors.Wrapf(err, "couldn't evaluate expression for %T result", result)
	}
	result, ok := raw.(T)
	if !ok {
		return result, errors.Errorf(
			"expression result %+v is a %T, but a %T is needed", raw, raw, result,
		)
	}
	return result, nil
}

func (e Expr) eval(env any, options ...exprl.Option) (result any, err error) {
	program, err := exprl.Compile(
		strings.TrimPrefix(string(e), ExprPrefix),
		append([]exprl.Option{exprl.Env(env)}, options...)...,
	)
	if err != nil {
		return nil, errors.Wrapf(
			err, "couldn't compile expression %s with environment %+v", e, env,
		)
	}
	output, err := exprl.Run(program, env)
	if err != nil {
		return nil, errors.Wrapf(
			err, "couldn't run expression %s with environment %+v", e, env,
		)
	}
	return output, nil
}
