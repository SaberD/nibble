FROM golang:1.24-bookworm

ARG USERNAME=vscode
ARG USER_UID=1000
ARG USER_GID=1000

ENV DEBIAN_FRONTEND=noninteractive

# Base system utilities.
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        sudo

# Scanner build dependency.
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        libpcap-dev

# node and python
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        nodejs \
        npm \
        python3-pip \
        python3-venv

# dev user
RUN groupadd --gid ${USER_GID} ${USERNAME} \
    && useradd --uid ${USER_UID} --gid ${USER_GID} -m -s /bin/bash ${USERNAME} \
    && usermod -aG sudo ${USERNAME} \
    && echo "${USERNAME} ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/${USERNAME} \
    && chmod 440 /etc/sudoers.d/${USERNAME}

# goreleaser
RUN curl -sL https://github.com/goreleaser/goreleaser/releases/latest/download/goreleaser_Linux_x86_64.tar.gz | tar xz \
    && mv goreleaser /usr/local/bin/goreleaser

# Recording/demo tooling.
RUN apt-get update \
    && apt-get install -y \
        chromium \
        fonts-noto-color-emoji \
    && go install github.com/charmbracelet/vhs@latest \
    && cp /go/bin/vhs /usr/local/bin/vhs \
    && chown -R ${USERNAME}:${USERNAME} /go/pkg /go/bin

USER ${USERNAME}
WORKDIR /workspace
