#!/usr/bin/env python3
"""Private logical backup and isolated restore drill; never replaces a live database."""
import argparse
import datetime
import hashlib
import json
import os
from pathlib import Path
import secrets
import shutil
import subprocess
import time
from service_ops import ROOT, execute

COMPOSE=['docker','compose','-f',str(ROOT/'deploy/compose/compose.yaml')]

def digest(path):
    with path.open('rb') as stream:return hashlib.file_digest(stream,'sha256').hexdigest()

def manifest_for(root,names):
    return {'version':1,'files':{name:digest(root/name) for name in names}}

def verify_manifest(root,manifest):
    for name,expected in manifest['files'].items():
        path=(root/name).resolve()
        if not path.is_relative_to(root.resolve()) or path.is_symlink():raise ValueError('invalid backup path')
        if digest(path)!=expected:raise ValueError('backup integrity mismatch')
    if 'database.dump' not in manifest['files']:raise ValueError('database archive missing')

def binary_command(command,path,restore=False):
    logs=path.parent/'diagnostics.log'
    with logs.open('ab') as diagnostics,path.open('rb' if restore else 'xb') as data:
        result=subprocess.run(command,stdin=data if restore else subprocess.DEVNULL,stdout=diagnostics if restore else data,stderr=diagnostics,timeout=300,cwd=ROOT)
    if result.returncode:raise RuntimeError('database command failed; inspect private backup diagnostics')

def backup():
    os.umask(0o077)
    stamp=datetime.datetime.now(datetime.timezone.utc).strftime('%Y%m%dT%H%M%SZ')
    root=ROOT/'.local/backups'/(stamp+'-'+secrets.token_hex(3));root.mkdir(parents=True,mode=0o700)
    start=time.monotonic()
    # pg_dump uses one consistent MVCC snapshot across every included table.
    binary_command(COMPOSE+['exec','-T','postgres','pg_dump','-U','cgms','-d','cgms','--format=custom'],root/'database.dump')
    config=json.loads(execute(COMPOSE+['config','--format','json']))
    names=['database.dump']
    # Preserve workspace service/rules files and executable without
    # copying credentials. Database-pinned rule snapshots remain in the dump.
    paths=sorted(p for p in (ROOT/'config').rglob('*') if p.is_file())+[ROOT/'.local/bin/server']
    for source in paths:
        name=str(source.relative_to(ROOT));target=root/name;target.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(source,target);names.append(name)
    manifest=manifest_for(root,names)
    manifest.update({'artifact_scope':'unverified workspace configuration and binary snapshots; not deployed runtime identity','snapshot_started_utc':stamp,'backup_seconds':round(time.monotonic()-start,3),'images':{k:v.get('image') for k,v in config['services'].items()},'postgres_image':config['services']['postgres']['image']})
    (root/'manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
    print('Private backup: '+str(root.relative_to(ROOT)))
    return root

def restore_drill(root):
    os.umask(0o077)
    root=root.resolve()
    if not root.is_relative_to((ROOT/'.local/backups').resolve()):raise ValueError('drill requires a local managed backup')
    manifest=json.loads((root/'manifest.json').read_text());verify_manifest(root,manifest)
    name='cgms-restore-'+secrets.token_hex(6)
    start=time.monotonic();created=False
    try:
        # Own this unique name before contacting Docker: timeout can occur after
        # daemon-side creation even when no success reached this process.
        created=True
        execute(['docker','run','-d','--name',name,'--network','none','--tmpfs','/var/lib/postgresql/data:rw,size=512m','--env','POSTGRES_HOST_AUTH_METHOD=trust','--env','POSTGRES_USER=cgms','--env','POSTGRES_DB=cgms',manifest['postgres_image'],'postgres','-c','listen_addresses='])
        deadline=time.monotonic()+60
        while True:
            try:execute(['docker','exec',name,'pg_isready','-U','cgms','-d','cgms'],timeout=5);break
            except RuntimeError:
                if time.monotonic()>=deadline:raise TimeoutError('restore database readiness expired')
                time.sleep(.25)
        binary_command(['docker','exec','-i',name,'pg_restore','-U','cgms','-d','cgms','--exit-on-error','--single-transaction','--no-owner','--no-privileges'],root/'database.dump',restore=True)
        # Read all restored rows and constraints; only aggregates leave the
        # isolated container. No application identifiers or state are logged.
        sql="SELECT count(*) FROM information_schema.tables WHERE table_schema='public'; SELECT count(*) FROM matches; SELECT count(*) FROM command_results; SELECT count(*) FROM financial_debts; SELECT count(*) FROM pg_constraint WHERE connamespace='public'::regnamespace AND NOT convalidated;"
        counts=[int(v) for v in execute(['docker','exec',name,'psql','-U','cgms','-d','cgms','-At','-v','ON_ERROR_STOP=1','-c',sql]).splitlines()]
        if counts[0]<10 or counts[-1]!=0:raise RuntimeError('restored schema validation failed')
        result={'restore_seconds':round(time.monotonic()-start,3),'tables':counts[0],'matches':counts[1],'command_receipts':counts[2],'debts':counts[3],'unvalidated_constraints':counts[4],'scope':'isolated logical restore, not host-loss recovery or application replay'}
        (root/'restore-report.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result))
    finally:
        if created:execute(['docker','rm','-f',name])

if __name__=='__main__':
    parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('action',choices=['backup','drill']);parser.add_argument('--backup',type=Path);args=parser.parse_args()
    if args.action=='backup':backup()
    else:restore_drill(args.backup or backup())
