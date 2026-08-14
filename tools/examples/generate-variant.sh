#!/bin/bash

command="$1"
prefix="$2"
type="$3"
format="$4"

script_dir="$(dirname "$(realpath "$BASH_SOURCE")")"
repo_dir="$(dirname "$(dirname "$script_dir")")"
function optikit-dev-dsn {
  echo "go run \"$repo_dir/main.go\" dev dsn $command"
}

yq '.variants | keys | .[]' optikit-design.yml |
  go tool rush -k -e \
    "
      echo \"$command $prefix $type for variant {} as $format\"
      $(optikit-dev-dsn "$command") --variant=\"{}\" $prefix-$type --format=$format \"_$type:{}.$format\"
    "

if [[ $? != 0 ]]; then
  echo "A variant couldn't be generated; you can find error messages above by searching for [ERRO]"
fi
