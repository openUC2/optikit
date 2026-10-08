package designs

import (
	"cmp"
	"slices"
)

// InputsSpec

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
	zeroes := make(InputValues)
	for id, spec := range s {
		zeroes[id] = varKindZeroValues[spec.Kind]
	}
	return zeroes
}

// InputVarSpec

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
