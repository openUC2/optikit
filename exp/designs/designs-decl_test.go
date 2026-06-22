package designs

import (
	"errors"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/openUC2/optikit/exp/fs"
)

// Design Decl

var loadDesignDeclTests = map[string][]error{
	"subdesigns/flashlight":                   nil,
	"subdesigns/mounted-diagonal-mirror":      nil,
	"subdesigns/mounted-lens":                 nil,
	"subdesigns/mounted-slide-holder":         nil,
	"subdesigns/projector-screen":             nil,
	"microscope-relative-translation-anchors": nil,
	"microscope-absolute-translation-anchors": nil,
	"invalid-missing-translation-anchor": {
		errors.New(
			"invalid components spec: component light-source depends on nonexistent translation anchor " +
				"sample-holder",
		),
	},
}

func TestDesignDecls(t *testing.T) {
	t.Parallel()
	cwd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	examplesRoot, err := os.OpenRoot(path.Join(path.Dir(path.Dir(cwd)), "examples"))
	if err != nil {
		t.Error(err)
	}
	examplesFS := fs.AttachPath(examplesRoot.FS(), cwd)

	for p, errs := range loadDesignDeclTests {
		t.Run(p, func(t *testing.T) {
			t.Parallel()

			t.Logf("load %s", p)
			designDecl, err := LoadDesignDecl(examplesFS, path.Join("designs", p, DesignDeclFile))
			if err != nil {
				t.Error(err)
			}

			t.Logf("check %s", p)
			if got, want := renderErrors(designDecl.Check()), renderErrors(errs); !cmp.Equal(
				got, want,
			) {
				t.Errorf("diff (-want +got):\n%+v", cmp.Diff(want, got))
			}
		})
	}
}

func renderErrors(errs []error) string {
	if len(errs) == 0 {
		return ""
	}
	return fmt.Sprintf("%s", errors.Join(errs...))
}

// CompsSpec

var designFlattenTests = map[string]string{
	"microscope-relative-translation-anchors": "microscope-absolute-translation-anchors",
}

func TestDesignFlatten(t *testing.T) {
	t.Parallel()
	cwd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	examplesRoot, err := os.OpenRoot(path.Join(path.Dir(path.Dir(cwd)), "examples"))
	if err != nil {
		t.Error(err)
	}
	examplesFS := fs.AttachPath(examplesRoot.FS(), cwd)

	for in, out := range designFlattenTests {
		t.Run(in, func(t *testing.T) {
			t.Parallel()

			t.Logf("load %s", in)
			inDecl, err := LoadDesignDecl(examplesFS, path.Join("designs", in, DesignDeclFile))
			if err != nil {
				t.Error(err)
			}

			t.Logf("load %s", out)
			outDecl, err := LoadDesignDecl(examplesFS, path.Join("designs", out, DesignDeclFile))
			if err != nil {
				t.Error(err)
			}

			t.Logf("check %s", in)
			if got, want := inDecl.Components.Flattened(), outDecl.Components; !cmp.Equal(got, want) {
				t.Errorf("diff (-want +got):\n%+v", cmp.Diff(want, got))
			}
		})
	}
}

// CompPoseRotSpec

var compPoseRotUC2Tests = []struct {
	in   CompPoseRotSpec
	errs []error
}{
	{
		in: CompPoseRotSpec{
			Type: RotTypeUC2,
			Grid: CompPoseRotGridSpec{Z: DirZPos, X: DirXPos},
		},
	},
	{
		in: CompPoseRotSpec{
			Type: RotTypeUC2,
			Grid: CompPoseRotGridSpec{Z: DirZPos, X: DirYPos},
		},
	},
	{
		in: CompPoseRotSpec{
			Type: RotTypeUC2,
			Grid: CompPoseRotGridSpec{Z: DirZPos, X: DirXNeg},
		},
	},
	{
		in: CompPoseRotSpec{
			Type: RotTypeUC2,
			Grid: CompPoseRotGridSpec{Z: DirZPos, X: DirYNeg},
		},
	},
	{
		in: CompPoseRotSpec{
			Type: RotTypeUC2,
			Grid: CompPoseRotGridSpec{Z: DirZNeg, X: DirXPos},
		},
	},
	{
		in: CompPoseRotSpec{
			Type: RotTypeUC2,
			Grid: CompPoseRotGridSpec{Z: DirZNeg, X: DirYPos},
		},
	},
	{
		in: CompPoseRotSpec{
			Type: RotTypeUC2,
			Grid: CompPoseRotGridSpec{Z: DirZNeg, X: DirXNeg},
		},
	},
	{
		in: CompPoseRotSpec{
			Type: RotTypeUC2,
			Grid: CompPoseRotGridSpec{Z: DirZNeg, X: DirYNeg},
		},
	},
	{
		in: CompPoseRotSpec{
			Type: RotTypeUC2,
			Grid: CompPoseRotGridSpec{Z: DirYPos, X: DirXPos},
		},
		errs: []error{errors.New("invalid value for component's z-axis: +y")},
	},
	{
		in: CompPoseRotSpec{
			Type: RotTypeUC2,
			Grid: CompPoseRotGridSpec{Z: DirYNeg, X: DirXPos},
		},
		errs: []error{errors.New("invalid value for component's z-axis: -y")},
	},
	{
		in: CompPoseRotSpec{
			Type: RotTypeUC2,
			Grid: CompPoseRotGridSpec{Z: DirYNeg, X: DirZNeg},
		},
		errs: []error{
			errors.New("invalid value for component's z-axis: -y"),
			errors.New("invalid value for component's x-axis: -z"),
		},
	},
	{
		in: CompPoseRotSpec{
			Type: RotTypeUC2,
			Grid: CompPoseRotGridSpec{Z: DirZPos, X: DirZPos},
		},
		errs: []error{
			errors.New("invalid value for component's x-axis: +z"),
			errors.New("component's z and x axes are coaxial: z=+z, x=+z"),
		},
	},
	{
		in: CompPoseRotSpec{
			Type: RotTypeUC2,
			Grid: CompPoseRotGridSpec{Z: DirZPos, X: DirZNeg},
		},
		errs: []error{
			errors.New("invalid value for component's x-axis: -z"),
			errors.New("component's z and x axes are coaxial: z=+z, x=-z"),
		},
	},
}

func TestCompPoseRotUC2Mats(t *testing.T) {
	t.Parallel()
	for _, test := range compPoseRotUC2Tests {
		name := fmt.Sprintf("{z=%s x=%s}", test.in.Grid.Z, test.in.Grid.X)
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			t.Log(name)

			t.Logf("%s (check)", name)
			if got, want := renderErrors(test.in.Check()), renderErrors(test.errs); !cmp.Equal(
				got, want,
			) {
				t.Errorf("diff (-want +got):\n%+v", cmp.Diff(want, got))
			}

			if len(test.errs) > 0 {
				return
			}
			mat := test.in.TransfMat()

			t.Logf("%s (no scaling or reflection)", name)
			if got, want := mat.Determinant3x3(), 1.0; !cmp.Equal(got, want) {
				t.Errorf("diff (-want +got:\n%+v)", cmp.Diff(want, got))
			}

			testRotMatAxes(t, name, test.in.Grid.Z, test.in.Grid.X, mat)
		})
	}
}

func TestCompPoseRotGridMats(t *testing.T) {
	t.Parallel()
	for _, z := range directions {
		for _, x := range directions {
			name := fmt.Sprintf("{z=%s x=%s}", z, x)
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				t.Log(name)
				spec := CompPoseRotSpec{
					Type: RotTypeGrid,
					Grid: CompPoseRotGridSpec{Z: z, X: x},
				}

				if z == x || z == negate[x] {
					if got, want := renderErrors(spec.Check()), renderErrors([]error{
						fmt.Errorf("component's z and x axes are coaxial: z=%s, x=%s", z, x),
					}); !cmp.Equal(got, want) {
						t.Errorf("diff (-want +got):\n%+v", cmp.Diff(want, got))
					}
					return
				}

				mat := spec.TransfMat()

				t.Logf("%s (no scaling or reflection)", name)
				if got, want := mat.Determinant3x3(), 1.0; !cmp.Equal(got, want) {
					t.Errorf("diff (-want +got:\n%+v)", cmp.Diff(want, got))
				}

				testRotMatAxes(t, name, z, x, mat)
			})
		}
	}
}

var directions = []string{
	DirXPos,
	DirXNeg,
	DirYPos,
	DirYNeg,
	DirZPos,
	DirZNeg,
}

var negate = map[string]string{
	DirXPos: DirXNeg,
	DirXNeg: DirXPos,
	DirYPos: DirYNeg,
	DirYNeg: DirYPos,
	DirZPos: DirZNeg,
	DirZNeg: DirZPos,
}
