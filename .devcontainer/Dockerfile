FROM golang:1.24-bookworm

ARG USERNAME=vscode
ARG USER_UID=1000
ARG USER_GID=1000

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        ffmpeg \
        fonts-noto-color-emoji \
        libasound2 \
        libatk-bridge2.0-0 \
        libatk1.0-0 \
        libcups2 \
        libdbus-1-3 \
        libgbm1 \
        libgtk-3-0 \
        libnspr4 \
        libnss3 \
        libpangocairo-1.0-0 \
        libpcap-dev \
        libxcomposite1 \
        libxdamage1 \
        libxfixes3 \
        libxinerama1 \
        libxkbcommon0 \
        libxrandr2 \
        libxtst6 \
        nodejs \
        npm \
        python3-pip \
        python3-venv \
        sudo \
    && rm -rf /var/lib/apt/lists/*

RUN groupadd --gid ${USER_GID} ${USERNAME} \
    && useradd --uid ${USER_UID} --gid ${USER_GID} -m -s /bin/bash ${USERNAME} \
    && usermod -aG sudo ${USERNAME} \
    && echo "${USERNAME} ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/${USERNAME} \
    && chmod 440 /etc/sudoers.d/${USERNAME}

RUN curl -sL https://github.com/goreleaser/goreleaser/releases/latest/download/goreleaser_Linux_x86_64.tar.gz | tar xz \
    && mv goreleaser /usr/local/bin/goreleaser

ARG TTYD_VERSION=1.7.7
RUN arch="$(uname -m)" \
    && case "${arch}" in \
        x86_64) ttyd_asset="ttyd.x86_64" ;; \
        aarch64|arm64) ttyd_asset="ttyd.aarch64" ;; \
        *) echo "Unsupported architecture: ${arch}" && exit 1 ;; \
    esac \
    && curl -fsSL "https://github.com/tsl0922/ttyd/releases/download/${TTYD_VERSION}/${ttyd_asset}" -o /usr/local/bin/ttyd \
    && chmod 0755 /usr/local/bin/ttyd

RUN npm install -g svg-term-cli \
    && go install github.com/charmbracelet/vhs@latest \
    && cp /go/bin/vhs /usr/local/bin/vhs \
    && chown -R ${USERNAME}:${USERNAME} /go/pkg /go/bin

USER ${USERNAME}
WORKDIR /workspace
