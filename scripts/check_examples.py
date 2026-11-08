"""Check published example assets and their recorded provenance without network access."""
import hashlib
import json
from pathlib import Path
import struct

ROOT = Path(__file__).resolve().parents[1]

def main():
    cases = json.loads((ROOT / "examples/cases.json").read_text(encoding="utf-8"))
    assert len({c["id"] for c in cases}) == len(cases), "Duplicate case IDs"
    for case in cases:
        stem = ROOT / "examples/results" / case["id"]
        image = stem.with_suffix("." + case["format"])
        record = json.loads(stem.with_suffix(".json").read_text(encoding="utf-8"))
        data = image.read_bytes()
        assert record["case"] == case["id"]
        assert record["scenario"] == case["scenario"]
        assert hashlib.sha256(data).hexdigest() == record["sha256"], "Image changed: " + case["id"]
        assert len(data) == record["actual"]["bytes"]
        assert record["model"] and record["model"] != "auto", "Missing resolved model"
        assert record["visual_review"]["status"] in {"pass", "partial", "fail"}, "Pending visual review"
        if case["format"] == "png":
            assert data.startswith(b"\x89PNG\r\n\x1a\n")
            dimensions = struct.unpack(">II", data[16:24])
        else:
            assert data[:6] in (b"GIF87a", b"GIF89a")
            dimensions = struct.unpack("<HH", data[6:10])
            assert record["actual"]["frames"] > 1, "GIF is not animated"
        assert dimensions == (record["actual"]["width"], record["actual"]["height"])
        if case.get("reference"):
            assert (ROOT / "examples/results" / case["reference"]).is_file()
        assert (ROOT / "examples/briefs" / (case["id"] + ".txt")).is_file()
        print(case["id"] + ": asset, hash, dimensions and review record verified")

if __name__ == "__main__":
    main()
