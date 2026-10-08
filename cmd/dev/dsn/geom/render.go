package geom

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v3"

	"github.com/openUC2/optikit/cmd/compat"
	"github.com/openUC2/optikit/exp/designs"
	ofs "github.com/openUC2/optikit/exp/fs"
	"github.com/openUC2/optikit/internal/optikit"
)

func renderObjA(versions compat.Versions) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		inputs, err := parseInputVars(c.StringSlice("input"))
		if err != nil {
			return errors.Wrap(err, "couldn't parse input variables")
		}
		design, err := optikit.LoadFSDesign(
			ctx, c.String("cwd"), inputs, false,
		)
		if err != nil {
			return err
		}

		fsys := ofs.AttachPath(os.DirFS(c.String("cwd")), c.String("cwd"))
		result, err := optikit.RenderObjects(
			ctx, fsys, design, designs.AssmID(c.String("assembly")), c.String("format"), versions.Tool,
		)
		if err != nil {
			return err
		}

		return produceOutput(c.Args().First(), result)
	}
}

func produceOutput(outputPath string, output []byte) error {
	if outputPath == "" {
		fmt.Println(string(output))
		return nil
	}
	const perms = 0o644
	return os.WriteFile(outputPath, output, perms)
}

func renderAssmGA(ctx context.Context, c *cli.Command) error {
	inputs, err := parseInputVars(c.StringSlice("input"))
	if err != nil {
		return errors.Wrap(err, "couldn't parse input variables")
	}
	design, err := optikit.LoadFSDesign(ctx, c.String("cwd"), inputs, false)
	if err != nil {
		return err
	}

	result, err := optikit.RenderAssemblyGraph(
		ctx, design, designs.AssmID(c.String("assembly")), c.String("format"), true,
	)
	if err != nil {
		return err
	}

	return produceOutput(c.Args().First(), result)
}

func parseInputVars(args []string) (map[designs.VarName]any, error) {
	inputs := make(map[designs.VarName]any)
	for _, arg := range args {
		varName, varValue, ok := strings.Cut(arg, ":")
		if !ok {
			return nil, errors.Errorf("%s couldn't be parsed as varname:varvalue", arg)
		}
		inputs[designs.VarName(varName)] = varValue
	}
	return inputs, nil
}
