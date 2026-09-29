#!/bin/bash

script_dir=$(dirname "$(realpath "$BASH_SOURCE")")
repo_dir="$(dirname "$(dirname "$script_dir")")"
function geom {
  go run "$repo_dir/main.go" dev dsn geom "$@"
}

geom render-obj --format=step "_objects.step"
