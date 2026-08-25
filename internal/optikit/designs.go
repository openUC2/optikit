package optikit

import (
	gerrors "errors"
	"maps"
	"os"

	"github.com/pkg/errors"

	"github.com/openUC2/optikit/exp/designs"
	ofs "github.com/openUC2/optikit/exp/fs"
)

func LoadFSDesign(
	path string, variant designs.VariantID, inputs map[designs.VarName]any, isolate bool,
) (d *designs.FSDesign, err error) {
	fsys := os.DirFS(path)
	if isolate {
		pathRoot, err := os.OpenRoot(path)
		if err != nil {
			return d, err
		}
		fsys = pathRoot.FS()
	}
	designFS := ofs.AttachPath(fsys, path)
	de, err := designs.LoadFSDesignExpr(designFS, ".")
	if err != nil {
		return d, err
	}
	errs := de.Check()
	if len(errs) > 0 {
		return d, gerrors.Join(errs...)
	}

	d = &designs.FSDesign{
		FS: de.FS,
	}
	i := make(map[designs.VarName]any)
	maps.Copy(i, inputs)
	if d.Design, err = de.Instantiated(designs.InstSpec{
		Variant: variant,
		Inputs:  i,
	}); err != nil {
		return d, errors.Wrapf(err, "couldn't instantiate with variant %s & inputs %+v", variant, i)
	}
	if errs = d.Check(); len(errs) > 0 {
		return d, gerrors.Join(errs...)
	}

	return d, err
}

func LoadDesignDecl(
	path, variant string, inputs map[string]any,
) (d designs.DesignDecl, err error) {
	pathRoot, err := os.OpenRoot(path)
	if err != nil {
		return d, err
	}
	designFS := ofs.AttachPath(pathRoot.FS(), path)
	de, err := designs.LoadDesignExprDecl(designFS, designs.DesignExprDeclFile)
	if err != nil {
		return d, err
	}

	errs := de.Check()
	if len(errs) > 0 {
		return d, gerrors.Join(errs...)
	}

	i := make(map[designs.VarName]any)
	for name, value := range inputs {
		i[designs.VarName(name)] = value
	}
	if d, err = de.Instantiated(designs.InstSpec{
		Variant: designs.VariantID(variant),
		Inputs:  i,
	}); err != nil {
		return d, err
	}
	if errs = d.Check(); len(errs) > 0 {
		return d, gerrors.Join(errs...)
	}

	return d, err
}
