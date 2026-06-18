package designs

import (
	"io/fs"

	"github.com/goccy/go-yaml"
	"github.com/pkg/errors"

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

type CompsSpec map[string]CompSpec

// CompSpec declares a component of an Optikit design.
type CompSpec struct {
	// Type is the type of component in the design. It can be either `location` or `design`.
	Type string `yaml:"type"`
	// Design is the path of the design which the component (of type `design`) instantiates. If it's
	// specified as an absolute path, then it will be relative to the root directory of the Optikits
	// design.
	Design string `yaml:"design,omitempty"`
	// Geometry declares the geometry of the component.
	Geometry CompGeomSpec `yaml:"geometry,omitempty"`
	// Tags is a list of human-readable string tags for describing the component to software.
	Tags []string `yaml:"tags,omitempty"`
}

// CompGeomSpec defines declares a Optikit design's component's geometry.
type CompGeomSpec struct {
	// Position declares the position of the component.
	Position CompGeomPositionSpec `yaml:"position,omitempty"`
}

// CompGeomPositionSpec declares the position of the component as an offset relative to an
// "anchor" component, as an x-y-z offset along the design's coordinate axes.
type CompGeomPositionSpec struct {
	// Anchor is the ID of the component whose position will be the relative reference point for
	// setting the position of the "target" component. If empty, it will be the origin of the overall
	// design's coordinate axes.
	Anchor string `yaml:"anchor,omitempty"`
	// Units is the length unit for the offset. It can be `cm` or empty. If empty (i.e. unitless), it
	// will be in UC2 grid units.
	Units string `yaml:"units,omitempty"`
	// Offset is the difference between the component's position and the anchor's position, in the
	// specified units.
	Offset Coordinates `yaml:"offset,omitempty"`
}

// Coordinates is a 3-component vector in an X-Y-Z coordinate system.
type Coordinates struct {
	X float64 `yaml:"x,omitempty"`
	Y float64 `yaml:"y,omitempty"`
	Z float64 `yaml:"z,omitempty"`
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
		anchor := component.Geometry.Position.Anchor
		if _, exists := s[anchor]; anchor != "" && !exists {
			errs = append(errs, errors.Errorf(
				"component %s depends on nonexistent position anchor %s", id, anchor,
			))
		}
	}
	return errs
}
