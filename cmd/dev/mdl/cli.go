// Package mdl provides subcommands for the development design
package mdl

import (
	"fmt"

	"github.com/urfave/cli/v3"

	"github.com/openUC2/optikit/internal/optikit"
)

func MakeCmd(_ optikit.Versions) *cli.Command {
	return &cli.Command{
		Name:    "mdl",
		Aliases: []string{"model"},
		Usage:   "Facilitates development and maintenance of model for a design primitive",
		Commands: []*cli.Command{
			{
				Name: "convert",
				Usage: "Convert the model file to an output format (manually set or inferred from " +
					"output file extension if output_file is specified)",
				ArgsUsage: "input_file [output_file]",
				Action:    convertA,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "input-format",
						Value: formatStep,
						Usage: fmt.Sprintf("Manually set format of input_file (%s)", formatStep),
					},
					&cli.StringFlag{
						Name: "output-format",
						Usage: fmt.Sprintf(
							"Manually set format of output_file (%s, %s, or %s)",
							formatStep, formatGLTF, formatGLB,
						),
					},
				},
			},
		},
	}
}
