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
while IFS= read -r variant; do
  if [ -f variant-instantiations.yml ]; then
    inputs="$(yq ".$variant.inputs" variant-instantiations.yml)"
    inst_args="--variant=$variant $(yq 'to_entries | map(. | "--input=\(.key):\(.value)") | join(" ")' <<<"$inputs")"
    inputs_string="($(yq 'to_entries | map(. | "\(.key)=\(.value)") | join(" ")' <<<"$inputs"))"
  else
    inputs_string="()"
    inst_args="--variant=$variant"
  fi
  while IFS= read -r directive; do # directive: "command prefix kind format"
    directive="$(tr ' ' '\t' <<<"$directive")"
    echo -e "$variant\t$inputs_string\t$inst_args\t$directive"
  done <<<"$directives"
done <<<"$variants" |
  go tool rush -k -e -d "\t" \
    "
      echo '{1}: {4} {5} {6} to {7} with inputs {2}'
      go run '$repo_dir/main.go' dev dsn {4} {3} {5}-{6} --format={7} '_{6}:{1}.{7}'
    "

if [[ "$?" != 0 ]]; then
  echo "A variant couldn't be generated; you can find error messages above by searching for [ERRO]"
fi
