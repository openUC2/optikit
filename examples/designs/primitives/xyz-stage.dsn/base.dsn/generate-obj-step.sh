#!/bin/bash

main="../../../../../main.go"

script_dir=$(dirname "$(realpath "$BASH_SOURCE")")
function geom {
  go run "$main" dev dsn geom "$@"
}

cd "$script_dir"
geom render-obj --format=step "_objects.step"
