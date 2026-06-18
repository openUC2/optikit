package main

import (
	"log"
	"os"
	"runtime/debug"

	"github.com/carlmjohnson/versioninfo"
	"github.com/urfave/cli/v2"

	"github.com/openUC2/optikit/cmd/dev"
	"github.com/openUC2/optikit/internal/optikit"
)

func main() {
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

var defaultWorkspaceBase, _ = os.UserHomeDir()

var ocliVersions optikit.Versions = optikit.Versions{
	Tool:               toolVersion,
	MinSupportedDesign: dsnMinVersion,
}

var app = &cli.App{
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
			EnvVars: []string{"OPTIKIT_WORKSPACE"},
		},
		&cli.BoolFlag{
			Name:    "ignore-tool-version",
			Value:   false,
			Usage:   "Ignore the version of the optikit tool in version compatibility checks",
			EnvVars: []string{"OPTIKIT_IGNORE_TOOL_VERSION"},
		},
		&cli.BoolFlag{
			Name:  "parallel",
			Value: true,
			Usage: "Allow parallel execution of I/O-bound tasks, such as downloading container images " +
				"or starting containers",
			EnvVars: []string{"OPTIKIT_PARALLEL"},
		},
	},
	Suggest: true,
}

// Versioning

const (
	// dsnMinVersion is the minimum supported Optikit version among designs. A design with a
	// lower Optikit version cannot be used.
	dsnMinVersion = "v0.0.0"
	// fallbackVersion is the version reported which the Optikit tool reports itself as if its actual
	// version is unknown.
	fallbackVersion = "v0.0.0-dev"
)

var (
	toolVersion = determineVersion(buildSummary, fallbackVersion)
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
