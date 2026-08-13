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
      $(comp) --variant=\"{}\" render-dsns-g --format=dot \"_designs-graph:{}.dot\"
      $(comp) --variant=\"{}\" render-dsns-g --format=svg \"_designs-graph:{}.svg\"
    "

if [[ $? != 0 ]]; then
  echo "A variant couldn't be generated; you can find error messages above by searching for [ERRO]"
fi
