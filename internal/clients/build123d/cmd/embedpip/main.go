package main

import (
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/kluctl/go-embed-python/pip"
)

func main() {
	inputFile := "requirements.frozen.txt"
	if len(os.Args) > 1 && os.Args[1] != "" {
		inputFile = os.Args[1]
	}

	var platforms []string
	if len(os.Args) > 2 {
		platforms = os.Args[2:]
	}
	err := createEmbeddedPipPackages(inputFile, "./data/", platforms)
	if err != nil {
		panic(err)
	}
}

func createEmbeddedPipPackages(requirementsFile, targetDir string, platforms []string) error {
	if len(platforms) == 0 || platforms[0] == "" || platforms[0] == "*" {
		platforms = slices.Collect(maps.Keys(supportedPlatforms))
	}

	for _, goPlatform := range platforms {
		pipPlatforms := supportedPlatforms[goPlatform]
		s := strings.Split(goPlatform, "-")
		goOs, goArch := s[0], s[1]
		err := pip.CreateEmbeddedPipPackages(requirementsFile, goOs, goArch, pipPlatforms, targetDir)
		if err != nil {
			return err
		}
	}
	return nil
}

var supportedPlatforms = map[string][]string{
	"linux-amd64": {
		"manylinux_2_27_x86_64",
		"manylinux_2_28_x86_64",
		"manylinux_2_31_x86_64",
		"manylinux2014_x86_64",
	},
	// Note: we can't support linux-arm64 yet because lib3mf isn't yet built for linux-arm64!
	// See https://github.com/3MFConsortium/lib3mf/issues/443 for details.
	// "linux-arm64": {
	// 	"manylinux_2_27_aarch64",
	// 	"manylinux_2_28_aarch64",
	// 	"manylinux_2_31_aarch64",
	// 	"manylinux2014_aarch64",
	// },
	"darwin-amd64": {"macosx_11_0_x86_64", "macosx_12_0_x86_64"},
	"darwin-arm64": {"macosx_11_0_arm64", "macosx_12_0_arm64"},
}
