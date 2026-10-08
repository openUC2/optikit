package optikit

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/openUC2/optikit/exp/designs"
	"github.com/openUC2/optikit/internal/clients/gltf"
)

var renderDesignDeclTests = map[string][]struct {
	Assembly      designs.AssmID
	Instantiation designs.InstSpec
}{
	"cube-mounted/lens.dsn": {
		{
			Assembly:      "3d:x",
			Instantiation: designs.InstSpec{Inputs: map[designs.VarName]any{"offset": -11}},
		},
	},
	"microscopes/simple-3d.dsn": {{}},
}

type graphRenderer func(
	ctx context.Context, design *designs.FSDesign, assembly designs.AssmID, format string,
	recurse bool,
) (result []byte, err error)

var renderers = []struct {
	filename string
	renderer graphRenderer
}{
	{
		filename: "_components-graph",
		renderer: func(
			ctx context.Context, design *designs.FSDesign, _ designs.AssmID, format string, recurse bool,
		) (result []byte, err error) {
			return RenderComponentsGraph(ctx, design, format, recurse)
		},
	},
	{
		filename: "_designs-graph",
		renderer: func(
			ctx context.Context, design *designs.FSDesign, _ designs.AssmID, format string, recurse bool,
		) (result []byte, err error) {
			return RenderDesignsGraph(ctx, design, format, recurse)
		},
	},
	{
		filename: "_positions-graph",
		renderer: RenderAssemblyGraph,
	},
}

func TestRenderGraphs(t *testing.T) { //nolint:tparallel // graphviz isn't concurrency-safe
	t.Parallel()
	cwd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	examplesPath := path.Join(path.Dir(path.Dir(cwd)), "examples")

	for design, renderings := range renderDesignDeclTests {
		dp := path.Join(examplesPath, "designs", design)
		for _, rendering := range renderings {
			for _, renderer := range renderers {
				name := fmt.Sprintf("%s:%s %s", design, rendering, renderer.filename)
				t.Run(name, func(t *testing.T) {
					checkGraph(
						t, dp, design, rendering.Assembly, rendering.Instantiation.Inputs,
						renderer.filename, renderer.renderer,
					)
				})
			}
		}
	}
}

func checkGraph(
	t *testing.T, dp, design string,
	assembly designs.AssmID, inputs map[designs.VarName]any,
	filename string, renderer graphRenderer,
) {
	t.Helper()

	t.Logf("load %s:%s:%+v", design, assembly, inputs)
	d, err := LoadFSDesign(t.Context(), dp, inputs, false)
	if err != nil {
		t.Error(err)
		return
	}
	var want, got []byte

	for _, format := range []string{"dot", "svg"} {
		t.Logf("render %s:%s to %s", design, assembly, format)
		if got, err = renderer(t.Context(), d, assembly, format, true); err != nil {
			t.Error(err)
		}
		if want, err = loadGraph(dp, filename, assembly, format); err != nil {
			t.Error(err)
		}
		if !cmp.Equal(got, want) {
			t.Errorf("diff (-want +got):\n%+v", cmp.Diff(want, got))
		}
	}
}

func loadGraph(dp, name string, assembly designs.AssmID, format string) ([]byte, error) {
	if assembly != "" {
		name += ":" + string(assembly)
	}
	name += "." + format
	return os.ReadFile(filepath.Clean(path.Join(dp, name)))
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

	for design, reports := range reportTests {
		dp := path.Join(examplesPath, "designs", design)
		for _, report := range reports {
			name := fmt.Sprintf("%s:%s", design, report)
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				t.Logf("load %s:%s", design, report)
				design, err := LoadFSDesign(t.Context(), dp, report.Instantiation.Inputs, false)
				if err != nil {
					t.Error(err)
					return
				}

				for _, format := range []string{formatGLB, formatGLTF} {
					checkGLTF(t, report.Assembly, design, dp, format == formatGLTF)
				}
			})
		}
	}
}

func checkGLTF(
	t *testing.T, assembly designs.AssmID, design *designs.FSDesign, dp string, asText bool,
) {
	t.Helper()

	format := formatGLB
	if asText {
		format = formatGLTF
	}
	objectName := "_objects"
	if assembly != "" {
		objectName += ":" + string(assembly)
	}
	t.Logf("render %s to %s", objectName, format)
	objectName += "." + format

	var want, got []byte
	var err error
	if got, err = RenderObjectsGLB(t.Context(), design, assembly, asText); err != nil {
		t.Error(err)
		return
	}
	if want, err = os.ReadFile(filepath.Clean(path.Join(dp, objectName))); err != nil {
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

	for design, renderings := range renderDesignDeclTests {
		dp := path.Join(examplesPath, "designs", design)
		for _, rendering := range renderings {
			name := fmt.Sprintf("%s:%s", design, rendering)
			t.Logf("load %s:%s", design, rendering)
			d, err := LoadFSDesign(t.Context(), dp, rendering.Instantiation.Inputs, false)
			if err != nil {
				t.Error(err)
				return
			}

			for _, format := range []string{formatGLTF, formatGLB} {
				t.Run(name, func(t *testing.T) {
					t.Parallel()

					var buf []byte
					t.Logf("round-trip %s loading and encoding of %s:%s", format, design, rendering)
					if buf, err = RenderObjectsGLB(
						t.Context(), d, rendering.Assembly, format == formatGLTF,
					); err != nil {
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
