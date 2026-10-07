#!/usr/bin/env python3
"""Fresh workspace/volume production-Compose rehearsal on the existing Docker host."""
import http.cookiejar
import json
import os
from pathlib import Path
import secrets
import shutil
import socket
import ssl
import subprocess
import sys
import time
import urllib.request
ROOT=Path(__file__).resolve().parents[1]
sys.path.insert(0,str(ROOT/'xops/makefile'))
from service_ops import execute,wait_ready

def main():
    os.umask(0o077)
    project='cgms-rehearsal-'+secrets.token_hex(5)
    target=ROOT/'.local/rehearsals'/project;target.mkdir(parents=True,mode=0o700)
    for name in ['config','deploy','client/build/web']:
        shutil.copytree(ROOT/name,target/name)
    for name in ['.dockerignore','.local/bin/server','xops/makefile/service_ops.py']:
        path=target/name;path.parent.mkdir(parents=True,exist_ok=True);shutil.copy2(ROOT/name,path)
    execute([sys.executable,str(target/'xops/makefile/service_ops.py'),'init'])
    for name in ['tls_certificate','tls_key']:
        shutil.copyfile(ROOT/'.local/secrets'/name,target/'.local/secrets'/name)
    with socket.socket() as reservation:
        reservation.bind(('127.0.0.1',0));port=reservation.getsockname()[1]
    selected=target/'config/backend/production.json'
    settings=json.loads(selected.read_text());settings['origin']=f'https://localhost:{port}';selected.write_text(json.dumps(settings,indent=2)+'\n')
    env=dict(os.environ,CGMS_CONFIG_FILE=str(selected),CGMS_UID=str(os.getuid()),CGMS_GID=str(os.getgid()))
    base=['docker','compose','-f',str(target/'deploy/compose/compose.yaml'),'-f',str(target/'deploy/compose/production.yaml')]
    config=json.loads(execute(base+['config','--format','json'],env=env))
    config['name']=project
    config['services']={name:service for name,service in config['services'].items() if name in ['postgres','redis','backend','web']}
    for collection in ['volumes','networks']:
        for name,value in config[collection].items():value['name']=project+'_'+name
    config['services']['web']['ports'][0]['published']=str(port)
    path=target/'resolved-compose.json';path.write_text(json.dumps(config,indent=2)+'\n')
    command=['docker','compose','-p',project,'-f',str(path)]
    started=time.monotonic()
    try:
        execute(command+['build','backend'],timeout=180)
        execute(command+['up','-d'],timeout=180)
        wait_ready(command,['postgres','redis','backend','web'],timeout=90)
        context=ssl.create_default_context(cafile=str(target/'.local/secrets/tls_certificate'))
        jar=http.cookiejar.CookieJar();client=urllib.request.build_opener(urllib.request.HTTPSHandler(context=context),urllib.request.HTTPCookieProcessor(jar))
        origin=f'https://localhost:{port}'
        def request(resource,body=None,csrf=None):
            headers={'Origin':origin,'Content-Type':'application/json'}
            if csrf:headers['X-CSRF-Token']=csrf
            with client.open(urllib.request.Request(origin+resource,data=None if body is None else json.dumps(body).encode(),headers=headers),timeout=10) as response:
                return response.status,response.read()
        assert request('/health/ready')[0]==204
        _,raw=request('/v1/guests',{});guest=json.loads(raw)
        status,raw=request('/v1/rooms',{'command_id':secrets.token_hex(12),'capacity':3,'games_per_match':3},guest['csrf_token']);assert status==200 and json.loads(raw)['capacity']==3
        ids=execute(command+['ps','-q']).splitlines()
        inspection=json.loads(execute(['docker','inspect']+ids))
        for item in inspection:
            service=item['Config']['Labels']['com.docker.compose.service']
            bindings=item['HostConfig'].get('PortBindings') or {}
            if service!='web':assert not bindings,'private service published a port'
            if service=='backend':assert item['HostConfig']['ReadonlyRootfs'] and 'ALL' in item['HostConfig']['CapDrop'] and item['Config']['User']!='0'
        report={'project':project,'seconds':round(time.monotonic()-started,3),'scope':'fresh workspace, fresh secrets and volumes on existing Docker host; not clean OS/host','tls_hostname_verified':True,'guest_and_room_created':True,'private_ports_verified':True,'backend_readonly_capdrop_nonroot':True,'postgres_volume':project+'_postgres_data'}
        (target/'report.json').write_text(json.dumps(report,indent=2)+'\n');print(json.dumps(report))
    except Exception:
        execute(command+['logs','--no-color','--tail','100'],timeout=15)
        raise
    finally:
        # Project-unique containers/network only; retain its durable volumes.
        execute(command+['down'],timeout=60)

if __name__=='__main__':main()
