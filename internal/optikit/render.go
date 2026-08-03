package optikit

import (
	"bytes"
	"context"
	"encoding/json"
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
		return RenderObjectsGLB(design, false)
	case "gltf":
		return RenderObjectsGLB(design, true)
	case "step":
		return RenderObjectsSTEP(ctx, design)
	}
}

func RenderObjectsGLB(
	design *designs.FSDesign, asText bool,
) (result []byte, err error) {
	doc := gltf.NewDocument()
	if result, err = doc.Assemble(design, asText, designs.UC2GridSpacings); err != nil {
		return nil, err
	}
	return result, nil
}

func RenderObjectsSTEP(
	ctx context.Context, design *designs.FSDesign,
) (result []byte, err error) {
	primsReport, err := ReportPrimitives(ctx, design, designs.UC2GridSpacings, "json")
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

func RenderComponentsGraph(
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
	gn := make(map[string]graphviz.NodeMetadata)
	gg.AddNode("")
	if gg, gn, err = populateComponentsGraph(gg, nil, design, ""); err != nil {
		return nil, errors.Wrapf(err, "couldn't populate components graph for design %s", design.Path())
	}
	gvg, err := gvc.NewStrictDigraph("", gg, gn)
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

func populateComponentsGraph(
	gg structures.StrictEdgeDigraph[string, string], nodeMetadata map[string]graphviz.NodeMetadata,
	design *designs.FSDesign, rootID designs.CompID,
) (structures.StrictEdgeDigraph[string, string], map[string]graphviz.NodeMetadata, error) {
	if nodeMetadata == nil {
		nodeMetadata = make(map[string]graphviz.NodeMetadata)
	}

	comps := design.Decl.Components
	compIDs := slices.Sorted(maps.Keys(comps))
	for _, id := range compIDs {
		component := comps[id]
		toID := string(designs.JoinCompIDs(rootID, id))
		gg.AddNode(toID)
		nodeMetadata[toID] = graphviz.NodeMetadata{
			Label: string(id),
		}
		edgeLabel := ""
		if component.Type == "design" {
			edgeLabel = component.Design
			if component.Instantiation.Variant != "" {
				edgeLabel = fmt.Sprintf("%s:%s", edgeLabel, component.Instantiation.Variant)
			}
		}
		gg.AddEdge(string(rootID), string(toID), edgeLabel)
	}

	for _, compID := range compIDs {
		component := comps[compID]
		if component.Type != designs.CompTypeDesign {
			continue
		}

		subdesign, err := design.LoadCompFSDesign(compID)
		if err != nil {
			return nil, nil, errors.Wrapf(
				err, "couldn't load subdesign %s for component %s", component.Design, compID,
			)
		}

		if gg, nodeMetadata, err = populateComponentsGraph(
			gg, nodeMetadata, subdesign, designs.JoinCompIDs(rootID, compID),
		); err != nil {
			return nil, nil, errors.Wrapf(
				err, "couldn't populate components graph by recursing into subdesign %s for component %s",
				component.Design, compID,
			)
		}
	}
	return gg, nodeMetadata, nil
}

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
	gvg, err := gvc.NewStrictDigraph("", gg, nil)
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
		if component.Type != designs.CompTypeDesign {
			continue
		}

		subdesign, err := design.LoadCompFSDesign(compID)
		if err != nil {
			return nil, errors.Wrapf(
				err, "couldn't load subdesign %s for component %s", component.Design, compID,
			)
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

	flattened := comps.TranslFlattened()
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
