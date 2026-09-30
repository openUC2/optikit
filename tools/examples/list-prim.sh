#!/bin/bash

script_dir=$(dirname "$(realpath "$BASH_SOURCE")")
repo_dir="$(dirname "$(dirname "$script_dir")")"

ls -d "$PWD"/*.stp
