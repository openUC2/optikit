script_dir=$(dirname "$(realpath "$BASH_SOURCE")")
main="$(realpath "$script_dir/../../../../main.go")"
comp() { go run "$main" dev dsn comp "$@"; }
geom() { go run "$main" dev dsn geom "$@"; }

pwd

comp render-comp-g --format=dot _components-graph.dot
comp render-comp-g --format=svg _components-graph.svg

comp render-dsn-g --format=dot _designs-graph.dot
comp render-dsn-g --format=svg _designs-graph.svg

geom render-assm-g --format=dot _assembly-graph.dot
geom render-assm-g --format=svg _assembly-graph.svg

geom report-assm --format=yaml _assembly.yml

geom render-obj --format=gltf _objects.gltf
geom render-obj --format=glb _objects.glb
