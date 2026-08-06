#!/bin/bash

main="../../../../main.go"

script_dir=$(dirname "$(realpath "$BASH_SOURCE")")
function comp {
  go run "$main" dev dsn comp "$@"
}

cd "$script_dir"
yq '.variants | keys | .[]' optikit-design.yml | while read -r variant; do
  comp --variant="$variant" render-dsns-g --format=dot "_designs-graph:$variant.dot"
  comp --variant="$variant" render-dsns-g --format=svg "_designs-graph:$variant.svg"
done
