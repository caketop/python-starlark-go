# https://github.com/pypa/manylinux/#docker-images
ARG BASE_IMAGE=quay.io/pypa/manylinux_2_28_x86_64

# local toolchains
ARG GO_URL=https://go.dev/dl/go1.25.4.linux-amd64.tar.gz
ARG CLANG_URL=https://github.com/llvm/llvm-project/releases/download/llvmorg-15.0.0/clang+llvm-15.0.0-x86_64-linux-gnu-rhel-8.4.tar.xz
ARG SHELLCHECK_URL=https://github.com/koalaman/shellcheck/releases/download/v0.11.0/shellcheck-v0.11.0.linux.x86_64.tar.xz
ARG MCFLY_URL=https://github.com/cantino/mcfly/releases/download/v0.9.3/mcfly-v0.9.3-x86_64-unknown-linux-musl.tar.gz

ARG USERNAME=builder
ARG USER_UID=501
ARG USER_GID=$USER_UID

FROM ${BASE_IMAGE} AS packages

# manylinux_2_28_x86_64 is Alma Linux; no deb infra
# ENV DEBIAN_FRONTEND=noninteractive
# RUN apt-get update && apt-get upgrade -y && \
#  apt-get install -y \
#  build-essential gdb less libffi-dev valgrind \
#  curl ca-certificates gnupg2 tar g++ gcc libc6-dev make pkg-config

RUN dnf -y upgrade && \
  dnf -y install \
  gcc gcc-c++ make glibc-devel libstdc++-devel \
  gdb less libffi-devel valgrind \
  curl ca-certificates gnupg2 tar \
  pkgconf-pkg-config

RUN curl -sS https://starship.rs/install.sh | sh -s -- -y


FROM packages AS go_install

ARG GO_URL

ADD ${GO_URL} /usr/src/go.tar.gz

RUN tar -C /opt -xvf /usr/src/go.tar.gz


FROM packages AS builder

COPY --from=go_install /opt/go/ /opt/go/

ENV PATH=/opt/go/bin:/opt/valgrind/bin:$PATH

ARG USERNAME USER_UID USER_GID

RUN groupadd -g ${USER_GID} ${USERNAME} && useradd -m -u ${USER_UID} -g ${USERNAME} -s /bin/bash ${USERNAME}

USER ${USERNAME}
WORKDIR /home/${USERNAME}
ENV USER=${USERNAME} SHELL=/bin/bash GOPATH=/home/${USERNAME}/go


FROM builder AS go_tools

RUN go install golang.org/x/tools/gopls@latest
RUN go install github.com/go-delve/delve/cmd/dlv@latest
RUN go install github.com/ramya-rao-a/go-outline@latest
RUN go install github.com/josharian/impl@latest
RUN go install github.com/fatih/gomodifytags@latest
RUN go install github.com/haya14busa/goplay/cmd/goplay@latest
RUN go install github.com/cweill/gotests/...@latest
RUN go install honnef.co/go/tools/cmd/staticcheck@latest
RUN go install mvdan.cc/gofumpt@latest
RUN go install mvdan.cc/sh/v3/cmd/shfmt@latest
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.45.2


FROM builder AS misc

ARG SHELLCHECK_URL MCFLY_URL CLANG_URL USERNAME

ADD --chown=${USERNAME} ${SHELLCHECK_URL} /usr/src/shellcheck.tar.xz

RUN mkdir /tmp/shellcheck && tar -C /tmp/shellcheck --strip-components=1 -xvf /usr/src/shellcheck.tar.xz

ADD --chown=${USERNAME} ${MCFLY_URL} /usr/src/mcfly.tar.gz

RUN mkdir /tmp/mcfly && tar -C /tmp/mcfly -xvf /usr/src/mcfly.tar.gz

ADD --chown=${USERNAME} ${CLANG_URL} /usr/src/clang.tar.xz

RUN mkdir /tmp/clang && tar -C /tmp/clang --strip-components=1 -xvf /usr/src/clang.tar.xz


FROM builder AS devcontainer

ARG USERNAME

RUN echo "set auto-load safe-path /" > /home/${USERNAME}/.gdbinit

COPY --from=go_tools --chown=${USERNAME} /home/${USERNAME}/go/bin/ /home/${USERNAME}/go/bin

RUN python3.11 -m venv /home/${USERNAME}/venv

ENV PATH=/home/${USERNAME}/venv/bin:/opt/python/cp311-cp311/bin:${PATH}

RUN python3.11 -m pip install black ipython isort memray pytest pytest-memray pytest-valgrind tox

COPY --from=misc --chown=${USERNAME} /tmp/shellcheck/shellcheck /home/${USERNAME}/.local/bin/shellcheck
COPY --from=misc --chown=${USERNAME} /tmp/mcfly/mcfly /home/${USERNAME}/.local/bin/mcfly
COPY --from=misc --chown=${USERNAME} /tmp/clang/bin/clang-format /home/${USERNAME}/.local/bin/clang-format

RUN echo 'eval "$(mcfly init bash)"' >> ~/.bashrc && touch ~/.bash_history
RUN echo 'eval "$(starship init bash)"' >> ~/.bashrc

ENV PATH=/home/${USERNAME}/go/bin:/home/${USERNAME}/.local/bin:$PATH
