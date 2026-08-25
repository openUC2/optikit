package example

//go:generate -command mdl go run ../../../../../main.go dev mdl

//go:generate pwd

//go:generate mdl convert --output-format=glb "PRT - 2100 - MASINS - V04 - B.stp"

//go:generate ./generate-variants.sh generate-variants.directives
