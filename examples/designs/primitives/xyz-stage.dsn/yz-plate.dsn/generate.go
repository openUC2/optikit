package example

//go:generate -command comp go run ../../../../../main.go dev dsn comp
//go:generate -command geom go run ../../../../../main.go dev dsn geom
//go:generate -command py uv run poe

//go:generate pwd

//go:generate py convert glb "BUY - Manual Z stage - LZ40 --- base.stp"
//go:generate py convert glb "BUY - Manual Z stage - LZ40 --- converter.stp"
//go:generate py convert glb "BUY - XYZ linear table - LD40 --- bracket.stp"
//go:generate py convert glb "BUY - XYZ linear table - LD40 --- Y plate.stp"
//go:generate py convert glb "PRT - 3024 - MOTHOLZST.stp"
