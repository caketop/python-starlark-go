#!/bin/sh

set -e
set -x

GO_VERSION=1.18.1

if [ -e /etc/alpine-release ]; then
  apk add bash curl git libc6-compat
fi

if [ -e /etc/debian_version ]; then
  apt-get update
  DEBIAN_FRONTEND=noninteractive apt-get install -y curl git
fi

git clone https://github.com/asdf-vm/asdf.git ~/.asdf --branch v0.10.0
. ~/.asdf/asdf.sh

asdf plugin add golang
asdf install golang "$GO_VERSION"

ln -s ~/.asdf/installs/golang/${GO_VERSION}/go/bin/go /usr/local/bin/go
ln -s ~/.asdf/installs/golang/${GO_VERSION}/go/bin/gofmt /usr/local/bin/gofmt
