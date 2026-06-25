package example

//go:generate -command geom go run ../../../../main.go dev dsn geom

//go:generate geom render-pos-g --format=dot _positions-graph.dot
//go:generate geom render-pos-g --format=svg _positions-graph.svg
//go:generate geom render-pos-p _positions-plot.html

//go:generate geom report-prim --format=json _primitives.json
//go:generate geom report-prim --format=yaml _primitives.yml
//go:generate geom render-obj _objects.step
