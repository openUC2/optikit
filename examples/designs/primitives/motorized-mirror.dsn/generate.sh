script_dir=$(dirname "$(realpath "$BASH_SOURCE")")
main="$(realpath "$script_dir/../../../../main.go")"
comp() { go run "$main" dev dsn comp "$@"; }
geom() { go run "$main" dev dsn geom "$@"; }

pwd

comp render-comp-g --format=dot _components-graph.dot
comp render-comp-g --format=svg _components-graph.svg

geom report-assm --format=yaml _assembly.yml

geom render-obj --format=gltf _objects.gltf
geom render-obj --format=glb _objects.glb
