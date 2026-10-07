#!/usr/bin/env python3
"""Build two path-independent offline simulator binaries and record provenance."""
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys

ROOT = Path(__file__).resolve().parent.parent
ARTIFACTS = ROOT / "sims" / "artifacts" / "build"
PINNED_GO = "go1.27.1"


def run(argv, cwd=ROOT, env=None):
    return subprocess.check_output(argv, cwd=cwd, env=env, stderr=subprocess.STDOUT)


def sha(data):
    return hashlib.sha256(data).hexdigest()


def main():
    os.umask(0o077)
    if len(sys.argv) != 2:
        raise ValueError("usage: python3 xops/sim_build.py BUILD_ID")
    build_id = sys.argv[1]
    if not build_id or any(c not in "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_" for c in build_id):
        raise ValueError("invalid build ID")
    destination = ARTIFACTS / build_id
    if any(p.is_symlink() for p in (destination, *destination.parents)):
        raise ValueError("symlink build path")
    destination.mkdir(parents=True, exist_ok=False)
    version = run(["go", "version"]).decode().strip()
    if version.split()[2] != PINNED_GO:
        raise ValueError("compiler does not match pinned Go 1.27.1")
    revision = run(["git", "rev-parse", "HEAD"]).decode().strip()
    patch = run(["git", "diff", "--binary", "HEAD", "--"])
    untracked = sorted(run(["git", "ls-files", "--others", "--exclude-standard", "-z"]).split(b"\0"))
    dirty = hashlib.sha256(patch)
    captured_untracked = {}
    for name in untracked:
        if not name:
            continue
        path = ROOT / os.fsdecode(name)
        if path.is_symlink() or not path.is_file():
            raise ValueError("unsupported untracked source entry")
        content = path.read_bytes()
        captured_untracked[os.fsdecode(name)] = content
        dirty.update(len(name).to_bytes(8, "big") + name)
        dirty.update(len(content).to_bytes(8, "big") + content)
    dirty_hash = dirty.hexdigest()
    names = sorted(p for p in (ROOT / "backend").rglob("*") if p.is_file())
    source = {}
    for path in names:
        if path.is_symlink():
            raise ValueError("symlink source")
        source[str(path.relative_to(ROOT / "backend"))] = path.read_bytes()
        relative_repo = str(path.relative_to(ROOT))
        if relative_repo in captured_untracked and captured_untracked[relative_repo] != source[str(path.relative_to(ROOT / "backend"))]:
            raise ValueError("untracked source changed during provenance capture")
    source_hash = sha(json.dumps({name: sha(data) for name, data in source.items()}, sort_keys=True, separators=(",", ":")).encode())
    if run(["git", "diff", "--binary", "HEAD", "--"]) != patch:
        raise ValueError("source changed during provenance capture; run again with a fresh build ID")
    goroot = Path(run(["go", "env", "GOROOT"]).decode().strip())
    go_executable = Path(shutil.which("go")).resolve()
    compiler = goroot / "pkg" / "tool" / "linux_amd64" / "compile"
    module_path = "github.com/metaphy6/cgms/backend/internal/sim"
    ldflags = f"-X {module_path}.SourceRevision={revision} -X {module_path}.DirtyPatchHash={dirty_hash}"
    flags = ["build", "-a", "-trimpath", "-buildvcs=false", "-ldflags", ldflags]
    hashes = []
    for name in ("one", "different-workspace-two"):
        workspace = destination / name / "backend"
        workspace.mkdir(parents=True)
        for relative, data in source.items():
            path = workspace / relative
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_bytes(data)
        env = os.environ.copy()
        env.update(GOTOOLCHAIN="local", GOPROXY="off", GOSUMDB="off", CGO_ENABLED="0", GOOS="linux", GOARCH="amd64", GOCACHE=str(destination / name / "cache"))
        output = destination / name / "sim"
        run(["go", *flags, "-o", str(output), "./cmd/sim"], cwd=workspace, env=env)
        hashes.append(sha(output.read_bytes()))
    if hashes[0] != hashes[1]:
        raise ValueError("different-path clean builds produced different binaries")
    shutil.copyfile(destination / "one" / "sim", destination / "sim")
    (destination / "sim").chmod(0o555)
    (destination / ".dockerignore").write_text("*\n!sim\n")
    provenance = {"schema_version": "cgms-build-v1", "source_revision": revision, "dirty_patch_hash": dirty_hash, "source_snapshot_hash": source_hash, "go_version": version, "go_executable_sha256": sha(go_executable.read_bytes()), "compiler_sha256": sha(compiler.read_bytes()), "module_sha256": sha(source["go.mod"]), "platform": "linux/amd64", "cgo_enabled": False, "goproxy": "off", "gosumdb": "off", "flags": flags, "clean_build_paths": ["one/backend", "different-workspace-two/backend"], "binary_hashes": hashes, "identical": True}
    (destination / "build-provenance.json").write_text(json.dumps(provenance, indent=2) + "\n")
    print("Simulator build verified:", destination.relative_to(ROOT), hashes[0])


if __name__ == "__main__":
    try:
        main()
    except subprocess.CalledProcessError as exc:
        sys.stderr.buffer.write(exc.output)
        print("Simulator build subprocess failed with exit", exc.returncode, file=sys.stderr)
        sys.exit(6)
    except Exception as exc:
        print("Simulator build failed:", type(exc).__name__, file=sys.stderr)
        sys.exit(6)
