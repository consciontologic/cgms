#!/usr/bin/env python3
"""Install the verified Linux x64 Flutter toolchain inside ignored workspace state."""
import hashlib
import json
from pathlib import Path
import tarfile
import urllib.request
ROOT=Path(__file__).resolve().parents[1]
VERSION='3.47.5'
SHA256='2132e990f236f8d22e7c6314b29a191a95b10d7cbcfec9b4e2e303d996652cbb'
BASE='https://storage.googleapis.com/flutter_infra_release/releases'

def main():
    data=json.load(urllib.request.urlopen(BASE+'/releases_linux.json',timeout=30))
    release=next(r for r in data['releases'] if r['version']==VERSION and r['channel']=='stable')
    if release['sha256']!=SHA256:raise ValueError('official release manifest no longer matches reviewed pin')
    directory=ROOT/'.local/toolchains';directory.mkdir(parents=True,exist_ok=True)
    archive=directory/f'flutter_linux_{VERSION}-stable.tar.xz'
    if not archive.exists():
        partial=archive.with_suffix('.partial')
        with urllib.request.urlopen(BASE+'/'+release['archive'],timeout=60) as response,partial.open('wb') as out:
            while block:=response.read(1024*1024):out.write(block)
        partial.rename(archive)
    with archive.open('rb') as stream:actual=hashlib.file_digest(stream,'sha256').hexdigest()
    if actual!=SHA256:raise ValueError('archive checksum mismatch; inspect/remove only the failed workspace archive')
    target=directory/f'flutter-{VERSION}';target.mkdir(exist_ok=True)
    with tarfile.open(archive) as stream:stream.extractall(target,filter='data')
    (directory/f'flutter-{VERSION}.release.json').write_text(json.dumps(release,indent=2)+'\n')
    print(f'Verified local Flutter {VERSION}; set PATH for the command/session to {target}/flutter/bin')

if __name__=='__main__':main()
