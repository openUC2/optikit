#!/bin/bash -eu

script_dir="$(dirname "$(realpath "$BASH_SOURCE")")"
tool_dir="$(dirname "$script_dir")/tools/gltf-checker"

find "$script_dir" -type f | grep -E "\.gltf$|\.glb$" | while read -r file; do
  echo "$file"
  npm --prefix="$tool_dir" -s run start <"$file"
done
