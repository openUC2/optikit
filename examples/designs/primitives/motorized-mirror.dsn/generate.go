package example

//go:generate -command comp go run ../../../../main.go dev dsn comp
//go:generate -command geom go run ../../../../main.go dev dsn geom
//go:generate -command py uv run poe

//go:generate pwd

//go:generate py convert glb "BUY - Galvo scanner XY --- single motor.stp"
//go:generate py convert glb "PRT - 2165 - CLPPCE.stp"
//go:generate py convert glb "PRT - 2180 - DISRNG.stp"

//go:generate comp render-comp-g --format=dot _components-graph.dot
//go:generate comp render-comp-g --format=svg _components-graph.svg

//go:generate geom report-prim --format=yaml _primitives.yml
//go:generate geom report-prim --format=json _primitives.json

//go:generate geom render-obj --format=gltf _objects.gltf
//go:generate geom render-obj --format=glb _objects.glb
