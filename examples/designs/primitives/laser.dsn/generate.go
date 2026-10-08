package example

//go:generate -command comp go run ../../../../main.go dev dsn comp
//go:generate -command geom go run ../../../../main.go dev dsn geom

//go:generate pwd

//go:generate comp report-comp --format=yaml _components.yml
//go:generate comp report-comp --format=json _components.json

//go:generate ./generate-newswitch-config.sh

//go:generate geom report-assm --format=yaml _assembly.yml
//go:generate geom report-assm --format=json _assembly.json

//go:generate geom render-obj --format=gltf _objects.gltf
//go:generate geom render-obj --format=glb _objects.glb
