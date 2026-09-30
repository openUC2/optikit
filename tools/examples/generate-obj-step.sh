#!/bin/bash

script_dir=$(dirname "$(realpath "$BASH_SOURCE")")

uv run poe assemble _primitives.json _objects.step
