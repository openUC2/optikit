#!/bin/bash

script_dir=$(dirname "$(realpath "$BASH_SOURCE")")

uv run poe assemble _assembly.json _objects.step
