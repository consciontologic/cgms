import importlib.util
import json
from pathlib import Path
import subprocess
import tempfile
import unittest
from unittest.mock import Mock, call, patch

SPEC = importlib.util.spec_from_file_location('service_ops', Path(__file__).with_name('service_ops.py'))
ops = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(ops)

class ServiceOperations(unittest.TestCase):
    def test_make_lifecycle_commands_dispatch_with_selected_mode(self):
        for target, action in [('up', 'up'), ('down', 'stop'), ('restart', 'up')]:
            for mode in ['local', 'production']:
                with self.subTest(target=target, mode=mode):
                    result = subprocess.run(
                        ['make', '--no-print-directory', '-n', target,
                         'PYTHON=python3', 'MODE='+mode],
                        cwd=ops.ROOT, text=True, capture_output=True,
                    )
                    self.assertEqual(result.returncode, 0, result.stderr)
                    self.assertEqual(result.stdout.strip(),
                                     'python3 xops/makefile/service_ops.py '+action+' --mode='+mode)

    def test_up_prepares_both_builds_before_recreating_configured_stack(self):
        for telemetry in [False, True]:
            with self.subTest(telemetry=telemetry):
                command = ['docker', 'compose'] + (['--profile', 'telemetry'] if telemetry else [])
                events = Mock()
                with patch.object(ops.sys, 'argv', ['service_ops.py', 'up']), \
                     patch.dict(ops.os.environ), \
                     patch.object(ops, 'compose_command', return_value=command), \
                     patch.object(ops, 'initialize', events.initialize), \
                     patch.object(ops, 'build_web', events.build_web), \
                     patch.object(ops, 'execute', events.execute), \
                     patch.object(ops, 'wait_ready', events.wait_ready), \
                     patch.object(ops, 'telemetry_ready', events.telemetry_ready):
                    ops.main()
                events.initialize.assert_called_once_with()
                events.build_web.assert_called_once_with()
                events.execute.assert_any_call(command+['build', 'backend'], timeout=180)
                recreate = call.execute(command+['up', '-d', '--force-recreate'], timeout=180)
                self.assertIn(recreate, events.mock_calls)
                self.assertLess(events.mock_calls.index(call.initialize()), events.mock_calls.index(recreate))
                self.assertLess(events.mock_calls.index(call.build_web()), events.mock_calls.index(recreate))
                services = ['backend', 'web', 'postgres', 'redis'] + (sorted(ops.TELEMETRY) if telemetry else [])
                events.wait_ready.assert_called_once_with(command, services, timeout=90)
                if telemetry:
                    events.telemetry_ready.assert_called_once_with(command)
                else:
                    events.telemetry_ready.assert_not_called()

    def test_failed_web_build_does_not_change_running_services(self):
        with patch.object(ops.sys, 'argv', ['service_ops.py', 'up']), \
             patch.dict(ops.os.environ), \
             patch.object(ops, 'initialize'), \
             patch.object(ops, 'build_web', side_effect=RuntimeError('web build failed')), \
             patch.object(ops, 'execute') as execute, \
             patch.object(ops, 'wait_ready'), \
             patch.object(ops, 'synchronize_telemetry') as synchronize:
            with self.assertRaisesRegex(RuntimeError, 'web build failed'):
                ops.main()
            synchronize.assert_not_called()
            self.assertFalse(any('up' in c.args[0] or 'stop' in c.args[0]
                                 for c in execute.call_args_list))

    def test_stop_retains_all_volumes_without_building(self):
        command = ['docker', 'compose']
        with patch.object(ops.sys, 'argv', ['service_ops.py', 'stop']), \
             patch.dict(ops.os.environ), \
             patch.object(ops, 'compose_command', return_value=command), \
             patch.object(ops, 'initialize') as initialize, \
             patch.object(ops, 'build_web') as build_web, \
             patch.object(ops, 'execute') as execute:
            ops.main()
            execute.assert_called_once_with(command+['stop']+sorted(ops.SERVICES))
            initialize.assert_not_called()
            build_web.assert_not_called()

    def test_setting_drives_profile(self):
        with tempfile.TemporaryDirectory() as temp:
            cfg=Path(temp)/'config.json'
            cfg.write_text(json.dumps({'observability':{'enabled':False}}))
            self.assertNotIn('--profile',ops.compose_command(cfg,'local'))
            cfg.write_text(json.dumps({'observability':{'enabled':True}}))
            self.assertIn('telemetry',ops.compose_command(cfg,'local'))
            cfg.write_text(json.dumps({'observability':{'enabled':'false'}}))
            with self.assertRaises(ValueError):ops.compose_command(cfg,'local')
    def test_bounded_readiness_and_volume_preservation(self):
        calls=[]
        def fake(args,**kwargs):
            calls.append(args)
            return '[]' if 'ps' in args else ''
        with patch.object(ops,'execute',side_effect=fake),patch.object(ops.time,'sleep'):
            with self.assertRaises(TimeoutError):ops.wait_ready(['docker','compose'],['backend'],timeout=0)
        self.assertFalse(any('-v' in c or 'down' in c for c in calls))
    def test_disabled_telemetry_stops_existing_collectors(self):
        with patch.object(ops,'execute') as run:
            ops.synchronize_telemetry(['docker','compose'])
            run.assert_called_once_with(['docker','compose','stop']+sorted(ops.TELEMETRY))
        with patch.object(ops,'execute') as run:
            ops.synchronize_telemetry(['docker','compose','--profile','telemetry'])
            run.assert_not_called()

    def test_unknown_service_rejected(self):
        with self.assertRaises(ValueError):ops.restart(['docker','compose'],'database; rm',timeout=0)
    def test_restart_failure_captures_logs(self):
        calls=[]
        def fake(args,**kwargs):
            calls.append(args)
            if 'up' in args:raise RuntimeError('failed')
            return ''
        with patch.object(ops,'execute',side_effect=fake):
            with self.assertRaises(RuntimeError):ops.restart(['docker','compose'],'backend',timeout=1)
        self.assertTrue(any('logs' in c for c in calls))

    def test_running_telemetry_is_not_ready_until_probe_passes(self):
        calls=[]
        def fake(args,**kwargs):
            calls.append(args)
            if 'ps' in args:return '[{"Service":"collector","State":"running","Health":""}]'
            if 'wget' in args:raise RuntimeError('not listening')
            return ''
        with patch.object(ops,'execute',side_effect=fake),patch.object(ops.time,'sleep'):
            with self.assertRaises(TimeoutError):ops.restart(['docker','compose'],'collector',timeout=.01)
        self.assertTrue(any('wget' in c and 'http://collector:13133/' in c for c in calls))
        self.assertTrue(any('logs' in c for c in calls))
        self.assertFalse(any('down' in c or '-v' in c for c in calls))

    def test_restart_probes_only_selected_telemetry_service(self):
        calls=[]
        def fake(args,**kwargs):
            calls.append(args)
            if 'ps' in args:return '[{"Service":"grafana","State":"running","Health":""}]'
            return ''
        with patch.object(ops,'execute',side_effect=fake):ops.restart(['docker','compose'],'grafana',timeout=1)
        probes=[c[-1] for c in calls if 'wget' in c]
        self.assertEqual(probes,['http://grafana:3000/api/health'])


class CapturedFailure(unittest.TestCase):
    def test_timeout_keeps_private_output(self):
        import subprocess
        with tempfile.TemporaryDirectory() as temp, patch.object(ops,'ROOT',Path(temp)), patch.object(ops.subprocess,'run',side_effect=subprocess.TimeoutExpired(['slow'],1,output=b'failure evidence')):
            with self.assertRaises(RuntimeError):ops.execute(['slow'],timeout=1)
            logs=list((Path(temp)/'.local/logs').glob('*.log'))
            self.assertEqual(len(logs),1)
            self.assertEqual(logs[0].read_text(),'failure evidence')
            self.assertEqual(logs[0].stat().st_mode & 0o777,0o600)

if __name__=='__main__':unittest.main()
