package example

//go:generate -command comp go run ../../../../main.go dev dsn comp
//go:generate -command geom go run ../../../../main.go dev dsn geom
//go:generate -command py uv run poe

//go:generate pwd

//go:generate comp render-comp-g --format=dot _components-graph.dot
//go:generate comp render-comp-g --format=svg _components-graph.svg

//go:generate comp render-dsn-g --format=dot _designs-graph.dot
//go:generate comp render-dsn-g --format=svg _designs-graph.svg

//go:generate comp report-comp --format=yaml _components.yml
//go:generate comp report-comp --format=json _components.json

//go:generate py config _components.yml _newswitch-config.yml

//go:generate geom report-assm --format=yaml _assembly.yml
//go:generate geom report-assm --format=json _assembly.json

//go:generate geom render-assm-g --format=dot _assembly-graph.dot
//go:generate geom render-assm-g --format=svg _assembly-graph.svg

//go:generate geom render-obj --format=gltf _objects.gltf
//go:generate geom render-obj --format=glb _objects.glb
