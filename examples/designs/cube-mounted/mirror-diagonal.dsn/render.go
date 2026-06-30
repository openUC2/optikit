package example

//go:generate -command geom go run ../../../../main.go dev dsn geom
//go:generate -command mdl go run ../../../../main.go dev mdl

//go:generate mdl convert --output-format=glb "SUB - 0022 - MIR45TH2 - virt ass.stp"

//go:generate geom --variant=xy render-pos-p _positions-plot:xy.html
//go:generate geom --variant=_z render-pos-p _positions-plot:_z.html
