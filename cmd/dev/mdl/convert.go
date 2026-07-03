package mdl

import (
	"cmp"
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v3"

	"github.com/openUC2/optikit/internal/optikit"
)

func convertA(ctx context.Context, c *cli.Command) error {
	inputPath := c.Args().First()
	outputPath := c.Args().Get(1)
	inputFormat, err := detectFormat(inputPath, c.String("input-format"))
	if err != nil {
		return errors.Wrap(err, "unknown input format")
	}
	outputFormat, err := detectFormat(outputPath, c.String("output-format"))
	if err != nil {
		return errors.Wrap(err, "unknown output format")
	}
	if outputPath == "" {
		outputPath = fmt.Sprintf(
			"%s.%s", strings.TrimSuffix(inputPath, path.Ext(inputPath)), outputFormat,
		)
	}

	if err = optikit.ConvertModel(
		ctx, inputFormat, inputPath, outputFormat, outputPath,
	); err != nil {
		return err
	}
	return nil
}

func detectFormat(f, format string) (string, error) {
	switch ext := strings.TrimPrefix(cmp.Or(format, path.Ext(f)), "."); strings.ToLower(ext) {
	default:
		return "", errors.Errorf("unknown file format for %s with file extension %s", f, ext)
	case "stp":
		return "step", nil
	case "gltf", "glb", "step":
		return ext, nil
	}
}
