package example

//go:generate -command py uv run poe

//go:generate pwd

//go:generate py convert glb "SUB - 0023 - LEND40F50 - V04 - virt ass.stp"

//go:generate ./generate-variants.sh generate-variants.directives
