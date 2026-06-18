// Package designs implements the Optikit designs specification for deployment and composition of
// Optikit packages.
package designs

import (
	"cmp"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/pkg/errors"
	"golang.org/x/mod/semver"

	ffs "github.com/openUC2/optikit/exp/fs"
)

// A FSDesign is an Optikit design stored at the root of a [fs.FS] filesystem.
type FSDesign struct {
	// Design is the design at the root of the filesystem.
	Design
	// FS is a filesystem which contains the design's contents.
	FS ffs.PathedFS
}

// A Design is an Optikit design, a complete specification of all package deployments which should
// be active on a Docker host.
type Design struct {
	// Decl is the Optikit design definition for the design.
	Decl DesignDecl
	// Version is the version or pseudoversion of the design.
	Version string
}

// FSDesign

// LoadFSDesign loads a FSDesign from the specified directory path in the provided base filesystem.
func LoadFSDesign(fsys ffs.PathedFS, subdirPath string) (p *FSDesign, err error) {
	p = &FSDesign{}
	if p.FS, err = fsys.Sub(subdirPath); err != nil {
		return nil, errors.Wrapf(
			err, "couldn't enter directory %s from fs at %s", subdirPath, fsys.Path(),
		)
	}
	if p.Design.Decl, err = LoadDesignDecl(p.FS, DesignDeclFile); err != nil {
		return nil, errors.Errorf("couldn't load design declaration")
	}
	return p, nil
}

// LoadFSDesignContaining loads the FSDesign containing the specified sub-directory path in the
// provided base filesystem.
// The provided path should use the host OS's path separators.
// The sub-directory path does not have to actually exist.
func LoadFSDesignContaining(path string) (*FSDesign, error) {
	designCandidatePath, err := filepath.Abs(path)
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't convert '%s' into an absolute path", path)
	}
	for {
		if fsDesign, err := LoadFSDesign(ffs.DirFS(designCandidatePath), "."); err == nil {
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

// LoadFSDesigns loads all FSDesigns from the provided base filesystem matching the specified search
// pattern. The search pattern should be a [doublestar] pattern, such as `**`, matching design
// directories to search for.
// In the embedded [Design] of each loaded FSDesign, the version is *not* initialized.
func LoadFSDesigns(fsys ffs.PathedFS, searchPattern string) ([]*FSDesign, error) {
	searchPattern = path.Join(searchPattern, DesignDeclFile)
	designDeclFiles, err := doublestar.Glob(fsys, searchPattern)
	if err != nil {
		return nil, errors.Wrapf(
			err, "couldn't search for design declaration files matching %s/%s",
			fsys.Path(), searchPattern,
		)
	}

	orderedDesigns := make([]*FSDesign, 0, len(designDeclFiles))
	designs := make(map[string]*FSDesign)
	for _, designDeclFilePath := range designDeclFiles {
		if path.Base(designDeclFilePath) != DesignDeclFile {
			continue
		}
		design, err := LoadFSDesign(fsys, path.Dir(designDeclFilePath))
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
func (p *FSDesign) Exists() bool {
	return ffs.DirExists(p.FS.Path())
}

// Remove deletes the cache from the OS's filesystem, if it exists.
func (p *FSDesign) Remove() error {
	return os.RemoveAll(p.FS.Path())
}

// Path returns either the design's path (if specified) or its path on the filesystem.
func (p *FSDesign) Path() string {
	if p.Decl.Design.Path == "" {
		return p.FS.Path()
	}
	return p.Decl.Design.Path
}

// Design

// Path returns the design path of the Design instance.
func (p Design) Path() string {
	return p.Decl.Design.Path
}

// VersionQuery represents the Design instance as a version query.
func (p Design) VersionQuery() string {
	return fmt.Sprintf("%s@%s", p.Path(), p.Version)
}

// Check looks for errors in the construction of the design.
func (p Design) Check() (errs []error) {
	errs = append(errs, errsWrap(p.Decl.Check(), "invalid design declaration")...)
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
