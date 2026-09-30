#!/bin/bash

novariant="$(ls -d "$PWD"/_primitives.json 2>/dev/null)"
variants="$(ls -d "$PWD"/_primitives:*.json 2>/dev/null)"
if [[ -n "$novariant" ]]; then
  cat <<<"$novariant"
fi
if [[ -n "$variants" ]]; then
  cat <<<"$variants"
fi
