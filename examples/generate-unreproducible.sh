#!/bin/bash

generator="$1.sh"

find . -type f -name "$generator" | while read -r script; do
  echo "$script"
  "$script"
done
