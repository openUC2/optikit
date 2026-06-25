from types import SimpleNamespace

import json
import tempfile
import sys
import build123d as b


def assemble_prims(prims_report: list[SimpleNamespace]) -> b.Compound:
    compounds: list[b.Compound] = []
    for prim in prims_report:
        compound = b.import_step(prim.model)
        match prim.rotation.type:
            case "extrinsic":
                ordering = b.Extrinsic[prim.rotation.order.upper()]
            case "intrinsic":
                ordering = b.Intrinsic[prim.rotation.order.upper()]
            case _:
                raise ValueError(f"unknown rotation type {prim.rotation.type}")
        angles = (
            prim.rotation.angles.x,
            prim.rotation.angles.y,
            prim.rotation.angles.z,
        )
        location = b.Location(
            position=prim.position, orientation=angles, ordering=ordering
        )
        compound.move(location)
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
