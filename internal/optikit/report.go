package optikit

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"math"
	"slices"

	"github.com/goccy/go-yaml"
	"github.com/pkg/errors"
	"github.com/ungerik/go3d/float64/mat4"
	"github.com/ungerik/go3d/float64/vec3"

	"github.com/openUC2/optikit/exp/designs"
)

// Primitives

func ReportPrimitives(
	ctx context.Context, design *designs.FSDesign, gridSpacings designs.ContinuousXYZ[float64],
	format string,
) (result []byte, err error) {
	d, err := design.Flattened(gridSpacings)
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't flatten design %s", d.Path())
	}
	prims, err := d.Primitives()
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't determine primitives of design %s", d.Path())
	}
	report := make([]PrimReport, 0, len(prims))
	for _, compID := range slices.Sorted(maps.Keys(prims)) {
		comp := prims[compID]
		m, err := comp.Pose.TransfMat(designs.UC2GridSpacings)
		if err != nil {
			return nil, err
		}
		r := PrimReport{
			ID:           compID,
			Type:         cmp.Or(comp.Primitive.Type, "static"),
			StaticModels: comp.Primitive.StaticModels,
			Position:     m.MulVec3(&vec3.Zero),
			Rotation:     NewPrimRotReport(m),
		}
		report = append(report, r)
	}

	switch format {
	default:
		return nil, fmt.Errorf("unknown output format %s", format)
	case "json":
		if result, err = json.MarshalIndent(report, "", "  "); err != nil {
			return nil, err
		}
		return result, nil
	case "yaml":
		if result, err = yaml.Marshal(report); err != nil {
			return nil, err
		}
	}
	return result, nil
}

type PrimReport struct {
	ID           designs.CompID                   `json:"id"            yaml:"id"`
	Type         string                           `json:"type"          yaml:"type"`
	StaticModels designs.CompPrimStaticModelsSpec `json:"static-models" yaml:"static-models"`
	Position     vec3.T                           `json:"position"      yaml:"position,flow"`
	Rotation     PrimRotReport                    `json:"rotation"      yaml:"rotation"`
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
	const roundingPrecision = 10
	return PrimRotReport{
		Type:  "extrinsic",
		Order: "zxy",
		Angles: designs.ContinuousXYZ[float64]{
			X: roundFloat(radToDeg(x), roundingPrecision),
			Y: roundFloat(radToDeg(y), roundingPrecision),
			Z: roundFloat(radToDeg(z), roundingPrecision),
		},
	}
}

func roundFloat(value float64, roundingPrecision uint) float64 {
	power := math.Pow(10, float64(roundingPrecision)) //nolint:mnd
	return math.Round(value*power) / power
}

func radToDeg(rad float64) float64 {
	return rad * (180.0 / math.Pi) //nolint:mnd
}
