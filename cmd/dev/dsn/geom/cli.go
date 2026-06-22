// Package geom provides subcommands for the development design's geometry
package geom

import (
	"github.com/urfave/cli/v3"

	"github.com/openUC2/optikit/internal/optikit"
)

func MakeCmd(_ optikit.Versions) *cli.Command {
	return &cli.Command{
		Name:    "geom",
		Aliases: []string{"geometry"},
		Usage:   "Facilitates development and maintenance of the design's geometry",
		Commands: []*cli.Command{
			{
				Name:      "render-pos-g",
				Aliases:   []string{"render-positions-graph"},
				Usage:     "Render a graph of the position relationships between the components",
				ArgsUsage: "output_file",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "format",
						Value: "dot",
						Usage: "Render output format (dot or svg)",
					},
				},
				Action: renderPosGA,
			},
			{
				Name:      "render-pos-p",
				Aliases:   []string{"render-positions-plot"},
				Usage:     "Render a scatterplot of the positions of the components, into an HTML file",
				ArgsUsage: "output_file",
				Action:    renderPosPA,
			},
		},
	}
}
