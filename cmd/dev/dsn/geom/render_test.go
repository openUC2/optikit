package geom

import (
	"os"
	"path"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var renderDesignDeclTests = map[string][]error{
	"microscope-relative-translation-anchors": nil,
}

func TestRenderPositionGraph(t *testing.T) {
	t.Parallel()
	cwd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	examplesPath := path.Join(path.Dir(path.Dir(path.Dir(path.Dir(cwd)))), "examples")

	for p := range renderDesignDeclTests {
		t.Run(p, func(t *testing.T) {
			t.Parallel()
			dp := path.Join(examplesPath, "designs", p)

			t.Logf("load %s", p)
			designDecl, err := loadDesignDecl(dp)
			if err != nil {
				t.Error(err)
			}
			var want, got []byte

			for _, format := range []string{"dot", "svg"} {
				t.Logf("render %s %s", p, format)
				if got, err = renderPositionGraph(t.Context(), designDecl, format); err != nil {
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
