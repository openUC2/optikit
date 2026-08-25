package optikit

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/openUC2/optikit/exp/designs"
)

var reports = map[string][]designs.InstSpec{ // design -> instantiations
	"primitives/cube-skeleton.dsn": {{}},
	"primitives/axes.dsn": {
		{Variant: "ZposXpos"},
		{Variant: "ZposYpos"},
		{Variant: "ZposXneg"},
		{Variant: "ZposYneg"},
		{Variant: "ZnegXpos"},
		{Variant: "ZnegYpos"},
		{Variant: "ZnegXneg"},
		{Variant: "ZnegYneg"},
		{Variant: "YposXpos"},
		{Variant: "YposZneg"},
		{Variant: "YposXneg"},
		{Variant: "YposZpos"},
		{Variant: "YnegXpos"},
		{Variant: "YnegZneg"},
		{Variant: "YnegXneg"},
		{Variant: "YnegZpos"},
		{Variant: "XposZneg"},
		{Variant: "XposYpos"},
		{Variant: "XposZpos"},
		{Variant: "XposYneg"},
		{Variant: "XnegZneg"},
		{Variant: "XnegYpos"},
		{Variant: "XnegZpos"},
		{Variant: "XnegYneg"},
	},
	"cube-mounted/lens.dsn": {
		{Variant: "x", Inputs: map[designs.VarName]any{"offset": -11}},
		{Variant: "z", Inputs: map[designs.VarName]any{"offset": 7}},
	},
	"cube-mounted/mirror-diagonal.dsn": {{Variant: "_z"}, {Variant: "xy"}},
	"cube-mounted/slide-holder.dsn": {
		{Variant: "x", Inputs: map[designs.VarName]any{"offset": -12}},
		{Variant: "z", Inputs: map[designs.VarName]any{"offset": 7}},
	},
	"microscopes/simple-3d.dsn":                 {{}},
	"microscopes/simple-rel-transl-anchors.dsn": {{}},
	"microscopes/simple-abs-transl-anchors.dsn": {{}},
}

func TestReportPrims(t *testing.T) {
	t.Parallel()
	cwd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	examplesPath := path.Join(path.Dir(path.Dir(cwd)), "examples")

	for design, instantiations := range reports {
		for _, instantiation := range instantiations {
			name := fmt.Sprintf("%s:%s", design, instantiation)
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				dp := path.Join(examplesPath, "designs", design)

				t.Logf("load %s:%s", design, instantiation)
				design, err := LoadFSDesign(dp, instantiation.Variant, instantiation.Inputs, false)
				if err != nil {
					t.Error(err)
					return
				}

				for format := range fileExts {
					checkPrimitives(t, instantiation.Variant, design, dp, format)
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
	t *testing.T, variant designs.VariantID, design *designs.FSDesign, dp, format string,
) {
	t.Helper()

	reportName := "_primitives"
	if variant != "" {
		reportName += ":" + string(variant)
	}
	t.Logf("report %s to %s", reportName, format)
	reportName += "." + fileExts[format]

	var want, got []byte
	var err error
	if got, err = ReportPrimitives(t.Context(), design, designs.UC2GridSpacings, format); err != nil {
		t.Error(err)
	}
	if want, err = os.ReadFile(filepath.Clean(path.Join(dp, reportName))); err != nil {
		t.Error(err)
	}
	if !cmp.Equal(got, want) {
		t.Errorf("diff (-want +got):\n%+v", cmp.Diff(want, got))
	}
}
