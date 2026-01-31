#!/bin/bash
# path: scripts/build.sh
# start this script on project directory
# run : sh scripts/build.sh

set -euo pipefail # exit on error

go mod init CHIMERA
go get gopkg.in/yaml.v3



mkdir -p runtimes/python-3.14.2
cd runtimes/python-3.14.2
python -m venv .venv
# shellcheck source=/dev/null
# current path is runtimes/python-3.14.2
source .venv/bin/activate
python -m pip install --upgrade pip setuptools wheel
pip install requests
deactivate
cd ../..

mkdir -p runtimes/node-v25.4.0
cd runtimes/node-v25.4.0
pnpm init
pnpm add puppeteer socket.io-client
cd ../..
