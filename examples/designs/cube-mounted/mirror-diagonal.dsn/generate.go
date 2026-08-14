package example

//go:generate -command mdl go run ../../../../main.go dev mdl

//go:generate pwd

//go:generate mdl convert --output-format=glb "SUB - 0022 - MIR45TH2 - virt ass.stp"

//go:generate ./generate-variant.sh comp render components-graph dot
//go:generate ./generate-variant.sh comp render components-graph svg
//go:generate ./generate-variant.sh comp render designs-graph dot
//go:generate ./generate-variant.sh comp render designs-graph svg
//go:generate ./generate-variant.sh geom render positions-graph dot
//go:generate ./generate-variant.sh geom render positions-graph svg
//go:generate ./generate-variant.sh geom render positions-plot html
//go:generate ./generate-variant.sh geom report primitives yml
//go:generate ./generate-variant.sh geom report primitives json
//go:generate ./generate-variant.sh geom render objects gltf
//go:generate ./generate-variant.sh geom render objects glb
