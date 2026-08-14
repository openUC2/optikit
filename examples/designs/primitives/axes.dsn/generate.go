package example

//go:generate pwd

//go:generate ./generate-variant.sh geom report primitives yml
//go:generate ./generate-variant.sh geom report primitives json
//go:generate ./generate-variant.sh geom render objects gltf
//go:generate ./generate-variant.sh geom render objects glb
