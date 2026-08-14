package example

//go:generate -command mdl go run ../../../../main.go dev mdl

//go:generate pwd

//go:generate mdl convert --output-format=glb "SUB - 0023 - LEND40F50 - V04 - virt ass.stp"

//go:generate ./generate-variants.sh generate-variants.directives
