"""Maintainer build only; installed users need no Python/Go runtime."""
import hashlib
import json
import os
from pathlib import Path
import subprocess
import zipfile

ROOT = Path(__file__).resolve().parents[1]
VERSION = (ROOT / "VERSION").read_text(encoding="utf-8").strip()
GO = os.environ.get("GO", "go")
PLATFORMS = [("windows", "amd64"), ("windows", "arm64"), ("darwin", "amd64"), ("darwin", "arm64")]

def main():
    skill = ROOT / "codex-draw"
    binary_dir = skill / "bin"
    binary_dir.mkdir(exist_ok=True)
    names = []
    for system, arch in PLATFORMS:
        name = f"codex-draw-{system}-{arch}" + (".exe" if system == "windows" else "")
        env = dict(os.environ, GOOS=system, GOARCH=arch, CGO_ENABLED="0")
        subprocess.run([GO, "build", "-trimpath", "-buildvcs=false", "-ldflags", f"-s -w -X main.version={VERSION}", "-o", str(binary_dir / name), "./codex-draw/src"], cwd=ROOT, env=env, check=True)
        names.append(name)
    out = ROOT / "dist"
    out.mkdir(exist_ok=True)
    archive = out / f"codex-draw-v{VERSION}.zip"
    allowed_dirs = {"agents", "references", "registry"}
    files = [p for p in skill.rglob("*") if p.is_file() and (p.relative_to(skill).parts[0] in allowed_dirs or p.name in {"SKILL.md", "配置图片服务.cmd", "配置图片服务.command"})]
    files += [binary_dir / n for n in names]
    with zipfile.ZipFile(archive, "w", zipfile.ZIP_DEFLATED, compresslevel=9) as z:
        for p in sorted(files):
            rel = p.relative_to(ROOT).as_posix()
            info = zipfile.ZipInfo(rel, (2026, 1, 1, 0, 0, 0))
            info.create_system = 3
            info.compress_type = zipfile.ZIP_DEFLATED
            executable = p.name.endswith(".command") or "darwin-" in p.name
            info.external_attr = (0o100755 if executable else 0o100644) << 16
            z.writestr(info, p.read_bytes())
        for name in ["LICENSE", "NOTICE.md"]:
            info = zipfile.ZipInfo("codex-draw/" + name, (2026, 1, 1, 0, 0, 0))
            info.create_system = 3
            info.external_attr = 0o100644 << 16
            info.compress_type = zipfile.ZIP_DEFLATED
            z.writestr(info, (ROOT / name).read_bytes())
    digest = hashlib.sha256(archive.read_bytes()).hexdigest()
    (out / "SHA256SUMS").write_text(f"{digest}  {archive.name}\n", encoding="ascii")
    with zipfile.ZipFile(archive) as z:
        assert "codex-draw/SKILL.md" in z.namelist()
        assert all("codex-draw/bin/" + name in z.namelist() for name in names)
        assert not any("/src/" in name or "config.json" in name for name in z.namelist())
    print(json.dumps({"archive": str(archive), "sha256": digest, "platforms": names}, indent=2))

if __name__ == "__main__":
    main()
