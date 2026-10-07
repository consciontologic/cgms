"""Filesystem regressions for bounded campaign storage accounting."""
import importlib.util
import os
import signal
from contextlib import ExitStack, redirect_stdout, contextmanager
import io
from pathlib import Path
import tempfile
import unittest
from unittest import mock

spec = importlib.util.spec_from_file_location(
    'campaign_storage', Path(__file__).with_name('sim_gameplay_evaluate.py'))
campaign = importlib.util.module_from_spec(spec)
spec.loader.exec_module(campaign)


def legacy_storage_bytes(root):
    total = 0
    for path in root.rglob('*'):
        campaign.require(not path.is_symlink(), 'campaign symlink')
        if path.is_file():
            total += path.stat().st_size
    return total


class StorageAccounting(unittest.TestCase):
    def setUp(self):
        campaign.ARTIFACTS.mkdir(parents=True, exist_ok=True)
        self.directory = tempfile.TemporaryDirectory(
            prefix='storage-regression-', dir=campaign.ARTIFACTS)
        self.addCleanup(self.directory.cleanup)
        self.root = Path(self.directory.name)

    def test_empty_and_nested_readable_trees_match_legacy(self):
        self.assertEqual(campaign.storage_bytes(self.root), 0)
        (self.root/'empty').mkdir()
        (self.root/'nested'/'deeper').mkdir(parents=True)
        (self.root/'empty-file').touch()
        (self.root/'.hidden').write_bytes(b'123')
        (self.root/'nested'/'deeper'/'unicode-é').write_bytes(bytes(range(256)))
        expected = 259
        self.assertEqual(legacy_storage_bytes(self.root), expected)
        self.assertEqual(campaign.storage_bytes(self.root), expected)

    def test_hardlinks_count_logical_bytes_for_each_path(self):
        source = self.root/'original'
        source.write_bytes(b'12345')
        os.link(source, self.root/'hardlink')
        self.assertEqual(campaign.storage_bytes(self.root), 10)
        self.assertEqual(campaign.storage_bytes(self.root), legacy_storage_bytes(self.root))

    def test_file_directory_and_dangling_symlinks_are_rejected(self):
        (self.root/'file').write_bytes(b'x')
        (self.root/'directory').mkdir()
        link = self.root/'link'
        for target in ('file', 'directory', 'missing'):
            with self.subTest(target=target):
                link.symlink_to(self.root/target)
                try:
                    with self.assertRaisesRegex(ValueError, 'campaign symlink'):
                        campaign.storage_bytes(self.root)
                    with self.assertRaisesRegex(ValueError, 'campaign symlink'):
                        legacy_storage_bytes(self.root)
                finally:
                    link.unlink()

    def test_nonregular_fifo_is_not_counted_or_opened(self):
        os.mkfifo(self.root/'fifo')
        (self.root/'data').write_bytes(b'abcd')
        self.assertEqual(campaign.storage_bytes(self.root), 4)
        self.assertEqual(campaign.storage_bytes(self.root), legacy_storage_bytes(self.root))

    def test_root_scan_permission_error_propagates(self):
        with mock.patch.object(campaign.os, 'scandir', side_effect=PermissionError('denied')):
            with self.assertRaises(PermissionError):
                campaign.storage_bytes(self.root)

    def test_nested_scan_permission_error_does_not_undercount(self):
        protected = self.root/'unreadable'
        protected.mkdir()
        (protected/'payload').write_bytes(b'important bytes')
        original = os.scandir

        def scan(path):
            if Path(path) == protected:
                raise PermissionError('denied nested tree')
            return original(path)

        with mock.patch.object(campaign.os, 'scandir', side_effect=scan):
            with self.assertRaises(PermissionError):
                campaign.storage_bytes(self.root)

    def test_scan_io_error_propagates(self):
        with mock.patch.object(campaign.os, 'scandir', side_effect=OSError('filesystem I/O failure')):
            with self.assertRaises(OSError):
                campaign.storage_bytes(self.root)


    def run_scan_failure_campaign(self, stubborn=False, error_type=PermissionError):
        art = self.root/'campaign-artifacts'
        art.mkdir()
        binary = self.root/'binary'
        binary.write_bytes(b'nonexecuted synthetic binary')
        config = self.root/'config'
        config.write_text('{}')
        source = self.root/'source'
        source.write_text('{}')
        seats = ['p0', 'p1', 'p2']
        profile = {'population': 3, 'block_ids': ['fixture'],
                   'rotations': [{'id': 'r'+str(i), 'seats': seats[i:]+seats[:i]} for i in range(3)],
                   'games_per_match': 1, 'budgets': {'workers': 1}, 'root_seed': 'fixture'}
        profile_path = self.root/'profile.json'
        profile_path.write_bytes(campaign.encoded(profile))
        plan = {'schema': campaign.SCHEMA, 'config': str(config), 'profiles': [str(profile_path)],
                'source_manifest': str(source), 'shard_blocks': 1, 'wall_seconds': 100000,
                'storage_bytes': 10000000, 'worker_check_units': [], 'worker_check_workers': 1,
                'analysis_seed': 'fixture', 'resamples': 100}
        plan_path = self.root/'plan.json'
        plan_path.write_bytes(campaign.encoded(plan))
        signals = []
        kills = []

        class Child:
            pid = 12345
            returncode = None
            polls = 0

            def poll(self):
                self.polls += 1
                if self.polls > 8:
                    raise AssertionError('unbounded child wait')
                return self.returncode

            def send_signal(self, sig):
                signals.append(sig)
                if not stubborn:
                    self.returncode = 130

        child = Child()
        def kill(pid, sig):
            kills.append((pid, sig))
            child.returncode = -signal.SIGKILL

        calls = 0
        def storage(_):
            nonlocal calls
            calls += 1
            if calls == 1:
                return 0  # Admit the stage, then fail while its child is alive.
            raise error_type('storage scan denied')

        tick = 0
        def clock():
            nonlocal tick
            tick += 31
            return tick

        argv = ['campaign', str(binary), str(plan_path), 'scan-failure']
        with ExitStack() as stack:
            for patch in (
                mock.patch.object(campaign, 'ARTIFACTS', art),
                mock.patch.object(campaign.sys, 'argv', argv),
                mock.patch.object(campaign, 'storage_bytes', side_effect=storage),
                mock.patch.object(campaign.subprocess, 'Popen', return_value=child),
                mock.patch.object(campaign.time, 'sleep'),
                mock.patch.object(campaign.time, 'monotonic', side_effect=clock),
                mock.patch.object(campaign.os, 'killpg', side_effect=kill),
                mock.patch.object(campaign, 'write_analysis'),
            ):
                stack.enter_context(patch)
            stack.enter_context(redirect_stdout(io.StringIO()))
            self.assertEqual(campaign.main(), 5)
        state = campaign.load(art/'scan-failure'/'state.json')
        self.assertEqual(signals, [signal.SIGINT])
        self.assertIsNotNone(child.returncode)
        self.assertFalse(state['invocation_active'])
        self.assertIsNone(state['storage_bytes'])
        self.assertEqual(state['storage_error'], 'unavailable')
        self.assertEqual(state['status'], 'incomplete')
        self.assertEqual(state['units'][0]['stages']['run']['status'], 'incomplete')
        if stubborn:
            self.assertEqual(kills, [(child.pid, signal.SIGKILL)])
        else:
            self.assertEqual(kills, [])
            self.assertEqual(state['units'][0]['stages']['run']['reason'], 'storage-scan-failed')

    def test_live_child_is_stopped_and_scan_failure_journaled(self):
        self.run_scan_failure_campaign()

    def test_scan_failure_stubborn_child_has_bounded_termination(self):
        self.run_scan_failure_campaign(stubborn=True)



    def test_atomic_rename_rescans_once_and_discards_partial_total(self):
        (self.root/'stable').write_bytes(b'12345')
        partial = self.root/'partial'
        partial.write_bytes(b'abcdefghij')
        original = os.scandir
        calls = 0

        @contextmanager
        def racing_scan(path):
            nonlocal calls
            calls += 1
            with original(path) as entries:
                snapshot = list(entries)
                if calls == 1:
                    # Reproduce a completed atomic publication after enumeration
                    # but before metadata lookup for its old path.
                    snapshot.sort(key=lambda e: e.name != 'stable')
                    partial.rename(self.root/'published')
                yield iter(snapshot)

        with mock.patch.object(campaign.os, 'scandir', side_effect=racing_scan):
            self.assertEqual(campaign.storage_bytes(self.root), 15)
        self.assertEqual(calls, 2)

    def test_persistent_missing_entry_propagates_after_two_scans(self):
        with mock.patch.object(campaign.os, 'scandir', side_effect=FileNotFoundError('publication churn')) as scan:
            with self.assertRaises(FileNotFoundError):
                campaign.storage_bytes(self.root)
        self.assertEqual(scan.call_count, 2)

    def test_persistent_missing_entry_stops_child_and_journals_unknown_storage(self):
        self.run_scan_failure_campaign(error_type=FileNotFoundError)


if __name__ == '__main__':
    unittest.main()
