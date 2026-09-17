#!/bin/bash

script="$(realpath "$BASH_SOURCE")"
script_dir="$(dirname "$script")"

device_configs="$(
  yq '
    filter(.results.newswitch-device-id) |
    map({
      "id": .results.newswitch-device-id,
      "config-from-file": load(.static-models.newswitch-device),
      "config-override": .results.newswitch-device-config
    }) |
    map({
      "id": .id,
      "config": .config-from-file *+ .config-override
    }) |
    .[] as $item ireduce ({}; .[$item | .id] = ($item | .config) )
  ' <_components.yml
)"
minimum_supported_version="$(
  yq '
    . |
    map(.newswitch-version) |
    max
  ' <<<"$device_configs"
)"
config="$(
  yq "
    {
      \"newswitch-version\": \"$minimum_supported_version\",
      \"devices\": .
    } |
    del(.devices.[].newswitch-version)
  " <<<"$device_configs"
)"
cat >"_newswitch-config.yml" <<<"$config"
