package example

//go:generate -command comp go run ../../../../../main.go dev dsn comp
//go:generate -command geom go run ../../../../../main.go dev dsn geom
//go:generate -command mdl go run ../../../../../main.go dev mdl

//go:generate pwd

//go:generate mdl convert --output-format=glb "BUY - Lens - f50 D40 sA2.1 sI10.1 cx-cx.stp"
//go:generate mdl convert --output-format=glb "PRT - 2003 - NUTD40.stp"
//go:generate mdl convert --output-format=glb "PRT - 2023 - INSLEND40F50 - V04.stp"

//go:generate comp render-comps-g --format=dot _components-graph.dot
//go:generate comp render-comps-g --format=svg _components-graph.svg

//go:generate geom report-prim --format=json _primitives.json
//go:generate geom report-prim --format=yaml _primitives.yml

//go:generate geom render-obj --format=gltf _objects.gltf
//go:generate geom render-obj --format=glb _objects.glb
