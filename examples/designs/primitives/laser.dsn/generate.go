package example

//go:generate -command comp go run ../../../../main.go dev dsn comp
//go:generate -command geom go run ../../../../main.go dev dsn geom
//go:generate -command py uv run poe

//go:generate pwd

//go:generate py convert glb "BUY - Laser 488 nm - Bosion Laser.stp"

//go:generate comp report-comp --format=yaml _components.yml
//go:generate comp report-comp --format=json _components.json

//go:generate ./generate-newswitch-config.sh

//go:generate geom report-prim --format=yaml _primitives.yml
//go:generate geom report-prim --format=json _primitives.json

//go:generate geom render-obj --format=gltf _objects.gltf
//go:generate geom render-obj --format=glb _objects.glb
