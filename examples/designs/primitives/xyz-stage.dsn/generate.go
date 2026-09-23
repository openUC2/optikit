package example

//go:generate -command comp go run ../../../../main.go dev dsn comp
//go:generate -command geom go run ../../../../main.go dev dsn geom
//go:generate -command mdl go run ../../../../main.go dev mdl

//go:generate pwd

//go:generate mdl convert --output-format=glb "BUY - Manual Z stage - LZ40 --- slide.stp"

//go:generate comp render-comp-g --format=dot _components-graph.dot
//go:generate comp render-comp-g --format=svg _components-graph.svg

//go:generate comp report-comp --format=yaml _components.yml
//go:generate comp report-comp --format=json _components.json

//go:generate geom report-prim --format=yaml _primitives.yml
//go:generate geom report-prim --format=json _primitives.json

//go:generate geom render-pos-g --format=dot _positions-graph.dot
//go:generate geom render-pos-g --format=svg _positions-graph.svg

//go:generate geom render-obj --format=gltf _objects.gltf
//go:generate geom render-obj --format=glb _objects.glb
