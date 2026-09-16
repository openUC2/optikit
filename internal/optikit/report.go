package optikit

import (
	"cmp"
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"maps"
	"math"
	"path"
	"slices"

	"github.com/goccy/go-yaml"
	"github.com/pkg/errors"
	"github.com/ungerik/go3d/float64/mat4"
	"github.com/ungerik/go3d/float64/vec3"

	"github.com/openUC2/optikit/exp/designs"
)

// CompReport

type CompReport struct {
	ID   designs.CompID `json:"id"   yaml:"id"`
	Kind string         `json:"kind" yaml:"kind"`

	// Design components:
	Design        string           `json:"design,omitzero"        yaml:"design,omitempty"`
	Instantiation designs.InstSpec `json:"instantiation,omitzero" yaml:"instantiation,omitempty"`

	// Primitive components:
	StaticModels designs.CompPrimStaticModelsSpec `json:"static-models,omitzero" yaml:"static-models,omitempty"`

	Position vec3.T         `json:"position,omitzero" yaml:"position,omitzero,flow"`
	Rotation RotReport      `json:"rotation,omitzero" yaml:"rotation,omitempty"`
	Results  map[string]any `json:"results,omitempty" yaml:"results,omitempty"`
}

type RotReport struct {
	// Kind should be either "intrinsic" or "extrinsic"
	Kind string `json:"kind" yaml:"kind"`
	// Order should be xyz, xzy, yzx, yxz, zxy, zyx, xyx, xzx, yzy, yxy, zxz, or zyz.
	// xyz, xzy, yzx, yxz, zxy, and zyx orders are Tait-Bryan angles, while
	// xyx, xzx, yzy, yxy, zxz, and zyz orders are pure Euler angles.
	// If the kind is flipped and the order is reversed, then the overall rotation remains the same.
	// For example, a rotation matrix defined as extrinsic ZXY (where Y, X, and Z are the rotation
	// matrices for rotations about the world's Z-axis, X-axis, and Y-axis, respectively) corresponds
	// to extrinsic rotations about the y-axis, then the x-axis, then the z-axis, in that order.
	Order string `json:"order" yaml:"order"`
	// Angles is in units of degrees
	Angles designs.ContinuousXYZ[float64] `json:"angles,omitzero" yaml:"angles,omitempty"`
}

func NewRotReport(m mat4.T) RotReport {
	y, x, z := m.ExtractEulerAngles()
	const roundingPrecision = 10
	return RotReport{
		Kind:  "extrinsic",
		Order: "zxy",
		Angles: designs.ContinuousXYZ[float64]{
			X: roundFloat(radToDeg(x), roundingPrecision),
			Y: roundFloat(radToDeg(y), roundingPrecision),
			Z: roundFloat(radToDeg(z), roundingPrecision),
		},
	}
}

func roundFloat(value float64, roundingPrecision uint) float64 {
	power := math.Pow(10, float64(roundingPrecision)) //nolint:mnd // base-10 isn't magic...
	return math.Round(value*power) / power
}

func radToDeg(rad float64) float64 {
	return rad * (180.0 / math.Pi) //nolint:mnd // the entire function is a magic number conversion...
}

// Components

func ReportComponents(
	ctx context.Context, design *designs.FSDesign, gridSpacings designs.ContinuousXYZ[float64],
) (report []CompReport, err error) {
	d, err := design.Flattened(ctx, gridSpacings)
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't flatten design %s", design.Path())
	}
	comps := d.Decl.Components
	report = make([]CompReport, 0, len(comps))
	for _, compID := range slices.Sorted(maps.Keys(comps)) {
		r, err := reportComp(compID, comps[compID], gridSpacings)
		if err != nil {
			return nil, errors.Wrapf(err, "couldn't make report for component %s", compID)
		}
		report = append(report, r)
	}
	return report, nil
}

func reportComp(
	compID designs.CompID, comp designs.CompSpec, gridSpacings designs.ContinuousXYZ[float64],
) (report CompReport, err error) {
	report = CompReport{
		ID:      compID,
		Results: comp.Results,
	}
	switch kind := comp.Kind; kind {
	default:
		return CompReport{}, errors.Errorf("unknown component kind %s", kind)
	case designs.CompKindLocation:
		report.Kind = kind
	case designs.CompKindDesign:
		report.Kind = kind
		report.Design = comp.Design
		report.Instantiation = comp.Instantiation
	case designs.CompKindPrimitive:
		report.Kind = path.Join(kind, cmp.Or(comp.Primitive.Kind, "static"))
		report.StaticModels = comp.Primitive.StaticModels
	}
	if comp.Pose != (designs.CompPoseSpec{}) {
		m, err := comp.Pose.TransfMat(gridSpacings)
		if err != nil {
			return CompReport{}, err
		}
		report.Position = m.MulVec3(&vec3.Zero)
		report.Rotation = NewRotReport(m)
	}
	return report, nil
}

func SerializeReport(
	ctx context.Context, report []CompReport, format string,
) (result []byte, err error) {
	switch format {
	default:
		return nil, fmt.Errorf("unknown output format %s", format)
	case "json":
		if result, err = json.Marshal(
			report,
			json.Deterministic(true), jsontext.Multiline(true),
			jsontext.CanonicalizeRawFloats(true), jsontext.CanonicalizeRawInts(true),
			jsontext.WithIndent("  "),
		); err != nil {
			return nil, err
		}
		return result, nil
	case "yaml":
		if result, err = yaml.MarshalContext(ctx, report); err != nil {
			return nil, err
		}
		return result, nil
	}
}

// Primitives

func ReportPrimitives(
	ctx context.Context, design *designs.FSDesign, gridSpacings designs.ContinuousXYZ[float64],
) (report []CompReport, err error) {
	d, err := design.Flattened(ctx, gridSpacings)
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't flatten design %s", design.Path())
	}
	comps := d.Decl.Components
	report = make([]CompReport, 0, len(comps))
	for _, compID := range slices.Sorted(maps.Keys(comps)) {
		comp := comps[compID]
		if comp.Kind != designs.CompKindPrimitive {
			continue
		}

		r, err := reportComp(compID, comps[compID], gridSpacings)
		if err != nil {
			return nil, errors.Wrapf(err, "couldn't make report for component %s", compID)
		}
		report = append(report, r)
	}
	return report, nil
}
