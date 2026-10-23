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

	if _, err := ConvertTo(s.Min, s.Kind); s.Min != nil && err != nil {
		errs = append(errs, errors.Wrap(err, "type error for min constraint"))
	}
	if _, err := ConvertTo(s.Max, s.Kind); s.Max != nil && err != nil {
		errs = append(errs, errors.Wrap(err, "type error for max constraint"))
	}

	if s.Kind == VarKindBool {
		// Note: the Enum field is ignored for variable kind bool, so we don't check enum values.
		return errs
	}

	for _, value := range s.Enum {
		if _, err := ConvertTo(value, s.Kind); err != nil {
			errs = append(errs, errors.Wrapf(
				err, "couldn't convert enum value %v of type %T to kind %s", value, value, s.Kind,
			))
		}
	}
	return errs
}

// CheckValue checks the specified value against the InputSpec. value is assumed to already exactly
// match the type associated with the InputVarSpec's kind, i.e. it's already been type-converted
// via the ConvertTo function (if necessary).
func (s InputVarSpec) CheckValue(value any) (errs []error) {
	switch v := value.(type) {
	default:
		errs = append(errs, fmt.Errorf(varValueErrFmt, value, value, s.Kind))
		return errs
	case bool:
		if s.Kind != VarKindBool {
			errs = append(errs, fmt.Errorf(varValueErrFmt, value, value, s.Kind))
		}
		return errs // we won't check enum values!
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

	if len(s.Enum) == 0 {
		return errs
	}
	for _, enumValue := range s.Enum {
		convertedEnumValue, err := ConvertTo(enumValue, s.Kind)
		if err != nil {
			errs = append(errs, fmt.Errorf("can't convert enum value %v to kind %s", enumValue, s.Kind))
			continue
		}

		if convertedEnumValue == value {
			return errs
		}
	}
	errs = append(errs, fmt.Errorf(
		"value %v doesn't match any allowed enum values (%+v)", value, s.Enum,
	))
	return errs
}

const varValueErrFmt = "value %v of type %T not allowed for variable of kind %s"

func (s InputVarSpec) checkValueRange[T cmp.Ordered](value T) (errs []error) {
	min, err := ConvertTo(s.Min, s.Kind)
	if s.Min != nil && err != nil {
		errs = append(errs, errors.Wrapf(
			err, "couldn't convert min constraint %v of type %T to kind %s", s.Min, s.Min, s.Kind,
		))
	}
	max, err := ConvertTo(s.Max, s.Kind)
	if s.Max != nil && err != nil {
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

	if s.Min != nil && value < typedMin {
		errs = append(errs, errors.Errorf(
			"value %v is less than the minimum constraint %v", value, typedMin,
		))
	}
	if s.Max != nil && value > typedMax {
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
