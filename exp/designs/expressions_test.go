package designs

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

var makeEnvTests = map[string]struct {
	inputs     map[VarName]any
	inputSpecs map[VarName]InputVarSpec
	env        any
	result     ExprEnv
}{
	"inputs-simple": {
		inputs: map[VarName]any{
			"offset": 7,
		},
		inputSpecs: map[VarName]InputVarSpec{
			"offset": {
				Kind:  "float64",
				Units: "mm",
				Min:   -5,
				Max:   5,
			},
		},
		result: ExprEnv{
			Inputs: map[VarName]ExprEnvInput{
				"offset": {
					Value: 7.0,
					Kind:  "float64",
					Units: "mm",
					Min:   -5,
					Max:   5,
				},
			},
		},
	},
}

func TestMakeEnv(t *testing.T) {
	for p, test := range makeEnvTests {
		t.Run(p, func(t *testing.T) {
			t.Logf("%s", p)
			result, err := MakeExprEnv(test.inputs, test.inputSpecs)
			if err != nil {
				t.Error(err)
				return
			}
			if got, want := result, test.result; !cmp.Equal(got, want) {
				t.Errorf("diff (-want +got):\n%+v", cmp.Diff(want, got))
			}
		})
	}
}

// TODO: test ensureKind

var evaluateAsAnyTests = map[string]struct {
	expression Expr
	env        ExprEnv
	result     any
}{
	"constant-int": {
		expression: "5 + 2",
		result:     7,
	},
	"constant-float64": {
		expression: "5 + 2.5",
		result:     7.5,
	},
	"inputs-simple-int": {
		expression: "inputs.offset.value + 2",
		env: ExprEnv{
			Inputs: map[VarName]ExprEnvInput{
				"offset": {Value: 7},
			},
		},
		result: 9,
	},
	"inputs-simple-float64": {
		expression: "inputs.offset.value + 2",
		env: ExprEnv{
			Inputs: map[VarName]ExprEnvInput{
				"offset": {Value: 7.0},
			},
		},
		result: 9.0,
	},
}

func TestEvaluateAsAny(t *testing.T) {
	for p, test := range evaluateAsAnyTests {
		t.Run(p, func(t *testing.T) {
			t.Logf("%s", p)
			result, err := test.expression.evalAsAny(test.env.ToMap())
			if err != nil {
				t.Error(err)
				return
			}
			if got, want := result, test.result; !cmp.Equal(got, want) {
				t.Errorf("diff (-want +got):\n%+v", cmp.Diff(want, got))
			}
		})
	}
}

var evaluateAsIntTests = map[string]struct {
	expression Expr
	env        ExprEnv
	result     int
}{
	"constant": {
		expression: "5 + 3",
		result:     8,
	},
	"inputs-simple-int": {
		expression: "inputs.offset.value + 2",
		env: ExprEnv{
			Inputs: map[VarName]ExprEnvInput{
				"offset": {Value: 7},
			},
		},
		result: 9,
	},
	"inputs-simple-float64": {
		expression: "inputs.offset.value + 2",
		env: ExprEnv{
			Inputs: map[VarName]ExprEnvInput{
				"offset": {Value: 7.0},
			},
		},
		result: 9.0,
	},
	"inputs-simple-float64-2": {
		expression: "inputs.offset.value + 2.0",
		env: ExprEnv{
			Inputs: map[VarName]ExprEnvInput{
				"offset": {Value: 7},
			},
		},
		result: 9,
	},
}

func TestEvaluateAsInt(t *testing.T) {
	for p, test := range evaluateAsIntTests {
		t.Run(p, func(t *testing.T) {
			t.Logf("%s", p)
			result, err := test.expression.evalAs[int](test.env.ToMap())
			if err != nil {
				t.Error(err)
				return
			}
			if got, want := result, test.result; !cmp.Equal(got, want) {
				t.Errorf("diff (-want +got):\n%+v", cmp.Diff(want, got))
			}
		})
	}
}

var evaluateAsFloat64Tests = map[string]struct {
	expression Expr
	env        ExprEnv
	result     float64
}{
	"constant-float64": {
		expression: "5 + 2.5",
		result:     7.5,
	},
	"inputs-simple-int": {
		expression: "inputs.offset.value + 2",
		env: ExprEnv{
			Inputs: map[VarName]ExprEnvInput{
				"offset": {Value: 7},
			},
		},
		result: 9,
	},
	"inputs-simple-float64": {
		expression: "inputs.offset.value + 2",
		env: ExprEnv{
			Inputs: map[VarName]ExprEnvInput{
				"offset": {Value: 7.0},
			},
		},
		result: 9.0,
	},
	"inputs-simple-float64-2": {
		expression: "inputs.offset.value + 2.0",
		env: ExprEnv{
			Inputs: map[VarName]ExprEnvInput{
				"offset": {Value: 7},
			},
		},
		result: 9.0,
	},
}

func TestEvaluateAsFloat(t *testing.T) {
	for p, test := range evaluateAsFloat64Tests {
		t.Run(p, func(t *testing.T) {
			t.Logf("%s", p)
			result, err := test.expression.evalAs[float64](test.env.ToMap())
			if err != nil {
				t.Error(err)
				return
			}
			if got, want := result, test.result; !cmp.Equal(got, want) {
				t.Errorf("diff (-want +got):\n%+v", cmp.Diff(want, got))
			}
		})
	}
}
