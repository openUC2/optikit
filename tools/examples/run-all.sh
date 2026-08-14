#!/bin/bash

run_dir="$1"
generator="$2.sh"
args="${@:3}"

if [ "$run_dir" = "" ]; then
  run_dir="."
fi

find "$run_dir" -name "$generator" | while read -r script; do
  (
    cd "$(dirname "$script")"
    echo "$script"
    "./$generator" $args
  )
done
