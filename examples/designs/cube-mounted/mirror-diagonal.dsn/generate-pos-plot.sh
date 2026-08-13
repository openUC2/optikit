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
      echo \"{}\"
      $(geom) --variant=\"{}\" render-pos-p \"_positions-plot:{}.html\"
    "

if [[ $? != 0 ]]; then
  echo "A variant couldn't be generated; you can find error messages above by searching for [ERRO]"
fi
