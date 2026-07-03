#!/bin/bash

main="../../../../main.go"

script_dir=$(dirname "$(realpath "$BASH_SOURCE")")
function geom {
  go run "$main" dev dsn geom "$@"
}

cd "$script_dir"
yq '.variants | keys | .[]' optikit-design.yml | while read -r variant; do
  geom --variant="$variant" report-prim --format=json "_primitives:$variant.json"
  geom --variant="$variant" report-prim --format=yaml "_primitives:$variant.yml"
done
