#!/usr/bin/env python3
"""Bounded local Compose operations. Never deletes a volume or runs down -v."""
import argparse
import json
import os
from pathlib import Path
import secrets
import subprocess
import sys
import time

ROOT=Path(__file__).resolve().parents[2]
SERVICES={'backend','web','postgres','redis','collector','tempo','prometheus','grafana'}
TELEMETRY={'collector','tempo','prometheus','grafana'}

def execute(args,timeout=120,env=None,cwd=ROOT):
    try:
        result=subprocess.run(args,cwd=cwd,env=env,text=True,stdout=subprocess.PIPE,stderr=subprocess.STDOUT,timeout=timeout)
        output,code=result.stdout,result.returncode
    except subprocess.TimeoutExpired as error:
        output=error.stdout or ''
        if isinstance(output,bytes):output=output.decode('utf-8',errors='replace')
        code=124
    # Operational output can contain service diagnostics. Retain in private files;
    # report only exit status and log path to the public terminal.
    logs=ROOT/'.local/logs';logs.mkdir(parents=True,exist_ok=True,mode=0o700)
    path=logs/(str(time.time_ns())+'.log')
    fd=os.open(path,os.O_WRONLY|os.O_CREAT|os.O_EXCL,0o600)
    with os.fdopen(fd,'w') as stream:stream.write(output)
    if code:raise RuntimeError(f'command failed ({code}); private log: {path.relative_to(ROOT)}')
    return output

def compose_command(config,mode):
    c=json.loads(Path(config).read_text())
    enabled=c.get('observability',{}).get('enabled')
    if type(enabled) is not bool:raise ValueError('observability.enabled must be boolean')
    command=['docker','compose','-f',str(ROOT/'deploy/compose/compose.yaml')]
    if mode=='production':command+=['-f',str(ROOT/'deploy/compose/production.yaml')]
    if enabled:command+=['--profile','telemetry']
    return command

def wait_ready(command,services,timeout=60):
    deadline=time.monotonic()+timeout
    while time.monotonic()<deadline:
        raw=execute(command+['ps','--format','json'],timeout=10)
        try:
            rows=json.loads(raw)
            if isinstance(rows,dict):rows=[rows]
        except json.JSONDecodeError:rows=[json.loads(line) for line in raw.splitlines() if line]
        ready={r['Service'] for r in rows if r.get('State')=='running' and r.get('Health','') in ('','healthy')}
        if set(services)<=ready:return
        time.sleep(.25)
    raise TimeoutError('bounded service readiness expired')

def restart(command,service,timeout=60):
    if service not in SERVICES:raise ValueError('unsupported service')
    try:
        execute(command+['up','-d','--no-deps','--force-recreate',service])
        deadline=time.monotonic()+timeout
        wait_ready(command,[service],timeout)
        if service in TELEMETRY:
            telemetry_ready(command,[service],timeout=max(0,deadline-time.monotonic()))
    except Exception:
        execute(command+['logs','--no-color','--tail','200',service],timeout=15)
        raise

def synchronize_telemetry(command):
    if '--profile' not in command:
        execute(command+['stop']+sorted(TELEMETRY))

def telemetry_ready(command,services=None,timeout=60):
    # Probe across the private network with the already-present backend wget.
    endpoints={'collector':'http://collector:13133/','tempo':'http://tempo:3200/ready','prometheus':'http://prometheus:9090/-/ready','grafana':'http://grafana:3000/api/health'}
    deadline=time.monotonic()+timeout
    for service in (sorted(TELEMETRY) if services is None else services):
        endpoint=endpoints[service]
        while time.monotonic()<deadline:
            try:
                execute(command+['exec','-T','backend','wget','-q','-T','2','-O','/dev/null',endpoint],timeout=5)
                break
            except RuntimeError:
                time.sleep(min(1,max(0,deadline-time.monotonic())))
        else:raise TimeoutError('telemetry readiness expired')

def initialize():
    directory=ROOT/'.local/secrets';directory.mkdir(parents=True,exist_ok=True,mode=0o700)
    password=directory/'database_password'
    if password.exists():secret=password.read_text().strip()
    else:
        secret=secrets.token_hex(32)
        fd=os.open(password,os.O_WRONLY|os.O_CREAT|os.O_EXCL,0o600)
        with os.fdopen(fd,'w') as stream:stream.write(secret)
    for name,value in [('database_url','postgres://cgms:'+secret+'@postgres/cgms?sslmode=disable'),('grafana_password',secrets.token_hex(32))]:
        path=directory/name
        if not path.exists():
            fd=os.open(path,os.O_WRONLY|os.O_CREAT|os.O_EXCL,0o600)
            with os.fdopen(fd,'w') as stream:stream.write(value)

    os.chmod(directory/'grafana_password',0o640)
    redis=directory/'redis_config'
    if redis.exists():redis_secret=redis.read_text().split()[1]
    else:
        redis_secret=secrets.token_hex(32)
        fd=os.open(redis,os.O_WRONLY|os.O_CREAT|os.O_EXCL,0o600)
        with os.fdopen(fd,'w') as stream:stream.write('requirepass '+redis_secret+'\n')
    redis_url=directory/'redis_url'
    if not redis_url.exists():
        fd=os.open(redis_url,os.O_WRONLY|os.O_CREAT|os.O_EXCL,0o600)
        with os.fdopen(fd,'w') as stream:stream.write('redis://:'+redis_secret+'@redis:6379')

def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('action',choices=['init','up','restart','go','web','stop','config'])
    parser.add_argument('--service',choices=sorted(SERVICES),default='backend')
    parser.add_argument('--mode',choices=['local','production'],default='local')
    parser.add_argument('--config',type=Path)
    args=parser.parse_args()
    if args.action=='init':initialize();print('Private local secrets initialized; existing values retained.');return
    config=(args.config or ROOT/'config/backend'/f'{args.mode}.json').resolve()
    command=compose_command(config,args.mode)
    os.environ['CGMS_CONFIG_FILE']=str(config)
    os.environ['CGMS_UID']=str(os.getuid());os.environ['CGMS_GID']=str(os.getgid())
    # Reject invalid settings before Compose can launch anything. Secrets are
    # validated inside the container against its mounted /run/secrets paths.
    if args.action in ('up','go','config'):
        (ROOT/'.local/bin').mkdir(parents=True,exist_ok=True)
        execute(['go','-C','backend','build','-o','../.local/bin/server','./cmd/server'],env=dict(os.environ,CGO_ENABLED='0'))
        execute([str(ROOT/'.local/bin/server'),'check-config',str(config)])
    if args.action=='config':execute(command+['config','--quiet']);print('Configuration valid.');return
    if args.action=='stop':execute(command+['stop']+sorted(SERVICES));print('Services stopped; volumes retained.');return
    if args.action=='restart' and args.service in TELEMETRY and '--profile' not in command:
        raise ValueError('telemetry service disabled by selected configuration')
    if args.action=='up':
        initialize()
        build_web()
    synchronize_telemetry(command)
    if args.action=='web':build_web()
    if args.action in ('up','go'):execute(command+['build','backend'],timeout=180)
    try:
        if args.action=='up':
            execute(command+['up','-d','--force-recreate'],timeout=180)
            wait_ready(command,['backend','web','postgres','redis']+(sorted(TELEMETRY) if '--profile' in command else []),timeout=90)
            if '--profile' in command: telemetry_ready(command)
        else:restart(command, 'web' if args.action=='web' else 'backend' if args.action=='go' else args.service)
    except Exception:
        execute(command+['logs','--no-color','--tail','200'],timeout=15)
        raise
    print('Services ready. Volumes retained; use browser hard reload after web.re.')

def build_web():
    execute(['flutter','build','web','--no-web-resources-cdn'],timeout=240,cwd=ROOT/'client')

if __name__=='__main__':
    try:main()
    except (ValueError,RuntimeError,TimeoutError,OSError,subprocess.TimeoutExpired) as error:
        print(str(error),file=sys.stderr);sys.exit(1)
