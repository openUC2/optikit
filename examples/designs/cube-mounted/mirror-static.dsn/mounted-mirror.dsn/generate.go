package example

//go:generate -command py uv run poe

//go:generate pwd

//go:generate py convert glb "PRT - 2100 - MASINS - V04 - B.stp"

//go:generate ./generate-variants.sh generate-variants.directives
