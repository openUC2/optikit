import os
import pathlib
import sys
from collections.abc import Mapping, Sequence
from types import SimpleNamespace
from typing import Any

import deepmerge
import yaml


def load_yaml(input_path: str, wd: str):
    p = pathlib.Path(os.path.join(wd, input_path))
    report_text = p.read_text()
    return yaml.load(report_text, Loader=yaml.CLoader)


def assemble_devices(
    components_report: Sequence[Mapping[str, Mapping[str, Any]]], wd: str
) -> tuple[str, Mapping[str, Mapping[str, Any]]]:
    version = ""
    devices: Mapping[str, Mapping[str, object]] = {}
    for component in components_report:
        if "results" not in component:
            continue

        results = component["results"]
        if "newswitch-device-id" not in results:
            continue

        id = results["newswitch-device-id"]

        config: dict[str, Any] = {}
        if (
            "static-models" in component
            and "newswitch-device" in component["static-models"]
        ):
            config = load_yaml(component["static-models"]["newswitch-device"], wd)
        if "newswitch-device-config" in results:
            override = results["newswitch-device-config"]
            deepmerge.always_merger.merge(config, override)

        v: str = config["newswitch-version"]
        version = max(version, v)
        config.pop("newswitch-version")

        devices[id] = config

    return (version, devices)


def assemble(
    components_report: Sequence[Mapping[str, object]], wd: str
) -> Mapping[str, object]:
    (version, devices) = assemble_devices(components_report["components"], wd)
    return {
        "newswitch-version": version,
        "devices": devices,
    }


def export_assembly(
    config: Mapping[str, object], output_path: str | None, wd: str | None
):
    serialized = yaml.dump(config, Dumper=yaml.CDumper)
    if output_path is None:
        print(serialized)
        return

    with open(os.path.join(wd, output_path), "w", encoding="utf-8") as f:
        f.write(serialized)


def clean_hyphens(d: Mapping[str, object]):
    dehyphenated = {k.replace("-", "_"): v for k, v in d.items()}
    return SimpleNamespace(**dehyphenated)


def assemble_file(input_path: str, output_path: str, wd: str):
    report: Sequence[Mapping[str, Mapping[str, Any]]] = load_yaml(input_path, wd)
    assembled = assemble(report, wd)
    export_assembly(assembled, output_path, wd)


def main_file():
    (input_path, output_path) = sys.argv[1:3]
    wd = os.getcwd()
    poe_wd = os.getenv("POE_PWD")
    if poe_wd is not None and poe_wd != "":
        wd = poe_wd

    assemble_file(input_path, output_path, wd)


if __name__ == "__main__":
    main_file()
