package optikit

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"math"
	"slices"

	"github.com/goccy/go-yaml"
	"github.com/ungerik/go3d/float64/mat4"
	"github.com/ungerik/go3d/float64/vec3"

	"github.com/openUC2/optikit/exp/designs"
)

// Primitives

func ReportPrimitives(
	ctx context.Context, comps designs.CompsSpec,
	format string,
) (result []byte, err error) {
	prims := comps.Primitives()
	report := make([]PrimReport, 0, len(prims))
	for _, compID := range slices.Sorted(maps.Keys(prims)) {
		comp := prims[compID]
		m, err := comp.Pose.TransfMat(designs.UC2GridSpacings)
		if err != nil {
			return nil, err
		}
		r := PrimReport{
			Type:     comp.Primitive.Type,
			Model:    comp.Primitive.Model,
			Position: m.MulVec3(&vec3.Zero),
			Rotation: NewPrimRotReport(m),
		}
		report = append(report, r)
	}

	switch format {
	default:
		return nil, fmt.Errorf("unknown output format %s", format)
	case "json":
		if result, err = json.Marshal(report); err != nil {
			return nil, err
		}
	case "yaml":
		if result, err = yaml.Marshal(report); err != nil {
			return nil, err
		}
	}
	return result, nil
}

type PrimReport struct {
	Type     string        `json:"type"     yaml:"type"`
	Model    string        `json:"model"    yaml:"model"`
	Position vec3.T        `json:"position" yaml:"position,flow"`
	Rotation PrimRotReport `json:"rotation" yaml:"rotation"`
}

type PrimRotReport struct {
	// Type should be either "intrinsic" or "extrinsic"
	Type string `json:"type" yaml:"type"`
	// Order should be xyz, xzy, yzx, yxz, zxy, zyx, xyx, xzx, yzy, yxy, zxz, or zyz.
	// xyz, xzy, yzx, yxz, zxy, and zyx orders are Tait-Bryan angles, while
	// xyx, xzx, yzy, yxy, zxz, and zyz orders are pure Euler angles.
	// If the type is flipped and the order is reversed, then the overall rotation remains the same.
	// For example, a rotation matrix defined as extrinsic ZXY (where Y, X, and Z are the rotation
	// matrices for rotations about the world's Z-axis, X-axis, and Y-axis, respectively) corresponds
	// to extrinsic rotations about the y-axis, then the x-axis, then the z-axis, in that order.
	Order string `json:"order" yaml:"order"`
	// Angles is in units of degrees
	Angles designs.ContinuousXYZ[float64] `json:"angles" yaml:"angles,flow"`
}

func NewPrimRotReport(m mat4.T) PrimRotReport {
	y, x, z := m.ExtractEulerAngles()
	return PrimRotReport{
		Type:  "extrinsic",
		Order: "zxy",
		Angles: designs.ContinuousXYZ[float64]{
			X: radToDeg(x),
			Y: radToDeg(y),
			Z: radToDeg(z),
		},
	}
}

func radToDeg(rad float64) float64 {
	return rad * (180.0 / math.Pi)
}
