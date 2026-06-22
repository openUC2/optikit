package echarts

import (
	"slices"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/ungerik/go3d/float64/mat3"
	"github.com/ungerik/go3d/float64/mat4"
	"github.com/ungerik/go3d/float64/vec3"
)

type Chart3D struct {
	c      *charts.Line3D
	Limits vec3.Box
}

func NewChart3D() *Chart3D {
	chart := charts.NewLine3D()
	chart.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "1600px",
			Height: "800px",
		}),
	)
	return &Chart3D{
		c: chart,
	}
}

func (c *Chart3D) AddObject(name string, pose mat4.T, originMarkerSize float64) *Chart3D {
	markers := make([]vec3.Box, len(vec3.Zero))
	for i := range len(markers) {
		markers[i] = vec3.Box{
			Min: mat3.Ident[i].Scaled(-1 * originMarkerSize),
			Max: mat3.Ident[i].Scaled(originMarkerSize),
		}
	}
	o := pose.MulVec3(&vec3.Zero)
	placed := make([]vec3.Box, len(vec3.Zero))
	for i := range len(placed) {
		placed[i] = vec3.Box{
			Min: pose.MulVec3(&markers[i].Min),
			Max: pose.MulVec3(&markers[i].Max),
		}
	}
	for i := range len(placed) {
		c.c.AddSeries(name, []opts.Chart3DData{
			{Value: toChart3DValue(placed[i].Min)},
			{Value: toChart3DValue(o)},
			{Value: toChart3DValue(placed[i].Max)},
		})
		c.Limits.Join(&placed[i])
	}
	data := make([]opts.Chart3DData, 0)
	for i, axis := range []string{"x", "y"} {
		data = append(data, opts.Chart3DData{
			Name:  "+" + axis,
			Value: toChart3DValue(placed[i].Max),
		})
	}
	data = append(data, opts.Chart3DData{
		Name:  "+x",
		Value: toChart3DValue(placed[0].Max),
	}, opts.Chart3DData{
		Name:  "o",
		Value: toChart3DValue(o),
	})
	c.c.AddSeries(name, data)
	return c
}

func toChart3DValue(v vec3.T) []any {
	return []any{v[0], v[1], v[2]}
}

func (c *Chart3D) MakeAxesIsometric() {
	center := c.Limits.Center()
	extent := vec3.Sub(&c.Limits.Max, &c.Limits.Min)
	maxExtent := slices.Max(extent.Slice())
	halfDiagonal := vec3.UnitXYZ.Scaled(maxExtent / 2) //nolint:mnd
	limits := vec3.Box{
		Min: vec3.Sub(&center, &halfDiagonal),
		Max: vec3.Add(&center, &halfDiagonal),
	}
	c.c.SetGlobalOptions(
		charts.WithXAxis3DOpts(opts.XAxis3D{
			Min: limits.Min[0],
			Max: limits.Max[0],
		}),
		charts.WithYAxis3DOpts(opts.YAxis3D{
			Min: limits.Min[1],
			Max: limits.Max[1],
		}),
		charts.WithZAxis3DOpts(opts.ZAxis3D{
			Min: limits.Min[2],
			Max: limits.Max[2],
		}),
	)
}

func (c *Chart3D) Render() []byte {
	return c.c.RenderContent()
}
