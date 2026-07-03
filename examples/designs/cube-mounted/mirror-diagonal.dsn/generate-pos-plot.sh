#!/bin/bash

main="../../../../main.go"

script_dir=$(dirname "$(realpath "$BASH_SOURCE")")
function geom {
  go run "$main" dev dsn geom "$@"
}

cd "$script_dir"
yq '.variants | keys | .[]' optikit-design.yml | while read -r variant; do
  geom --variant="$variant" render-pos-p "_positions-plot:$variant.html"
done
