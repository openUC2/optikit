package geom

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/openUC2/optikit/internal/optikit"
)

func renderObjA(ctx context.Context, c *cli.Command) error {
	designDecl, err := optikit.LoadDesignDecl(c.String("cwd"), c.String("variant"))
	if err != nil {
		return err
	}

	result, err := optikit.RenderObjects(ctx, designDecl.Components, c.String("format"))
	if err != nil {
		return err
	}

	return produceOutput(c.Args().First(), result)
}

func produceOutput(outputPath string, output []byte) error {
	if outputPath == "" {
		fmt.Println(string(output))
		return nil
	}
	const perms = 0o644
	return os.WriteFile(outputPath, output, perms)
}

func renderPosGA(ctx context.Context, c *cli.Command) error {
	designDecl, err := optikit.LoadDesignDecl(c.String("cwd"), c.String("variant"))
	if err != nil {
		return err
	}

	result, err := optikit.RenderPositionGraph(ctx, designDecl.Components, c.String("format"))
	if err != nil {
		return err
	}

	return produceOutput(c.Args().First(), result)
}

func renderPosPA(ctx context.Context, c *cli.Command) error {
	designDecl, err := optikit.LoadDesignDecl(c.String("cwd"), c.String("variant"))
	if err != nil {
		return err
	}

	result, err := optikit.RenderPositionPlot(designDecl.Components)
	if err != nil {
		return err
	}

	return produceOutput(c.Args().First(), result)
}
