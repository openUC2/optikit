package optikit

import (
	"errors"
	"os"

	"github.com/openUC2/optikit/exp/designs"
	ofs "github.com/openUC2/optikit/exp/fs"
)

func LoadFSDesign(path, variant string, isolate bool) (d *designs.FSDesign, err error) {
	fsys := os.DirFS(path)
	if isolate {
		pathRoot, err := os.OpenRoot(path)
		if err != nil {
			return d, err
		}
		fsys = pathRoot.FS()
	}
	designFS := ofs.AttachPath(fsys, path)
	if d, err = designs.LoadFSDesign(designFS, "."); err != nil {
		return d, err
	}

	errs := d.Check()
	if len(errs) > 0 {
		return d, errors.Join(errs...)
	}

	if d.Decl.NeedsInstantiation() {
		if d.Decl.Components, err = d.Decl.Instantiate(designs.InstSpec{
			Variant: designs.VariantID(variant),
		}); err != nil {
			return d, err
		}
	}

	return d, err
}

func LoadDesignDecl(path, variant string) (d designs.DesignDecl, err error) {
	pathRoot, err := os.OpenRoot(path)
	if err != nil {
		return d, err
	}
	designFS := ofs.AttachPath(pathRoot.FS(), path)
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
