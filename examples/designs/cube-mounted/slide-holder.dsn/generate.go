package example

//go:generate -command mdl go run ../../../../main.go dev mdl

//go:generate mdl convert --output-format=glb "SUB - 0027 - SAMHOL - V04 - virt ass.stp"

//go:generate ./generate-pos-plot.sh
