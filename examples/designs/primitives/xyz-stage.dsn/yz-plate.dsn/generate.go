package example

//go:generate -command comp go run ../../../../../main.go dev dsn comp
//go:generate -command geom go run ../../../../../main.go dev dsn geom
//go:generate -command mdl go run ../../../../../main.go dev mdl

//go:generate pwd

//go:generate mdl convert --output-format=glb "BUY - Manual Z stage - LZ40 --- base.stp"
//go:generate mdl convert --output-format=glb "BUY - Manual Z stage - LZ40 --- converter.stp"
//go:generate mdl convert --output-format=glb "BUY - XYZ linear table - LD40 --- bracket.stp"
//go:generate mdl convert --output-format=glb "BUY - XYZ linear table - LD40 --- Y plate.stp"
//go:generate mdl convert --output-format=glb "PRT - 3024 - MOTHOLZST.stp"
