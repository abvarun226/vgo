#!/usr/bin/env python3
# PYTHON_ARGCOMPLETE_OK

import os
import sys
import argparse
import argcomplete

GO_DIR = '/usr/local/go'
GO_SYMLINK = '/usr/local/go/active'

help = '''
Usage:
python {} <version>
version: Go version to be set {}
'''

instructions = '''
export GO_BIN="/usr/local/go/active/bin"
export GO_PATH="${HOME}/Documents/gopath"
export PATH="${PATH}:${GO_BIN}:${GO_PATH}"
'''

def check():
    if GO_SYMLINK in os.environ['PATH']:
        return
    print('[ERROR] go path is not set in PATH. Add the following to ~/.zshrc file')
    print(instructions)

def set(version: str):
    versions = list()
    if version not in versions:
        print('\n[ERROR] unknown go version: {}. List of available versions are: {}\n'.format(version, '/'.join(versions)))
        return
    
    if os.path.exists(GO_SYMLINK):
        os.remove(GO_SYMLINK)

    gopath = GO_DIR + '/go' + version
    if not os.path.exists(gopath):
        print('\n[ERROR] go path not found: {}\n'.format(gopath))
        return

    os.symlink(gopath, GO_SYMLINK)
    os.system('go version')

def list():
    directories = os.listdir(GO_DIR)
    versions = []
    for dir in directories:
        if dir == 'active':
            continue
        versions.append(dir.replace('go', ''))
    return versions

if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('version', choices=list(), help='go version to be activated')
    argcomplete.autocomplete(parser)
    args = parser.parse_args()

    check()
    set(args.version)