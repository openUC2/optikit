package geom

import (
	"context"

	"github.com/urfave/cli/v3"

	"github.com/openUC2/optikit/internal/optikit"
)

func reportPrimA(ctx context.Context, c *cli.Command) error {
	designDecl, err := optikit.LoadDesignDecl(c.String("cwd"), c.String("variant"))
	if err != nil {
		return err
	}

	result, err := optikit.ReportPrimitives(ctx, designDecl.Components, c.String("format"))
	if err != nil {
		return err
	}

	return produceOutput(c.Args().First(), result)
}
