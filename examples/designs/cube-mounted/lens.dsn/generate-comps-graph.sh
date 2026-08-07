#!/bin/bash

main="../../../../main.go"

script_dir=$(dirname "$(realpath "$BASH_SOURCE")")
function comp {
  go run "$main" dev dsn comp "$@"
}

cd "$script_dir"
yq '.variants | keys | .[]' optikit-design.yml | while read -r variant; do
  comp --variant="$variant" render-comps-g --format=dot "_components-graph:$variant.dot"
  comp --variant="$variant" render-comps-g --format=svg "_components-graph:$variant.svg"
done
