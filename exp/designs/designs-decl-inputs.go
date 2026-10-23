package designs

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/pkg/errors"
	"github.com/ungerik/go3d/float64/quaternion"
)

// InputsSpec

// Check looks for errors in the construction of the InputsSpec.
func (s InputsSpec) Check() (errs []error) {
	for varName, inputSpec := range s {
		inputErrs := inputSpec.Check()
		for _, err := range inputErrs {
			errs = append(errs, errors.Wrapf(err, "checks failed for input variable %s", varName))
		}
	}
	return errs
}

// Cloned returns a deep copy of the InputsSpec.
func (s InputsSpec) Cloned() InputsSpec {
	result := make(InputsSpec)
	for varName, spec := range s {
		result[varName] = spec.Cloned()
	}
	return result
}

// Merged returns a new InputsSpec created by applying the specified overlay, without modifying
// this current InputsSpec or the overlay.
func (s InputsSpec) Merged(overlay InputsSpec) InputsSpec {
	merged := s.Cloned()
	for id, o := range overlay {
		already, alreadyHas := merged[id]
		if !alreadyHas {
			merged[id] = o
			continue
		}

		merged[id] = already.Merged(o)
	}
	return merged
}

// Defaults returns a map of zero values for all input variables.
func (s InputsSpec) ZeroValues() InputValues {
	// TODO: if we want to allow making an input variable required, then this should only make zero
	// values for non-required input variables
	zeroes := make(InputValues)
	for id, spec := range s {
		zeroes[id] = varKindZeroValues[spec.Kind]
	}
	return zeroes
}

// InputVarSpec

// Check checks whether the InputVarSpec is well-formed.
func (s InputVarSpec) Check() (errs []error) {
	switch s.Kind {
	default:
		errs = append(errs, fmt.Errorf("unknown variable kind %s", s.Kind))
	case VarKindBool, VarKindInt, VarKindFloat64, VarKindString, VarKindQuaternion:
		break
	}

	switch s.Min.(type) {
	default:
		errs = append(errs, fmt.Errorf("unsupported type %T for min constraint", s.Min))
	case nil:
		break // i.e. unspecified min
	case int, int64, uint64, float64:
		if err := checkVarConstraintType(s.Kind, s.Min); err != nil {
			errs = append(errs, errors.Wrap(err, "type error for min constraint"))
		}
	}

	switch s.Max.(type) {
	default:
		errs = append(errs, fmt.Errorf("unsupported type %T for max constraint", s.Max))
	case nil:
		break // i.e. unspecified max
	case int, int64, uint64, float64:
		if err := checkVarConstraintType(s.Kind, s.Max); err != nil {
			errs = append(errs, errors.Wrap(err, "type error for max constraint"))
		}
	}

	return errs
}

func checkVarConstraintType(kind VarKind, constraint any) error {
	switch constraint.(type) {
	default:
		return fmt.Errorf("unsupported type %T for constraint %+v", constraint, constraint)
	case float64:
		if kind != VarKindFloat64 {
			return fmt.Errorf(varConstraintErrFmt, constraint, constraint, kind)
		}
	case int, int64, uint64:
		if kind != VarKindInt && kind != VarKindFloat64 {
			return fmt.Errorf(varConstraintErrFmt, constraint, constraint, kind)
		}
	}
	return nil
}

const varConstraintErrFmt = "constraint %+v of type %T not allowed for variable of kind %s"

// CheckValue checks the specified value against the InputSpec.
func (s InputVarSpec) CheckValue(value any) (errs []error) {
	switch v := value.(type) {
	default:
		errs = append(errs, fmt.Errorf(varValueErrFmt, value, value, s.Kind))
	case bool:
		if s.Kind != VarKindBool {
			errs = append(errs, fmt.Errorf(varValueErrFmt, value, value, s.Kind))
		}
		return errs
	case int:
		if s.Kind != VarKindInt {
			errs = append(errs, fmt.Errorf(varValueErrFmt, value, value, s.Kind))
			break
		}
		errs = append(errs, s.checkValueRange(v)...)
	case float64:
		if s.Kind != VarKindFloat64 {
			errs = append(errs, fmt.Errorf(varValueErrFmt, value, value, s.Kind))
			break
		}
		errs = append(errs, s.checkValueRange(v)...)
	case string:
		if s.Kind != VarKindString {
			errs = append(errs, fmt.Errorf(varValueErrFmt, value, value, s.Kind))
		}
	case quaternion.T:
		if s.Kind != VarKindQuaternion {
			errs = append(errs, fmt.Errorf(varValueErrFmt, value, value, s.Kind))
		}
	}
	return errs
}

const varValueErrFmt = "value %v of type %T not allowed for variable of kind %s"

func (s InputVarSpec) checkValueRange[T cmp.Ordered](value T) (errs []error) {
	min, err := convertTo(s.Min, s.Kind)
	if err != nil {
		errs = append(errs, errors.Wrapf(
			err, "couldn't convert min constraint %v of type %T to kind %s", s.Min, s.Min, s.Kind,
		))
	}
	max, err := convertTo(s.Max, s.Kind)
	if err != nil {
		errs = append(errs, errors.Wrapf(
			err, "couldn't convert max constraint %v of type %T to kind %s", s.Max, s.Max, s.Kind,
		))
	}
	typedMin, ok := min.(T)
	if !ok {
		errs = append(errs, errors.Errorf(
			"min constraint %v of type %T should instead be of type %T", min, min, value,
		))
	}
	typedMax, ok := max.(T)
	if !ok {
		errs = append(errs, errors.Errorf(
			"max constraint %v of type %T should instead be of type %T", max, max, value,
		))
	}
	if len(errs) > 0 {
		return errs
	}
	zero := varKindZeroValues[s.Kind]
	if typedMin == zero && typedMax == zero {
		return nil
	}

	if value < typedMin {
		errs = append(errs, errors.Errorf(
			"value %v is less than the minimum constraint %v", value, typedMin,
		))
	}
	if value > typedMax {
		errs = append(errs, errors.Errorf(
			"value %v is greater than the maximum constraint %v", value, typedMax,
		))
	}
	return errs
}

// Cloned returns a deep copy of the InputSpec.
func (s InputVarSpec) Cloned() InputVarSpec {
	return InputVarSpec{
		Description: s.Description,
		Kind:        s.Kind,
		Units:       s.Units,
		Min:         s.Min,
		Max:         s.Max,
		Tags:        slices.Clone(s.Tags),
	}
}

// Merged returns a new InputSpec created by applying the specified overlay, without modifying
// this current InputSpec or the overlay.
func (s InputVarSpec) Merged(overlay InputVarSpec) InputVarSpec {
	return InputVarSpec{
		Description: cmp.Or(overlay.Description, s.Description),
		Kind:        cmp.Or(overlay.Kind, s.Kind),
		Units:       cmp.Or(overlay.Units, s.Units),
		Min:         cmp.Or(overlay.Min, s.Min),
		Max:         cmp.Or(overlay.Max, s.Max),
		Tags:        s.Tags.Merged(overlay.Tags),
	}
}
