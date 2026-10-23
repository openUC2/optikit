#!/bin/bash -eu

script_dir=$(dirname "$(realpath "$BASH_SOURCE")")
main="$(realpath "$script_dir/../../../../../main.go")"
comp() { go run "$main" dev dsn comp "$@"; }
geom() { go run "$main" dev dsn geom "$@"; }

pwd

geom report-assm --format=yaml _assembly.yml
