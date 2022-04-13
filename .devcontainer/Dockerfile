ARG BASE_IMAGE=quay.io/pypa/manylinux_2_24_x86_64
ARG GO_URL=https://go.dev/dl/go1.18.1.linux-amd64.tar.gz

ARG USERNAME=builder
ARG USER_UID=1000
ARG USER_GID=$USER_UID

FROM ${BASE_IMAGE}

ENV DEBIAN_FRONTENT=noninteractive

RUN apt-get update && apt-get upgrade -y && apt-get install libffi-dev

ARG GO_URL USERNAME USER_UID USER_GID

ADD ${GO_URL} /usr/src/go.tar.gz

RUN tar -C /opt -xvf /usr/src/go.tar.gz

ENV PATH=/opt/go/bin:$PATH

RUN groupadd -g ${USER_GID} ${USERNAME} && useradd -m -u ${USER_UID} -g ${USERNAME} -s /bin/bash ${USERNAME}

USER ${USERNAME}
WORKDIR /home/${USERNAME}
ENV USER=${USERNAME} SHELL=/bin/bash
