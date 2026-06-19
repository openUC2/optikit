package geom

import (
	"context"
	gerrors "errors"
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/openUC2/optikit/exp/designs"
	"github.com/openUC2/optikit/exp/fs"
	"github.com/openUC2/optikit/exp/structures"
	"github.com/openUC2/optikit/internal/clients/graphviz"
)

func renderPosGA(c *cli.Context) (err error) {
	designDecl, err := loadDesignDecl(c.String("cwd"))
	if err != nil {
		return err
	}

	result, err := renderPositionGraph(c.Context, designDecl, c.String("format"))
	if err != nil {
		return err
	}
	outputPath := c.Args().First()
	if outputPath == "" {
		fmt.Println(string(result))
		return nil
	}
	const perms = 0o644
	return os.WriteFile(outputPath, result, perms)
}

func loadDesignDecl(path string) (d designs.DesignDecl, err error) {
	pathRoot, err := os.OpenRoot(path)
	if err != nil {
		return d, err
	}
	designFS := fs.AttachPath(pathRoot.FS(), path)
	if d, err = designs.LoadDesignDecl(designFS, designs.DesignDeclFile); err != nil {
		return d, err
	}
	errs := d.Check()
	if len(errs) > 0 {
		return d, gerrors.Join(errs...)
	}
	return d, err
}

func renderPositionGraph(
	ctx context.Context, designDecl designs.DesignDecl, format string,
) ([]byte, error) {
	gvc, err := graphviz.New(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = gvc.Close()
	}()

	gg := make(structures.StrictEdgeDigraph[string, string])
	for fromID, from := range designDecl.Components.TranslDigraph() {
		gg.AddNode(string(fromID))
		for toID, edge := range from {
			gg.AddEdge(string(fromID), string(toID), edge.String())
		}
	}
	gvg, err := gvc.NewStrictDigraph("", gg)
	if err != nil {
		return nil, err
	}

	var result []byte
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
