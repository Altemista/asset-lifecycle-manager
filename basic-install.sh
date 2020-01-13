#!/usr/bin/env bash

set -e
set -u

url=https://github.com/Altemista/asset-lifecycle-manager/releases/latest/download

curl -fsSL https://github.com/operator-framework/operator-lifecycle-manager/releases/latest/download/install.sh | bash -s 0.12.0
curl -fsSL ${url}/aolm.yaml | kubectl apply -f -
