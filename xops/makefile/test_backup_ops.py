import importlib.util
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch
import json
SPEC=importlib.util.spec_from_file_location('backup_ops',Path(__file__).with_name('backup_ops.py'))
ops=importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(ops)

class BackupManifest(unittest.TestCase):
    def test_modified_backup_rejected(self):
        with tempfile.TemporaryDirectory() as directory:
            root=Path(directory)
            (root/'database.dump').write_bytes(b'archive')
            manifest=ops.manifest_for(root,['database.dump'])
            ops.verify_manifest(root,manifest)
            (root/'database.dump').write_bytes(b'changed')
            with self.assertRaises(ValueError):ops.verify_manifest(root,manifest)

    def test_manifest_path_cannot_escape(self):
        with tempfile.TemporaryDirectory() as directory:
            with self.assertRaises(ValueError):ops.verify_manifest(Path(directory),{'files':{'../secret':'hash'}})

class RestoreCleanup(unittest.TestCase):
    def test_uncertain_container_creation_is_cleaned(self):
        with tempfile.TemporaryDirectory() as temporary:
            root=Path(temporary);backup=root/'.local/backups/test';backup.mkdir(parents=True)
            (backup/'database.dump').write_bytes(b'archive')
            manifest=ops.manifest_for(backup,['database.dump']);manifest['postgres_image']='pinned-test-image'
            (backup/'manifest.json').write_text(json.dumps(manifest))
            calls=[]
            def execute(command,**kwargs):
                calls.append(command)
                if 'run' in command:raise TimeoutError('uncertain daemon outcome')
                return ''
            with patch.object(ops,'ROOT',root),patch.object(ops,'execute',side_effect=execute):
                with self.assertRaises(TimeoutError):ops.restore_drill(backup)
            self.assertEqual(calls[-1][:3],['docker','rm','-f'])
            self.assertEqual(calls[-1][-1],calls[0][4])
