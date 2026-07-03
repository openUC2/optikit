// Package dev provides subcommands for developing pallets, packages, etc.
package dev

import (
	"os"

	"github.com/urfave/cli/v3"

	"github.com/openUC2/optikit/cmd/dev/dsn"
	"github.com/openUC2/optikit/cmd/dev/mdl"
	"github.com/openUC2/optikit/internal/optikit"
)

var defaultWorkingDir, _ = os.Getwd()

func MakeCmd(versions optikit.Versions) *cli.Command {
	return &cli.Command{
		Name:    "dev",
		Aliases: []string{"development"},
		Usage:   "Facilitates development and maintenance in the current working directory",
		Commands: []*cli.Command{
			dsn.MakeCmd(versions),
			mdl.MakeCmd(versions),
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "cwd",
				Value:   defaultWorkingDir,
				Usage:   "Path of the current working directory",
				Sources: cli.EnvVars("OPTIKIT_CWD"),
			},
		},
	}
}
