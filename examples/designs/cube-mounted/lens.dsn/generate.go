package example

//go:generate -command mdl go run ../../../../main.go dev mdl

//go:generate pwd

//go:generate mdl convert --output-format=glb "SUB - 0023 - LEND40F50 - V04 - virt ass.stp"

//go:generate ./generate-comps-graph.sh
//go:generate ./generate-dsns-graph.sh
//go:generate ./generate-pos-graph.sh
//go:generate ./generate-pos-plot.sh
//go:generate ./generate-prim-report.sh
//go:generate ./generate-obj-gltf.sh
//go:generate ./generate-obj-glb.sh
