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

var renderDesignDeclTests = map[string][]designs.InstSpec{ // design -> instantiations
	"primitives/cube-skeleton.dsn": {{}},
	"cube-mounted/lens.dsn":        {{Variant: "x"}, {Variant: "z"}},
	"cube-mounted/slide-holder.dsn": {
		{Variant: "x", Inputs: map[designs.VarName]any{"offset": -12}},
		{Variant: "z", Inputs: map[designs.VarName]any{"offset": 7}},
	},
	"microscopes/simple-3d.dsn":                 {{}},
	"microscopes/simple-rel-transl-anchors.dsn": {{}},
	"microscopes/simple-abs-transl-anchors.dsn": {{}},
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

func TestRenderGraphs(t *testing.T) { //nolint:tparallel // graphviz isn't concurrency-safe
	t.Parallel()
	cwd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	examplesPath := path.Join(path.Dir(path.Dir(cwd)), "examples")

	for design, instantiations := range renderDesignDeclTests {
		dp := path.Join(examplesPath, "designs", design)
		for _, instantiation := range instantiations {
			for _, renderer := range renderers {
				name := fmt.Sprintf("%s:%s %s", design, instantiation, renderer.filename)
				t.Run(name, func(t *testing.T) {
					checkGraph(
						t, dp, design, instantiation.Variant, instantiation.Inputs,
						renderer.filename, renderer.renderer,
					)
				})
			}
		}
	}
}

func checkGraph(
	t *testing.T, dp, design string,
	variant designs.VariantID, inputs map[designs.VarName]any,
	filename string, renderer graphRenderer,
) {
	t.Helper()

	t.Logf("load %s:%s:%+v", design, variant, inputs)
	d, err := LoadFSDesign(dp, variant, inputs, false)
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

func loadGraph(dp, name string, variant designs.VariantID, format string) ([]byte, error) {
	if variant != "" {
		name += ":" + string(variant)
	}
	name += "." + format
	return os.ReadFile(path.Join(dp, name))
}

const (
	formatGLTF = "gltf"
	formatGLB  = "glb"
)

func TestRenderObjectsGLTF(t *testing.T) {
	t.Parallel()
	cwd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	examplesPath := path.Join(path.Dir(path.Dir(cwd)), "examples")

	for design, instantiations := range reports {
		dp := path.Join(examplesPath, "designs", design)
		for _, instantiation := range instantiations {
			name := fmt.Sprintf("%s:%s", design, instantiation)
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				t.Logf("load %s:%s", design, instantiation)
				design, err := LoadFSDesign(dp, instantiation.Variant, instantiation.Inputs, false)
				if err != nil {
					t.Error(err)
					return
				}

				for _, format := range []string{formatGLB, formatGLTF} {
					checkGLTF(t, instantiation.Variant, design, dp, format == formatGLTF)
				}
			})
		}
	}
}

func checkGLTF(
	t *testing.T, variant designs.VariantID, design *designs.FSDesign, dp string, asText bool,
) {
	t.Helper()

	format := formatGLB
	if asText {
		format = formatGLTF
	}
	objectName := "_objects"
	if variant != "" {
		objectName += ":" + string(variant)
	}
	t.Logf("render %s to %s", objectName, format)
	objectName += "." + format

	var want, got []byte
	var err error
	if got, err = RenderObjectsGLB(design, asText); err != nil {
		t.Error(err)
		return
	}
	if want, err = os.ReadFile(path.Join(dp, objectName)); err != nil {
		t.Error(err)
		return
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

	for design, instantiations := range renderDesignDeclTests {
		dp := path.Join(examplesPath, "designs", design)
		for _, instantiation := range instantiations {
			name := fmt.Sprintf("%s:%s", design, instantiation)
			t.Logf("load %s:%s", design, instantiation)
			d, err := LoadFSDesign(dp, instantiation.Variant, instantiation.Inputs, false)
			if err != nil {
				t.Error(err)
				return
			}

			for _, format := range []string{formatGLTF, formatGLB} {
				t.Run(name, func(t *testing.T) {
					t.Parallel()

					var buf []byte
					t.Logf("round-trip %s loading and encoding of %s:%s", format, design, instantiation)
					if buf, err = RenderObjectsGLB(d, format == formatGLTF); err != nil {
						t.Error(err)
					}
					roundtripDoc(t, buf, format == formatGLTF)

					// Note: glb-to-gltf-to-glb and gltf-to-glb-to-gltf roundtripping don't necessarily work
					// due to potential gltf extensions (e.g. from OnShape's glTF/glb export), so we don't
					// require it.
				})
			}
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
