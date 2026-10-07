"""Recovery ordering for the opt-in, retained-storage process drill."""
import importlib.util
import io
from pathlib import Path
import unittest
spec=importlib.util.spec_from_file_location('operations_probe',Path(__file__).with_name('test_operations_integration.py'))
probe=importlib.util.module_from_spec(spec)
spec.loader.exec_module(probe)

class GroupRestartTest(unittest.TestCase):
    def test_restart_is_attempted_after_uncertain_stop_failure(self):
        calls=[]
        def operation(*args):
            calls.append(args)
            if args[0]=='stop':raise RuntimeError('uncertain stop')
        def request(*args):return io.BytesIO(b'{"version":4}')
        with self.assertRaisesRegex(RuntimeError,'uncertain stop'):
            probe.grouped_restart(operation,request,'csrf','room','match')
        self.assertEqual(calls,[('stop','backend','postgres','redis'),('start','--wait','--wait-timeout','90','postgres','redis','backend')])

class AssetRoutingTest(unittest.TestCase):
    def test_asset_rejects_spa_fallback_even_with_ok_status(self):
        import tempfile
        with tempfile.TemporaryDirectory(dir=probe.ROOT/'.local') as directory:
            root=Path(directory);(root/'worker.js').write_bytes(b'worker asset')
            class Response(io.BytesIO):
                headers={'Content-Type':'text/html'}
                status=200
            with self.assertRaisesRegex(AssertionError,'asset MIME'):
                probe.verify_assets(lambda path:Response(b'<html>fallback</html>'),root)
