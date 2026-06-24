package optikit

import (
	"context"
	"fmt"

	"github.com/openUC2/optikit/exp/designs"
	"github.com/openUC2/optikit/exp/structures"
	"github.com/openUC2/optikit/internal/clients/build123d"
	"github.com/openUC2/optikit/internal/clients/echarts"
	"github.com/openUC2/optikit/internal/clients/graphviz"
)

// Objects

func RenderObjects(
	ctx context.Context, comps designs.CompsSpec, format string,
) (result []byte, err error) {
	cqc, err := build123d.New()
	if err != nil {
		return nil, err
	}
	defer func() {
		err = cqc.Close()
	}()

	if result, err = cqc.Run(); err != nil {
		return nil, err
	}
	return result, nil
}

// Graphs

func RenderPositionGraph(
	ctx context.Context, comps designs.CompsSpec, format string,
) (result []byte, err error) {
	gvc, err := graphviz.New(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = gvc.Close()
	}()

	gg := make(structures.StrictEdgeDigraph[string, string])
	for fromID, from := range comps.TranslDigraph() {
		gg.AddNode(string(fromID))
		for toID, edge := range from {
			gg.AddEdge(string(fromID), string(toID), edge.String())
		}
	}
	gvg, err := gvc.NewStrictDigraph("", gg)
	if err != nil {
		return nil, err
	}

	switch format {
	default:
		return nil, fmt.Errorf("unknown output format %s", format)
	case "dot":
		if result, err = gvg.DOT(ctx); err != nil {
			return nil, err
		}
	case "svg":
		if result, err = gvg.SVG(ctx); err != nil {
			return nil, err
		}
	}
	return result, nil
}

// Plots

func RenderPositionPlot(comps designs.CompsSpec) (result []byte, err error) {
	c := echarts.NewChart3D()

	for id, cdecl := range comps.Flattened() {
		mat, err := cdecl.Pose.TransfMat(designs.UC2GridSpacings)
		if err != nil {
			return nil, err
		}
		c.AddObject(string(id), mat, designs.UC2GridSpacings.X/2) //nolint:mnd
	}
	c.MakeAxesIsometric()

	return c.Render(), nil
}
