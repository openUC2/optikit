from types import SimpleNamespace

import json
import tempfile
import sys
import build123d as b


axes = {
    "x": (1, 0, 0),
    "y": (0, 1, 0),
    "z": (0, 0, 1),
}


def assemble_prims(prims_report: list[SimpleNamespace]) -> b.Compound:
    compounds: list[b.Compound] = []
    for prim in prims_report:
        compound = b.import_step(prim.model)
        ordering = prim.rotation.order.lower()
        match prim.rotation.type:
            case "intrinsic":
                pass
            case "extrinsic":
                # build123d seems to compose rotations with intrinsic axes:
                ordering = ordering[::-1]
            case _:
                raise ValueError(f"unknown rotation type {prim.rotation.type}")
        angles = {
            "x": prim.rotation.angles.x,
            "y": prim.rotation.angles.y,
            "z": prim.rotation.angles.z,
        }
        for axis in ordering[::-1]:
            compound.location *= b.Location((0, 0, 0), axes[axis], angles[axis])
        compound.location.position = prim.position
        compounds.append(compound)
    return b.Compound(label="design", children=compounds)


def export_assembly(design_assembly: b.Compound):
    with tempfile.NamedTemporaryFile(suffix=".step") as fp:
        b.export_step(design_assembly, fp.name)
        print(fp.read().decode("utf-8"))


prims_report: list[SimpleNamespace] = json.loads(
    "".join([line.rstrip("\r\n") for line in sys.stdin]),
    object_hook=SimpleNamespace,
)
export_assembly(assemble_prims(prims_report))
