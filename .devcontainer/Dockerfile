ARG BASE_IMAGE=quay.io/pypa/manylinux_2_24_x86_64
ARG GO_URL=https://go.dev/dl/go1.18.1.linux-amd64.tar.gz

ARG USERNAME=builder
ARG USER_UID=1000
ARG USER_GID=$USER_UID


FROM ${BASE_IMAGE} AS packages

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get upgrade -y && apt-get install -y \
  build-essential clang-format gdb libffi-dev valgrind \
  curl ca-certificates gnupg2 tar g++ gcc libc6-dev make pkg-config


FROM packages AS go_install

ARG GO_URL

ADD ${GO_URL} /usr/src/go.tar.gz

RUN tar -C /opt -xvf /usr/src/go.tar.gz


FROM packages AS builder

COPY --from=go_install /opt/go/ /opt/go/
ENV PATH=/opt/go/bin:$PATH

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
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.45.2


FROM builder AS devcontainer

ARG USERNAME

COPY --from=go_tools /home/${USERNAME}/go/bin/ /home/${USERNAME}/go/bin

RUN python3.10 -m pip install --user black isort tox

ENV PATH=/home/${USERNAME}/go/bin:/home/${USERNAME}/.local/bin:$PATH
