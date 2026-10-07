"""Failure evidence must survive the probe runner's captured subprocesses."""
import contextlib
import io
import subprocess
import unittest
from unittest.mock import patch

import offline_probe


class CaptureTests(unittest.TestCase):
    def test_cardplay_report_requires_both_complete_populations(self):
        valid = {'mode': 'verify', 'parity': 'passed', 'reports': [
            {'population': n, 'games': 3, 'frames': 20, 'restarts': 19, 'parity': 'passed'}
            for n in (3, 4)]}
        offline_probe.validate_cardplay_report(valid)
        valid['reports'][1]['games'] = 0
        with self.assertRaises(ValueError):
            offline_probe.validate_cardplay_report(valid)

    def test_success_returns_captured_output(self):
        result = subprocess.CompletedProcess(['probe'], 0, 'ok\n')
        with patch('offline_probe.subprocess.run', return_value=result), contextlib.redirect_stdout(io.StringIO()):
            self.assertEqual(offline_probe.run(['probe']), 'ok')

    def test_nonzero_preserves_output_and_fails(self):
        result = subprocess.CompletedProcess(['probe'], 7, 'diagnostic\n')
        output = io.StringIO()
        with patch('offline_probe.subprocess.run', return_value=result), contextlib.redirect_stdout(output):
            with self.assertRaisesRegex(RuntimeError, '7'):
                offline_probe.run(['probe'])
        self.assertIn('diagnostic', output.getvalue())

    def test_timeout_preserves_partial_output_and_fails(self):
        error = subprocess.TimeoutExpired(['probe'], 1, output=b'partial diagnostic\n')
        output = io.StringIO()
        with patch('offline_probe.subprocess.run', side_effect=error), contextlib.redirect_stdout(output):
            with self.assertRaises(subprocess.TimeoutExpired):
                offline_probe.run(['probe'], timeout=1)
        self.assertIn('partial diagnostic', output.getvalue())


if __name__ == '__main__':
    unittest.main()
