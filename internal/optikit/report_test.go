package optikit

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/openUC2/optikit/cmd"
	"github.com/openUC2/optikit/exp/designs"
)

var reportTests = map[string][]struct {
	Assembly      designs.AssmID
	Instantiation designs.InstSpec
}{
	"primitives/cube-skeleton.dsn": {{Assembly: ""}},
	"primitives/axes.dsn": {
		{Assembly: "ZposXpos"},
		{Assembly: "ZposYpos"},
		{Assembly: "ZposXneg"},
		{Assembly: "ZposYneg"},
		{Assembly: "ZnegXpos"},
		{Assembly: "ZnegYpos"},
		{Assembly: "ZnegXneg"},
		{Assembly: "ZnegYneg"},
		{Assembly: "YposXpos"},
		{Assembly: "YposZneg"},
		{Assembly: "YposXneg"},
		{Assembly: "YposZpos"},
		{Assembly: "YnegXpos"},
		{Assembly: "YnegZneg"},
		{Assembly: "YnegXneg"},
		{Assembly: "YnegZpos"},
		{Assembly: "XposZneg"},
		{Assembly: "XposYpos"},
		{Assembly: "XposZpos"},
		{Assembly: "XposYneg"},
		{Assembly: "XnegZneg"},
		{Assembly: "XnegYpos"},
		{Assembly: "XnegZpos"},
		{Assembly: "XnegYneg"},
	},
	"cube-mounted/lens.dsn": {
		{
			Assembly:      "x",
			Instantiation: designs.InstSpec{Inputs: map[designs.VarName]any{"offset": -11}},
		},
		{
			Assembly:      "z",
			Instantiation: designs.InstSpec{Inputs: map[designs.VarName]any{"offset": 7}},
		},
	},
	"cube-mounted/mirror-diagonal.dsn": {{Assembly: "3d:_z"}, {Assembly: "3d:xy"}},
	"cube-mounted/slide-holder.dsn": {
		{
			Assembly:      "x",
			Instantiation: designs.InstSpec{Inputs: map[designs.VarName]any{"offset": -12}},
		},
		{
			Assembly:      "z",
			Instantiation: designs.InstSpec{Inputs: map[designs.VarName]any{"offset": 7}},
		},
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

	for design, reports := range reportTests {
		for _, report := range reports {
			name := fmt.Sprintf("%s:%s", design, report)
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				dp := path.Join(examplesPath, "designs", design)

				t.Logf("load %s:%s", design, report)
				design, err := LoadFSDesign(t.Context(), dp, report.Instantiation.Inputs, false)
				if err != nil {
					t.Error(err)
					return
				}

				for format := range fileExts {
					checkComponents(t, design, dp, format)
					checkAssembly(t, report.Assembly, design, dp, format)
				}
			})
		}
	}
}

var fileExts = map[string]string{
	"json": "json",
	"yaml": "yml",
}

func checkComponents(
	t *testing.T, design *designs.FSDesign, dp, format string,
) {
	t.Helper()

	reportName := "_components"
	t.Logf("report %s to %s", reportName, format)
	reportName += "." + fileExts[format]

	var want, got []byte
	var err error
	report, err := ReportComponents(t.Context(), design, designs.UC2GridSpacings, cmd.FallbackVersion)
	if err != nil {
		t.Error(err)
		return
	}
	if got, err = SerializeReport(t.Context(), report, format); err != nil {
		t.Error(err)
		return
	}
	if want, err = os.ReadFile(filepath.Clean(path.Join(dp, reportName))); err != nil {
		t.Error(err)
		return
	}
	if !cmp.Equal(got, want) {
		t.Errorf("diff (-want +got):\n%+v", cmp.Diff(want, got))
	}
}

func checkAssembly(
	t *testing.T, assembly designs.AssmID, design *designs.FSDesign, dp, format string,
) {
	t.Helper()

	reportName := "_assembly"
	if assembly != "" {
		reportName += ":" + string(assembly)
	}
	t.Logf("report %s to %s", reportName, format)
	reportName += "." + fileExts[format]

	var want, got []byte
	var err error
	report, err := ReportAssembly(
		t.Context(), design, assembly, designs.UC2GridSpacings, cmd.FallbackVersion,
	)
	if err != nil {
		t.Error(err)
		return
	}
	if got, err = SerializeReport(t.Context(), report, format); err != nil {
		t.Error(err)
		return
	}
	if want, err = os.ReadFile(filepath.Clean(path.Join(dp, reportName))); err != nil {
		t.Error(err)
		return
	}
	if !cmp.Equal(got, want) {
		t.Errorf("diff (-want +got):\n%+v", cmp.Diff(want, got))
	}
}
