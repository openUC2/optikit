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
	{
		design:  "primitives/axes.dsn",
		variant: "ZposXpos",
	},
	{
		design:  "primitives/axes.dsn",
		variant: "ZposYpos",
	},
	{
		design:  "primitives/axes.dsn",
		variant: "ZposXneg",
	},
	{
		design:  "primitives/axes.dsn",
		variant: "ZposYneg",
	},
	{
		design:  "primitives/axes.dsn",
		variant: "ZnegXpos",
	},
	{
		design:  "primitives/axes.dsn",
		variant: "ZnegYpos",
	},
	{
		design:  "primitives/axes.dsn",
		variant: "ZnegXneg",
	},
	{
		design:  "primitives/axes.dsn",
		variant: "ZnegYneg",
	},
	{
		design:  "primitives/axes.dsn",
		variant: "YposXpos",
	},
	{
		design:  "primitives/axes.dsn",
		variant: "YposZneg",
	},
	{
		design:  "primitives/axes.dsn",
		variant: "YposXneg",
	},
	{
		design:  "primitives/axes.dsn",
		variant: "YposZpos",
	},
	{
		design:  "primitives/axes.dsn",
		variant: "YnegXpos",
	},
	{
		design:  "primitives/axes.dsn",
		variant: "YnegZneg",
	},
	{
		design:  "primitives/axes.dsn",
		variant: "YnegXneg",
	},
	{
		design:  "primitives/axes.dsn",
		variant: "YnegZpos",
	},
	{
		design:  "primitives/axes.dsn",
		variant: "XposZneg",
	},
	{
		design:  "primitives/axes.dsn",
		variant: "XposYpos",
	},
	{
		design:  "primitives/axes.dsn",
		variant: "XposZpos",
	},
	{
		design:  "primitives/axes.dsn",
		variant: "XposYneg",
	},
	{
		design:  "primitives/axes.dsn",
		variant: "XnegZneg",
	},
	{
		design:  "primitives/axes.dsn",
		variant: "XnegYpos",
	},
	{
		design:  "primitives/axes.dsn",
		variant: "XnegZpos",
	},
	{
		design:  "primitives/axes.dsn",
		variant: "XnegYneg",
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
				reportName := "_primitives"
				if test.variant != "" {
					reportName += ":" + test.variant
				}
				t.Logf("report %s to %s", reportName, format)
				reportName += "." + fileExts[format]
				if got, err = ReportPrimitives(t.Context(), designDecl.Components, format); err != nil {
					t.Error(err)
				}
				if want, err = os.ReadFile(path.Join(dp, reportName)); err != nil {
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
