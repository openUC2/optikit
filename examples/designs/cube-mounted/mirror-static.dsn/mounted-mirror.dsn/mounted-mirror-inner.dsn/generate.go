package example

//go:generate -command comp go run ../../../../../../main.go dev dsn comp
//go:generate -command geom go run ../../../../../../main.go dev dsn geom
//go:generate -command py uv run poe

//go:generate pwd

//go:generate py convert glb "BUY - Adhesive pad - D22x1.stp"
//go:generate py convert glb "BUY - Mirror - D24x3.stp"
//go:generate py convert glb "PRT - 2110 - MASINSMIRHOLLOW - C.stp"
//go:generate py convert glb "PRT - 2111 - MASINSMIRHOLUPP - C.stp"

//go:generate comp render-comp-g --format=dot _components-graph.dot
//go:generate comp render-comp-g --format=svg _components-graph.svg

//go:generate geom report-prim --format=yaml _primitives.yml
//go:generate geom report-prim --format=json _primitives.json

//go:generate geom render-obj --format=gltf _objects.gltf
//go:generate geom render-obj --format=glb _objects.glb
