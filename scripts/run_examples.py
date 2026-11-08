"""Reproduce documented cases. Offline by default; --live makes billable requests."""
import argparse
from datetime import datetime, timezone
import hashlib
import json
from pathlib import Path
import subprocess
import tempfile
import time

ROOT = Path(__file__).resolve().parents[1]

def run(command):
    result = subprocess.run([str(x) for x in command], cwd=ROOT, capture_output=True, encoding="utf-8")
    if result.returncode:
        raise RuntimeError(result.stderr.strip() or result.stdout.strip() or "CLI failed")
    return json.loads(result.stdout)

def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--cli", required=True, type=Path, help="codex-draw executable")
    parser.add_argument("--image2-cli", type=Path, help="optional legacy execution backend; recorded in each result")
    parser.add_argument("--case", action="append", dest="cases")
    parser.add_argument("--live", action="store_true", help="send one potentially billable request per case; no retry")
    args = parser.parse_args()
    cases = json.loads((ROOT / "examples/cases.json").read_text(encoding="utf-8"))
    known = {c["id"] for c in cases}
    if args.cases and not set(args.cases) <= known:
        parser.error("unknown case")
    results = ROOT / "examples/results"
    scratch = ROOT / "output/draw/example-tests"
    results.mkdir(parents=True, exist_ok=True)
    scratch.mkdir(parents=True, exist_ok=True)
    for case in cases:
        name = case["id"]
        if args.cases and name not in args.cases:
            continue
        brief = ROOT / "examples/briefs" / (name + ".txt")
        output = results / (name + "." + case["format"])
        common = (["--scenario", case["scenario"], "--brief-file", brief] if case["scenario"] else ["--prompt-file", brief]) + ["--model", case["model"], "--size", "1024x1024", "--quality", "auto", "--out", output, "--timeout", "720", "--max-attempts", "1"]
        reference = results / case["reference"] if case.get("reference") else None
        if reference:
            if not reference.is_file():
                raise RuntimeError("Generate reference first: " + case["reference"])
            common += ["--image", reference]
        if case["format"] == "gif":
            common += ["--allow-experimental", "--frame-count", str(case["frames"]), "--frame-duration-ms", str(case["frame_duration_ms"])]
        plan_command = [args.cli, "plan", *common] if case["scenario"] else [args.cli, "generate", "--dry-run", *common]
        if args.live:
            plan = run(plan_command)
        else:
            # Dry-run still validates overwrite rules; use a temporary output target.
            with tempfile.TemporaryDirectory(prefix="plan-", dir=scratch) as temporary:
                plan_command[plan_command.index("--out") + 1] = Path(temporary) / output.name
                plan = run(plan_command)
        if not case["scenario"]:
            plan["prompt"] = plan["payload"]["prompt"]
            plan["model"] = plan["payload"]["model"]
        prompt_file = scratch / (name + ".txt")
        prompt_file.write_text(plan["prompt"], encoding="utf-8")
        if not args.live:
            print(json.dumps({"case": name, "offline_plan": "passed", "model_candidate": plan["model"]}, ensure_ascii=False), flush=True)
            continue
        if output.exists():
            raise RuntimeError("Refusing to overwrite " + str(output))
        if args.image2_cli:
            command = [args.image2_cli, "edit" if reference else "generate", "--prompt-file", prompt_file, "--model", case["model"], "--size", "1024x1024", "--quality", "auto", "--out", output, "--timeout", "720", "--max-attempts", "1"]
            if reference:
                command += ["--image", reference]
            if case["format"] == "gif":
                command += ["--output-format", "gif", "--frame-count", str(case["frames"]), "--frame-duration-ms", str(case["frame_duration_ms"])]
            backend = "codex-image2 (legacy configured backend)"
        else:
            command = [args.cli, "draw" if case["scenario"] else "generate", *common]
            backend = "codex-draw"
        started = time.monotonic()
        print(json.dumps({"case": name, "status": "requesting", "backend": backend}, ensure_ascii=False), flush=True)
        generated = run(command)
        actual = run([args.cli, "inspect", output])
        record = {"case": name, "scenario": case["scenario"], "generated_at_utc": datetime.now(timezone.utc).isoformat(), "generation_backend": backend, "planner_and_inspector": "codex-draw", "model": generated.get("model", case["model"]), "requested": {"size": "1024x1024", "format": case["format"], "quality": "auto", "max_attempts": 1}, "actual": actual, "elapsed_seconds": round(time.monotonic()-started, 1), "sha256": hashlib.sha256(output.read_bytes()).hexdigest(), "prompt": plan["prompt"], "visual_review": "pending"}
        if reference:
            record["reference"] = case["reference"]
        if case["format"] == "gif":
            record["requested"].update(frame_count=case["frames"], frame_duration_ms=case["frame_duration_ms"])
        (results / (name + ".json")).write_text(json.dumps(record, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
        print(json.dumps({"case": name, "status": "generated", "model": record["model"], "actual": actual}, ensure_ascii=False), flush=True)

if __name__ == "__main__":
    main()
