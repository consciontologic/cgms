"""Infrastructure tests never require or silently skip Docker."""
import importlib.util
from pathlib import Path
import unittest
from unittest.mock import patch, Mock
spec = importlib.util.spec_from_file_location('postgres_runner', Path(__file__).with_name('postgres_integration.py'))
runner = importlib.util.module_from_spec(spec)
spec.loader.exec_module(runner)

class RunnerTest(unittest.TestCase):
    def test_cleanup_attempts_all_resources_after_failure(self):
        calls = []
        def command(args, **kwargs):
            calls.append(args)
            if args[:2] == ['docker', 'rm']:
                raise RuntimeError('failure')
            return ''
        with patch.object(runner, 'command', side_effect=command):
            self.assertFalse(runner.cleanup('container', 'volume', 'network'))
        self.assertEqual([c[1] for c in calls], ['rm', 'volume', 'network'])

    def test_go_arguments_keep_integration_and_finite_timeout(self):
        argv = runner.test_command(['-race', '-timeout=2m', './internal/storage/...'])
        self.assertIn('-tags=integration', argv)
        self.assertEqual(argv[-3:], ['-race', '-timeout=2m', './internal/storage/...'])
        self.assertIn('-timeout=5m', argv)

    def test_pin_is_digest(self):
        self.assertRegex(runner.IMAGE, r'^postgres@sha256:[a-f0-9]{64}$')

    def test_webfixture_uses_only_dedicated_loopback_command(self):
        argv = runner.test_command(['--webfixture'])
        self.assertEqual(argv[:3], ['go', 'run', './cmd/webfixture'])
        self.assertIn('127.0.0.1:8081', argv)
        self.assertEqual(argv[-1], str(runner.ROOT/'client/build/web'))
        with self.assertRaises(ValueError):
            runner.test_command(['--webfixture', '--listen=0.0.0.0:80'])

    def test_economy_fixture_is_explicit_and_loopback_only(self):
        argv = runner.test_command(['--webfixture-economy'])
        self.assertEqual(argv[:3], ['go', 'run', './cmd/webfixture'])
        self.assertIn('127.0.0.1:8081', argv)
        self.assertIn('-economy', argv)
        self.assertNotIn('-economy', runner.test_command(['--webfixture']))
        with self.assertRaises(ValueError):
            runner.test_command(['--webfixture-economy', '--listen=0.0.0.0:80'])

    def test_test_failure_propagates_and_always_cleans(self):
        child = Mock()
        child.wait.return_value = 7
        child.poll.return_value = 7
        commands = []
        def command(args, **kwargs):
            commands.append(args)
            return '127.0.0.1:54321' if args[:2] == ['docker', 'port'] else ''
        with patch.object(runner, 'command', side_effect=command), patch.object(runner, 'ready'), patch.object(runner.subprocess, 'Popen', return_value=child) as start:
            self.assertEqual(runner.main(['-race', './...']), 7)
        self.assertEqual([a[1] for a in commands[-3:]], ['rm', 'volume', 'network'])
        env = start.call_args.kwargs['env']
        self.assertTrue(env['CGMS_TEST_DATABASE_URL'].startswith('postgres://cgms:'))
        self.assertNotIn('POSTGRES_PASSWORD', env)
        self.assertIn('-tags=integration', start.call_args.args[0])
        self.assertNotIn(env['CGMS_TEST_DATABASE_URL'], str(start.call_args.args[0]))
        run = next(c for c in commands if c[:2] == ['docker', 'run'])
        self.assertRegex(run[run.index('-p')+1], r'^127\.0\.0\.1:[1-9][0-9]*:5432$')

    def test_only_browser_fixture_gets_longer_bounded_lifetime(self):
        for args, timeout in [(['--webfixture'], 3600), (['--webfixture-economy'], 3600), (['./internal/game'], 900)]:
            with self.subTest(args=args):
                child = Mock()
                child.wait.return_value = child.poll.return_value = 0
                def command(argv, **kwargs):
                    return '127.0.0.1:54321' if argv[:2] == ['docker', 'port'] else ''
                with patch.object(runner, 'command', side_effect=command), patch.object(runner, 'ready'), patch.object(runner.subprocess, 'Popen', return_value=child):
                    self.assertEqual(runner.main(args), 0)
                child.wait.assert_called_once_with(timeout=timeout)

    def test_missing_docker_is_failure_not_skip(self):
        with patch.object(runner, 'command', side_effect=FileNotFoundError):
            self.assertEqual(runner.main([]), 1)

    def test_child_exit_race_does_not_bypass_cleanup(self):
        child = Mock()
        child.pid = 12345
        child.poll.return_value = None
        child.wait.side_effect = [runner.subprocess.TimeoutExpired('go', 900), 0]
        commands = []
        def command(args, **kwargs):
            commands.append(args)
            return '127.0.0.1:54321' if args[:2] == ['docker', 'port'] else ''
        with patch.object(runner, 'command', side_effect=command), patch.object(runner, 'ready'), patch.object(runner.subprocess, 'Popen', return_value=child), patch.object(runner.os, 'killpg', side_effect=ProcessLookupError):
            self.assertEqual(runner.main([]), 1)
        self.assertEqual([a[1] for a in commands[-3:]], ['rm', 'volume', 'network'])

    def test_optional_redis_has_private_url_and_cleanup(self):
        child = Mock()
        child.wait.return_value = child.poll.return_value = 0
        commands = []
        def command(args, **kwargs):
            commands.append(args)
            if args[:2] == ['docker', 'port']:
                return '127.0.0.1:54321'
            if args[:2] == ['docker', 'exec']:
                return 'PONG'
            return ''
        with patch.object(runner, 'command', side_effect=command), patch.object(runner, 'ready'), patch.object(runner.subprocess, 'Popen', return_value=child) as start:
            self.assertEqual(runner.main(['--redis', './...']), 0)
        env = start.call_args.kwargs['env']
        self.assertTrue('CGMS_TEST_REDIS_URL' in env, 'missing private Redis configuration')
        self.assertNotIn('--redis', start.call_args.args[0])
        self.assertTrue(any(c[:3] == ['docker', 'rm', '-f'] and c[-1].endswith('-redis') for c in commands))
