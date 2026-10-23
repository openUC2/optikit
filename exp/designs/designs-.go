// Package designs implements the Optikit designs specification for .
package designs

import (
	"cmp"
	"context"
	gerrors "errors"
	"fmt"
	"maps"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/pkg/errors"
	"golang.org/x/mod/semver"

	ffs "github.com/openuc2/optikit/exp/fs"
)

// A FSDesignExpr is an Optikit design stored at the root of a [fs.FS] filesystem.
// Various parameters of the design are string expressions which can be evaluated to get a FSDesign.
type FSDesignExpr struct {
	// DesignExpr is the design at the root of the filesystem.
	DesignExpr
	// FS is a filesystem which contains the design's contents.
	FS ffs.PathedFS
}

// A FSDesign is an Optikit design stored at the root of a [fs.FS] filesystem.
type FSDesign struct {
	// Design is the design at the root of the filesystem.
	Design
	// FS is a filesystem which contains the design's contents.
	FS ffs.PathedFS
}

// A DesignExpr is an Optikit design, a complete specification of all package deployments which
// should be active on a Docker host.
// Various parameters of the Decl are string expressions which can be evaluated to get a Design.
type DesignExpr struct {
	// Decl is the Optikit design definition for the design.
	Decl DesignExprDecl
	// Version is the version or pseudoversion of the design.
	Version string
}

// A Design is an Optikit design, a complete specification of all package deployments which should
// be active on a Docker host.
type Design struct {
	// Decl is the Optikit design definition for the design.
	Decl DesignDecl
	// Version is the version or pseudoversion of the design.
	Version string
}

// FSDesignExpr

// LoadFSDesignExpr loads a FSDesignExpr from the specified directory path in the provided base
// filesystem.
func LoadFSDesignExpr(
	ctx context.Context, fsys ffs.PathedFS, subdirPath string,
) (p *FSDesignExpr, err error) {
	p = &FSDesignExpr{}
	if p.FS, err = fsys.Sub(subdirPath); err != nil {
		return nil, errors.Wrapf(
			err, "couldn't enter directory %s from fs at %s", subdirPath, fsys.Path(),
		)
	}
	if p.Decl, err = LoadDesignExprDecl(ctx, p.FS, DesignExprDeclFile); err != nil {
		return nil, errors.Wrapf(err, "couldn't load design declaration from %s", fsys.Path())
	}
	return p, nil
}

// LoadFSDesignExprContaining loads the FSDesignExpr containing the specified sub-directory path in
// the provided base filesystem.
// The provided path should use the host OS's path separators.
// The sub-directory path does not have to actually exist.
func LoadFSDesignExprContaining(ctx context.Context, path string) (*FSDesignExpr, error) {
	designCandidatePath, err := filepath.Abs(path)
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't convert '%s' into an absolute path", path)
	}
	for {
		if fsDesign, err := LoadFSDesignExpr(ctx, ffs.DirFS(designCandidatePath), "."); err == nil {
			return fsDesign, nil
		}

		designCandidatePath = filepath.Dir(designCandidatePath)
		if designCandidatePath == "/" || designCandidatePath == "." {
			// we can't go up anymore!
			return nil, errors.Errorf(
				"no design declaration file found in any parent directory of %s", path,
			)
		}
	}
}

// LoadFSDesignExprs loads all FSDesignExprs from the provided base filesystem matching the
// specified search pattern. The search pattern should be a [doublestar] pattern, such as `**`,
// matching design directories to search for.
// In the embedded [Design] of each loaded FSDesign, the version is *not* initialized.
func LoadFSDesignExprs(
	ctx context.Context, fsys ffs.PathedFS, searchPattern string,
) ([]*FSDesignExpr, error) {
	searchPattern = path.Join(searchPattern, DesignExprDeclFile)
	designDeclFiles, err := doublestar.Glob(fsys, searchPattern)
	if err != nil {
		return nil, errors.Wrapf(
			err, "couldn't search for design declaration files matching %s/%s",
			fsys.Path(), searchPattern,
		)
	}

	orderedDesigns := make([]*FSDesignExpr, 0, len(designDeclFiles))
	designs := make(map[string]*FSDesignExpr)
	for _, designDeclFilePath := range designDeclFiles {
		if path.Base(designDeclFilePath) != DesignExprDeclFile {
			continue
		}
		design, err := LoadFSDesignExpr(ctx, fsys, path.Dir(designDeclFilePath))
		if err != nil {
			return nil, errors.Wrapf(
				err, "couldn't load design from %s/%s", fsys.Path(), designDeclFilePath,
			)
		}

		orderedDesigns = append(orderedDesigns, design)
		designs[design.Path()] = design
	}

	return orderedDesigns, nil
}

// Exists checks whether the design actually exists on the OS's filesystem.
func (d *FSDesignExpr) Exists() bool {
	return ffs.DirExists(d.FS.Path())
}

// Remove deletes the design from the OS's filesystem, if it exists.
func (d *FSDesignExpr) Remove() error {
	return os.RemoveAll(d.FS.Path())
}

// Path returns either the design's path (if specified) or its path on the filesystem.
func (d *FSDesignExpr) Path() string {
	if d.Decl.Design.Path == "" {
		return d.FS.Path()
	}
	return d.Decl.Design.Path
}

// Cloned returns a new design which is a deep copy of the FSDesign.
func (d *FSDesignExpr) Cloned() *FSDesignExpr {
	return &FSDesignExpr{
		Decl:    d.Decl.Cloned(),
		Version: d.Version,
		FS:      d.FS,
	}
}

// LoadFSDesignExpr loads the subdesign at the specified filesystem path, relative to the current
// design.
func (d *FSDesignExpr) LoadFSDesignExpr(
	ctx context.Context, subdesign string,
) (*FSDesignExpr, error) {
	return LoadFSDesignExpr(ctx, d.FS, subdesign)
}

// FSDesign

// Check looks for errors in the construction of the design.
func (d *FSDesign) Check(ctx context.Context) (errs []error) {
	errs = append(errs, errsWrap(d.Design.Check(), "invalid design")...)
	checks := make(map[CompID]checkSubdesign)
	for compID, comp := range d.Decl.Components {
		if comp.Kind != CompKindDesign {
			continue
		}

		subdesign, err := d.LoadCompFSDesign(ctx, compID)
		if err != nil {
			errs = append(errs, errors.Wrapf(
				err, "couldn't load subdesign %s for component %s", comp.Design, compID,
			))
			continue
		}
		checks[compID] = checkSubdesign{
			name:   comp.Design,
			design: subdesign,
		}
	}
	errs = append(errs, errsWrap(
		d.checkCompInsts(ctx, checks),
		"failed checks for input variable instantiation values",
	)...)

	for compID, check := range checks {
		errs = append(errs, errsWrap(
			check.design.Check(ctx), // recurse
			fmt.Sprintf("failed checks for component %s with subdesign %s", compID, check.name),
		)...)
	}
	return errs
}

type checkSubdesign struct {
	name   string
	design *FSDesign
}

func (d *FSDesign) checkCompInsts(
	ctx context.Context, checks map[CompID]checkSubdesign,
) (errs []error) {
	for compID, comp := range d.Decl.Components {
		if comp.Kind != CompKindDesign {
			continue
		}

		if comp.Design == "" {
			errs = append(errs, errors.Errorf("component %s has an undefined subdesign", compID))
			continue
		}

		check, ok := checks[compID]
		if !ok {
			errs = append(errs, errors.Errorf(
				"couldn't find component %s's subdesign %s among pre-loaded subdesigns", compID, check.name,
			))
			continue
		}

		errs = append(errs, d.checkCompInst(ctx, compID, comp, check)...)
	}
	return errs
}

func (d *FSDesign) checkCompInst(
	ctx context.Context, compID CompID, comp CompSpec, check checkSubdesign,
) (errs []error) {
	inputVars := check.design.Design.Decl.Inputs
	for varName, value := range comp.Instantiation.Inputs {
		inputVar, ok := inputVars[varName]
		if !ok {
			errs = append(errs, errors.Errorf(
				"input variable %s was not defined in subdesign %s", varName, check.name,
			))
			continue
		}

		typedValue, err := ConvertTo(value, inputVar.Kind)
		if err != nil {
			errs = append(errs, errors.Wrapf(
				err, "couldn't convert value %+v to kind %s", value, inputVar.Kind,
			))
		}
		errs = append(errs, errsWrap(
			inputVar.CheckValue(typedValue),
			fmt.Sprintf("failed checks in component %s for input variable %s", compID, varName),
		)...)
	}
	return errs
}

// Exists checks whether the design actually exists on the OS's filesystem.
func (d *FSDesign) Exists() bool {
	return ffs.DirExists(d.FS.Path())
}

// Remove deletes the design from the OS's filesystem, if it exists.
func (d *FSDesign) Remove() error {
	return os.RemoveAll(d.FS.Path())
}

// Path returns either the design's path (if specified) or its path on the filesystem.
func (d *FSDesign) Path() string {
	if d.Decl.Design.Path == "" {
		return d.FS.Path()
	}
	return d.Decl.Design.Path
}

// Cloned returns a new design which is a deep copy of the FSDesign.
func (d *FSDesign) Cloned() *FSDesign {
	return &FSDesign{
		Decl:    d.Decl.Cloned(),
		Version: d.Version,
		FS:      d.FS,
	}
}

// LoadFSDesign loads the subdesign at the specified filesystem path, relative to the current
// design.
func (d *FSDesign) LoadFSDesignExpr(ctx context.Context, subdesign string) (*FSDesignExpr, error) {
	return LoadFSDesignExpr(ctx, d.FS, subdesign)
}

// WithDescendants returns a new design in which all design components have been recursively
// supplemented with their constituent primitive components as descendants.
func (d *FSDesign) WithDescendants(ctx context.Context) (supplemented *FSDesign, err error) {
	design := d.Cloned()
	if err = design.addSubdesignComps(ctx); err != nil {
		return nil, errors.Wrap(err, "couldn't recurse into subdesigns of components")
	}
	// fmt.Printf("all components: %+v\n", slices.Collect(maps.Keys(design.Decl.Components)))
	for assmID, assm := range design.Decl.Assemblies {
		if err = design.addAssmSubdesignComps(ctx, assm.Children, ""); err != nil {
			return nil, errors.Wrapf(err, "couldn't recurse into subdesigns of assembly %s", assmID)
		}
		// fmt.Printf("assembly %s\n", assmID)
		// assm.Children.Print(1)
		// fmt.Println()
	}
	return design, nil
}

// addSubdesignComps recursively adds components from subdesigns and prefixes the IDs of
// those components, as well as any file path references in those components. It does not modify
// assemblies.
func (d *FSDesign) addSubdesignComps(ctx context.Context) error {
	for compID, comp := range d.Decl.Components {
		if comp.Kind != CompKindDesign {
			continue
		}

		subdesign, err := d.LoadCompFSDesign(ctx, compID)
		if err != nil {
			return errors.Wrapf(
				err, "couldn't load subdesign %s for component %s", comp.Design, compID,
			)
		}
		if err = subdesign.addSubdesignComps(ctx); err != nil {
			return errors.Wrapf(
				err, "couldn't flatten subdesign %s for component %s", comp.Design, compID,
			)
		}
		for subcompID, subcomp := range subdesign.Decl.Components {
			switch subcomp.Kind {
			case CompKindDesign:
				subcomp.Design = prefixNonempty(subcomp.Design, comp.Design)
			case CompKindPrimitive:
				subcomp.Primitive.StaticModels = subcomp.Primitive.StaticModels.Prefixed(comp.Design)
			}
			d.Decl.Components[JoinCompIDs(compID, subcompID)] = subcomp
		}
	}
	return nil
}

// addAssmSubdesignComps recursively adds assembly components from subdesigns into the assembly
// subtree. All component IDs (including mate connector IDs) of components from subdesigns are fully
// prefixed.
func (d *FSDesign) addAssmSubdesignComps(
	ctx context.Context, subtree AssmCompsSpec, parentID CompID,
) error {
	for compID, assmComp := range subtree {
		comp, ok := d.Decl.Components[compID]
		if !ok {
			return errors.Errorf("couldn't find component %s", compID)
		}
		if assmComp.Pose.Base.Kind == AssmCompPoseBaseKindMate {
			assmComp.Pose.Base.Mate = JoinCompIDs(parentID, assmComp.Pose.Base.Mate)
			subtree[compID] = assmComp
		}
		if comp.Kind != CompKindDesign {
			if err := d.addAssmSubdesignComps(ctx, assmComp.Children, compID); err != nil {
				return errors.Wrapf(
					err, "couldn't add assembly components from subdesigns of components under %s", compID,
				)
			}
			continue
		}

		subdesign, err := d.LoadCompFSDesign(ctx, compID)
		if err != nil {
			return errors.Wrapf(err, "couldn't load subdesign %s for component %s", comp.Design, compID)
		}
		subassm, ok := subdesign.Decl.Assemblies[assmComp.Assm]
		if !ok {
			return errors.Errorf(
				"couldn't find assembly %s in flattened subdesign %s", assmComp.Assm, comp.Design,
			)
		}

		maps.Copy(assmComp.Children, subassm.Children.Prefixed(compID))
		if err := d.addAssmSubdesignComps(ctx, assmComp.Children, compID); err != nil {
			return errors.Wrapf(
				err, "couldn't add assembly components from subdesigns of components under %s", compID,
			)
		}
	}

	return nil
}

// RelocatedMates returns a new design in which all design components have been recursively
// supplemented with their constituent primitive components, and all mate-posed components have been
// relocated to be descendants of their respective mate connectors.
func (d *FSDesign) RelocatedMates(ctx context.Context) (restructured *FSDesign, err error) {
	design, err := d.WithDescendants(ctx)
	if err != nil {
		return nil, err
	}
	for assmID, assm := range design.Decl.Assemblies {
		relocations := make(map[CompID]map[CompID]compRelocation)
		identifyMatedCompRelocations(assm.Children, relocations)
		if err = relocateMatedComps(assm.Children, relocations); err != nil {
			return nil, errors.Wrapf(
				err, "couldn't relocate mate-posed components for assembly %s", assmID,
			)
		}
		if len(relocations) > 0 {
			return nil, errors.Errorf(
				"couldn't apply some relocations: %+v", slices.Collect(maps.Keys(relocations)),
			)
		}
		// fmt.Printf("assembly %s\n", assmID)
		// assm.Children.Print(1)
		// fmt.Println()
	}
	return design, nil
}

type compRelocation struct {
	assmComp   AssmCompSpec
	prevParent CompID
}

// identifyMatedCompRelocations recursively identifies all necessary relocations of mate-posed
// components in a subtree. It assumes that the subtree has already been fully populated with all
// descendant components.
func identifyMatedCompRelocations(
	subtree AssmCompsSpec, relocations map[CompID]map[CompID]compRelocation,
) {
	for compID, assmComp := range subtree {
		for gcCompID, gcAssmComp := range assmComp.Children {
			if gcAssmComp.Pose.Base.Kind != AssmCompPoseBaseKindMate {
				continue
			}

			mateConnID := gcAssmComp.Pose.Base.Mate
			if relocations[mateConnID] == nil {
				relocations[mateConnID] = make(map[CompID]compRelocation)
			}
			relocations[mateConnID][gcCompID] = compRelocation{
				assmComp:   gcAssmComp,
				prevParent: compID,
			}
		}
	}

	for _, assmComp := range subtree {
		identifyMatedCompRelocations(assmComp.Children, relocations)
	}
}

// relocateMatedComps recursively moves mate-posed components in a subtree to be non-mate
// children of their required mate connectors, so that all descendant components have poses of base
// kind `parent` or `assembly`. It assumes that the subtree has already been fully populated with
// all descendant components.
func relocateMatedComps(
	tree AssmCompsSpec, relocations map[CompID]map[CompID]compRelocation,
) error {
	if len(relocations) == 0 {
		// Skip the tree-traversal computation needed for getting the editable subtrees
		return nil
	}

	// Build an map of subtrees for in-place restructuring, unlike tree.Subtrees() which returns a
	// clone which doesn't allow in-place restructuring
	subtrees, err := tree.EditableSubtrees()
	if err != nil {
		return errors.Wrap(err, "couldn't enumerate all subtrees of assembly tree")
	}
	subtrees[""] = AssmCompSpec{Children: tree}

	g, err := tree.PoseDigraph()
	if err != nil {
		return errors.Wrap(err, "couldn't construct pose digraph from tree")
	}

	// We relocate items in breadth-first tree traversal order because some mate-posed components may
	// themselves have mate connectors for other components, and if we relocate them before those
	// components, the logic is slightly easier to understand.
	nextParents := make([]CompID, 0, len(g))
	nextParents = append(nextParents, "")
	for len(nextParents) > 0 {
		parent := nextParents[0]
		nextParents = nextParents[1:]
		for child := range g[parent] {
			nextParents = append(nextParents, child)
		}

		for newChild, relocation := range relocations[parent] {
			nextParents = append(nextParents, newChild)

			// Modify the child
			child := subtrees[newChild]
			child.Pose.Base = AssmCompPoseBaseSpec{}

			// Move the child
			subtrees[parent].Children[newChild] = child
			delete(subtrees[relocation.prevParent].Children, newChild)
			delete(relocations[parent], newChild)
		}
		delete(relocations, parent)
	}

	return nil
}

// Canonicalized returns a new design in which all descendant components have been restructured so
// that all component poses are of kind `affine` with base kind `parent` or `assembly`.
func (d *FSDesign) Canonicalized(
	ctx context.Context, gridSpacings ContinuousXYZ[float64],
) (restructured *FSDesign, err error) {
	design, err := d.RelocatedMates(ctx)
	if err != nil {
		return nil, err
	}
	for assmID, assm := range design.Decl.Assemblies {
		// fmt.Printf("assembly %s\n", assmID)
		// assm.Children.Print(1)
		// fmt.Println()
		relocations := make(map[CompID]map[CompID]compRelocation)
		identifyRelTranslCompRelocations(assm.Children, relocations)
		if err = relocateRelTranslComps(assm.Children, relocations, gridSpacings); err != nil {
			return nil, errors.Wrapf(
				err, "couldn't relocate relative-position components for assembly %s", assmID,
			)
		}
		if len(relocations) > 0 {
			return nil, errors.Errorf(
				"couldn't apply some relocations: %+v", slices.Collect(maps.Keys(relocations)),
			)
		}
		// fmt.Printf("assembly %s\n", assmID)
		// assm.Children.Print(1)
		// fmt.Println()
	}
	return design, nil
}

// identifyRelTranslCompRelocations recursively identifies all necessary relocations of
// relative-position-posed components in a subtree. It assumes that the subtree has already been
// fully populated with all descendant components.
func identifyRelTranslCompRelocations(
	subtree AssmCompsSpec, relocations map[CompID]map[CompID]compRelocation,
) {
	for compID, assmComp := range subtree {
		for gcCompID, gcAssmComp := range assmComp.Children {
			if gcAssmComp.Pose.Kind != AssmCompPoseKindRelTransl {
				continue
			}

			parentCompID, _, found := strings.Cut(string(gcCompID), "/")
			if !found {
				parentCompID = ""
			}
			if relocations[CompID(parentCompID)] == nil {
				relocations[CompID(parentCompID)] = make(map[CompID]compRelocation)
			}
			relocations[CompID(parentCompID)][gcCompID] = compRelocation{
				assmComp:   gcAssmComp,
				prevParent: compID,
			}
		}
	}

	for _, assmComp := range subtree {
		identifyRelTranslCompRelocations(assmComp.Children, relocations)
	}
}

// relocateRelTranslComps recursively moves relative-position posed components in a subtree to be
// affine-posed children of their parent design components, so that all descendant components have
// poses of kind `affine`. It assumes that the subtree has already been fully populated with
// all descendant components, and that mate-posed components have already been relocated to be under
// their respective mate connectors.
func relocateRelTranslComps(
	tree AssmCompsSpec, relocations map[CompID]map[CompID]compRelocation,
	gridSpacings ContinuousXYZ[float64],
) error {
	if len(relocations) == 0 {
		// Skip the tree-traversal computation needed for getting the editable subtrees
		return nil
	}

	// Build an map of subtrees for in-place restructuring, unlike tree.Subtrees() which returns a
	// clone which doesn't allow in-place restructuring
	subtrees, err := tree.EditableSubtrees()
	if err != nil {
		return errors.Wrap(err, "couldn't enumerate all subtrees of assembly tree")
	}
	subtrees[""] = AssmCompSpec{Children: tree}

	g, err := tree.PoseDigraph()
	if err != nil {
		return errors.Wrap(err, "couldn't construct pose digraph from tree")
	}
	// We relocate items in breadth-first tree traversal order because some components may themselves
	// be translation anchors for other components and thus will need to be relocated before those
	// components.
	nextParents := make([]CompID, 0, len(g))
	nextParents = append(nextParents, "")
	for len(nextParents) > 0 {
		parent := nextParents[0]
		nextParents = nextParents[1:]
		for child := range g[parent] {
			nextParents = append(nextParents, child)
		}

		if len(relocations[parent]) == 0 {
			continue // skip the tree-traversal computation needed for flattening of the parent
		}
		flattenedParent, err := subtrees[parent].Children.Flattened(gridSpacings)
		if err != nil {
			return errors.Wrapf(err, "couldn't compute flattening of subtree %s", parent)
		}

		for newChild, relocation := range relocations[parent] {
			nextParents = append(nextParents, newChild)

			// Modify the child
			child := subtrees[newChild]
			child.Pose.Translation = flattenedParent[relocation.prevParent].Pose.Translation.Added(
				child.Pose.Translation,
			)
			child.Pose.Kind = ""

			// Move the child
			subtrees[parent].Children[newChild] = child
			delete(subtrees[relocation.prevParent].Children, newChild)
			delete(relocations[parent], newChild)
		}
		delete(relocations, parent)
	}

	return nil
}

// Flattened returns a new design in which all descendant components have been restructured
// so that all components are direct children of the assembly, and their poses are of kind `affine`
// with base kind `parent`.
func (d *FSDesign) Flattened(
	ctx context.Context, gridSpacings ContinuousXYZ[float64],
) (flattened *FSDesign, err error) {
	design, err := d.RelocatedMates(ctx)
	if err != nil {
		return nil, err
	}
	for assmID, assm := range design.Decl.Assemblies {
		if assm.Children, err = assm.Children.Flattened(gridSpacings); err != nil {
			return nil, errors.Wrapf(err, "couldn't flatten assembly %s", assmID)
		}
		design.Decl.Assemblies[assmID] = assm
		// fmt.Printf("assembly %s\n", assmID)
		// assm.Children.Print(1)
		// fmt.Println()
	}
	return design, nil
}

// LoadCompFSDesign loads and returns the FSDesign for the specified component.
func (d *FSDesign) LoadCompFSDesign(
	ctx context.Context, compID CompID,
) (subdesign *FSDesign, err error) {
	component := d.Decl.Components[compID]
	if component.Kind != CompKindDesign {
		return nil, errors.Errorf(
			"component %s of kind %s does not have an associated design", compID, component.Kind,
		)
	}

	subdesignExpr, err := d.LoadFSDesignExpr(ctx, component.Design)
	if err != nil {
		return nil, errors.Wrapf(
			err, "couldn't load subdesign %s for component %s", component.Design, compID,
		)
	}
	errs := subdesignExpr.Check()
	if len(errs) > 0 {
		return nil, gerrors.Join(errs...)
	}
	subdesign = &FSDesign{FS: subdesignExpr.FS}
	if subdesign.Design, err = subdesignExpr.Instantiated(component.Instantiation); err != nil {
		return nil, errors.Wrapf(
			err, "couldn't instantiate subdesign %s for component %s as %s",
			component.Design, compID, component.Instantiation,
		)
	}
	if errs = subdesign.Check(ctx); len(errs) > 0 {
		return nil, gerrors.Join(errs...)
	}
	return subdesign, nil
}

// DesignExpr

// Check looks for errors in the construction of the design.
func (d DesignExpr) Check() (errs []error) {
	errs = append(errs, errsWrap(d.Decl.Check(), "invalid design declaration")...)
	return errs
}

func (d DesignExpr) Instantiated(instantiation InstSpec) (Design, error) {
	dd, err := d.Decl.Instantiated(instantiation)
	if err != nil {
		return Design{}, err
	}
	return Design{
		Decl:    dd,
		Version: d.Version,
	}, nil
}

// Design

// Path returns the design path of the Design instance.
func (d Design) Path() string {
	return d.Decl.Design.Path
}

// VersionQuery represents the Design instance as a version query.
func (d Design) VersionQuery() string {
	return fmt.Sprintf("%s@%s", d.Path(), d.Version)
}

// Check looks for errors in the construction of the design.
func (d Design) Check() (errs []error) {
	errs = append(errs, errsWrap(d.Decl.Check(), "invalid design declaration")...)
	return errs
}

func errsWrap(errs []error, message string) []error {
	wrapped := make([]error, 0, len(errs))
	for _, err := range errs {
		wrapped = append(wrapped, errors.Wrap(err, message))
	}
	return wrapped
}

// CompareDesigns returns an integer comparing two [Design] instances according to their paths and
// versions. The result will be 0 if the r and s have the same paths and versions; -1 if r has a
// path which alphabetically comes before the path of s or if the paths are the same but r has a
// lower version than s; or +1 if r has a path which alphabetically comes after the path of s or if
// the paths are the same but r has a higher version than s.
func CompareDesigns(r, s Design) int {
	if result := cmp.Compare(r.Path(), s.Path()); result != 0 {
		return result
	}
	if result := semver.Compare(r.Version, s.Version); result != 0 {
		return result
	}
	return 0
}
