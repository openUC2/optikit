package comp

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/openUC2/optikit/internal/optikit"
)

func renderCompsGA(ctx context.Context, c *cli.Command) error {
	design, err := optikit.LoadFSDesign(c.String("cwd"), c.String("variant"), false)
	if err != nil {
		return err
	}

	result, err := optikit.RenderComponentsGraph(ctx, design, c.String("format"), true)
	if err != nil {
		return err
	}

	return produceOutput(c.Args().First(), result)
}

func renderDsnsGA(ctx context.Context, c *cli.Command) error {
	design, err := optikit.LoadFSDesign(c.String("cwd"), c.String("variant"), false)
	if err != nil {
		return err
	}

	result, err := optikit.RenderDesignsGraph(ctx, design, c.String("format"), true)
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
