#!/bin/bash

script_dir="$(dirname "$(realpath "$BASH_SOURCE")")"
tool_dir="$(dirname "$script_dir")/tools/gltf-checker"

find "$script_dir" -type f | grep -E "\.gltf$|\.glb$" |
  go tool rush -k -e \
    "echo \"{}\"; npm --prefix=\"$tool_dir\" -s run start <\"{}\""

if [[ $? != 0 ]]; then
  echo "An asset failed validation; you can find error messages above by searching for [ERRO]"
fi
