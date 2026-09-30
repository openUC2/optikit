import datetime
import json
import os
import pathlib
import sys
import tempfile
from types import SimpleNamespace

import build123d as b

axes = {
    "x": (1, 0, 0),
    "y": (0, 1, 0),
    "z": (0, 0, 1),
}


def assemble_prims(prims_report: list[SimpleNamespace], wd: str) -> b.Compound:
    compounds: list[b.Compound] = []
    for prim in prims_report.components:
        if not hasattr(prim, "static_models"):
            continue
        if not hasattr(prim.static_models, "step"):
            continue
        if not hasattr(prim, "rotation"):
            continue

        compound = b.import_step(os.path.join(wd, prim.static_models.step))
        ordering = prim.rotation.order.lower()
        match prim.rotation.kind:
            case "intrinsic":
                pass
            case "extrinsic":
                # build123d seems to compose rotations with intrinsic axes:
                ordering = ordering[::-1]
            case _:
                raise ValueError(f"unknown rotation kind {prim.rotation.kind}")
        angles = {"x": 0, "y": 0, "z": 0}
        if hasattr(prim.rotation, "angles"):
            if hasattr(prim.rotation.angles, "x"):
                angles["x"] = prim.rotation.angles.x
            if hasattr(prim.rotation.angles, "y"):
                angles["y"] = prim.rotation.angles.y
            if hasattr(prim.rotation.angles, "z"):
                angles["z"] = prim.rotation.angles.z
        position = (0, 0, 0)
        if hasattr(prim, "position"):
            position = tuple(prim.position)
        compound.location *= b.Location(position)
        for axis in ordering[::-1]:
            compound.location *= b.Location((0, 0, 0), axes[axis], angles[axis])
        compounds.append(compound)
    return b.Compound(label="design", children=compounds)


def export_assembly(design_assembly: b.Compound, output_path: str | None):
    if output_path is None:
        with tempfile.NamedTemporaryFile(suffix=".step") as fp:
            success = b.export_step(
                design_assembly,
                fp.name,
                timestamp=datetime.datetime.fromtimestamp(0, datetime.UTC),
            )
            if not success:
                raise RuntimeError("Unknown error in STEP export process")
            print(fp.read().decode("utf-8"))
            return

    success = b.export_step(
        design_assembly,
        output_path,
        timestamp=datetime.datetime.fromtimestamp(0, datetime.UTC),
    )
    if not success:
        raise RuntimeError("Unknown error in STEP export process")


def clean_hyphens(d: dict[str, object]):
    dehyphenated = {k.replace("-", "_"): v for k, v in d.items()}
    return SimpleNamespace(**dehyphenated)


def read_json_report(report_text: str) -> list[SimpleNamespace]:
    return json.loads(report_text, object_hook=clean_hyphens)


def assemble_file(input_path: str, output_path: str, wd: str):
    p = pathlib.Path(os.path.join(wd, input_path))
    report_text = p.read_text()
    assembled = assemble_prims(read_json_report(report_text), wd)
    export_assembly(assembled, output_path)


def main_file():
    (input_path, output_path) = sys.argv[1:3]
    wd = os.getcwd()
    poe_wd = os.getenv("POE_PWD")
    if poe_wd is not None and poe_wd != "":
        wd = poe_wd

    assemble_file(input_path, output_path, wd)


def main_batch():
    (input_prefix, output_prefix) = sys.argv[1:3]
    for line in sys.stdin:
        input_path = line.strip()
        if not input_path:
            continue

        wd = os.path.dirname(input_path)
        filename = os.path.basename(input_path)
        (before, _) = os.path.splitext(filename)
        filename = output_prefix + before.removeprefix(input_prefix) + ".step"
        output_path = os.path.join(wd, filename)
        print(output_path)
        assemble_file(input_path, output_path, wd)


def main_json_stdio():
    report_text = "".join([line.rstrip("\r\n") for line in sys.stdin])
    assembled = assemble_prims(read_json_report(report_text), os.getcwd())
    export_assembly(assembled, None)


if __name__ == "__main__":
    main_json_stdio()
