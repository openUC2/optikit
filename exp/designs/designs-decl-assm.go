package designs

import (
	"cmp"
	"fmt"
	"maps"
	"math"
	"slices"

	"github.com/pkg/errors"
	"github.com/ungerik/go3d/float64/mat3"
	"github.com/ungerik/go3d/float64/mat4"
	"github.com/ungerik/go3d/float64/vec3"

	"github.com/openuc2/optikit/exp/structures"
)

// AssmExprsSpec

// Check looks for errors in the construction of the assemblies spec.
func (s AssmExprsSpec) Check() (errs []error) {
	for assmID, assm := range s {
		assmErrs := assm.Check()
		for _, err := range assmErrs {
			errs = append(errs, errors.Wrapf(err, "checks failed for assembly %s", assmID))
		}
	}
	return errs
}

// Cloned returns a deep copy of the AssmExprsSpec.
func (s AssmExprsSpec) Cloned() AssmExprsSpec {
	result := make(AssmExprsSpec)
	for variantID, spec := range s {
		result[variantID] = spec.Cloned()
	}
	return result
}

// Merged returns a new AssmExprsSpec created by applying the specified overlay, without modifying
// this current AssmExprsSpec or the overlay.
/*func (s AssmExprsSpec) Merged(overlay AssmExprsSpec) AssmExprsSpec {
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
func (s AssmExprsSpec) Evaluated(env ExprEnv) (result AssmsSpec, err error) {
	result = make(AssmsSpec)
	for id, c := range s {
		if result[id], err = c.Evaluated(env); err != nil {
			return nil, errors.Wrapf(err, "couldn't evaluate expressions in assembly %s", id)
		}
	}
	return result, nil
}

// AssmsSpec

// Check looks for errors in the construction of the assemblies spec.
func (s AssmsSpec) Check() (errs []error) {
	for assmID, assm := range s {
		assmErrs := assm.Check()
		for _, err := range assmErrs {
			errs = append(errs, errors.Wrapf(err, "checks failed for assembly %s", assmID))
		}
	}
	return errs
}

// Cloned returns a deep copy of the AssmsSpec.
func (s AssmsSpec) Cloned() AssmsSpec {
	result := make(AssmsSpec)
	for variantID, spec := range s {
		result[variantID] = spec.Cloned()
	}
	return result
}

// Flattened returns a new AssmsSpec in which, for each assembly, each non-origin component's pose
// base is just the root coordinate system of the assembly itself, and all components (besides
// components dependent on mate connectors) are direct children of the assembly itself.
// Components dependent on mate connectors are kept as children of their parent components, but
// their component subtree is flattened.
func (s AssmsSpec) Flattened(gridSpacings ContinuousXYZ[float64]) (flattened AssmsSpec, err error) {
	flattened = s.Cloned()
	for assmID, assm := range s {
		if assm.Children, err = assm.Children.Flattened(gridSpacings); err != nil {
			return nil, errors.Wrapf(err, "couldn't flatten assembly %s", assmID)
		}
		flattened[assmID] = assm
	}
	return flattened, nil
}

// AssmExprSpec

// Check looks for errors in the construction of the assembly spec.
func (s AssmExprSpec) Check() (errs []error) {
	// TODO: implement
	return errs
}

// Cloned returns a deep copy of the AssmExprSpec.
func (s AssmExprSpec) Cloned() AssmExprSpec {
	return AssmExprSpec{
		Description: s.Description,
		Children:    s.Children.Cloned(),
		Tags:        slices.Clone(s.Tags),
	}
}

// Evaluated evaluates the expressions with the given ExprEnv into an AssmSpec.
func (s AssmExprSpec) Evaluated(env ExprEnv) (result AssmSpec, err error) {
	result = AssmSpec{
		Tags: s.Tags,
	}
	if result.Children, err = s.Children.Evaluated(env); err != nil {
		return result, errors.Wrap(err, "couldn't evaluate expressions in children section")
	}
	return result, nil
}

// AssmSpec

// Check looks for errors in the construction of the assembly spec.
func (s AssmSpec) Check() (errs []error) {
	// TODO: check that no component is included multiple times at multiple places in the tree
	return errs
}

// Cloned returns a deep copy of the AssmSpec.
func (s AssmSpec) Cloned() AssmSpec {
	return AssmSpec{
		Description: s.Description,
		Children:    s.Children.Cloned(),
		Tags:        slices.Clone(s.Tags),
	}
}

// AssmCompExprsSpec

// Cloned returns a deep copy of the AssmCompExprsSpec.
func (s AssmCompExprsSpec) Cloned() AssmCompExprsSpec {
	result := make(AssmCompExprsSpec)
	for compID, spec := range s {
		result[compID] = spec.Cloned()
	}
	return result
}

// Evaluated evaluates the parameter expressions with the given ExprEnv into a CompsSpec.
func (s AssmCompExprsSpec) Evaluated(env ExprEnv) (result AssmCompsSpec, err error) {
	result = make(AssmCompsSpec)
	for id, c := range s {
		if result[id], err = c.Evaluated(env); err != nil {
			return nil, errors.Wrapf(err, "couldn't evaluate expressions in component %s", id)
		}
	}
	return result, nil
}

// AssmCompsSpec

// Cloned returns a deep copy of the AssmCompsSpec.
func (s AssmCompsSpec) Cloned() AssmCompsSpec {
	result := make(AssmCompsSpec)
	for compID, spec := range s {
		result[compID] = spec.Cloned()
	}
	return result
}

type PoseDigraph = structures.StrictEdgeDigraph[CompID, AssmCompSpec]

// PoseDigraph returns a StrictEdgeDigraph representing the tree structure of the assembly.
// It returns an error if any component ID appears in multiple nodes of the tree.
func (s AssmCompsSpec) PoseDigraph() (PoseDigraph, error) {
	g := make(PoseDigraph)
	g.AddNode("") // origin
	return s.poseDigraph(g, "", s)
}

func (s AssmCompsSpec) poseDigraph(
	g PoseDigraph, rootID CompID, children AssmCompsSpec,
) (modified PoseDigraph, err error) {
	for childID, child := range children {
		if _, ok := g[childID]; ok {
			return g, errors.Errorf(
				"child %s of parent %s was already added to the pose digraph", childID, rootID,
			)
		}
		if child.Disabled {
			continue
		}

		g.AddNode(childID)
		g.AddEdge(rootID, childID, child)
	}
	for childID, child := range children {
		if child.Disabled {
			continue
		}

		if g, err = s.poseDigraph(g, childID, child.Children); err != nil {
			return g, errors.Wrapf(err, "couldn't recurse into subtree rooted at %s", childID)
		}
	}
	return g, nil
}

// Subtrees gathers all direct and indirect descendants of the assembly into a single map keyed
// by those descendants' component IDs. The descendants' declared poses and children are left
// unmodified, unlike the Flattened method. In other words, the returned map includes every subtree
// of the AssmCompsSpec, keyed by the component ID of the root of that subtree.
// The subtrees are cloned, so they cannot be used to restructure or edit the assembly tree.
func (s AssmCompsSpec) Subtrees() map[CompID]AssmCompSpec {
	collected := s.Cloned()
	for _, assmComp := range s {
		if assmComp.Disabled {
			continue
		}

		maps.Insert(collected, maps.All(assmComp.Children.Subtrees()))
	}
	return collected
}

// EditableSubtrees returns a similar output as Subtrees, except the values can be used to
// restructure/edit the assembly tree.
func (s AssmCompsSpec) EditableSubtrees() (map[CompID]AssmCompSpec, error) {
	subtrees := make(map[CompID]AssmCompSpec)
	g, err := s.PoseDigraph()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't construct pose digraph from tree")
	}
	nextParents := make([]CompID, 0, len(g))
	nextParents = append(nextParents, "")
	for len(nextParents) > 0 {
		parent := nextParents[0]
		nextParents = nextParents[1:]
		for child := range g[parent] {
			subtrees[child] = g[parent][child]
			nextParents = append(nextParents, child)
		}
	}

	return subtrees, nil
}

// Flattened returns a new AssmCompsSpec in which each non-origin component's pose base
// is just the root coordinate system of the assembly itself, and all components (besides
// components dependent on mate connectors) are direct children of the assembly itself.
// Components dependent on mate connectors are kept as children of their parent components, but
// their component subtree is flattened.
func (s AssmCompsSpec) Flattened(gridSpacings ContinuousXYZ[float64]) (AssmCompsSpec, error) {
	flattened := make(AssmCompsSpec)
	g, err := s.PoseDigraph()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't construct pose digraph from assembly")
	}

	nextParents := make([]CompID, 0, len(g)) // for breadth-first tree traversal
	nextParents = append(nextParents, "")    // add the root node
	for len(nextParents) > 0 {
		parent := nextParents[0]
		nextParents = nextParents[1:]
		for child := range g[parent] {
			if _, ok := flattened[child]; ok {
				return flattened, errors.Errorf(
					"child %s of parent %s was already added to the flattened assembly", child, parent,
				)
			}

			c := g[parent][child].Cloned()
			fc, err := flattenedChild(
				parent, flattened[parent].Pose, child, c.Cloned(), gridSpacings,
			)
			if err != nil {
				return flattened, errors.Wrapf(err, "couldn't flatten child %s of parent %s", child, parent)
			}
			if c.Pose.Base.Kind == AssmCompPoseBaseKindMate {
				flattened[parent].Children[child] = fc
				// Note: we treat this child as a leaf node for the breadth-first tree traversal because its
				// subtree has already been flattened.
				continue
			}

			flattened[child] = fc
			nextParents = append(nextParents, child)
		}
	}
	return flattened, nil
}

func flattenedChild(
	parentID CompID, parentPose AssmCompPoseSpec, childID CompID, child AssmCompSpec,
	gridSpacings ContinuousXYZ[float64],
) (flattened AssmCompSpec, err error) {
	var basePose AssmCompPoseSpec
	switch child.Pose.Base.Kind {
	default:
		return flattened, errors.Errorf("unknown pose base kind %s", child.Pose.Base.Kind)
	case AssmCompPoseBaseKindMate:
		if child.Children, err = child.Children.Flattened(gridSpacings); err != nil {
			return child, errors.Wrapf(
				err, "couldn't flatten subtree of component %s mated under %s", childID, parentID,
			)
		}
		for grandchild, gc := range child.Children {
			gc.Pose.Base = AssmCompPoseBaseSpec{Kind: AssmCompPoseBaseKindParent}
			child.Children[grandchild] = gc
		}
		return child, nil
	case "", AssmCompPoseBaseKindParent:
		if parentID == "" {
			child.Children = make(AssmCompsSpec)
			return child, nil
		}
		basePose = parentPose
	}

	switch child.Pose.Kind {
	default:
		return flattened, errors.Errorf("unknown pose kind %s", child.Pose.Kind)
	case "", AssmCompPoseKindAffine:
		if child.Pose, err = popPose(
			childID, parentID, child.Pose, parentPose, gridSpacings,
		); err != nil {
			return flattened, errors.Wrapf(
				err, "couldn't bring affine-type pose of child %s up to the assembly's frame", childID,
			)
		}
	case AssmCompPoseKindRelTransl:
		child.Pose.Translation = basePose.Translation.Added(child.Pose.Translation)
		child.Pose.Kind = ""
	}
	child.Pose.Base = AssmCompPoseBaseSpec{}
	child.Children = make(AssmCompsSpec)
	return child, nil
}

func popPose(
	childID, parentID CompID,
	childPose, parentPose AssmCompPoseSpec,
	gridSpacings ContinuousXYZ[float64],
) (AssmCompPoseSpec, error) {
	parentMat, err := parentPose.TransfMat(gridSpacings)
	if err != nil {
		return AssmCompPoseSpec{}, errors.Wrapf(
			err, "couldn't determine affine matrix of parent %s", parentID,
		)
	}
	mat, err := childPose.TransfMat(gridSpacings)
	if err != nil {
		return AssmCompPoseSpec{}, errors.Wrapf(err, "couldn't compute pose of child %s", childID)
	}
	result := &mat4.T{}
	result.AssignMul(&parentMat, &mat)
	return NewPose(*result, gridSpacings), nil
}

// Prefixed returns a copy of the AssmCompsSpec but with the specified parent CompID prefixed to all
// component IDs of all descendants.
func (s AssmCompsSpec) Prefixed(parent CompID) AssmCompsSpec {
	prefixed := make(AssmCompsSpec)
	for compID, comp := range s {
		c := comp.Cloned()
		c.Children = c.Children.Prefixed(parent)
		prefixed[JoinCompIDs(parent, compID)] = c
	}
	return prefixed
}

// Print prints the entire tree structure.
func (s AssmCompsSpec) Print(indent int) {
	for _, compID := range slices.Sorted(maps.Keys(s)) {
		for range indent {
			fmt.Printf("  ")
		}
		fmt.Printf("%s %s\n", compID, s[compID].Pose.ShortString())
		s[compID].Children.Print(indent + 1)
	}
}

// AssmCompExprSpec

// Cloned returns a deep copy of the CompExprSpec.
func (s AssmCompExprSpec) Cloned() AssmCompExprSpec {
	return AssmCompExprSpec{
		Disabled: s.Disabled,
		Assm:     s.Assm,
		Pose:     s.Pose,
		Children: s.Children.Cloned(),
	}
}

// Evaluated evaluates the expressions with the given ExprEnv into a AssmCompSpec.
func (s AssmCompExprSpec) Evaluated(env ExprEnv) (result AssmCompSpec, err error) {
	result = AssmCompSpec{}
	if result.Disabled, err = s.Disabled.evalAs[bool](env.ToMap()); err != nil {
		return result, errors.Wrap(err, "couldn't evaluate expression for whether to disable component")
	}
	if result.Assm, err = s.Assm.evalAsString[AssmID](env.ToMap()); err != nil {
		return result, errors.Wrap(err, "couldn't evaluate expression for assembly to include")
	}
	if result.Pose, err = s.Pose.Evaluated(env); err != nil {
		return result, errors.Wrap(err, "couldn't evaluate expressions in pose section")
	}
	if result.Children, err = s.Children.Evaluated(env); err != nil {
		return result, errors.Wrap(err, "couldn't evaluate expressions in children section")
	}
	return result, nil
}

// AssmCompSpec

// Cloned returns a deep copy of the AssmCompSpec.
func (s AssmCompSpec) Cloned() AssmCompSpec {
	return AssmCompSpec{
		Disabled: s.Disabled,
		Assm:     s.Assm,
		Pose:     s.Pose,
		Children: s.Children.Cloned(),
	}
}

func (s AssmCompSpec) MateChildren() AssmCompsSpec {
	mates := make(AssmCompsSpec)
	for id, comp := range s.Children {
		if comp.Pose.Base.Kind != AssmCompPoseBaseKindMate {
			continue
		}
		mates[id] = comp
	}
	return mates
}

// AssmCompPoseExprSpec

// Merged returns a new CompPoseExprSpec created by applying the specified overlay, without
// modifying this current CompsPoseExprSpec or the overlay.
/*func (s AssmCompPoseExprSpec) Merged(overlay AssmCompPoseExprSpec) AssmCompPoseExprSpec {
	return AssmCompPoseExprSpec{
		Rotation:    s.Rotation.Merged(overlay.Rotation),
		Translation: s.Translation.Merged(overlay.Translation),
	}
}*/

// Evaluated evaluates the pose expressions with the given ExprEnv into a CompPoseSpec.
func (s AssmCompPoseExprSpec) Evaluated(env ExprEnv) (result AssmCompPoseSpec, err error) {
	result = AssmCompPoseSpec{
		Kind: s.Kind,
		Base: s.Base,
	}
	if result.Rotation, err = s.Rotation.Evaluated(env); err != nil {
		return AssmCompPoseSpec{}, errors.Wrap(err, "couldn't evaluate rotation")
	}
	if result.Translation, err = s.Translation.Evaluated(env); err != nil {
		return AssmCompPoseSpec{}, errors.Wrap(err, "couldn't evaluate translation")
	}
	return result, nil
}

// AssmCompPoseSpec

// NewPose builds a new AssmCompPoseSpec from an affine transformation matrix with respect to the
// assembly's coordinate system.
// The translation component is decomposed into a discrete component (for any non-zero grid
// spacings) and any remaining non-discrete component.
func NewPose(mat mat4.T, gridSpacings ContinuousXYZ[float64]) AssmCompPoseSpec {
	return AssmCompPoseSpec{
		Kind:        AssmCompPoseKindAffine,
		Base:        AssmCompPoseBaseSpec{},
		Rotation:    NewPoseRot(mat),
		Translation: NewPoseTransl(mat, gridSpacings),
	}
}

func (s AssmCompPoseSpec) ShortString() string {
	baseKind := s.Base.Kind
	switch baseKind {
	case "", AssmCompPoseBaseKindParent:
		baseKind = ""
	case AssmCompPoseBaseKindMate:
		baseKind = fmt.Sprintf("mate: %s", s.Base.Mate)
	}

	kind := s.Kind
	switch kind {
	case "", AssmCompPoseKindAffine:
		return baseKind
	case AssmCompPoseKindRelTransl:
		kind = "rel-pos"
	}

	if baseKind == "" {
		return kind
	}
	return fmt.Sprintf("%s %s", kind, baseKind)
}

// TransfMat returns a homogeneous affine transformation matrix representing the pose of the
// component relative to the frame of the base.
// The pose must be of kind `affine` (or “, which is equivalent to `affine`), or else an error will
// be returned.
// The translation component of the matrix is in mm.
// This is the matrix H^a_b for homogeneous pose vectors p^a_h and p^b_h, which are homogeneous
// representations of vectors p^a and p^b, where p^b is in the frame of the component and p^b is in
// the frame of the overall assembly. In other words, this matrix can be multiplied with a point in
// the frame of the component to get the position of that point in the frame of the overall
// assembly.
func (s AssmCompPoseSpec) TransfMat(gridSpacings ContinuousXYZ[float64]) (mat4.T, error) {
	switch s.Kind {
	default:
		return mat4.T{}, errors.Errorf("unsupported pose kind %s", s.Kind)
	case "", AssmCompPoseKindAffine:
		// proceed
	}
	m := s.Rotation.TransfMat()
	offsetGrid := AsMM(s.Translation.OffsetGrid, gridSpacings).AsVec3()
	offsetMM := s.Translation.OffsetMM.AsVec3()
	translation := vec3.Add(&offsetGrid, &offsetMM)
	m.SetTranslation(&translation)
	return m, nil
}

// AssmCompPoseRotExprSpec

// Merged returns a new CompPoseRotExprSpec created by applying the specified overlay, without
// modifying this current CompsPoseExprSpec or the overlay.
/*func (s AssmCompPoseRotExprSpec) Merged(overlay AssmCompPoseRotExprSpec) AssmCompPoseRotExprSpec {
	t := cmp.Or(overlay.Kind, s.Kind)
	switch t {
	default:
		return AssmCompPoseRotExprSpec{}
	case RotKindUC2, RotKindGrid:
		return AssmCompPoseRotExprSpec{
			Kind: t,
			Grid: s.Grid.Merged(overlay.Grid),
		}
	case RotKindEuler:
		return AssmCompPoseRotExprSpec{
			Kind:  t,
			Euler: s.Euler.Merged(overlay.Euler),
		}
	case RotKindQuaternion:
		return AssmCompPoseRotExprSpec{
			Kind:       t,
			Quaternion: cmp.Or(overlay.Quaternion, s.Quaternion),
		}
	}
}*/

// Evaluated evaluates the pose expressions with the given ExprEnv into a CompPoseRotSpec.
// Parameters not associated with the CompPoseRotExprSpec's kind are excluded from the result; for
// example, if the rotation kind is "quaternion", then the result's Grid field will be zero.
func (s AssmCompPoseRotExprSpec) Evaluated(env ExprEnv) (result AssmCompPoseRotSpec, err error) {
	result.Kind = s.Kind
	switch result.Kind {
	case "":
		result = AssmCompPoseRotSpec{}
	case RotKindUC2, RotKindGrid:
		result.Grid = s.Grid
	case RotKindEuler:
		if result.Euler, err = s.Euler.EvaluatedFloat64(env.ToMap()); err != nil {
			return AssmCompPoseRotSpec{}, errors.Wrap(err, "couldn't evaluate euler")
		}
	case RotKindQuaternion:
		if s.Quaternion != "" {
			evaluated, err := s.Quaternion.evalAsAny(env.ToMap())
			if err != nil {
				return AssmCompPoseRotSpec{}, errors.Wrapf(
					err, "couldn't evaluate quat expr %s", s.Quaternion,
				)
			}
			if result.Quaternion, err = convertToQuaternion(evaluated); err != nil {
				return AssmCompPoseRotSpec{}, errors.Wrapf(
					err, "evaluated quat expr %s as array %+v, but couldn't convert it to a quaternion",
					s.Quaternion, evaluated,
				)
			}
		}
	}
	return result, nil
}

// AssmCompPoseRotSpec

// NewPoseRot builds a CompPoseRotSpec from a transformation matrix. If the transformation matrix
// specifies an axis-aligned rotation, then the result will be of kind "grid" (note: it will never
// be of kind "uc2"). Otherwise, the result will be of kind "quaternion".
func NewPoseRot(mat mat4.T) AssmCompPoseRotSpec {
	z := mat.MulVec3(&vec3.UnitZ)
	zDir, zAxisAligned := BasisDirs[z]
	x := mat.MulVec3(&vec3.UnitX)
	xDir, xAxisAligned := BasisDirs[x]
	if zAxisAligned && xAxisAligned {
		return AssmCompPoseRotSpec{
			Kind: RotKindGrid,
			Grid: RotGridSpec{
				Z: zDir,
				X: xDir,
			},
		}
	}
	rotMat := mat3.T{}
	for col := 0; col < rotMat.Cols(); col++ {
		for row := 0; row < rotMat.Rows(); row++ {
			rotMat[col][row] = mat.Get(col, row)
		}
	}
	return AssmCompPoseRotSpec{
		Kind:   RotKindMatrix,
		Matrix: rotMat,
	}
}

// Check looks for errors in the construction of the component orientation spec.
func (s AssmCompPoseRotSpec) Check() (errs []error) {
	switch s.Kind {
	default:
		return []error{errors.Errorf("invalid rotation kind: %s", s.Kind)}
	case "":
		return nil
	case RotKindUC2:
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
	case RotKindGrid:
		return s.Grid.Check()
	case RotKindQuaternion:
		const tolerance = 1e-6
		if !s.Quaternion.IsUnitQuat(tolerance) {
			return append(errs, errors.Errorf("quaternion is not a unit quaternion: %+v", s.Quaternion))
		}
		return nil
	case RotKindMatrix:
		const tolerance = 1e-6
		determinant := s.Matrix.Determinant()
		if determinant < (1.0-tolerance) || determinant > (1.0+tolerance) {
			return append(errs, errors.Errorf("matrix does not have determinant of 1: %+v", s.Quaternion))
		}
		return nil
	}
}

// TransfMat returns a homogeneous transformation matrix representing the orientation of the
// component relative to the frame of the design. If the rotation kind is empty, then it'll return
// the identity; if the rotation kind is unknown, then it'll return zero; otherwise, it assumes that the
// component orientation spec is valid.
// The first column is the component's x-axis, represented in the coordinate system of the overall
// design. The second and third columns are the y- and z-axes, respectively.
func (s AssmCompPoseRotSpec) TransfMat() mat4.T {
	switch s.Kind {
	default:
		return mat4.Zero
	case "":
		return mat4.Ident
	case RotKindUC2, RotKindGrid:
		return GridRotMats[cmp.Or(s.Grid.Z, DirZPos)][cmp.Or(s.Grid.X, DirXPos)]
	case RotKindEuler:
		mat := mat4.Zero
		mat.AssignEulerRotation(
			degToRad(s.Euler.Y), degToRad(s.Euler.X), degToRad(s.Euler.Z),
		)
		return mat
	case RotKindQuaternion:
		mat := mat4.Zero
		mat.AssignQuaternion(&s.Quaternion)
		return mat
	case RotKindMatrix:
		mat := mat4.Ident
		for col := 0; col < s.Matrix.Cols(); col++ {
			for row := 0; row < s.Matrix.Rows(); row++ {
				mat[col][row] = s.Matrix.Get(col, row)
			}
		}
		return mat
	}
}

func degToRad(deg float64) float64 {
	return deg * (math.Pi / 180.0) //nolint:mnd // the entire function is a magic number conversion...
}

// AssmCompPoseTranslExprSpec

// Evaluated evaluates the pose expressions with the given ExprEnv into a CompPoseSpec.
func (s AssmCompPoseTranslExprSpec) Evaluated(
	env ExprEnv,
) (result AssmCompPoseTranslSpec, err error) {
	result = AssmCompPoseTranslSpec{}
	if result.OffsetGrid, err = s.OffsetGrid.EvaluatedInt(env.ToMap()); err != nil {
		return AssmCompPoseTranslSpec{}, errors.Wrap(err, "couldn't evaluate offsetGrid")
	}
	if result.OffsetMM, err = s.OffsetMM.EvaluatedFloat64(env.ToMap()); err != nil {
		return AssmCompPoseTranslSpec{}, errors.Wrap(err, "couldn't evaluate offsetMM")
	}
	return result, nil
}

// Merged returns a new CompPoseTranslSpec created by applying the specified overlay, without modifying
// this current CompsPoseSpec or the overlay.
/*func (s AssmCompPoseTranslExprSpec) Merged(
	overlay AssmCompPoseTranslExprSpec,
) AssmCompPoseTranslExprSpec {
	return AssmCompPoseTranslExprSpec{
		Base:       cmp.Or(overlay.Base, s.Base),
		OffsetGrid: s.OffsetGrid.Merged(overlay.OffsetGrid),
		OffsetMM:   s.OffsetMM.Merged(overlay.OffsetMM),
	}
}*/

// AssmCompPoseTranslSpec

// NewPoseTransl builds a new CompPoseTranslSpec from a transformation matrix.
// The translation component is decomposed into a discrete component (for any non-zero grid
// spacings) and any remaining non-discrete component.
func NewPoseTransl(mat mat4.T, gridSpacings ContinuousXYZ[float64]) AssmCompPoseTranslSpec {
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
	return AssmCompPoseTranslSpec{
		OffsetGrid: gridded,
		OffsetMM:   mm,
	}
}

// Merged returns a new CompPoseTranslSpec created by applying the specified overlay, without
// modifying this current CompsPoseSpec or the overlay.
/*func (s AssmCompPoseTranslSpec) Merged(overlay AssmCompPoseTranslSpec) AssmCompPoseTranslSpec {
	return AssmCompPoseTranslSpec{
		Base:       cmp.Or(overlay.Base, s.Base),
		OffsetGrid: s.OffsetGrid.Merged(overlay.OffsetGrid),
		OffsetMM:   s.OffsetMM.Merged(overlay.OffsetMM),
	}
}*/

// String returns an abbreviated representation of the CompPoseTranslSpec.
func (s AssmCompPoseTranslSpec) String() string {
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
// It assumes that the pose Bases and basis vectors are the same between the two
// CompPoseTranslSpecs.
func (s AssmCompPoseTranslSpec) Added(t AssmCompPoseTranslSpec) AssmCompPoseTranslSpec {
	return AssmCompPoseTranslSpec{
		OffsetGrid: s.OffsetGrid.Added(t.OffsetGrid),
		OffsetMM:   s.OffsetMM.Added(t.OffsetMM),
	}
}
