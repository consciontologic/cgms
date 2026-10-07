#!/usr/bin/env python3
"""Reproduce P01 host FFI parity; optional Android ARM64 compilation, no installs."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import platform
import re
import shutil
import subprocess
import time

ROOT = Path(__file__).resolve().parents[1]


def run(argv, *, cwd=ROOT, env=None, timeout=180):
    print(f"cwd: {cwd}\nrun: {' '.join(map(str, argv))}", flush=True)
    try:
        completed = subprocess.run(argv, cwd=cwd, env=env, text=True,
                                   stdout=subprocess.PIPE, stderr=subprocess.STDOUT,
                                   timeout=timeout)
    except subprocess.TimeoutExpired as error:
        output = error.stdout or b''
        if isinstance(output, bytes):
            output = output.decode('utf-8', errors='replace')
        print(output, flush=True)
        raise
    if completed.returncode:
        print(completed.stdout, flush=True)
        raise RuntimeError(f"command exited {completed.returncode}")
    return completed.stdout.strip()


def validate_cardplay_report(report):
    """Reject incomplete differential evidence before publishing a passing report."""
    rows = report.get('reports', [])
    if (report.get('mode') != 'verify' or report.get('parity') != 'passed'
            or len(rows) != 2 or {r.get('population') for r in rows} != {3, 4}
            or any(r.get('games') != 3 or r.get('parity') != 'passed'
                   or r.get('frames', 0) < 2
                   or r.get('restarts') != r.get('frames', 0) - 1 for r in rows)):
        raise ValueError('incomplete card-play differential evidence')


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('run_id', help='fresh output name under sims/artifacts/offline-probe')
    parser.add_argument('--android-ndk', type=Path, help='existing NDK directory; compile only')
    args = parser.parse_args()
    if not re.fullmatch(r'[a-z0-9][a-z0-9-]{0,39}', args.run_id):
        parser.error('run_id must be 1..40 lowercase letters/digits/hyphens')
    if platform.system() != 'Linux':
        parser.error('this host runner currently measures Linux only')
    for program in ('go', 'dart', 'cc'):
        if shutil.which(program) is None:
            parser.error(f'missing {program}; install/configure separately')
    compiler = None
    if args.android_ndk:
        compiler = args.android_ndk.resolve() / 'toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android23-clang'
        if not compiler.is_file():
            parser.error('expected Linux x86_64 NDK with Android API 23 ARM64 compiler')
    output = ROOT / 'sims/artifacts/offline-probe' / args.run_id
    output.mkdir(parents=True, exist_ok=False, mode=0o700)
    fixtures = output / 'fixtures'
    fixtures.mkdir(mode=0o700)
    env = dict(os.environ, CGMS_PROBE_FIXTURES=str(fixtures), CGO_ENABLED='1', GOPROXY='off', GOSUMDB='off')
    go_version = run(['go', 'version'])
    dart_version = run(['dart', '--version'])
    print(run(['go', 'test', '-count=1', '-race', './internal/offlineprobe'], cwd=ROOT/'backend', env=env))
    print(run(['go', 'vet', './internal/offlineprobe', './cmd/offlineffi'], cwd=ROOT/'backend', env=env))
    library = output / 'libcgmsprobe.so'
    run(['go', 'build', '-trimpath', '-buildmode=c-shared', '-o', str(library), './cmd/offlineffi'], cwd=ROOT/'backend', env=env)
    print(run(['dart', 'analyze', 'client/offline_probe']))
    measurements = []
    finance_measurements = []
    for _ in range(2):
        raw = run(['dart', 'run', 'client/offline_probe/probe_test.dart', str(library), str(fixtures)])
        measurements.append(json.loads(raw))
        raw = run(['dart', 'run', 'client/offline_probe/finance_test.dart', str(library), str(fixtures)])
        finance_measurements.append(json.loads(raw))
    cardplay_runs = []
    cardplay_measurements = []
    for _ in range(2):
        verified = json.loads(run(['dart', 'run', 'client/offline_probe/cardplay_test.dart',
                                   str(fixtures), str(library), 'verify']))
        validate_cardplay_report(verified)
        cardplay_runs.append(verified)
        for mode in ('dart', 'native'):
            cardplay_measurements.append(json.loads(run(
                ['dart', 'run', 'client/offline_probe/cardplay_test.dart',
                 str(fixtures), str(library), mode])))
    android = {'status': 'not_requested'}
    if compiler:
        android_lib = output / 'android-arm64' / 'libcgmsprobe.so'
        android_lib.parent.mkdir()
        android_env = dict(env, GOOS='android', GOARCH='arm64', CC=str(compiler))
        started = time.monotonic()
        run(['go', 'build', '-trimpath', '-buildmode=c-shared', '-o', str(android_lib), './cmd/offlineffi'], cwd=ROOT/'backend', env=android_env)
        android = {'status':'compiled_not_executed', 'api':23, 'abi':'arm64-v8a',
                   'build_seconds':round(time.monotonic()-started, 3),
                   'library_bytes':android_lib.stat().st_size,
                   'ndk':(args.android_ndk/'source.properties').read_text()}
    artifacts = {str(p.relative_to(output)): hashlib.sha256(p.read_bytes()).hexdigest()
                 for p in sorted(output.rglob('*')) if p.is_file()}
    sources = {}
    for path in ['backend/internal/offlineprobe', 'backend/cmd/offlineffi',
                 'backend/internal/game', 'backend/internal/bots',
                 'backend/internal/canonical', 'backend/internal/randomstream',
                 'client/offline_probe']:
        for p in sorted((ROOT/path).glob('*')):
            if p.is_file():
                sources[str(p.relative_to(ROOT))] = hashlib.sha256(p.read_bytes()).hexdigest()
    sources['xops/offline_probe.py'] = hashlib.sha256(Path(__file__).read_bytes()).hexdigest()
    for path in ['backend/go.mod', 'backend/go.sum']:
        sources[path] = hashlib.sha256((ROOT/path).read_bytes()).hexdigest()
    result = {'schema':'cgms-offline-probe-run-v1', 'go':go_version, 'dart':dart_version,
              'host':platform.platform(), 'machine':platform.machine(),
              'host_library_bytes':library.stat().st_size,
              'runs':measurements, 'finance_runs':finance_measurements,
              'cardplay_runs':cardplay_runs, 'cardplay_measurements':cardplay_measurements, 'android':android, 'artifacts_sha256':artifacts,
              'sources_sha256':sources,
              'limitations':['Independent Dart board candidate covers only the documented passive opening/pass policy; not complete rules or competent bots', 'No device runtime or lifecycle evidence',
                             'RSS samples are not a leak proof', 'Six distinct card-play games are fixed fixture verification, not bot evaluation samples',
                             'Whole-match timing includes complete state/event/observation hashing; decision timing includes JSON and native FFI overhead',
                             'Dart-only and native measurements run in separate fresh processes; RSS includes runtimes and is not engine-only memory']}
    (output/'report.json').write_text(json.dumps(result, indent=2)+'\n')
    print(json.dumps({'parity':'passed', 'fresh_processes':2, 'android':android['status'],
                      'report':str(output/'report.json')}))


if __name__ == '__main__':
    main()
