import os
import os.path
import pathlib
import sys

import build123d as b


def load_model(input_format: str, input_path: str) -> b.Compound:
    match input_format:
        case "step":
            compound = b.import_step(input_path)
        case _:
            raise ValueError(f"unknown input format {input_format}")
    return b.Compound(label=pathlib.Path(input_path).stem, children=[compound]).unwrap()


def export_model(model: b.Compound, output_format: str, output_path: str):
    match output_format:
        case "step":
            b.export_step(model, output_path)
        case "gltf":
            # Undo GLTF export functions's rotation of the model:
            model.location *= b.Location((0, 0, 0), (1, 0, 0), 90)
            b.export_gltf(model, output_path)
        case "glb":
            # Undo GLTF export functions's rotation of the model:
            model.location *= b.Location((0, 0, 0), (1, 0, 0), 90)
            b.export_gltf(model, output_path, binary=True)
        case _:
            raise ValueError(f"unknown output format {output_format}")


def convert_detect(output_format: str, input_path: str):
    (output_path, input_ext) = os.path.splitext(input_path)
    match input_ext:
        case ".stp" | ".step":
            input_format = "step"
        case ".gltf":
            input_format = "gltf"
        case ".glb":
            input_format = "glb"
        case _:
            raise ValueError(f"unknown output format {output_format}")

    output_path += "." + output_format
    poe_wd = os.getenv("POE_PWD")
    if poe_wd is not None and poe_wd != "":
        input_path = os.path.join(poe_wd, input_path)
        output_path = os.path.join(poe_wd, output_path)
    export_model(load_model(input_format, input_path), output_format, output_path)


def main_detect():
    (output_format, input_path) = sys.argv[1:3]
    convert_detect(output_format, input_path)


def main_batch():
    output_format = sys.argv[1]
    for line in sys.stdin:
        input_path = line.strip()
        if not input_path:
            continue

        print(input_path)
        convert_detect(output_format, input_path)


def main_args():
    (input_format, input_path, output_format, output_path) = sys.argv[1:5]
    export_model(load_model(input_format, input_path), output_format, output_path)


if __name__ == "__main__":
    main_detect()
