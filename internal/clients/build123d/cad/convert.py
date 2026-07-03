import sys
import build123d as b


def load_model(input_format: str, input_path: str) -> b.Compound:
    match input_format:
        case "step":
            compound = b.import_step(input_path)
        case _:
            raise ValueError(f"unknown input format {input_format}")
    return b.Compound(label="design", children=[compound])


def export_model(model: b.Compound, output_format: str, output_path: str):
    with open(output_path, "w") as fp:
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


(input_format, input_path, output_format, output_path) = sys.argv[1:5]
export_model(load_model(input_format, input_path), output_format, output_path)
