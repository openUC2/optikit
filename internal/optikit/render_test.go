package optikit

import (
	"context"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/openUC2/optikit/exp/designs"
	"github.com/openUC2/optikit/internal/clients/gltf"
)

var renderDesignDeclTests = []struct {
	design  string
	variant string
}{
	{
		design: "primitives/cube-skeleton.dsn",
	},
	{
		design:  "cube-mounted/lens.dsn",
		variant: "x",
	},
	{
		design:  "cube-mounted/lens.dsn",
		variant: "z",
	},
	{
		design:  "cube-mounted/mirror-diagonal.dsn",
		variant: "xy",
	},
	{
		design:  "cube-mounted/mirror-diagonal.dsn",
		variant: "_z",
	},
	{
		design: "microscopes/simple-3d.dsn",
	},
	{
		design: "microscopes/simple-rel-transl-anchors.dsn",
	},
	{
		design: "microscopes/simple-abs-transl-anchors.dsn",
	},
}

type graphRenderer func(
	ctx context.Context, design *designs.FSDesign, format string, recurse bool,
) (result []byte, err error)

var renderers = []struct {
	filename string
	renderer graphRenderer
}{
	{
		filename: "_components-graph",
		renderer: RenderComponentsGraph,
	},
	{
		filename: "_designs-graph",
		renderer: RenderDesignsGraph,
	},
	{
		filename: "_positions-graph",
		renderer: RenderPositionGraph,
	},
}

func TestRenderGraphs(t *testing.T) {
	t.Parallel() //nolint:tparallel // graphviz is concurrency-unsafe, we can't parallelize subtests
	cwd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	examplesPath := path.Join(path.Dir(path.Dir(cwd)), "examples")

	for _, test := range renderDesignDeclTests {
		dp := path.Join(examplesPath, "designs", test.design)
		for _, renderer := range renderers {
			name := fmt.Sprintf("%s:%s %s", test.design, test.variant, renderer.filename)
			t.Run(name, func(t *testing.T) {
				checkGraph(t, dp, test.design, test.variant, renderer.filename, renderer.renderer)
			})
		}
	}
}

func checkGraph(
	t *testing.T, dp, design, variant, filename string, renderer graphRenderer,
) {
	t.Helper()

	t.Logf("load %s:%s", design, variant)
	d, err := LoadFSDesign(dp, variant, false)
	if err != nil {
		t.Error(err)
		return
	}
	var want, got []byte

	for _, format := range []string{"dot", "svg"} {
		t.Logf("render %s:%s to %s", design, variant, format)
		if got, err = renderer(t.Context(), d, format, true); err != nil {
			t.Error(err)
		}
		if want, err = loadGraph(dp, filename, variant, format); err != nil {
			t.Error(err)
		}
		if !cmp.Equal(got, want) {
			t.Errorf("diff (-want +got):\n%+v", cmp.Diff(want, got))
		}
	}
}

func loadGraph(dp, name, variant, format string) ([]byte, error) {
	if variant != "" {
		name += ":" + variant
	}
	name += "." + format
	return os.ReadFile(path.Join(dp, name))
}

func TestRenderObjectsGLTF(t *testing.T) {
	t.Parallel()
	cwd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	examplesPath := path.Join(path.Dir(path.Dir(cwd)), "examples")

	for design, variants := range reports {
		dp := path.Join(examplesPath, "designs", design)
		for _, variant := range variants {
			name := fmt.Sprintf("%s:%s", design, variant)
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				t.Logf("load %s:%s", design, variant)
				design, err := LoadFSDesign(dp, variant, false)
				if err != nil {
					t.Error(err)
					return
				}

				for _, asText := range []bool{true, false} {
					checkGLTF(t, variant, design, dp, asText)
				}
			})
		}
	}
}

func checkGLTF(t *testing.T, variant string, design *designs.FSDesign, dp string, asText bool) {
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
	if got, err = RenderObjectsGLB(design, asText); err != nil {
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
	t.Parallel()
	cwd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	examplesPath := path.Join(path.Dir(path.Dir(cwd)), "examples")

	for design, variants := range reports {
		dp := path.Join(examplesPath, "designs", design)
		for _, variant := range variants {
			name := fmt.Sprintf("%s:%s", design, variant)
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				t.Logf("load %s:%s", design, variant)
				d, err := LoadFSDesign(dp, variant, false)
				if err != nil {
					t.Error(err)
					return
				}

				var buf []byte

				t.Logf("round-trip glb loading and encoding of %s:%s", design, variant)
				if buf, err = RenderObjectsGLB(d, false); err != nil {
					t.Error(err)
				}
				roundtripDoc(t, buf, false)

				t.Logf("round-trip gltf loading and encoding of %s:%s", design, variant)
				if buf, err = RenderObjectsGLB(d, true); err != nil {
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
