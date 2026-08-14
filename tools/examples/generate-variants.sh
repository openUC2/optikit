#!/bin/bash

script="$(realpath "$BASH_SOURCE")"

case "$#" in
4)
  # Run commands by treating cli args as a single directive line
  set -eu
  "$script" <<<"$@"
  exit 0
  ;;
1)
  # Run commands from directives file
  set -eu

  directives="$1"
  "$script" <"$directives"

  exit 0
  ;;
0)
  # Run commands from directives in stdin
  ;;
*)
  >&2 echo "Cannot understand command invocation with $# arguments!"
  exit 1
  ;;
esac

script_dir="$(dirname "$script")"
repo_dir="$(dirname "$(dirname "$script_dir")")"

directives="$(cat)"
variants="$(yq '.variants | keys | .[]' optikit-design.yml)"
args=""
while IFS= read -r variant; do
  while IFS= read -r directive; do # directive: "command prefix type format"
    echo "$variant $directive"
  done <<<"$directives"
done <<<"$variants" |
  go tool rush -k -e \
    "
      echo \"variant {1}: {2} {3} {4} to {5}\"
      go run \"$repo_dir/main.go\" dev dsn {2} --variant={1} {3}-{4} --format={5} _{4}:{1}.{5}
    "

if [[ "$?" != 0 ]]; then
  echo "A variant couldn't be generated; you can find error messages above by searching for [ERRO]"
fi
