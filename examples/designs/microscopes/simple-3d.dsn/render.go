package example

//go:generate -command geom go run ../../../../main.go dev dsn geom

//go:generate geom render-obj _objects.step
//go:generate geom render-pos-g --format=dot _positions-graph.dot
//go:generate geom render-pos-g --format=svg _positions-graph.svg
//go:generate geom render-pos-p _positions-plot.html
