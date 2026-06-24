package designs

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/openUC2/optikit/exp/fs"
)

// Design Decl

var loadDesignDeclTests = map[string][]error{
	"cube-mounted/lens.dsn":                     nil,
	"cube-mounted/mirror-diagonal.dsn":          nil,
	"cube-mounted/slide-holder.dsn":             nil,
	"primitives/cube-skeleton.dsn":              nil,
	"primitives/flashlight.dsn":                 nil,
	"primitives/projector-screen.dsn":           nil,
	"microscopes/simple-rel-transl-anchors.dsn": nil,
	"microscopes/simple-abs-transl-anchors.dsn": nil,
	"microscopes/simple-3d.dsn":                 nil,
	"microscopes/invalid-missing-transl-anchor.dsn": {
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
				return
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

var compsSpecMergeTests = map[string]struct {
	in      CompsSpec
	overlay CompsSpec
	out     func(CompsSpec) CompsSpec
}{
	"none": {
		in: map[CompID]CompSpec{
			"a": exampleCompSpec,
			"b": exampleCompSpec,
		},
		overlay: CompsSpec{},
		out: func(s CompsSpec) CompsSpec {
			return s
		},
	},
	"b": {
		in: map[CompID]CompSpec{
			"a": exampleCompSpec,
			"b": exampleCompSpec,
		},
		overlay: CompsSpec{
			"b": CompSpec{
				Pose: CompPoseSpec{
					Rotation: CompPoseRotSpec{
						Type: "foofoo",
						Grid: CompPoseRotGridSpec{
							Z: "Z!",
						},
					},
				},
			},
		},
		out: func(s CompsSpec) CompsSpec {
			merged := maps.Clone(s)
			b := merged["b"]
			b.Pose.Rotation.Type = "foofoo"
			b.Pose.Rotation.Grid.Z = "Z!"
			merged["b"] = b
			return merged
		},
	},
	"c": {
		in: map[CompID]CompSpec{
			"a": exampleCompSpec,
			"b": exampleCompSpec,
		},
		overlay: map[CompID]CompSpec{
			"c": {
				Type: "test",
			},
		},
		out: func(s CompsSpec) CompsSpec {
			merged := maps.Clone(s)
			merged["c"] = CompSpec{
				Type: "test",
			}
			return merged
		},
	},
	"b,c": {
		in: map[CompID]CompSpec{
			"a": exampleCompSpec,
			"b": exampleCompSpec,
		},
		overlay: CompsSpec{
			"b": CompSpec{
				Pose: CompPoseSpec{
					Rotation: CompPoseRotSpec{
						Type: "foofoo",
						Grid: CompPoseRotGridSpec{
							Z: "Z!",
						},
					},
				},
			},
			"c": CompSpec{
				Type: "test",
			},
		},
		out: func(s CompsSpec) CompsSpec {
			merged := maps.Clone(s)
			b := merged["b"]
			b.Pose.Rotation.Type = "foofoo"
			b.Pose.Rotation.Grid.Z = "Z!"
			merged["b"] = b
			merged["c"] = CompSpec{
				Type: "test",
			}
			return merged
		},
	},
}

func TestCompsSpecMerge(t *testing.T) {
	t.Parallel()

	for name, test := range compsSpecMergeTests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			t.Logf("%s", name)
			if got, want := test.in.Merged(test.overlay), test.out(test.in); !cmp.Equal(got, want) {
				t.Errorf("diff (-want +got):\n%+v", cmp.Diff(want, got))
			}
		})
	}
}

var designFlattenTests = map[string]string{
	"microscopes/simple-abs-transl-anchors.dsn": "microscopes/simple-abs-transl-anchors.dsn",
	"microscopes/simple-rel-transl-anchors.dsn": "microscopes/simple-abs-transl-anchors.dsn",
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
				return
			}

			t.Logf("load %s", out)
			outDecl, err := LoadDesignDecl(examplesFS, path.Join("designs", out, DesignDeclFile))
			if err != nil {
				t.Error(err)
				return
			}

			t.Logf("check %s", in)
			if got, want := inDecl.Components.Flattened(), outDecl.Components; !cmp.Equal(got, want) {
				t.Errorf("diff (-want +got):\n%+v", cmp.Diff(want, got))
			}
		})
	}
}

// CompSpec

var exampleCompSpec = CompSpec{
	Type:   "foo",
	Design: "bar",
	Pose: CompPoseSpec{
		Rotation: CompPoseRotSpec{
			Type: "foobar",
			Grid: CompPoseRotGridSpec{
				Z: "zz",
				X: "xx",
			},
		},
		Translation: CompPoseTranslSpec{
			Anchor: "bazz",
			OffsetGrid: DiscreteXYZ[int]{
				X: 1,
				Y: 2,
				Z: 3,
			},
			OffsetCM: ContinuousXYZ[float64]{
				X: 1.1,
				Y: 2.2,
				Z: 3.3,
			},
		},
	},
}

var compSpecMergeTests = map[string]struct {
	in      CompSpec
	overlay CompSpec
	out     func(CompSpec) CompSpec
}{
	"none": {
		in:      exampleCompSpec,
		overlay: CompSpec{},
		out: func(s CompSpec) CompSpec {
			return s
		},
	},
	"Type": {
		in: exampleCompSpec,
		overlay: CompSpec{
			Type: "foobar",
		},
		out: func(s CompSpec) CompSpec {
			s.Type = "foobar"
			return s
		},
	},
	"Pose.Rotation": {
		in: exampleCompSpec,
		overlay: CompSpec{
			Pose: CompPoseSpec{
				Rotation: CompPoseRotSpec{
					Type: "foofoo",
					Grid: CompPoseRotGridSpec{
						Z: "Z!",
					},
				},
			},
		},
		out: func(s CompSpec) CompSpec {
			s.Pose.Rotation.Type = "foofoo"
			s.Pose.Rotation.Grid.Z = "Z!"
			return s
		},
	},
	"Pose.Translation": {
		in: exampleCompSpec,
		overlay: CompSpec{
			Pose: CompPoseSpec{
				Translation: CompPoseTranslSpec{
					Anchor: "maybe",
					OffsetGrid: DiscreteXYZ[int]{
						X: -1,
					},
					OffsetCM: ContinuousXYZ[float64]{
						Z: 11.1,
					},
				},
			},
		},
		out: func(s CompSpec) CompSpec {
			s.Pose.Translation.Anchor = "maybe"
			s.Pose.Translation.OffsetGrid.X = -1
			s.Pose.Translation.OffsetCM.Z = 11.1
			return s
		},
	},
}

func TestCompSpecMerge(t *testing.T) {
	t.Parallel()

	for name, test := range compSpecMergeTests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			t.Logf("%s", name)
			if got, want := test.in.Merged(test.overlay), test.out(test.in); !cmp.Equal(got, want) {
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
