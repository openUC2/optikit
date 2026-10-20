package designs

import (
	"cmp"
	"context"
	"encoding/json/v2"
	"fmt"
	"maps"
	"path"
	"slices"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/goccy/go-yaml"
	"github.com/pkg/errors"
)

// CompID

// JoinCompIDs concatenates component IDs with slash ("/") delimiters into a path-style name.
func JoinCompIDs(elem ...CompID) CompID {
	elems := make([]string, 0, len(elem))
	for _, e := range elem {
		elems = append(elems, string(e))
	}
	return CompID(path.Join(elems...))
}

// CompExprsSpec

// Check looks for errors in the construction of the components spec.
func (s CompExprsSpec) Check() (errs []error) {
	for range s {
		// TODO: check for validity of instantiation...or maybe we must do this in FSDesign
	}
	return errs
}

// Cloned returns a deep copy of the CompExprsSpec.
func (s CompExprsSpec) Cloned() CompExprsSpec {
	result := make(CompExprsSpec)
	for compID, spec := range s {
		result[compID] = spec.Cloned()
	}
	return result
}

// Merged returns a new CompExprsSpec created by applying the specified overlay, without modifying
// this current CompExprsSpec or the overlay.
/*func (s CompExprsSpec) Merged(overlay CompExprsSpec) CompExprsSpec {
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
}*/

// Evaluated evaluates the parameter expressions with the given ExprEnv into a CompsSpec.
func (s CompExprsSpec) Evaluated(env ExprEnv) (result CompsSpec, err error) {
	result = make(CompsSpec)
	for id, c := range s {
		if result[id], err = c.Evaluated(env); err != nil {
			return nil, errors.Wrapf(err, "couldn't evaluate expressions in component %s", id)
		}
	}
	return result, nil
}

// CompsSpec

// Check looks for errors in the construction of the components spec.
func (s CompsSpec) Check() (errs []error) {
	for range s {
		// TODO: check for validity of instantiation...or maybe we must do this in FSDesign
	}
	return errs
}

// Cloned returns a deep copy of the CompsSpec.
func (s CompsSpec) Cloned() CompsSpec {
	result := make(CompsSpec)
	for compID, spec := range s {
		result[compID] = spec.Cloned()
	}
	return result
}

// CompExprSpec

// Cloned returns a deep copy of the CompExprSpec.
func (s CompExprSpec) Cloned() CompExprSpec {
	return CompExprSpec{
		Kind:          s.Kind,
		Design:        s.Design,
		Instantiation: s.Instantiation.Cloned(),
		Primitive:     s.Primitive.Cloned(),
		Results:       maps.Clone(s.Results),
		Tags:          slices.Clone(s.Tags),
	}
}

// Merged returns a new CompExprSpec created by applying the specified overlay, without modifying
// this current CompExprSpec or the overlay.
/*func (s CompExprSpec) Merged(overlay CompExprSpec) CompExprSpec {
	resultsMerged := maps.Clone(s.Results)
	for key, value := range overlay.Results {
		if value == "" {
			continue
		}
		resultsMerged[key] = value
	}
	return CompExprSpec{
		Kind:          cmp.Or(overlay.Kind, s.Kind),
		Design:        cmp.Or(overlay.Design, s.Design),
		Instantiation: s.Instantiation.Merged(overlay.Instantiation),
		Primitive:     s.Primitive.Merged(overlay.Primitive),
		Results:       resultsMerged,
		Tags:          s.Tags.Merged(overlay.Tags),
	}
}*/

// Evaluated evaluates the expressions with the given ExprEnv into a CompSpec.
func (s CompExprSpec) Evaluated(env ExprEnv) (result CompSpec, err error) {
	result = CompSpec{
		Kind:      s.Kind,
		Design:    s.Design,
		Primitive: s.Primitive,
		Results:   make(map[string]any),
		Tags:      s.Tags,
	}
	if result.Instantiation, err = s.Instantiation.Evaluated(env); err != nil {
		return result, errors.Wrap(err, "couldn't evaluate expressions in instantiation section")
	}
	for exprName, expr := range s.Results {
		if expr == "" {
			continue
		}

		value, err := expr.evalAsAny(env.ToMap())
		if err != nil {
			return CompSpec{}, errors.Wrapf(err, "couldn't evaluate result %s as %s", exprName, expr)
		}
		result.Results[exprName] = value
	}
	return result, nil
}

// CompSpec

// Cloned returns a deep copy of the CompSpec.
func (s CompSpec) Cloned() CompSpec {
	return CompSpec{
		Kind:          s.Kind,
		Design:        s.Design,
		Instantiation: s.Instantiation.Cloned(),
		Primitive:     s.Primitive.Cloned(),
		Results:       maps.Clone(s.Results),
		Tags:          slices.Clone(s.Tags),
	}
}

// InstExprSpec

// Cloned returns a deep copy of the InstExprSpec.
func (s InstExprSpec) Cloned() InstExprSpec {
	return InstExprSpec{
		Inputs: maps.Clone(s.Inputs),
	}
}

// Merged returns a new InstExprSpec created by applying the specified overlay, without modifying
// this current InstExprSpec or the overlay.
/*func (s InstExprSpec) Merged(overlay InstExprSpec) InstExprSpec {
	merged := InstExprSpec{}
	mergedInputs := maps.Clone(s.Inputs)
	for name, o := range overlay.Inputs {
		already, alreadyHas := mergedInputs[name]
		if !alreadyHas {
			mergedInputs[name] = o
			continue
		}

		mergedInputs[name] = cmp.Or(o, already)
	}
	merged.Inputs = mergedInputs
	return merged
}*/

// Evaluated evaluates the expressions with the given ExprEnv into a CompSpec.
func (s InstExprSpec) Evaluated(env ExprEnv) (result InstSpec, err error) {
	result.Inputs = make(InputValues)
	for varName, expr := range s.Inputs {
		if expr == "" {
			continue
		}

		value, err := expr.evalAsAny(env.ToMap())
		if err != nil {
			return InstSpec{}, errors.Wrapf(
				err, "couldn't evaluate input %s as expression %s", varName, expr,
			)
		}
		result.Inputs[varName] = value
	}
	return result, nil
}

// InstSpec

// String returns an abbreviated string representation of the InstSpec.
func (s InstSpec) String() string {
	result := ":"
	if len(s.Inputs) > 0 {
		inputs := make([]string, 0, len(s.Inputs))
		for _, varName := range slices.Sorted(maps.Keys(s.Inputs)) {
			inputs = append(inputs, fmt.Sprintf("%s=%s", varName, s.Inputs[varName]))
		}
		result += fmt.Sprintf("(%s)", strings.Join(inputs, " "))
	}
	if result == ":" {
		return ""
	}
	return result
}

// Cloned returns a deep copy of the InstSpec.
func (s InstSpec) Cloned() InstSpec {
	return InstSpec{
		Inputs: maps.Clone(s.Inputs),
	}
}

// InputValues

// Merged returns a new InputValues created by applying the specified overlay, without modifying
// this current InputValues or the overlay.
func (s InputValues) Merged(overlay InputValues) InputValues {
	merged := maps.Clone(s)
	for name, o := range overlay {
		already, alreadyHas := merged[name]
		if !alreadyHas {
			merged[name] = o
			continue
		}

		merged[name] = cmp.Or(o, already)
	}
	return merged
}

// CompPrimSpec

// Cloned returns a deep copy of the CompPrimSpec.
func (s CompPrimSpec) Cloned() CompPrimSpec {
	return CompPrimSpec{
		Kind:         s.Kind,
		StaticModels: s.StaticModels.Cloned(),
	}
}

// Merged returns a new CompPrimSpec created by applying the specified overlay, without modifying
// this current CompsPoseSpec or the overlay.
/*func (s CompPrimSpec) Merged(overlay CompPrimSpec) CompPrimSpec {
	return CompPrimSpec{
		Kind:         cmp.Or(overlay.Kind, s.Kind),
		StaticModels: s.StaticModels.Merged(overlay.StaticModels),
	}
}*/

// CompPrimStaticModelsSpec

func (s CompPrimStaticModelsSpec) Cloned() CompPrimStaticModelsSpec {
	return CompPrimStaticModelsSpec{
		GLTF:     s.GLTF,
		STEP:     s.STEP,
		Optiland: s.Optiland,
		Other:    maps.Clone(s.Other),
	}
}

func (s CompPrimStaticModelsSpec) MarshalJSON() ([]byte, error) {
	// Implements json.Marshaler
	m := make(map[string]any)
	if err := mapstructure.Decode(s, &m); err != nil {
		return nil, errors.Wrap(err, "couldn't marshal primitive static models struct as a map")
	}
	b, err := json.Marshal(m, json.Deterministic(true))
	if err != nil {
		return nil, errors.Wrap(err, "couldn't marshal primitive static models map as yaml")
	}
	return b, nil
}

func (s CompPrimStaticModelsSpec) MarshalYAML(ctx context.Context) (any, error) {
	// Implements yaml.BytesMarshalerContext
	m := make(map[string]any)
	if err := mapstructure.Decode(s, &m); err != nil {
		return nil, errors.Wrap(err, "couldn't marshal primitive static models struct as a map")
	}
	return m, nil
}

func yamlUnmarshalCompPrimStaticModelsSpec(
	ctx context.Context, s *CompPrimStaticModelsSpec, data []byte,
) error {
	m := make(map[string]any)
	if err := yaml.UnmarshalContext(ctx, data, &m); err != nil {
		return errors.Wrap(err, "couldn't unmarshal primitive static models yaml as a map")
	}
	if err := mapstructure.Decode(m, s); err != nil {
		return errors.Wrap(err, "couldn't unmarshal primitive static models map as a struct")
	}
	return nil
}

// Merged returns a new CompPrimStaticModelsSpec created by applying the specified overlay, without modifying
// this current CompsPoseSpec or the overlay.
/*func (s CompPrimStaticModelsSpec) Merged(
	overlay CompPrimStaticModelsSpec,
) CompPrimStaticModelsSpec {
	otherMerged := maps.Clone(s.Other)
	for key, value := range overlay.Other {
		if value == "" {
			continue
		}
		otherMerged[key] = value
	}

	return CompPrimStaticModelsSpec{
		GLTF:     cmp.Or(overlay.GLTF, s.GLTF),
		STEP:     cmp.Or(overlay.STEP, s.STEP),
		Optiland: cmp.Or(overlay.Optiland, s.Optiland),
		Other:    otherMerged,
	}
}*/

func (s CompPrimStaticModelsSpec) Prefixed(pathPrefix string) CompPrimStaticModelsSpec {
	otherPrefixed := make(map[string]string)
	for key, value := range s.Other {
		otherPrefixed[key] = prefixNonempty(value, pathPrefix)
	}

	return CompPrimStaticModelsSpec{
		GLTF:     prefixNonempty(s.GLTF, pathPrefix),
		STEP:     prefixNonempty(s.STEP, pathPrefix),
		Optiland: prefixNonempty(s.Optiland, pathPrefix),
		Other:    otherPrefixed,
	}
}

func prefixNonempty(s, pathPrefix string) string {
	if s == "" {
		return ""
	}

	return path.Clean(path.Join(pathPrefix, s))
}
