package designs

import (
	"io/fs"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	ffs "github.com/openUC2/optikit/exp/fs"
)

// DesignDeclFile is the name of the file defining each Optikit design.
const DesignDeclFile = "optikit-design.yml"

// A DesignDecl defines an Optikit design.
type DesignDecl struct {
	// Optikit indicates that the design was written assuming the semantics of a given version
	// of Optikit. The version must be a valid Optikit version, and it sets the minimum version of
	// Optikit required to use the design. The Optikit tool refuses to use designs declaring newer
	// Optikit versions for any operations beyond printing information. The Optikit version of the
	// design must be greater than or equal to the Optikit version of every required Optikit design.
	Optikit string `yaml:"optikit-version"`
	// Design defines the basic metadata for the design.
	Design DesignSpec `yaml:"design,omitempty"`
}

// DesignSpec defines the basic metadata for an Optikit design.
type DesignSpec struct {
	// Path is the design path, which acts as the canonical name for the design. It should just be the
	// path of the VCS repository for the design.
	Path string `yaml:"path"`
	// Description is a short description of the design to be shown to users.
	Description string `yaml:"description"`
	// ReadmeFile is the name of a readme file to be shown to users.
	ReadmeFile string `yaml:"readme-file"`
}

// DesignDecl

// loadDesignDecl loads a DesignDecl from the specified file path in the provided base filesystem.
func loadDesignDecl(fsys ffs.PathedFS, filePath string) (DesignDecl, error) {
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
	return errsWrap(d.Design.Check(), "invalid design spec")
}

// DesignSpec

// Check looks for errors in the construction of the design spec.
func (s DesignSpec) Check() (errs []error) {
	if s.Path == "" {
		errs = append(errs, errors.Errorf("design spec is missing `path` parameter"))
	}
	return errs
}
