#!/usr/bin/env python3
"""Compile the current HEAD reader in ignored state and rehearse schema compatibility."""
import io
import os
from pathlib import Path
import subprocess
import sys
import tarfile
ROOT=Path(__file__).resolve().parents[1]
sys.path.insert(0,str(ROOT/'xops/makefile'))
from service_ops import execute

def main():
    revision=execute(['git','rev-parse','HEAD']).strip()
    directory=ROOT/'.local/rollback'/revision;directory.mkdir(parents=True,exist_ok=True)
    archive=subprocess.run(['git','archive',revision,'backend'],cwd=ROOT,stdout=subprocess.PIPE,stderr=subprocess.PIPE,timeout=30,check=True).stdout
    with tarfile.open(fileobj=io.BytesIO(archive)) as tar:tar.extractall(directory,filter='data')
    helper=directory/'backend/cmd/rollback-reader';helper.mkdir(parents=True,exist_ok=True)
    (helper/'main.go').write_text((ROOT/'xops/rollback_reader.go.template').read_text())
    binary=directory/'rollback-reader'
    execute(['go','build','-o',str(binary),'./cmd/rollback-reader'],cwd=directory/'backend',timeout=180)
    print('Archived reader revision: '+revision,flush=True)
    env=dict(os.environ,CGMS_ROLLBACK_READER=str(binary))
    return subprocess.run([sys.executable,str(ROOT/'xops/postgres_integration.py'),'-tags=integration,rollback','./internal/matchstore','-run','Test(Pinned|Economy)RollbackReaders','-v'],env=env,cwd=ROOT,timeout=300).returncode

if __name__=='__main__':sys.exit(main())
