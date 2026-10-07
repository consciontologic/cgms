#!/usr/bin/env python3
"""Run Flutter hot reload through the local same-origin Nginx proxy."""
from contextlib import contextmanager
import json
import os
from pathlib import Path
import signal
import subprocess
import sys
import time
import urllib.request
import urllib.error

from service_ops import ROOT, execute, restart

@contextmanager
def development_proxy(base):
    try:
        restart(base+['-f',str(ROOT/'deploy/compose/development.yaml')],'web')
        yield
    finally:
        restart(base,'web')

def main():
    # Bind to Docker's private host bridge, not every LAN interface.
    bridge=json.loads(execute(['docker','network','inspect','bridge']))[0]
    gateway=bridge['IPAM']['Config'][0]['Gateway']
    directory=ROOT/'.local/logs';directory.mkdir(parents=True,exist_ok=True,mode=0o700)
    path=directory/(str(time.time_ns())+'-flutter-dev.log')
    fd=os.open(path,os.O_WRONLY|os.O_CREAT|os.O_EXCL,0o600)
    base=['docker','compose','-f',str(ROOT/'deploy/compose/compose.yaml')]
    child=None
    with os.fdopen(fd,'w') as log:
        try:
            child=subprocess.Popen(['flutter','run','-d','web-server','--web-hostname',gateway,'--web-port','8082'],cwd=ROOT/'client',stdout=log,stderr=subprocess.STDOUT,stdin=subprocess.PIPE,text=True,start_new_session=True)
            deadline=time.monotonic()+120
            while True:
                if child.poll() is not None:raise RuntimeError('Flutter development server exited; inspect '+str(path.relative_to(ROOT)))
                try:
                    with urllib.request.urlopen('http://'+gateway+':8082',timeout=2) as response:
                        if response.status==200:break
                except (OSError,urllib.error.URLError):
                    if time.monotonic()>=deadline:raise TimeoutError('Flutter readiness expired; inspect '+str(path.relative_to(ROOT)))
                    time.sleep(.25)
            with development_proxy(base):
                print('Flutter dev: http://localhost:8080. Type r + Enter for reload, R for restart, q to stop. Private log: '+str(path.relative_to(ROOT)),flush=True)
                for line in sys.stdin:
                    if line.strip()=='q':break
                    if line.strip() in ('r','R'):
                        child.stdin.write(line);child.stdin.flush()
                    if child.poll() is not None:raise RuntimeError('Flutter dev stopped')
        finally:
            if child is not None and child.poll() is None:
                os.killpg(child.pid,signal.SIGTERM)
                try:child.wait(timeout=10)
                except subprocess.TimeoutExpired:os.killpg(child.pid,signal.SIGKILL);child.wait(timeout=5)

if __name__=='__main__':
    try:main()
    except (OSError,RuntimeError,TimeoutError,KeyboardInterrupt) as error:
        print(str(error),file=sys.stderr);sys.exit(1)
