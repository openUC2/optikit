package geom

import (
	"context"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v3"

	"github.com/openUC2/optikit/exp/designs"
	"github.com/openUC2/optikit/internal/optikit"
)

func reportPrimA(ctx context.Context, c *cli.Command) error {
	inputs, err := parseInputVars(c.StringSlice("input"))
	if err != nil {
		return errors.Wrap(err, "couldn't parse input variables")
	}
	design, err := optikit.LoadFSDesign(
		ctx, c.String("cwd"), designs.VariantID(c.String("variant")), inputs, false,
	)
	if err != nil {
		return err
	}

	format := c.String("format")
	if format == "yml" {
		format = "yaml"
	}
	result, err := optikit.ReportPrimitives(ctx, design, designs.UC2GridSpacings, format)
	if err != nil {
		return err
	}

	return produceOutput(c.Args().First(), result)
}
