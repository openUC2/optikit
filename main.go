package main

import (
	"context"
	"log"
	"os"
	"runtime/debug"

	"github.com/carlmjohnson/versioninfo"
	"github.com/urfave/cli/v3"

	"github.com/openUC2/optikit/cmd"
	"github.com/openUC2/optikit/cmd/compat"
	"github.com/openUC2/optikit/cmd/dev"
)

func main() {
	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

var defaultWorkspaceBase, _ = os.UserHomeDir()

var ocliVersions compat.Versions = compat.Versions{
	Tool:               toolVersion,
	MinSupportedDesign: cmd.DSNMinVersion,
}

var app = &cli.Command{
	Name:    "optikit",
	Version: toolVersion,
	Usage:   "Manages pallets and package deployments",
	Commands: []*cli.Command{
		dev.MakeCmd(ocliVersions),
	},
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "workspace",
			Aliases: []string{"ws"},
			Value:   defaultWorkspaceBase,
			Usage:   "Path of the optikit workspace",
			Sources: cli.EnvVars("OPTIKIT_WORKSPACE"),
		},
		&cli.BoolFlag{
			Name:    "ignore-tool-version",
			Value:   false,
			Usage:   "Ignore the version of the optikit tool in version compatibility checks",
			Sources: cli.EnvVars("OPTIKIT_IGNORE_TOOL_VERSION"),
		},
		&cli.BoolFlag{
			Name:  "parallel",
			Value: true,
			Usage: "Allow parallel execution of I/O-bound tasks, such as downloading container images " +
				"or starting containers",
			Sources: cli.EnvVars("OPTIKIT_PARALLEL"),
		},
	},
	EnableShellCompletion: true,
	Suggest:               true,
}

// Versioning

var (
	toolVersion = determineVersion(buildSummary, cmd.FallbackVersion)
	// buildSummary should be overridden by ldflags, such as with GoReleaser's "Summary".
	buildSummary = ""
)

// determineVersion returns either a semver, a pseudoversion, or a Git hash based on information
// available from Go's `debug.ReadBuildInfo()`.
func determineVersion(override, fallback string) string {
	if override != "" {
		return override
	}

	const dirtySuffix = "-dirty"
	// Determine any version tags, if available
	if info, ok := debug.ReadBuildInfo(); ok &&
		info.Main.Version != "" && info.Main.Version != "(devel)" {
		v := info.Main.Version
		if versioninfo.DirtyBuild {
			v += dirtySuffix
		}
		return v
	}
	if v := versioninfo.Version; v != "unknown" && v != "(devel)" {
		if versioninfo.DirtyBuild {
			v += dirtySuffix
		}
		return v
	}

	// Fall back to whatever is available
	if r := versioninfo.Revision; r != "unknown" && r != "" {
		if versioninfo.DirtyBuild {
			r += dirtySuffix
		}
		return r
	}
	return fallback
}
