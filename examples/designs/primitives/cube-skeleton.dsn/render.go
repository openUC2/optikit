package example

//go:generate -command geom go run ../../../../main.go dev dsn geom
//go:generate -command mdl go run ../../../../main.go dev mdl

//go:generate mdl convert --output-format=glb "PRT - 1003 - CUBHLF111 - V04.stp"

//go:generate geom render-pos-g --format=dot _positions-graph.dot
//go:generate geom render-pos-g --format=svg _positions-graph.svg
//go:generate geom render-pos-p _positions-plot.html

//go:generate geom report-prim --format=json _primitives.json
//go:generate geom report-prim --format=yaml _primitives.yml
//go:generate geom render-obj _objects.step
