package example

//go:generate -command geom go run ../../../../main.go dev dsn geom

//go:generate geom --variant=xy render-pos-p _positions-plot:xy.html
//go:generate geom --variant=_z render-pos-p _positions-plot:_z.html
