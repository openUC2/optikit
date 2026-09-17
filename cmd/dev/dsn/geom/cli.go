// Package geom provides subcommands for the development design's geometry
package geom

import (
	"fmt"
	"slices"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/openUC2/optikit/internal/optikit"
)

func MakeCmd(versions optikit.Versions) *cli.Command {
	return &cli.Command{
		Name:    "geom",
		Aliases: []string{"geometry"},
		Usage:   "Facilitates development and maintenance of the design's geometry",
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
		Commands: slices.Concat(
			makeReportCmds(versions),
			makeRenderCmds(versions),
		),
	}
}

func makeReportCmds(versions optikit.Versions) []*cli.Command {
	return []*cli.Command{
		{
			Name:    "report-prim",
			Aliases: []string{"report-primitives"},
			Usage: "Generate a report of the model files and poses of all primitives in the " +
				"design",
			ArgsUsage: argsUsageOutputFile,
			Flags: []cli.Flag{
				makeRenderOutputFormatFlag("json", "yaml", "yml"),
			},
			Action: reportPrimA(versions),
		},
	}
}

func makeRenderCmds(versions optikit.Versions) []*cli.Command {
	return []*cli.Command{
		{
			Name:      "render-obj",
			Aliases:   []string{"render-objects"},
			Usage:     "Render the assembly as a 3D model object",
			ArgsUsage: argsUsageOutputFile,
			Flags: []cli.Flag{
				makeRenderOutputFormatFlag("glb", "gltf", "step"),
			},
			Action: renderObjA(versions),
		},
		{
			Name:      "render-pos-g",
			Aliases:   []string{"render-positions-graph"},
			Usage:     "Render a graph of the position relationships between the components",
			ArgsUsage: argsUsageOutputFile,
			Flags: []cli.Flag{
				makeRenderOutputFormatFlag("dot", "svg"),
			},
			Action: renderPosGA,
		},
		{
			Name:      "render-pos-p",
			Aliases:   []string{"render-positions-plot"},
			Usage:     "Render a scatterplot of the positions of the components",
			ArgsUsage: argsUsageOutputFile,
			Flags: []cli.Flag{
				makeRenderOutputFormatFlag("html"),
			},
			Action: renderPosPA,
		},
	}
}

const argsUsageOutputFile = "output_file"

func makeRenderOutputFormatFlag(formats ...string) *cli.StringFlag {
	return &cli.StringFlag{
		Name:  "format",
		Value: formats[0],
		Usage: fmt.Sprintf("Render output format (%s)", strings.Join(formats, ", ")),
	}
}
