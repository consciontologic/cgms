#!/usr/bin/env python3
"""Opt-in probe of an already-started local or TLS CGMS Compose stack.
Creates a three-seat match; retains it to test persisted restart state.
Never prints cookies, CSRF, invitations, match payloads or secret URLs.
"""
import argparse
import hashlib
import base64
import http.cookiejar
import json
from pathlib import Path
import secrets
import socket
import subprocess
import sys
import time
import struct
import ssl
import urllib.error
import urllib.parse
import urllib.request

ROOT=Path(__file__).resolve().parents[1]

def verify_assets(request,root):
    """Verify bytes and MIME for emitted artifacts; never infer renderer use."""
    paths=sorted(p for p in root.rglob('*') if p.suffix in ('.js','.wasm'))
    fonts=sorted(root.rglob('*.ttf'))
    if fonts:paths.append(fonts[0])
    assert paths,'missing built assets'
    worker_assets=[]
    for path in paths:
        relative=path.relative_to(root).as_posix()
        with request('/'+relative) as response:
            mime=response.headers.get('Content-Type','').split(';')[0]
            allowed={'.js':('application/javascript','text/javascript'),'.wasm':('application/wasm',),'.ttf':('font/ttf','application/octet-stream','application/x-font-ttf')}
            assert mime in allowed[path.suffix],f'asset MIME mismatch: {relative}'
            assert hashlib.sha256(response.read()).digest()==hashlib.sha256(path.read_bytes()).digest(),f'asset routing/bytes mismatch: {relative}'
        if 'worker' in path.name:worker_assets.append(relative)
    try:request('/main.dart.js.map')
    except urllib.error.HTTPError as error:assert error.code==404
    else:raise AssertionError('source map exposed')
    print(f'PASS: {len(paths)} emitted JS/WASM/font assets match local bytes and MIME; worker-named assets={worker_assets}; routing only, not renderer or service-worker execution')

def grouped_restart(operation,request,csrf,room_id,match_id):
    """Stop/restart only test service processes; retain every data volume."""
    room=json.load(request('/v1/rooms/'+room_id))
    snapshot=json.load(request('/v1/matches/'+match_id))
    # Let initial deal/initiative/turn scheduling reach an acknowledged stable
    # boundary before measuring process-loss recovery.
    stable=0
    deadline=time.monotonic()+15
    while stable<3:
        time.sleep(.5)
        current=json.load(request('/v1/matches/'+match_id))
        stable=stable+1 if current==snapshot else 0
        snapshot=current
        assert time.monotonic()<deadline,'match did not reach quiescent checkpoint'
    command={'command_id':secrets.token_hex(12),'capacity':3,'games_per_match':3}
    receipt=json.load(request('/v1/rooms',command,csrf))
    started=time.monotonic()
    try:
        operation('stop','backend','postgres','redis')
        try:request('/health/ready')
        except urllib.error.HTTPError as error:assert error.code>=500
        except (TimeoutError,OSError):pass
        else:raise AssertionError('stopped authority falsely ready')
    finally:
        operation('start','--wait','--wait-timeout','90','postgres','redis','backend')
    deadline=time.monotonic()+90
    while True:
        try:
            if request('/health/ready').status==204:break
        except (urllib.error.URLError,OSError):pass
        assert time.monotonic()<deadline,'grouped process restart failed to recover'
        time.sleep(.5)
    assert json.load(request('/v1/rooms/'+room_id))==room,'grouped restart lost room/session'
    assert json.load(request('/v1/matches/'+match_id))==snapshot,'grouped restart changed acknowledged projection'
    assert json.load(request('/v1/rooms',command,csrf))==receipt,'grouped restart lost idempotency receipt'
    print(f'PASS: grouped backend/PostgreSQL/Redis process restart recovered in {time.monotonic()-started:.3f}s with identical room, match projection and idempotent receipt; retained storage only')

def probe(base,certificate=None,restart=False,idle=0,failures=False,group_restart=False,assets=False):
    parsed=urllib.parse.urlparse(base)
    context=ssl.create_default_context(cafile=str(certificate)) if certificate else None
    def client():
        jar=http.cookiejar.CookieJar()
        handlers=[urllib.request.HTTPCookieProcessor(jar)]
        if context:handlers.append(urllib.request.HTTPSHandler(context=context))
        opener=urllib.request.build_opener(*handlers)
        def request(path,body=None,csrf=None,origin=None):
            headers={'Origin':origin or base,'Content-Type':'application/json'}
            if csrf:headers['X-CSRF-Token']=csrf
            data=None if body is None else json.dumps(body).encode()
            return opener.open(urllib.request.Request(base+path,data=data,headers=headers),timeout=5)
        response=request('/v1/guests',{})
        guest=json.load(response)
        return jar,request,guest['csrf_token']
    actors=[client() for _ in range(3)]
    jar,request,csrf=actors[0]
    if assets:verify_assets(request,ROOT/'client/build/web')
    for path in ('/health/live','/health/ready'):
        assert request(path).status==204
    with request('/') as response:
        assert response.status==200
        if context:
            assert response.headers['Strict-Transport-Security']=='max-age=31536000'
            assert "frame-ancestors 'none'" in response.headers['Content-Security-Policy']
        else:assert response.headers.get('Strict-Transport-Security') is None
    for path in ('/internal/metrics',):
        try:request(path)
        except urllib.error.HTTPError as error:assert error.code==404
        else:raise AssertionError('private endpoint exposed')
    try:request('/v1/session',origin='https://foreign.invalid')
    except urllib.error.HTTPError as error:assert error.code==403
    else:raise AssertionError('foreign origin accepted')
    room=json.load(request('/v1/rooms',{'command_id':secrets.token_hex(12),'capacity':3,'games_per_match':3},csrf))
    room_id=room['id']
    for _,join,join_csrf in actors[1:]:
        invitation=json.load(request('/v1/rooms/'+room_id+'/invitations',{},csrf))['invitation']
        join('/v1/rooms/'+room_id+'/join',{'invitation':invitation},join_csrf).close()
    match=json.load(request('/v1/rooms/'+room_id+'/matches',{'command_id':secrets.token_hex(12)},csrf))['match_id']
    cookie='; '.join(c.name+'='+c.value for c in jar)
    connection=socket.create_connection((parsed.hostname,parsed.port),timeout=5)
    if context:connection=context.wrap_socket(connection,server_hostname=parsed.hostname)
    with connection:
        key=base64.b64encode(secrets.token_bytes(16)).decode()
        message=f'GET /v1/matches/{match}/stream?after=0 HTTP/1.1\r\nHost: {parsed.netloc}\r\nOrigin: {base}\r\nCookie: {cookie}\r\nConnection: Upgrade\r\nUpgrade: websocket\r\nSec-WebSocket-Version: 13\r\nSec-WebSocket-Key: {key}\r\n\r\n'
        connection.sendall(message.encode())
        response=connection.recv(16384)
        assert response.startswith(b'HTTP/1.1 101 '),'WebSocket upgrade failed'
        if idle:
            buffer=bytearray(response.split(b'\r\n\r\n',1)[1])
            def read_exact(n):
                while len(buffer)<n:
                    chunk=connection.recv(65536)
                    assert chunk,'idle stream disconnected'
                    buffer.extend(chunk)
                value=bytes(buffer[:n]);del buffer[:n];return value
            connection.settimeout(30)
            deadline=time.monotonic()+idle
            pings=0
            while time.monotonic()<deadline:
                first,second=read_exact(2)
                length=second&127
                if length==126:length=struct.unpack('!H',read_exact(2))[0]
                if length==127:length=struct.unpack('!Q',read_exact(8))[0]
                assert length<2**20,'unexpected frame size'
                payload=read_exact(length)
                assert first&15!=8,'idle stream closed'
                if first&15==9:
                    pings+=1
                    mask=secrets.token_bytes(4)
                    connection.sendall(bytes([0x8a,0x80|len(payload)])+mask+bytes(v^mask[i%4] for i,v in enumerate(payload)))
            assert pings>=2,'missing periodic heartbeat'
            print('PASS: quiet WebSocket survives proxy idle timeout with heartbeats')
    if restart:
        before=json.load(request('/v1/rooms/'+room_id))
        for _ in range(2):
            subprocess.run([sys.executable,str(ROOT/'xops/makefile/service_ops.py'),'restart','--service','backend','--mode','production' if context else 'local'],check=True,cwd=ROOT)
            after=json.load(request('/v1/rooms/'+room_id))
            assert after==before,'persisted room or session lost after restart'
        print('PASS: two backend recreations preserve authenticated room and match identity')
    if failures or group_restart:
        compose=['docker','compose','-f',str(ROOT/'deploy/compose/compose.yaml')]
        if context:compose+=['-f',str(ROOT/'deploy/compose/production.yaml')]
        def operation(*args):
            result=subprocess.run(compose+list(args),cwd=ROOT,stdout=subprocess.PIPE,stderr=subprocess.STDOUT,timeout=120)
            assert result.returncode==0,'controlled service operation failed'
        if group_restart:grouped_restart(operation,request,csrf,room_id,match)
    if failures:
        before=json.load(request('/v1/rooms/'+room_id))
        try:
            operation('stop','redis')
            assert json.load(request('/v1/rooms/'+room_id))==before,'Redis loss changed durable authority'
        finally:operation('start','redis')
        print('PASS: Redis outage preserves authenticated PostgreSQL-backed room read')
        failed_id=secrets.token_hex(12)
        body={'command_id':failed_id,'capacity':3,'games_per_match':3}
        try:
            operation('stop','postgres')
            try:
                request('/v1/rooms',body,csrf)
            except urllib.error.HTTPError as error:
                assert error.code>=500,'database outage did not fail unavailable'
            except (TimeoutError,OSError):pass
            else:raise AssertionError('database outage falsely acknowledged room creation')
            try:request('/health/ready')
            except urllib.error.HTTPError as error:assert error.code==503
            except (TimeoutError,OSError):pass
            else:raise AssertionError('database outage falsely ready')
        finally:operation('start','postgres')
        deadline=time.monotonic()+90
        while True:
            try:
                if request('/health/ready').status==204:break
            except (urllib.error.URLError,OSError):pass
            assert time.monotonic()<deadline,'authority failed to recover after database restart'
            time.sleep(.5)
        assert json.load(request('/v1/rooms/'+room_id))==before,'database restart lost acknowledged room'
        recovered=json.load(request('/v1/rooms',body,csrf))
        retried=json.load(request('/v1/rooms',body,csrf))
        assert recovered==retried,'post-outage retry duplicated room creation'
        print('PASS: database outage rejects writes/readiness; restart preserves acknowledged room and idempotent retry')
    logs=subprocess.check_output(['docker','compose','-f',str(ROOT/'deploy/compose/compose.yaml'),'logs','--no-color','backend'],cwd=ROOT,text=True,stderr=subprocess.STDOUT)
    for account_jar,_,account_csrf in actors:
        for private in [account_csrf]+[item.value for item in account_jar]:
            assert private not in logs,'private session material leaked into backend logs'
    print('PASS: live logs contain none of the probe session cookies or CSRF tokens')
    print('PASS: readiness, assets, private endpoint denial, origin rejection, three-seat start, WebSocket 101'+(', trusted local test TLS and production headers' if context else ', local profile without HSTS'))

if __name__=='__main__':
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--base',default='http://localhost:8080')
    parser.add_argument('--certificate',type=Path)
    parser.add_argument('--restart',action='store_true')
    parser.add_argument('--idle',type=int,default=0)
    parser.add_argument('--failures',action='store_true')
    parser.add_argument('--group-restart',action='store_true')
    parser.add_argument('--assets',action='store_true')
    args=parser.parse_args()
    probe(args.base,args.certificate,args.restart,args.idle,args.failures,args.group_restart,args.assets)
