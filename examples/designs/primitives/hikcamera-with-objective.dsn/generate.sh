#!/bin/bash -eu

py() { uv run poe "$@"; }

pwd

./generate-assm.sh generate-assm.directives
py config _assembly:wiring.yml _newswitch-config.yml
