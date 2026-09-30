#!/bin/bash

run_dir="$1"
lister="$2.sh"
process_command="${@:3}"

script_dir=$(dirname "$(realpath "$BASH_SOURCE")")

if [ "$run_dir" = "" ]; then
  run_dir="."
fi

input_files=""
while read -r script; do
  echo "$script"
  new_files="$(cd "$(dirname "$script")" && "./$lister")"
  input_files="$(cat <<<"$input_files" && cat <<<"$new_files")"
done < <(find "$run_dir" -name "$lister")
$process_command <<<"$input_files"
