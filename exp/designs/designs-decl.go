package designs

import (
	"cmp"
	"fmt"
	"io/fs"

	"github.com/goccy/go-yaml"
	"github.com/pkg/errors"
	"github.com/ungerik/go3d/float64/mat4"
	"github.com/ungerik/go3d/float64/vec3"

	ffs "github.com/openUC2/optikit/exp/fs"
	"github.com/openUC2/optikit/exp/structures"
)

// DesignDeclFile is the name of the file defining each Optikit design.
const DesignDeclFile = "optikit-design.yml"

// A DesignDecl declares an Optikit design.
type DesignDecl struct {
	// Optikit indicates that the design was written assuming the semantics of a given version
	// of Optikit. The version must be a valid Optikit version, and it sets the minimum version of
	// Optikit required to use the design. The Optikit tool refuses to use designs declaring newer
	// Optikit versions for any operations beyond printing information. The Optikit version of the
	// design must be greater than or equal to the Optikit version of every required Optikit design.
	Optikit string `yaml:"optikit-version"`
	// Design defines the basic metadata for the design.
	Design DesignSpec `yaml:"design,omitempty"`
	// Components declares the design's constituent components as a mapping from the ID of each
	// component to the declaration of that component.
	Components CompsSpec `yaml:"components,omitempty"`
}

// DesignSpec defines the basic metadata for an Optikit design.
type DesignSpec struct {
	// Path is the design path, which acts as the canonical name for the design.
	Path string `yaml:"path,omitempty"`
	// Description is a short description of the design to be shown to users.
	Description string `yaml:"description,omitempty"`
	// Tags is a list of human-readable string tags for describing the design to software.
	Tags []string `yaml:"tags,omitempty"`
}

type (
	CompID    string
	CompsSpec map[CompID]CompSpec
)

// CompSpec declares a component of an Optikit design.
type CompSpec struct {
	// Type is the type of component in the design. It can be either `location` or `design`.
	Type string `yaml:"type"`
	// Design is the path of the design which the component (of type `design`) instantiates. If it's
	// specified as an absolute path, then it will be relative to the root directory of the Optikits
	// design.
	Design string `yaml:"design,omitempty"`
	// Pose declares the geometry of the component.
	Pose CompPoseSpec `yaml:"pose,omitempty"`
	// Tags is a list of human-readable string tags for describing the component to software.
	Tags []string `yaml:"tags,omitempty"`
}

// CompPoseSpec defines declares a Optikit design's component's geometry.
// A zero value indicates that the component has no geometric pose.
type CompPoseSpec struct {
	// Rotation declares the orientation of the component as a rotation.
	Rotation CompPoseRotSpec `yaml:"rotation,omitempty"`
	// Translation declares the position of the component as a linear translation.
	Translation CompPoseTranslSpec `yaml:"translation,omitempty"`
}

// CompPoseRotSpec declares the orientation of the component as a rotation relative to the overall
// design's orientation.
type CompPoseRotSpec struct {
	// Type is the type of orientation of the component. It can be either `` (implying a component
	// without any spatial geometry), `uc2` (implying a UC2 cube), or `grid` (for any orientation
	// aligned with the design's axes, even if violating UC2 cube orientation constraints).
	// If the type is uc2, then Grid.Z is only allowed to be +z or -z, and Grid.X is not allowed to
	// be +z or -z.
	Type string `yaml:"type"`
	// Grid declares the orientation parameters of the component if its rotation type is `uc2` or
	// `grid`.
	Grid CompPoseRotGridSpec `yaml:"grid,omitempty"`
}

const (
	RotTypeUC2  = "uc2"
	RotTypeGrid = "grid"
)

// CompPoseRotGridSpec specifies the component's orientation relative to the design's orientation by
// two discrete parameters: the orientation of the component's z-axis, and the orientation of the
// component's x-axis.
// The component's y-axis is derived from the component's x- and z-axes via the right-hand rule.
type CompPoseRotGridSpec struct {
	// Z specifies the axis of the design's coordinate system which the component's coordinate
	// system's +z direction should point in. The zero value is interpreted as +z.
	Z string `yaml:"z,omitempty"`
	// X specifies the axis of the design's coordinate system which the component's coordinate
	// system's +x direction should point in. The zero value is interpreted as +x.
	X string `yaml:"x,omitempty"`
}

// CompPoseTranslSpec declares the position of the component as linear translation relative to an
// "anchor" component, as an x-y-z offset along the overall design's coordinate axes.
type CompPoseTranslSpec struct {
	// Anchor is the ID of the component whose position will be linearly translated by the specified
	// offsets in order to determine the position of this component.
	// If empty, it will be the origin of the overall design's coordinate axes.
	Anchor CompID `yaml:"anchor,omitempty"`
	// OffsetGrid is an offset from the anchor's position towards the component's position, in the
	// design's coordinate axes.
	OffsetGrid DiscreteXYZ[int] `yaml:"offset-grid,omitempty"`
	// OffsetCM is an additional offset from the anchor's position towards the component's position,
	// in centimeters, after first applying the grid offset.
	OffsetCM ContinuousXYZ[float64] `yaml:"offset-cm,omitempty"`
}

// DesignDecl

// LoadDesignDecl loads a DesignDecl from the specified file path in the provided base filesystem.
func LoadDesignDecl(fsys ffs.PathedFS, filePath string) (DesignDecl, error) {
	bytes, err := fs.ReadFile(fsys, filePath)
	if err != nil {
		return DesignDecl{}, errors.Wrapf(
			err, "couldn't read design config file %s/%s", fsys.Path(), filePath,
		)
	}
	config := DesignDecl{}
	if err = yaml.Unmarshal(bytes, &config); err != nil {
		return DesignDecl{}, errors.Wrap(err, "couldn't parse design config")
	}
	return config, nil
}

// Check looks for errors in the construction of the design configuration.
func (d DesignDecl) Check() (errs []error) {
	errs = append(errs, errsWrap(d.Design.Check(), "invalid design spec")...)
	errs = append(errs, errsWrap(d.Components.Check(), "invalid components spec")...)
	return errs
}

// DesignSpec

// Check looks for errors in the construction of the design spec.
func (s DesignSpec) Check() (errs []error) {
	return errs
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

// Flattened returns a new CompsSpec in which each non-origin component's translation anchor is just
// the root (origin) node.
// It assumes that the CompsSpec does not have any errors such as a nonexistent translation anchor
// required by a CompPosesTranslSpec.
func (s CompsSpec) Flattened() CompsSpec {
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

// CompPoseSpec

// TransfMat returns a homogeneous affine transformation matrix representing the pose of the
// component relative to the frame of the overall design, but only if the pose's translation is
// specified with the overall design's coordinate system's origin as the anchor. If anything else is
// the anchor, then this method returns an error instead.
// The translation component of the matrix is in cm.
// This is the matrix H^a_b for homogeneous pose vectors p^a_h and p^b_h, which are homogeneous
// representations of vectors p^a and p^b, where p^b is in the frame of the component and p^b is in
// the frame of the overall design. In other words, this matrix can be multiplied with a point in
// the frame of the component to get the position of that point in the frame of the overall design.
func (s CompPoseSpec) TransfMat(gridSpacings ContinuousXYZ[float64]) (mat4.T, error) {
	if s.Translation.Anchor != "" {
		return mat4.Zero, errors.New("translation anchor is not the overall design's origin!")
	}
	m := s.Rotation.TransfMat()
	offsetGrid := s.Translation.OffsetGrid.AsCM(gridSpacings).AsVec3()
	offsetCM := s.Translation.OffsetCM.AsVec3()
	translation := vec3.Add(&offsetGrid, &offsetCM)
	m.SetTranslation(&translation)
	return m, nil
}

// CompPoseRotSpec

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
	}
}

// CompPoseRotGridSpec

func (s CompPoseRotGridSpec) Check() (errs []error) {
	if s.Z[1] == s.X[1] {
		errs = append(errs, errors.Errorf("component's z and x axes are coaxial: z=%s, x=%s", s.Z, s.X))
	}
	return errs
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
	}
}

// CompPoseTranslSpec

func (s CompPoseTranslSpec) String() string {
	switch {
	case s.OffsetGrid == gridZero && s.OffsetCM == cmZero:
		return ""
	case s.OffsetGrid == gridZero:
		return fmt.Sprintf("%s cm", s.OffsetCM.String())
	case s.OffsetCM == cmZero:
		return s.OffsetGrid.String()
	default:
		return fmt.Sprintf("%s + %s cm", s.OffsetGrid.String(), s.OffsetCM.String())
	}
}

var (
	gridZero DiscreteXYZ[int]
	cmZero   ContinuousXYZ[float64]
)

func (s CompPoseTranslSpec) Added(t CompPoseTranslSpec) CompPoseTranslSpec {
	return CompPoseTranslSpec{
		Anchor:     s.Anchor,
		OffsetGrid: s.OffsetGrid.Added(t.OffsetGrid),
		OffsetCM:   s.OffsetCM.Added(t.OffsetCM),
	}
}
