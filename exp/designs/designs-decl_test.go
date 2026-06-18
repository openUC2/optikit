package designs

import (
	"errors"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/openUC2/optikit/exp/fs"
)

var loadDesignDeclTests = map[string][]error{
	"microscope-1":                                    nil,
	"microscope-1/subdesigns/flashlight":              nil,
	"microscope-1/subdesigns/mounted-diagonal-mirror": nil,
	"microscope-1/subdesigns/mounted-lens":            nil,
	"microscope-1/subdesigns/mounted-slide-holder":    nil,
	"microscope-1/subdesigns/projector-screen":        nil,
	"invalid-1": {
		fmt.Errorf(
			"invalid components spec: component light-source depends on nonexistent position anchor " +
				"sample-holder",
		),
	},
}

func TestSetAdd(t *testing.T) {
	t.Parallel()
	cwd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	examplesRoot, err := os.OpenRoot(path.Join(path.Dir(path.Dir(cwd)), "examples"))
	if err != nil {
		t.Error(err)
	}
	examplesFS := fs.AttachPath(examplesRoot.FS(), cwd)

	for p, errs := range loadDesignDeclTests {
		t.Run(p, func(t *testing.T) {
			t.Parallel()

			t.Logf("load %s", p)
			designDecl, err := LoadDesignDecl(examplesFS, path.Join("designs", p, DesignDeclFile))
			if err != nil {
				t.Error(err)
			}

			t.Logf("check %s", p)
			if got, want := renderErrors(designDecl.Check()), renderErrors(errs); !cmp.Equal(
				got,
				want,
			) {
				t.Errorf("diff (-want +got):\n%+v", cmp.Diff(want, got))
			}
		})
	}
}

func renderErrors(errs []error) string {
	return fmt.Sprintf("%s", errors.Join(errs...))
}
