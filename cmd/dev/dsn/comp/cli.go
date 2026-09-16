// Package comp provides subcommands for the development design's composition
package comp

import (
	"fmt"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/openUC2/optikit/internal/optikit"
)

func MakeCmd(_ optikit.Versions) *cli.Command {
	return &cli.Command{
		Name:    "comp",
		Aliases: []string{"composition"},
		Usage:   "Facilitates development and maintenance of the design's composition",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "variant",
				Aliases: []string{"v"},
				Usage:   "Select design variant",
			},
			&cli.StringSliceFlag{
				Name:    "input",
				Aliases: []string{"i"},
				Usage:   "Set value of input variable, using format `variablename:variablevalue`",
			},
		},
		Commands: cmds,
	}
}

var cmds = []*cli.Command{
	{
		Name:      "report-comp",
		Aliases:   []string{"report-components"},
		Usage:     "Generate a flattened report of all components in the design",
		ArgsUsage: argsUsageOutputFile,
		Flags: []cli.Flag{
			makeRenderOutputFormatFlag("json", "yaml", "yml"),
		},
		Action: reportCompA,
	},
	{
		Name:      "render-comp-g",
		Aliases:   []string{"render-components-graph"},
		Usage:     "Render a graph of the composition relationships between the components",
		ArgsUsage: "output_file",
		Flags: []cli.Flag{
			makeRenderOutputFormatFlag("dot", "svg"),
		},
		Action: renderCompGA,
	},
	{
		Name:      "render-dsn-g",
		Aliases:   []string{"render-designs-graph"},
		Usage:     "Render a graph of the composition relationships between designs",
		ArgsUsage: "output_file",
		Flags: []cli.Flag{
			makeRenderOutputFormatFlag("dot", "svg"),
		},
		Action: renderDsnGA,
	},
}

const argsUsageOutputFile = "output_file"

func makeRenderOutputFormatFlag(formats ...string) *cli.StringFlag {
	return &cli.StringFlag{
		Name:  "format",
		Value: formats[0],
		Usage: fmt.Sprintf("Render output format (%s)", strings.Join(formats, ", ")),
	}
}
