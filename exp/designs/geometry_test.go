package designs

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/ungerik/go3d/float64/mat4"
	"github.com/ungerik/go3d/float64/vec3"
)

// TestGridRotMats checks that GridRotMats is correctly defined by testing its expected properties.
func TestGridRotMats(t *testing.T) {
	t.Parallel()
	for z, mats := range GridRotMats {
		for x, mat := range mats {
			name := fmt.Sprintf("{z=%s x=%s}", z, x)
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				t.Log(name)
				t.Logf("%s (no scaling or reflection)", name)
				if got, want := mat.Determinant3x3(), 1.0; !cmp.Equal(got, want) {
					t.Errorf("diff (-want +got:\n%+v)", cmp.Diff(want, got))
				}
				testRotMatAxes(t, name, z, x, mat)
			})
		}
	}
}

func testRotMatAxes(t *testing.T, name, z, x string, mat mat4.T) {
	t.Helper()

	unitX := vec3.UnitX
	unitY := vec3.UnitY
	unitZ := vec3.UnitZ

	t.Logf("%s (z-axis)", name)
	wantZ := BasisVec3s[z]
	if got, want := mat.MulVec3(&unitZ), wantZ; !cmp.Equal(got, want) {
		t.Errorf("diff (-want +got:\n%+v)", cmp.Diff(want, got))
	}

	t.Logf("%s (x-axis)", name)
	wantX := BasisVec3s[x]
	if got, want := mat.MulVec3(&unitX), wantX; !cmp.Equal(got, want) {
		t.Errorf("diff (-want +got:\n%+v)", cmp.Diff(want, got))
	}

	t.Logf("%s (y-axis)", name)
	wantY := vec3.Cross(&wantZ, &wantX)
	if got, want := mat.MulVec3(&unitY), wantY; !cmp.Equal(got, want) {
		t.Errorf("diff (-want +got:\n%+v)", cmp.Diff(want, got))
	}
	if got, want := vec3.Cross(&wantX, &wantY), wantZ; !cmp.Equal(got, want) {
		t.Errorf("diff (-want +got:\n%+v)", cmp.Diff(want, got))
	}
}
