#!/bin/bash

main="../../../../main.go"

script_dir=$(dirname "$(realpath "$BASH_SOURCE")")
function comp {
  echo "go run \"$main\" dev dsn comp"
}

cd "$script_dir"
yq '.variants | keys | .[]' optikit-design.yml |
  go tool rush -k -e \
    "
      $(comp) --variant=\"{}\" render-comps-g --format=dot \"_components-graph:{}.dot\"
      $(comp) --variant=\"{}\" render-comps-g --format=svg \"_components-graph:{}.svg\"
    "

if [[ $? != 0 ]]; then
  echo "A variant couldn't be generated; you can find error messages above by searching for [ERRO]"
fi
