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
assemblies="$(yq '.assemblies | keys | .[]' optikit-design.yml)"
while IFS= read -r assembly; do
  if [ -f assembly-instantiations.yml ]; then
    inputs="$(yq ".$assembly.inputs" assembly-instantiations.yml)"
    inst_args="$(yq 'to_entries | map(. | "--input=\(.key):\(.value)") | join(" ")' <<<"$inputs")"
    inputs_string="($(yq 'to_entries | map(. | "\(.key)=\(.value)") | join(" ")' <<<"$inputs"))"
  else
    inputs_string="()"
    inst_args=""
  fi
  while IFS= read -r directive; do # directive: "command prefix kind format"
    directive="$(tr ' ' '\t' <<<"$directive")"
    command="$(cut -f1 <<<"$directive")"
    after_command="$(cut -f2- <<<"$directive")"
    if [ "$command" == "geom" ]; then
      assm_arg="--assembly=$assembly"
      assembly_suffix=":$assembly"
    else
      assm_arg=""
      assembly_suffix=""
    fi
    echo -e "$assembly_suffix\t$inputs_string\t$inst_args\t$command\t$assm_arg\t$after_command"
  done <<<"$directives"
done <<<"$assemblies" |
  go tool rush -k -e -d "\t" \
    "
      echo '{4}{1} {6} {7} to {8} with inputs {2}'
      go run '$repo_dir/main.go' dev dsn {4} {5} {3} {6}-{7} --format={8} '_{7}{1}.{8}'
    "

if [[ "$?" != 0 ]]; then
  echo "An assembly couldn't be generated; you can find error messages above by searching for [ERRO]"
fi
