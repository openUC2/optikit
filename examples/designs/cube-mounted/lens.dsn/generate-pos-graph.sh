#!/bin/bash

main="../../../../main.go"

script_dir=$(dirname "$(realpath "$BASH_SOURCE")")
function geom {
  echo "go run \"$main\" dev dsn geom"
}

cd "$script_dir"
yq '.variants | keys | .[]' optikit-design.yml |
  go tool rush -k -e \
    "
      $(geom) --variant=\"{}\" render-pos-g --format=dot \"_positions-graph:{}.dot\"
      $(geom) --variant=\"{}\" render-pos-g --format=svg \"_positions-graph:{}.svg\"
    "

if [[ $? != 0 ]]; then
  echo "A variant couldn't be generated; you can find error messages above by searching for [ERRO]"
fi
