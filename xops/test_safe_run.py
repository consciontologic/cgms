"""Isolated wrapper regressions; no project tracking breadcrumbs are overwritten."""
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest


class SafeRunTests(unittest.TestCase):
    def run_wrapper(self, command_status, capture_status):
        with tempfile.TemporaryDirectory(prefix='safe-run-test-', dir='/tmp/agent-runs') as td:
            root = Path(td)
            (root / 'xops/agent').mkdir(parents=True)
            (root / 'xops/lib').mkdir()
            shutil.copy(Path(__file__).parent / 'agent/safe-run.sh', root / 'xops/agent/safe-run.sh')
            shutil.copy(Path(__file__).parent / 'lib/log.sh', root / 'xops/lib/log.sh')
            (root / 'bin').mkdir()
            tee = root / 'bin/tee'
            tee.write_text('#!/bin/sh\ncat > "$1"\nexit ' + str(capture_status) + '\n')
            tee.chmod(0o700)
            env = dict(os.environ, PATH=str(root / 'bin') + os.pathsep + os.environ['PATH'], AVB_RUN_DIR=str(root / 'runs'))
            result = subprocess.run(['bash', str(root / 'xops/agent/safe-run.sh'), 'fixture', '--', 'sh', '-c', 'echo captured; exit ' + str(command_status)], cwd=root, env=env, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL, timeout=10)
            status = int(next((root / 'runs').glob('*.exit')).read_text())
            breadcrumb = root / 'docs/tracking/state/last_failure.json'
            failure = json.loads(breadcrumb.read_text()) if breadcrumb.exists() else None
            return result.returncode, status, failure

    def test_capture_failure_never_reports_success(self):
        code, status, failure = self.run_wrapper(0, 73)
        self.assertEqual((code, status), (73, 73))
        self.assertEqual(failure['command_exit_code'], 0)
        self.assertEqual(failure['capture_exit_code'], 73)
        self.assertFalse(failure['resolved'])

    def test_original_command_failure_is_preserved(self):
        code, status, failure = self.run_wrapper(7, 73)
        self.assertEqual((code, status), (7, 7))
        self.assertEqual(failure['command_exit_code'], 7)
        self.assertEqual(failure['capture_exit_code'], 73)

    def test_success_requires_successful_capture(self):
        self.assertEqual(self.run_wrapper(0, 0), (0, 0, None))
