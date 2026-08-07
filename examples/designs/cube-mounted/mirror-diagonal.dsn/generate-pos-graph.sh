#!/bin/bash

main="../../../../main.go"

script_dir=$(dirname "$(realpath "$BASH_SOURCE")")
function geom {
  go run "$main" dev dsn geom "$@"
}

cd "$script_dir"
yq '.variants | keys | .[]' optikit-design.yml | while read -r variant; do
  geom --variant="$variant" render-pos-g --format=dot "_positions-graph:$variant.dot"
  geom --variant="$variant" render-pos-g --format=svg "_positions-graph:$variant.svg"
done
