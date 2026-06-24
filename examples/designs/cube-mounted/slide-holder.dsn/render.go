package example

//go:generate -command geom go run ../../../../main.go dev dsn geom

//go:generate geom --variant=z render-pos-p _positions-plot:z.html
//go:generate geom --variant=y render-pos-p _positions-plot:y.html
//go:generate geom --variant=x render-pos-p _positions-plot:x.html
