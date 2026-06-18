// Package dsn provides subcommands for the development design
package dsn

import (
	"github.com/urfave/cli/v2"

	"github.com/openUC2/optikit/internal/app/optikit"
)

func MakeCmd(_ optikit.Versions) *cli.Command {
	return &cli.Command{
		Name:    "dsn",
		Aliases: []string{"design"},
		Usage: "Facilitates development and maintenance of an Optikit design in the current working " +
			"directory",
	}
}
