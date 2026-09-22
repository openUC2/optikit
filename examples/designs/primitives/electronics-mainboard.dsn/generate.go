package example

//go:generate -command comp go run ../../../../main.go dev dsn comp
//go:generate -command geom go run ../../../../main.go dev dsn geom
//go:generate -command mdl go run ../../../../main.go dev mdl

//go:generate pwd

//go:generate mdl convert --output-format=glb "PRT - 1009 - CASMAIBRD2X2BAS - B.stp"
//go:generate mdl convert --output-format=glb "PRT - 1010 - CASMAIBRD2X2LID - B.stp"

//go:generate geom report-prim --format=yaml _primitives.yml
//go:generate geom report-prim --format=json _primitives.json

//go:generate geom render-obj --format=gltf _objects.gltf
//go:generate geom render-obj --format=glb _objects.glb
