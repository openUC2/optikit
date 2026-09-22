// Package cmd provides subcommands
package cmd

const (
	// dsnMinVersion is the minimum supported Optikit version among designs. A design with a
	// lower Optikit version cannot be used.
	DSNMinVersion = "v0.0.0"
	// fallbackVersion is the version reported which the Optikit tool reports itself as if its actual
	// version is unknown.
	FallbackVersion = "v0.0.0-dev"
)
