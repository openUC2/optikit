package geom

import (
	"context"

	"github.com/urfave/cli/v3"

	"github.com/openUC2/optikit/internal/optikit"
)

func reportPrimA(ctx context.Context, c *cli.Command) error {
	design, err := optikit.LoadFSDesign(c.String("cwd"), c.String("variant"), false)
	if err != nil {
		return err
	}

	result, err := optikit.ReportPrimitives(ctx, design, c.String("format"))
	if err != nil {
		return err
	}

	return produceOutput(c.Args().First(), result)
}
