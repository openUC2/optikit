package geom

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/openUC2/optikit/internal/optikit"
)

func renderPosGA(ctx context.Context, c *cli.Command) (err error) {
	designDecl, err := optikit.LoadDesignDecl(c.String("cwd"), c.String("variant"))
	if err != nil {
		return err
	}

	result, err := optikit.RenderPositionGraph(ctx, designDecl.Components, c.String("format"))
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

func renderPosPA(ctx context.Context, c *cli.Command) (err error) {
	designDecl, err := optikit.LoadDesignDecl(c.String("cwd"), c.String("variant"))
	if err != nil {
		return err
	}

	result, err := optikit.RenderPositionPlot(designDecl.Components)
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
