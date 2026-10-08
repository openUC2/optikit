package optikit

import (
	"context"
	"fmt"
	"maps"
	"path"
	"slices"

	"github.com/pkg/errors"

	"github.com/openUC2/optikit/exp/designs"
	ffs "github.com/openUC2/optikit/exp/fs"
	"github.com/openUC2/optikit/exp/structures"
	"github.com/openUC2/optikit/internal/clients/gltf"
	"github.com/openUC2/optikit/internal/clients/graphviz"
)

// Objects

func RenderObjects(
	ctx context.Context,
	fsys ffs.PathedFS, design *designs.FSDesign, assembly designs.AssmID, format string,
	optikitVersion string,
) (result []byte, err error) {
	switch format {
	default:
		return nil, errors.Errorf("unknown format %s", format)
	case "glb":
		return RenderObjectsGLB(ctx, design, assembly, false)
	case "gltf":
		return RenderObjectsGLB(ctx, design, assembly, true)
	}
}

func RenderObjectsGLB(
	ctx context.Context, design *designs.FSDesign, assembly designs.AssmID, asText bool,
) (result []byte, err error) {
	doc := gltf.NewDocument()
	if result, err = doc.Assemble(
		ctx, design, assembly, designs.UC2GridSpacings, asText,
	); err != nil {
		return nil, err
	}
	return result, nil
}

// Graphs

const (
	RenderFormatDOT = "dot"
	RenderFormatSVG = "svg"
)

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
	var gn map[string]graphviz.NodeMetadata
	gg.AddNode("")
	if gg, gn, err = populateComponentsGraph(ctx, gg, nil, design, ""); err != nil {
		return nil, errors.Wrapf(err, "couldn't populate components graph for design %s", design.Path())
	}
	gvg, err := gvc.NewStrictDigraph("", gg, gn)
	if err != nil {
		return nil, err
	}

	switch format {
	default:
		return nil, fmt.Errorf("unknown output format %s", format)
	case RenderFormatDOT:
		if result, err = gvg.DOT(ctx); err != nil {
			return nil, err
		}
	case RenderFormatSVG:
		if result, err = gvg.SVG(ctx); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func populateComponentsGraph(
	ctx context.Context,
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
		if component.Kind == designs.CompKindDesign {
			edgeLabel = component.Design
			// TODO: move this into the assembly graph
			// if component.Instantiation.Variant != "" {
			// 	edgeLabel = fmt.Sprintf("%s:%s", edgeLabel, component.Instantiation.Variant)
			// }
		}
		gg.AddEdge(string(rootID), toID, edgeLabel)
	}

	for _, compID := range compIDs {
		component := comps[compID]
		if component.Kind != designs.CompKindDesign {
			continue
		}

		subdesign, err := design.LoadCompFSDesign(ctx, compID)
		if err != nil {
			return nil, nil, errors.Wrapf(
				err, "couldn't load subdesign %s for component %s", component.Design, compID,
			)
		}

		if gg, nodeMetadata, err = populateComponentsGraph(
			ctx, gg, nodeMetadata, subdesign, designs.JoinCompIDs(rootID, compID),
		); err != nil {
			return nil, nil, errors.Wrapf(
				err, "couldn't populate components graph by recursing into subdesign %s for component %s",
				component.Design, compID,
			)
		}
	}
	return gg, nodeMetadata, nil
}

func RenderDesignsGraph(
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

	gg := make(structures.NonStrictEdgeDigraph[string, string])
	var gn map[string]graphviz.NodeMetadata
	gg.AddNode("")
	if gg, gn, err = populateDesignsGraph(ctx, gg, nil, &designs.FSDesign{
		Design: design.Design,
		FS:     ffs.AttachPath(design.FS, ""),
	}, ""); err != nil {
		return nil, errors.Wrapf(err, "couldn't populate designs graph for design %s", design.Path())
	}
	gvg, err := gvc.NewNonStrictDigraph("", gg, gn)
	if err != nil {
		return nil, err
	}

	switch format {
	default:
		return nil, fmt.Errorf("unknown output format %s", format)
	case RenderFormatDOT:
		if result, err = gvg.DOT(ctx); err != nil {
			return nil, err
		}
	case RenderFormatSVG:
		if result, err = gvg.SVG(ctx); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func populateDesignsGraph(
	ctx context.Context,
	gg structures.NonStrictEdgeDigraph[string, string], nodeMetadata map[string]graphviz.NodeMetadata,
	design *designs.FSDesign, rootID string,
) (structures.NonStrictEdgeDigraph[string, string], map[string]graphviz.NodeMetadata, error) {
	if nodeMetadata == nil {
		nodeMetadata = make(map[string]graphviz.NodeMetadata)
	}

	comps := design.Decl.Components
	compIDs := slices.Sorted(maps.Keys(comps))
	for _, id := range compIDs {
		component := comps[id]
		if component.Kind != designs.CompKindDesign {
			continue
		}

		child := path.Join(design.FS.Path(), component.Design)
		// TODO: move this into the assembly graph
		// if component.Instantiation.Variant != "" {
		// 	child += ":" + string(component.Instantiation.Variant)
		// }
		gg.AddNode(child)
		nodeMetadata[child] = graphviz.NodeMetadata{
			Label: component.Design,
		}
		// TODO: move this into the assembly graph
		// if component.Instantiation.Variant != "" {
		// 	nodeMetadata[child] = graphviz.NodeMetadata{
		// 		Label: fmt.Sprintf("%s:%s", component.Design, string(component.Instantiation.Variant)),
		// 	}
		// }
		gg.AddEdge(rootID, child, string(id))
	}

	for _, compID := range compIDs {
		component := comps[compID]
		if component.Kind != designs.CompKindDesign {
			continue
		}

		child := path.Join(design.FS.Path(), component.Design)
		// TODO: move this into the assembly graph
		// if component.Instantiation.Variant != "" {
		// 	child += ":" + string(component.Instantiation.Variant)
		// }
		subdesign, err := design.LoadCompFSDesign(ctx, compID)
		if err != nil {
			return nil, nil, errors.Wrapf(
				err, "couldn't load subdesign %s for component %s", component.Design, compID,
			)
		}

		if gg, nodeMetadata, err = populateDesignsGraph(
			ctx, gg, nodeMetadata, subdesign, child,
		); err != nil {
			return nil, nil, errors.Wrapf(
				err, "couldn't populate designs graph by recursing into subdesign %s for component %s",
				component.Design, compID,
			)
		}
	}
	return gg, nodeMetadata, nil
}

func RenderAssemblyGraph(
	ctx context.Context, design *designs.FSDesign, assembly designs.AssmID, format string,
	recurse bool,
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
	if gg, err = populateAssemblyGraph(ctx, gg, design, assembly, recurse, ""); err != nil {
		return nil, errors.Wrapf(err, "couldn't populate position graph for design %s", design.Path())
	}
	gvg, err := gvc.NewStrictDigraph("", gg, nil)
	if err != nil {
		return nil, err
	}

	switch format {
	default:
		return nil, fmt.Errorf("unknown output format %s", format)
	case RenderFormatDOT:
		if result, err = gvg.DOT(ctx); err != nil {
			return nil, err
		}
	case RenderFormatSVG:
		if result, err = gvg.SVG(ctx); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func populateAssemblyGraph(
	ctx context.Context,
	gg structures.StrictEdgeDigraph[string, string],
	design *designs.FSDesign, assembly designs.AssmID,
	recurse bool, nodePrefix designs.CompID,
) (structures.StrictEdgeDigraph[string, string], error) {
	assm, ok := design.Decl.Assemblies[assembly]
	if !ok {
		return nil, errors.Errorf("couldn't find assembly %s in design %s", assembly, design.Path())
	}
	tg, err := assm.Children.PoseDigraph()
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't build scene graph of assembly %s", assembly)
	}

	fromIDs := slices.Sorted(maps.Keys(tg))
	for _, fromID := range fromIDs {
		from := tg[fromID]
		fromID = designs.JoinCompIDs(nodePrefix, fromID)
		gg.AddNode(string(fromID))
		for _, toID := range slices.Sorted(maps.Keys(from)) {
			edge := from[toID]
			toID = designs.JoinCompIDs(nodePrefix, toID)
			gg.AddEdge(string(fromID), string(toID), edge.Pose.String())
		}
	}
	if !recurse {
		return gg, nil
	}

	for _, compID := range fromIDs {
		assmComp := assm.Children[compID]
		comp := design.Decl.Components[compID]
		if comp.Kind != designs.CompKindDesign {
			continue
		}

		// TODO: annotate the node or parent edge with the included assembly
		subdesign, err := design.LoadCompFSDesign(ctx, compID)
		if err != nil {
			return nil, errors.Wrapf(
				err, "couldn't load subdesign %s for component %s", comp.Design, compID,
			)
		}

		if gg, err = populateAssemblyGraph(
			ctx, gg, subdesign, assmComp.Assm, recurse, designs.JoinCompIDs(nodePrefix, compID),
		); err != nil {
			return nil, errors.Wrapf(
				err, "couldn't populate position graph by recursing into subdesign %s for component %s",
				comp.Design, compID,
			)
		}
	}
	return gg, nil
}
