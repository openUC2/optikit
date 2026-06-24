package optikit

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/google/go-cmp/cmp"
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
