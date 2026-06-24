package build123d

import (
	"embed"
	"io/fs"
)

//go:embed cad
var _source embed.FS
var Source, _ = fs.Sub(_source, "cad")
