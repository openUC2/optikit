import tempfile
import build123d as b

with b.BuildPart() as box_builder:
    b.Box(1, 1, 1)

with tempfile.NamedTemporaryFile() as fp:
    b.export_step(box_builder.part, fp.name)
    print(fp.read().decode("utf-8"))
