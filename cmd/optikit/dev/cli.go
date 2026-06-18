// Package dev provides subcommands for developing pallets, packages, etc.
package dev

import (
	"os"

	"github.com/urfave/cli/v2"

	"github.com/openUC2/optikit/cmd/optikit/dev/dsn"
	"github.com/openUC2/optikit/internal/app/optikit"
)

var defaultWorkingDir, _ = os.Getwd()

func MakeCmd(versions optikit.Versions) *cli.Command {
	return &cli.Command{
		Name:    "dev",
		Aliases: []string{"development"},
		Usage:   "Facilitates development and maintenance in the current working directory",
		Subcommands: []*cli.Command{
			dsn.MakeCmd(versions),
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "cwd",
				Value:   defaultWorkingDir,
				Usage:   "Path of the current working directory",
				EnvVars: []string{"OPTIKIT_CWD"},
			},
		},
	}
}
