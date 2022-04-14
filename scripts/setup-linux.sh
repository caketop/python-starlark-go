#!/usr/bin/env bash

set -xeuo pipefail

if [ ! -f "/usr/bin/go" ]; then
    curl -vvv -L -O https://go.dev/dl/go1.18.1.linux-amd64.tar.gz
    tar -xf go1.18.1.linux-amd64.tar.gz
    mv go /usr/local
    ln -s /usr/local/go/bin/go /usr/bin/go
fi
