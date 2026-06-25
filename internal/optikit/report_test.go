package optikit

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var reportPrimTests = []struct {
	design  string
	variant string
}{
	{
		design: "primitives/cube-skeleton.dsn",
	},
}

func TestReportPrims(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	examplesPath := path.Join(path.Dir(path.Dir(cwd)), "examples")

	for _, test := range reportPrimTests {
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

			for _, format := range []string{"json", "yaml"} {
				t.Logf("report %s:%s to %s", test.design, test.variant, format)
				if got, err = ReportPrimitives(t.Context(), designDecl.Components, format); err != nil {
					t.Error(err)
				}
				if want, err = os.ReadFile(path.Join(dp, "_primitives."+fileExts[format])); err != nil {
					t.Error(err)
				}
				if !cmp.Equal(got, want) {
					t.Errorf("diff (-want +got):\n%+v", cmp.Diff(want, got))
				}
			}
		})
	}
}

var fileExts = map[string]string{
	"json": "json",
	"yaml": "yml",
}
