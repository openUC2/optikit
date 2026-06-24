package optikit

import (
	"errors"
	"os"

	"github.com/openUC2/optikit/exp/designs"
	"github.com/openUC2/optikit/exp/fs"
)

func LoadDesignDecl(path, variant string) (d designs.DesignDecl, err error) {
	pathRoot, err := os.OpenRoot(path)
	if err != nil {
		return d, err
	}
	designFS := fs.AttachPath(pathRoot.FS(), path)
	if d, err = designs.LoadDesignDecl(designFS, designs.DesignDeclFile); err != nil {
		return d, err
	}

	errs := d.Check()
	if len(errs) > 0 {
		return d, errors.Join(errs...)
	}

	if d.NeedsInstantiation() {
		if d.Components, err = d.Instantiate(designs.InstSpec{
			Variant: designs.VariantID(variant),
		}); err != nil {
			return d, err
		}
	}

	return d, err
}
