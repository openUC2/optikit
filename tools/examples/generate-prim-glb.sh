#!/bin/bash

script_dir=$(dirname "$(realpath "$BASH_SOURCE")")
repo_dir="$(dirname "$(dirname "$script_dir")")"
function py {
  uv run poe "$@"
}

ls *.stp | while read -r prim_step; do
  py convert glb "$prim_step"
done
