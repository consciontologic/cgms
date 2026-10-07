#!/usr/bin/env python3
"""Generate or verify simulator runtime shapes from the single Go module."""
import os
from pathlib import Path
import subprocess
import sys

root = Path(__file__).resolve().parent.parent
if sys.argv[1:] not in ([], ["--check"]):
    raise SystemExit("usage: python3 xops/sim_schema.py [--check]")
env = os.environ.copy()
env.update(GOTOOLCHAIN="local", GOPROXY="off", GOSUMDB="off")
raise SystemExit(subprocess.call(["go", "run", "./cmd/sim-schema", "--out", str(root / "config/schemas"), *sys.argv[1:]], cwd=root / "backend", env=env))
