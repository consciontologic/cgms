#!/usr/bin/env python3
"""Bounded disposable PostgreSQL integration tests; Docker failures are fatal."""
import os
from pathlib import Path
import secrets
import signal
import socket
import subprocess
import sys
import time
import uuid

ROOT = Path(__file__).resolve().parent.parent
# Resolved from the preinstalled postgres:16-alpine RepoDigests on 2026-09-30.
REDIS_IMAGE = 'redis@sha256:8b81dd37ff027bec4e516d41acfbe9fe2460070dc6d4a4570a2ac5b9d59df065'
IMAGE = 'postgres@sha256:20edbde7749f822887a1a022ad526fde0a47d6b2be9a8364433605cf65099416'


def command(args, **kwargs):
    return subprocess.run(args, check=True, text=True, stdout=subprocess.PIPE,
                          stderr=subprocess.PIPE, timeout=30, **kwargs).stdout.strip()


def cleanup(container, volume, network, created=("container", "volume", "network")):
    good = True
    for kind, args in zip(('redis', 'container', 'volume', 'network'),
                          (['docker', 'rm', '-f', container+'-redis'], ['docker', 'rm', '-f', container], ['docker', 'volume', 'rm', volume],
                           ['docker', 'network', 'rm', network])):
        if kind not in created:
            continue
        try:
            command(args)
        except Exception:
            print('❌ PostgreSQL resource cleanup failed: ' + args[1], file=sys.stderr)
            good = False
    return good


def test_command(args):
    if args and args[0] in ('--webfixture', '--webfixture-economy'):
        if len(args) != 1:
            raise ValueError('webfixture accepts no runner overrides')
        return ['go', 'run', './cmd/webfixture', '-listen', '127.0.0.1:8081',
                '-assets', str(ROOT/'client/build/web')] + (['-economy'] if args[0] == '--webfixture-economy' else [])
    return ['go', 'test', '-tags=integration', '-count=1', '-timeout=5m'] + (args or ['./...'])


def ready(container):
    deadline = time.monotonic() + 40
    while time.monotonic() < deadline:
        try:
            command(['docker', 'exec', container, 'pg_isready', '-h', '127.0.0.1', '-U', 'cgms', '-d', 'cgms'])
            return
        except subprocess.CalledProcessError:
            time.sleep(.25)
    raise TimeoutError('PostgreSQL readiness timed out')


def main(args=None):
    args = sys.argv[1:] if args is None else args
    with_redis = '--redis' in args
    args = [arg for arg in args if arg != '--redis']
    token = 'cgms-integration-' + uuid.uuid4().hex
    child = None
    created = []
    result = 1
    def interrupted(signum, frame):
        raise InterruptedError('integration run interrupted')
    old_handlers = {s: signal.signal(s, interrupted) for s in (signal.SIGINT, signal.SIGTERM)}
    try:
        command(['docker', 'image', 'inspect', IMAGE])  # No implicit pull or installation.
        created.append('volume')
        command(['docker', 'volume', 'create', token])
        created.append('network')
        command(['docker', 'network', 'create', token])
        env = dict(os.environ, POSTGRES_PASSWORD=secrets.token_hex(24))
        # Pin the selected port across crash/restart. A bind race fails startup
        # safely; Docker must never silently choose a replacement port.
        with socket.socket() as reservation:
            reservation.bind(('127.0.0.1', 0))
            port = reservation.getsockname()[1]
        created.append('container')
        command(['docker', 'run', '-d', '--pull=never', '--name', token,
                 '--network', token, '-p', f'127.0.0.1:{port}:5432', '--memory=512m', '--cpus=2',
                 '--mount', 'type=volume,src='+token+',dst=/var/lib/postgresql/data',
                 '-e', 'POSTGRES_USER=cgms', '-e', 'POSTGRES_DB=cgms',
                 '-e', 'POSTGRES_PASSWORD', IMAGE], env=env)
        ready(token)
        endpoint = command(['docker', 'port', token, '5432/tcp'])
        if not endpoint.startswith('127.0.0.1:') or not endpoint.rsplit(':', 1)[1].isdigit():
            raise ValueError('database is not loopback-only')
        env['CGMS_TEST_DATABASE_URL'] = 'postgres://cgms:'+env.pop('POSTGRES_PASSWORD')+'@'+endpoint+'/cgms?sslmode=disable'
        # Tests may docker kill/start this exact disposable container for WAL recovery.
        env['CGMS_TEST_POSTGRES_CONTAINER'] = token
        if with_redis:
            command(['docker', 'image', 'inspect', REDIS_IMAGE])
            with socket.socket() as reservation:
                reservation.bind(('127.0.0.1', 0))
                redis_port = reservation.getsockname()[1]
            redis_name = token+'-redis'
            redis_env = dict(os.environ, CGMS_REDIS_PASSWORD=secrets.token_hex(24))
            created.append('redis')
            command(['docker', 'run', '-d', '--pull=never', '--name', redis_name,
                     '--network', token, '-p', f'127.0.0.1:{redis_port}:6379',
                     '--memory=128m', '--cpus=1', '-e', 'CGMS_REDIS_PASSWORD',
                     REDIS_IMAGE, 'sh', '-c',
                     'umask 077; printf "requirepass %s\\n" "$CGMS_REDIS_PASSWORD" > /tmp/cgms-redis.conf; exec redis-server /tmp/cgms-redis.conf --save "" --appendonly no --maxmemory 64mb --maxmemory-policy allkeys-lru'], env=redis_env)
            deadline = time.monotonic()+30
            while True:
                try:
                    pong = command(['docker', 'exec', redis_name, 'sh', '-c',
                                    'REDISCLI_AUTH="$CGMS_REDIS_PASSWORD" redis-cli ping'])
                    if pong == 'PONG':
                        break
                except subprocess.CalledProcessError:
                    pass
                if time.monotonic() >= deadline:
                    raise TimeoutError('Redis readiness timed out')
                time.sleep(.1)
            redis_endpoint = command(['docker', 'port', redis_name, '6379/tcp'])
            if not redis_endpoint.startswith('127.0.0.1:'):
                raise ValueError('Redis is not loopback-only')
            env['CGMS_TEST_REDIS_URL'] = 'redis://:'+redis_env['CGMS_REDIS_PASSWORD']+'@'+redis_endpoint
            env['CGMS_TEST_REDIS_CONTAINER'] = redis_name
        if args == ['--self-test']:
            sql = ['docker', 'exec', token, 'psql', '-U', 'cgms', '-d', 'cgms', '-Atqc']
            command(sql + ['CREATE TABLE durable_probe(v int); INSERT INTO durable_probe VALUES(42);'])
            command(['docker', 'kill', '--signal=KILL', token])
            command(['docker', 'start', token]); ready(token)
            if command(['docker', 'port', token, '5432/tcp']) != endpoint:
                raise ValueError('published endpoint changed after restart')
            if command(sql + ['SELECT v FROM durable_probe']) != '42':
                raise ValueError('durability probe failed')
            print('✅ PostgreSQL committed row survived SIGKILL/restart on disk volume')
            result = 0
        else:
            print('🔎 Running PostgreSQL integration tests (loopback, disposable disk volume)', flush=True)
            child = subprocess.Popen(test_command(args), cwd=ROOT/'backend', env=env, start_new_session=True)
            # Interactive multi-browser QA needs a longer lifetime; ordinary tests
            # retain their hard wall bound even if their Go timeout is overridden.
            result = child.wait(timeout=3600 if args in (['--webfixture'], ['--webfixture-economy']) else 900)
    except (Exception, KeyboardInterrupt) as error:
        # Never print captured commands/environments containing test credentials.
        print('❌ PostgreSQL integration failed: '+type(error).__name__, file=sys.stderr)
        if isinstance(error, subprocess.CalledProcessError):
            print('❌ Failed operation: '+' '.join(error.cmd[:3]), file=sys.stderr)
    finally:
        for s in old_handlers:
            signal.signal(s, signal.SIG_IGN)
        if child is not None and child.poll() is None:
            try:
                os.killpg(child.pid, signal.SIGTERM)
            except ProcessLookupError:
                pass  # Child exited between poll and signal; still reap/clean.
            try:
                child.wait(timeout=5)
            except subprocess.TimeoutExpired:
                try:
                    os.killpg(child.pid, signal.SIGKILL)
                except ProcessLookupError:
                    pass
                child.wait(timeout=5)
        good = cleanup(token, token, token, created)
        if not good:
            result = 1
        for s, handler in old_handlers.items():
            signal.signal(s, handler)
    return result

if __name__ == '__main__':
    sys.exit(main())
