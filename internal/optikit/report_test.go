package optikit

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/openUC2/optikit/exp/designs"
)

var reports = map[string][]string{ // design -> variants
	"primitives/cube-skeleton.dsn": {""},
	"primitives/axes.dsn": {
		"ZposXpos",
		"ZposYpos",
		"ZposXneg",
		"ZposYneg",
		"ZnegXpos",
		"ZnegYpos",
		"ZnegXneg",
		"ZnegYneg",
		"YposXpos",
		"YposZneg",
		"YposXneg",
		"YposZpos",
		"YnegXpos",
		"YnegZneg",
		"YnegXneg",
		"YnegZpos",
		"XposZneg",
		"XposYpos",
		"XposZpos",
		"XposYneg",
		"XnegZneg",
		"XnegYpos",
		"XnegZpos",
		"XnegYneg",
	},
	"cube-mounted/lens.dsn":            {"x", "y", "z"},
	"cube-mounted/mirror-diagonal.dsn": {"_z", "xy"},
	"cube-mounted/slide-holder.dsn":    {"x", "y", "z"},
}

func TestReportPrims(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	examplesPath := path.Join(path.Dir(path.Dir(cwd)), "examples")

	for design, variants := range reports {
		for _, variant := range variants {
			name := fmt.Sprintf("%s:%s", design, variant)
			t.Run(name, func(t *testing.T) {
				dp := path.Join(examplesPath, "designs", design)

				t.Logf("load %s:%s", design, variant)
				design, err := LoadFSDesign(dp, variant, false)
				if err != nil {
					t.Error(err)
					return
				}

				for format := range fileExts {
					checkPrimitives(t, variant, design, dp, format)
				}
			})
		}
	}
}

var fileExts = map[string]string{
	"json": "json",
	"yaml": "yml",
}

func checkPrimitives(
	t *testing.T, variant string, design *designs.FSDesign, dp, format string,
) {
	t.Helper()

	reportName := "_primitives"
	if variant != "" {
		reportName += ":" + variant
	}
	t.Logf("report %s to %s", reportName, format)
	reportName += "." + fileExts[format]

	var want, got []byte
	var err error
	if got, err = ReportPrimitives(t.Context(), design, format); err != nil {
		t.Error(err)
	}
	if want, err = os.ReadFile(path.Join(dp, reportName)); err != nil {
		t.Error(err)
	}
	if !cmp.Equal(got, want) {
		t.Errorf("diff (-want +got):\n%+v", cmp.Diff(want, got))
	}
}
