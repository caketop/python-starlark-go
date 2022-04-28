#!/bin/sh

set -ex

GO_VERSION=1.18.1

if [ -e /etc/alpine-release ] && [ -z "$BASH_VERSION" ]; then
  apk add bash curl git go
  exit 0
fi

if [ -z "$BASH_VERSION" ]; then
  exec bash "$0"
fi

set -eou pipefail

if [ -e /etc/debian_version ]; then
  apt-get update
  DEBIAN_FRONTEND=noninteractive apt-get install -y curl git
fi

install_go() {
  git clone https://github.com/asdf-vm/asdf.git ~/.asdf --branch v0.10.0

  # shellcheck disable=SC1090
  . ~/.asdf/asdf.sh

  asdf plugin add golang
  asdf install golang "$GO_VERSION"

  ln -s ~/.asdf/installs/golang/${GO_VERSION}/go/bin/go /usr/local/bin/go
  ln -s ~/.asdf/installs/golang/${GO_VERSION}/go/bin/gofmt /usr/local/bin/gofmt
}

(install_go)

go version

env | sort
