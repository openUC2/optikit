package example

//go:generate -command comp go run ../../../../main.go dev dsn comp
//go:generate -command geom go run ../../../../main.go dev dsn geom

//go:generate pwd

//go:generate comp render-comp-g --format=dot _components-graph.dot
//go:generate comp render-comp-g --format=svg _components-graph.svg

//go:generate comp render-dsn-g --format=dot _designs-graph.dot
//go:generate comp render-dsn-g --format=svg _designs-graph.svg

//go:generate comp report-comp --format=yaml _components.yml
//go:generate comp report-comp --format=json _components.json

//go:generate geom report-prim --format=yaml _primitives.yml
//go:generate geom report-prim --format=json _primitives.json

//go:generate geom render-pos-g --format=dot _positions-graph.dot
//go:generate geom render-pos-g --format=svg _positions-graph.svg
//go:generate geom render-pos-p _positions-plot.html

//go:generate geom render-obj --format=gltf _objects.gltf
//go:generate geom render-obj --format=glb _objects.glb
