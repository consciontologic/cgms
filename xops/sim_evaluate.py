#!/usr/bin/env python3
"""Bounded local Phase 0 evaluation; private invocation details stay in artifacts."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess
import sys
import time

ROOT = Path(__file__).resolve().parent.parent
ARTIFACTS = ROOT / "sims" / "artifacts"


def require(condition, message):
    if not condition:
        raise ValueError(message)


def inspect_manifest(directory, expected_code):
    directory = Path(directory)
    m = json.loads((directory / "manifest.json").read_text())
    require(m["schema_version"] == "cgms-manifest-v1" and m["privacy"] == "restricted", "manifest format")
    require(m["exit_code"] == expected_code, "unexpected manifest exit")
    require(m["status"] == {0: "complete", 3: "unsupported"}[expected_code], "manifest status")
    seen = set()
    for a in m["artifacts"]:
        rel = Path(a["path"])
        require(not rel.is_absolute() and ".." not in rel.parts and a["path"] not in seen, "artifact path")
        seen.add(a["path"])
        path = directory / rel
        require(not any(p.is_symlink() for p in (path, *path.parents)), "artifact symlink")
        b = path.read_bytes()
        require(len(b) == a["bytes"] and hashlib.sha256(b).hexdigest() == a["sha256"], "artifact integrity")
        if a["privacy"] == "public" or path.suffix == ".md":
            text = b.decode()
            require(not m.get("root_seed") or m["root_seed"] not in text, "private root in public artifact")
            require(re.search(r"deck-[12]-(hearts|spades|clubs|diamonds)-\d+", text) is None, "physical identity in public artifact")
    require("report.md" in seen, "report missing")
    require("Observer: public" in (directory / "report.md").read_text(), "report observer")
    return m


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("binary")
    parser.add_argument("output_id")
    parser.add_argument("--samples", type=int, default=4)
    args = parser.parse_args()
    require(re.fullmatch(r"[a-zA-Z0-9][a-zA-Z0-9_-]{0,79}", args.output_id), "invalid output id")
    require(1 <= args.samples <= 1000, "sample bound")
    binary = Path(args.binary).resolve(strict=True)
    os.umask(0o077)
    dest = ARTIFACTS / args.output_id
    require(not any(p.is_symlink() for p in (dest, *dest.parents)), "output symlink")
    dest.mkdir(parents=True, exist_ok=False)
    started = time.monotonic()
    summary = {"schema": "cgms-evaluation-v1", "status": "running", "samples_per_population": args.samples, "claim_scope": "partial-scenarios-only", "runs": []}
    # Fixed synthetic private root is replay input, never printed or placed in reports.
    root_seed = "871936420195728361"
    def invoke(name, argv, expected):
        output = dest / name
        capture = dest / (name + "-capture")
        spec = {"argv": [str(binary), *argv, "--review-hold", "--out", str(output)], "cwd": str(ROOT), "capture": str(capture)}
        path = dest / (name + "-spec.json")
        path.write_text(json.dumps(spec))
        result = subprocess.run([sys.executable, str(ROOT / "xops/sim_private.py"), str(path)], cwd=ROOT, capture_output=True)
        require(result.returncode == expected, "unexpected process status for " + name)
        m = inspect_manifest(output, expected)
        require(m["review_hold"], "evaluation evidence hold missing")
        if m["command"] not in ("compare", "replay"):
            outcomes = json.loads((output / "outcomes.json").read_text())
            require(len(outcomes) == m["planned"] and m["finalized"] == 0, "incomplete denominators")
            require(all(not o["game_complete"] for o in outcomes), "fabricated completion")
            report = (output / "report.md").read_text()
            require("Finalized games: 0 /" in report and "3-player population" in report and "4-player population" in report, "population report")
            require(m["provenance"]["binary_sha256"] == hashlib.sha256(binary.read_bytes()).hexdigest(), "binary provenance")
        summary["runs"].append({"name": name, "exit_code": expected, "planned": m["planned"], "finalized": m["finalized"], "manifest_sha256": hashlib.sha256((output / "manifest.json").read_bytes()).hexdigest()})
        return output
    try:
        base = ["--config", "config/rules/game-rules.json", "--seed", root_seed]
        def population(pop):
            return ["--players", str(pop), "--bots", ",".join(["legal-random@v1"] * pop)]
        q10 = invoke("q10", ["run", *base, *population(3), "--scenario", "sims/scenarios/q10-reciprocal-thirds.json"], 0)
        invoke("q10-replay", ["replay", "--manifest", str(q10 / "manifest.json")], 0)
        for pop in (3, 4):
            scenario = "sims/scenarios/narrow-combat-bots" + ("-4p" if pop == 4 else "") + ".json"
            common = [*base, *population(pop), "--scenario", scenario]
            run = invoke(f"bots-{pop}p", ["run", *common], 3)
            invoke(f"bots-{pop}p-replay", ["replay", "--manifest", str(run / "manifest.json")], 3)
            one = invoke(f"batch-{pop}p-w1", ["batch", *common, "--games", str(args.samples), "--workers", "1"], 3)
            invoke(f"batch-{pop}p-replay", ["replay", "--manifest", str(one / "manifest.json")], 3)
            four = invoke(f"batch-{pop}p-w4", ["batch", *common, "--games", str(args.samples), "--workers", "4"], 3)
            for p in one.rglob("*"):
                if p.name in ("outcomes.json", "schedule.json", "initial.json", "checkpoint.json", "trace.jsonl"):
                    require(p.read_bytes() == (four / p.relative_to(one)).read_bytes(), "worker-count nondeterminism")
            invoke(f"compare-{pop}p", ["compare", "--baseline", str(one / "manifest.json"), "--candidate", str(four / "manifest.json")], 0)
            profile = json.loads((ROOT / f"sims/experiments/baseline-{pop}p.json").read_text())
            profile["block_ids"] = ["evaluation-000"]
            profile["retention"]["review_hold"] = True
            profile["budgets"]["max_output_bytes"] = 67108864
            profile_path = dest / f"tournament-{pop}p-profile.json"
            profile_path.write_text(json.dumps(profile))
            tournament = invoke(f"tournament-{pop}p", ["tournament", *base, "--scenario", scenario, "--experiment", str(profile_path)], 3)
            require(json.loads((tournament / "manifest.json").read_text())["review_hold"], "review hold missing")
            profile["id"] = f"evaluation-grid-{pop}p"
            profile["grid"] = [{"parameter": "bot.immediate-score", "values": [5, 10]}]
            profile["budgets"]["max_matches"] = 4 * pop
            sweep_path = dest / f"sweep-{pop}p-profile.json"
            sweep_path.write_text(json.dumps(profile))
            invoke(f"sweep-{pop}p", ["sweep", *base, "--scenario", scenario, "--experiment", str(sweep_path)], 3)
        summary["status"] = "passed"
        summary["worker_determinism"] = "exact outcomes, schedules, initial states, checkpoints and traces at workers 1 and 4"
        summary["complete_games"] = 0
        summary["replayed_seeded_partial_prefixes_per_population"] = args.samples
    except Exception as error:
        summary["status"] = "failed"
        # Only our static validation label is exposed; never exception payloads
        # from JSON, OS or subprocess containing private data.
        summary["failure_type"] = type(error).__name__
        raise
    finally:
        summary["elapsed_seconds"] = round(time.monotonic() - started, 3)
        path = dest / "summary.json.partial"
        path.write_text(json.dumps(summary, indent=2) + "\n")
        path.replace(dest / "summary.json")
        print(json.dumps({"evaluation": args.output_id, "status": summary["status"], "checked_runs": len(summary["runs"]), "complete_games": 0, "elapsed_seconds": summary["elapsed_seconds"]}))


if __name__ == "__main__":
    try:
        main()
    except Exception:
        print("Evaluation failed; restricted artifacts retain evidence.", file=sys.stderr)
        sys.exit(1)
