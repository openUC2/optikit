#!/bin/bash

run_dir="$1"

if [ "$run_dir" = "" ]; then
  run_dir="."
fi

script_dir="$(dirname "$(realpath "$BASH_SOURCE")")"
tool_dir="$script_dir"

find "$run_dir" -type f | grep -E "\.gltf$|\.glb$" |
  go tool rush -k \
    "
      echo \"{}\"
      npm --prefix=\"$tool_dir\" -s run start <\"{}\"
    "

if [[ $? != 0 ]]; then
  echo "An asset failed validation; you can find error messages above by searching for [ERRO]"
fi
