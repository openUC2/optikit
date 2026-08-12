#!/bin/bash

generator="$1.sh"

script_dir="$(dirname "$(realpath "$BASH_SOURCE")")"

find "$script_dir" -type f -name "$generator" | while read -r script; do
  echo "$script"
  "$script"
done
