#!/usr/bin/env python3
"""Run a protected simulator specification without copying private argv into logs."""
import json
import os
from pathlib import Path
import signal
import stat
import subprocess
import sys

ROOT = Path(__file__).resolve().parent.parent
ARTIFACTS = ROOT / "sims" / "artifacts"


def protected(path):
    path = Path(path).absolute()
    if ".." in path.parts or not path.is_relative_to(ARTIFACTS) or any(p.is_symlink() for p in (path, *path.parents)):
        raise ValueError("invalid protected path")
    return path


def main():
    os.umask(0o077)
    child = None
    canceled = False

    def cancel(signum, frame):
        nonlocal canceled
        canceled = True
        if child is not None and child.poll() is None:
            child.send_signal(signal.SIGINT)

    # Install before Popen so SIGINT cannot strand a just-started child.
    signal.signal(signal.SIGINT, cancel)
    signal.signal(signal.SIGTERM, cancel)
    try:
        if len(sys.argv) != 2:
            return 2
        spec_path = protected(sys.argv[1])
        info = spec_path.stat()
        if not stat.S_ISREG(info.st_mode) or info.st_mode & 0o077 or info.st_size > 1048576:
            return 2
        def unique(pairs):
            result = {}
            for key, value in pairs:
                if key in result:
                    raise ValueError("duplicate key")
                result[key] = value
            return result
        spec = json.loads(spec_path.read_text(), object_pairs_hook=unique)
        if set(spec) != {"argv", "cwd", "capture"} or not isinstance(spec["argv"], list) or not spec["argv"] or not all(isinstance(x, str) for x in spec["argv"]) or not isinstance(spec["cwd"], str) or not isinstance(spec["capture"], str):
            return 2
        capture = protected(spec["capture"])
        with open(str(capture) + ".argv.json", "x") as saved:
            json.dump(spec, saved, ensure_ascii=False)
        with open(str(capture) + ".log", "xb") as log:
            child = subprocess.Popen(spec["argv"], cwd=spec["cwd"], stdout=log, stderr=subprocess.STDOUT)
            if canceled and child.poll() is None:
                child.send_signal(signal.SIGINT)
            code = child.wait()
        if code < 0:
            code = 130
        print("🔎 Simulator status:", code, "(details retained in restricted artifacts)")
        return code
    except (ValueError, TypeError):
        print("❌ Invalid protected simulator specification; private details withheld", file=sys.stderr)
        return 2
    except Exception:
        if child is not None and child.poll() is None:
            child.send_signal(signal.SIGINT)
            child.wait()
        print("❌ Protected simulator invocation failed; private details withheld", file=sys.stderr)
        return 6


if __name__ == "__main__":
    sys.exit(main())
