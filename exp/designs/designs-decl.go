package designs

import (
	"cmp"
	"fmt"
	"io/fs"
	"maps"
	"math"
	"path"
	"slices"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/pkg/errors"
	"github.com/ungerik/go3d/float64/mat4"
	"github.com/ungerik/go3d/float64/quaternion"
	"github.com/ungerik/go3d/float64/vec3"

	ffs "github.com/openUC2/optikit/exp/fs"
	"github.com/openUC2/optikit/exp/structures"
)

// DesignExprDeclFile is the name of the file defining each Optikit design.
const DesignExprDeclFile = "optikit-design.yml"

// A DesignExprDecl declares an Optikit design.
// Some parameters are string expressions which can be evaluated to produce a DesignDecl.
type DesignExprDecl struct {
	// Optikit indicates that the design was written assuming the semantics of a given version
	// of Optikit. The version must be a valid Optikit version, and it sets the minimum version of
	// Optikit required to use the design. The Optikit tool refuses to use designs declaring newer
	// Optikit versions for any operations beyond printing information. The Optikit version of the
	// design must be greater than or equal to the Optikit version of every required Optikit design.
	Optikit string `json:"optikit-version" yaml:"optikit-version"`
	// Design defines the basic metadata for the design.
	Design DesignSpec `json:"design" yaml:"design,omitempty"`
	// Components declares the design's constituent components as a mapping from the ID of each
	// component to the declaration of that component.
	// Some component parameters are string expressions which can be evaluated to produce a CompSpec.
	Components CompExprsSpec `json:"components" yaml:"components,omitempty"`
	// Inputs declares the design's input variables as a mapping from the name of each variable to the
	// declaration of that input variable.
	Inputs InputsSpec `json:"inputs,omitempty" yaml:"inputs,omitempty"`
	// Variants declares the design's variants as a mapping from the ID of each variant to the
	// declaration of that variant.
	Variants VariantsSpec `json:"variants" yaml:"variants,omitempty"`
}

// A DesignDecl declares an Optikit design.
type DesignDecl struct {
	// Optikit indicates that the design was written assuming the semantics of a given version
	// of Optikit. The version must be a valid Optikit version, and it sets the minimum version of
	// Optikit required to use the design. The Optikit tool refuses to use designs declaring newer
	// Optikit versions for any operations beyond printing information. The Optikit version of the
	// design must be greater than or equal to the Optikit version of every required Optikit design.
	Optikit string `json:"optikit-version" yaml:"optikit-version"`
	// Design defines the basic metadata for the design.
	Design DesignSpec `json:"design" yaml:"design,omitempty"`
	// Components declares the design's constituent components as a mapping from the ID of each
	// component to the declaration of that component.
	Components CompsSpec `json:"components" yaml:"components,omitempty"`
	// Inputs declares the design's input variables as a mapping from the name of each variable to the
	// declaration of that input variable.
	Inputs InputsSpec `json:"inputs,omitempty" yaml:"inputs,omitempty"`
	// Variants declares the design's variants as a mapping from the ID of each variant to the
	// declaration of that variant.
	Variants VariantsSpec `json:"variants" yaml:"variants,omitempty"`
}

// DesignSpec declares the basic metadata for an Optikit design.
type DesignSpec struct {
	// Path is the design path, which acts as the canonical name for the design.
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
	// Description is a short description of the design to be shown to users.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// Tags is a list of human-readable string tags for describing the design to software.
	Tags []string `json:"tags,omitempty" yaml:"tags,omitempty"`
}

type (
	CompID        string
	CompExprsSpec map[CompID]CompExprSpec
	CompsSpec     map[CompID]CompSpec
)

// CompExprSpec declares a component of an Optikit design.
// Some parameters are string expressions which can be evaluated to produce a CompSpec.
type CompExprSpec struct {
	// Type is the type of component in the design. It can be either `location`, `primitive`, or
	// `design`.
	Type string `json:"type" yaml:"type"`
	// Design is the path of the design which the component (of type `design`) instantiates, relative
	// to the root directory of the Optikit design.
	Design string `json:"design,omitempty" yaml:"design,omitempty"`
	// Instantiation declares information about how the design is to be instantiated to create the
	// component (of type `design`).
	// Some instantiation parameters are string expressions which can be evaluated to produce a
	// InstSpec.
	Instantiation InstExprSpec `json:"instantiation" yaml:"instantiation,omitempty"`
	// Primitive declares information about the model primitive which the component (of type
	// `primitive`) is.
	Primitive CompPrimSpec `json:"primitive" yaml:"primitive,omitempty"`
	// Pose declares the geometry of the component.
	// Some pose parameters are string expressions which can be evaluated to produce a CompPoseSpec.
	Pose CompPoseExprSpec `json:"pose" yaml:"pose,omitempty"`
}

// CompSpec declares a component of an Optikit design.
type CompSpec struct {
	// Type is the type of component in the design. It can be either `location`, `primitive`, or
	// `design`.
	Type string `json:"type" yaml:"type"`
	// Design is the path of the design which the component (of type `design`) instantiates, relative
	// to the root directory of the Optikit design.
	Design string `json:"design,omitempty" yaml:"design,omitempty"`
	// Instantiation declares information about how the design is to be instantiated to create the
	// component (of type `design`).
	Instantiation InstSpec `json:"instantiation" yaml:"instantiation,omitempty"`
	// Primitive declares information about the model primitive which the component (of type
	// `primitive`) is.
	Primitive CompPrimSpec `json:"primitive" yaml:"primitive,omitempty"`
	// Pose declares the geometry of the component.
	Pose CompPoseSpec `json:"pose" yaml:"pose,omitempty"`
}

const (
	CompTypeLocation  = "location"
	CompTypePrimitive = "primitive"
	CompTypeDesign    = "design"
)

// InstExprSpec declares how an indeterminate design is made determinate by specifying a particular
// design variant, particular values of input variables, and particular feature flags.
// The input values are string expressions which can be evaluated into concrete values.
type InstExprSpec struct {
	// Variant declares which design variant (if any) of a design will be used. The value will only
	// be evaluated as an expression if it begins with the prefix `~ `; otherwise, it will be treated
	// as a string literal directly encoding the variant ID (rather than an expression to be evaluated
	// into a variant ID).
	Variant Expr `json:"variant,omitempty" yaml:"variant,omitempty"`
	// Inputs instantiates the design's input variables to particular values, which are provided as
	// expr expressions to be evaluated into concrete values.
	Inputs map[VarName]Expr `json:"inputs,omitempty" yaml:"inputs,omitempty"`
}

// InstSpec declares how an indeterminate design is made determinate by specifying a particular
// design variant, particular values of input variables, and particular feature flags.
type InstSpec struct {
	// Variant declares which design variant (if any) of a design will be used.
	Variant VariantID `json:"variant,omitempty" yaml:"variant,omitempty"`
	// Inputs instantiates the design's input variables to particular values.
	Inputs InputValues `json:"inputs,omitempty" yaml:"inputs,omitempty"`
}

type InputValues map[VarName]any

type CompPrimSpec struct {
	// Type is the type of primitive. It can be `static`.
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
	// StaticModels declares the paths of the model files (in whatever formats are available) which the
	// primitive represents, relative to the root directory of the Optikit design.
	StaticModels CompPrimStaticModelsSpec `json:"static-models" yaml:"static-models,omitempty"`
}

// CompPrimStaticModelSpec declares equivalent files in alternate file formats representing the same
// primitive model.
type CompPrimStaticModelsSpec struct {
	GLTF string `json:"gltf,omitempty" yaml:"gltf,omitempty"`
	STEP string `json:"step,omitempty" yaml:"step,omitempty"`
}

// CompPoseExprSpec declares a Optikit design's component's geometry.
// The pose parameters are string expressions which can be evaluated to generate a CompPoseSpec.
// A zero value indicates that the component has no geometric pose.
type CompPoseExprSpec struct {
	// Rotation declares the orientation of the component as a rotation.
	// The rotation parameters are string expressions which can be evaluated to generate a
	// CompPoseRotSpec.
	Rotation CompPoseRotExprSpec `json:"rotation" yaml:"rotation,omitempty"`
	// Translation declares the position of the component as a linear translation.
	// The translation parameters are string expressions which can be evaluated to generate a
	// CompPoseTranslSpec.
	Translation CompPoseTranslExprSpec `json:"translation" yaml:"translation,omitempty"`
}

// CompPoseSpec declares a Optikit design's component's geometry.
// A zero value indicates that the component has no geometric pose.
type CompPoseSpec struct {
	// Rotation declares the orientation of the component as a rotation.
	Rotation CompPoseRotSpec `json:"rotation" yaml:"rotation,omitempty"`
	// Translation declares the position of the component as a linear translation.
	Translation CompPoseTranslSpec `json:"translation" yaml:"translation,omitempty"`
}

// CompPoseRotExprSpec declares the orientation of the component as a rotation relative to the
// overall design's orientation.
// The pose parameters are string expressions which can be evaluated to generate a CompPoseRotSpec.
type CompPoseRotExprSpec struct {
	// Type is the type of orientation of the component. It can be either `` (implying a component
	// without any spatial geometry), `uc2` (implying a UC2 cube), `grid` (for any orientation
	// aligned with the design's axes, even if violating UC2 cube orientation constraints),
	// 'euler' (for arbitrary rotations in the extrinsic z-x-y Euler angle order), or
	// `quaternion` (for arbitrary rotations).
	// If the type is uc2, then Grid.Z is only allowed to be +z or -z, and Grid.X is not allowed to
	// be +z or -z.
	Type string `json:"type,omitempty" yaml:"type"`
	// Grid declares the orientation parameters of the component if its rotation type is `uc2` or
	// `grid`.
	Grid CompPoseRotGridSpec `json:"grid" yaml:"grid,omitempty"`
	// Euler declares the orientation parameters of the component if its rotation type is
	// `euler`. Angles should be in the extrinsic Z-X-Y order, which is equivalent to the
	// intrinsic Y-X-Z order.
	Euler ExprXYZ `json:"euler" yaml:"euler,omitempty"`
	// Quaternion declares the orientation parameters of the component if its rotation type is
	// `quaternion`.
	// The quaternion should be a string expression which evaluates into a 4-component numeric array.
	Quaternion Expr `json:"quaternion" yaml:"quaternion,omitempty"`
}

// CompPoseRotSpec declares the orientation of the component as a rotation relative to the
// overall design's orientation.
type CompPoseRotSpec struct {
	// Type is the type of orientation of the component. It can be either `` (implying a component
	// without any spatial geometry), `uc2` (implying a UC2 cube), `grid` (for any orientation
	// aligned with the design's axes, even if violating UC2 cube orientation constraints), or
	// `quaternion` (for arbitrary rotations).
	// If the type is uc2, then Grid.Z is only allowed to be +z or -z, and Grid.X is not allowed to
	// be +z or -z.
	Type string `json:"type,omitempty" yaml:"type"`
	// Grid declares the orientation parameters of the component if its rotation type is `uc2` or
	// `grid`.
	Grid CompPoseRotGridSpec `json:"grid" yaml:"grid,omitempty"`
	// Euler declares the orientation parameters of the component if its rotation type is
	// `euler`. Angles should be in the extrinsic Z-X-Y order, which is equivalent to the
	// intrinsic Y-X-Z order.
	Euler ContinuousXYZ[float64] `json:"euler" yaml:"euler,omitempty"`
	// Quaternion declares the orientation parameters of the component if its rotation type is
	// `quaternion`.
	Quaternion quaternion.T `json:"quaternion" yaml:"quaternion,omitempty"`
}

const (
	RotTypeUC2        = "uc2"
	RotTypeGrid       = "grid"
	RotTypeEuler      = "euler"
	RotTypeQuaternion = "quaternion"
)

// CompPoseRotGridSpec specifies the component's orientation relative to the design's orientation by
// two discrete parameters: the orientation of the component's z-axis, and the orientation of the
// component's x-axis.
// The component's y-axis is derived from the component's x- and z-axes via the right-hand rule.
type CompPoseRotGridSpec struct {
	// Z specifies the axis of the design's coordinate system which the component's coordinate
	// system's +z direction should point in. The zero value is interpreted as +z.
	Z string `json:"z,omitempty" yaml:"z,omitempty"`
	// X specifies the axis of the design's coordinate system which the component's coordinate
	// system's +x direction should point in. The zero value is interpreted as +x.
	X string `json:"x,omitempty" yaml:"x,omitempty"`
}

// CompPoseTranslExprSpec declares the position of the component as linear translation relative to
// an "anchor" component, as an x-y-z offset along the overall design's coordinate axes.
// The pose parameters are string expressions which can be evaluated to generate a
// CompPoseTranslSpec.
type CompPoseTranslExprSpec struct {
	// Anchor is the ID of the component whose position will be linearly translated by the specified
	// offsets in order to determine the position of this component.
	// If empty, it will be the origin of the overall design's coordinate axes.
	Anchor CompID `json:"anchor,omitempty" yaml:"anchor,omitempty"`
	// OffsetGrid is an offset from the anchor's position towards the component's position, in the
	// design's coordinate axes.
	OffsetGrid ExprXYZ `json:"offset-grid" yaml:"offset-grid,omitempty"`
	// OffsetMM is an additional offset from the anchor's position towards the component's position,
	// in millimeters, after first applying the grid offset.
	OffsetMM ExprXYZ `json:"offset-mm" yaml:"offset-mm,omitempty"`
}

// CompPoseTranslSpec declares the position of the component as linear translation relative to
// an "anchor" component, as an x-y-z offset along the overall design's coordinate axes.
type CompPoseTranslSpec struct {
	// Anchor is the ID of the component whose position will be linearly translated by the specified
	// offsets in order to determine the position of this component.
	// If empty, it will be the origin of the overall design's coordinate axes.
	Anchor CompID `json:"anchor,omitempty" yaml:"anchor,omitempty"`
	// OffsetGrid is an offset from the anchor's position towards the component's position, in the
	// design's coordinate axes.
	OffsetGrid DiscreteXYZ[int] `json:"offset-grid" yaml:"offset-grid,omitempty"`
	// OffsetMM is an additional offset from the anchor's position towards the component's position,
	// in millimeters, after first applying the grid offset.
	OffsetMM ContinuousXYZ[float64] `json:"offset-mm" yaml:"offset-mm,omitempty"`
}

type (
	VarName    string
	VarType    string
	InputsSpec map[VarName]InputVarSpec
)

const (
	VarTypeBool       = "bool"
	VarTypeInt        = "int"
	VarTypeFloat64    = "float64"
	VarTypeString     = "string"
	VarTypeQuaternion = "quaternion"
)

var varTypeZeroValues = map[VarType]any{
	VarTypeBool:       false,
	VarTypeInt:        0,
	VarTypeFloat64:    0,
	VarTypeString:     "",
	VarTypeQuaternion: quaternion.T{},
}

// An InputVarSpec declares an input variable of a design, which can be referenced in
// expression-based fields in other parts of the design.
type InputVarSpec struct {
	// Description is a short description of the variable to be shown to users.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// Type is a string indicating the expected type of the variable, for type-checking. Allowed
	// values are: bool, int, float64, string
	Type VarType `json:"type,omitempty" yaml:"type,omitempty"`
	// Units is a string indicating the expected units of the variable, to be shown to users.
	Units string `json:"units,omitempty" yaml:"units,omitempty"`
	// Min is the minimum allowed value of the variable. It should be either an int or a float64.
	Min any `json:"min,omitempty" yaml:"min,omitempty"`
	// Max is the maximum allowed value of the variable. It should be either an int or a float64.
	Max any `json:"max,omitempty" yaml:"max,omitempty"`
}

type (
	VariantID    string
	VariantsSpec map[VariantID]VariantSpec
)

// A VariantSpec declares a design variant.
type VariantSpec struct {
	// Description is a short description of the variant to be shown to users.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// Components declares any modifications to the design's components. Non-zero values here will
	// overwrite non-zero values in the design's components; new components here will also be added to
	// the design.
	Components CompExprsSpec `json:"components,omitempty" yaml:"components,omitempty"`
	// Inputs declares any modifications to the design's input variables. Non-zero values here will
	// overwrite non-zero values in the design's input variables; new components here will also be
	// added to the design.
	Inputs InputsSpec `json:"inputs,omitempty" yaml:"inputs,omitempty"`
}

// DesignExprDecl

// LoadDesignExprDecl loads a DesignExprDecl from the specified file path in the provided base
// filesystem.
func LoadDesignExprDecl(fsys ffs.PathedFS, filePath string) (DesignExprDecl, error) {
	bytes, err := fs.ReadFile(fsys, filePath)
	if err != nil {
		return DesignExprDecl{}, errors.Wrapf(
			err, "couldn't read design config file %s/%s", fsys.Path(), filePath,
		)
	}
	config := DesignExprDecl{}
	if err = yaml.Unmarshal(bytes, &config); err != nil {
		return DesignExprDecl{}, errors.Wrap(err, "couldn't parse design declaration with expressions")
	}
	return config, nil
}

// Check looks for errors in the construction of the design configuration.
func (d DesignExprDecl) Check() (errs []error) {
	errs = append(errs, errsWrap(d.Design.Check(), "invalid design spec")...)
	errs = append(errs, errsWrap(d.Components.Check(), "invalid components spec")...)
	return errs
}

// Cloned returns a deep copy of the DesignExprDecl.
func (d DesignExprDecl) Cloned() DesignExprDecl {
	d.Design.Tags = slices.Clone(d.Design.Tags)
	d.Components = maps.Clone(d.Components)
	d.Inputs = maps.Clone(d.Inputs)
	d.Variants = maps.Clone(d.Variants)
	return d
}

// Instantiated returns a CompExprsSpec which has been modified with design variants, input
// variables, and feature flags, as specified by the provided instantiation parameters.
func (d DesignExprDecl) Instantiated(instantiation InstSpec) (dd DesignDecl, err error) {
	s := d.Components
	dd.Inputs = d.Inputs
	if instantiation.Variant != "" {
		v, has := d.Variants[instantiation.Variant]
		if !has {
			return dd, errors.Errorf("requested variant not found: %s", instantiation.Variant)
		}
		dd.Inputs = dd.Inputs.Merged(v.Inputs)
		s = d.Components.Merged(v.Components)
	}

	inputEnv, err := MakeExprEnv(instantiation.Inputs, dd.Inputs)
	if err != nil {
		return dd, errors.Wrapf(
			err, "couldn't make expression env with inputs %+v", instantiation.Inputs,
		)
	}
	if dd.Components, err = s.Evaluated(inputEnv); err != nil {
		return dd, errors.Wrapf(err, "couldn't evaluate expressions with inputs %+v", inputEnv)
	}
	return dd, nil
}

// DesignDecl

// LoadDesignDecl loads a DesignExprDecl from the specified file path in the provided base
// filesystem.
func LoadDesignDecl(fsys ffs.PathedFS, filePath string) (DesignDecl, error) {
	bytes, err := fs.ReadFile(fsys, filePath)
	if err != nil {
		return DesignDecl{}, errors.Wrapf(
			err, "couldn't read design config file %s/%s", fsys.Path(), filePath,
		)
	}
	config := DesignDecl{}
	if err = yaml.Unmarshal(bytes, &config); err != nil {
		return DesignDecl{}, errors.Wrap(err, "couldn't parse design declaration")
	}
	return config, nil
}

// Check looks for errors in the construction of the design configuration.
func (d DesignDecl) Check() (errs []error) {
	errs = append(errs, errsWrap(d.Design.Check(), "invalid design spec")...)
	// TODO: make the components check account for declared input variables
	errs = append(errs, errsWrap(d.Components.Check(), "invalid components spec")...)
	return errs
}

// Cloned returns a deep copy of the DesignDecl.
func (d DesignDecl) Cloned() DesignDecl {
	d.Design.Tags = slices.Clone(d.Design.Tags)
	d.Components = maps.Clone(d.Components)
	d.Inputs = maps.Clone(d.Inputs)
	d.Variants = maps.Clone(d.Variants)
	return d
}

// DesignSpec

// Check looks for errors in the construction of the design spec.
func (s DesignSpec) Check() (errs []error) {
	return errs
}

// CompExprsSpec

// Check looks for errors in the construction of the components spec.
func (s CompExprsSpec) Check() (errs []error) {
	for id, component := range s {
		anchor := component.Pose.Translation.Anchor
		if _, exists := s[anchor]; anchor != "" && !exists {
			errs = append(errs, errors.Errorf(
				"component %s depends on nonexistent translation anchor %s", id, anchor,
			))
		}
		// TODO: check for validity of instantiation...or maybe we must do this in FSDesign
	}
	return errs
}

// Merged returns a new CompExprsSpec created by applying the specified overlay, without modifying
// this current CompExprsSpec or the overlay.
func (s CompExprsSpec) Merged(overlay CompExprsSpec) CompExprsSpec {
	merged := maps.Clone(s)
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
	for id, component := range s {
		anchor := component.Pose.Translation.Anchor
		if _, exists := s[anchor]; anchor != "" && !exists {
			errs = append(errs, errors.Errorf(
				"component %s depends on nonexistent translation anchor %s", id, anchor,
			))
		}
		// TODO: check for validity of instantiation...or maybe we must do this in FSDesign
	}
	return errs
}

// Poses returns a map from component IDs to their poses.
func (s CompsSpec) Poses() map[CompID]CompPoseSpec {
	poses := make(map[CompID]CompPoseSpec)
	for id, component := range s {
		poses[id] = component.Pose
	}
	return poses
}

type TranslDigraph = structures.StrictEdgeDigraph[CompID, CompPoseTranslSpec]

// TranslDigraph returns a StrictEdgeDigraph of the translation relationships between components.
// It assumes that the CompsSpec does not have any errors such as a nonexistent translation anchor
// required by a CompPosesTranslSpec.
func (s CompsSpec) TranslDigraph() TranslDigraph {
	g := make(TranslDigraph)
	g.AddNode("") // origin
	for compName, comp := range s {
		g.AddNode(compName)
		anchor := comp.Pose.Translation.Anchor
		g.AddEdge(anchor, compName, comp.Pose.Translation)
	}
	return g
}

// TranslFlattened returns a new CompsSpec in which each non-origin component's translation anchor
// is just the root (origin) node.
// It assumes that the CompsSpec does not have any errors such as a nonexistent translation anchor
// required by a CompPosesTranslSpec.
func (s CompsSpec) TranslFlattened() CompsSpec {
	flattened := make(CompsSpec)
	g := s.TranslDigraph()
	nextParents := make([]CompID, 0, len(g))
	nextParents = append(nextParents, "") // add the root node
	for len(nextParents) > 0 {
		parent := nextParents[0]
		parentPos := flattened[parent].Pose.Translation
		nextParents = nextParents[1:]
		for child := range g[parent] {
			nextParents = append(nextParents, child)
			c := s[child]
			c.Pose.Translation = c.Pose.Translation.Added(parentPos)
			c.Pose.Translation.Anchor = ""
			flattened[child] = c
		}
	}
	return flattened
}

// Primitives returns the primitive-type components in this CompsSpec.
func (s CompsSpec) Primitives() CompsSpec {
	prims := make(CompsSpec)
	for id, c := range s {
		if c.Type != "primitive" {
			continue
		}
		prims[id] = c
	}
	return prims
}

// CompID

// JoinCompIDs concatenates component IDs with slash ("/") delimiters into a path-style name.
func JoinCompIDs(elem ...CompID) CompID {
	elems := make([]string, 0, len(elem))
	for _, e := range elem {
		elems = append(elems, string(e))
	}
	return CompID(path.Join(elems...))
}

// CompExprSpec

// Merged returns a new CompExprSpec created by applying the specified overlay, without modifying
// this current CompExprSpec or the overlay.
func (s CompExprSpec) Merged(overlay CompExprSpec) CompExprSpec {
	return CompExprSpec{
		Type:          cmp.Or(overlay.Type, s.Type),
		Design:        cmp.Or(overlay.Design, s.Design),
		Instantiation: s.Instantiation.Merged(overlay.Instantiation),
		Primitive:     s.Primitive.Merged(overlay.Primitive),
		Pose:          s.Pose.Merged(overlay.Pose),
	}
}

// Evaluated evaluates the expressions with the given ExprEnv into a CompSpec.
func (s CompExprSpec) Evaluated(env ExprEnv) (result CompSpec, err error) {
	result = CompSpec{
		Type:      s.Type,
		Design:    s.Design,
		Primitive: s.Primitive,
	}
	if result.Instantiation, err = s.Instantiation.Evaluated(env); err != nil {
		return result, errors.Wrap(err, "couldn't evaluate expressions in instantiation section")
	}
	if result.Pose, err = s.Pose.Evaluated(env); err != nil {
		return result, errors.Wrap(err, "couldn't evaluate expressions in pose section")
	}
	return result, nil
}

// InstExprSpec

// Merged returns a new InstExprSpec created by applying the specified overlay, without modifying
// this current InstExprSpec or the overlay.
func (s InstExprSpec) Merged(overlay InstExprSpec) InstExprSpec {
	merged := InstExprSpec{
		Variant: cmp.Or(overlay.Variant, s.Variant),
	}
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
}

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
	if result.Variant, err = s.Variant.evalAsString[VariantID](env.ToMap()); err != nil {
		return result, errors.Wrapf(err, "couldn't evaluate variant expression %s", s.Variant)
	}
	return result, nil
}

// InstSpec

// String returns an abbreviated string representation of the InstSpec.
func (s InstSpec) String() string {
	result := ":"
	if s.Variant != "" {
		result += string(s.Variant)
	}
	if len(s.Inputs) > 0 {
		inputs := make([]string, len(s.Inputs), 0)
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

// Merged returns a new CompPrimSpec created by applying the specified overlay, without modifying
// this current CompsPoseSpec or the overlay.
func (s CompPrimSpec) Merged(overlay CompPrimSpec) CompPrimSpec {
	return CompPrimSpec{
		Type:         cmp.Or(overlay.Type, s.Type),
		StaticModels: s.StaticModels.Merged(overlay.StaticModels),
	}
}

// CompPrimStaticModelsSpec

// Merged returns a new CompPrimStaticModelsSpec created by applying the specified overlay, without modifying
// this current CompsPoseSpec or the overlay.
func (s CompPrimStaticModelsSpec) Merged(
	overlay CompPrimStaticModelsSpec,
) CompPrimStaticModelsSpec {
	return CompPrimStaticModelsSpec{
		GLTF: cmp.Or(overlay.GLTF, s.GLTF),
		STEP: cmp.Or(overlay.STEP, s.STEP),
	}
}

func (s CompPrimStaticModelsSpec) Prefixed(pathPrefix string) CompPrimStaticModelsSpec {
	return CompPrimStaticModelsSpec{
		GLTF: path.Clean(path.Join(pathPrefix, s.GLTF)),
		STEP: path.Clean(path.Join(pathPrefix, s.STEP)),
	}
}

// CompPoseExprSpec

// Merged returns a new CompPoseExprSpec created by applying the specified overlay, without
// modifying this current CompsPoseExprSpec or the overlay.
func (s CompPoseExprSpec) Merged(overlay CompPoseExprSpec) CompPoseExprSpec {
	return CompPoseExprSpec{
		Rotation:    s.Rotation.Merged(overlay.Rotation),
		Translation: s.Translation.Merged(overlay.Translation),
	}
}

// Evaluated evaluates the pose expressions with the given ExprEnv into a CompPoseSpec.
func (s CompPoseExprSpec) Evaluated(env ExprEnv) (result CompPoseSpec, err error) {
	if result.Rotation, err = s.Rotation.Evaluated(env); err != nil {
		return CompPoseSpec{}, errors.Wrap(err, "couldn't evaluate rotation")
	}
	if result.Translation, err = s.Translation.Evaluated(env); err != nil {
		return CompPoseSpec{}, errors.Wrap(err, "couldn't evaluate translation")
	}
	return result, nil
}

// CompPoseSpec

// NewPose builds a new CompPoseSpec from a transformation matrix.
// The translation component is decomposed into a discrete component (for any non-zero grid
// spacings) and any remaining non-discrete component.
func NewPose(mat mat4.T, gridSpacings ContinuousXYZ[float64]) CompPoseSpec {
	return CompPoseSpec{
		Rotation:    NewPoseRot(mat),
		Translation: NewPoseTransl(mat, gridSpacings),
	}
}

// TransfMat returns a homogeneous affine transformation matrix representing the pose of the
// component relative to the frame of the overall design, but only if the pose's translation is
// specified with the overall design's coordinate system's origin as the anchor. If anything else is
// the anchor, then this method returns an error instead.
// The translation component of the matrix is in mm.
// This is the matrix H^a_b for homogeneous pose vectors p^a_h and p^b_h, which are homogeneous
// representations of vectors p^a and p^b, where p^b is in the frame of the component and p^b is in
// the frame of the overall design. In other words, this matrix can be multiplied with a point in
// the frame of the component to get the position of that point in the frame of the overall design.
func (s CompPoseSpec) TransfMat(gridSpacings ContinuousXYZ[float64]) (mat4.T, error) {
	if s.Translation.Anchor != "" {
		return mat4.Zero, errors.New("translation anchor is not the overall design's origin!")
	}
	m := s.Rotation.TransfMat()
	offsetGrid := AsMM(s.Translation.OffsetGrid, gridSpacings).AsVec3()
	offsetMM := s.Translation.OffsetMM.AsVec3()
	translation := vec3.Add(&offsetGrid, &offsetMM)
	m.SetTranslation(&translation)
	return m, nil
}

// CompPoseRotExprSpec

// Merged returns a new CompPoseRotExprSpec created by applying the specified overlay, without
// modifying this current CompsPoseExprSpec or the overlay.
func (s CompPoseRotExprSpec) Merged(overlay CompPoseRotExprSpec) CompPoseRotExprSpec {
	t := cmp.Or(overlay.Type, s.Type)
	switch t {
	default:
		return CompPoseRotExprSpec{}
	case RotTypeUC2, RotTypeGrid:
		return CompPoseRotExprSpec{
			Type: t,
			Grid: s.Grid.Merged(overlay.Grid),
		}
	case RotTypeEuler:
		return CompPoseRotExprSpec{
			Type:  t,
			Euler: s.Euler.Merged(overlay.Euler),
		}
	case RotTypeQuaternion:
		return CompPoseRotExprSpec{
			Type:       t,
			Quaternion: cmp.Or(overlay.Quaternion, s.Quaternion),
		}
	}
}

// Evaluated evaluates the pose expressions with the given ExprEnv into a CompPoseRotSpec.
// Parameters not associated with the CompPoseRotExprSpec's type are excluded from the result; for
// example, if the rotation type is "quaternion", then the result's Grid field will be zero.
func (s CompPoseRotExprSpec) Evaluated(env ExprEnv) (result CompPoseRotSpec, err error) {
	result.Type = s.Type
	switch result.Type {
	case RotTypeUC2, RotTypeGrid:
		result.Grid = s.Grid
	case RotTypeEuler:
		if result.Euler, err = s.Euler.EvaluatedFloat64(env.ToMap()); err != nil {
			return CompPoseRotSpec{}, errors.Wrap(err, "couldn't evaluate euler")
		}
	case RotTypeQuaternion:
		if s.Quaternion != "" {
			evaluated, err := s.Quaternion.evalAsAny(env.ToMap())
			if err != nil {
				return CompPoseRotSpec{}, errors.Wrapf(
					err, "couldn't evaluate quat expr %s", s.Quaternion,
				)
			}
			if result.Quaternion, err = convertToQuaternion(evaluated); err != nil {
				return CompPoseRotSpec{}, errors.Wrapf(
					err, "evaluated quat expr %s as array %+v, but couldn't convert it to a quaternion",
					s.Quaternion, evaluated,
				)
			}
		}
	}
	return result, nil
}

// CompPoseRotSpec

// NewPoseRot builds a CompPoseRotSpec from a transformation matrix. If the transformation matrix
// specifies an axis-aligned rotation, then the result will be of type "grid" (note: it will never
// be of type "uc2"). Otherwise, the result will be of type "quaternion".
func NewPoseRot(mat mat4.T) CompPoseRotSpec {
	z := mat.MulVec3(&vec3.UnitZ)
	zDir, zAxisAligned := BasisDirs[z]
	x := mat.MulVec3(&vec3.UnitX)
	xDir, xAxisAligned := BasisDirs[x]
	if zAxisAligned && xAxisAligned {
		return CompPoseRotSpec{
			Type: RotTypeGrid,
			Grid: CompPoseRotGridSpec{
				Z: zDir,
				X: xDir,
			},
		}
	}
	return CompPoseRotSpec{
		Type:       RotTypeQuaternion,
		Quaternion: mat.Quaternion(),
	}
}

// Check looks for errors in the construction of the component orientation spec.
func (s CompPoseRotSpec) Check() (errs []error) {
	switch s.Type {
	default:
		return []error{errors.Errorf("invalid rotation type: %s", s.Type)}
	case "":
		return nil
	case RotTypeUC2:
		switch s.Grid.Z {
		case "", DirZPos, DirZNeg:
		default:
			errs = append(errs, errors.Errorf("invalid value for component's z-axis: %s", s.Grid.Z))
		}
		switch s.Grid.X {
		case "", DirXPos, DirYPos, DirXNeg, DirYNeg:
		default:
			errs = append(errs, errors.Errorf("invalid value for component's x-axis: %s", s.Grid.X))
		}
		return append(errs, s.Grid.Check()...)
	case RotTypeGrid:
		return s.Grid.Check()
	case RotTypeQuaternion:
		const tolerance = 1e-6
		if !s.Quaternion.IsUnitQuat(tolerance) {
			return append(errs, errors.Errorf("quaternion is not a unit quaternion: %+v", s.Quaternion))
		}
		return nil
	}
}

// CompPoseRotGridSpec

// Check looks for errors in the construction of the component grid orientation spec.
func (s CompPoseRotGridSpec) Check() (errs []error) {
	if s.Z[1] == s.X[1] {
		errs = append(errs, errors.Errorf("component's z and x axes are coaxial: z=%s, x=%s", s.Z, s.X))
	}
	return errs
}

// Merged returns a new CompPoseRotGridSpec created by applying the specified overlay, without
// modifying this current CompsPoseSpec or the overlay.
func (s CompPoseRotGridSpec) Merged(overlay CompPoseRotGridSpec) CompPoseRotGridSpec {
	return CompPoseRotGridSpec{
		Z: cmp.Or(overlay.Z, s.Z),
		X: cmp.Or(overlay.X, s.X),
	}
}

// TransfMat returns a homogeneous transformation matrix representing the orientation of the
// component relative to the frame of the design. If the rotation type is empty, then it'll return
// a zero matrix; otherwise, it assumes that the component orientation spec is valid.
// The first column is the component's x-axis, represented in the coordinate system of the overall
// design. The second and third columns are the y- and z-axes, respectively.
func (s CompPoseRotSpec) TransfMat() mat4.T {
	switch s.Type {
	default:
		return mat4.T{}
	case RotTypeUC2, RotTypeGrid:
		return GridRotMats[cmp.Or(s.Grid.Z, DirZPos)][cmp.Or(s.Grid.X, DirXPos)]
	case RotTypeEuler:
		mat := mat4.Zero
		mat.AssignEulerRotation(
			degToRad(s.Euler.Y), degToRad(s.Euler.X), degToRad(s.Euler.Z),
		)
		return mat
	case RotTypeQuaternion:
		mat := mat4.Zero
		mat.AssignQuaternion(&s.Quaternion)
		return mat
	}
}

func degToRad(deg float64) float64 {
	return deg * (math.Pi / 180.0) //nolint:mnd // the entire function is a magic number conversion...
}

// CompPoseTranslExprSpec

// Evaluated evaluates the pose expressions with the given ExprEnv into a CompPoseSpec.
func (s CompPoseTranslExprSpec) Evaluated(env ExprEnv) (result CompPoseTranslSpec, err error) {
	result = CompPoseTranslSpec{
		Anchor: s.Anchor,
	}
	if result.OffsetGrid, err = s.OffsetGrid.EvaluatedInt(env.ToMap()); err != nil {
		return CompPoseTranslSpec{}, errors.Wrap(err, "couldn't evaluate offsetGrid")
	}
	if result.OffsetMM, err = s.OffsetMM.EvaluatedFloat64(env.ToMap()); err != nil {
		return CompPoseTranslSpec{}, errors.Wrap(err, "couldn't evaluate offsetMM")
	}
	return result, nil
}

// Merged returns a new CompPoseTranslSpec created by applying the specified overlay, without modifying
// this current CompsPoseSpec or the overlay.
func (s CompPoseTranslExprSpec) Merged(overlay CompPoseTranslExprSpec) CompPoseTranslExprSpec {
	return CompPoseTranslExprSpec{
		Anchor:     cmp.Or(overlay.Anchor, s.Anchor),
		OffsetGrid: s.OffsetGrid.Merged(overlay.OffsetGrid),
		OffsetMM:   s.OffsetMM.Merged(overlay.OffsetMM),
	}
}

// CompPoseTranslSpec

// NewPoseTransl builds a new CompPoseTranslSpec from a transformation matrix.
// The translation component is decomposed into a discrete component (for any non-zero grid
// spacings) and any remaining non-discrete component.
func NewPoseTransl(mat mat4.T, gridSpacings ContinuousXYZ[float64]) CompPoseTranslSpec {
	transl := mat.MulVec3(&vec3.Zero)
	var gridded DiscreteXYZ[int]
	if spacing := gridSpacings.X; spacing != 0 {
		gridded.X = int(transl[0] / spacing)
	}
	if spacing := gridSpacings.Y; spacing != 0 {
		gridded.Y = int(transl[1] / spacing)
	}
	if spacing := gridSpacings.Z; spacing != 0 {
		gridded.Z = int(transl[2] / spacing)
	}
	griddedMM := AsMM(gridded, gridSpacings)
	var mm ContinuousXYZ[float64]
	mm.X = transl[0] - griddedMM.X
	mm.Y = transl[1] - griddedMM.Y
	mm.Z = transl[2] - griddedMM.Z
	return CompPoseTranslSpec{
		OffsetGrid: gridded,
		OffsetMM:   mm,
	}
}

// Merged returns a new CompPoseTranslSpec created by applying the specified overlay, without
// modifying this current CompsPoseSpec or the overlay.
func (s CompPoseTranslSpec) Merged(overlay CompPoseTranslSpec) CompPoseTranslSpec {
	return CompPoseTranslSpec{
		Anchor:     cmp.Or(overlay.Anchor, s.Anchor),
		OffsetGrid: s.OffsetGrid.Merged(overlay.OffsetGrid),
		OffsetMM:   s.OffsetMM.Merged(overlay.OffsetMM),
	}
}

// String returns an abbreviated representation of the CompPoseTranslSpec.
func (s CompPoseTranslSpec) String() string {
	switch {
	case s.OffsetGrid == gridZero && s.OffsetMM == mmZero:
		return ""
	case s.OffsetGrid == gridZero:
		return fmt.Sprintf("%s mm", s.OffsetMM.String())
	case s.OffsetMM == mmZero:
		return s.OffsetGrid.String()
	default:
		return fmt.Sprintf("%s + %s mm", s.OffsetGrid.String(), s.OffsetMM.String())
	}
}

var (
	gridZero DiscreteXYZ[int]
	mmZero   ContinuousXYZ[float64]
)

// Added returns the vector sum of the translation specified by this CompPoseTranslSpec and the
// translation specified by the provided CompPoseTranslSpec.
func (s CompPoseTranslSpec) Added(t CompPoseTranslSpec) CompPoseTranslSpec {
	return CompPoseTranslSpec{
		Anchor:     s.Anchor,
		OffsetGrid: s.OffsetGrid.Added(t.OffsetGrid),
		OffsetMM:   s.OffsetMM.Added(t.OffsetMM),
	}
}

// InputsSpec

// Merged returns a new InputsSpec created by applying the specified overlay, without modifying
// this current InputsSpec or the overlay.
func (s InputsSpec) Merged(overlay InputsSpec) InputsSpec {
	merged := maps.Clone(s)
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
		zeroes[id] = varTypeZeroValues[spec.Type]
	}
	return zeroes
}

// InputVarSpec

// Merged returns a new InputSpec created by applying the specified overlay, without modifying
// this current InputSpec or the overlay.
func (s InputVarSpec) Merged(overlay InputVarSpec) InputVarSpec {
	return InputVarSpec{
		Description: cmp.Or(overlay.Description, s.Description),
		Type:        cmp.Or(overlay.Type, s.Type),
		Units:       cmp.Or(overlay.Units, s.Units),
		Min:         cmp.Or(overlay.Min, s.Min),
		Max:         cmp.Or(overlay.Max, s.Max),
	}
}
