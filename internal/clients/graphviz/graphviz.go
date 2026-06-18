package graphviz

import (
	"bytes"
	"context"
	gerrors "errors"
	"fmt"
	"maps"
	"slices"

	"github.com/goccy/go-graphviz"
	"github.com/pkg/errors"

	"github.com/openUC2/optikit/exp/structures"
)

type Client struct {
	gv *graphviz.Graphviz

	graphs structures.Set[*Graph]
}

func New(ctx context.Context) (c *Client, err error) {
	c = new(Client)

	if c.gv, err = graphviz.New(ctx); err != nil {
		return nil, err
	}
	c.gv.SetLayout(graphviz.DOT)

	c.graphs = make(structures.Set[*Graph])
	return c, nil
}

func (c *Client) Close() error {
	errs := make([]error, 0)
	for graph := range c.graphs {
		name := graph.name
		errs = append(errs, errors.Wrapf(graph.Close(), "couldn't close graph %s", name))
	}
	errs = append(errs, c.gv.Close())
	return gerrors.Join(errs...)
}

type Graph struct {
	name string

	c *Client
	g *graphviz.Graph
}

func (c *Client) NewStrictDigraph(
	name string, graph structures.StrictEdgeDigraph[string, string],
) (g *Graph, err error) {
	g = &Graph{
		name: name,
		c:    c,
	}
	if g.g, err = c.gv.Graph(
		graphviz.WithDirectedType(graphviz.StrictDirected), graphviz.WithName(name),
	); err != nil {
		return nil, err
	}

	nodes := make(map[string]*graphviz.Node)
	errs := make([]error, 0)
	for _, node := range slices.Sorted(maps.Keys(graph)) {
		n, err := g.g.CreateNodeByName(node)
		if err != nil {
			errs = append(errs, errors.Wrapf(err, "couldn't create node %s", node))
		}
		nodes[node] = n
	}
	if len(errs) > 0 {
		_ = g.g.Close()
		return nil, gerrors.Join(errs...)
	}
	for _, from := range slices.Sorted(maps.Keys(graph)) {
		edges := graph[from]
		for _, to := range slices.Sorted(maps.Keys(edges)) {
			edge := edges[to]
			e, err := g.g.CreateEdgeByName(
				fmt.Sprintf("%s -> %s", from, to), nodes[from], nodes[to],
			)
			if err != nil {
				errs = append(errs, errors.Wrapf(err, "couldn't create edge %s", edge))
			}
			e.SetLabel(edge)
		}
	}
	if len(errs) > 0 {
		_ = g.g.Close()
		return nil, gerrors.Join(errs...)
	}

	c.graphs.Add(g)
	return g, nil
}

func (g *Graph) Close() error {
	g.c.graphs.Remove(g)
	return g.g.Close()
}

func (g *Graph) DOT(ctx context.Context) ([]byte, error) {
	var buf bytes.Buffer
	if err := g.c.gv.Render(ctx, g.g, graphviz.XDOT, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (g *Graph) SVG(ctx context.Context) ([]byte, error) {
	var buf bytes.Buffer
	if err := g.c.gv.Render(ctx, g.g, graphviz.SVG, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// func (g *Graph) PNG(ctx context.Context) (r []byte, err error) {
// 	// Note: go-graphviz's PNG renderer is broken. For visual examples, see
// 	// https://github.com/goccy/go-graphviz/issues/106 . In the meantime, we render the SVG to PNG.
// 	if r, err = g.SVG(ctx); err != nil {
// 		return nil, err
// 	}
//
// 	svgctx, err := resvg.NewContext(ctx)
// 	defer func() {
// 		err = svgctx.Close()
// 	}()
// 	renderer, err := svgctx.NewRenderer()
// 	defer func() {
// 		err = renderer.Close()
// 	}()
//
// 	// Note: fonts are broken in resvg-go!
// 	if err = renderer.LoadSystemFonts(); err != nil {
// 		return nil, err
// 	}
// 	return renderer.Render(r)
// }
