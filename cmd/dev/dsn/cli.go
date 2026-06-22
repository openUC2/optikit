// Package dsn provides subcommands for the development design
package dsn

import (
	"github.com/urfave/cli/v3"

	"github.com/openUC2/optikit/cmd/dev/dsn/geom"
	"github.com/openUC2/optikit/internal/optikit"
)

func MakeCmd(versions optikit.Versions) *cli.Command {
	return &cli.Command{
		Name:    "dsn",
		Aliases: []string{"design"},
		Usage: "Facilitates development and maintenance of an Optikit design in the current working " +
			"directory",
		Commands: []*cli.Command{
			geom.MakeCmd(versions),
		},
	}
}
