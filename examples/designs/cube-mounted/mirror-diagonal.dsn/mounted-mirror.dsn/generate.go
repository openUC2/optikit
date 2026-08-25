package example

//go:generate -command comp go run ../../../../../main.go dev dsn comp
//go:generate -command geom go run ../../../../../main.go dev dsn geom
//go:generate -command mdl go run ../../../../../main.go dev mdl

//go:generate pwd

//go:generate mdl convert --output-format=glb "BUY - Mirror - 40x30x2.stp"
//go:generate mdl convert --output-format=glb "BUY - Adhesive pad - 41x29x1.stp"
//go:generate mdl convert --output-format=glb "PRT - 2022 - INSMIR45TH2.stp"

//go:generate comp render-comps-g --format=dot _components-graph.dot
//go:generate comp render-comps-g --format=svg _components-graph.svg

//go:generate geom report-prim --format=yaml _primitives.yml
//go:generate geom report-prim --format=json _primitives.json

//go:generate geom render-obj --format=gltf _objects.gltf
//go:generate geom render-obj --format=glb _objects.glb
