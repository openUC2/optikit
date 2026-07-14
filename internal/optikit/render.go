package optikit

import (
	"bytes"
	"context"
	"encoding/json"
	gerrors "errors"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"

	"github.com/openUC2/optikit/exp/designs"
	ffs "github.com/openUC2/optikit/exp/fs"
	"github.com/openUC2/optikit/exp/structures"
	"github.com/openUC2/optikit/internal/clients/build123d"
	"github.com/openUC2/optikit/internal/clients/echarts"
	"github.com/openUC2/optikit/internal/clients/gltf"
	"github.com/openUC2/optikit/internal/clients/graphviz"
)

// Objects

func RenderObjects(
	ctx context.Context, fsys ffs.PathedFS, design *designs.FSDesign, format string,
) (result []byte, err error) {
	switch format {
	default:
		return nil, errors.Errorf("unknown format %s", format)
	case "glb":
		return RenderObjectsGLB(fsys, design.Decl.Components, false)
	case "gltf":
		return RenderObjectsGLB(fsys, design.Decl.Components, true)
	case "step":
		return RenderObjectsSTEP(ctx, design)
	}
}

func RenderObjectsGLB(
	fsys ffs.PathedFS, comps designs.CompsSpec, asText bool,
) (result []byte, err error) {
	doc := gltf.NewDocument()
	if result, err = doc.Assemble(fsys, comps, asText, designs.UC2GridSpacings); err != nil {
		return nil, err
	}
	return result, nil
}

func RenderObjectsSTEP(
	ctx context.Context, design *designs.FSDesign,
) (result []byte, err error) {
	primsReport, err := ReportPrimitives(ctx, design, "json")
	if err != nil {
		return nil, err
	}

	bc, err := build123d.New()
	if err != nil {
		return nil, err
	}
	defer func() {
		err = bc.Close()
	}()

	if result, err = bc.Assemble(primsReport); err != nil {
		return nil, err
	}
	return result, nil
}

// Graphs

func RenderPositionGraph(
	ctx context.Context, design *designs.FSDesign, format string, recurse bool,
) (result []byte, err error) {
	gvc, err := graphviz.New(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		cerr := gvc.Close()
		if cerr != nil {
			err = cerr
		}
	}()

	gg := make(structures.StrictEdgeDigraph[string, string])
	if gg, err = populatePositionGraph(gg, design, recurse, ""); err != nil {
		return nil, errors.Wrapf(err, "couldn't populate position graph for design %s", design.Path())
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

func populatePositionGraph(
	gg structures.StrictEdgeDigraph[string, string], design *designs.FSDesign,
	recurse bool, nodePrefix designs.CompID,
) (structures.StrictEdgeDigraph[string, string], error) {
	tg := design.Decl.Components.TranslDigraph()
	fromIDs := slices.Sorted(maps.Keys(tg))
	for _, fromID := range fromIDs {
		from := tg[fromID]
		fromID = designs.JoinCompIDs(nodePrefix, fromID)
		gg.AddNode(string(fromID))
		for _, toID := range slices.Sorted(maps.Keys(from)) {
			edge := from[toID]
			toID = designs.JoinCompIDs(nodePrefix, toID)
			gg.AddEdge(string(fromID), string(toID), edge.String())
		}
	}
	if !recurse {
		return gg, nil
	}

	for _, compID := range fromIDs {
		component := design.Decl.Components[compID]
		if component.Type != "design" {
			continue
		}

		subdesign, err := design.LoadFSDesign(component.Design)
		if err != nil {
			return nil, errors.Wrapf(
				err, "couldn't load subdesign %s for component %s", component.Design, compID,
			)
		}
		errs := subdesign.Check()
		if len(errs) > 0 {
			return nil, gerrors.Join(errs...)
		}
		if subdesign.Decl.NeedsInstantiation() {
			if subdesign.Decl.Components, err = subdesign.Decl.Instantiate(designs.InstSpec{
				Variant: component.Instantiation.Variant,
			}); err != nil {
				return nil, errors.Wrapf(
					err, "couldn't instantiate variant %s of subdesign %s for component %s",
					component.Instantiation.Variant, component.Design, compID,
				)
			}
		}

		if gg, err = populatePositionGraph(
			gg, subdesign, recurse, designs.JoinCompIDs(nodePrefix, compID),
		); err != nil {
			return nil, errors.Wrapf(
				err, "couldn't populate position graph by recursing into subdesign %s for component %s",
				component.Design, compID,
			)
		}
	}
	return gg, nil
}

// Plots

func RenderPositionPlot(comps designs.CompsSpec) (result []byte, err error) {
	c := echarts.NewChart3D()

	flattened := comps.Flattened()
	for _, id := range slices.Sorted(maps.Keys(flattened)) {
		cdecl := flattened[id]
		mat, err := cdecl.Pose.TransfMat(designs.UC2GridSpacings)
		if err != nil {
			return nil, err
		}
		c.AddObject(string(id), mat, designs.UC2GridSpacings.X/2) //nolint:mnd
	}
	c.MakeAxesIsometric()

	return formatPositionPlot(c.Render())
}

func formatPositionPlot(html []byte) (formatted []byte, err error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return nil, err
	}

	chartNode := doc.Find("div.container div.item")
	id, hasID := chartNode.Attr("id") // goecharts randomly generates this, so it's not reproducible
	if !hasID {
		return nil, errors.Errorf("couldn't determine randomly-generated ID of chart node!")
	}
	chartNode.SetAttr("id", "chart")

	scriptNode := doc.Find("script[type=\"text/javascript\"]")
	script := scriptNode.Text()
	script = strings.ReplaceAll(script, id, "chart") // make the HTML source reproducible!
	pattern := regexp.MustCompile(`(?m)^    let option_chart = (?P<options>.+);?$`)
	options := []byte{}
	for _, submatches := range pattern.FindAllSubmatchIndex([]byte(script), -1) {
		options = pattern.ExpandString(options, "$options", script, submatches)
	}
	var indented bytes.Buffer
	if err = json.Indent(&indented, options, "    ", "  "); err != nil {
		return nil, errors.Wrapf(err, "couldn't format chart options: %s", script)
	}
	script = pattern.ReplaceAllString(script, "    let option_chart = "+indented.String()+";")
	scriptNode.SetText(script)

	rendered, err := doc.Html()
	if err != nil {
		return nil, err
	}
	rendered = strings.ReplaceAll(rendered, "&#34;", `"`)
	rendered = strings.ReplaceAll(rendered, "&#39;", `'`)
	return []byte(rendered), nil
}
