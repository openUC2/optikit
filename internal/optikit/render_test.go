package optikit

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/openUC2/optikit/exp/designs"
	ffs "github.com/openUC2/optikit/exp/fs"
	"github.com/openUC2/optikit/internal/clients/gltf"
)

var renderDesignDeclTests = []struct {
	design  string
	variant string
}{
	{
		design: "microscopes/simple-rel-transl-anchors.dsn",
	},
	{
		design: "microscopes/simple-abs-transl-anchors.dsn",
	},
	{
		design: "primitives/cube-skeleton.dsn",
	},
}

func TestRenderPositionGraph(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	examplesPath := path.Join(path.Dir(path.Dir(cwd)), "examples")

	for _, test := range renderDesignDeclTests {
		name := fmt.Sprintf("%s:%s", test.design, test.variant)
		t.Run(name, func(t *testing.T) {
			dp := path.Join(examplesPath, "designs", test.design)

			t.Logf("load %s:%s", test.design, test.variant)
			designDecl, err := LoadDesignDecl(dp, test.variant)
			if err != nil {
				t.Error(err)
				return
			}
			var want, got []byte

			for _, format := range []string{"dot", "svg"} {
				t.Logf("render %s:%s to %s", test.design, test.variant, format)
				if got, err = RenderPositionGraph(t.Context(), designDecl.Components, format); err != nil {
					t.Error(err)
				}
				if want, err = os.ReadFile(path.Join(dp, "_positions-graph."+format)); err != nil {
					t.Error(err)
				}
				if !cmp.Equal(got, want) {
					t.Errorf("diff (-want +got):\n%+v", cmp.Diff(want, got))
				}
			}
		})
	}
}

func TestRenderObjectsGLTF(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	examplesPath := path.Join(path.Dir(path.Dir(cwd)), "examples")

	for design, variants := range reports {
		dp := path.Join(examplesPath, "designs", design)
		fsys := ffs.AttachPath(os.DirFS(dp), dp)
		for _, variant := range variants {
			name := fmt.Sprintf("%s:%s", design, variant)
			t.Run(name, func(t *testing.T) {
				t.Logf("load %s:%s", design, variant)
				designDecl, err := LoadDesignDecl(dp, variant)
				if err != nil {
					t.Error(err)
					return
				}

				for _, asText := range []bool{true, false} {
					checkGLTF(t, fsys, variant, designDecl, dp, asText)
				}
			})
		}
	}
}

func checkGLTF(
	t *testing.T, fsys ffs.PathedFS, variant string,
	designDecl designs.DesignDecl, dp string, asText bool,
) {
	t.Helper()

	format := "glb"
	if asText {
		format = "gltf"
	}
	objectName := "_objects"
	if variant != "" {
		objectName += ":" + variant
	}
	t.Logf("render %s to %s", objectName, format)
	objectName += "." + format

	var want, got []byte
	var err error
	if got, err = RenderObjectsGLB(fsys, designDecl.Components, asText); err != nil {
		t.Error(err)
	}
	if want, err = os.ReadFile(path.Join(dp, objectName)); err != nil {
		t.Error(err)
	}
	if !cmp.Equal(got, want) {
		t.Errorf("diff (-want +got):\n%+v", cmp.Diff(want, got))
	}
}

func TestGLTFRoundtrip(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	examplesPath := path.Join(path.Dir(path.Dir(cwd)), "examples")

	for design, variants := range reports {
		dp := path.Join(examplesPath, "designs", design)
		fsys := ffs.AttachPath(os.DirFS(dp), dp)
		for _, variant := range variants {
			name := fmt.Sprintf("%s:%s", design, variant)
			t.Run(name, func(t *testing.T) {
				t.Logf("load %s:%s", design, variant)
				designDecl, err := LoadDesignDecl(dp, variant)
				if err != nil {
					t.Error(err)
					return
				}

				var buf []byte

				t.Logf("round-trip glb loading and encoding of %s:%s", design, variant)
				if buf, err = RenderObjectsGLB(fsys, designDecl.Components, false); err != nil {
					t.Error(err)
				}
				roundtripDoc(t, buf, false)

				t.Logf("round-trip gltf loading and encoding of %s:%s", design, variant)
				if buf, err = RenderObjectsGLB(fsys, designDecl.Components, true); err != nil {
					t.Error(err)
				}
				roundtripDoc(t, buf, true)

				// Note: glb-to-gltf-to-glb and gltf-to-glb-to-gltf roundtripping don't necessarily work
				// due to potential gltf extensions (e.g. from OnShape's glTF/glb export), so we don't
				// require it.
			})
		}
	}
}

func roundtripDoc(t *testing.T, want []byte, asText bool) {
	t.Helper()

	var doc *gltf.Document
	var got []byte
	var err error

	if doc, err = gltf.Load(want); err != nil {
		t.Error(err)
		return
	}
	if got, err = doc.Encode(asText); err != nil {
		t.Error(err)
		return
	}
	if !cmp.Equal(got, want) {
		t.Errorf("diff (-want +got):\n%+v", cmp.Diff(want, got))
	}
}
