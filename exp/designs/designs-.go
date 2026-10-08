// Package designs implements the Optikit designs specification for .
package designs

import (
	"cmp"
	"context"
	gerrors "errors"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/pkg/errors"
	"github.com/ungerik/go3d/float64/mat4"
	"golang.org/x/mod/semver"

	ffs "github.com/openUC2/optikit/exp/fs"
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
		DesignExpr: DesignExpr{
			Decl:    d.Decl.Cloned(),
			Version: d.Version,
		},
		FS: d.FS,
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
		Design: Design{
			Decl:    d.Decl.Cloned(),
			Version: d.Version,
		},
		FS: d.FS,
	}
}

// LoadFSDesign loads the subdesign at the specified filesystem path, relative to the current
// design.
func (d *FSDesign) LoadFSDesignExpr(ctx context.Context, subdesign string) (*FSDesignExpr, error) {
	return LoadFSDesignExpr(ctx, d.FS, subdesign)
}

// Flattened returns a new design in which all design components have been recursively supplemented
// with their constituent primitive components, and each component's pose base in an assembly is
// just that assembly.
func (d *FSDesign) Flattened(
	ctx context.Context, gridSpacings ContinuousXYZ[float64],
) (
	flattened *FSDesign, err error,
) {
	if flattened, err = d.flattenComponents(ctx, d.Cloned()); err != nil {
		return nil, errors.Wrap(err, "couldn't flatten components")
	}
	for assmID, assm := range flattened.Decl.Assemblies {
		if flattened, err = flattened.flattenAssembly(
			ctx, flattened, assmID, assm, gridSpacings,
		); err != nil {
			return nil, errors.Wrapf(err, "couldn't recursively flatten assembly %s", assmID)
		}
	}
	return flattened, nil
}

func (d *FSDesign) flattenComponents(
	ctx context.Context, flattened *FSDesign,
) (modified *FSDesign, err error) {
	for compID, comp := range d.Decl.Components {
		if comp.Kind != CompKindDesign {
			continue
		}

		subdesign, err := d.LoadCompFSDesign(ctx, compID)
		if err != nil {
			return nil, errors.Wrapf(
				err, "couldn't load subdesign %s for component %s", comp.Design, compID,
			)
		}
		subflattened, err := subdesign.flattenComponents(ctx, subdesign.Cloned())
		if err != nil {
			return nil, errors.Wrapf(
				err, "couldn't flatten subdesign %s for component %s", comp.Design, compID,
			)
		}
		for subcompID, subcomp := range subflattened.Decl.Components {
			switch subcomp.Kind {
			case CompKindDesign:
				subcomp.Design = prefixNonempty(subcomp.Design, comp.Design)
			case CompKindPrimitive:
				subcomp.Primitive.StaticModels = subcomp.Primitive.StaticModels.Prefixed(comp.Design)
			}
			flattenedID := JoinCompIDs(compID, subcompID)
			flattened.Decl.Components[flattenedID] = subcomp
		}
	}
	return flattened, nil
}

func (d *FSDesign) flattenAssembly(
	ctx context.Context, flattened *FSDesign, assmID AssmID, assm AssmSpec,
	gridSpacings ContinuousXYZ[float64],
) (modified *FSDesign, err error) {
	if assm.Children, err = assm.Children.Flattened(gridSpacings); err != nil {
		return nil, errors.Wrapf(
			err, "couldn't flatten assembly %s prior to flattening subdesigns", assmID,
		)
	}
	flattened.Decl.Assemblies[assmID] = assm

	for compID, assmComp := range assm.Children.Cloned() {
		comp, ok := flattened.Decl.Components[compID]
		if !ok {
			return nil, errors.Errorf(
				"couldn't find component %s required by assembly %s", compID, assmID,
			)
		}
		if comp.Kind != CompKindDesign {
			continue
		}

		if flattened, err = d.flattenSubtree(
			ctx, flattened, assmID, compID, assmComp, comp, gridSpacings,
		); err != nil {
			return flattened, errors.Wrapf(err, "couldn't flatten assembly at %s as subdesign", compID)
		}
	}
	return flattened, nil
}

func (d *FSDesign) flattenSubtree(
	ctx context.Context,
	flattened *FSDesign, assmID AssmID, compID CompID, assmComp AssmCompSpec, comp CompSpec,
	gridSpacings ContinuousXYZ[float64],
) (modified *FSDesign, err error) {
	mat, err := assmComp.Pose.TransfMat(gridSpacings)
	if err != nil {
		return nil, errors.Wrapf(
			err, "couldn't compute transformation matrix for pose of component %s", compID,
		)
	}
	subdesign, err := d.LoadCompFSDesign(ctx, compID)
	if err != nil {
		return nil, errors.Wrapf(
			err, "couldn't load subdesign %s for component %s", comp.Design, compID,
		)
	}
	subflattened, err := subdesign.Flattened(ctx, gridSpacings)
	if err != nil {
		return nil, errors.Wrapf(
			err, "couldn't flatten subdesign %s for component %s", comp.Design, compID,
		)
	}
	subassm, ok := subflattened.Decl.Assemblies[assmComp.Assm]
	if !ok {
		return flattened, errors.Errorf(
			"couldn't find assembly %s in flattened subdesign %s", assmComp.Assm, comp.Design,
		)
	}
	for subcompID, subassmComp := range subassm.Children {
		flattenedID := JoinCompIDs(compID, subcompID)

		submat, err := subassmComp.Pose.TransfMat(gridSpacings)
		if err != nil {
			return nil, errors.Wrapf(
				err, "couldn't compute transformation matrix for pose of subcomponent %s", subcompID,
			)
		}
		flattenedSubmat := mat4.Ident
		flattenedSubmat.AssignMul(&mat, &submat)
		subassmComp.Pose = NewPose(flattenedSubmat, gridSpacings)

		flattened.Decl.Assemblies[assmID].Children[flattenedID] = subassmComp
	}
	return flattened, nil
}

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
	if errs = subdesign.Check(); len(errs) > 0 {
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
