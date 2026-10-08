package designs

import (
	"bytes"
	"context"
	"io/fs"
	"slices"

	"github.com/goccy/go-yaml"
	"github.com/pkg/errors"
	"github.com/ungerik/go3d/float64/quaternion"

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
	Design DesignSpec `json:"design,omitzero" yaml:"design,omitempty"`
	// Inputs declares the design's input variables as a mapping from the name of each variable to the
	// declaration of that input variable.
	Inputs InputsSpec `json:"inputs,omitempty" yaml:"inputs,omitempty"`
	// Components declares the design's constituent components as a mapping from the ID of each
	// component to the declaration of that component.
	// Some component parameters are string expressions which can be evaluated to produce a CompSpec.
	Components CompExprsSpec `json:"components,omitempty" yaml:"components,omitempty"`
	// Assemblies declares the design's assemblies as a mapping from the ID of each assembly to the
	// declaration of that assembly.
	Assemblies AssmExprsSpec `json:"assemblies,omitempty" yaml:"assemblies,omitempty"`
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
	Design DesignSpec `json:"design,omitzero" yaml:"design,omitempty"`
	// Inputs declares the design's input variables as a mapping from the name of each variable to the
	// declaration of that input variable.
	Inputs InputsSpec `json:"inputs,omitempty" yaml:"inputs,omitempty"`
	// Components declares the design's constituent components as a mapping from the ID of each
	// component to the declaration of that component.
	Components CompsSpec `json:"components,omitempty" yaml:"components,omitempty"`
	// Assemblies declares the design's assemblies as a mapping from the ID of each assembly to the
	// declaration of that assembly.
	Assemblies AssmsSpec `json:"assemblies,omitempty" yaml:"assemblies,omitempty"`
}

// DesignSpec declares the basic metadata for an Optikit design.
type DesignSpec struct {
	// Path is the design path, which acts as the canonical name for the design.
	Path string `json:"path,omitzero" yaml:"path,omitempty"`
	// Description is a short description of the design to be shown to users.
	Description string `json:"description,omitzero" yaml:"description,omitempty"`
	// Tags is a list of human-readable string tags for describing the design to software.
	Tags Tags `json:"tags,omitempty" yaml:"tags,omitempty"`
}

type Tags []string

type (
	VarName    string
	InputsSpec map[VarName]InputVarSpec
)

var varKindZeroValues = map[VarKind]any{
	VarKindBool:       false,
	VarKindInt:        0,
	VarKindFloat64:    0,
	VarKindString:     "",
	VarKindQuaternion: quaternion.T{},
}

// An InputVarSpec declares an input variable of a design, which can be referenced in
// expression-based fields in other parts of the design.
type InputVarSpec struct {
	// Description is a short description of the variable to be shown to users.
	Description string `json:"description,omitzero" yaml:"description,omitempty"`
	// Kind is a string indicating the expected type of the variable, for type-checking. Allowed
	// values are: bool, int, float64, string
	Kind VarKind `json:"kind,omitzero" yaml:"kind,omitempty"`
	// Units is a string indicating the expected units of the variable, to be shown to users.
	Units string `json:"units,omitzero" yaml:"units,omitempty"`
	// Min is the minimum allowed value of the variable. It should be either an int or a float64.
	Min any `json:"min,omitzero" yaml:"min,omitempty"`
	// Max is the maximum allowed value of the variable. It should be either an int or a float64.
	Max any `json:"max,omitzero" yaml:"max,omitempty"`
	// Tags is a list of human-readable string tags for describing the input variable to software.
	Tags Tags `json:"tags,omitempty" yaml:"tags,omitempty"`
}

type (
	VarKind string
)

const (
	VarKindBool       = "bool"
	VarKindInt        = "int"
	VarKindFloat64    = "float64"
	VarKindString     = "string"
	VarKindQuaternion = "quaternion"
)

type (
	CompID        string
	CompExprsSpec map[CompID]CompExprSpec
	CompsSpec     map[CompID]CompSpec
)

// CompExprSpec declares a component of an Optikit design.
// Some parameters are string expressions which can be evaluated to produce a CompSpec.
type CompExprSpec struct {
	// Kind is the type of component in the design. It can be either `location`, `primitive`, or
	// `design`.
	Kind string `json:"kind" yaml:"kind"`
	// Design is the path of the design which the component (of kind `design`) instantiates, relative
	// to the root directory of the Optikit design.
	Design string `json:"design,omitzero" yaml:"design,omitempty"`
	// Instantiation declares information about how the design is to be instantiated to create the
	// component (of kind `design`).
	// Some instantiation parameters are string expressions which can be evaluated to produce a
	// InstSpec.
	Instantiation InstExprSpec `json:"instantiation,omitzero" yaml:"instantiation,omitempty"`
	// Primitive declares information about the model primitive which the component (of kind
	// `primitive`) is.
	Primitive CompPrimSpec `json:"primitive,omitzero" yaml:"primitive,omitempty"`
	// Results declares expressions to be evaluated with a given set of input variables.
	Results map[string]Expr `json:"results,omitempty" yaml:"results,omitempty"`
	// Tags is a list of human-readable string tags for describing the component to software.
	Tags Tags `json:"tags,omitempty" yaml:"tags,omitempty"`
}

// CompSpec declares a component of an Optikit design.
type CompSpec struct {
	// Kind is the type of component in the design. It can be either `location`, `primitive`, or
	// `design`.
	Kind string `json:"kind" yaml:"kind"`
	// Design is the path of the design which the component (of kind `design`) instantiates, relative
	// to the root directory of the Optikit design.
	Design string `json:"design,omitzero" yaml:"design,omitempty"`
	// Instantiation declares information about how the design is to be instantiated to create the
	// component (of kind `design`).
	Instantiation InstSpec `json:"instantiation,omitzero" yaml:"instantiation,omitempty"`
	// Primitive declares information about the model primitive which the component (of kind
	// `primitive`) is.
	Primitive CompPrimSpec `json:"primitive,omitzero" yaml:"primitive,omitempty"`
	// Results declares the results of expressions evaluated with a given set of input variables.
	Results map[string]any `json:"results,omitempty" yaml:"results,omitempty"`
	// Tags is a list of human-readable string tags for describing the component to software.
	Tags Tags `json:"tags,omitempty" yaml:"tags,omitempty"`
}

const (
	CompKindLocation  = "location"
	CompKindDesign    = "design"
	CompKindPrimitive = "primitive"
)

// InstExprSpec declares how an indeterminate design is made determinate by specifying a particular
// design variant, particular values of input variables, and particular feature flags.
// The input values are string expressions which can be evaluated into concrete values.
type InstExprSpec struct {
	// Inputs instantiates the design's input variables to particular values, which are provided as
	// expr expressions to be evaluated into concrete values.
	Inputs map[VarName]Expr `json:"inputs,omitempty" yaml:"inputs,omitempty"`
}

// InstSpec declares how an indeterminate design is made determinate by specifying a particular
// design variant, particular values of input variables, and particular feature flags.
type InstSpec struct {
	// Inputs instantiates the design's input variables to particular values.
	Inputs InputValues `json:"inputs,omitempty" yaml:"inputs,omitempty"`
}

type InputValues map[VarName]any

type CompPrimSpec struct {
	// Kind is the type of primitive. It can be `static`.
	Kind string `json:"kind,omitzero" yaml:"kind,omitempty"`
	// StaticModels declares the paths of the model files (in whatever formats are available) which the
	// primitive represents, relative to the root directory of the Optikit design.
	StaticModels CompPrimStaticModelsSpec `json:"static-models,omitzero" yaml:"static-models,omitempty"`
}

// CompPrimStaticModelSpec declares equivalent files in alternate file formats representing the same
// primitive model.
type CompPrimStaticModelsSpec struct {
	GLTF     string            `mapstructure:"gltf,omitempty"`
	STEP     string            `mapstructure:"step,omitempty"`
	Optiland string            `mapstructure:"optiland,omitempty"`
	Other    map[string]string `mapstructure:",remain"`
}

type (
	AssmID        string
	AssmExprsSpec map[AssmID]AssmExprSpec
	AssmsSpec     map[AssmID]AssmSpec
)

// AssmExprSpec declares an assembly of an Optikit design.
// Some parameters are string expressions which can be evaluated to produce an AssmSpec.
type AssmExprSpec struct {
	// Description is a short description of the assembly to be shown to users.
	Description string `json:"description,omitzero" yaml:"description,omitempty"`
	// Children declares the top-level components of the assembly.
	Children AssmCompExprsSpec `json:"children,omitzero" yaml:"children,omitempty"`
	// Tags is a list of human-readable string tags for describing the component to software.
	Tags Tags `json:"tags,omitempty" yaml:"tags,omitempty"`
}

// AssmSpec declares an assembly of an Optikit design.
type AssmSpec struct {
	// Description is a short description of the assembly to be shown to users.
	Description string `json:"description,omitzero" yaml:"description,omitempty"`
	// Children declares the top-level components of the assembly.
	Children AssmCompsSpec `json:"children,omitzero" yaml:"children,omitempty"`
	// Tags is a list of human-readable string tags for describing the component to software.
	Tags Tags `json:"tags,omitempty" yaml:"tags,omitempty"`
}

type (
	AssmCompExprsSpec map[CompID]AssmCompExprSpec
	AssmCompsSpec     map[CompID]AssmCompSpec
)

// AssmCompExprSpec declares the placement of a component in an assembly.
// Some parameters are string expressions which can be evaluated to produce a AssmCompSpec.
type AssmCompExprSpec struct {
	// Assm declares the assembly of the component which will be included.
	Assm Expr `json:"assembly,omitzero" yaml:"assembly,omitempty"`
	// Pose declares the geometry of the component.
	// Some pose parameters are string expressions which can be evaluated to produce a AssmCompPoseSpec.
	Pose AssmCompPoseExprSpec `json:"pose,omitzero" yaml:"pose,omitempty"`
	// Children declares the components whose poses depend in some way on this component.
	Children AssmCompExprsSpec `json:"children,omitempty" yaml:"children,omitempty"`
}

// AssmCompSpec declares the placement of a component in an assembly.
type AssmCompSpec struct {
	// Assm declares the assembly of the component which will be included, if the component is an
	// instance of a design.
	Assm AssmID `json:"assembly,omitzero" yaml:"assembly,omitempty"`
	// Pose declares the geometry of the component.
	Pose AssmCompPoseSpec `json:"pose,omitzero" yaml:"pose,omitempty"`
	// Children declares the components whose poses depend in some way on this component.
	Children AssmCompsSpec `json:"children,omitempty" yaml:"children,omitempty"`
}

// AssmCompPoseExprSpec declares a Optikit design's component's geometry.
// The pose parameters are string expressions which can be evaluated to generate a CompPoseSpec.
// A zero value indicates that the component has no geometric pose.
type AssmCompPoseExprSpec struct {
	// Kind declares the kind of geometric transformation for the component's pose. It can be
	// either `` (equivalent to `affine`), `affine` (for a standard affine transformation
	// relative to the base coordinate system), or `relative-position` (in which the component
	// is rotated with respect to the assembly's basis vectors and then translated from the base's
	// origin).
	Kind string `json:"kind,omitzero" yaml:"kind,omitempty"`
	// Base declares the base coordinate system for the geometric transformation which sets the
	// component's pose.
	Base AssmCompPoseBaseSpec `json:"base,omitzero" yaml:"base,omitempty"`
	// Rotation declares the orientation of the component as a rotation.
	// The rotation parameters are string expressions which can be evaluated to generate a
	// CompPoseRotSpec.
	Rotation AssmCompPoseRotExprSpec `json:"rotation,omitzero" yaml:"rotation,omitempty"`
	// Translation declares the position of the component as a linear translation.
	// The translation parameters are string expressions which can be evaluated to generate a
	// CompPoseTranslSpec.
	Translation AssmCompPoseTranslExprSpec `json:"translation,omitzero" yaml:"translation,omitempty"`
}

// AssmCompPoseSpec declares a Optikit design's component's geometry.
// A zero value indicates that the component has no geometric pose.
type AssmCompPoseSpec struct {
	// Kind declares the kind of geometric transformation for the component's pose. It can be either
	// `` (equivalent to `affine`), `affine` (for a standard affine transformation relative to the
	// base coordinate system), or `relative-position` (in which the component is rotated with respect
	// to the assembly's basis vectors and then translated from the base's origin).
	Kind string `json:"kind,omitzero" yaml:"kind,omitempty"`
	// Base declares the base coordinate system for the geometric transformation which sets the
	// component's pose.
	Base AssmCompPoseBaseSpec `json:"base,omitzero" yaml:"base,omitempty"`
	// Rotation declares the orientation of the component as a rotation.
	Rotation AssmCompPoseRotSpec `json:"rotation,omitzero" yaml:"rotation,omitempty"`
	// Translation declares the position of the component as a linear translation.
	Translation AssmCompPoseTranslSpec `json:"translation,omitzero" yaml:"translation,omitempty"`
}

const (
	AssmCompPoseKindAffine           = "affine"
	AssmCompPoseKindRelativePosition = "relative-position"
)

// AssmCompPoseBaseSpec declares the coordinate system used as the base for the rotation or
// translation. The rotation or translation will be relative to the base's rotation or
// translation.
type AssmCompPoseBaseSpec struct {
	// Kind is the type of base. It can be either `` (equivalent to `parent`), `parent` (to use the
	// component's parent's coordinate system as the base), `assembly` (to use the overall assembly's
	// coordinate system as the base) or `mate` (to use the coordinate system of a specified mate
	// component of the component's parent as the base). If the component has no parent component,
	// then its parent is automatically just the parent assembly.
	Kind string `json:"kind,omitzero" yaml:"kind,omitempty"`
}

const (
	AssmCompPoseBaseKindParent   = "parent"
	AssmCompPoseBaseKindAssembly = "assembly"
	AssmCompPoseBaseKindMate     = "mate"
)

// AssmCompPoseRotExprSpec declares the orientation of the component as a rotation relative to the
// specified base's orientation.
// The pose parameters are string expressions which can be evaluated to generate a CompPoseRotSpec.
type AssmCompPoseRotExprSpec struct {
	// Kind is the type of orientation of the component. It can be either `` (implying a component
	// without any rotation), `uc2` (implying a UC2 cube), `grid` (for any orientation
	// aligned with the design's axes, even if violating UC2 cube orientation constraints),
	// 'euler' (for arbitrary rotations in the extrinsic z-x-y Euler angle order), or
	// `quaternion` (for arbitrary rotations).
	// If the kind is uc2, then Grid.Z is only allowed to be +z or -z, and Grid.X is not allowed to
	// be +z or -z.
	Kind string `json:"kind,omitzero" yaml:"kind,omitempty"`
	// Grid declares the orientation parameters of the component if its rotation kind is `uc2` or
	// `grid`.
	Grid RotGridSpec `json:"grid,omitzero" yaml:"grid,omitempty"`
	// Euler declares the orientation parameters of the component if its rotation kind is
	// `euler`. Angles should be in the extrinsic Z-X-Y order, which is equivalent to the
	// intrinsic Y-X-Z order.
	Euler ExprXYZ `json:"euler,omitzero" yaml:"euler,omitempty"`
	// Quaternion declares the orientation parameters of the component if its rotation kind is
	// `quaternion`.
	// The quaternion should be a string expression which evaluates into a 4-component numeric array.
	Quaternion Expr `json:"quaternion,omitzero" yaml:"quaternion,omitempty"`
}

// AssmCompPoseRotSpec declares the orientation of the component as a rotation relative to the
// specified base's orientation.
type AssmCompPoseRotSpec struct {
	// Kind is the type of orientation of the component. It can be either `` (implying no rotation),
	// `uc2` (implying a UC2 cube), `grid` (for any orientation aligned with the design's axes, even
	// if violating UC2 cube orientation constraints), or `quaternion` (for arbitrary rotations).
	// If the kind is uc2, then Grid.Z is only allowed to be +z or -z, and Grid.X is not allowed to
	// be +z or -z.
	Kind string `json:"kind,omitzero" yaml:"kind,omitempty"`
	// Grid declares the orientation parameters of the component if its rotation kind is `uc2` or
	// `grid`.
	Grid RotGridSpec `json:"grid,omitzero" yaml:"grid,omitempty"`
	// Euler declares the orientation parameters of the component if its rotation kind is
	// `euler`. Angles should be in the extrinsic Z-X-Y order, which is equivalent to the
	// intrinsic Y-X-Z order.
	Euler ContinuousXYZ[float64] `json:"euler,omitzero" yaml:"euler,omitempty"`
	// Quaternion declares the orientation parameters of the component if its rotation kind is
	// `quaternion`.
	Quaternion quaternion.T `json:"quaternion,omitzero" yaml:"quaternion,omitzero"`
}

const (
	RotKindUC2        = "uc2"
	RotKindGrid       = "grid"
	RotKindEuler      = "euler"
	RotKindQuaternion = "quaternion"
)

// AssmCompPoseTranslExprSpec declares the position of the component as linear translation relative to
// an "anchor" component, as an x-y-z offset along the specified base's coordinate axes.
// The pose parameters are string expressions which can be evaluated to generate a
// CompPoseTranslSpec.
type AssmCompPoseTranslExprSpec struct {
	// OffsetGrid is an offset from the base's position towards the component's position, in the
	// design's coordinate axes.
	OffsetGrid ExprXYZ `json:"offset-grid,omitzero" yaml:"offset-grid,omitempty"`
	// OffsetMM is an additional offset from the base's position towards the component's position,
	// in millimeters, after first applying the grid offset.
	OffsetMM ExprXYZ `json:"offset-mm,omitzero" yaml:"offset-mm,omitempty"`
}

// AssmCompPoseTranslSpec declares the position of the component as linear translation relative to
// an "anchor" component, as an x-y-z offset along the specified base's coordinate axes.
type AssmCompPoseTranslSpec struct {
	// OffsetGrid is an offset from the base's position towards the component's position, in the
	// design's coordinate axes.
	OffsetGrid DiscreteXYZ[int] `json:"offset-grid,omitzero" yaml:"offset-grid,omitempty"`
	// OffsetMM is an additional offset from the base's position towards the component's position,
	// in millimeters, after first applying the grid offset.
	OffsetMM ContinuousXYZ[float64] `json:"offset-mm,omitzero" yaml:"offset-mm,omitempty"`
}

// DesignExprDecl

// LoadDesignExprDecl loads a DesignExprDecl from the specified file path in the provided base
// filesystem.
func LoadDesignExprDecl(
	ctx context.Context, fsys ffs.PathedFS, filePath string,
) (DesignExprDecl, error) {
	b, err := fs.ReadFile(fsys, filePath)
	if err != nil {
		return DesignExprDecl{}, errors.Wrapf(
			err, "couldn't read design config file %s/%s", fsys.Path(), filePath,
		)
	}
	config := DesignExprDecl{}
	decoder := yaml.NewDecoder(bytes.NewReader(b), customYAMLUnmarshalers()...)
	if err = decoder.DecodeContext(ctx, &config); err != nil {
		return DesignExprDecl{}, errors.Wrap(err, "couldn't parse design declaration with expressions")
	}
	return config, nil
}

func customYAMLUnmarshalers() []yaml.DecodeOption {
	return []yaml.DecodeOption{
		yaml.CustomUnmarshalerContext(yamlUnmarshalCompPrimStaticModelsSpec),
	}
}

// Check looks for errors in the construction of the design configuration.
func (d DesignExprDecl) Check() (errs []error) {
	errs = append(errs, errsWrap(d.Design.Check(), "invalid design spec")...)
	errs = append(errs, errsWrap(d.Components.Check(), "invalid components spec")...)
	errs = append(errs, errsWrap(d.Assemblies.Check(), "invalid assemblies spec")...)
	return errs
}

// Cloned returns a deep copy of the DesignExprDecl.
func (d DesignExprDecl) Cloned() DesignExprDecl {
	d.Design.Tags = slices.Clone(d.Design.Tags)
	d.Inputs = d.Inputs.Cloned()
	d.Components = d.Components.Cloned()
	d.Assemblies = d.Assemblies.Cloned()
	return d
}

// Instantiated returns a CompExprsSpec which has been modified with design variants, input
// variables, and feature flags, as specified by the provided instantiation parameters.
func (d DesignExprDecl) Instantiated(instantiation InstSpec) (dd DesignDecl, err error) {
	dd.Inputs = d.Inputs

	inputEnv, err := MakeExprEnv(instantiation.Inputs, dd.Inputs)
	if err != nil {
		return dd, errors.Wrapf(
			err, "couldn't make expression env with inputs %+v", instantiation.Inputs,
		)
	}
	if dd.Components, err = d.Components.Evaluated(inputEnv); err != nil {
		return dd, errors.Wrapf(err, "couldn't evaluate expressions with inputs %+v", inputEnv)
	}
	if dd.Assemblies, err = d.Assemblies.Evaluated(inputEnv); err != nil {
		return dd, errors.Wrapf(err, "couldn't evaluate expressions with inputs %+v", inputEnv)
	}
	return dd, nil
}

// DesignDecl

// LoadDesignDecl loads a DesignExprDecl from the specified file path in the provided base
// filesystem.
func LoadDesignDecl(ctx context.Context, fsys ffs.PathedFS, filePath string) (DesignDecl, error) {
	b, err := fs.ReadFile(fsys, filePath)
	if err != nil {
		return DesignDecl{}, errors.Wrapf(
			err, "couldn't read design config file %s/%s", fsys.Path(), filePath,
		)
	}
	config := DesignDecl{}
	decoder := yaml.NewDecoder(bytes.NewReader(b), customYAMLUnmarshalers()...)
	if err = decoder.DecodeContext(ctx, &config); err != nil {
		return DesignDecl{}, errors.Wrap(err, "couldn't parse design declaration")
	}
	return config, nil
}

// Check looks for errors in the construction of the design configuration.
func (d DesignDecl) Check() (errs []error) {
	errs = append(errs, errsWrap(d.Design.Check(), "invalid design spec")...)
	// TODO: make the components check account for declared input variables
	errs = append(errs, errsWrap(d.Components.Check(), "invalid components spec")...)
	// TODO: make the assemblies check account for declared input variables
	errs = append(errs, errsWrap(d.Assemblies.Check(), "invalid assemblies spec")...)
	return errs
}

// Cloned returns a deep copy of the DesignDecl.
func (d DesignDecl) Cloned() DesignDecl {
	d.Design.Tags = slices.Clone(d.Design.Tags)
	d.Inputs = d.Inputs.Cloned()
	d.Components = d.Components.Cloned()
	d.Assemblies = d.Assemblies.Cloned()
	return d
}

// DesignSpec

// Check looks for errors in the construction of the design spec.
func (s DesignSpec) Check() (errs []error) {
	return errs
}

// Tags

// Merged returns a new Tags created by applying the specified overlay, without modifying this
// current Tags or the overlay.
func (t Tags) Merged(overlay Tags) Tags {
	set := make(structures.Set[string])
	merged := make(Tags, 0, len(t)+len(overlay))
	for _, tag := range t {
		set.Add(tag)
		merged = append(merged, tag)
	}
	for _, tag := range overlay {
		if set.Has(tag) {
			continue
		}

		set.Add(tag)
		merged = append(merged, tag)
	}
	return merged
}
