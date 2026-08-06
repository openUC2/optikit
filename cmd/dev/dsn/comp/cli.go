// Package comp provides subcommands for the development design's composition
package comp

import (
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
				Name:  "variant",
				Usage: "Select design variant",
			},
		},
		Commands: cmds,
	}
}

var cmds = []*cli.Command{
	{
		Name:      "render-comps-g",
		Aliases:   []string{"render-components-graph"},
		Usage:     "Render a graph of the composition relationships between the components",
		ArgsUsage: "output_file",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "format",
				Value: "dot",
				Usage: "Render output format (dot or svg)",
			},
		},
		Action: renderCompsGA,
	},
	{
		Name:      "render-dsns-g",
		Aliases:   []string{"render-designs-graph"},
		Usage:     "Render a graph of the composition relationships between designs",
		ArgsUsage: "output_file",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "format",
				Value: "dot",
				Usage: "Render output format (dot or svg)",
			},
		},
		Action: renderDsnsGA,
	},
}
