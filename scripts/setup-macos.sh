#!/usr/bin/env bash

set -xeuo pipefail

if [ ! -f "/usr/local/bin/go" ]; then
    curl -vvv -L -O https://go.dev/dl/go1.18.1.darwin-amd64.tar.gz
    tar -xf go1.18.1.darwin-amd64.tar.gz
    mv go /tmp
    ln -s /tmp/go/bin/go /usr/local/bin/go
fi
