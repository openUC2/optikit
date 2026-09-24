package example

//go:generate -command comp go run ../../../../../main.go dev dsn comp
//go:generate -command geom go run ../../../../../main.go dev dsn geom
//go:generate -command mdl go run ../../../../../main.go dev mdl

//go:generate pwd

//go:generate geom report-prim --format=yaml _primitives.yml
//go:generate geom report-prim --format=json _primitives.json
