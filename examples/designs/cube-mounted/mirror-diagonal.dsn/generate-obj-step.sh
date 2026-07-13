#!/bin/bash

main="../../../../main.go"

script_dir=$(dirname "$(realpath "$BASH_SOURCE")")
function geom {
  go run "$main" dev dsn geom "$@"
}

cd "$script_dir"
yq '.variants | keys | .[]' optikit-design.yml | while read -r variant; do
  echo "$variant"
  geom --variant="$variant" render-obj --format=step "_objects:$variant.step"
done
