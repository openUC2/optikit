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
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "variant",
				Usage: "Select design variant",
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "report-prim",
				Aliases: []string{"report-primitives"},
				Usage: "Generate a report of the model files and poses of all primitives in the " +
					"design",
				ArgsUsage: "output_file",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "format",
						Value: "json",
						Usage: "Render output format (json or yaml)",
					},
				},
				Action: reportPrimA,
			},
			{
				Name:      "render-obj",
				Aliases:   []string{"render-objects"},
				Usage:     "Render the assembly as a 3D model object",
				ArgsUsage: "output_file",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "format",
						Value: "step",
						Usage: "Render output format (step)",
					},
				},
				Action: renderObjA,
			},
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
